package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/app/usecase"
	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

// =============================================================================
// HANDLE MESSAGE USECASE TESTS
// =============================================================================
//
// Tests command routing logic:
// - /lapor → routes to ReportActivityUsecase
// - /leaderboard → disabled command fallback
// - regular chat messages → returns empty string (no response)
// - unknown slash commands → fallback response
//
// =============================================================================

type mockReportUsecase struct {
	called   bool
	userID   string
	name     string
	response string
}

func (m *mockReportUsecase) Execute(ctx context.Context, userID, name string) (string, error) {
	m.called = true
	m.userID = userID
	m.name = name
	return m.response, nil
}

type mockLeaderboardUsecase struct {
	called   bool
	response string
}

func (m *mockLeaderboardUsecase) Execute(ctx context.Context) (string, error) {
	m.called = true
	return m.response, nil
}

// mockReportRepo implements domain.ReportRepository for testing
type mockReportRepo struct {
	domain.ReportRepository
	reports          map[string]*domain.Report
	activityCounts   []domain.ActivityLeaderboardEntry
	dailyCountByKind map[string]int
	deletedLogKind   string
	goals            map[string]*domain.WeeklyGoal
}

func (m *mockReportRepo) GetReport(ctx context.Context, userID string) (*domain.Report, error) {
	return m.reports[userID], nil
}

func (m *mockReportRepo) UpsertReport(ctx context.Context, report *domain.Report) error {
	m.reports[report.UserID] = report
	return nil
}

func (m *mockReportRepo) UpsertReportWithActivity(ctx context.Context, report *domain.Report, activityDate time.Time) error {
	m.reports[report.UserID] = report
	return nil
}

func (m *mockReportRepo) GetAllReports(ctx context.Context) ([]*domain.Report, error) {
	var result []*domain.Report
	for _, r := range m.reports {
		result = append(result, r)
	}
	return result, nil
}

func (m *mockReportRepo) LogActivity(ctx context.Context, userID string, activityDate time.Time) error {
	return nil
}

func (m *mockReportRepo) GetActivityCountsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]domain.ActivityLeaderboardEntry, error) {
	return m.activityCounts, nil
}

func (m *mockReportRepo) ResolveLIDToPhone(ctx context.Context, lid string) string {
	return lid
}

func (m *mockReportRepo) InitTable(ctx context.Context) error {
	return nil
}

func (m *mockReportRepo) GetInactiveUsers(ctx context.Context, days int) ([]*domain.Report, error) {
	return nil, nil
}

func (m *mockReportRepo) ResetAllReports(ctx context.Context) error {
	m.reports = make(map[string]*domain.Report)
	return nil
}

func (m *mockReportRepo) UpsertStravaAccount(ctx context.Context, account *domain.StravaAccount) error {
	return nil
}

func (m *mockReportRepo) GetStravaAccountByAthleteID(ctx context.Context, athleteID int64) (*domain.StravaAccount, error) {
	return nil, nil
}

func (m *mockReportRepo) GetStravaAccountByUserID(ctx context.Context, userID string) (*domain.StravaAccount, error) {
	return nil, nil
}

func (m *mockReportRepo) DeleteActivityLog(ctx context.Context, userID string, activityDate time.Time) error {
	return nil
}

func (m *mockReportRepo) DeleteActivityLogByKind(ctx context.Context, userID string, activityDate time.Time, kind string) error {
	m.deletedLogKind = kind
	if m.dailyCountByKind != nil {
		m.dailyCountByKind[kind] = 0
	}
	return nil
}

func (m *mockReportRepo) DeleteLatestActivityLog(ctx context.Context, userID string, activityDate time.Time) (int, error) {
	return 0, nil
}

func (m *mockReportRepo) DeleteLatestActivityLogByKind(ctx context.Context, userID string, activityDate time.Time, kind string) (int, error) {
	m.deletedLogKind = kind
	if m.dailyCountByKind == nil {
		return 0, nil
	}
	if m.dailyCountByKind[kind] > 0 {
		m.dailyCountByKind[kind]--
	}
	return m.dailyCountByKind[kind], nil
}

