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
	reports = domain.DedupReportsByUserID(reports, domain.SortBySeasonRank)

	now := time.Now()
	displayDate := domain.GetToday(now)

	_, sessionStart := GetCurrentSessionInfo(now)
	startDate := time.Date(sessionStart.Year(), sessionStart.Month(), sessionStart.Day(), 0, 0, 0, 0, time.UTC)
	challengeDay := int(displayDate.Sub(startDate).Hours()/24) + 1

	domain.SortReports(reports, domain.SortBySeasonRank)

	activeCount, lostCount := countStreakStatus(reports, now)

	sb := strings.Builder{}
	dateStr := displayDate.Format("02-01-2006")
	seasonNumber, _ := GetCurrentSessionInfo(now)
	sb.WriteString(fmt.Sprintf("Season %d Hidup Sehat SWE Growth – Day %d (%s)\n\n", seasonNumber, challengeDay, dateStr))

	sb.WriteString(fmt.Sprintf("Recap day %d:\n", challengeDay))
	sb.WriteString(fmt.Sprintf("%d peoples keep the streak 🔥\n", activeCount))
	sb.WriteString(fmt.Sprintf("%d lose the streak 💔\n", lostCount))
	sb.WriteString("\nUpdate klasemen season sementara:\n")

	currentWeekStart := domain.GetStartOfISOWeek(now)
	rank := 1
	for _, r := range reports {
		if !domain.HasSeasonActivity(r) {
			continue
		}
		lastWeekStart := domain.GetStartOfISOWeek(r.LastReportDate)
		weeksSinceLastReport := int(math.Round(currentWeekStart.Sub(lastWeekStart).Hours() / (24 * 7)))

		cyclePrefix := ""
		if r.CenturionCycles > 0 {
			cyclePrefix = fmt.Sprintf("[S1-C%d] ", r.CenturionCycles+1)
		}

		if weeksSinceLastReport <= 1 {
			sb.WriteString(fmt.Sprintf("%d. %s%s — %d pts (%d hari season, %d hari lifetime, %d minggu streak 🔥)\n", rank, cyclePrefix, r.Name, r.SeasonalPoints, r.SeasonalActivityCount, r.TotalActiveDays(), r.Streak))
		} else {
			sb.WriteString(fmt.Sprintf("%d. %s%s — %d pts (%d hari season, %d hari lifetime, 💔)\n", rank, cyclePrefix, r.Name, r.SeasonalPoints, r.SeasonalActivityCount, r.TotalActiveDays()))
		}
		rank++
	}

	if rank == 1 {
		sb.WriteString("Belum ada hunter aktif season ini.\n")
	}

	sb.WriteString("\nYang udah keringetan langsung update/posting aja nanti dimasukkin klasemen 💪\n\nSemangat🔥")

	return sb.String(), nil
}

func (uc *GetLeaderboardUsecase) ExecuteSeasonal(ctx context.Context) (string, error) {
	reports, err := uc.repo.GetAllReports(ctx)
	if err != nil {
		return "", err
	}
	reports = domain.DedupReportsByUserID(reports, domain.SortBySeasonRank)

	now := time.Now()
	seasonNumber, _ := GetCurrentSessionInfo(now)

	domain.SortReports(reports, domain.SortBySeasonRank)
	active := domain.FilterReports(reports, domain.HasSeasonActivity)

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("🏆 *Season %d Leaderboard*\n\n", seasonNumber))

	if len(active) == 0 {
		sb.WriteString("Belum ada yang aktif di season ini.\n")
		sb.WriteString("Jadilah yang pertama dengan /lapor! 💪")
		return sb.String(), nil
	}

	sb.WriteString(fmt.Sprintf("Peserta aktif: %d\n\n", len(active)))

	for rank, r := range active {
		cyclePrefix := ""
		if r.CenturionCycles > 0 {
			cyclePrefix = fmt.Sprintf("[S1-C%d] ", r.CenturionCycles+1)
		}
		sb.WriteString(fmt.Sprintf("%d. %s%s — %d pts (%d hari)\n", rank+1, cyclePrefix, r.Name, r.SeasonalPoints, r.SeasonalActivityCount))
	}

	sb.WriteString("\nSeasonal ranking dihitung dari poin yang diraih di season ini.\n")
	sb.WriteString("Semangat naikin rank-mu! 🚀")

	return sb.String(), nil
}

