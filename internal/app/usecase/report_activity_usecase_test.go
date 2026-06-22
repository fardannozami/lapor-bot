package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/app/usecase"
	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type mockRepo struct {
	domain.ReportRepository
	reports            map[string]*domain.Report
	activityCounts     []domain.ActivityLeaderboardEntry
	dailyCounts        map[string]int
	activityKindCounts map[string]map[string]int
}

func (m *mockRepo) GetReport(ctx context.Context, userID string) (*domain.Report, error) {
	return m.reports[userID], nil
}

func (m *mockRepo) UpsertReport(ctx context.Context, report *domain.Report) error {
	m.reports[report.UserID] = report
	return nil
}

func (m *mockRepo) UpsertReportWithActivity(ctx context.Context, report *domain.Report, activityDate time.Time) error {
	return m.UpsertReportWithActivityKind(ctx, report, activityDate, domain.ActivityKindRegularReport)
}

func (m *mockRepo) UpsertReportWithActivityKind(ctx context.Context, report *domain.Report, activityDate time.Time, kind string) error {
	m.reports[report.UserID] = report
	key := report.UserID + "|" + activityDate.Format(time.DateOnly)
	if m.dailyCounts == nil {
		m.dailyCounts = make(map[string]int)
	}
	m.dailyCounts[key]++
	if m.activityKindCounts == nil {
		m.activityKindCounts = make(map[string]map[string]int)
	}
	if m.activityKindCounts[key] == nil {
		m.activityKindCounts[key] = make(map[string]int)
	}
	m.activityKindCounts[key][kind]++
	return nil
}

func (m *mockRepo) GetAllReports(ctx context.Context) ([]*domain.Report, error) {
	var result []*domain.Report
	for _, r := range m.reports {
		result = append(result, r)
	}
	return result, nil
}

func (m *mockRepo) LogActivity(ctx context.Context, userID string, activityDate time.Time) error {
	return nil
}

func (m *mockRepo) GetActivityCountsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]domain.ActivityLeaderboardEntry, error) {
	return m.activityCounts, nil
}

func (m *mockRepo) ResolveLIDToPhone(ctx context.Context, lid string) string {
	return lid
}

func (m *mockRepo) InitTable(ctx context.Context) error {
	return nil
}

func (m *mockRepo) GetInactiveUsers(ctx context.Context, days int) ([]*domain.Report, error) {
	return nil, nil
}

func (m *mockRepo) ResetAllReports(ctx context.Context) error {
	m.reports = make(map[string]*domain.Report)
	return nil
}

func (m *mockRepo) UpsertStravaAccount(ctx context.Context, account *domain.StravaAccount) error {
	return nil
}

func (m *mockRepo) GetStravaAccountByAthleteID(ctx context.Context, athleteID int64) (*domain.StravaAccount, error) {
	return nil, nil
}

func (m *mockRepo) GetStravaAccountByUserID(ctx context.Context, userID string) (*domain.StravaAccount, error) {
	return nil, nil
}

func (m *mockRepo) GetUserActivityDates(ctx context.Context, userID string) ([]time.Time, error) {
	return nil, nil
}

func (m *mockRepo) DeleteActivityLog(ctx context.Context, userID string, activityDate time.Time) error {
	return nil
}

func (m *mockRepo) DeleteLatestActivityLog(ctx context.Context, userID string, activityDate time.Time) (int, error) {
	return 0, nil
}

func (m *mockRepo) DeleteReport(ctx context.Context, userID string) error {
	return nil
}

func (m *mockRepo) GetDailyActivityCount(ctx context.Context, userID string, date time.Time) (int, error) {
	key := userID + "|" + date.Format(time.DateOnly)
	if m.dailyCounts == nil {
		return 0, nil
	}
	return m.dailyCounts[key], nil
}

func (m *mockRepo) GetDailyActivityCountByKind(ctx context.Context, userID string, date time.Time, kind string) (int, error) {
	key := userID + "|" + date.Format(time.DateOnly)
	if m.activityKindCounts != nil && m.activityKindCounts[key] != nil {
		return m.activityKindCounts[key][kind], nil
	}
	if kind == domain.ActivityKindRegularReport {
		return m.GetDailyActivityCount(ctx, userID, date)
	}
	return 0, nil
}

