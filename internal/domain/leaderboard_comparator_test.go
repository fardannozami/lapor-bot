package domain

import (
	"testing"
)

func makeReport(name string, seasonalPoints, seasonalActivity, streak, totalPoints, maxStreak, activityCount, centurionCycles, str, sta, agi, vit int) *Report {
	return &Report{
		UserID:                name,
		Name:                  name,
		SeasonalPoints:        seasonalPoints,
		SeasonalActivityCount: seasonalActivity,
		Streak:                streak,
		TotalPoints:           totalPoints,
		MaxStreak:             maxStreak,
		ActivityCount:         activityCount,
		CenturionCycles:       centurionCycles,
		Str:                   str,
		Sta:                   sta,
		Agi:                   agi,
		Vit:                   vit,
	}
}

func TestCompareSeasonRank_PointsPrimary(t *testing.T) {
	a := makeReport("Alice", 300, 10, 3, 500, 5, 10, 0, 10, 10, 10, 10)
	b := makeReport("Bob", 200, 20, 5, 600, 8, 20, 0, 20, 20, 20, 20)

	if !CompareReports(a, b, SortBySeasonRank) {
		t.Errorf("higher seasonal points should rank first")
	}
}

func TestCompareSeasonRank_ActivityTiebreak(t *testing.T) {
	a := makeReport("Alice", 300, 15, 2, 0, 0, 15, 0, 0, 0, 0, 0)
	b := makeReport("Bob", 300, 10, 5, 0, 0, 10, 0, 0, 0, 0, 0)

	if !CompareReports(a, b, SortBySeasonRank) {
		t.Errorf("with equal points, higher activity count should rank first")
	}
}

func TestCompareSeasonRank_StreakTiebreak(t *testing.T) {
	a := makeReport("Alice", 300, 10, 5, 0, 0, 10, 0, 0, 0, 0, 0)
	b := makeReport("Bob", 300, 10, 2, 0, 0, 10, 0, 0, 0, 0, 0)

	if !CompareReports(a, b, SortBySeasonRank) {
		t.Errorf("with equal points and activity, higher streak should rank first")
	}
}

func TestCompareSeasonRank_NameFinalTiebreak(t *testing.T) {
	a := makeReport("Alice", 300, 10, 3, 100, 5, 10, 0, 10, 10, 10, 10)
	b := makeReport("Bob", 300, 10, 3, 100, 5, 10, 0, 10, 10, 10, 10)

	if !CompareReports(a, b, SortBySeasonRank) {
		t.Errorf("with all metrics equal, name should be the final tiebreak (Alice before Bob)")
	}
}

func TestCompareLifetimeXP_PointsPrimary(t *testing.T) {
	a := makeReport("Alice", 0, 0, 0, 1000, 5, 10, 0, 0, 0, 0, 0)
	b := makeReport("Bob", 0, 0, 0, 500, 10, 20, 0, 0, 0, 0, 0)

	if !CompareReports(a, b, SortByLifetimeXP) {
		t.Errorf("higher lifetime points should rank first")
	}
}

func TestCompareLifetimeXP_ActiveDaysTiebreak(t *testing.T) {
	a := makeReport("Alice", 0, 0, 0, 500, 5, 150, 1, 0, 0, 0, 0)
	b := makeReport("Bob", 0, 0, 0, 500, 10, 50, 0, 0, 0, 0, 0)

	if !CompareReports(a, b, SortByLifetimeXP) {
		t.Errorf("with equal points, higher total active days should rank first")
	}
}

func TestCompareWeeklyStreak_StreakPrimary(t *testing.T) {
	a := makeReport("Alice", 100, 5, 8, 200, 10, 5, 0, 0, 0, 0, 0)
	b := makeReport("Bob", 300, 20, 3, 500, 5, 20, 0, 0, 0, 0, 0)

	if !CompareReports(a, b, SortByWeeklyStreak) {
		t.Errorf("higher weekly streak should rank first regardless of points")
	}
}

