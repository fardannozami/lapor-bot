package domain

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode"
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

var attributeKeywordSets = []struct {
	attr     AttributeType
	keywords []string
}{
	{AttrStr, []string{
		"beban", "weight", "weightlifting", "strength", "gym", "angkat", "powerlifting",
		"push up", "pushup", "pull up", "pullup", "squat", "chair squat", "sit to stand",
		"push day", "pull day", "push workout", "pull workout", "deadlift", "bench press",
		"row", "leg day", "plank", "desk plank", "lunges", "lunge", "glute bridge",
		"calf raises", "calf raise", "jinjit", "reverse crunch", "wall sit", "dead bug", "core",
	}},
	{AttrSta, []string{
		"lari", "berlari", "run", "running", "jogging", "marathon", "sepeda", "bersepeda",
		"cycle", "cycling", "bike", "kardio", "cardio", "renang", "berenang", "swim",
		"swimming", "hiking", "trekking", "jalan jauh", "stairs", "tangga", "naik turun tangga",
		"jumping jacks", "jumpingjacks", "step up", "stepup", "skipping", "rope jump",
		"mountain climber", "mountainclimber",
	}},
	{AttrAgi, []string{
		"bola", "sepak bola", "soccer", "football", "futsal", "basket", "basketball",
		"bulutangkis", "bulu tangkis", "badminton", "tenis", "tennis", "padel", "padle",
		"pickleball", "sprint", "hiit", "tabata", "muay thai", "muaythai", "boxing",
		"tinju", "shadow boxing", "shadowboxing", "calisthenics", "voli", "volleyball",
		"high knees", "highknees", "mountain climber", "mountainclimber", "lateral shuffle",
		"lateralshuffle", "squat jump", "jump squat", "ladder drill", "shuttle run",
	}},
	{AttrVit, []string{
		"yoga", "pilates", "stretching", "stretch", "peregangan", "mobility", "recovery",
		"walk", "walking", "jalan kaki", "jalan santai", "meditasi", "meditation", "meditate",
		"napas", "breathing", "deep breathing", "pemulihan", "mobility flow", "bird dog",
		"balance", "keseimbangan", "shoulder mobility", "neck mobility",
	}},
}

// DetermineAttributes parses an activity text to find matching RPG attributes.
// Matching is table-driven and boundary-aware so unrelated words like
// "legacy" or "rundown" do not accidentally grant STR/STA.
// Returns an empty slice when no keywords match — callers should handle
// the fallback via ResolveReportAttributes instead of hardcoding a default.
func DetermineAttributes(text string) []AttributeType {
	normalized := normalizeActivityText(text)
	var attrs []AttributeType

	for _, set := range attributeKeywordSets {
		if containsAnyActivityKeyword(normalized, set.keywords) {
			attrs = append(attrs, set.attr)
		}
	}

	return attrs
}

// ResolveReportAttributes determines which attributes are activated by an
// activity, falling back to the user's job primary attribute when no keywords
// match. Returns (attributes, ok). When ok is false the caller should prompt
// the user to pick a job first.
func ResolveReportAttributes(text string, jobClass string) ([]AttributeType, bool) {
	attrs := DetermineAttributes(text)
	if len(attrs) > 0 {
		return attrs, true
	}

	// No keywords matched — fall back to job primary attribute.
	if jobClass != "" {
		primary := JobClassPrimaryAttribute(jobClass)
		if primary != "" {
			return []AttributeType{primary}, true
		}
		// Mage has no single primary attribute — give all attributes.
		return []AttributeType{AttrStr, AttrSta, AttrAgi, AttrVit}, true
	}

	// No job selected — user needs to pick one.
	return nil, false
}