func (m *mockRepo) SetGoal(ctx context.Context, goal *domain.WeeklyGoal) error {
	return nil
}

func (m *mockRepo) GetActiveGoal(ctx context.Context, userID string, now time.Time) (*domain.WeeklyGoal, error) {
	return nil, nil
}

func (m *mockRepo) DeleteActiveGoal(ctx context.Context, userID string, now time.Time) error {
	return nil
}

func (m *mockRepo) DeleteExpiredGoals(ctx context.Context, now time.Time) (int64, error) {
	return 0, nil
}

func (m *mockRepo) GetGoalActivities(ctx context.Context, userID string, startDate, endDate time.Time) ([]domain.GoalActivity, error) {
	return nil, nil
}

func (m *mockRepo) RecordGoalActivity(ctx context.Context, userID string, activityDate time.Time, activityText string) (bool, error) {
	return false, nil
}

// =============================================================================
// STREAK LOGIC TESTS
// =============================================================================
//
// Streak Rules:
// 1. First report ever → Streak = 1, ActivityCount = 1
// 2. Report consecutive day (yesterday) → Streak++, ActivityCount++
// 3. Miss a day then report → Streak resets to 1, ActivityCount++ (total preserved)
// 4. Report same day twice → Rejected ("sudah laporan hari ini")
//
// ActivityCount = Total days ever reported (never resets)
// Streak = Consecutive days in current streak (resets on missed day)
//
// =============================================================================

func TestStreak_FirstReport(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report), dailyCounts: make(map[string]int)}
	uc := usecase.NewReportActivityUsecase(repo)
	ctx := context.Background()

	// First ever report
	msg, err := uc.Execute(ctx, "user1", "Alice", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	r := repo.reports["user1"]
	if r.Streak != 1 {
		t.Errorf("First report: expected Streak=1, got %d", r.Streak)
	}
	if r.ActivityCount != 1 {
		t.Errorf("First report: expected ActivityCount=1, got %d", r.ActivityCount)
	}
	if r.Name != "Alice" {
		t.Errorf("First report: expected Name='Alice', got '%s'", r.Name)
	}

	// Check response message
	expected := "Laporan diterima, Alice sudah berkeringat 1 hari. Lanjutkan 🔥 (streak 1 minggu)"
	if !containsSubstring(msg, expected) {
		t.Errorf("Expected message to contain '%s', got '%s'", expected, msg)
	}
	if !containsSubstring(msg, "💬 _\"") {
		t.Errorf("Expected every #lapor response to contain random motivation, got '%s'", msg)
	}
	// Should also have unlocked the first season badge.
	if !containsSubstring(msg, "Awakened Hunter") {
		t.Errorf("Expected message to contain season badge 'Awakened Hunter', got '%s'", msg)
	}
	// Points: 10 base + 5 first season + 10 Awakened Hunter + 25 E-Rank Consistency = 50
	if r.TotalPoints != 50 {
		t.Errorf("First report: expected TotalPoints=50, got %d", r.TotalPoints)
	}
	if r.SeasonalPoints != 50 {
		t.Errorf("First report: expected SeasonalPoints=50, got %d", r.SeasonalPoints)
	}
}

func TestStreak_ConsecutiveDay_StreakIncreases(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report), dailyCounts: make(map[string]int)}
	uc := usecase.NewReportActivityUsecase(repo)
	ctx := context.Background()

	// Setup: user reported last week (consecutive week)
	lastWeek := time.Now().AddDate(0, 0, -7)
	repo.reports["user1"] = &domain.Report{
		UserID:         "user1",
		Name:           "Bob",
		Streak:         5,
		ActivityCount:  10,
		LastReportDate: lastWeek,
	}

	// Report today (consecutive day)
	_, err := uc.Execute(ctx, "user1", "Bob", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	r := repo.reports["user1"]
	if r.Streak != 6 {
		t.Errorf("Consecutive week: expected Streak=6 (5+1), got %d", r.Streak)
	}
	if r.ActivityCount != 11 {
		t.Errorf("Consecutive week: expected ActivityCount=11 (10+1), got %d", r.ActivityCount)
	}
}

