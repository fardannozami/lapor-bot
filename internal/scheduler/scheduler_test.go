package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestDailyScheduleNext(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")

	// Mock "now" at 14:00 WIB on June 14
	now := time.Date(2026, time.June, 14, 14, 0, 0, 0, loc)

	tests := []struct {
		name     string
		schedule DailySchedule
		expected time.Time
	}{
		{
			name:     "later today",
			schedule: DailySchedule{Hour: 15, Minute: 9, Loc: loc},
			expected: time.Date(2026, time.June, 14, 15, 9, 0, 0, loc),
		},
		{
			name:     "already past, tomorrow",
			schedule: DailySchedule{Hour: 10, Minute: 0, Loc: loc},
			expected: time.Date(2026, time.June, 15, 10, 0, 0, 0, loc),
		},
		{
			name:     "exact same minute is tomorrow",
			schedule: DailySchedule{Hour: 14, Minute: 0, Loc: loc},
			expected: time.Date(2026, time.June, 15, 14, 0, 0, 0, loc),
		},
		{
			name:     "near midnight",
			schedule: DailySchedule{Hour: 23, Minute: 58, Loc: loc},
			expected: time.Date(2026, time.June, 14, 23, 58, 0, 0, loc),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.schedule.Next(now)
			if !got.Equal(tt.expected) {
				t.Errorf("Next() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWeeklyScheduleNext(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")

	// Mock "now" at Monday 06:00 WIB on June 15, 2026
	nowMondayMorning := time.Date(2026, time.June, 15, 6, 0, 0, 0, loc)

	tests := []struct {
		name     string
		now      time.Time
		schedule WeeklySchedule
		expected time.Time
	}{
		{
			name:     "Monday 06:00 to Monday 07:00 (same day, later)",
			now:      nowMondayMorning,
			schedule: WeeklySchedule{Weekday: time.Monday, Hour: 7, Minute: 0, Loc: loc},
			expected: time.Date(2026, time.June, 15, 7, 0, 0, 0, loc),
		},
		{
			name:     "Monday 07:00 to Monday 07:00 (exact time, rolls to next week)",
			now:      time.Date(2026, time.June, 15, 7, 0, 0, 0, loc),
			schedule: WeeklySchedule{Weekday: time.Monday, Hour: 7, Minute: 0, Loc: loc},
			expected: time.Date(2026, time.June, 22, 7, 0, 0, 0, loc),
		},
		{
			name:     "Monday 08:00 to Monday 07:00 (same day, past, rolls to next week)",
			now:      time.Date(2026, time.June, 15, 8, 0, 0, 0, loc),
			schedule: WeeklySchedule{Weekday: time.Monday, Hour: 7, Minute: 0, Loc: loc},
			expected: time.Date(2026, time.June, 22, 7, 0, 0, 0, loc),
		},
		{
			name:     "Sunday 23:59 to Monday 07:00 (next day)",
			now:      time.Date(2026, time.June, 14, 23, 59, 0, 0, loc),
			schedule: WeeklySchedule{Weekday: time.Monday, Hour: 7, Minute: 0, Loc: loc},
			expected: time.Date(2026, time.June, 15, 7, 0, 0, 0, loc),
		},
		{
			name:     "Tuesday 12:00 to Monday 07:00 (rolls to next week)",
			now:      time.Date(2026, time.June, 16, 12, 0, 0, 0, loc),
			schedule: WeeklySchedule{Weekday: time.Monday, Hour: 7, Minute: 0, Loc: loc},
			expected: time.Date(2026, time.June, 22, 7, 0, 0, 0, loc),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.schedule.Next(tt.now)
			if !got.Equal(tt.expected) {
				t.Errorf("Next() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseDaily(t *testing.T) {
	loc := time.UTC

	tests := []struct {
		input   string
		wantErr bool
		wantH   int
		wantM   int
	}{
		{"15:09", false, 15, 9},
		{"00:00", false, 0, 0},
		{"23:59", false, 23, 59},
		{"24:00", true, 0, 0},
		{"15:60", true, 0, 0},
		{"not-a-time", true, 0, 0},
		{"", true, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDaily(tt.input, loc)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got.Hour != tt.wantH || got.Minute != tt.wantM {
				t.Errorf("got %d:%d, want %d:%d", got.Hour, got.Minute, tt.wantH, tt.wantM)
			}
		})
	}
}

func TestSchedulerRecovery(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var fired int32
	sched := NewScheduler(ctx)

	sched.AddJob(&Job{
		Name:    "recovery-test",
		Freq:    &DailySchedule{Hour: 9, Minute: 0, Loc: loc},
		Recover: true,
		Fn: func(ctx context.Context) error {
			atomic.AddInt32(&fired, 1)
			return nil
		},
	})

	sched.Start()

	// Wait for the job to fire (recovery) and settle
	time.Sleep(200 * time.Millisecond)
	sched.Stop()

	count := atomic.LoadInt32(&fired)
	if count == 0 {
		t.Error("Recovery job should have fired at least once")
	}
}

func TestSchedulerCancellation(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	ctx, cancel := context.WithCancel(context.Background())

	var fired int32
	sched := NewScheduler(ctx)

	// Schedule far in the future, no recovery
	sched.AddJob(&Job{
		Name:    "no-recovery-future",
		Freq:    &DailySchedule{Hour: 23, Minute: 59, Loc: loc},
		Recover: false,
		Fn: func(ctx context.Context) error {
			atomic.AddInt32(&fired, 1)
			return nil
		},
	})

	sched.Start()

	// Immediately cancel
	time.Sleep(50 * time.Millisecond)
	cancel()
	sched.Stop()

	count := atomic.LoadInt32(&fired)
	if count != 0 {
		t.Errorf("Job should not have fired after cancellation, got %d", count)
	}
}

func TestSchedulerNoRecoveryWhenClose(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var fired int32
	sched := NewScheduler(ctx)

	// Schedule 1 minute from now, no recovery
	now := time.Now().In(loc)
	futureMinute := (now.Minute() + 1) % 60
	futureHour := now.Hour()
	if futureMinute == 0 {
		futureHour = (futureHour + 1) % 24
	}

	sched.AddJob(&Job{
		Name:    "future-no-recovery",
		Freq:    &DailySchedule{Hour: futureHour, Minute: futureMinute, Loc: loc},
		Recover: false,
		Fn: func(ctx context.Context) error {
			atomic.AddInt32(&fired, 1)
			return nil
		},
	})

	sched.Start()

	// Should not fire (it's scheduled for next cycle, and Recover is false)
	time.Sleep(200 * time.Millisecond)
	sched.Stop()

	count := atomic.LoadInt32(&fired)
	if count != 0 {
		t.Errorf("Job with no recovery should not have fired, got %d", count)
	}
}