func TestCompareWeeklyStreak_MaxStreakTiebreak(t *testing.T) {
	a := makeReport("Alice", 100, 5, 4, 200, 10, 5, 0, 0, 0, 0, 0)
	b := makeReport("Bob", 300, 20, 4, 500, 5, 20, 0, 0, 0, 0, 0)

	if !CompareReports(a, b, SortByWeeklyStreak) {
		t.Errorf("with equal streak, higher max streak should rank first")
	}
}

func TestCompareDailyStreak_ActivityCountPrimary(t *testing.T) {
	a := makeReport("Alice", 100, 5, 2, 200, 10, 15, 0, 0, 0, 0, 0)
	b := makeReport("Bob", 300, 20, 5, 500, 5, 5, 0, 0, 0, 0, 0)

	if !CompareReports(a, b, SortByDailyStreak) {
		t.Errorf("higher activity count (daily streak proxy) should rank first")
	}
}

func TestCompareWeeklyActivity_ActivityPrimary(t *testing.T) {
	a := makeReport("Alice", 100, 20, 3, 200, 5, 20, 0, 0, 0, 0, 0)
	b := makeReport("Bob", 300, 5, 5, 500, 10, 5, 0, 0, 0, 0, 0)

	if !CompareReports(a, b, SortByWeeklyActivity) {
		t.Errorf("higher activity count should rank first for weekly activity")
	}
}

func TestCompareWeeklyActivity_StreakTiebreak(t *testing.T) {
	a := makeReport("Alice", 100, 10, 5, 200, 5, 10, 0, 0, 0, 0, 0)
	b := makeReport("Bob", 300, 10, 2, 500, 5, 10, 0, 0, 0, 0, 0)

	if !CompareReports(a, b, SortByWeeklyActivity) {
		t.Errorf("with equal activity, higher streak should rank first")
	}
}

func TestCompareWeeklyActivity_PointsTiebreak(t *testing.T) {
	a := makeReport("Alice", 300, 10, 3, 200, 5, 10, 0, 0, 0, 0, 0)
	b := makeReport("Bob", 100, 10, 3, 500, 5, 10, 0, 0, 0, 0, 0)

	if !CompareReports(a, b, SortByWeeklyActivity) {
		t.Errorf("with equal activity and streak, higher seasonal points should rank first")
	}
}

func TestCompareAttributeOverall_AveragePrimary(t *testing.T) {
	a := makeReport("Alice", 0, 0, 0, 200, 0, 0, 0, 40, 40, 40, 40)
	b := makeReport("Bob", 0, 0, 0, 500, 0, 0, 0, 10, 10, 10, 10)

	if !CompareReports(a, b, SortByAttributeOverall) {
		t.Errorf("higher attribute average should rank first")
	}
}

func TestCompareAttributeOverall_PointsTiebreak(t *testing.T) {
	a := makeReport("Alice", 0, 0, 0, 300, 0, 0, 0, 20, 20, 20, 20)
	b := makeReport("Bob", 0, 0, 0, 100, 0, 0, 0, 20, 20, 20, 20)

	if !CompareReports(a, b, SortByAttributeOverall) {
		t.Errorf("with equal attribute average, higher total points should rank first")
	}
}

func TestCompareAttributeSTR(t *testing.T) {
	a := makeReport("Alice", 0, 0, 0, 100, 0, 0, 0, 50, 1, 1, 1)
	b := makeReport("Bob", 0, 0, 0, 500, 0, 0, 0, 10, 50, 50, 50)

	if !CompareReports(a, b, SortByAttributeSTR) {
		t.Errorf("higher STR should rank first for STR attribute leaderboard")
	}
}

func TestCompareAttributeSTA(t *testing.T) {
	a := makeReport("Alice", 0, 0, 0, 100, 0, 0, 0, 1, 50, 1, 1)
	b := makeReport("Bob", 0, 0, 0, 500, 0, 0, 0, 50, 10, 50, 50)

	if !CompareReports(a, b, SortByAttributeSTA) {
		t.Errorf("higher STA should rank first for STA attribute leaderboard")
	}
}

