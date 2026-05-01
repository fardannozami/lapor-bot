package domain

import (
	"context"
	"time"
)

type Report struct {
	UserID          string    `json:"user_id" db:"user_id"`
	Name            string    `json:"name" db:"name"`
	Streak          int       `json:"streak" db:"streak"`
	ActivityCount   int       `json:"activity_count" db:"activity_count"`
	LastReportDate  time.Time `json:"last_report_date" db:"last_report_date"`
	MaxStreak       int       `json:"max_streak" db:"max_streak"`
	TotalPoints     int       `json:"total_points" db:"total_points"`
	Achievements    string    `json:"achievements" db:"achievements"`
	ComebackStreak  int       `json:"comeback_streak" db:"comeback_streak"`
	InactiveDays    int       `json:"inactive_days" db:"inactive_days"`
	CenturionCycles int       `json:"centurion_cycles" db:"centurion_cycles"`
}

type ActivityLeaderboardEntry struct {
	UserID        string
	Name          string
	ActivityCount int
}

// ReportCutoffOffset is the spare time allowed for late-night reporting.
// For example, if offset is 30m, 00:29 AM is still considered "yesterday".
const ReportCutoffOffset = 30 * time.Minute

// GetToday returns the normalized "today" (midnight) based on the cutoff offset.
func GetToday(t time.Time) time.Time {
	// Shift time back by offset then truncate to date
	shifted := t.Add(-ReportCutoffOffset)
	return time.Date(shifted.Year(), shifted.Month(), shifted.Day(), 0, 0, 0, 0, time.UTC)
}

// GetStartOfISOWeek returns the Monday of the ISO week containing t.
func GetStartOfISOWeek(t time.Time) time.Time {
	t = GetToday(t)
	// ISO week starts on Monday. Go's Weekday() returns 0 for Sunday, 1 for Monday, etc.
	// We want to shift back to the most recent Monday.
	weekday := int(t.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}
	daysToSubtract := weekday - 1
	return t.AddDate(0, 0, -daysToSubtract)
}

// GetStartOfSundayWeek returns the Sunday that starts the week containing t.
func GetStartOfSundayWeek(t time.Time) time.Time {
	t = GetToday(t)
	return t.AddDate(0, 0, -int(t.Weekday()))
}

type ReportRepository interface {
	GetReport(ctx context.Context, userID string) (*Report, error)
	UpsertReport(ctx context.Context, report *Report) error
	UpsertReportWithActivity(ctx context.Context, report *Report, activityDate time.Time) error
	GetAllReports(ctx context.Context) ([]*Report, error)
	LogActivity(ctx context.Context, userID string, activityDate time.Time) error
	GetActivityCountsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]ActivityLeaderboardEntry, error)
	GetInactiveUsers(ctx context.Context, days int) ([]*Report, error)
	ResetAllReports(ctx context.Context) error
	InitTable(ctx context.Context) error
	ResolveLIDToPhone(ctx context.Context, lid string) string

	// Strava Integration
	UpsertStravaAccount(ctx context.Context, account *StravaAccount) error
	GetStravaAccountByAthleteID(ctx context.Context, athleteID int64) (*StravaAccount, error)
	GetStravaAccountByUserID(ctx context.Context, userID string) (*StravaAccount, error)
}