func (m *mockReportRepo) GetUserActivityDates(ctx context.Context, userID string) ([]time.Time, error) {
	return nil, nil
}

func (m *mockReportRepo) GetUserActivityDatesByKind(ctx context.Context, userID string, kind string) ([]time.Time, error) {
	return nil, nil
}

func (m *mockReportRepo) DeleteReport(ctx context.Context, userID string) error {
	delete(m.reports, userID)
	return nil
}

func (m *mockReportRepo) GetDailyActivityCount(ctx context.Context, userID string, date time.Time) (int, error) {
	return 0, nil
}

func (m *mockReportRepo) GetDailyActivityCountByKind(ctx context.Context, userID string, date time.Time, kind string) (int, error) {
	if m.dailyCountByKind == nil {
		return 0, nil
	}
	return m.dailyCountByKind[kind], nil
}

func (m *mockReportRepo) SaveDailyQuest(ctx context.Context, userID, questDate, tasksJSON string) error {
	return nil
}

func (m *mockReportRepo) GetDailyQuest(ctx context.Context, userID, questDate string) (string, error) {
	return "", nil
}

func (m *mockReportRepo) SetGoal(ctx context.Context, goal *domain.WeeklyGoal) error {
	if m.goals == nil {
		m.goals = make(map[string]*domain.WeeklyGoal)
	}
	m.goals[goal.UserID+"|"+goal.StartAt.Format(time.RFC3339)] = goal
	return nil
}

func (m *mockReportRepo) GetActiveGoal(ctx context.Context, userID string, now time.Time) (*domain.WeeklyGoal, error) {
	for _, goal := range m.goals {
		if goal.UserID == userID && !now.Before(goal.StartAt) && now.Before(goal.EndAt) {
			return goal, nil
		}
	}
	return nil, nil
}

func (m *mockReportRepo) DeleteActiveGoal(ctx context.Context, userID string, now time.Time) error {
	for key, goal := range m.goals {
		if goal.UserID == userID && !now.Before(goal.StartAt) && now.Before(goal.EndAt) {
			delete(m.goals, key)
		}
	}
	return nil
}

func (m *mockReportRepo) DeleteExpiredGoals(ctx context.Context, now time.Time) (int64, error) {
	var deleted int64
	for key, goal := range m.goals {
		if !goal.EndAt.After(now) {
			delete(m.goals, key)
			deleted++
		}
	}
	return deleted, nil
}

func (m *mockReportRepo) GetGoalActivities(ctx context.Context, userID string, startDate, endDate time.Time) ([]domain.GoalActivity, error) {
	return nil, nil
}

func (m *mockReportRepo) RecordGoalActivity(ctx context.Context, userID string, activityDate time.Time, activityText string) (bool, error) {
	return false, nil
}

func TestHandleMessage_LaporCommand(t *testing.T) {
	repo := &mockReportRepo{reports: make(map[string]*domain.Report)}
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, usecase.NewCancelReportUsecase(repo), updateNameUC, nil, usecase.NewBroadcastUpdateUsecase(), usecase.NewGetMotivationUsecase(), usecase.NewGetHelpUsecase())

	ctx := context.Background()

	// Test /lapor command
	msg, err := handleUC.Execute(ctx, "user123", "TestUser", "/lapor")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should route to report usecase and return response
	expected := "Laporan diterima, TestUser sudah berkeringat 1 hari. Lanjutkan 🔥 (streak 1 minggu)"
	if !containsSubstring(msg.Text, expected) {
		t.Errorf("Expected message to contain '%s', got '%s'", expected, msg.Text)
	}

	// Verify user was created
	r := repo.reports["user123"]
	if r == nil {
		t.Error("Report should have been created")
	} else if r.Name != "TestUser" {
		t.Errorf("Expected name 'TestUser', got '%s'", r.Name)
	}
}