func TestStreak_MissedDay_StreakResets(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report), dailyCounts: make(map[string]int)}
	uc := usecase.NewReportActivityUsecase(repo)
	ctx := context.Background()

	// Setup: user last reported 2 weeks ago (missed 1 whole week)
	twoWeeksAgo := time.Now().AddDate(0, 0, -14)
	repo.reports["user1"] = &domain.Report{
		UserID:         "user1",
		Name:           "Charlie",
		Streak:         20, // Had a 20-week streak
		ActivityCount:  25,
		LastReportDate: twoWeeksAgo,
	}

	// Report today after missing days
	_, err := uc.Execute(ctx, "user1", "Charlie", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	r := repo.reports["user1"]
	if r.Streak != 1 {
		t.Errorf("Missed week: expected Streak=1 (reset), got %d", r.Streak)
	}
	if r.ActivityCount != 26 {
		t.Errorf("Missed week: expected ActivityCount=26 (25+1, preserved), got %d", r.ActivityCount)
	}
}

func TestStreak_SameDay_SecondReport(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report), dailyCounts: make(map[string]int)}
	uc := usecase.NewReportActivityUsecase(repo)
	ctx := context.Background()

	// Setup: user already reported today
	now := time.Now()
	repo.reports["user1"] = &domain.Report{
		UserID:         "user1",
		Name:           "Diana",
		Streak:         7,
		ActivityCount:  15,
		LastReportDate: now,
	}
	// Simulate that the user already has 1 activity logged today
	repo.dailyCounts["user1|"+domain.GetToday(now).Format(time.DateOnly)] = 1

	// Report again same day (second report)
	msg, err := uc.Execute(ctx, "user1", "Diana", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should accept but with halved XP
	if !containsSubstring(msg, "Laporan diterima (laporan ke-2 hari ini)") {
		t.Errorf("Same day: expected repeat report message, got '%s'", msg)
	}
	if !containsSubstring(msg, "repeat report, ½ XP") {
		t.Errorf("Same day: expected halved XP note, got '%s'", msg)
	}

	// Values should NOT change (streak and activity count stay same)
	r := repo.reports["user1"]
	if r.Streak != 7 {
		t.Errorf("Same day: Streak should remain 7, got %d", r.Streak)
	}
	if r.ActivityCount != 15 {
		t.Errorf("Same day: ActivityCount should remain 15, got %d", r.ActivityCount)
	}
}

func TestReportLimit_IsSeparateFromSideQuestLimit(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report), dailyCounts: make(map[string]int)}
	uc := usecase.NewReportActivityUsecase(repo)
	ctx := context.Background()
	now := time.Now()
	todayKey := "user1|" + domain.GetToday(now).Format(time.DateOnly)
	repo.reports["user1"] = &domain.Report{
		UserID:         "user1",
		Name:           "Diana",
		Streak:         7,
		ActivityCount:  15,
		LastReportDate: now,
	}
	repo.dailyCounts[todayKey] = 3
	repo.activityKindCounts = map[string]map[string]int{
		todayKey: {
			domain.ActivityKindSideQuest: 3,
		},
	}

	msg, err := uc.Execute(ctx, "user1", "Diana", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !containsSubstring(msg, "Laporan diterima") {
		t.Fatalf("regular report should still be accepted after 3 side quests, got %q", msg)
	}
	if repo.activityKindCounts[todayKey][domain.ActivityKindRegularReport] != 1 {
		t.Fatalf("regular report should use the regular-report slot after side quests")
	}
}

func TestReportActivity_AnnouncesLifetimeTierAndSeasonRankUp(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report), dailyCounts: make(map[string]int)}
	uc := usecase.NewReportActivityUsecase(repo)
	ctx := context.Background()
	now := time.Now()
	repo.reports["user1"] = &domain.Report{
		UserID:                "user1",
		Name:                  "Diana",
		Streak:                1,
		ActivityCount:         15,
		SeasonalActivityCount: 5,
		TotalPoints:           1149,
		SeasonalPoints:        149,
		LastReportDate:        now,
	}

	msg, err := uc.Execute(ctx, "user1", "Diana", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !containsSubstring(msg, "TIER LIFETIME UP") || !containsSubstring(msg, "E-Tier Hunter") || !containsSubstring(msg, "D-Tier Hunter") {
		t.Fatalf("expected lifetime tier-up announcement, got %q", msg)
	}
	if !containsSubstring(msg, "RANK SEASON UP") || !containsSubstring(msg, "Warrior") || !containsSubstring(msg, "Elite") {
		t.Fatalf("expected season rank-up announcement, got %q", msg)
	}
}

