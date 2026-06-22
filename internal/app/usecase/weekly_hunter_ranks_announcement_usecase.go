package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type WeeklyHunterRanksAnnouncementUsecase struct {
	repo domain.ReportRepository
}

func NewWeeklyHunterRanksAnnouncementUsecase(repo domain.ReportRepository) *WeeklyHunterRanksAnnouncementUsecase {
	return &WeeklyHunterRanksAnnouncementUsecase{repo: repo}
}

func (u *WeeklyHunterRanksAnnouncementUsecase) Execute(ctx context.Context, now time.Time) (string, error) {
	reports, err := u.repo.GetAllReports(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get reports: %w", err)
	}

	seasonNumber, _ := GetCurrentSessionInfo(now)
	nextReset := GetNextResetTime(now)

	domain.SortReports(reports, domain.SortBySeasonRank)

	var sb strings.Builder
	sb.WriteString("📢 *PENGUMUMAN MINGGUAN: RANK & JOB HUNTER* 🏹\n")
	sb.WriteString(fmt.Sprintf("Season %d | Reset badge/rank: %s\n", seasonNumber, nextReset.Format("02-01-2006")))
	sb.WriteString("Level & EXP lifetime tetap aman.\n\n")

	rank := 1
	for _, r := range reports {
		if !domain.HasSeasonActivity(r) {
			continue
		}

		jobStr := domain.FormatJobClass(r.JobClass)

		sb.WriteString(fmt.Sprintf(
			"%d. %s — %s (%s) | %d pts | %d hari\n",
			rank,
			r.Name,
			domain.FormatSeasonRank(r.SeasonalPoints),
			jobStr,
			r.SeasonalPoints,
			r.SeasonalActivityCount,
		))

		rank++
	}

	if rank == 1 {
		sb.WriteString("Belum ada hunter aktif season ini. Mulai dengan /lapor 💪")
		return sb.String(), nil
	}

	sb.WriteString("\nTetap konsisten berlatih untuk menaikkan rank-mu! Semangat🔥")
	sb.WriteString("\n\n🌐 Lihat klasemen & stats: https://lapor-bot.web.id/")

	return sb.String(), nil
}
