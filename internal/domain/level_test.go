package domain

import "testing"

func TestGetLevel_UsesLongTermLifetimeThresholds(t *testing.T) {
	tests := []struct {
		points int
		name   string
	}{
		{points: 0, name: "Newbie"},
		{points: 50, name: "Fighter"},
		{points: 199, name: "Fighter"},
		{points: 200, name: "Warrior"},
		{points: 499, name: "Warrior"},
		{points: 500, name: "Champion"},
		{points: 999, name: "Champion"},
		{points: 1000, name: "Legend"},
		{points: 2499, name: "Legend"},
		{points: 2500, name: "Immortal"},
		{points: 3500, name: "Immortal"},
		{points: 4999, name: "Immortal"},
		{points: 5000, name: "Titan"},
		{points: 9999, name: "Titan"},
		{points: 10000, name: "God"},
		{points: 19999, name: "God"},
		{points: 20000, name: "Cosmic"},
	}

	for _, tt := range tests {
		if got := GetLevel(tt.points).Name; got != tt.name {
			t.Fatalf("GetLevel(%d) = %q, want %q", tt.points, got, tt.name)
		}
	}
}

func TestGetNextLevel_ReturnsNilAtMaxLifetimeTier(t *testing.T) {
	next, remaining := GetNextLevel(20000)
	if next != nil || remaining != 0 {
		t.Fatalf("GetNextLevel(20000) = %+v, %d; want nil, 0", next, remaining)
	}
}

func TestGetNextSeasonRank_UsesSeasonThresholds(t *testing.T) {
	next, remaining := GetNextSeasonRank(915)
	if next == nil {
		t.Fatal("expected next season rank")
	}
	if next.Name != "S-Rank Hunter" || remaining != 85 {
		t.Fatalf("next rank = %s, remaining = %d; want S-Rank Hunter, 85", next.Name, remaining)
	}
}