func (uc *GetLeaderboardUsecase) ExecuteRanks(ctx context.Context) (string, error) {
	reports, err := uc.repo.GetAllReports(ctx)
	if err != nil {
		return "", err
	}
	reports = domain.DedupReportsByUserID(reports, domain.SortBySeasonRank)

	now := time.Now()
	seasonNumber, _ := GetCurrentSessionInfo(now)
	nextReset := GetNextResetTime(now)

	domain.SortReports(reports, domain.SortBySeasonRank)
	active := domain.FilterReports(reports, domain.HasSeasonActivity)

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("🏹 *Season %d Ranks*\n", seasonNumber))
	sb.WriteString(fmt.Sprintf("Reset badge/rank: %s\n", nextReset.Format("02-01-2006")))
	sb.WriteString("Level & EXP lifetime tetap aman.\n\n")

	if len(active) == 0 {
		sb.WriteString("Belum ada hunter aktif season ini. Mulai dengan /lapor 💪")
		return sb.String(), nil
	}

	maxRank := 10
	if len(active) < maxRank {
		maxRank = len(active)
	}

	for rank := 0; rank < maxRank; rank++ {
		r := active[rank]
		badges := countBadges(r.SeasonalAchievements)
		sb.WriteString(fmt.Sprintf(
			"%d. %s — %s | %d pts | %d hari | %d badge\n",
			rank+1,
			r.Name,
			domain.FormatSeasonRank(r.SeasonalPoints),
			r.SeasonalPoints,
			r.SeasonalActivityCount,
			badges,
		))
	}

	sb.WriteString("\nRank dihitung dari seasonal points. Badge season ikut menambah poin, lalu reset saat season baru.")

	return sb.String(), nil
}

// ExecuteLifetime returns a leaderboard ranked by lifetime total points (XP).
// Lifetime points never reset across seasons.
func (uc *GetLeaderboardUsecase) ExecuteLifetime(ctx context.Context) (string, error) {
	reports, err := uc.repo.GetAllReports(ctx)
	if err != nil {
		return "", err
	}
	reports = domain.DedupReportsByUserID(reports, domain.SortByLifetimeXP)

	seasonNumber, _ := GetCurrentSessionInfo(time.Now())

	domain.SortReports(reports, domain.SortByLifetimeXP)
	active := domain.FilterReports(reports, domain.HasAnyActivity)

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("⭐ *Lifetime XP Leaderboard*\n"))
	sb.WriteString(fmt.Sprintf("Season %d — total EXP sejak bergabung\n\n", seasonNumber))

	if len(active) == 0 {
		sb.WriteString("Belum ada hunter. Mulai dengan /lapor! 💪")
		return sb.String(), nil
	}

	maxRank := 10
	if len(active) < maxRank {
		maxRank = len(active)
	}

	for rank := 0; rank < maxRank; rank++ {
		r := active[rank]
		xpProg := domain.GetNumericLevelProgress(r.TotalPoints)
		sb.WriteString(fmt.Sprintf(
			"%d. %s — Lv.%d (%d EXP, %d hari lifetime, %d minggu streak max 🔥)\n",
			rank+1,
			r.Name,
			xpProg.Level,
			r.TotalPoints,
			r.TotalActiveDays(),
			r.MaxStreak,
		))
	}

	sb.WriteString("\nLifetime EXP tidak pernah reset. Terus berkeringat untuk naik level! 💪")

	return sb.String(), nil
}