func TestStreak_SameDay_MaxReportsReached(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report), dailyCounts: make(map[string]int)}
	uc := usecase.NewReportActivityUsecase(repo)
	ctx := context.Background()

	now := time.Now()
	repo.reports["user1"] = &domain.Report{
		UserID:         "user1",
		Name:           "Diana",
		Streak:         7,
		ActivityCount:  15,
		LastReportDate: now,
	}
	// Simulate that the user already has 3 activities logged today (max)
	repo.dailyCounts["user1|"+domain.GetToday(now).Format(time.DateOnly)] = 3

	// Fourth report should be rejected
	msg, err := uc.Execute(ctx, "user1", "Diana", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !containsSubstring(msg, "batas laporan utama 3x hari ini") {
		t.Errorf("Expected max limit message, got '%s'", msg)
	}
}

func TestStreak_SameDay_UsesStoredNameWhenIncomingNameUnknown(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report), dailyCounts: make(map[string]int)}
	uc := usecase.NewReportActivityUsecase(repo)
	ctx := context.Background()

	repo.reports["user1"] = &domain.Report{
		UserID:         "user1",
		Name:           "Budi",
		Streak:         3,
		ActivityCount:  8,
		LastReportDate: time.Now(),
	}

	msg, err := uc.Execute(ctx, "user1", "Unknown", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !containsSubstring(msg, "Budi") {
		t.Fatalf("Duplicate #lapor should use stored name, got %q", msg)
	}
	if containsSubstring(msg, "Unknown") {
		t.Fatalf("Duplicate #lapor should not show Unknown, got %q", msg)
	}
}

func TestStreak_SameDay_HealsUnknownStoredName(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report), dailyCounts: make(map[string]int)}
	uc := usecase.NewReportActivityUsecase(repo)
	ctx := context.Background()

	repo.reports["user1"] = &domain.Report{
		UserID:         "user1",
		Name:           "Unknown",
		Streak:         3,
		ActivityCount:  8,
		LastReportDate: time.Now(),
	}

	msg, err := uc.Execute(ctx, "user1", "Budi", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !containsSubstring(msg, "Budi") {
		t.Fatalf("Duplicate #lapor should use incoming valid name, got %q", msg)
	}
	if repo.reports["user1"].Name != "Budi" {
		t.Fatalf("Duplicate #lapor should persist healed name, got %q", repo.reports["user1"].Name)
	}
	if repo.reports["user1"].ActivityCount != 8 {
		t.Fatalf("Duplicate #lapor must not change activity count, got %d", repo.reports["user1"].ActivityCount)
	}
}

func TestStreak_FirstReport_DoesNotPersistUnknownName(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report), dailyCounts: make(map[string]int)}
	uc := usecase.NewReportActivityUsecase(repo)
	ctx := context.Background()

	msg, err := uc.Execute(ctx, "user1", "Unknown", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if repo.reports["user1"].Name != "Teman" {
		t.Fatalf("First #lapor should persist safe fallback name, got %q", repo.reports["user1"].Name)
	}
	if containsSubstring(msg, "Unknown") {
		t.Fatalf("First #lapor should not show Unknown fallback, got %q", msg)
	}
}

func TestStreak_ExistingUnknownStoredNameWithUnknownIncomingNormalizesToTeman(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report), dailyCounts: make(map[string]int)}
	uc := usecase.NewReportActivityUsecase(repo)
	ctx := context.Background()

	repo.reports["user1"] = &domain.Report{
		UserID:         "user1",
		Name:           "Unknown",
		Streak:         3,
		ActivityCount:  8,
		LastReportDate: time.Now().AddDate(0, 0, -7),
	}

	msg, err := uc.Execute(ctx, "user1", "Unknown", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if repo.reports["user1"].Name != "Teman" {
		t.Fatalf("Expected stored name normalized to Teman, got %q", repo.reports["user1"].Name)
	}
	if containsSubstring(msg, "Unknown") {
		t.Fatalf("Response should not show Unknown, got %q", msg)
	}
}

