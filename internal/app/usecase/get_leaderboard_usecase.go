package usecase

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type GetLeaderboardUsecase struct {
	repo domain.ReportRepository
}

func NewGetLeaderboardUsecase(repo domain.ReportRepository) *GetLeaderboardUsecase {
	return &GetLeaderboardUsecase{repo: repo}
}

func (uc *GetLeaderboardUsecase) Execute(ctx context.Context) (string, error) {
	reports, err := uc.repo.GetAllReports(ctx)
	if err != nil {
		return "", err
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// Global Challenge Day Calculation (Optional: Fix a start date or assume max streak represents it?
	// The prompt says "Day 37 (06-02-2026)".
	// Let's use the current Max Streak or a fixed start date if provided.
	// For now, let's look at the highest streak in the DB to infer "Day X" or just use the current highest streak as the "Day".

	// Logic for "Keep the streak" vs "Lose the streak":
	// Keep streak: Reported Today OR Reported Yesterday (still have time to report today).
	// Lose streak: Last report < Yesterday.
	// New submission: Reported Today AND Streak == 1 (and maybe created today?).

	var keepStreak []*domain.Report
	var loseStreak []*domain.Report
	// New submission logic is tricky without a "CreatedDate". We'll skip specific "New submission" category for now or map it to "Keep Streak" with streak=1.
	// Let's simplify:
	// Active (Fire): Reported Today or Yesterday.
	// Lost (Broken Heart): Reported before Yesterday.

	yesterday := today.AddDate(0, 0, -1)

	for _, r := range reports {
		lastReportDate := time.Date(r.LastReportDate.Year(), r.LastReportDate.Month(), r.LastReportDate.Day(), 0, 0, 0, 0, r.LastReportDate.Location())

		if lastReportDate.Equal(today) || lastReportDate.Equal(yesterday) {
			keepStreak = append(keepStreak, r)
		} else {
			loseStreak = append(loseStreak, r)
		}
	}

	// Sort by Streak Descending
	sort.Slice(keepStreak, func(i, j int) bool {
		return keepStreak[i].Streak > keepStreak[j].Streak
	})
	sort.Slice(loseStreak, func(i, j int) bool {
		return loseStreak[i].Streak > loseStreak[j].Streak
	})

	// Header
	// "Day 37" - Let's approximate this by the max streak found or just "Leaderboard"
	maxStreak := 0
	if len(reports) > 0 {
		// Find max streak overall
		for _, r := range reports {
			if r.Streak > maxStreak {
				maxStreak = r.Streak
			}
		}
	}

	sb := strings.Builder{}
	dateStr := now.Format("02-01-2006")
	sb.WriteString(fmt.Sprintf("30 Days of Sweat Challenge – Day %d (%s)\n\n", maxStreak, dateStr))

	// Recap
	// "22 peoples keep the streak 🔥"
	sb.WriteString(fmt.Sprintf("Recap day %d:\n", maxStreak))
	sb.WriteString(fmt.Sprintf("%d peoples keep the streak 🔥\n", len(keepStreak)))
	sb.WriteString(fmt.Sprintf("%d lose the streak 💔\n", len(loseStreak)))
	// New submission/Left behind omitted for simplicity unless we add more tracking
	sb.WriteString("\nUpdate klasemen sementara:\n")

	rank := 1
	// Active
	for _, r := range keepStreak {
		sb.WriteString(fmt.Sprintf("%d. %s - %d days streak 🔥\n", rank, r.Name, r.Streak))
		rank++
	}

	// Lost
	for _, r := range loseStreak {
		sb.WriteString(fmt.Sprintf("%d. @%s - Day %d 💔\n", rank, r.Name, r.Streak))
		rank++
	}

	sb.WriteString("\nYang udah keringetan langsung update/posting aja nanti dimasukkin klasemen 💪\n\nSemangat🔥")

	return sb.String(), nil
}
