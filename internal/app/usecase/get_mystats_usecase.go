package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type GetMyStatsUsecase struct {
	repo domain.ReportRepository
}

func NewGetMyStatsUsecase(repo domain.ReportRepository) *GetMyStatsUsecase {
	return &GetMyStatsUsecase{repo: repo}
}

func (uc *GetMyStatsUsecase) Execute(ctx context.Context, userID, name string) (string, error) {
	report, err := uc.repo.GetReport(ctx, userID)
	if err != nil {
		return "", err
	}

	if report == nil {
		return fmt.Sprintf("Halo %s, kamu belum pernah laporan aktivitas. Yuk mulai dengan ketik #lapor!", name), nil
	}

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("📊 Statistik kamu, %s:\n\n", report.Name))
	sb.WriteString(fmt.Sprintf("🔥 Streak saat ini: %d hari\n", report.Streak))
	sb.WriteString(fmt.Sprintf("🏆 Streak tertinggi: %d hari\n", report.MaxStreak))
	sb.WriteString(fmt.Sprintf("📅 Total hari aktif: %d\n", report.ActivityCount))
	sb.WriteString(fmt.Sprintf("⭐ Total poin: %d\n", report.TotalPoints))

	// Helper to get achievement name by ID
	getAchName := func(id string) string {
		for _, a := range domain.AllAchievements {
			if a.ID == id {
				return a.Name
			}
		}
		return id
	}

	if report.Achievements != "" {
		sb.WriteString("\n🏅 Achievements:\n")
		ids := strings.Split(report.Achievements, ",")
		for _, id := range ids {
			trimmedID := strings.TrimSpace(id)
			if trimmedID != "" {
				sb.WriteString(fmt.Sprintf("✅ %s\n", getAchName(trimmedID)))
			}
		}
	} else {
		sb.WriteString("\n🏅 Belum ada achievement. Terus semangat! 💪")
	}

	return sb.String(), nil
}
