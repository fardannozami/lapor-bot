package usecase

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

const MaxDailyReports = 3

type ReportActivityUsecase struct {
	repo  domain.ReportRepository
	locks sync.Map
}

type reportActivityOptions struct {
	sideQuestCount int
	activityText   string
	now            time.Time
}

func NewReportActivityUsecase(repo domain.ReportRepository) *ReportActivityUsecase {
	return &ReportActivityUsecase{repo: repo}
}

func (uc *ReportActivityUsecase) userLock(userID string) *sync.Mutex {
	lock, _ := uc.locks.LoadOrStore(userID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func (uc *ReportActivityUsecase) Execute(ctx context.Context, userID, name string, workout *domain.HevyWorkout) (string, error) {
	return uc.execute(ctx, userID, name, workout, reportActivityOptions{})
}

func (uc *ReportActivityUsecase) ExecuteSideQuest(ctx context.Context, userID, name, activityText string, completedCount int, now time.Time) (string, error) {
	if completedCount < 1 {
		completedCount = 1
	}
	return uc.execute(ctx, userID, name, nil, reportActivityOptions{
		sideQuestCount: completedCount,
		activityText:   activityText,
		now:            now,
	})
}

func (uc *ReportActivityUsecase) execute(ctx context.Context, userID, name string, workout *domain.HevyWorkout, opts reportActivityOptions) (string, error) {
	lock := uc.userLock(userID)
	lock.Lock()
	defer lock.Unlock()

	report, err := uc.repo.GetReport(ctx, userID)
	if err != nil {
		return "", err
	}

	now := opts.now
	if now.IsZero() {
		now = time.Now()
	}
	today := domain.GetToday(now)
	isSideQuest := opts.sideQuestCount > 0

	var dailyCount int
	if report != nil && domain.GetToday(report.LastReportDate).Equal(today) {
		dailyCount, err = uc.repo.GetDailyActivityCount(ctx, userID, today)
		if err != nil {
			return "", err
		}
	}

	if dailyCount >= MaxDailyReports {
		msg := fmt.Sprintf("%s sudah laporan %dx hari ini, ayo jangan curang! 😉\n\nBatas harian adalah %d laporan: laporan pertama untuk progres utama, laporan ke-2 dan ke-3 hanya bonus ½ XP. Kalau tadi salah input, pakai #cancel untuk hapus laporan terakhir atau #cancel-all untuk hapus semua laporan hari ini. 🙏", report.Name, MaxDailyReports, MaxDailyReports)
		msg += fmt.Sprintf("\n\n💬 _\"%s\"_", RandomQuote())
		return msg, nil
	}

	isRepeatReport := dailyCount > 0
	isFullReport := !isRepeatReport

	streakFreezeUsed := false

	if report != nil {
		storedName := report.Name
		name = reportName(storedName, name)
		shouldUpdateStoredName := isUnknownName(storedName) && strings.TrimSpace(storedName) != name
		if shouldUpdateStoredName {
			report.Name = name
		}

		lastReport := report.LastReportDate
		lastReportDate := domain.GetToday(lastReport)

		if lastReportDate.Equal(today) {
			if shouldUpdateStoredName {
				if err := uc.repo.UpsertReport(ctx, report); err != nil {
					return "", err
				}
			}
			report.LastReportDate = now
			name = report.Name
		} else {
			daysSinceLastReport := int(math.Round(today.Sub(lastReportDate).Hours() / 24))

			currentWeekStart := domain.GetStartOfISOWeek(today)
			lastWeekStart := domain.GetStartOfISOWeek(lastReportDate)
			weeksSinceLastReport := int(math.Round(currentWeekStart.Sub(lastWeekStart).Hours() / (24 * 7)))

			if isFullReport {
				if currentWeekStart.Equal(lastWeekStart) {
				} else if weeksSinceLastReport == 1 {
					report.Streak++
					report.ComebackStreak++
				} else if weeksSinceLastReport == 2 && report.StreakFreezes > 0 {
					report.StreakFreezes--
					report.Streak++
					report.ComebackStreak++
					streakFreezeUsed = true
				} else {
					report.InactiveDays = daysSinceLastReport
					report.ComebackStreak = 1
					report.Streak = 1
				}
				report.ActivityCount++
				report.SeasonalActivityCount++
				if report.Streak > report.SeasonalMaxStreak {
					report.SeasonalMaxStreak = report.Streak
				}

				if report.ActivityCount > 100 {
					report.ActivityCount = 1
					report.CenturionCycles++
				}
			}

			report.LastReportDate = now
			name = report.Name
		}
	} else {
		name = reportName("", name)
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

	if report.MaxStreak == 0 {
		report.MaxStreak = 1
	}
	oldNumericLevel := domain.NumericLevelFromTotalPoints(report.TotalPoints)

	newRecord := false
	if isFullReport && report.Streak > report.MaxStreak {
		report.MaxStreak = report.Streak
		if report.Streak > 1 {
			newRecord = true
		}
	}

	basePoints := 10
	streakBonus := 0
	seasonalFirstBonus := 0
	if isFullReport {
		streakBonus = 2 * (report.Streak - 1)
		if streakBonus < 0 {
			streakBonus = 0
		}
		if report.SeasonalActivityCount == 1 {
			seasonalFirstBonus = 5
		}
	}
	if isSideQuest {
		basePoints = 5
		streakBonus = 0
		seasonalFirstBonus = 0
		report.TotalSideQuests += opts.sideQuestCount
		report.SeasonalSideQuests += opts.sideQuestCount
	}
	reportPoints := basePoints + streakBonus + seasonalFirstBonus
	if isRepeatReport && !isSideQuest {
		reportPoints = reportPoints / 2
	}
	report.TotalPoints += reportPoints
	report.SeasonalPoints += reportPoints

	var newAchievements []domain.Achievement
	var comebackAchievements []domain.ComebackAchievement
	pointsGained := 0

	if isFullReport && !isSideQuest {
		newAchievements = domain.CheckNewSeasonAchievements(report)
		for _, ach := range newAchievements {
			report.SeasonalAchievements = domain.AddAchievement(report.SeasonalAchievements, ach.ID)
			if !domain.HasAchievement(report.Achievements, ach.ID) {
				report.Achievements = domain.AddAchievement(report.Achievements, ach.ID)
			}
			report.TotalPoints += ach.Points
			report.SeasonalPoints += ach.Points
			pointsGained += ach.Points
		}

		comebackAchievements = domain.CheckComebackAchievements(report)
		for _, ach := range comebackAchievements {
			report.Achievements = domain.AddAchievement(report.Achievements, ach.ID)
			report.TotalPoints += ach.Points
			pointsGained += ach.Points
		}
	}

	freezeAwarded := false
	if isFullReport && !isSideQuest {
		for _, ach := range newAchievements {
			if ach.ID == "streak_4" && report.StreakFreezes < 2 {
				report.StreakFreezes++
				freezeAwarded = true
			}
		}
	}

	report.Level = domain.NumericLevelFromTotalPoints(report.TotalPoints)
	leveledUp := report.Level > oldNumericLevel

	if err := uc.repo.UpsertReportWithActivity(ctx, report, today); err != nil {
		return "", err
	}
	goalCompleted, err := NewGoalUsecase(uc.repo).RecordActivity(ctx, userID, now, goalActivityTextWithFallback(workout, opts.activityText))
	if err != nil {
		return "", err
	}

	isComeback := isFullReport && report.InactiveDays > 3 && report.Streak == 1
	var response string

	if isComeback {
		response = fmt.Sprintf("🎉 WELCOME BACK, %s! 🎉\n", name)
		response += fmt.Sprintf("Kamu kembali setelah %d hari absen. Itu butuh keberanian! 💪\n", report.InactiveDays)
		response += "Streak kamu direset, tapi totalmu tetap tersimpan.\n"
		if report.JobClass != "" {
			response += fmt.Sprintf("🧭 Job: %s\n", domain.FormatJobClass(report.JobClass))
		}
		response += fmt.Sprintf("\n📊 Level: Lv.%d • %s (Total: %d pts)\n", report.Level, domain.FormatLevel(report.TotalPoints), report.TotalPoints)
		response += fmt.Sprintf("📅 Total hari aktif: %d\n", report.ActivityCount)

		nextComebackAch := uc.getNextComebackTarget(report)
		if nextComebackAch != nil {
			response += fmt.Sprintf("\n🔥 Comeback Challenge dimulai! Raih %d minggu berturut-turut untuk unlock \"%s\"!", nextComebackAch.MinComebackStreak, nextComebackAch.Name)
		} else {
			response += "\n🔥 Comeback Challenge dimulai! Ayo bangun streak-mu kembali!"
		}
	} else {
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
		if isSideQuest {
			expBreakdown += " (side quest, ½ XP)"
		} else if isRepeatReport {
			expBreakdown += " (repeat report, ½ XP)"
		}

		if isSideQuest {
			response = fmt.Sprintf("Side quest diterima, %s%s menyelesaikan %d side quest. Tetap dihitung untuk streak, stats, leaderboard, dan goal. 🔥\n%s",
				cyclePrefix, name, opts.sideQuestCount, expBreakdown)
		} else if isRepeatReport {
			response = fmt.Sprintf("Laporan diterima (laporan ke-%d hari ini), %s%s sudah berkeringat %d hari. Lanjutkan 🔥 (streak %d minggu)\n%s",
				dailyCount+1, cyclePrefix, name, report.ActivityCount, report.Streak, expBreakdown)
		} else {
			response = fmt.Sprintf("Laporan diterima, %s%s sudah berkeringat %d hari. Lanjutkan 🔥 (streak %d minggu)\n%s",
				cyclePrefix, name, report.ActivityCount, report.Streak, expBreakdown)
		}
		if report.JobClass != "" {
			response += fmt.Sprintf("\n🧭 Job: %s", domain.FormatJobClass(report.JobClass))
		}
	}

	if isFullReport && report.ActivityCount == 1 && report.CenturionCycles > 0 && !isComeback {
		response = "🔥 *ERA BARU DIMULAI!* 🔥\n" +
			fmt.Sprintf("Selamat %s, Anda telah menyelesaikan 100 hari sebelumnya.\n", name) +
			fmt.Sprintf("Sekarang Anda memulai *Siklus %d (Hari ke-1)*. Terus jaga konsistensi! 💪\n\n", report.CenturionCycles+1) +
			response
	} else if isFullReport && report.ActivityCount == 100 {
		response = "🎊 *LUAR BIASA!* 🎊\n" +
			fmt.Sprintf("Selamat %s, Anda mencapai *HARI KE-100*! 💯\n", name) +
			"Anda sekarang resmi menyandang gelar *CENTURION 🛡️*. Nama Anda telah diabdikan dalam jajaran legenda grup!\n\n" +
			response
	}

	if workout != nil {
		response += "\n" + uc.formatWorkout(workout)
	}

	if streakFreezeUsed {
		response += fmt.Sprintf("\n\n❄️ *Streak Freeze terpakai!* Streak kamu aman. (Sisa freeze: %d)", report.StreakFreezes)
	}

	if newRecord {
		response += fmt.Sprintf("\n\n🏆 New Personal Best Streak: %d minggu!", report.MaxStreak)
	}

	if goalCompleted {
		response += "\n\n🎯 *Goal minggu ini tercapai!* Konsistensi harianmu sudah sesuai target. Mantap, champ! 🏆"
	}

	if leveledUp {
		response += fmt.Sprintf("\n\n⚔️ *LEVEL UP!* Lv.%d → Lv.%d", oldNumericLevel, report.Level)
	}

	if len(newAchievements)+len(comebackAchievements) > 0 {
		response += "\n\n🏅 *Badge baru:*"
		for _, ach := range newAchievements {
			response += fmt.Sprintf("\n%s %s (+%d pts)", ach.DisplayEmoji, ach.Name, ach.Points)
		}
		for _, ach := range comebackAchievements {
			response += fmt.Sprintf("\n%s %s (+%d pts)", ach.DisplayEmoji, ach.Name, ach.Points)
		}

		if freezeAwarded {
			response += fmt.Sprintf("\n\n❄️ Bonus: +1 Streak Freeze! (Total: %d)", report.StreakFreezes)
		}

		response += "\nDetail badge & cerita unlock: #achievements"
	}

	response += fmt.Sprintf("\n\n💬 _\"%s\"_", RandomQuote())

	totalPointsGained := reportPoints + pointsGained
	if totalPointsGained > 0 {
		response += fmt.Sprintf("\n\n💰 Total: +%d points (Lifetime: %d | Season: %d)", totalPointsGained, report.TotalPoints, report.SeasonalPoints)
	}

	response += fmt.Sprintf("\n%s", domain.FormatNumericLevelProgressBar(report.TotalPoints))
	response += fmt.Sprintf("\n%s", domain.FormatProgressBar(report.TotalPoints))

	return response, nil
}

func (uc *ReportActivityUsecase) ExecuteYesterday(ctx context.Context, userID, name string, workout *domain.HevyWorkout) (string, error) {
	lock := uc.userLock(userID)
	lock.Lock()
	defer lock.Unlock()

	report, err := uc.repo.GetReport(ctx, userID)
	if err != nil {
		return "", err
	}

	now := time.Now()
	today := domain.GetToday(now)
	yesterday := today.AddDate(0, 0, -1)

	dailyCount, err := uc.repo.GetDailyActivityCount(ctx, userID, yesterday)
	if err != nil {
		return "", err
	}
	if dailyCount >= MaxDailyReports {
		return fmt.Sprintf("%s sudah mencapai batas maksimal %dx laporan untuk hari kemarin. 🙏", report.Name, MaxDailyReports), nil
	}

	if report != nil && domain.GetToday(report.LastReportDate).Equal(today) {
		return fmt.Sprintf("%s sudah laporan hari ini. Gunakan #lapor untuk laporan tambahan hari ini. 💪", report.Name), nil
	}

	if report != nil {
		lastReportDate := domain.GetToday(report.LastReportDate)
		daysSinceLastReport := int(math.Round(today.Sub(lastReportDate).Hours() / 24))
		if daysSinceLastReport < 2 {
			return fmt.Sprintf("%s sudah laporan dalam 2 hari terakhir. Tidak perlu lapor ulang untuk hari kemarin. 🙏", report.Name), nil
		}
	}

	streakFreezeUsed := false

	if report != nil {
		storedName := report.Name
		name = reportName(storedName, name)
		shouldUpdateStoredName := isUnknownName(storedName) && strings.TrimSpace(storedName) != name
		if shouldUpdateStoredName {
			report.Name = name
		}

		lastReportDate := domain.GetToday(report.LastReportDate)
		currentWeekStart := domain.GetStartOfISOWeek(yesterday)
		lastWeekStart := domain.GetStartOfISOWeek(lastReportDate)
		weeksSinceLastReport := int(math.Round(currentWeekStart.Sub(lastWeekStart).Hours() / (24 * 7)))

		if currentWeekStart.Equal(lastWeekStart) {
		} else if weeksSinceLastReport == 1 {
			report.Streak++
			report.ComebackStreak++
		} else if weeksSinceLastReport == 2 && report.StreakFreezes > 0 {
			report.StreakFreezes--
			report.Streak++
			report.ComebackStreak++
			streakFreezeUsed = true
		} else {
			report.InactiveDays = int(math.Round(today.Sub(lastReportDate).Hours()/24)) - 1
			report.ComebackStreak = 1
			report.Streak = 1
		}
		report.ActivityCount++
		report.SeasonalActivityCount++
		if report.Streak > report.SeasonalMaxStreak {
			report.SeasonalMaxStreak = report.Streak
		}

		if report.ActivityCount > 100 {
			report.ActivityCount = 1
			report.CenturionCycles++
		}

		report.LastReportDate = now.AddDate(0, 0, -1)
		name = report.Name
	} else {
		name = reportName("", name)
		report = &domain.Report{
			UserID:                userID,
			Name:                  name,
			Streak:                1,
			ActivityCount:         1,
			SeasonalActivityCount: 1,
			SeasonalMaxStreak:     1,
			LastReportDate:        now.AddDate(0, 0, -1),
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

	if report.MaxStreak == 0 {
		report.MaxStreak = 1
	}
	oldNumericLevel := domain.NumericLevelFromTotalPoints(report.TotalPoints)

	newRecord := false
	if report.Streak > report.MaxStreak {
		report.MaxStreak = report.Streak
		if report.Streak > 1 {
			newRecord = true
		}
	}

	basePoints := 5
	streakBonus := 0
	seasonalFirstBonus := 0
	if report.Streak > 1 {
		streakBonus = (2 * (report.Streak - 1)) / 2
		if streakBonus < 0 {
			streakBonus = 0
		}
	}
	if report.SeasonalActivityCount == 1 {
		seasonalFirstBonus = 2
	}
	reportPoints := basePoints + streakBonus + seasonalFirstBonus
	report.TotalPoints += reportPoints
	report.SeasonalPoints += reportPoints

	newAchievements := domain.CheckNewSeasonAchievements(report)
	pointsGained := 0

	for _, ach := range newAchievements {
		report.SeasonalAchievements = domain.AddAchievement(report.SeasonalAchievements, ach.ID)
		if !domain.HasAchievement(report.Achievements, ach.ID) {
			report.Achievements = domain.AddAchievement(report.Achievements, ach.ID)
		}
		report.TotalPoints += ach.Points
		report.SeasonalPoints += ach.Points
		pointsGained += ach.Points
	}

	comebackAchievements := domain.CheckComebackAchievements(report)
	for _, ach := range comebackAchievements {
		report.Achievements = domain.AddAchievement(report.Achievements, ach.ID)
		report.TotalPoints += ach.Points
		pointsGained += ach.Points
	}

	freezeAwarded := false
	for _, ach := range newAchievements {
		if ach.ID == "streak_4" && report.StreakFreezes < 2 {
			report.StreakFreezes++
			freezeAwarded = true
		}
	}

	report.Level = domain.NumericLevelFromTotalPoints(report.TotalPoints)
	leveledUp := report.Level > oldNumericLevel

	if err := uc.repo.UpsertReportWithActivity(ctx, report, yesterday); err != nil {
		return "", err
	}
	goalCompleted, err := NewGoalUsecase(uc.repo).RecordActivity(ctx, userID, yesterday, goalActivityText(workout))
	if err != nil {
		return "", err
	}

	isComeback := report.InactiveDays > 3 && report.Streak == 1
	var response string

	if isComeback {
		response = fmt.Sprintf("🎉 WELCOME BACK, %s! 🎉\n", name)
		response += fmt.Sprintf("Kamu kembali setelah %d hari absen. Itu butuh keberanian! 💪\n", report.InactiveDays)
		response += "Streak kamu direset, tapi totalmu tetap tersimpan.\n"
		if report.JobClass != "" {
			response += fmt.Sprintf("🧭 Job: %s\n", domain.FormatJobClass(report.JobClass))
		}
		response += fmt.Sprintf("\n📊 Level: Lv.%d • %s (Total: %d pts)\n", report.Level, domain.FormatLevel(report.TotalPoints), report.TotalPoints)
		response += fmt.Sprintf("📅 Total hari aktif: %d\n", report.ActivityCount)

		nextComebackAch := uc.getNextComebackTarget(report)
		if nextComebackAch != nil {
			response += fmt.Sprintf("\n🔥 Comeback Challenge dimulai! Raih %d minggu berturut-turut untuk unlock \"%s\"!", nextComebackAch.MinComebackStreak, nextComebackAch.Name)
		} else {
			response += "\n🔥 Comeback Challenge dimulai! Ayo bangun streak-mu kembali!"
		}
	} else {
		cyclePrefix := ""
		if report.CenturionCycles > 0 {
			cyclePrefix = fmt.Sprintf("[C%d] ", report.CenturionCycles+1)
		}

		expBreakdown := fmt.Sprintf("⭐ +%d pts", basePoints)
		if streakBonus > 0 {
			expBreakdown += fmt.Sprintf(" (streak bonus +%d)", streakBonus)
		}
		if seasonalFirstBonus > 0 {
			expBreakdown += " (first season report +2)"
		}
		expBreakdown += " (lapor kemarin, ½ XP)"

		response = fmt.Sprintf("Laporan kemarin diterima, %s%s sudah berkeringat %d hari. Lanjutkan 🔥 (streak %d minggu)\n%s",
			cyclePrefix, name, report.ActivityCount, report.Streak, expBreakdown)
		if report.JobClass != "" {
			response += fmt.Sprintf("\n🧭 Job: %s", domain.FormatJobClass(report.JobClass))
		}
	}

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

	if workout != nil {
		response += "\n" + uc.formatWorkout(workout)
	}

	if streakFreezeUsed {
		response += fmt.Sprintf("\n\n❄️ *Streak Freeze terpakai!* Streak kamu aman. (Sisa freeze: %d)", report.StreakFreezes)
	}

	if newRecord {
		response += fmt.Sprintf("\n\n🏆 New Personal Best Streak: %d minggu!", report.MaxStreak)
	}

	if goalCompleted {
		response += "\n\n🎯 *Goal minggu ini tercapai!* Konsistensi harianmu sudah sesuai target. Mantap, champ! 🏆"
	}

	if leveledUp {
		response += fmt.Sprintf("\n\n⚔️ *LEVEL UP!* Lv.%d → Lv.%d", oldNumericLevel, report.Level)
	}

	if len(newAchievements)+len(comebackAchievements) > 0 {
		response += "\n\n🏅 *Badge baru:*"
		for _, ach := range newAchievements {
			response += fmt.Sprintf("\n%s %s (+%d pts)", ach.DisplayEmoji, ach.Name, ach.Points)
		}
		for _, ach := range comebackAchievements {
			response += fmt.Sprintf("\n%s %s (+%d pts)", ach.DisplayEmoji, ach.Name, ach.Points)
		}

		if freezeAwarded {
			response += fmt.Sprintf("\n\n❄️ Bonus: +1 Streak Freeze! (Total: %d)", report.StreakFreezes)
		}

		response += "\nDetail badge & cerita unlock: #achievements"
	}

	response += fmt.Sprintf("\n\n💬 _\"%s\"_", RandomQuote())

	totalPointsGained := reportPoints + pointsGained
	if totalPointsGained > 0 {
		response += fmt.Sprintf("\n\n💰 Total: +%d points (Lifetime: %d | Season: %d)", totalPointsGained, report.TotalPoints, report.SeasonalPoints)
	}

	response += fmt.Sprintf("\n%s", domain.FormatNumericLevelProgressBar(report.TotalPoints))
	response += fmt.Sprintf("\n%s", domain.FormatProgressBar(report.TotalPoints))

	return response, nil
}

func reportName(storedName, incomingName string) string {
	storedName = strings.TrimSpace(storedName)
	incomingName = strings.TrimSpace(incomingName)

	if !isUnknownName(storedName) {
		return storedName
	}
	if !isUnknownName(incomingName) {
		return incomingName
	}
	return "Teman"
}

func isUnknownName(name string) bool {
	clean := strings.TrimSpace(name)
	return clean == "" ||
		clean == "-" ||
		strings.EqualFold(clean, "unknown") ||
		strings.EqualFold(clean, "teman") ||
		strings.EqualFold(clean, "user")
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
