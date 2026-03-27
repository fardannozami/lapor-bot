package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type ReportActivityUsecase struct {
	repo domain.ReportRepository
}

func NewReportActivityUsecase(repo domain.ReportRepository) *ReportActivityUsecase {
	return &ReportActivityUsecase{repo: repo}
}

func (uc *ReportActivityUsecase) Execute(ctx context.Context, userID, name string) (string, error) {
	report, err := uc.repo.GetReport(ctx, userID)
	if err != nil {
		return "", err
	}

	now := time.Now()
	today := domain.GetToday(now)

	if report != nil {
		lastReport := report.LastReportDate
		lastReportDate := domain.GetToday(lastReport)

		if lastReportDate.Equal(today) {
			return fmt.Sprintf("%s sudah laporan hari ini, ayo jangan curang! 😉", name), nil
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
		} else {
			// Missed week(s) — streak resets
			report.InactiveDays = daysSinceLastReport
			report.ComebackStreak = 1
			report.Streak = 1
		}
		report.ActivityCount++
		report.Name = name
		report.LastReportDate = now
	} else {
		report = &domain.Report{
			UserID:         userID,
			Name:           name,
			Streak:         1,
			ActivityCount:  1,
			LastReportDate: now,
			MaxStreak:      0,
			TotalPoints:    0,
			Achievements:   "",
			ComebackStreak: 1,
			InactiveDays:   0,
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

	// 2. Store old level to detect level-up
	oldLevel := domain.GetLevel(report.TotalPoints)

	// 3. Check for new standard achievements
	newAchievements := domain.CheckNewAchievements(report)
	var unlockedNames []string
	pointsGained := 0

	for _, ach := range newAchievements {
		report.Achievements = domain.AddAchievement(report.Achievements, ach.ID)
		report.TotalPoints += ach.Points
		pointsGained += ach.Points
		unlockedNames = append(unlockedNames, fmt.Sprintf("%s (%d pts)", ach.Name, ach.Points))
	}

	// 4. Check for new comeback achievements
	comebackAchievements := domain.CheckComebackAchievements(report)
	for _, ach := range comebackAchievements {
		report.Achievements = domain.AddAchievement(report.Achievements, ach.ID)
		report.TotalPoints += ach.Points
		pointsGained += ach.Points
		unlockedNames = append(unlockedNames, fmt.Sprintf("%s (%d pts)", ach.Name, ach.Points))
	}

	// 5. Detect level-up
	newLevel := domain.GetLevel(report.TotalPoints)
	leveledUp := newLevel.Tier > oldLevel.Tier

	if err := uc.repo.UpsertReport(ctx, report); err != nil {
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
		// Normal report message
		response = fmt.Sprintf("Laporan diterima, %s sudah berkeringat %d hari. Lanjutkan 🔥 (streak %d minggu)", name, report.ActivityCount, report.Streak)
	}

	// Append Gamification Notifications
	if newRecord {
		response += fmt.Sprintf("\n\n🏆 New Personal Best Streak: %d minggu!", report.Streak)
	}

	if leveledUp {
		response += fmt.Sprintf("\n\n⬆️ LEVEL UP! %s → %s %s", oldLevel.Name, newLevel.Name, newLevel.Icon)
	}

	if len(unlockedNames) > 0 {
		response += "\n\n🎉 ACHIEVEMENT UNLOCKED! 🎉"
		for _, name := range unlockedNames {
			response += fmt.Sprintf("\n🏅 %s", name)
		}
	}

	if pointsGained > 0 {
		response += fmt.Sprintf("\n\n⭐ +%d points (Total: %d)", pointsGained, report.TotalPoints)
	}

	// Show progress bar
	progressBar := domain.FormatProgressBar(report.TotalPoints)
	response += fmt.Sprintf("\n%s", progressBar)

	return response, nil
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
