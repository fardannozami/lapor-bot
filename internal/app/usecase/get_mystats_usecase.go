package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type GetMyStatsUsecase struct {
	repo domain.ReportRepository
	now  func() time.Time
}

func NewGetMyStatsUsecase(repo domain.ReportRepository) *GetMyStatsUsecase {
	return &GetMyStatsUsecase{
		repo: repo,
		now:  time.Now,
	}
}

func (uc *GetMyStatsUsecase) Execute(ctx context.Context, userID, name string) (string, error) {
	report, err := uc.repo.GetReport(ctx, userID)
	if err != nil {
		return "", err
	}

	if report == nil {
		return fmt.Sprintf("Halo %s, kamu belum pernah laporan aktivitas. Yuk mulai dengan ketik #lapor!", name), nil
	}

	now := uc.now()
	weekStart := domain.GetStartOfSundayWeek(now)
	weekEnd := weekStart.AddDate(0, 0, 7)
	weeklyEntries, err := uc.repo.GetActivityCountsByDateRange(ctx, weekStart, weekEnd)
	if err != nil {
		return "", err
	}

	_, sessionStart := GetCurrentSessionInfo(now)
	seasonStart := time.Date(sessionStart.Year(), sessionStart.Month(), sessionStart.Day(), 0, 0, 0, 0, time.UTC)
	seasonEnd := GetNextResetTime(now)
	seasonEntries, err := uc.repo.GetActivityCountsByDateRange(ctx, seasonStart, seasonEnd)
	if err != nil {
		return "", err
	}

	getActivityCount := func(entries []domain.ActivityLeaderboardEntry) int {
		for _, entry := range entries {
			if entry.UserID == userID {
				return entry.ActivityCount
			}
		}
		return 0
	}

	weeklyCount := getActivityCount(weeklyEntries)
	seasonCount := getActivityCount(seasonEntries)

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("📊 Statistik kamu, %s:\n\n", report.Name))

	// Level display
	sb.WriteString(fmt.Sprintf("🎖️ Level: %s\n", domain.FormatLevel(report.TotalPoints)))
	sb.WriteString(fmt.Sprintf("📈 %s\n\n", domain.FormatProgressBar(report.TotalPoints)))

	sb.WriteString(fmt.Sprintf("🔥 Streak saat ini: %d minggu\n", report.Streak))
	sb.WriteString(fmt.Sprintf("🏆 Streak tertinggi: %d minggu\n", report.MaxStreak))
	streakFreezeInfo := fmt.Sprintf("❄️ Streak Freeze: %d", report.StreakFreezes)
	if report.StreakFreezes == 0 {
		streakFreezeInfo += " (capai 4 minggu streak untuk dapat +1 freeze!)"
	} else if report.StreakFreezes == 1 {
		streakFreezeInfo += " — otomatis melindungi 1 minggu absen"
	} else {
		streakFreezeInfo += " (max!)"
	}
	sb.WriteString(streakFreezeInfo + "\n")
	sb.WriteString(fmt.Sprintf("📅 Total hari aktif (lifetime): %d\n", report.ActivityCount))
	sb.WriteString(fmt.Sprintf("🗓️ Total hari season: %d\n", seasonCount))
	sb.WriteString(fmt.Sprintf("📆 Total mingguan: %d\n", weeklyCount))
	sb.WriteString(fmt.Sprintf("⭐ Total poin (lifetime): %d\n", report.TotalPoints))
	sb.WriteString(fmt.Sprintf("🌟 Seasonal poin: %d\n", report.SeasonalPoints))
	sb.WriteString(fmt.Sprintf("📊 Seasonal activity: %d\n", report.SeasonalActivityCount))

	// Comeback status
	if report.InactiveDays > 0 && report.ComebackStreak > 0 && report.ComebackStreak < 30 {
		sb.WriteString(fmt.Sprintf("\n🔄 Comeback streak: %d minggu (setelah %d hari absen)\n", report.ComebackStreak, report.InactiveDays))

		// Show next comeback achievement target
		for _, a := range domain.AllComebackAchievements {
			if !domain.HasAchievement(report.Achievements, a.ID) &&
				report.InactiveDays >= a.MinInactiveDays {
				remaining := a.MinComebackStreak - report.ComebackStreak
				if remaining > 0 {
					sb.WriteString(fmt.Sprintf("🎯 %d minggu lagi untuk unlock \"%s\"!\n", remaining, a.Name))
				}
				break
			}
		}
	}

	sb.WriteString("\n🏅 *Achievements & Progress*\n")

	// Show all standard achievements with status
	for _, a := range domain.AllAchievements {
		if domain.HasAchievement(report.Achievements, a.ID) {
			sb.WriteString(fmt.Sprintf("%s %s — %s (%d pts)\n", a.DisplayEmoji, a.Name, a.Description, a.Points))
		} else {
			sb.WriteString(fmt.Sprintf("🔒 %s — %s (%d pts)\n", a.Name, a.Description, a.Points))
		}
	}

	// Show comeback achievements with progress
	sb.WriteString("\n🔄 *Comeback Achievements*\n")
	sb.WriteString("_Khusus untuk yang kembali setelah absen!_\n")
	for _, a := range domain.AllComebackAchievements {
		if domain.HasAchievement(report.Achievements, a.ID) {
			sb.WriteString(fmt.Sprintf("%s %s — %s (%d pts)\n", a.DisplayEmoji, a.Name, a.Description, a.Points))
		} else if report.InactiveDays >= a.MinInactiveDays {
			remaining := a.MinComebackStreak - report.ComebackStreak
			if remaining > 0 {
				sb.WriteString(fmt.Sprintf("🎯 %s — %d minggu lagi (%d pts)\n", a.Name, remaining, a.Points))
			} else {
				sb.WriteString(fmt.Sprintf("%s %s — Siap unlock! (%d pts)\n", a.DisplayEmoji, a.Name, a.Points))
			}
		} else {
			sb.WriteString(fmt.Sprintf("🔒 %s — butuh absen >%d hari dulu (%d pts)\n", a.Name, a.MinInactiveDays, a.Points))
		}
	}

	return sb.String(), nil
}