// ExecuteStreakMasters returns a leaderboard ranked by streak.
// streakType must be "weekly" or "daily" — each produces its own ranking.
func (uc *GetLeaderboardUsecase) ExecuteStreakMasters(ctx context.Context, streakType string) (string, error) {
	reports, err := uc.repo.GetAllReports(ctx)
	if err != nil {
		return "", err
	}

	seasonNumber, _ := GetCurrentSessionInfo(time.Now())

	var key domain.LeaderboardSortKey
	var title, icon, metricLabel string

	normalizedType := strings.ToLower(streakType)
	switch normalizedType {
	case "daily":
		key = domain.SortByDailyStreak
		title = "Streak Masters — Daily"
		icon = "📅"
		metricLabel = "hari beruntun"
	case "weekly", "":
		key = domain.SortByWeeklyStreak
		title = "Streak Masters — Weekly"
		icon = "🔥"
		metricLabel = "minggu beruntun"
	default:
		return "", fmt.Errorf("invalid streak type %q: expected daily or weekly", streakType)
	}
	reports = domain.DedupReportsByUserID(reports, key)
	if normalizedType == "daily" {
		return uc.executeDailyStreakMasters(ctx, reports, seasonNumber, title, icon, metricLabel)
	}
	if normalizedType == "weekly" || normalizedType == "" {
		return uc.executeWeeklyStreakMasters(ctx, reports, seasonNumber, title, icon, metricLabel)
	}

	domain.SortReports(reports, key)
	active := domain.FilterReports(reports, domain.HasStreakActivity)

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s *%s*\n", icon, title))
	sb.WriteString(fmt.Sprintf("Season %d\n\n", seasonNumber))

	if len(active) == 0 {
		sb.WriteString("Belum ada hunter dengan streak. Mulai dengan /lapor! 💪")
		return sb.String(), nil
	}

	maxRank := 10
	if len(active) < maxRank {
		maxRank = len(active)
	}

	for rank := 0; rank < maxRank; rank++ {
		r := active[rank]
		sb.WriteString(formatStreakEntry(rank+1, r, normalizedType, metricLabel))
	}

	sb.WriteString("\nStreak dihitung dari konsistensi laporan. Jangan sampai putus! 🔥")

	return sb.String(), nil
}

// ExecuteAttribute returns a leaderboard ranked by attribute values.
// attrType must be "overall", "str", "sta", "agi", or "vit".
func (uc *GetLeaderboardUsecase) ExecuteAttribute(ctx context.Context, attrType string) (string, error) {
	reports, err := uc.repo.GetAllReports(ctx)
	if err != nil {
		return "", err
	}

	seasonNumber, _ := GetCurrentSessionInfo(time.Now())

	key, label, icon, err := resolveAttributeSortKey(attrType)
	if err != nil {
		return "", err
	}
	reports = domain.DedupReportsByUserID(reports, key)

	domain.SortReports(reports, key)
	active := domain.FilterReports(reports, domain.HasAttributeActivity)

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s *Attribute Leaderboard — %s*\n", icon, label))
	sb.WriteString(fmt.Sprintf("Season %d\n\n", seasonNumber))

	if len(active) == 0 {
		sb.WriteString("Belum ada hunter dengan attribute. Laporkan aktivitas untuk naikin attribute! 💪")
		return sb.String(), nil
	}

	maxRank := 10
	if len(active) < maxRank {
		maxRank = len(active)
	}

	attrEnum := domain.AttrStr
	switch key {
	case domain.SortByAttributeOverall:
		attrEnum = ""
	case domain.SortByAttributeSTA:
		attrEnum = domain.AttrSta
	case domain.SortByAttributeAGI:
		attrEnum = domain.AttrAgi
	case domain.SortByAttributeVIT:
		attrEnum = domain.AttrVit
	}

	for rank := 0; rank < maxRank; rank++ {
		r := active[rank]
		val := r.AttributeValue(attrEnum)
		sb.WriteString(fmt.Sprintf(
			"%d. %s — %d %s pts (STR %d | STA %d | AGI %d | VIT %d)\n",
			rank+1,
			r.Name,
			val,
			label,
			domain.ClampedAttribute(r.Str),
			domain.ClampedAttribute(r.Sta),
			domain.ClampedAttribute(r.Agi),
			domain.ClampedAttribute(r.Vit),
		))
	}

	sb.WriteString("\nAttribute naik dari laporan aktivitas. Variasikan latihanmu! ⚔️")

	return sb.String(), nil
}