func TestHandleMessage_LaporCaseInsensitive(t *testing.T) {
	repo := &mockReportRepo{reports: make(map[string]*domain.Report)}
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, usecase.NewCancelReportUsecase(repo), updateNameUC, nil, usecase.NewBroadcastUpdateUsecase(), usecase.NewGetMotivationUsecase(), usecase.NewGetHelpUsecase())

	ctx := context.Background()

	testCases := []string{"/LAPOR", "/Lapor", "/LaPor", "/lapor"}
	for _, cmd := range testCases {
		// Reset repo for each test
		repo.reports = make(map[string]*domain.Report)

		msg, err := handleUC.Execute(ctx, "user1", "User", cmd)
		if err != nil {
			t.Fatalf("Unexpected error for '%s': %v", cmd, err)
		}
		if msg.Text == "" {
			t.Errorf("Command '%s' should return a response", cmd)
		}
	}
}

func TestHandleMessage_LaporWithTrailingText(t *testing.T) {
	repo := &mockReportRepo{reports: make(map[string]*domain.Report)}
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, usecase.NewCancelReportUsecase(repo), updateNameUC, nil, usecase.NewBroadcastUpdateUsecase(), usecase.NewGetMotivationUsecase(), usecase.NewGetHelpUsecase())

	ctx := context.Background()

	// Command with trailing text should still work
	msg, err := handleUC.Execute(ctx, "user1", "User", "/lapor hari ini olahraga lari")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if msg.Text == "" {
		t.Error("/lapor with trailing text should still be recognized")
	}
}

func TestHandleMessage_LeaderboardCommand(t *testing.T) {
	repo := &mockReportRepo{reports: make(map[string]*domain.Report)}
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, usecase.NewCancelReportUsecase(repo), updateNameUC, nil, usecase.NewBroadcastUpdateUsecase(), usecase.NewGetMotivationUsecase(), usecase.NewGetHelpUsecase())

	ctx := context.Background()

	// Setup some test data
	repo.reports["user1"] = &domain.Report{
		UserID:        "user1",
		Name:          "Alice",
		Streak:        5,
		ActivityCount: 5,
	}

	// /leaderboard command is currently disabled — should be silent
	msg, err := handleUC.Execute(ctx, "user1", "User", "/leaderboard")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if msg.Text != "" {
		t.Errorf("Disabled command should be silently ignored, got '%s'", msg.Text)
	}
}

func TestHandleMessage_LeaderboardCaseInsensitive(t *testing.T) {
	repo := &mockReportRepo{reports: make(map[string]*domain.Report)}
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, usecase.NewCancelReportUsecase(repo), updateNameUC, nil, usecase.NewBroadcastUpdateUsecase(), usecase.NewGetMotivationUsecase(), usecase.NewGetHelpUsecase())

	ctx := context.Background()

	testCases := []string{"/LEADERBOARD", "/Leaderboard", "/LeaderBoard", "/leaderboard"}
	for _, cmd := range testCases {
		msg, err := handleUC.Execute(ctx, "user1", "User", cmd)
		if err != nil {
			t.Fatalf("Unexpected error for '%s': %v", cmd, err)
		}
		if msg.Text != "" {
			t.Errorf("Disabled command '%s' should be silently ignored, got '%s'", cmd, msg.Text)
		}
	}
}

func TestHandleMessage_WeeklyLeaderboardCommand(t *testing.T) {
	repo := &mockReportRepo{
		reports: make(map[string]*domain.Report),
		activityCounts: []domain.ActivityLeaderboardEntry{
			{Name: "Alice", ActivityCount: 3},
			{Name: "Budi", ActivityCount: 1},
		},
	}
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, usecase.NewCancelReportUsecase(repo), updateNameUC, nil, usecase.NewBroadcastUpdateUsecase(), usecase.NewGetMotivationUsecase(), usecase.NewGetHelpUsecase())

	ctx := context.Background()

	msg, err := handleUC.Execute(ctx, "user1", "User", "/leaderboard-weekly")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if msg.Text != "" {
		t.Fatalf("Disabled weekly leaderboard should be silently ignored, got %q", msg.Text)
	}
}

