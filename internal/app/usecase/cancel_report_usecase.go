package usecase

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type CancelReportUsecase struct {
	repo domain.ReportRepository
}

func NewCancelReportUsecase(repo domain.ReportRepository) *CancelReportUsecase {
	return &CancelReportUsecase{repo: repo}
}

func (uc *CancelReportUsecase) Execute(ctx context.Context, userID, name string) (string, error) {
	return uc.cancelToday(ctx, userID, name, false)
}

func (uc *CancelReportUsecase) ExecuteAll(ctx context.Context, userID, name string) (string, error) {
	return uc.cancelToday(ctx, userID, name, true)
}

func (uc *CancelReportUsecase) cancelToday(ctx context.Context, userID, name string, all bool) (string, error) {
	report, err := uc.repo.GetReport(ctx, userID)
	if err != nil {
		return "", err
	}

	if report == nil {
		return fmt.Sprintf("Halo %s, kamu belum pernah laporan. Belum ada yang bisa dibatalkan.", name), nil
	}

	now := time.Now()
	today := domain.GetToday(now)
	lastReportDate := domain.GetToday(report.LastReportDate)

	if !lastReportDate.Equal(today) {
		return fmt.Sprintf("Halo %s, kamu belum laporan hari ini. Tidak ada yang perlu dibatalkan.", report.Name), nil
	}

	dates, err := uc.repo.GetUserActivityDates(ctx, userID)
	if err != nil {
		return "", err
	}

	if len(dates) == 0 {
		return fmt.Sprintf("Halo %s, tidak ada aktivitas yang tercatat.", report.Name), nil
	}

	dailyCount, err := uc.repo.GetDailyActivityCount(ctx, userID, today)
	if err != nil {
		return "", err
	}
	if dailyCount == 0 {
		return fmt.Sprintf("Halo %s, tidak menemukan laporan untuk hari ini.", report.Name), nil
	}

	if !all && dailyCount > 1 {
		remainingReports, err := uc.repo.DeleteLatestActivityLog(ctx, userID, today)
		if err != nil {
			return "", err
		}
		if remainingReports == 0 {
			return uc.cancelToday(ctx, userID, name, true)
		}

		removeRepeatReportPoints(report)
		if err := uc.repo.UpsertReport(ctx, report); err != nil {
			return "", err
		}

		msg := fmt.Sprintf("✅ Laporan terakhir hari ini telah dibatalkan, %s.\n\n", report.Name)
		msg += fmt.Sprintf("📌 Sisa laporan hari ini: %d/%d\n", remainingReports, MaxDailyReports)
		msg += fmt.Sprintf("📅 Total hari aktif tetap: %d\n", report.ActivityCount)
		msg += fmt.Sprintf("⭐ Total poin sekarang: %d\n", report.TotalPoints)
		msg += "\nKalau ingin menghapus semua laporan hari ini, ketik /cancel-all."
		return msg, nil
	}

	remainingDates := removeDate(dates, today)
	if len(remainingDates) == len(dates) {
		return fmt.Sprintf("Halo %s, tidak menemukan laporan untuk hari ini.", report.Name), nil
	}

	if err := uc.repo.DeleteActivityLog(ctx, userID, today); err != nil {
		return "", err
	}

	var newReport *domain.Report
	if len(remainingDates) == 0 {
		newReport = &domain.Report{
			UserID:          userID,
			Name:            report.Name,
			ActivityCount:   0,
			LastReportDate:  time.Time{},
			Streak:          0,
			MaxStreak:       0,
			TotalPoints:     0,
			Achievements:    "",
			ComebackStreak:  0,
			InactiveDays:    0,
			CenturionCycles: 0,
		}
	} else {
		newReport = recalculateReportFromDates(userID, report.Name, remainingDates)
	}

	if err := uc.repo.UpsertReport(ctx, newReport); err != nil {
		return "", err
	}

	msg := fmt.Sprintf("✅ Semua laporan hari ini telah dibatalkan, %s.\n\n", report.Name)
	if !all {
		msg = fmt.Sprintf("✅ Laporan hari ini telah dibatalkan, %s.\n\n", report.Name)
	}
	msg += fmt.Sprintf("📅 Total hari aktif: %d\n", newReport.ActivityCount)
	msg += fmt.Sprintf("🔥 Streak saat ini: %d minggu\n", newReport.Streak)
	msg += fmt.Sprintf("⭐ Total poin: %d\n", newReport.TotalPoints)

	if newReport.InactiveDays > 0 {
		msg += fmt.Sprintf("🔄 Comeback streak: %d minggu\n", newReport.ComebackStreak)
	}

	msg += "\n_Kamu bisa lapor lagi hari ini dengan /lapor._"
	return msg, nil
}

