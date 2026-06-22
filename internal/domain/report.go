package domain

import (
	"context"
	"errors"
	"strings"
	"time"
)

var ErrActiveGoalExists = errors.New("active goal exists")

type Report struct {
	UserID                string    `json:"user_id" db:"user_id"`
	Name                  string    `json:"name" db:"name"`
	JobClass              string    `json:"job_class" db:"job_class"`
	Streak                int       `json:"streak" db:"streak"`
	ActivityCount         int       `json:"activity_count" db:"activity_count"`
	LastReportDate        time.Time `json:"last_report_date" db:"last_report_date"`
	MaxStreak             int       `json:"max_streak" db:"max_streak"`
	TotalPoints           int       `json:"total_points" db:"total_points"`
	Level                 int       `json:"level" db:"level"`
	Achievements          string    `json:"achievements" db:"achievements"`
	ComebackStreak        int       `json:"comeback_streak" db:"comeback_streak"`
	InactiveDays          int       `json:"inactive_days" db:"inactive_days"`
	CenturionCycles       int       `json:"centurion_cycles" db:"centurion_cycles"`
	SeasonalPoints        int       `json:"seasonal_points" db:"seasonal_points"`
	SeasonalActivityCount int       `json:"seasonal_activity_count" db:"seasonal_activity_count"`
	SeasonalMaxStreak     int       `json:"seasonal_max_streak" db:"seasonal_max_streak"`
	SeasonalAchievements  string    `json:"seasonal_achievements" db:"seasonal_achievements"`
	StreakFreezes         int       `json:"streak_freezes" db:"streak_freezes"`
	GoalsCompleted        int       `json:"goals_completed" db:"goals_completed"`
	TotalSideQuests       int       `json:"total_side_quests" db:"total_side_quests"`
	SeasonalSideQuests    int       `json:"seasonal_side_quests" db:"seasonal_side_quests"`
	Str                   int       `json:"str" db:"str"`
	Sta                   int       `json:"sta" db:"sta"`
	Agi                   int       `json:"agi" db:"agi"`
	Vit                   int       `json:"vit" db:"vit"`
}

type ActivityLeaderboardEntry struct {
	UserID        string
	Name          string
	ActivityCount int
}

type WeeklyGoal struct {
	UserID      string
	TargetDays  int
	Activity    string
	StartAt     time.Time
	EndAt       time.Time
	CreatedAt   time.Time
	CompletedAt *time.Time
}

type GoalActivity struct {
	Date     time.Time
	Activity string
}

type AttributeType string

const (
	AttrStr AttributeType = "STR"
	AttrSta AttributeType = "STA"
	AttrAgi AttributeType = "AGI"
	AttrVit AttributeType = "VIT"
)

// MinAttributeValue is the floor for any displayed attribute.
// A hunter starts at 1 (not 0) so the dashboard always shows a positive baseline.
const MinAttributeValue = 1

// ClampedAttribute returns v bounded below by MinAttributeValue.
// Use this whenever an attribute is displayed or serialized for the UI
// so that the "start from 1" invariant is enforced in a single place.
func ClampedAttribute(v int) int {
	if v < MinAttributeValue {
		return MinAttributeValue
	}
	return v
}

// DetermineAttributes parses an activity text to find matching RPG attributes.
func DetermineAttributes(text string) []AttributeType {
	text = strings.ToLower(text)
	var attrs []AttributeType

	if strings.Contains(text, "beban") || strings.Contains(text, "weight") || strings.Contains(text, "strength") || strings.Contains(text, "gym") || strings.Contains(text, "angkat") || strings.Contains(text, "powerlifting") || strings.Contains(text, "push") || strings.Contains(text, "pull") || strings.Contains(text, "leg") {
		attrs = append(attrs, AttrStr)
	}
	if strings.Contains(text, "lari") || strings.Contains(text, "run") || strings.Contains(text, "running") || strings.Contains(text, "sepeda") || strings.Contains(text, "cycle") || strings.Contains(text, "hiit") || strings.Contains(text, "kardio") || strings.Contains(text, "cardio") || strings.Contains(text, "renang") || strings.Contains(text, "swim") {
		attrs = append(attrs, AttrSta)
	}
	if strings.Contains(text, "bola") || strings.Contains(text, "futsal") || strings.Contains(text, "basket") || strings.Contains(text, "bulutangkis") || strings.Contains(text, "tenis") || strings.Contains(text, "sprint") || strings.Contains(text, "muaythai") || strings.Contains(text, "boxing") || strings.Contains(text, "calisthenics") || strings.Contains(text, "padel") || strings.Contains(text, "padle") {
		attrs = append(attrs, AttrAgi)
	}
	if strings.Contains(text, "yoga") || strings.Contains(text, "pilates") || strings.Contains(text, "stretching") || strings.Contains(text, "recovery") || strings.Contains(text, "jalan") || strings.Contains(text, "walk") || strings.Contains(text, "meditasi") {
		attrs = append(attrs, AttrVit)
	}

	if len(attrs) == 0 {
		// Default to VIT (General Health/Vitality) if no specific keywords match
		attrs = append(attrs, AttrVit)
	}

	return attrs
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

// GetStartOfISOWeekStrict returns the Monday of the ISO week containing t
// without applying the daily report cutoff. Use this for weekly windows that
// must roll over exactly at Monday 00:00.
func GetStartOfISOWeekStrict(t time.Time) time.Time {
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
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

	DeleteActivityLog(ctx context.Context, userID string, activityDate time.Time) error
	DeleteLatestActivityLog(ctx context.Context, userID string, activityDate time.Time) (int, error)
	GetUserActivityDates(ctx context.Context, userID string) ([]time.Time, error)
	DeleteReport(ctx context.Context, userID string) error

	GetDailyActivityCount(ctx context.Context, userID string, date time.Time) (int, error)

	SetGoal(ctx context.Context, goal *WeeklyGoal) error
	GetActiveGoal(ctx context.Context, userID string, now time.Time) (*WeeklyGoal, error)
	DeleteActiveGoal(ctx context.Context, userID string, now time.Time) error
	DeleteExpiredGoals(ctx context.Context, now time.Time) (int64, error)
	GetGoalActivities(ctx context.Context, userID string, startDate, endDate time.Time) ([]GoalActivity, error)
	RecordGoalActivity(ctx context.Context, userID string, activityDate time.Time, activityText string) (bool, error)

	// Strava Integration
	UpsertStravaAccount(ctx context.Context, account *StravaAccount) error
	GetStravaAccountByAthleteID(ctx context.Context, athleteID int64) (*StravaAccount, error)
	GetStravaAccountByUserID(ctx context.Context, userID string) (*StravaAccount, error)

	// Daily Quests
	SaveDailyQuest(ctx context.Context, userID, questDate, tasksJSON string) error
	GetDailyQuest(ctx context.Context, userID, questDate string) (tasksJSON string, err error)

	// Job Classes
	GetAllJobClasses(ctx context.Context) ([]JobClass, error)
	GetJobClass(ctx context.Context, id string) (*JobClass, error)
}
