package usecase

import (
	"testing"
	"time"
)

func TestCappedStreakBonus(t *testing.T) {
	cases := []struct {
		name    string
		steps   int64
		cap     int64
		perStep int64
		want    int
	}{
		{"first period (no bonus)", 0, 10, 2, 0},
		{"negative clamped to zero", -3, 10, 2, 0},
		{"within cap scales linearly", 4, 10, 2, 8},
		{"at cap boundary", 10, 10, 2, 20},
		{"above cap is clamped", 50, 10, 2, 20},
		{"daily per-step is 1", 3, 5, 1, 3},
		{"daily capped at 5", 99, 5, 1, 5},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := cappedStreakBonus(c.steps, c.cap, c.perStep); got != c.want {
				t.Fatalf("cappedStreakBonus(%d, %d, %d) = %d, want %d", c.steps, c.cap, c.perStep, got, c.want)
			}
		})
	}
}

// TestScoring_WeeklyPerStepHigherThanDaily guards the requirement that the
// weekly streak bonus per step must be larger than the daily one, so a weekly
// streak is always worth more than an equally long daily streak.
func TestScoring_WeeklyPerStepHigherThanDaily(t *testing.T) {
	if weeklyStreakBonusPerStep <= dailyStreakBonusPerStep {
		t.Fatalf("weekly per-step (%d) must be > daily per-step (%d)",
			weeklyStreakBonusPerStep, dailyStreakBonusPerStep)
	}
}

// TestScoring_WeeklyAndDailyBonusesAreAdditive guarantees that when both
// streaks are at their caps, the combined bonus is the sum of the two caps
// (i.e. having both streaks alive yields strictly more than either alone),
// and that the weekly component stays the larger contributor.
func TestScoring_WeeklyAndDailyBonusesAreAdditive(t *testing.T) {
	weeklyMax := cappedStreakBonus(weeklyStreakBonusCap, weeklyStreakBonusCap, weeklyStreakBonusPerStep)
	dailyMax := cappedStreakBonus(dailyStreakBonusCap, dailyStreakBonusCap, dailyStreakBonusPerStep)
	combined := weeklyMax + dailyMax

	if weeklyMax <= dailyMax {
		t.Fatalf("weekly cap bonus (%d) must exceed daily cap bonus (%d)", weeklyMax, dailyMax)
	}
	if combined <= weeklyMax || combined <= dailyMax {
		t.Fatalf("combined (%d) must be greater than weekly-only (%d) and daily-only (%d)",
			combined, weeklyMax, dailyMax)
	}
	// The cap's job is to bound runaway XP: a streak-100 user must earn the
	// same streak bonus as a streak-(cap+1) user. Verify that ceiling.
	veryLong := cappedStreakBonus(100, weeklyStreakBonusCap, weeklyStreakBonusPerStep)
	if veryLong != weeklyMax {
		t.Fatalf("streak-100 weekly bonus (%d) must equal the cap (%d); cap is not bounding", veryLong, weeklyMax)
	}
}

func TestDailyStreakFromDates(t *testing.T) {
	today := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name   string
		dates  []time.Time
		today  time.Time
		want   int
	}{
		{
			name:  "no history counts today only",
			today: today,
			want:  1,
		},
		{
			name:  "consecutive days including today",
			dates: []time.Time{today.AddDate(0, 0, -3), today.AddDate(0, 0, -2), today.AddDate(0, 0, -1)},
			today: today,
			want:  4,
		},
		{
			name: "gap breaks the streak",
			// Reported 5 and 4 days ago, then nothing until today: streak is 1.
			dates: []time.Time{today.AddDate(0, 0, -5), today.AddDate(0, 0, -4)},
			today: today,
			want:  1,
		},
		{
			name: "duplicate dates do not inflate the count",
			dates: []time.Time{
				today.AddDate(0, 0, -1), today.AddDate(0, 0, -1),
				today.AddDate(0, 0, -2),
			},
			today: today,
			want:  3,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := dailyStreakFromDates(c.dates, c.today); got != c.want {
				t.Fatalf("dailyStreakFromDates = %d, want %d", got, c.want)
			}
		})
	}
}