func TestStreak_LongGap_StreakResets(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report), dailyCounts: make(map[string]int)}
	uc := usecase.NewReportActivityUsecase(repo)
	ctx := context.Background()

	// Setup: user last reported 30 days ago
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	repo.reports["user1"] = &domain.Report{
		UserID:         "user1",
		Name:           "Eve",
		Streak:         36, // Was at max streak
		ActivityCount:  36,
		LastReportDate: thirtyDaysAgo,
	}

	// Report today after long absence
	_, err := uc.Execute(ctx, "user1", "Eve", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	r := repo.reports["user1"]
	if r.Streak != 1 {
		t.Errorf("Long gap: expected Streak=1 (reset), got %d", r.Streak)
	}
	if r.ActivityCount != 37 {
		t.Errorf("Long gap: expected ActivityCount=37 (36+1), got %d", r.ActivityCount)
	}
}

// =============================================================================
// LEADERBOARD DISPLAY LOGIC
// =============================================================================
//
// Active 🔥: Reported today OR yesterday (still has time to report today)
// Lost 💔: Last report was before yesterday (streak broken)
//
// Ranking: By SeasonalPoints, then SeasonalActivityCount, then Name.
// Lifetime days are displayed separately and include centurion cycles.
//
// =============================================================================

func TestLeaderboard_RanksBySeasonPoints(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report), dailyCounts: make(map[string]int)}
	uc := usecase.NewGetLeaderboardUsecase(repo)
	ctx := context.Background()

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	twoWeeksAgo := now.AddDate(0, 0, -14)

	// Setup: 3 users with different activity counts and streak status
	repo.reports["user1"] = &domain.Report{
		UserID:                "user1",
		Name:                  "HighTotal_LostStreak",
		Streak:                0,
		ActivityCount:         30,
		SeasonalPoints:        90,
		SeasonalActivityCount: 4,
		LastReportDate:        twoWeeksAgo,
	}
	repo.reports["user2"] = &domain.Report{
		UserID:                "user2",
		Name:                  "MediumTotal_ActiveStreak",
		Streak:                25,
		ActivityCount:         25,
		SeasonalPoints:        120,
		SeasonalActivityCount: 2,
		LastReportDate:        yesterday,
	}
	repo.reports["user3"] = &domain.Report{
		UserID:                "user3",
		Name:                  "LowTotal_ActiveStreak",
		Streak:                10,
		ActivityCount:         10,
		SeasonalPoints:        120,
		SeasonalActivityCount: 5,
		LastReportDate:        now,
	}

	// Get leaderboard
	result, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify ranking order in output (should be by season points, then season active days)
	pos1 := indexOf(result, "HighTotal_LostStreak")
	pos2 := indexOf(result, "MediumTotal_ActiveStreak")
	pos3 := indexOf(result, "LowTotal_ActiveStreak")

	if pos3 > pos2 || pos2 > pos1 {
		t.Errorf("Leaderboard should rank by season points then active days: got positions high=%d, medium=%d, low=%d", pos1, pos2, pos3)
	}

	// Verify emojis and format
	if !containsSubstring(result, "HighTotal_LostStreak — 90 pts (4 hari season, 30 hari lifetime, 💔)") {
		t.Errorf("Lost streak user should have 💔 emoji in ranking")
	}
	if !containsSubstring(result, "MediumTotal_ActiveStreak — 120 pts (2 hari season, 25 hari lifetime, 25 minggu streak 🔥)") {
		t.Errorf("Active streak user should have weeks streak and 🔥 emoji")
	}
}

// =============================================================================
// CENTURION PRESTIGE TESTS
// =============================================================================

