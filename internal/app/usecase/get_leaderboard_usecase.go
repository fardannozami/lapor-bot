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

	yesterday := today.AddDate(0, 0, -1)

	// Sort all reports by ActivityCount (total days) descending
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].ActivityCount > reports[j].ActivityCount
	})

	// Count active vs lost for recap
	activeCount := 0
	lostCount := 0
	for _, r := range reports {
		lastReportDate := time.Date(r.LastReportDate.Year(), r.LastReportDate.Month(), r.LastReportDate.Day(), 0, 0, 0, 0, r.LastReportDate.Location())
		if lastReportDate.Equal(today) || lastReportDate.Equal(yesterday) {
			activeCount++
		} else {
			lostCount++
		}
	}

	// Header
	// Use max activity count to represent the current "Day" of the challenge
	maxDay := 0
	if len(reports) > 0 {
		for _, r := range reports {
			if r.ActivityCount > maxDay {
				maxDay = r.ActivityCount
			}
		}
	}

	sb := strings.Builder{}
	dateStr := now.Format("02-01-2006")
	sb.WriteString(fmt.Sprintf("30 Days of Sweat Challenge – Day %d (%s)\n\n", maxDay, dateStr))

	// Recap
	sb.WriteString(fmt.Sprintf("Recap day %d:\n", maxDay))
	sb.WriteString(fmt.Sprintf("%d peoples keep the streak 🔥\n", activeCount))
	sb.WriteString(fmt.Sprintf("%d lose the streak 💔\n", lostCount))
	sb.WriteString("\nUpdate klasemen sementara:\n")

	// Single unified ranking by ActivityCount
	for rank, r := range reports {
		lastReportDate := time.Date(r.LastReportDate.Year(), r.LastReportDate.Month(), r.LastReportDate.Day(), 0, 0, 0, 0, r.LastReportDate.Location())
		isActive := lastReportDate.Equal(today) || lastReportDate.Equal(yesterday)
		
		if isActive {
			sb.WriteString(fmt.Sprintf("%d. %s - %d days 🔥\n", rank+1, r.Name, r.ActivityCount))
		} else {
			sb.WriteString(fmt.Sprintf("%d. %s - %d days 💔\n", rank+1, r.Name, r.ActivityCount))
		}
	}

	sb.WriteString("\nYang udah keringetan langsung update/posting aja nanti dimasukkin klasemen 💪\n\nSemangat🔥")

	return sb.String(), nil
}
