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

func (uc *GetWeeklyLeaderboardUsecase) Execute(ctx context.Context) (string, error) {
	now := uc.now()
	weekStart := domain.GetStartOfSundayWeek(now)
	weekEnd := weekStart.AddDate(0, 0, 7)

	entries, err := uc.repo.GetActivityCountsByDateRange(ctx, weekStart, weekEnd)
	if err != nil {
		return "", err
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].ActivityCount == entries[j].ActivityCount {
			return entries[i].Name < entries[j].Name
		}
		return entries[i].ActivityCount > entries[j].ActivityCount
	})

	sb := strings.Builder{}
	sb.WriteString("Leaderboard mingguan\n")
	sb.WriteString(fmt.Sprintf("Periode: %s s/d sebelum %s\n", weekStart.Format("02-01-2006"), weekEnd.Format("02-01-2006")))
	sb.WriteString("Data terhitung sampai saat ini.\n\n")
	sb.WriteString("Berikut posisi sementara mingguan:\n")

	if len(entries) == 0 {
		sb.WriteString("Belum ada yang masuk leaderboard minggu ini.\n")
		sb.WriteString("Jangan lupa bergerak minggu ini ya 💪")
		return sb.String(), nil
	}

	for rank, entry := range entries {
		sb.WriteString(fmt.Sprintf("%d. %s: %d hari\n", rank+1, entry.Name, entry.ActivityCount))
	}

	sb.WriteString("\nKalau belum masuk leaderboard, jangan lupa bergerak minggu ini ya 💪")

	return sb.String(), nil
}