func TestHandleMessage_NonCommand_ReturnsNoResponse(t *testing.T) {
	repo := &mockReportRepo{reports: make(map[string]*domain.Report)}
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, usecase.NewCancelReportUsecase(repo), updateNameUC, nil, usecase.NewBroadcastUpdateUsecase(), usecase.NewGetMotivationUsecase(), usecase.NewGetHelpUsecase())

	ctx := context.Background()

	testCases := []string{
		"hello",
		"random message",
		"lapor",
		"leaderboard",
		"#invalid",
	}

	for _, msgText := range testCases {
		result, err := handleUC.Execute(ctx, "user1", "User", msgText)
		if err != nil {
			t.Fatalf("Unexpected error for '%s': %v", msgText, err)
		}
		if result.Text != "" {
			t.Errorf("Non-command message '%s' should not return a response, got '%s'", msgText, result.Text)
		}
	}
}

func TestHandleMessage_HashCommand_WorksLikeSlash(t *testing.T) {
	repo := &mockReportRepo{reports: make(map[string]*domain.Report)}
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, usecase.NewCancelReportUsecase(repo), updateNameUC, nil, usecase.NewBroadcastUpdateUsecase(), usecase.NewGetMotivationUsecase(), usecase.NewGetHelpUsecase())

	ctx := context.Background()

	// Legacy "#" prefix should work exactly like "/" for all commands.
	tests := []struct {
		input        string
		wantContains string
	}{
		{"#lapor", "Laporan diterima"},
		{"#lapor Push Day", "Laporan diterima"},
		{"#lapor sidequest", "belum terdaftar"},
		{"#help", "Lapor Bot"},
		{"#tutorial", "Panduan"},
		{"#cancel", "belum pernah"},
		{"#cancel-all", "belum pernah"},
		{"#LAPOR", "Laporan diterima"},
	}
	for _, tt := range tests {
		// Reset for each test case
		repo.reports = make(map[string]*domain.Report)

		result, err := handleUC.Execute(ctx, "user1", "User", tt.input)
		if err != nil {
			t.Fatalf("Unexpected error for '%s': %v", tt.input, err)
		}
		if result.Text == "" {
			t.Errorf("Hash command '%s' should produce a response, got empty", tt.input)
			continue
		}
		if !containsSubstring(result.Text, tt.wantContains) {
			t.Errorf("Hash command '%s': expected to contain '%s', got '%s'", tt.input, tt.wantContains, result.Text)
		}
	}
}

func TestHandleMessage_UnknownHashCommand_Silent(t *testing.T) {
	repo := &mockReportRepo{reports: make(map[string]*domain.Report)}
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, usecase.NewCancelReportUsecase(repo), updateNameUC, nil, usecase.NewBroadcastUpdateUsecase(), usecase.NewGetMotivationUsecase(), usecase.NewGetHelpUsecase())

	ctx := context.Background()

	// Unknown hash commands should be silently ignored (bot does not react)
	tests := []string{
		"#random",
		"#makan",
		"#olahraga",
	}
	for _, msgText := range tests {
		result, err := handleUC.Execute(ctx, "user1", "User", msgText)
		if err != nil {
			t.Fatalf("Unexpected error for '%s': %v", msgText, err)
		}
		if result.Text != "" {
			t.Errorf("Unknown hash command '%s' should be silently ignored, got '%s'", msgText, result.Text)
		}
	}
}

func TestHandleMessage_UnknownSlashCommand_Silent(t *testing.T) {
	repo := &mockReportRepo{reports: make(map[string]*domain.Report)}
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, usecase.NewCancelReportUsecase(repo), updateNameUC, nil, usecase.NewBroadcastUpdateUsecase(), usecase.NewGetMotivationUsecase(), usecase.NewGetHelpUsecase())

	ctx := context.Background()

	testCases := []string{"/invalid", "/mystats", "/laporhalo", "/leaderboard"}
	for _, msgText := range testCases {
		result, err := handleUC.Execute(ctx, "user1", "User", msgText)
		if err != nil {
			t.Fatalf("Unexpected error for '%s': %v", msgText, err)
		}
		if result.Text != "" {
			t.Errorf("Unknown slash command '%s' should be silently ignored, got '%s'", msgText, result.Text)
		}
	}
}

