package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/app/usecase"
	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type mockRepo struct {
	reports map[string]*domain.Report
}

func (m *mockRepo) GetReport(ctx context.Context, userID string) (*domain.Report, error) {
	return m.reports[userID], nil
}

func (m *mockRepo) UpsertReport(ctx context.Context, report *domain.Report) error {
	m.reports[report.UserID] = report
	return nil
}

func (m *mockRepo) GetAllReports(ctx context.Context) ([]*domain.Report, error) {
	var result []*domain.Report
	for _, r := range m.reports {
		result = append(result, r)
	}
	return result, nil
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
	repo := &mockRepo{reports: make(map[string]*domain.Report)}
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
	// Should also have unlocked "Pemula" achievement
	if !containsSubstring(msg, "Pemula") {
		t.Errorf("Expected message to contain achievement 'Pemula', got '%s'", msg)
	}
}

func TestStreak_ConsecutiveDay_StreakIncreases(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report)}
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
	repo := &mockRepo{reports: make(map[string]*domain.Report)}
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

func TestStreak_SameDay_Rejected(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report)}
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

	// Try to report again same day
	msg, err := uc.Execute(ctx, "user1", "Diana", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should be rejected with warning
	expected := "Diana sudah laporan hari ini, ayo jangan curang! 😉"
	if msg != expected {
		t.Errorf("Same day: expected rejection message, got '%s'", msg)
	}

	// Values should NOT change
	r := repo.reports["user1"]
	if r.Streak != 7 {
		t.Errorf("Same day: Streak should remain 7, got %d", r.Streak)
	}
	if r.ActivityCount != 15 {
		t.Errorf("Same day: ActivityCount should remain 15, got %d", r.ActivityCount)
	}
}

func TestStreak_LongGap_StreakResets(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report)}
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
// Ranking: By ActivityCount (total days), NOT by streak
// Someone with 30 days 💔 ranks above someone with 25 days 🔥
//
// =============================================================================

func TestLeaderboard_RanksByActivityCount(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report)}
	uc := usecase.NewGetLeaderboardUsecase(repo)
	ctx := context.Background()

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	twoWeeksAgo := now.AddDate(0, 0, -14)

	// Setup: 3 users with different activity counts and streak status
	repo.reports["user1"] = &domain.Report{
		UserID:         "user1",
		Name:           "HighTotal_LostStreak",
		Streak:         0,
		ActivityCount:  30, // Highest total, but lost streak
		LastReportDate: twoWeeksAgo,
	}
	repo.reports["user2"] = &domain.Report{
		UserID:         "user2",
		Name:           "MediumTotal_ActiveStreak",
		Streak:         25,
		ActivityCount:  25, // Medium total, active streak
		LastReportDate: yesterday,
	}
	repo.reports["user3"] = &domain.Report{
		UserID:         "user3",
		Name:           "LowTotal_ActiveStreak",
		Streak:         10,
		ActivityCount:  10, // Lowest total, active streak
		LastReportDate: now,
	}

	// Get leaderboard
	result, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify ranking order in output (should be by ActivityCount)
	// HighTotal should appear before MediumTotal, which appears before LowTotal
	pos1 := indexOf(result, "HighTotal_LostStreak")
	pos2 := indexOf(result, "MediumTotal_ActiveStreak")
	pos3 := indexOf(result, "LowTotal_ActiveStreak")

	if pos1 > pos2 || pos2 > pos3 {
		t.Errorf("Leaderboard should rank by ActivityCount: got positions %d, %d, %d", pos1, pos2, pos3)
	}

	// Verify emojis and format
	if !containsSubstring(result, "HighTotal_LostStreak - 30 days (💔)") {
		t.Errorf("Lost streak user should have 💔 emoji in ranking")
	}
	if !containsSubstring(result, "MediumTotal_ActiveStreak - 25 days (25 weeks streak 🔥)") {
		t.Errorf("Active streak user should have weeks streak and 🔥 emoji")
	}
}

// =============================================================================
// CENTURION PRESTIGE TESTS
// =============================================================================

func TestCenturion_PrestigeTransition(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report)}
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

func TestLeaderboard_CenturionSorting(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report)}
	uc := usecase.NewGetLeaderboardUsecase(repo)
	ctx := context.Background()

	now := time.Now()

	// Setup: 
	// User A: Cycle 0, Day 50 (Experienced)
	// User B: Cycle 1, Day 1 (Just "lapped" the leaderboard)
	repo.reports["userA"] = &domain.Report{
		Name:           "OldGuard",
		ActivityCount:  50,
		CenturionCycles: 0,
		LastReportDate: now,
	}
	repo.reports["userB"] = &domain.Report{
		Name:           "PrestigePlayer",
		ActivityCount:  1,
		CenturionCycles: 1,
		LastReportDate: now,
	}

	result, _ := uc.Execute(ctx)

	// Since we sort by ActivityCount DESC, Day 50 (OldGuard) should be higher than Day 1 (PrestigePlayer)
	posA := indexOf(result, "OldGuard")
	posB := indexOf(result, "PrestigePlayer")

	if posA > posB {
		t.Errorf("Leaderboard should put Day 50 above Day 1 even if Day 1 is Cycle 2. Got positions A:%d, B:%d", posA, posB)
	}
	
	if !containsSubstring(result, "[S1-C2] PrestigePlayer") {
		t.Errorf("Leaderboard should show [S1-C2] badge for the Prestige player")
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