// SelectReportAttribute picks the single attribute that should actually be
// rewarded for a report, given the activity-matched attributes and the user's
// job. This keeps the attribute grant FAIR and CONSISTENT across reports:
//
//   - Every report grants the same total attribute budget (one attribute),
//     regardless of how many categories the activity text happens to match.
//     Without this, a mixed session ("gym lalu lari dan stretching") would
//     grant +1 to each of STR/STA/VIT while a focused session grants only +1,
//     making some workouts worth several times the attribute points of others.
//   - The activity directs the reward (a run still boosts STA, not STR), so
//     the grant reflects the sport that was reported.
//   - The job acts as a tie-breaker when several attributes match (the hunter's
//     specialty wins among the matches), and as the fallback when nothing
//     matches — so the user's job always shapes the gain.
//   - Jobs with no single primary attribute (Mage) distribute the gain across
//     the matched attributes using a stable per-report seed, instead of always
//     defaulting to the first match (which would bias Mage gains toward STR).
//
// seed must be stable for a given report slot (e.g. user+date+slot+activity)
// so the same report context always picks the same attribute, while different
// reports spread Mage's gains evenly across its candidate attributes.
//
// Returns the chosen attribute, or "" if attrs is empty.
func SelectReportAttribute(attrs []AttributeType, jobClass, seed string) AttributeType {
	if len(attrs) == 0 {
		return ""
	}
	if primary := JobClassPrimaryAttribute(jobClass); primary != "" {
		for _, a := range attrs {
			if a == primary {
				return a
			}
		}
		// Job's specialty is not among the matches — keep a deterministic
		// fallback (first matched) so the result is stable per input.
		return attrs[0]
	}
	// No single primary (Mage, or an unknown job): distribute fairly across
	// the matched attributes using the stable per-report seed.
	if len(attrs) > 1 {
		return attrs[stableAttributeIndex(seed, len(attrs))]
	}
	return attrs[0]
}

// stableAttributeIndex maps a seed string to a deterministic index in [0, n).
// Same seed always yields the same index (reproducible tests), and as the seed
// varies across report slots the indices spread uniformly over n.
func stableAttributeIndex(seed string, n int) int {
	if n <= 0 {
		return 0
	}
	const fnvOffset uint32 = 2166136261
	const fnvPrime uint32 = 16777619
	h := fnvOffset
	for i := 0; i < len(seed); i++ {
		h ^= uint32(seed[i])
		h *= fnvPrime
	}
	return int(h % uint32(n))
}

func normalizeActivityText(text string) string {
	text = strings.ToLower(text)
	var b strings.Builder
	lastSpace := true

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastSpace = false
			continue
		}
		if !lastSpace {
			b.WriteByte(' ')
			lastSpace = true
		}
	}

	return strings.TrimSpace(b.String())
}

func containsAnyActivityKeyword(normalizedText string, keywords []string) bool {
	for _, keyword := range keywords {
		if containsActivityKeyword(normalizedText, keyword) {
			return true
		}
	}
	return false
}

func containsActivityKeyword(normalizedText, keyword string) bool {
	normalizedKeyword := normalizeActivityText(keyword)
	if normalizedText == "" || normalizedKeyword == "" {
		return false
	}
	return strings.Contains(" "+normalizedText+" ", " "+normalizedKeyword+" ")
}

// ReportCutoffOffset is the spare time allowed for late-night reporting.
// For example, if offset is 30m, 00:29 AM is still considered "yesterday".
const ReportCutoffOffset = 30 * time.Minute

const (
	ActivityKindRegularReport = "regular_report"
	ActivityKindSideQuest     = "sidequest"
)

// ReportActivityEvent is the immutable fact written for every accepted /lapor
// or /lapor sidequest command. Aggregates can change shape over time, but this
// event preserves the awarded deltas for Season 2+ leaderboard rebuilds.
type ReportActivityEvent struct {
	EventID             string
	UserID              string
	SeasonNumber        int
	Kind                string
	ActivityDate        time.Time
	OccurredAt          time.Time
	PointsDelta         int
	RegularCountDelta   int
	SideQuestCountDelta int
	RuleVersion         int
	Source              string
	ActivityText        string
	MetadataJSON        string
}

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
	DeleteActivityLogByKind(ctx context.Context, userID string, activityDate time.Time, kind string) error
	DeleteLatestActivityLogByKind(ctx context.Context, userID string, activityDate time.Time, kind string) (int, error)
	GetUserActivityDates(ctx context.Context, userID string) ([]time.Time, error)
	GetUserActivityDatesByKind(ctx context.Context, userID string, kind string) ([]time.Time, error)
	DeleteReport(ctx context.Context, userID string) error

	GetDailyActivityCount(ctx context.Context, userID string, date time.Time) (int, error)
	GetDailyActivityCountByKind(ctx context.Context, userID string, date time.Time, kind string) (int, error)

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