func TestHandleMessage_WhitespaceHandling(t *testing.T) {
	repo := &mockReportRepo{reports: make(map[string]*domain.Report)}
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, usecase.NewCancelReportUsecase(repo), updateNameUC, nil, usecase.NewBroadcastUpdateUsecase(), usecase.NewGetMotivationUsecase(), usecase.NewGetHelpUsecase())

	ctx := context.Background()

	// Test with leading/trailing whitespace
	testCases := []string{
		"  /lapor",
		"/lapor  ",
		"  /lapor  ",
		"\t/lapor",
		"\n/lapor\n",
	}

	for i, cmd := range testCases {
		repo.reports = make(map[string]*domain.Report) // Reset
		msg, err := handleUC.Execute(ctx, "user1", "User", cmd)
		if err != nil {
			t.Fatalf("Test %d: Unexpected error for '%q': %v", i, cmd, err)
		}
		if msg.Text == "" {
			t.Errorf("Test %d: Command '%q' with whitespace should still work", i, cmd)
		}
	}
}

func TestHandleMessage_EmptyMessage(t *testing.T) {
	repo := &mockReportRepo{reports: make(map[string]*domain.Report)}
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, usecase.NewCancelReportUsecase(repo), updateNameUC, nil, usecase.NewBroadcastUpdateUsecase(), usecase.NewGetMotivationUsecase(), usecase.NewGetHelpUsecase())

	ctx := context.Background()

	result, err := handleUC.Execute(ctx, "user1", "User", "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Text != "" {
		t.Errorf("Empty message should return empty string, got '%s'", result.Text)
	}
}

func TestHandleMessage_GamificationCommands(t *testing.T) {
	repo := &mockReportRepo{reports: make(map[string]*domain.Report)}
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, usecase.NewCancelReportUsecase(repo), updateNameUC, nil, usecase.NewBroadcastUpdateUsecase(), usecase.NewGetMotivationUsecase(), usecase.NewGetHelpUsecase())

	ctx := context.Background()

	// Setup data
	repo.reports["user1"] = &domain.Report{
		UserID:        "user1",
		Name:          "Gamer",
		Streak:        10,
		ActivityCount: 10,
		MaxStreak:     10,
		TotalPoints:   50,
		Achievements:  "first_report",
	}

	// /mystats is now disabled — should get fallback
	msg, err := handleUC.Execute(ctx, "user1", "Gamer", "/mystats")
	if err != nil {
		t.Fatalf("Unexpected error for /mystats: %v", err)
	}
	if msg.Text != "" {
		t.Errorf("Disabled /mystats should be silently ignored, got: %s", msg.Text)
	}

	// /achievements is now disabled — should get fallback
	msg, err = handleUC.Execute(ctx, "user1", "Gamer", "/achievements")
	if err != nil {
		t.Fatalf("Unexpected error for /achievements: %v", err)
	}
	if msg.Text != "" {
		t.Errorf("Disabled /achievements should be silently ignored, got: %s", msg.Text)
	}
}

func TestHandleMessage_JobCommands(t *testing.T) {
	repo := &mockReportRepo{reports: make(map[string]*domain.Report)}
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, usecase.NewCancelReportUsecase(repo), updateNameUC, nil, usecase.NewBroadcastUpdateUsecase(), usecase.NewGetMotivationUsecase(), usecase.NewGetHelpUsecase())

	ctx := context.Background()

	// /jobs command is now disabled — should be silent
	msg, err := handleUC.Execute(ctx, "user1", "Hunter", "/jobs")
	if err != nil {
		t.Fatalf("unexpected error for /jobs: %v", err)
	}
	if msg.Text != "" {
		t.Fatalf("Disabled /jobs should be silently ignored, got %q", msg.Text)
	}

	repo.reports["user1"] = &domain.Report{
		UserID:      "user1",
		Name:        "Hunter",
		TotalPoints: 100,
	}

	// /job command is now disabled — should be silent
	msg, err = handleUC.Execute(ctx, "user1", "Hunter", "/job ranger")
	if err != nil {
		t.Fatalf("unexpected error for /job: %v", err)
	}
	if msg.Text != "" {
		t.Fatalf("Disabled /job should be silently ignored, got %q", msg.Text)
	}

	// /job command is now disabled — set job directly in repo for lapor test
	repo.reports["user1"].JobClass = "ranger"

	// /mystats doesn't show job anymore since mystats is disabled
	msg, err = handleUC.Execute(ctx, "user1", "Hunter", "/mystats")
	if err != nil {
		t.Fatalf("unexpected error for /mystats: %v", err)
	}
	if msg.Text != "" {
		t.Fatalf("Disabled /mystats should be silently ignored, got %q", msg.Text)
	}

	// /lapor still works and should show the job
	msg, err = handleUC.Execute(ctx, "user1", "Hunter", "/lapor")
	if err != nil {
		t.Fatalf("unexpected error for /lapor: %v", err)
	}
	if !containsSubstring(msg.Text, "Job: Ranger") {
		t.Fatalf("/lapor should include selected job, got %q", msg.Text)
	}
}

