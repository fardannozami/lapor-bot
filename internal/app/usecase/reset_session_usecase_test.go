package usecase

import (
	"strings"
	"testing"
	"time"
)

func TestGetCurrentSessionInfo_CurrentLaunchSeasonStartsAtOne(t *testing.T) {
	loc := time.FixedZone("WIB", 7*3600)

	tests := []struct {
		name      string
		now       time.Time
		season    int
		startDate string
	}{
		{
			name:      "launch period is season one",
			now:       time.Date(2026, time.June, 5, 12, 0, 0, 0, loc),
			season:    1,
			startDate: "2026-05-01",
		},
		{
			name:      "next reset increments to season two",
			now:       time.Date(2026, time.September, 1, 0, 0, 0, 0, loc),
			season:    2,
			startDate: "2026-09-01",
		},
		{
			name:      "following year keeps incrementing",
			now:       time.Date(2027, time.January, 1, 0, 0, 0, 0, loc),
			season:    3,
			startDate: "2027-01-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			season, start := GetCurrentSessionInfo(tt.now)
			if season != tt.season {
				t.Fatalf("expected season %d, got %d", tt.season, season)
			}
			if got := start.Format(time.DateOnly); got != tt.startDate {
				t.Fatalf("expected start %s, got %s", tt.startDate, got)
			}
		})
	}
}

func TestBuildSeasonResetAnnouncement_DoesNotMentionSeasonZero(t *testing.T) {
	announcement := buildSeasonResetAnnouncement(1)
	if strings.Contains(announcement, "Season 0") {
		t.Fatalf("season 1 announcement must not mention Season 0: %s", announcement)
	}
	if !strings.Contains(announcement, "SEASON 1 TELAH DIMULAI") {
		t.Fatalf("expected season 1 start announcement, got %s", announcement)
	}
}
