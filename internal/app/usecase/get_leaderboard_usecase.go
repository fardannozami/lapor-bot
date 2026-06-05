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

	// Session-aware start date (auto-cycles every 4 months)
	sessionNumber, sessionStart := GetCurrentSessionInfo(now)
	startDate := time.Date(sessionStart.Year(), sessionStart.Month(), sessionStart.Day(), 0, 0, 0, 0, time.UTC)

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
	currentWeekStart := domain.GetStartOfISOWeek(now)
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

	for rank, r := range reports {
		// Active if reported this week or last week
		lastWeekStart := domain.GetStartOfISOWeek(r.LastReportDate)
		weeksSinceLastReport := int(math.Round(currentWeekStart.Sub(lastWeekStart).Hours() / (24 * 7)))

		cyclePrefix := ""
		if r.CenturionCycles > 0 {
			cyclePrefix = fmt.Sprintf("[S1-C%d] ", r.CenturionCycles+1)
		}

		seasonalInfo := ""
		if r.SeasonalPoints > 0 {
			seasonalInfo = fmt.Sprintf(" | Season %d: %d pts", sessionNumber, r.SeasonalPoints)
		}

		if weeksSinceLastReport <= 1 {
			sb.WriteString(fmt.Sprintf("%d. %s%s - %d days (%d weeks streak 🔥)%s\n", rank+1, cyclePrefix, r.Name, r.ActivityCount, r.Streak, seasonalInfo))
		} else {
			sb.WriteString(fmt.Sprintf("%d. %s%s - %d days (💔)%s\n", rank+1, cyclePrefix, r.Name, r.ActivityCount, seasonalInfo))
		}
	}

	sb.WriteString("\nYang udah keringetan langsung update/posting aja nanti dimasukkin klasemen 💪\n\nSemangat🔥")

	return sb.String(), nil
}

// ExecuteSeasonal returns a leaderboard ranked by seasonal points.
func (uc *GetLeaderboardUsecase) ExecuteSeasonal(ctx context.Context) (string, error) {
	reports, err := uc.repo.GetAllReports(ctx)
	if err != nil {
		return "", err
	}

	now := time.Now()
	seasonNumber, sessionStart := GetCurrentSessionInfo(now)
	_ = sessionStart

	// Sort by seasonal points descending
	sort.Slice(reports, func(i, j int) bool {
		if reports[i].SeasonalPoints == reports[j].SeasonalPoints {
			return reports[i].SeasonalActivityCount > reports[j].SeasonalActivityCount
		}
		return reports[i].SeasonalPoints > reports[j].SeasonalPoints
	})

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("🏆 *Season %d Leaderboard*\n\n", seasonNumber))

	if len(reports) == 0 {
		sb.WriteString("Belum ada yang aktif di season ini.\n")
		sb.WriteString("Jadilah yang pertama dengan #lapor! 💪")
		return sb.String(), nil
	}

	activeInSeason := 0
	for _, r := range reports {
		if r.SeasonalActivityCount > 0 {
			activeInSeason++
		}
	}

	sb.WriteString(fmt.Sprintf("Peserta aktif: %d\n\n", activeInSeason))

	rank := 1
	for _, r := range reports {
		if r.SeasonalPoints == 0 && r.SeasonalActivityCount == 0 {
			continue
		}
		cyclePrefix := ""
		if r.CenturionCycles > 0 {
			cyclePrefix = fmt.Sprintf("[S1-C%d] ", r.CenturionCycles+1)
		}
		sb.WriteString(fmt.Sprintf("%d. %s%s — %d pts (%d hari)\n", rank, cyclePrefix, r.Name, r.SeasonalPoints, r.SeasonalActivityCount))
		rank++
	}

	sb.WriteString("\nSeasonal ranking dihitung dari poin yang diraih di season ini.\n")
	sb.WriteString("Semangat naikin rank-mu! 🚀")

	return sb.String(), nil
}

// ExecuteRanks returns a concise season ranking for the #ranks command.
func (uc *GetLeaderboardUsecase) ExecuteRanks(ctx context.Context) (string, error) {
	reports, err := uc.repo.GetAllReports(ctx)
	if err != nil {
		return "", err
	}

	now := time.Now()
	seasonNumber, _ := GetCurrentSessionInfo(now)
	nextReset := GetNextResetTime(now)

	sort.Slice(reports, func(i, j int) bool {
		if reports[i].SeasonalPoints == reports[j].SeasonalPoints {
			if reports[i].SeasonalActivityCount == reports[j].SeasonalActivityCount {
				return reports[i].Name < reports[j].Name
			}
			return reports[i].SeasonalActivityCount > reports[j].SeasonalActivityCount
		}
		return reports[i].SeasonalPoints > reports[j].SeasonalPoints
	})

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("🏹 *Season %d Ranks*\n", seasonNumber))
	sb.WriteString(fmt.Sprintf("Reset badge/rank: %s\n", nextReset.Format("02-01-2006")))
	sb.WriteString("Level & EXP lifetime tetap aman.\n\n")

	rank := 1
	for _, r := range reports {
		if r.SeasonalPoints == 0 && r.SeasonalActivityCount == 0 {
			continue
		}

		badges := 0
		if r.SeasonalAchievements != "" {
			badges = len(strings.Split(r.SeasonalAchievements, ","))
		}

		sb.WriteString(fmt.Sprintf(
			"%d. %s — %s | %d pts | %d hari | %d badge\n",
			rank,
			r.Name,
			domain.FormatSeasonRank(r.SeasonalPoints),
			r.SeasonalPoints,
			r.SeasonalActivityCount,
			badges,
		))

		rank++
		if rank > 10 {
			break
		}
	}

	if rank == 1 {
		sb.WriteString("Belum ada hunter aktif season ini. Mulai dengan #lapor 💪")
		return sb.String(), nil
	}

	sb.WriteString("\nRank dihitung dari seasonal points. Badge season ikut menambah poin, lalu reset saat season baru.")

	return sb.String(), nil
}