func TestHandleMessage_SetNameCommand(t *testing.T) {
	repo := &mockReportRepo{reports: make(map[string]*domain.Report)}
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, usecase.NewCancelReportUsecase(repo), updateNameUC, nil, usecase.NewBroadcastUpdateUsecase(), usecase.NewGetMotivationUsecase(), usecase.NewGetHelpUsecase())

	ctx := context.Background()

	// /setname command is now disabled — should get fallback
	msg, err := handleUC.Execute(ctx, "userVip", "OldName", "/setname King Budi")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if msg.Text != "" {
		t.Errorf("Disabled /setname should be silently ignored, got: %s", msg.Text)
	}

	msg, err = handleUC.Execute(ctx, "userVip", "King Budi", "/setname Budi Solo")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if msg.Text != "" {
		t.Errorf("Disabled /setname should be silently ignored, got: %s", msg.Text)
	}
}

func TestHandleMessage_LaporDoesNotUpdateName(t *testing.T) {
	repo := &mockReportRepo{reports: make(map[string]*domain.Report)}
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, usecase.NewCancelReportUsecase(repo), updateNameUC, nil, usecase.NewBroadcastUpdateUsecase(), usecase.NewGetMotivationUsecase(), usecase.NewGetHelpUsecase())

	ctx := context.Background()

	// setname is disabled — PushName is used directly in /lapor instead
	_, _ = handleUC.Execute(ctx, "user1", "InitialPushName", "/lapor")

	r := repo.reports["user1"]
	if r == nil || r.Name != "InitialPushName" {
		t.Errorf("Expected name from PushName 'InitialPushName', got '%s'", r.Name)
	}

	// Report again with different PushName — names don't auto-update either
	_, err := handleUC.Execute(ctx, "user1", "DifferentPushName", "/lapor")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	r = repo.reports["user1"]
	if r.Name != "InitialPushName" {
		t.Errorf("Expected name to remain from first report, but got '%s'", r.Name)
	}
}

func TestHandleMessage_LaporWithNonWhitespaceChars_ReturnsNoResponse(t *testing.T) {
	// /lapor must match a real command token so accidental slash words do not
	// create activity reports.
	repo := &mockReportRepo{reports: make(map[string]*domain.Report)}
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, usecase.NewCancelReportUsecase(repo), updateNameUC, nil, usecase.NewBroadcastUpdateUsecase(), usecase.NewGetMotivationUsecase(), usecase.NewGetHelpUsecase())

	ctx := context.Background()

	testCases := []string{
		"/laporhalo",
		"/lapor123",
		"/lapor.hari.ini",
		"/lapor-hari-ini",
	}

	for i, cmd := range testCases {
		repo.reports = make(map[string]*domain.Report)
		msg, err := handleUC.Execute(ctx, "user1", "User", cmd)
		if err != nil {
			t.Fatalf("Test %d: unexpected error for %q: %v", i, cmd, err)
		}
		if msg.Text != "" {
			t.Errorf("Test %d: %q should be silently ignored, got %q", i, cmd, msg.Text)
		}
		if repo.reports["user1"] != nil {
			t.Errorf("Test %d: %q should not create an activity report", i, cmd)
		}
	}
}

