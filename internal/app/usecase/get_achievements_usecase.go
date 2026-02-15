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

	// Calculate stats for each achievement
	stats := make(map[string]int)
	for _, r := range reports {
		if r.Achievements != "" {
			ids := strings.Split(r.Achievements, ",")
			for _, id := range ids {
				stats[strings.TrimSpace(id)]++
			}
		}
	}

	sb := strings.Builder{}
	sb.WriteString("🎖️ Daftar Achievement Challenge:\n\n")

	for _, ach := range domain.AllAchievements {
		count := stats[ach.ID]
		// Status icon based on unlock count
		var icon string
		if count > 0 {
			icon = "🏅"
		} else {
			icon = "🔒"
		}

		sb.WriteString(fmt.Sprintf("%s %s — %s (%d/%d member)\n", icon, ach.Name, ach.Description, count, totalMembers))
	}

	return sb.String(), nil
}
