package usecase

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type GetWeeklyLeaderboardUsecase struct {
	repo domain.ReportRepository
	now  func() time.Time
}

func NewGetWeeklyLeaderboardUsecase(repo domain.ReportRepository) *GetWeeklyLeaderboardUsecase {
	return &GetWeeklyLeaderboardUsecase{
		repo: repo,
		now:  time.Now,
	}
}

// weeklyEntry enriches the raw activity count with streak and point data
// so the leaderboard can apply the full tie-break chain:
// laporan duluan (activity count) → streak → poin.
type weeklyEntry struct {
	Name          string
	ActivityCount int
	Streak        int
	SeasonalPoints int
}

func (uc *GetWeeklyLeaderboardUsecase) Execute(ctx context.Context) (string, error) {
	now := uc.now()
	weekStart := domain.GetStartOfISOWeekStrict(now)
	weekEnd := weekStart.AddDate(0, 0, 7)

	entries, err := uc.repo.GetActivityCountsByDateRange(ctx, weekStart, weekEnd)
	if err != nil {
		return "", err
	}

	reports, err := uc.repo.GetAllReports(ctx)
	if err != nil {
		return "", err
	}

	enriched := enrichWeeklyEntries(entries, reports)

	sort.Slice(enriched, func(i, j int) bool {
		if enriched[i].ActivityCount != enriched[j].ActivityCount {
			return enriched[i].ActivityCount > enriched[j].ActivityCount
		}
		if enriched[i].Streak != enriched[j].Streak {
			return enriched[i].Streak > enriched[j].Streak
		}
		if enriched[i].SeasonalPoints != enriched[j].SeasonalPoints {
			return enriched[i].SeasonalPoints > enriched[j].SeasonalPoints
		}
		return enriched[i].Name < enriched[j].Name
	})

	sb := strings.Builder{}
	sb.WriteString("Leaderboard mingguan\n")
	sb.WriteString(fmt.Sprintf("Periode: %s s/d sebelum %s\n", weekStart.Format("02-01-2006"), weekEnd.Format("02-01-2006")))
	sb.WriteString("Data terhitung sampai saat ini.\n\n")
	sb.WriteString("Berikut posisi sementara mingguan:\n")

	if len(enriched) == 0 {
		sb.WriteString("Belum ada yang masuk leaderboard minggu ini.\n")
		sb.WriteString("Jangan lupa bergerak minggu ini ya 💪")
		return sb.String(), nil
	}

	for rank, entry := range enriched {
		sb.WriteString(fmt.Sprintf("%d. %s: %d hari\n", rank+1, entry.Name, entry.ActivityCount))
	}

	sb.WriteString("\nKalau belum masuk leaderboard, jangan lupa bergerak minggu ini ya 💪")

	return sb.String(), nil
}

func enrichWeeklyEntries(entries []domain.ActivityLeaderboardEntry, reports []*domain.Report) []weeklyEntry {
	reportByUserID := make(map[string]*domain.Report, len(reports))
	for _, r := range reports {
		reportByUserID[r.UserID] = r
	}

	result := make([]weeklyEntry, 0, len(entries))
	for _, e := range entries {
		entry := weeklyEntry{
			Name:          e.Name,
			ActivityCount: e.ActivityCount,
		}
		if r, ok := reportByUserID[e.UserID]; ok {
			entry.Streak = r.Streak
			entry.SeasonalPoints = r.SeasonalPoints
		}
		result = append(result, entry)
	}
	return result
}
