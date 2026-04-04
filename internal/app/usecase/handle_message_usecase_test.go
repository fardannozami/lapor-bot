package usecase_test

import (
	"context"
	"testing"

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
	reports map[string]*domain.Report
}

func (m *mockReportRepo) GetReport(ctx context.Context, userID string) (*domain.Report, error) {
	return m.reports[userID], nil
}

func (m *mockReportRepo) UpsertReport(ctx context.Context, report *domain.Report) error {
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

func (m *mockReportRepo) ResolveLIDToPhone(ctx context.Context, lid string) string {
	return lid
}

func (m *mockReportRepo) InitTable(ctx context.Context) error {
	return nil
}

func (m *mockReportRepo) GetInactiveUsers(ctx context.Context, days int) ([]*domain.Report, error) {
	return nil, nil
}

func TestHandleMessage_LaporCommand(t *testing.T) {
	repo := &mockReportRepo{reports: make(map[string]*domain.Report)}
	reportUC := usecase.NewReportActivityUsecase(repo)
	leaderboardUC := usecase.NewGetLeaderboardUsecase(repo)
	myStatsUC := usecase.NewGetMyStatsUsecase(repo)
	achievementsUC := usecase.NewGetAchievementsUsecase(repo)
	comebackUC := usecase.NewComebackChallengeUsecase(repo)
	updateNameUC := usecase.NewUpdateNameUsecase(repo)
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, updateNameUC)

	ctx := context.Background()

	// Test #lapor command
	msg, err := handleUC.Execute(ctx, "user123", "TestUser", "#lapor")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should route to report usecase and return response
	expected := "Laporan diterima, TestUser sudah berkeringat 1 hari. Lanjutkan 🔥 (streak 1 minggu)"
	if !containsSubstring(msg, expected) {
		t.Errorf("Expected message to contain '%s', got '%s'", expected, msg)
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
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, updateNameUC)

	ctx := context.Background()

	testCases := []string{"#LAPOR", "#Lapor", "#LaPor", "#lapor"}
	for _, cmd := range testCases {
		// Reset repo for each test
		repo.reports = make(map[string]*domain.Report)

		msg, err := handleUC.Execute(ctx, "user1", "User", cmd)
		if err != nil {
			t.Fatalf("Unexpected error for '%s': %v", cmd, err)
		}
		if msg == "" {
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
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, updateNameUC)

	ctx := context.Background()

	// Command with trailing text should still work
	msg, err := handleUC.Execute(ctx, "user1", "User", "#lapor hari ini olahraga lari")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if msg == "" {
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
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, updateNameUC)

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
	if msg == "" {
		t.Error("#leaderboard should return a response")
	}
	if !containsSubstring(msg, "30 Days of Sweat Challenge") {
		t.Errorf("Response should contain '30 Days of Sweat Challenge', got '%s'", msg)
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
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, updateNameUC)

	ctx := context.Background()

	testCases := []string{"#LEADERBOARD", "#Leaderboard", "#LeaderBoard", "#leaderboard"}
	for _, cmd := range testCases {
		msg, err := handleUC.Execute(ctx, "user1", "User", cmd)
		if err != nil {
			t.Fatalf("Unexpected error for '%s': %v", cmd, err)
		}
		if msg == "" {
			t.Errorf("Command '%s' should return a response", cmd)
		}
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
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, updateNameUC)

	ctx := context.Background()

	testCases := []string{
		"hello",
		"random message",
		"#invalid",
		"#help",
		"lapor",       // missing #
		"leaderboard", // missing #
	}

	for _, msg := range testCases {
		result, err := handleUC.Execute(ctx, "user1", "User", msg)
		if err != nil {
			t.Fatalf("Unexpected error for '%s': %v", msg, err)
		}
		if result != "" {
			t.Errorf("Unknown command '%s' should return empty string, got '%s'", msg, result)
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
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, updateNameUC)

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
		if msg == "" {
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
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, updateNameUC)

	ctx := context.Background()

	result, err := handleUC.Execute(ctx, "user1", "User", "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("Empty message should return empty string, got '%s'", result)
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
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, updateNameUC)

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
	if msg == "" || !containsSubstring(msg, "Statistik kamu, Gamer") {
		t.Errorf("#mystats response invalid: %s", msg)
	}

	// Test #achievements
	msg, err = handleUC.Execute(ctx, "user1", "Gamer", "#achievements")
	if err != nil {
		t.Fatalf("Unexpected error for #achievements: %v", err)
	}
	if msg == "" || !containsSubstring(msg, "Daftar Achievement") {
		t.Errorf("#achievements response invalid: %s", msg)
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
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, updateNameUC)

	ctx := context.Background()

	// 1. Test #setname for new user
	msg, err := handleUC.Execute(ctx, "userVip", "OldName", "#setname King Budi")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !containsSubstring(msg, "Namamu telah diatur sebagai King Budi") {
		t.Errorf("Unexpected response: %s", msg)
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
	if !containsSubstring(msg, "dari King Budi menjadi Budi Solo") {
		t.Errorf("Unexpected response: %s", msg)
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
	handleUC := usecase.NewHandleMessageUsecase(reportUC, leaderboardUC, myStatsUC, achievementsUC, comebackUC, updateNameUC)

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
