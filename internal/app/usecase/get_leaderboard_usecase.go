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
	displayDate := domain.GetToday(now)
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	challengeDay := int(displayDate.Sub(startDate).Hours()/24) + 1

	// Logic for "Keep the streak" vs "Lose the streak":
	// Keep streak: Reported Today OR Reported Yesterday (still have time to report today).
	// Lose streak: Last report < Yesterday.
	// New submission: Reported Today AND Streak == 1 (and maybe created today?).

	// Sort all reports by ActivityCount (total days) descending
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].ActivityCount > reports[j].ActivityCount
	})

	// Count active vs lost for recap
	activeCount := 0
	lostCount := 0
	currentWeekStart := domain.GetStartOfISOWeek(displayDate)
	for _, r := range reports {
		lastWeekStart := domain.GetStartOfISOWeek(r.LastReportDate)
		weeksSinceLastReport := int(math.Round(currentWeekStart.Sub(lastWeekStart).Hours() / (24 * 7)))

		// Active if reported this week or last week (still has chance to continue streak)
		if weeksSinceLastReport <= 1 {
			activeCount++
		} else {
			lostCount++
		}
	}

	// Header
	sb := strings.Builder{}
	dateStr := displayDate.Format("02-01-2006")
	sb.WriteString(fmt.Sprintf("30 Days of Sweat Challenge – Day %d (%s)\n\n", challengeDay, dateStr))

	// Recap
	sb.WriteString(fmt.Sprintf("Recap day %d:\n", challengeDay))
	sb.WriteString(fmt.Sprintf("%d peoples keep the streak 🔥\n", activeCount))
	sb.WriteString(fmt.Sprintf("%d lose the streak 💔\n", lostCount))
	sb.WriteString("\nUpdate klasemen sementara:\n")

	// Single unified ranking by ActivityCount
	for rank, r := range reports {
		// Active if reported this week or last week
		lastWeekStart := domain.GetStartOfISOWeek(r.LastReportDate)
		weeksSinceLastReport := int(math.Round(currentWeekStart.Sub(lastWeekStart).Hours() / (24 * 7)))

		if weeksSinceLastReport <= 1 {
			sb.WriteString(fmt.Sprintf("%d. %s - %d days (%d weeks streak 🔥)\n", rank+1, r.Name, r.ActivityCount, r.Streak))
		} else {
			sb.WriteString(fmt.Sprintf("%d. %s - %d days (💔)\n", rank+1, r.Name, r.ActivityCount))
		}
	}

	sb.WriteString("\nYang udah keringetan langsung update/posting aja nanti dimasukkin klasemen 💪\n\nSemangat🔥")

	return sb.String(), nil
}
