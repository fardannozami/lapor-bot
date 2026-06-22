package domain

import "sort"

// LeaderboardSortKey identifies which metric a leaderboard is ranked by.
// Each key maps to a dedicated comparator so sorting logic stays in one place
// and is testable in isolation.
type LeaderboardSortKey string

const (
	// SortBySeasonRank ranks by seasonal points, then seasonal activity count,
	// then weekly streak, then daily streak, then name.
	// This is the canonical season leaderboard.
	SortBySeasonRank LeaderboardSortKey = "season_rank"

	// SortByLifetimeXP ranks by total (lifetime) points — never resets.
	// Tie-break: total active days, then max streak, then name.
	SortByLifetimeXP LeaderboardSortKey = "lifetime_xp"

	// SortByWeeklyStreak ranks by current weekly streak (consecutive weeks).
	// Tie-break: lifetime max streak, then seasonal points, then name.
	SortByWeeklyStreak LeaderboardSortKey = "weekly_streak"

	// SortByDailyStreak ranks by current daily streak (consecutive days).
	// Since the bot stores streak as a weekly counter, we approximate daily
	// streak from total active days in the current cycle.
	// Tie-break: max streak, then total points, then name.
	SortByDailyStreak LeaderboardSortKey = "daily_streak"

	// SortByWeeklyActivity ranks by activity count in the current week.
	// Tie-break: weekly streak, then seasonal points, then name.
	// This follows the user's requested order: laporan duluan → streak → poin.
	SortByWeeklyActivity LeaderboardSortKey = "weekly_activity"

	// SortByAttributeOverall ranks by the average of all four attributes.
	// Tie-break: total points, then name.
	SortByAttributeOverall LeaderboardSortKey = "attribute_overall"

	// SortByAttributeSTR ranks by STR attribute value.
	SortByAttributeSTR LeaderboardSortKey = "attribute_str"

	// SortByAttributeSTA ranks by STA attribute value.
	SortByAttributeSTA LeaderboardSortKey = "attribute_sta"

	// SortByAttributeAGI ranks by AGI attribute value.
	SortByAttributeAGI LeaderboardSortKey = "attribute_agi"

	// SortByAttributeVIT ranks by VIT attribute value.
	SortByAttributeVIT LeaderboardSortKey = "attribute_vit"
)

// TotalActiveDays returns the lifetime count of active days, accounting for
// centurion cycles (each cycle = 100 days).
func (r *Report) TotalActiveDays() int {
	return r.CenturionCycles*100 + r.ActivityCount
}

// AttributeAverage returns the rounded average of all four RPG attributes.
// Values are clamped to MinAttributeValue before averaging so the baseline
// (start from 1) invariant holds.
func (r *Report) AttributeAverage() int {
	total := ClampedAttribute(r.Str) +
		ClampedAttribute(r.Sta) +
		ClampedAttribute(r.Agi) +
		ClampedAttribute(r.Vit)
	return total / 4
}

// AttributeValue returns the raw attribute value for the given type.
// For AttrOverall it returns the average.
func (r *Report) AttributeValue(attr AttributeType) int {
	switch attr {
	case AttrStr:
		return ClampedAttribute(r.Str)
	case AttrSta:
		return ClampedAttribute(r.Sta)
	case AttrAgi:
		return ClampedAttribute(r.Agi)
	case AttrVit:
		return ClampedAttribute(r.Vit)
	default:
		return r.AttributeAverage()
	}
}

// CompareReports compares two reports for leaderboard ordering.
// Returns true if report a should rank higher (come before) report b.
// This is the single source of truth for all leaderboard sorting.
func CompareReports(a, b *Report, key LeaderboardSortKey) bool {
	switch key {
	case SortBySeasonRank:
		return compareSeasonRank(a, b)
	case SortByLifetimeXP:
		return compareLifetimeXP(a, b)
	case SortByWeeklyStreak:
		return compareWeeklyStreak(a, b)
	case SortByDailyStreak:
		return compareDailyStreak(a, b)
	case SortByWeeklyActivity:
		return compareWeeklyActivity(a, b)
	case SortByAttributeOverall:
		return compareAttributeOverall(a, b)
	case SortByAttributeSTR, SortByAttributeSTA, SortByAttributeAGI, SortByAttributeVIT:
		return compareAttribute(a, b, key)
	default:
		return compareSeasonRank(a, b)
	}
}

// SortReports sorts a slice of reports in-place using the given sort key.
func SortReports(reports []*Report, key LeaderboardSortKey) {
	sort.Slice(reports, func(i, j int) bool {
		return CompareReports(reports[i], reports[j], key)
	})
}

// HasSeasonActivity returns true if a report has any seasonal engagement.
// Users with zero points AND zero activity are excluded from season boards.
func HasSeasonActivity(r *Report) bool {
	return r.SeasonalPoints > 0 || r.SeasonalActivityCount > 0
}

// HasAnyActivity returns true if a report has any lifetime engagement.
// Used to filter out completely inactive users from lifetime boards.
func HasAnyActivity(r *Report) bool {
	return r.TotalPoints > 0 || r.ActivityCount > 0 || r.CenturionCycles > 0
}

