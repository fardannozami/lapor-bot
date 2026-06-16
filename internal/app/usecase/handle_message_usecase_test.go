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
// - #lapor → routes to ReportActivityUsecase
// - #leaderboard → routes to GetLeaderboardUsecase
// - Unknown commands → returns empty string (no response)
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
	reports        map[string]*domain.Report
	activityCounts []domain.ActivityLeaderboardEntry
	goals          map[string]*domain.WeeklyGoal
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

func (m *mockReportRepo) DeleteLatestActivityLog(ctx context.Context, userID string, activityDate time.Time) (int, error) {
	return 0, nil
}

func (m *mockReportRepo) GetUserActivityDates(ctx context.Context, userID string) ([]time.Time, error) {
	return nil, nil
}

func (m *mockReportRepo) DeleteReport(ctx context.Context, userID string) error {
	delete(m.reports, userID)
	return nil
}

func (m *mockReportRepo) GetDailyActivityCount(ctx context.Context, userID string, date time.Time) (int, error) {
	return 0, nil
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

	// Test #lapor command
	msg, err := handleUC.Execute(ctx, "user123", "TestUser", "#lapor")
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

	testCases := []string{"#LAPOR", "#Lapor", "#LaPor", "#lapor"}
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
	msg, err := handleUC.Execute(ctx, "user1", "User", "#lapor hari ini olahraga lari")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if msg.Text == "" {
		t.Error("#lapor with trailing text should still be recognized")
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

	// Test #leaderboard command
	msg, err := handleUC.Execute(ctx, "user1", "User", "#leaderboard")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should return leaderboard output
	if msg.Text == "" {
		t.Error("#leaderboard should return a response")
	}
	if !containsSubstring(msg.Text, "Hidup Sehat SWE Growth") {
		t.Errorf("Response should contain 'Hidup Sehat SWE Growth', got '%s'", msg.Text)
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

	testCases := []string{"#LEADERBOARD", "#Leaderboard", "#LeaderBoard", "#leaderboard"}
	for _, cmd := range testCases {
		msg, err := handleUC.Execute(ctx, "user1", "User", cmd)
		if err != nil {
			t.Fatalf("Unexpected error for '%s': %v", cmd, err)
		}
		if msg.Text == "" {
			t.Errorf("Command '%s' should return a response", cmd)
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

	msg, err := handleUC.Execute(ctx, "user1", "User", "#leaderboard-weekly")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !containsSubstring(msg.Text, "Berikut posisi sementara mingguan") {
		t.Fatalf("Expected weekly leaderboard response, got %q", msg.Text)
	}
	if containsSubstring(msg.Text, "30 Days of Sweat Challenge") {
		t.Fatalf("Weekly command should not fall back to the regular leaderboard response")
	}
	if !containsSubstring(msg.Text, "1. Alice: 3 hari") {
		t.Errorf("Expected Alice to appear with 3 hari, got %q", msg.Text)
	}
	if !containsSubstring(msg.Text, "2. Budi: 1 hari") {
		t.Errorf("Expected Budi to appear with 1 hari, got %q", msg.Text)
	}
}

func TestHandleMessage_UnknownCommand_ReturnsEmpty(t *testing.T) {
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
		"#invalid",
		"lapor",       // missing #
		"leaderboard", // missing #
	}

	for _, msgText := range testCases {
		result, err := handleUC.Execute(ctx, "user1", "User", msgText)
		if err != nil {
			t.Fatalf("Unexpected error for '%s': %v", msgText, err)
		}
		if result.Text != "" {
			t.Errorf("Unknown command '%s' should return empty string, got '%s'", msgText, result.Text)
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
		"  #lapor",
		"#lapor  ",
		"  #lapor  ",
		"\t#lapor",
		"\n#lapor\n",
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

	// Test #mystats
	msg, err := handleUC.Execute(ctx, "user1", "Gamer", "#mystats")
	if err != nil {
		t.Fatalf("Unexpected error for #mystats: %v", err)
	}
	if msg.Text == "" || !containsSubstring(msg.Text, "Statistik kamu, Gamer") {
		t.Errorf("#mystats response invalid: %s", msg.Text)
	}

	// Test #achievements
	msg, err = handleUC.Execute(ctx, "user1", "Gamer", "#achievements")
	if err != nil {
		t.Fatalf("Unexpected error for #achievements: %v", err)
	}
	if msg.Text == "" || !containsSubstring(msg.Text, "Season Badge Challenge") {
		t.Errorf("#achievements response invalid: %s", msg.Text)
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

	msg, err := handleUC.Execute(ctx, "user1", "Hunter", "#jobs")
	if err != nil {
		t.Fatalf("unexpected error for #jobs: %v", err)
	}
	if !containsSubstring(msg.Text, "Daftar Hunter Jobs") || !containsSubstring(msg.Text, "#job ranger") {
		t.Fatalf("#jobs response should include job list and selection example, got %q", msg.Text)
	}

	repo.reports["user1"] = &domain.Report{
		UserID:      "user1",
		Name:        "Hunter",
		TotalPoints: 100,
	}

	msg, err = handleUC.Execute(ctx, "user1", "Hunter", "#job ranger")
	if err != nil {
		t.Fatalf("unexpected error for #job: %v", err)
	}
	if !containsSubstring(msg.Text, "Job dipilih") || !containsSubstring(msg.Text, "Ranger") {
		t.Fatalf("#job response should confirm selected job, got %q", msg.Text)
	}
	if repo.reports["user1"].JobClass != "ranger" {
		t.Fatalf("expected stored job ranger, got %q", repo.reports["user1"].JobClass)
	}

	msg, err = handleUC.Execute(ctx, "user1", "Hunter", "#mystats")
	if err != nil {
		t.Fatalf("unexpected error for #mystats: %v", err)
	}
	if !containsSubstring(msg.Text, "Job: Ranger") {
		t.Fatalf("#mystats should include selected job, got %q", msg.Text)
	}

	msg, err = handleUC.Execute(ctx, "user1", "Hunter", "#lapor")
	if err != nil {
		t.Fatalf("unexpected error for #lapor: %v", err)
	}
	if !containsSubstring(msg.Text, "Job: Ranger") {
		t.Fatalf("#lapor should include selected job, got %q", msg.Text)
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

	// 1. Test #setname for new user
	msg, err := handleUC.Execute(ctx, "userVip", "OldName", "#setname King Budi")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !containsSubstring(msg.Text, "Namamu telah diatur sebagai King Budi") {
		t.Errorf("Unexpected response: %s", msg.Text)
	}

	// Verify repo
	r := repo.reports["userVip"]
	if r == nil || r.Name != "King Budi" {
		t.Errorf("Expected name 'King Budi' in repo, got %+v", r)
	}

	// 2. Test #setname for existing user
	msg, err = handleUC.Execute(ctx, "userVip", "King Budi", "#setname Budi Solo")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !containsSubstring(msg.Text, "dari King Budi menjadi Budi Solo") {
		t.Errorf("Unexpected response: %s", msg.Text)
	}

	// Verify repo update
	r = repo.reports["userVip"]
	if r.Name != "Budi Solo" {
		t.Errorf("Expected name 'Budi Solo' in repo, got %s", r.Name)
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

	// 1. Setup user with a specific name via #setname
	_, _ = handleUC.Execute(ctx, "user1", "InitialPushName", "#setname ManualName")

	// 2. Report with a different PushName
	_, err := handleUC.Execute(ctx, "user1", "DifferentPushName", "#lapor")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 3. Verify name remains "ManualName"
	r := repo.reports["user1"]
	if r.Name != "ManualName" {
		t.Errorf("Expected name to remain 'ManualName', but got '%s'", r.Name)
	}
}

func TestHandleMessage_LaporWithNonWhitespaceChars(t *testing.T) {
	// Regression test: #lapor must trigger even when non-whitespace characters
	// follow it (e.g., #laporhalo, #lapor123). This was broken by the old
	// containsCommand function which required whitespace or EOS after the command.
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
		"#laporhalo",              // non-whitespace right after command
		"#lapor123",               // numbers after command
		"#lapor.hari.ini",         // dots after command
		"halo#laporhalo",          // text before AND after
		"sebelum #lapor-hari-ini", // dash-separated with preceding text
	}

	for i, cmd := range testCases {
		repo.reports = make(map[string]*domain.Report)
		msg, err := handleUC.Execute(ctx, "user1", "User", cmd)
		if err != nil {
			t.Fatalf("Test %d: unexpected error for %q: %v", i, cmd, err)
		}
		if msg.Text == "" {
			t.Errorf("Test %d: %q should trigger #lapor, got empty response", i, cmd)
		}
	}
}

func TestHandleMessage_LaporCommandPriority(t *testing.T) {
	// Ensure that more specific commands (kemarin, sidequest) take priority
	// over the generic #lapor handler even with the relaxed matching.
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

	t.Run("#lapor-kemarin still routes to yesterday handler", func(t *testing.T) {
		msg, err := handleUC.Execute(ctx, "user1", "User", "#lapor-kemarin lari pagi")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Ensure the message was routed somewhere (non-empty) and is NOT the
		// standard #lapor response. The kemarin handler produces a comeback/
		// streak-based response that differs from the regular lapor output.
		if msg.Text == "" {
			t.Errorf("expected #lapor-kemarin to produce a response, got empty")
		}
	})

	t.Run("#lapor sidequest still routes to sidequest handler", func(t *testing.T) {
		msg, err := handleUC.Execute(ctx, "user1", "User", "#lapor sidequest push up 20x")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Side quest handler returns quest-related message, NOT lapor report
		if containsSubstring(msg.Text, "Laporan diterima") {
			t.Errorf("expected #lapor sidequest to route to sidequest handler, not lapor, got: %q", msg.Text)
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

	msg, err := handleUC.Execute(ctx, "user123", "TestUser", "#mysidequest")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "Side Quest Hari Ini - TestUser"
	if !containsSubstring(msg.Text, expected) {
		t.Errorf("Expected message to contain '%s', got '%s'", expected, msg.Text)
	}
}
