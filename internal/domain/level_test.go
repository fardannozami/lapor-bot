package domain

import "testing"

func TestGetLevel_UsesLongTermLifetimeThresholds(t *testing.T) {
	tests := []struct {
		points int
		name   string
	}{
		{points: 0, name: "E-Tier Hunter"},
		{points: 1149, name: "E-Tier Hunter"},
		{points: 1150, name: "D-Tier Hunter"},
		{points: 4674, name: "D-Tier Hunter"},
		{points: 4675, name: "C-Tier Hunter"},
		{points: 11824, name: "C-Tier Hunter"},
		{points: 11825, name: "B-Tier Hunter"},
		{points: 23849, name: "B-Tier Hunter"},
		{points: 23850, name: "A-Tier Hunter"},
		{points: 41999, name: "A-Tier Hunter"},
		{points: 42000, name: "S-Tier Hunter"},
	}

	for _, tt := range tests {
		if got := GetLevel(tt.points).Name; got != tt.name {
			t.Fatalf("GetLevel(%d) = %q, want %q", tt.points, got, tt.name)
		}
	}
}

func TestGetNextLevel_ReturnsNilAtMaxLifetimeTier(t *testing.T) {
	next, remaining := GetNextLevel(42000)
	if next != nil || remaining != 0 {
		t.Fatalf("GetNextLevel(42000) = %+v, %d; want nil, 0", next, remaining)
	}
}

func TestGetNextSeasonRank_UsesSeasonThresholds(t *testing.T) {
	next, remaining := GetNextSeasonRank(915)
	if next == nil {
		t.Fatal("expected next season rank")
	}
	if next.Name != "Legend" || remaining != 335 {
		t.Fatalf("next rank = %s, remaining = %d; want Legend, 335", next.Name, remaining)
	}
}
