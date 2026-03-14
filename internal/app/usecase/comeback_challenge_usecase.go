package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type ComebackChallengeUsecase struct {
	repo domain.ReportRepository
}

func NewComebackChallengeUsecase(repo domain.ReportRepository) *ComebackChallengeUsecase {
	return &ComebackChallengeUsecase{repo: repo}
}

func (uc *ComebackChallengeUsecase) Execute(ctx context.Context, userID, name string) (string, error) {
	report, err := uc.repo.GetReport(ctx, userID)
	if err != nil {
		return "", err
	}

	if report == nil {
		return fmt.Sprintf("Halo %s, kamu belum pernah laporan aktivitas. Yuk mulai dengan ketik #lapor!", name), nil
	}

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("🔄 Comeback Challenge — %s\n\n", report.Name))

	if report.InactiveDays == 0 {
		sb.WriteString("Kamu belum pernah absen lama. Pertahankan streak-mu! 🔥\n\n")
		sb.WriteString(fmt.Sprintf("🔥 Streak saat ini: %d hari\n", report.Streak))
		sb.WriteString(fmt.Sprintf("📅 Total hari aktif: %d\n", report.ActivityCount))
		sb.WriteString(fmt.Sprintf("🎖️ Level: %s\n", domain.FormatLevel(report.TotalPoints)))
		return sb.String(), nil
	}

	// Show comeback status
	sb.WriteString(fmt.Sprintf("Kamu kembali setelah %d hari absen.\n", report.InactiveDays))
	sb.WriteString(fmt.Sprintf("Comeback streak saat ini: %d hari 🔥\n\n", report.ComebackStreak))

	// Show comeback achievements status
	sb.WriteString("🏅 Comeback Achievements:\n")
	for _, a := range domain.AllComebackAchievements {
		if domain.HasAchievement(report.Achievements, a.ID) {
			sb.WriteString(fmt.Sprintf("  ✅ %s — %s (%d pts)\n", a.Name, a.Description, a.Points))
		} else if report.InactiveDays >= a.MinInactiveDays {
			remaining := a.MinComebackStreak - report.ComebackStreak
			if remaining > 0 {
				sb.WriteString(fmt.Sprintf("  🎯 %s — %d hari lagi! (%d pts)\n", a.Name, remaining, a.Points))
			} else {
				// Should have been unlocked already, but just in case
				sb.WriteString(fmt.Sprintf("  🏅 %s — READY TO UNLOCK! (%d pts)\n", a.Name, a.Points))
			}
		} else {
			sb.WriteString(fmt.Sprintf("  🔒 %s — butuh absen >%d hari (%d pts)\n", a.Name, a.MinInactiveDays, a.Points))
		}
	}

	sb.WriteString(fmt.Sprintf("\n📊 Level: %s\n", domain.FormatLevel(report.TotalPoints)))
	sb.WriteString(fmt.Sprintf("📈 %s\n", domain.FormatProgressBar(report.TotalPoints)))

	// Motivational closer
	sb.WriteString("\n💪 Terus semangat! Setiap hari laporan membawamu lebih dekat ke achievement baru!")

	return sb.String(), nil
}