func TestCenturion_PrestigeTransition(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report), dailyCounts: make(map[string]int)}
	uc := usecase.NewReportActivityUsecase(repo)
	ctx := context.Background()

	// 1. Setup: User at Day 99
	lastWeek := time.Now().AddDate(0, 0, -7)
	repo.reports["user1"] = &domain.Report{
		UserID:         "user1",
		Name:           "Centurion-to-be",
		Streak:         12,
		ActivityCount:  99,
		LastReportDate: lastWeek,
	}

	// 2. Report Day 100
	msg, _ := uc.Execute(ctx, "user1", "Centurion-to-be", nil)
	r := repo.reports["user1"]
	if r.ActivityCount != 100 {
		t.Errorf("Expected day 100, got %d", r.ActivityCount)
	}
	if !containsSubstring(msg, "LUAR BIASA") || !containsSubstring(msg, "HARI KE-100") {
		t.Errorf("Expected 100-day celebration message, got: %s", msg)
	}

	// 3. Report Day 101 (should transition to Cycle 1, Day 1)
	r.LastReportDate = time.Now() // set to "today" relative to the logic inside uc.Execute(now)
	// Note: Execute uses time.Now(), so we need to be careful.
	// In the real app we can't easily spoof time.now() without a clock provider.
	// But in these tests, uc.Execute is called. Let's just call it again but simulate it's a new day/week.

	// Fast-forward the report object's date to "yesterday" so the next Execute() call increments the streak
	repo.reports["user1"].LastReportDate = time.Now().AddDate(0, 0, -7)
	repo.dailyCounts = make(map[string]int)
	repo.activityKindCounts = make(map[string]map[string]int)

	msg, _ = uc.Execute(ctx, "user1", "Centurion-to-be", nil)
	r = repo.reports["user1"]

	if r.ActivityCount != 1 {
		t.Errorf("Expected ActivityCount reset to 1 after 100, got %d", r.ActivityCount)
	}
	if r.CenturionCycles != 1 {
		t.Errorf("Expected CenturionCycles=1 after day 101, got %d", r.CenturionCycles)
	}
	if r.Streak != 14 { // 12 (original) + 1 (for day 100) + 1 (for day 101)
		t.Errorf("Expected streak to continue to 14, got %d", r.Streak)
	}
	if !containsSubstring(msg, "ERA BARU DIMULAI") || !containsSubstring(msg, "Siklus 2") {
		t.Errorf("Expected cycle transition message, got: %s", msg)
	}
	if !containsSubstring(msg, "[C2] Centurion-to-be") {
		t.Errorf("Expected [C2] badge in message, got: %s", msg)
	}
}

func TestLeaderboard_UsesSeasonSortingAndDisplaysTotalActiveDays(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report), dailyCounts: make(map[string]int)}
	uc := usecase.NewGetLeaderboardUsecase(repo)
	ctx := context.Background()

	now := time.Now()

	// Setup:
	// User A: Cycle 0, Day 50 (Experienced)
	// User B: Cycle 1, Day 1 (Just "lapped" the leaderboard)
	repo.reports["userA"] = &domain.Report{
		Name:                  "OldGuard",
		ActivityCount:         50,
		CenturionCycles:       0,
		SeasonalPoints:        100,
		SeasonalActivityCount: 5,
		LastReportDate:        now,
	}
	repo.reports["userB"] = &domain.Report{
		Name:                  "PrestigePlayer",
		ActivityCount:         1,
		CenturionCycles:       1,
		SeasonalPoints:        120,
		SeasonalActivityCount: 3,
		LastReportDate:        now,
	}

	result, _ := uc.Execute(ctx)

	// Season points are canonical, so PrestigePlayer ranks higher even though its cycle day is 1.
	posA := indexOf(result, "OldGuard")
	posB := indexOf(result, "PrestigePlayer")

	if posB > posA {
		t.Errorf("Leaderboard should put higher season points first. Got positions A:%d, B:%d", posA, posB)
	}

	if !containsSubstring(result, "[S1-C2] PrestigePlayer") {
		t.Errorf("Leaderboard should show [S1-C2] badge for the Prestige player")
	}
	if !containsSubstring(result, "101 hari lifetime") {
		t.Errorf("Leaderboard should display centurion-aware lifetime days, got: %s", result)
	}
}

// Helper functions
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func containsSubstring(s, substr string) bool {
	return indexOf(s, substr) >= 0
}