func TestHandleMessage_LaporCommandPriority(t *testing.T) {
	// Ensure that more specific commands (kemarin, sidequest) take priority
	// over the generic /lapor handler.
	repo := &mockReportRepo{reports: make(map[string]*domain.Report)}
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, usecase.NewCancelReportUsecase(repo), updateNameUC, nil, usecase.NewBroadcastUpdateUsecase(), usecase.NewGetMotivationUsecase(), usecase.NewGetHelpUsecase())

	ctx := context.Background()

	// Setup user data
	repo.reports["user1"] = &domain.Report{
		UserID:        "user1",
		Name:          "User",
		Streak:        3,
		ActivityCount: 3,
	}

	t.Run("/lapor-kemarin still routes to yesterday handler", func(t *testing.T) {
		msg, err := handleUC.Execute(ctx, "user1", "User", "/lapor-kemarin lari pagi")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Ensure the message was routed somewhere (non-empty) and is NOT the
		// standard /lapor response. The kemarin handler produces a comeback/
		// streak-based response that differs from the regular lapor output.
		if msg.Text == "" {
			t.Errorf("expected /lapor-kemarin to produce a response, got empty")
		}
	})

	t.Run("/lapor sidequest still routes to sidequest handler", func(t *testing.T) {
		msg, err := handleUC.Execute(ctx, "user1", "User", "/lapor sidequest push up 20x")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Side quest handler returns quest-related message, NOT lapor report
		if containsSubstring(msg.Text, "Laporan diterima") {
			t.Errorf("expected /lapor sidequest to route to sidequest handler, not lapor, got: %q", msg.Text)
		}
	})
}

func TestHandleMessage_MySideQuestCommand(t *testing.T) {
	repo := &mockReportRepo{reports: make(map[string]*domain.Report)}
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, usecase.NewCancelReportUsecase(repo), updateNameUC, nil, usecase.NewBroadcastUpdateUsecase(), usecase.NewGetMotivationUsecase(), usecase.NewGetHelpUsecase())

	ctx := context.Background()

	// 1. Setup user in repo
	repo.reports["user123"] = &domain.Report{
		UserID:      "user123",
		Name:        "TestUser",
		JobClass:    "fighter",
		TotalPoints: 10,
		Level:       1,
	}

	// /mysidequest command is now disabled — should be silent
	msg, err := handleUC.Execute(ctx, "user123", "TestUser", "/mysidequest")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if msg.Text != "" {
		t.Errorf("Disabled /mysidequest should be silently ignored, got '%s'", msg.Text)
	}
}

func TestHandleMessage_CancelSideQuestCommandsRouteToSideQuestCancel(t *testing.T) {
	tests := []string{
		"/cancel sidequest",
		"/cancel-sidequest",
		"/cancel all sidequest",
		"/cancel-all sidequest",
		"/cancel sidequest all",
		"#cancel sidequest",
		"#cancel-all sidequest",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			repo := &mockReportRepo{
				reports: make(map[string]*domain.Report),
				dailyCountByKind: map[string]int{
					domain.ActivityKindRegularReport: 1,
					domain.ActivityKindSideQuest:     2,
				},
			}
			repo.reports["user1"] = &domain.Report{
				UserID:             "user1",
				Name:               "User",
				ActivityCount:      1,
				TotalSideQuests:    2,
				SeasonalSideQuests: 2,
				LastReportDate:     time.Now(),
			}
			reportUC := usecase.NewReportActivityUsecase(repo)
			leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
			myStatsUC := usecase.NewGetMyStatsUsecase(repo)
			achievementsUC := usecase.NewGetAchievementsUsecase(repo)
			comebackUC := usecase.NewComebackChallengeUsecase(repo)
			updateNameUC := usecase.NewUpdateNameUsecase(repo)
			handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, usecase.NewCancelReportUsecase(repo), updateNameUC, nil, usecase.NewBroadcastUpdateUsecase(), usecase.NewGetMotivationUsecase(), usecase.NewGetHelpUsecase())

			msg, err := handleUC.Execute(context.Background(), "user1", "User", input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if msg.Text == "" {
				t.Fatal("expected cancel response")
			}
			if repo.deletedLogKind != domain.ActivityKindSideQuest {
				t.Fatalf("expected sidequest cancel, got %q", repo.deletedLogKind)
			}
			if repo.dailyCountByKind[domain.ActivityKindRegularReport] != 1 {
				t.Fatalf("regular report count should not change, got %d", repo.dailyCountByKind[domain.ActivityKindRegularReport])
			}
		})
	}
}
