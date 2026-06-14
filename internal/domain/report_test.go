package domain

import (
	"testing"
	"time"
)

func TestGetToday(t *testing.T) {
	// 2026-03-14 00:15:00 UTC
	t1 := time.Date(2026, 3, 14, 0, 15, 0, 0, time.UTC)
	// With 30m offset, 00:15 should be "yesterday" (2026-03-13)
	expected1 := time.Date(2026, 3, 13, 0, 0, 0, 0, time.UTC)
	got1 := GetToday(t1)
	if !got1.Equal(expected1) {
		t.Errorf("For 00:15, expected %v, got %v", expected1, got1)
	}

	// 2026-03-14 00:31:00 UTC
	t2 := time.Date(2026, 3, 14, 0, 31, 0, 0, time.UTC)
	// With 30m offset, 00:31 should be "today" (2026-03-14)
	expected2 := time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC)
	got2 := GetToday(t2)
	if !got2.Equal(expected2) {
		t.Errorf("For 00:31, expected %v, got %v", expected2, got2)
	}

	// 2026-03-14 23:59:00 UTC
	t3 := time.Date(2026, 3, 14, 23, 59, 0, 0, time.UTC)
	// Should still be today
	expected3 := time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC)
	got3 := GetToday(t3)
	if !got3.Equal(expected3) {
		t.Errorf("For 23:59, expected %v, got %v", expected3, got3)
	}
}

func TestGetStartOfISOWeekStrict_UsesMondayMidnightBoundary(t *testing.T) {
	tests := []struct {
		name string
		now  time.Time
		want time.Time
	}{
		{
			name: "Monday midnight starts new week",
			now:  time.Date(2026, time.June, 15, 0, 0, 0, 0, time.UTC),
			want: time.Date(2026, time.June, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Monday within report cutoff still starts new week",
			now:  time.Date(2026, time.June, 15, 0, 29, 0, 0, time.UTC),
			want: time.Date(2026, time.June, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Sunday belongs to previous Monday",
			now:  time.Date(2026, time.June, 14, 23, 59, 0, 0, time.UTC),
			want: time.Date(2026, time.June, 8, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetStartOfISOWeekStrict(tt.now)
			if !got.Equal(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