type dailyStreakEntry struct {
	report  *domain.Report
	current int
	longest int
}

type weeklyStreakEntry struct {
	report       *domain.Report
	dailyCurrent int
}

func (uc *GetLeaderboardUsecase) executeWeeklyStreakMasters(ctx context.Context, reports []*domain.Report, seasonNumber int, title, icon, metricLabel string) (string, error) {
	today := domain.GetToday(time.Now())
	entries := make([]weeklyStreakEntry, 0, len(reports))
	for _, report := range reports {
		if !domain.HasStreakActivity(report) {
			continue
		}
		dates, err := uc.repo.GetUserActivityDates(ctx, report.UserID)
		if err != nil {
			return "", err
		}
		current, _ := calculateDailyStreaks(dates, today)
		entries = append(entries, weeklyStreakEntry{report: report, dailyCurrent: current})
	}

	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.report.Streak != b.report.Streak {
			return a.report.Streak > b.report.Streak
		}
		if a.dailyCurrent != b.dailyCurrent {
			return a.dailyCurrent > b.dailyCurrent
		}
		if a.report.SeasonalPoints != b.report.SeasonalPoints {
			return a.report.SeasonalPoints > b.report.SeasonalPoints
		}
		if a.report.MaxStreak != b.report.MaxStreak {
			return a.report.MaxStreak > b.report.MaxStreak
		}
		return domain.CompareReports(a.report, b.report, domain.SortByWeeklyStreak)
	})

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s *%s*\n", icon, title))
	sb.WriteString(fmt.Sprintf("Season %d\n\n", seasonNumber))

	if len(entries) == 0 {
		sb.WriteString("Belum ada hunter dengan streak. Mulai dengan /lapor! 💪")
		return sb.String(), nil
	}

	maxRank := 10
	if len(entries) < maxRank {
		maxRank = len(entries)
	}

	for rank := 0; rank < maxRank; rank++ {
		entry := entries[rank]
		sb.WriteString(fmt.Sprintf(
			"%d. %s — %d %s (current %d hari, %d pts, max %d minggu)\n",
			rank+1,
			entry.report.Name,
			entry.report.Streak,
			metricLabel,
			entry.dailyCurrent,
			entry.report.SeasonalPoints,
			entry.report.MaxStreak,
		))
	}

	sb.WriteString("\nPeringkat: streak mingguan → current streak harian → poin season. Jangan sampai putus! 🔥")

	return sb.String(), nil
}