func TestCompareAttributeAGI(t *testing.T) {
	a := makeReport("Alice", 0, 0, 0, 100, 0, 0, 0, 1, 1, 50, 1)
	b := makeReport("Bob", 0, 0, 0, 500, 0, 0, 0, 50, 50, 10, 50)

	if !CompareReports(a, b, SortByAttributeAGI) {
		t.Errorf("higher AGI should rank first for AGI attribute leaderboard")
	}
}

func TestCompareAttributeVIT(t *testing.T) {
	a := makeReport("Alice", 0, 0, 0, 100, 0, 0, 0, 1, 1, 1, 50)
	b := makeReport("Bob", 0, 0, 0, 500, 0, 0, 0, 50, 50, 50, 10)

	if !CompareReports(a, b, SortByAttributeVIT) {
		t.Errorf("higher VIT should rank first for VIT attribute leaderboard")
	}
}

func TestSortReports_SortsInPlace(t *testing.T) {
	reports := []*Report{
		makeReport("Charlie", 100, 5, 1, 0, 0, 5, 0, 0, 0, 0, 0),
		makeReport("Alice", 300, 10, 3, 0, 0, 10, 0, 0, 0, 0, 0),
		makeReport("Bob", 200, 15, 2, 0, 0, 15, 0, 0, 0, 0, 0),
	}

	SortReports(reports, SortBySeasonRank)

	if reports[0].Name != "Alice" {
		t.Errorf("expected Alice first (300 pts), got %s", reports[0].Name)
	}
	if reports[1].Name != "Bob" {
		t.Errorf("expected Bob second (200 pts), got %s", reports[1].Name)
	}
	if reports[2].Name != "Charlie" {
		t.Errorf("expected Charlie third (100 pts), got %s", reports[2].Name)
	}
}

func TestDedupReportsByUserID_KeepsBestReportForSortKey(t *testing.T) {
	reports := []*Report{
		{UserID: "u1", Name: "Alice", SeasonalPoints: 10, SeasonalActivityCount: 1},
		{UserID: "u2", Name: "Bob", SeasonalPoints: 20, SeasonalActivityCount: 1},
		{UserID: "u1", Name: "Alice", SeasonalPoints: 30, SeasonalActivityCount: 2},
	}

	got := DedupReportsByUserID(reports, SortBySeasonRank)

	if len(got) != 2 {
		t.Fatalf("expected 2 reports after dedupe, got %d", len(got))
	}
	if got[0].UserID != "u1" || got[0].SeasonalPoints != 30 {
		t.Fatalf("expected best u1 row kept in original user order, got %+v", got[0])
	}
	if got[1].UserID != "u2" {
		t.Fatalf("expected u2 second, got %+v", got[1])
	}
}

func TestCompareReports_NameTieBreaksByUserID(t *testing.T) {
	a := makeReport("Same", 100, 1, 1, 100, 1, 1, 0, 1, 1, 1, 1)
	b := makeReport("Same", 100, 1, 1, 100, 1, 1, 0, 1, 1, 1, 1)
	a.UserID = "1"
	b.UserID = "2"

	if !CompareReports(a, b, SortBySeasonRank) {
		t.Errorf("with equal name and metrics, lower user_id should rank first")
	}
}

