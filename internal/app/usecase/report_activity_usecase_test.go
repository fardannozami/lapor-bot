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
	msg, err := uc.Execute(ctx, "user1", "Alice")
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
	expected := "Laporan diterima, Alice sudah berkeringat 1 hari. Lanjutkan 🔥 (streak 1 hari)"
	if msg != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, msg)
	}
}

func TestStreak_ConsecutiveDay_StreakIncreases(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report)}
	uc := usecase.NewReportActivityUsecase(repo)
	ctx := context.Background()

	// Setup: user reported yesterday
	yesterday := time.Now().AddDate(0, 0, -1)
	repo.reports["user1"] = &domain.Report{
		UserID:         "user1",
		Name:           "Bob",
		Streak:         5,
		ActivityCount:  10,
		LastReportDate: yesterday,
	}

	// Report today (consecutive day)
	_, err := uc.Execute(ctx, "user1", "Bob")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	r := repo.reports["user1"]
	if r.Streak != 6 {
		t.Errorf("Consecutive day: expected Streak=6 (5+1), got %d", r.Streak)
	}
	if r.ActivityCount != 11 {
		t.Errorf("Consecutive day: expected ActivityCount=11 (10+1), got %d", r.ActivityCount)
	}
}

func TestStreak_MissedDay_StreakResets(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report)}
	uc := usecase.NewReportActivityUsecase(repo)
	ctx := context.Background()

	// Setup: user last reported 3 days ago (missed 2 days)
	threeDaysAgo := time.Now().AddDate(0, 0, -3)
	repo.reports["user1"] = &domain.Report{
		UserID:         "user1",
		Name:           "Charlie",
		Streak:         20, // Had a 20-day streak
		ActivityCount:  25,
		LastReportDate: threeDaysAgo,
	}

	// Report today after missing days
	_, err := uc.Execute(ctx, "user1", "Charlie")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	r := repo.reports["user1"]
	if r.Streak != 1 {
		t.Errorf("Missed day: expected Streak=1 (reset), got %d", r.Streak)
	}
	if r.ActivityCount != 26 {
		t.Errorf("Missed day: expected ActivityCount=26 (25+1, preserved), got %d", r.ActivityCount)
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
	msg, err := uc.Execute(ctx, "user1", "Diana")
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
	_, err := uc.Execute(ctx, "user1", "Eve")
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
	weekAgo := now.AddDate(0, 0, -7)

	// Setup: 3 users with different activity counts and streak status
	repo.reports["user1"] = &domain.Report{
		UserID:         "user1",
		Name:           "HighTotal_LostStreak",
		Streak:         0,
		ActivityCount:  30, // Highest total, but lost streak
		LastReportDate: weekAgo,
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

	// Verify emojis
	if !containsSubstring(result, "HighTotal_LostStreak - 30 days 💔") {
		t.Errorf("Lost streak user should have 💔 emoji")
	}
	if !containsSubstring(result, "MediumTotal_ActiveStreak - 25 days 🔥") {
		t.Errorf("Active streak user should have 🔥 emoji")
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