func (uc *GetLeaderboardUsecase) executeDailyStreakMasters(ctx context.Context, reports []*domain.Report, seasonNumber int, title, icon, metricLabel string) (string, error) {
	today := domain.GetToday(time.Now())
	entries := make([]dailyStreakEntry, 0, len(reports))
	for _, report := range reports {
		dates, err := uc.repo.GetUserActivityDates(ctx, report.UserID)
		if err != nil {
			return "", err
		}
		current, longest := calculateDailyStreaks(dates, today)
		if current > 0 {
			entries = append(entries, dailyStreakEntry{report: report, current: current, longest: longest})
		}
	}

	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.current != b.current {
			return a.current > b.current
		}
		if a.longest != b.longest {
			return a.longest > b.longest
		}
		if a.report.TotalPoints != b.report.TotalPoints {
			return a.report.TotalPoints > b.report.TotalPoints
		}
		return domain.CompareReports(a.report, b.report, domain.SortByLifetimeXP)
	})

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s *%s*\n", icon, title))
	sb.WriteString(fmt.Sprintf("Season %d\n\n", seasonNumber))

	if len(entries) == 0 {
		sb.WriteString("Belum ada hunter dengan streak harian aktif. Mulai dengan /lapor! 💪")
		return sb.String(), nil
	}

	maxRank := 10
	if len(entries) < maxRank {
		maxRank = len(entries)
	}

	for rank := 0; rank < maxRank; rank++ {
		entry := entries[rank]
		sb.WriteString(fmt.Sprintf(
			"%d. %s — %d %s (best %d hari, %d pts)\n",
			rank+1,
			entry.report.Name,
			entry.current,
			metricLabel,
			entry.longest,
			entry.report.TotalPoints,
		))
	}

	sb.WriteString("\nStreak harian dihitung dari laporan beruntun per hari. Jangan sampai putus! 🔥")

	return sb.String(), nil
}

func calculateDailyStreaks(activityDates []time.Time, today time.Time) (int, int) {
	sort.Slice(activityDates, func(i, j int) bool {
		return activityDates[i].Before(activityDates[j])
	})

	activeDates := make(map[string]bool, len(activityDates))
	for _, date := range activityDates {
		activeDates[date.Format(time.DateOnly)] = true
	}

	current := 0
	for date := today; activeDates[date.Format(time.DateOnly)]; date = date.AddDate(0, 0, -1) {
		current++
	}

	longest := 0
	running := 0
	var previous time.Time
	for _, date := range activityDates {
		day := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
		if !previous.IsZero() && day.Equal(previous) {
			continue
		}
		if !previous.IsZero() && day.Equal(previous.AddDate(0, 0, 1)) {
			running++
		} else {
			running = 1
		}
		if running > longest {
			longest = running
		}
		previous = day
	}

	return current, longest
}

func countStreakStatus(reports []*domain.Report, now time.Time) (active, lost int) {
	currentWeekStart := domain.GetStartOfISOWeek(now)
	for _, r := range reports {
		lastWeekStart := domain.GetStartOfISOWeek(r.LastReportDate)
		weeksSinceLastReport := int(math.Round(currentWeekStart.Sub(lastWeekStart).Hours() / (24 * 7)))
		if weeksSinceLastReport <= 1 {
			active++
		} else {
			lost++
		}
	}
	return
}

func countBadges(achievements string) int {
	if achievements == "" {
		return 0
	}
	return len(strings.Split(achievements, ","))
}

func formatStreakEntry(rank int, r *domain.Report, streakType, metricLabel string) string {
	switch streakType {
	case "daily":
		return fmt.Sprintf("%d. %s — %d %s (max %d hari, %d pts)\n", rank, r.Name, r.ActivityCount, metricLabel, r.TotalActiveDays(), r.TotalPoints)
	default:
		return fmt.Sprintf("%d. %s — %d %s (max %d minggu, %d pts)\n", rank, r.Name, r.Streak, metricLabel, r.MaxStreak, r.SeasonalPoints)
	}
}

func resolveAttributeSortKey(attrType string) (key domain.LeaderboardSortKey, label, icon string, err error) {
	switch strings.ToLower(attrType) {
	case "overall", "":
		return domain.SortByAttributeOverall, "Overall", "🌟", nil
	case "str":
		return domain.SortByAttributeSTR, "STR", "💪", nil
	case "sta":
		return domain.SortByAttributeSTA, "STA", "🏃", nil
	case "agi":
		return domain.SortByAttributeAGI, "AGI", "⚡", nil
	case "vit":
		return domain.SortByAttributeVIT, "VIT", "💚", nil
	default:
		return "", "", "", fmt.Errorf("invalid attribute type %q: expected overall, str, sta, agi, or vit", attrType)
	}
}
