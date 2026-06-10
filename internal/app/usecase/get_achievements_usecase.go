package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

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

	seasonNumber, _ := GetCurrentSessionInfo(time.Now())

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("🎖️ *Season Badge Challenge — Season %d*\n", seasonNumber))
	sb.WriteString("Badge di bawah ini reset setiap season. Level & EXP lifetime tetap aman.\n")
	sb.WriteString("Notifikasi #lapor hanya menampilkan badge terbaru; detail lengkapnya ada di sini.\n\n")

	for _, ach := range domain.AllSeasonAchievements {
		count := stats[ach.ID]
		var icon string
		if count > 0 {
			icon = ach.DisplayEmoji
		} else {
			icon = "🔒"
		}

		sb.WriteString(fmt.Sprintf("%s *%s* (+%d pts)\n", icon, ach.Name, ach.Points))
		sb.WriteString(fmt.Sprintf("   Syarat: %s (%d/%d member)\n", ach.Description, count, totalMembers))
		if ach.UnlockMessage != "" {
			sb.WriteString(fmt.Sprintf("   _%s_\n", ach.UnlockMessage))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("🔄 *Comeback Badges*\n")
	for _, ach := range domain.AllComebackAchievements {
		sb.WriteString(fmt.Sprintf("%s *%s* (+%d pts)\n", ach.DisplayEmoji, ach.Name, ach.Points))
		sb.WriteString(fmt.Sprintf("   Syarat: absen minimal %d hari, lalu comeback streak %d minggu.\n", ach.MinInactiveDays, ach.MinComebackStreak))
		if ach.UnlockMessage != "" {
			sb.WriteString(fmt.Sprintf("   _%s_\n", ach.UnlockMessage))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("⚔️ Level numerik naik dari total EXP lifetime. Semakin tinggi level, semakin banyak EXP yang dibutuhkan.")

	return sb.String(), nil
}