func TestHasSeasonActivity(t *testing.T) {
	tests := []struct {
		name string
		r    *Report
		want bool
	}{
		{"zero everything", makeReport("A", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0), false},
		{"has points", makeReport("A", 10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0), true},
		{"has activity", makeReport("A", 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasSeasonActivity(tt.r); got != tt.want {
				t.Errorf("HasSeasonActivity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasAnyActivity(t *testing.T) {
	tests := []struct {
		name string
		r    *Report
		want bool
	}{
		{"zero everything", makeReport("A", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0), false},
		{"has total points", makeReport("A", 0, 0, 0, 10, 0, 0, 0, 0, 0, 0, 0), true},
		{"has activity count", makeReport("A", 0, 0, 0, 0, 0, 5, 0, 0, 0, 0, 0), true},
		{"has centurion cycles", makeReport("A", 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasAnyActivity(tt.r); got != tt.want {
				t.Errorf("HasAnyActivity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasStreakActivity(t *testing.T) {
	tests := []struct {
		name string
		r    *Report
		want bool
	}{
		{"zero streak", makeReport("A", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0), false},
		{"has current streak", makeReport("A", 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0), true},
		{"has max streak", makeReport("A", 0, 0, 0, 0, 5, 0, 0, 0, 0, 0, 0), true},
		{"has activity count", makeReport("A", 0, 0, 0, 0, 0, 5, 0, 0, 0, 0, 0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasStreakActivity(tt.r); got != tt.want {
				t.Errorf("HasStreakActivity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasAttributeActivity(t *testing.T) {
	tests := []struct {
		name string
		r    *Report
		want bool
	}{
		{"all zero", makeReport("A", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0), false},
		{"has STR", makeReport("A", 0, 0, 0, 0, 0, 0, 0, 5, 0, 0, 0), true},
		{"has STA", makeReport("A", 0, 0, 0, 0, 0, 0, 0, 0, 5, 0, 0), true},
		{"has AGI", makeReport("A", 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 0), true},
		{"has VIT", makeReport("A", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasAttributeActivity(tt.r); got != tt.want {
				t.Errorf("HasAttributeActivity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReport_TotalActiveDays(t *testing.T) {
	tests := []struct {
		name            string
		activityCount   int
		centurionCycles int
		want            int
	}{
		{"zero", 0, 0, 0},
		{"only activity", 50, 0, 50},
		{"with one cycle", 30, 1, 130},
		{"multiple cycles", 5, 3, 305},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Report{ActivityCount: tt.activityCount, CenturionCycles: tt.centurionCycles}
			if got := r.TotalActiveDays(); got != tt.want {
				t.Errorf("TotalActiveDays() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReport_AttributeAverage(t *testing.T) {
	tests := []struct {
		name               string
		str, sta, agi, vit int
		want               int
	}{
		{"all zero clamps to 1", 0, 0, 0, 0, 1},
		{"all ten", 10, 10, 10, 10, 10},
		{"mixed values", 40, 20, 10, 30, 25},
		{"one high", 100, 1, 1, 1, 25},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Report{Str: tt.str, Sta: tt.sta, Agi: tt.agi, Vit: tt.vit}
			if got := r.AttributeAverage(); got != tt.want {
				t.Errorf("AttributeAverage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterReports(t *testing.T) {
	reports := []*Report{
		makeReport("Active1", 100, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0),
		makeReport("Inactive", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0),
		makeReport("Active2", 50, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0),
	}

	filtered := FilterReports(reports, HasSeasonActivity)

	if len(filtered) != 2 {
		t.Fatalf("expected 2 active reports, got %d", len(filtered))
	}
	if filtered[0].Name != "Active1" {
		t.Errorf("expected Active1 first, got %s", filtered[0].Name)
	}
	if filtered[1].Name != "Active2" {
		t.Errorf("expected Active2 second, got %s", filtered[1].Name)
	}
}

func TestAttributeSortKeyFromType(t *testing.T) {
	tests := []struct {
		attr AttributeType
		want LeaderboardSortKey
	}{
		{AttrStr, SortByAttributeSTR},
		{AttrSta, SortByAttributeSTA},
		{AttrAgi, SortByAttributeAGI},
		{AttrVit, SortByAttributeVIT},
		{AttributeType("unknown"), SortByAttributeOverall},
	}
	for _, tt := range tests {
		if got := AttributeSortKeyFromType(tt.attr); got != tt.want {
			t.Errorf("AttributeSortKeyFromType(%q) = %v, want %v", tt.attr, got, tt.want)
		}
	}
}
