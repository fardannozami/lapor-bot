package usecase

import (
	"context"
	"fmt"
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
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if report != nil {
		lastReport := report.LastReportDate
		lastReportDate := time.Date(lastReport.Year(), lastReport.Month(), lastReport.Day(), 0, 0, 0, 0, lastReport.Location())

		if lastReportDate.Equal(today) {
			return fmt.Sprintf("%s sudah laporan hari ini, ayo jangan curang! 😉", name), nil
		}

		// Calculate streak (simplified: if last report was yesterday, increment. Else reset?
		// Requirement says "36 days streak", "Day 31 💔".
		// 💔 implies they broke the streak but we still track "Day X".
		// The prompt says "Recap day 37... 45 lose the streak 💔". This implies if they missed a day, they lose streak status but maybe the day count is preserved or it's just a display thing?
		// Let's assume for #lapor:
		// if last report was yesterday -> streak++
		// if last report was older -> reset streak to 1? or just increment count?
		// The prompt example "Laporan diterima, {wa name} sudah berkeringat {counting day} hari." suggests a cumulative count or current streak.
		// Let's implement: If reported yesterday -> Streak++. Else -> Streak = 1 (new streak).

		// Wait, the prompt says "Day 31 💔". This implies a Challenge context where "Day X" is the global challenge day, and "Streak" is personal.
		// BUT, the specific response for #lapor is: "sudah berkeringat {counting day} hari."
		// And the leaderboard splits into "Streak 🔥" and "Day X 💔".
		// This implies we track the Streak. If they report today, we update the streak.

		// Let's implement robust streak logic:
		// If last report was yesterday (today - 1 day), streak++.
		// If last report was today (already handled above).
		// If last report was before yesterday, streak = 1.

		yesterday := today.AddDate(0, 0, -1)
		if lastReportDate.Equal(yesterday) {
			report.Streak++
		} else {
			// Logic for broken streak.
			// If they missed a day, they restart?
			// "45 lose the streak 💔".
			// The leaderboard shows "Day 31 💔" for someone who lost it. identifying they are on Day 31 of the challenge but lost streak?
			// Or maybe "Day 31" is their last max streak?
			// Let's stick to standard streak logic: Reset to 1 if missed.
			// BUT, if we reset to 1, they become "1 days streak 🔥". They don't appear as "Day 31 💔".
			// "Day 31 💔" likely means they *were* on Day 31 but missed today/yesterday.

			// For the specific "#lapor" response: "{counting day} hari" likely refers to the NEW streak count.
			report.Streak = 1
		}
		report.Name = name // Update name if changed
		report.LastReportDate = now
	} else {
		report = &domain.Report{
			UserID:         userID,
			Name:           name,
			Streak:         1,
			LastReportDate: now,
		}
	}

	if err := uc.repo.UpsertReport(ctx, report); err != nil {
		return "", err
	}

	return fmt.Sprintf("Laporan diterima, %s sudah berkeringat %d hari. Lanjutkan 🔥", name, report.Streak), nil
}
