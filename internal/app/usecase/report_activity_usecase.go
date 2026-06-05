package usecase

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type ReportActivityUsecase struct {
	repo  domain.ReportRepository
	locks sync.Map
}

func NewReportActivityUsecase(repo domain.ReportRepository) *ReportActivityUsecase {
	return &ReportActivityUsecase{repo: repo}
}

func (uc *ReportActivityUsecase) userLock(userID string) *sync.Mutex {
	lock, _ := uc.locks.LoadOrStore(userID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func (uc *ReportActivityUsecase) Execute(ctx context.Context, userID, name string, workout *domain.HevyWorkout) (string, error) {
	lock := uc.userLock(userID)
	lock.Lock()
	defer lock.Unlock()

	report, err := uc.repo.GetReport(ctx, userID)
	if err != nil {
		return "", err
	}

	now := time.Now()
	today := domain.GetToday(now)

	streakFreezeUsed := false

	if report != nil {
		lastReport := report.LastReportDate
		lastReportDate := domain.GetToday(lastReport)

		if lastReportDate.Equal(today) {
			duplicateMsg := fmt.Sprintf("%s sudah laporan hari ini, ayo jangan curang! 😉", name)
			if workout != nil {
				duplicateMsg += "\n" + uc.formatWorkout(workout)
			}
			return duplicateMsg, nil
		}

		daysSinceLastReport := int(math.Round(today.Sub(lastReportDate).Hours() / 24))

		currentWeekStart := domain.GetStartOfISOWeek(today)
		lastWeekStart := domain.GetStartOfISOWeek(lastReportDate)
		weeksSinceLastReport := int(math.Round(currentWeekStart.Sub(lastWeekStart).Hours() / (24 * 7)))

		if currentWeekStart.Equal(lastWeekStart) {
			// Same week — streak stays the same, nothing to increment
		} else if weeksSinceLastReport == 1 {
			// Consecutive week — streak continues
			report.Streak++
			report.ComebackStreak++ // also increment comeback streak if active
		} else if weeksSinceLastReport == 2 && report.StreakFreezes > 0 {
			// Exactly 1 week missed — auto-consume streak freeze to protect streak
			report.StreakFreezes--
			report.Streak++
			report.ComebackStreak++
			streakFreezeUsed = true
		} else {
			// Missed week(s) — streak resets
			report.InactiveDays = daysSinceLastReport
			report.ComebackStreak = 1
			report.Streak = 1
		}
		report.ActivityCount++
		report.SeasonalActivityCount++
		if report.Streak > report.SeasonalMaxStreak {
			report.SeasonalMaxStreak = report.Streak
		}

		// Handle Centurion Cycle Transition
		isNewCycle := false
		if report.ActivityCount > 100 {
			report.ActivityCount = 1
			report.CenturionCycles++
			isNewCycle = true
		}

		// report.Name = name // Removed: don't update name during report
		report.LastReportDate = now
		name = report.Name // Use the stored name for the response message

		if isNewCycle {
			// Special handling for new cycle can be added here or in the response construction
		}
	} else {
		report = &domain.Report{
			UserID:                userID,
			Name:                  name,
			Streak:                1,
			ActivityCount:         1,
			SeasonalActivityCount: 1,
			SeasonalMaxStreak:     1,
			LastReportDate:        now,
			MaxStreak:             0,
			TotalPoints:           0,
			SeasonalPoints:        0,
			Achievements:          "",
			SeasonalAchievements:  "",
			ComebackStreak:        1,
			InactiveDays:          0,
			StreakFreezes:         1,
		}
	}

	// Logic for MaxStreak: it should be at least 1 if they reported once
	if report.MaxStreak == 0 {
		report.MaxStreak = 1
	}

	// 1. Update Max Streak
	newRecord := false
	if report.Streak > report.MaxStreak {
		report.MaxStreak = report.Streak
		if report.Streak > 1 {
			newRecord = true
		}
	}

	// 2. Calculate base EXP + bonuses
	basePoints := 10
	streakBonus := 2 * (report.Streak - 1)
	if streakBonus < 0 {
		streakBonus = 0
	}
	seasonalFirstBonus := 0
	if report.SeasonalActivityCount == 1 {
		seasonalFirstBonus = 5
	}
	reportPoints := basePoints + streakBonus + seasonalFirstBonus
	report.TotalPoints += reportPoints
	report.SeasonalPoints += reportPoints

	// 3. Store old level to detect level-up
	oldLevel := domain.GetLevel(report.TotalPoints)

	// 3. Check for new season badges. Badge progress resets every season, but
	// lifetime EXP/level remains preserved.
	newAchievements := domain.CheckNewSeasonAchievements(report)
	var unlockedNames []string
	pointsGained := 0

	for _, ach := range newAchievements {
		report.SeasonalAchievements = domain.AddAchievement(report.SeasonalAchievements, ach.ID)
		if !domain.HasAchievement(report.Achievements, ach.ID) {
			report.Achievements = domain.AddAchievement(report.Achievements, ach.ID)
		}
		report.TotalPoints += ach.Points
		report.SeasonalPoints += ach.Points
		pointsGained += ach.Points
		unlockedNames = append(unlockedNames, fmt.Sprintf("%s %s (%d pts)", ach.DisplayEmoji, ach.Name, ach.Points))
	}

	// 4. Check for new comeback achievements
	comebackAchievements := domain.CheckComebackAchievements(report)
	for _, ach := range comebackAchievements {
		report.Achievements = domain.AddAchievement(report.Achievements, ach.ID)
		report.TotalPoints += ach.Points
		pointsGained += ach.Points
		unlockedNames = append(unlockedNames, fmt.Sprintf("%s %s (%d pts)", ach.DisplayEmoji, ach.Name, ach.Points))
	}

	freezeAwarded := false
	for _, ach := range newAchievements {
		if ach.ID == "streak_4" && report.StreakFreezes < 2 {
			report.StreakFreezes++
			freezeAwarded = true
		}
	}

	// 5. Detect level-up
	newLevel := domain.GetLevel(report.TotalPoints)
	leveledUp := newLevel.Tier > oldLevel.Tier

	if err := uc.repo.UpsertReportWithActivity(ctx, report, today); err != nil {
		return "", err
	}

	// 6. Construct Response
	isComeback := report.InactiveDays > 3 && report.Streak == 1
	var response string

	if isComeback {
		// Special comeback message
		response = fmt.Sprintf("🎉 WELCOME BACK, %s! 🎉\n", name)
		response += fmt.Sprintf("Kamu kembali setelah %d hari absen. Itu butuh keberanian! 💪\n", report.InactiveDays)
		response += "Streak kamu direset, tapi totalmu tetap tersimpan.\n"
		if report.JobClass != "" {
			response += fmt.Sprintf("🧭 Job: %s\n", domain.FormatJobClass(report.JobClass))
		}
		response += fmt.Sprintf("\n📊 Level: %s (Total: %d pts)\n", domain.FormatLevel(report.TotalPoints), report.TotalPoints)
		response += fmt.Sprintf("📅 Total hari aktif: %d\n", report.ActivityCount)

		// Show comeback challenge info
		nextComebackAch := uc.getNextComebackTarget(report)
		if nextComebackAch != nil {
			response += fmt.Sprintf("\n🔥 Comeback Challenge dimulai! Raih %d minggu berturut-turut untuk unlock \"%s\"!", nextComebackAch.MinComebackStreak, nextComebackAch.Name)
		} else {
			response += "\n🔥 Comeback Challenge dimulai! Ayo bangun streak-mu kembali!"
		}
	} else {
		// Normal report message with EXP breakdown
		cyclePrefix := ""
		if report.CenturionCycles > 0 {
			cyclePrefix = fmt.Sprintf("[C%d] ", report.CenturionCycles+1)
		}

		expBreakdown := fmt.Sprintf("⭐ +%d pts", basePoints)
		if streakBonus > 0 {
			expBreakdown += fmt.Sprintf(" (streak bonus +%d)", streakBonus)
		}
		if seasonalFirstBonus > 0 {
			expBreakdown += " (first season report +5)"
		}

		response = fmt.Sprintf("Laporan diterima, %s%s sudah berkeringat %d hari. Lanjutkan 🔥 (streak %d minggu)\n%s",
			cyclePrefix, name, report.ActivityCount, report.Streak, expBreakdown)
		if report.JobClass != "" {
			response += fmt.Sprintf("\n🧭 Job: %s", domain.FormatJobClass(report.JobClass))
		}
	}

	// 7. Add Centurion Milestone Message
	if report.ActivityCount == 1 && report.CenturionCycles > 0 && !isComeback {
		response = "🔥 *ERA BARU DIMULAI!* 🔥\n" +
			fmt.Sprintf("Selamat %s, Anda telah menyelesaikan 100 hari sebelumnya.\n", name) +
			fmt.Sprintf("Sekarang Anda memulai *Siklus %d (Hari ke-1)*. Terus jaga konsistensi! 💪\n\n", report.CenturionCycles+1) +
			response
	} else if report.ActivityCount == 100 {
		response = "🎊 *LUAR BIASA!* 🎊\n" +
			fmt.Sprintf("Selamat %s, Anda mencapai *HARI KE-100*! 💯\n", name) +
			"Anda sekarang resmi menyandang gelar *CENTURION 🛡️*. Nama Anda telah diabdikan dalam jajaran legenda grup!\n\n" +
			response
	}

	// Append Workout Details if present
	if workout != nil {
		response += "\n" + uc.formatWorkout(workout)
	}

	// Append Gamification Notifications
	if streakFreezeUsed {
		response += fmt.Sprintf("\n\n❄️ *Streak Freeze terpakai!* Streak kamu aman. (Sisa freeze: %d)", report.StreakFreezes)
	}

	if newRecord {
		response += fmt.Sprintf("\n\n🏆 New Personal Best Streak: %d minggu!", report.Streak)
	}

	if leveledUp {
		response += fmt.Sprintf("\n\n⬆️ LEVEL UP! %s → %s %s", oldLevel.Name, newLevel.Name, newLevel.Icon)
	}

	if len(unlockedNames) > 0 {
		response += "\n\n🎉 *SEASON BADGE UNLOCKED!* 🎉"

		for _, ach := range newAchievements {
			response += fmt.Sprintf("\n\n%s *%s* (+%d pts)", ach.DisplayEmoji, ach.Name, ach.Points)
			if ach.UnlockMessage != "" {
				response += fmt.Sprintf("\n_%s_", ach.UnlockMessage)
			}
		}
		for _, ach := range comebackAchievements {
			response += fmt.Sprintf("\n\n%s *%s* (+%d pts)", ach.DisplayEmoji, ach.Name, ach.Points)
			if ach.UnlockMessage != "" {
				response += fmt.Sprintf("\n_%s_", ach.UnlockMessage)
			}
		}

		if freezeAwarded {
			response += fmt.Sprintf("\n\n❄️ Bonus: +1 Streak Freeze! (Total: %d)", report.StreakFreezes)
		}

		response += fmt.Sprintf("\n\n💬 _\"%s\"_", RandomQuote())
	}

	totalPointsGained := reportPoints + pointsGained
	if totalPointsGained > 0 {
		response += fmt.Sprintf("\n\n💰 Total: +%d points (Lifetime: %d | Season: %d)", totalPointsGained, report.TotalPoints, report.SeasonalPoints)
	}

	// Show progress bar
	progressBar := domain.FormatProgressBar(report.TotalPoints)
	response += fmt.Sprintf("\n%s", progressBar)

	return response, nil
}

func (uc *ReportActivityUsecase) formatWorkout(workout *domain.HevyWorkout) string {
	if workout == nil {
		return ""
	}

	res := fmt.Sprintf("\n🏋️‍♂️ *%s Detail:*\n", workout.Title)
	if len(workout.Exercises) > 0 {
		res += "\n"
		for _, ex := range workout.Exercises {
			res += fmt.Sprintf("%s\n", ex)
		}
	}
	if workout.Time != "" {
		res += fmt.Sprintf("... ⏱️ Time: %s", workout.Time)
	}
	return res
}

// getNextComebackTarget finds the next comeback achievement the user can target.
func (uc *ReportActivityUsecase) getNextComebackTarget(report *domain.Report) *domain.ComebackAchievement {
	for i := range domain.AllComebackAchievements {
		a := &domain.AllComebackAchievements[i]
		if !domain.HasAchievement(report.Achievements, a.ID) &&
			report.InactiveDays >= a.MinInactiveDays {
			return a
		}
	}
	return nil
}