// HasStreakActivity returns true if a report has any streak worth ranking.
// Used to filter the Streak Masters leaderboard.
func HasStreakActivity(r *Report) bool {
	return r.Streak > 0 || r.MaxStreak > 0 || r.SeasonalMaxStreak > 0 || r.ActivityCount > 0
}

// HasAttributeActivity returns true if a report has at least one attribute
// point across any of the four RPG attributes.
func HasAttributeActivity(r *Report) bool {
	return r.Str > 0 || r.Sta > 0 || r.Agi > 0 || r.Vit > 0
}

// FilterReports returns a new slice containing only reports that pass the predicate.
func FilterReports(reports []*Report, keep func(*Report) bool) []*Report {
	result := make([]*Report, 0, len(reports))
	for _, r := range reports {
		if keep(r) {
			result = append(result, r)
		}
	}
	return result
}

// --- Internal comparators ---
// Each follows the pattern: primary metric → tie-breakers → name (stable).

func compareSeasonRank(a, b *Report) bool {
	if a.SeasonalPoints != b.SeasonalPoints {
		return a.SeasonalPoints > b.SeasonalPoints
	}
	if a.SeasonalActivityCount != b.SeasonalActivityCount {
		return a.SeasonalActivityCount > b.SeasonalActivityCount
	}
	if a.Streak != b.Streak {
		return a.Streak > b.Streak
	}
	if a.TotalActiveDays() != b.TotalActiveDays() {
		return a.TotalActiveDays() > b.TotalActiveDays()
	}
	return a.Name < b.Name
}

func compareLifetimeXP(a, b *Report) bool {
	if a.TotalPoints != b.TotalPoints {
		return a.TotalPoints > b.TotalPoints
	}
	if a.TotalActiveDays() != b.TotalActiveDays() {
		return a.TotalActiveDays() > b.TotalActiveDays()
	}
	if a.MaxStreak != b.MaxStreak {
		return a.MaxStreak > b.MaxStreak
	}
	return a.Name < b.Name
}

func compareWeeklyStreak(a, b *Report) bool {
	if a.Streak != b.Streak {
		return a.Streak > b.Streak
	}
	if a.MaxStreak != b.MaxStreak {
		return a.MaxStreak > b.MaxStreak
	}
	if a.SeasonalPoints != b.SeasonalPoints {
		return a.SeasonalPoints > b.SeasonalPoints
	}
	return a.Name < b.Name
}

func compareDailyStreak(a, b *Report) bool {
	// Daily streak is approximated from total active days in the current cycle.
	aDaily := a.TotalActiveDays()
	bDaily := b.TotalActiveDays()
	if aDaily != bDaily {
		return aDaily > bDaily
	}
	if a.MaxStreak != b.MaxStreak {
		return a.MaxStreak > b.MaxStreak
	}
	if a.TotalPoints != b.TotalPoints {
		return a.TotalPoints > b.TotalPoints
	}
	return a.Name < b.Name
}

func compareWeeklyActivity(a, b *Report) bool {
	// User's requested order: laporan duluan (activity) → streak → poin
	if a.SeasonalActivityCount != b.SeasonalActivityCount {
		return a.SeasonalActivityCount > b.SeasonalActivityCount
	}
	if a.Streak != b.Streak {
		return a.Streak > b.Streak
	}
	if a.SeasonalPoints != b.SeasonalPoints {
		return a.SeasonalPoints > b.SeasonalPoints
	}
	return a.Name < b.Name
}

func compareAttributeOverall(a, b *Report) bool {
	aAvg := a.AttributeAverage()
	bAvg := b.AttributeAverage()
	if aAvg != bAvg {
		return aAvg > bAvg
	}
	if a.TotalPoints != b.TotalPoints {
		return a.TotalPoints > b.TotalPoints
	}
	return a.Name < b.Name
}

func compareAttribute(a, b *Report, key LeaderboardSortKey) bool {
	var attr AttributeType
	switch key {
	case SortByAttributeSTR:
		attr = AttrStr
	case SortByAttributeSTA:
		attr = AttrSta
	case SortByAttributeAGI:
		attr = AttrAgi
	case SortByAttributeVIT:
		attr = AttrVit
	default:
		attr = AttrStr
	}

	aVal := a.AttributeValue(attr)
	bVal := b.AttributeValue(attr)
	if aVal != bVal {
		return aVal > bVal
	}
	if a.TotalPoints != b.TotalPoints {
		return a.TotalPoints > b.TotalPoints
	}
	return a.Name < b.Name
}

// AttributeSortKeyFromType converts an AttributeType to its corresponding
// LeaderboardSortKey. Returns SortByAttributeOverall for unknown types.
func AttributeSortKeyFromType(attr AttributeType) LeaderboardSortKey {
	switch attr {
	case AttrStr:
		return SortByAttributeSTR
	case AttrSta:
		return SortByAttributeSTA
	case AttrAgi:
		return SortByAttributeAGI
	case AttrVit:
		return SortByAttributeVIT
	default:
		return SortByAttributeOverall
	}
}