func removeRepeatReportPoints(report *domain.Report) {
	const repeatReportPoints = 5
	report.TotalPoints -= repeatReportPoints
	if report.TotalPoints < 0 {
		report.TotalPoints = 0
	}
	report.SeasonalPoints -= repeatReportPoints
	if report.SeasonalPoints < 0 {
		report.SeasonalPoints = 0
	}
	report.Level = domain.NumericLevelFromTotalPoints(report.TotalPoints)
}

func removeDate(dates []time.Time, target time.Time) []time.Time {
	var result []time.Time
	for _, d := range dates {
		if !d.Equal(target) {
			result = append(result, d)
		}
	}
	return result
}

func recalculateReportFromDates(userID, name string, dates []time.Time) *domain.Report {
	if len(dates) == 0 {
		return nil
	}

	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	totalDates := len(dates)
	lastReportDate := dates[len(dates)-1]

	weeks := make([]time.Time, 0, len(dates))
	seenWeeks := make(map[string]bool)
	for _, d := range dates {
		ws := domain.GetStartOfISOWeek(d)
		key := ws.Format(time.DateOnly)
		if !seenWeeks[key] {
			seenWeeks[key] = true
			weeks = append(weeks, ws)
		}
	}

	streak, maxStreak, comebackStreak, inactiveDays := calculateStreaksFromWeeks(weeks)

	activityCount := totalDates
	centurionCycles := 0
	if totalDates > 100 {
		centurionCycles = (totalDates - 1) / 100
		activityCount = ((totalDates - 1) % 100) + 1
	}

	report := &domain.Report{
		UserID:          userID,
		Name:            name,
		ActivityCount:   activityCount,
		LastReportDate:  lastReportDate,
		Streak:          streak,
		MaxStreak:       maxStreak,
		ComebackStreak:  comebackStreak,
		InactiveDays:    inactiveDays,
		CenturionCycles: centurionCycles,
		TotalPoints:     0,
		Achievements:    "",
	}

	if maxStreak == 0 && streak > 0 {
		report.MaxStreak = streak
	}

	var allAchIDs []string
	for _, ach := range domain.CheckNewAchievements(report) {
		allAchIDs = append(allAchIDs, ach.ID)
		report.TotalPoints += ach.Points
	}
	for _, ach := range domain.CheckComebackAchievements(report) {
		allAchIDs = append(allAchIDs, ach.ID)
		report.TotalPoints += ach.Points
	}
	if len(allAchIDs) > 0 {
		report.Achievements = strings.Join(allAchIDs, ",")
	}

	return report
}

func calculateStreaksFromWeeks(weeks []time.Time) (streak, maxStreak, comebackStreak, inactiveDays int) {
	if len(weeks) == 0 {
		return 0, 0, 0, 0
	}

	streak = 1
	maxStreak = 1
	comebackStreak = 1
	inactiveDays = 0

	var lastGapDays int

	for i := 1; i < len(weeks); i++ {
		prevWeek := weeks[i-1]
		currWeek := weeks[i]
		weeksDiff := int(math.Round(currWeek.Sub(prevWeek).Hours() / (24 * 7)))

		if weeksDiff == 1 {
			streak++
			comebackStreak++
		} else {
			if streak > maxStreak {
				maxStreak = streak
			}
			lastGapDays = int(math.Round(currWeek.Sub(prevWeek.AddDate(0, 0, 7)).Hours() / 24))
			streak = 1
			comebackStreak = 1
		}
	}

	if streak > maxStreak {
		maxStreak = streak
	}

	if len(weeks) >= 2 {
		lastTwoDiff := int(math.Round(weeks[len(weeks)-1].Sub(weeks[len(weeks)-2]).Hours() / (24 * 7)))
		if lastTwoDiff > 1 {
			inactiveDays = int(math.Round(weeks[len(weeks)-1].Sub(weeks[len(weeks)-2].AddDate(0, 0, 7)).Hours() / 24))
		} else if lastGapDays > 0 {
			inactiveDays = lastGapDays
		}
	}

	return streak, maxStreak, comebackStreak, inactiveDays
}
