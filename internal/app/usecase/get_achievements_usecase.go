package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type GetAchievementsUsecase struct {
	repo domain.ReportRepository
}

func NewGetAchievementsUsecase(repo domain.ReportRepository) *GetAchievementsUsecase {
	return &GetAchievementsUsecase{repo: repo}
}

func (uc *GetAchievementsUsecase) Execute(ctx context.Context) (string, error) {
	reports, err := uc.repo.GetAllReports(ctx)
	if err != nil {
		return "", err
	}

	totalMembers := len(reports)
	if totalMembers == 0 {
		return "Belum ada member yang aktif.", nil
	}

	// Calculate season badge stats. Lifetime achievements are preserved in the
	// database, but the visible badge race resets every season.
	stats := make(map[string]int)
	for _, r := range reports {
		if r.SeasonalAchievements != "" {
			ids := strings.Split(r.SeasonalAchievements, ",")
			for _, id := range ids {
				stats[strings.TrimSpace(id)]++
			}
		}
	}

	sb := strings.Builder{}
	sb.WriteString("🎖️ *Season Badge Challenge*\n")
	sb.WriteString("Badge di bawah ini reset setiap season. Level & EXP lifetime tetap aman.\n\n")

	for _, ach := range domain.AllSeasonAchievements {
		count := stats[ach.ID]
		var icon string
		if count > 0 {
			icon = ach.DisplayEmoji
		} else {
			icon = "🔒"
		}

		sb.WriteString(fmt.Sprintf("%s %s — %s (%d/%d member)\n", icon, ach.Name, ach.Description, count, totalMembers))
	}

	return sb.String(), nil
}
