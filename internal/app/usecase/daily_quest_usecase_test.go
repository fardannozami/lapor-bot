package usecase

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
	"github.com/fardannozami/whatsapp-gateway/internal/queue"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

type mockQuestRepo struct {
	domain.ReportRepository
	report         *domain.Report
	reports        []*domain.Report
	quests         map[string]string
	dailyCount     int
	upsertedReport *domain.Report
}

func (m *mockQuestRepo) GetReport(ctx context.Context, userID string) (*domain.Report, error) {
	return m.report, nil
}

func (m *mockQuestRepo) UpsertReport(ctx context.Context, report *domain.Report) error {
	m.upsertedReport = report
	m.report = report
	return nil
}

func (m *mockQuestRepo) UpsertReportWithActivity(ctx context.Context, report *domain.Report, activityDate time.Time) error {
	m.upsertedReport = report
	m.report = report
	return nil
}

func (m *mockQuestRepo) GetDailyQuest(ctx context.Context, userID, questDate string) (string, error) {
	return m.quests[userID+":"+questDate], nil
}

func (m *mockQuestRepo) SaveDailyQuest(ctx context.Context, userID, questDate, tasksJSON string) error {
	if m.quests == nil {
		m.quests = make(map[string]string)
	}
	m.quests[userID+":"+questDate] = tasksJSON
	return nil
}

func (m *mockQuestRepo) GetDailyActivityCount(ctx context.Context, userID string, date time.Time) (int, error) {
	return m.dailyCount, nil
}

func (m *mockQuestRepo) ResolveLIDToPhone(ctx context.Context, lid string) string {
	return lid
}

func (m *mockQuestRepo) GetAllReports(ctx context.Context) ([]*domain.Report, error) {
	if len(m.reports) > 0 {
		return m.reports, nil
	}
	if m.report != nil {
		return []*domain.Report{m.report}, nil
	}
	return nil, nil
}

func (m *mockQuestRepo) GetActiveGoal(ctx context.Context, userID string, now time.Time) (*domain.WeeklyGoal, error) {
	return nil, nil
}

func (m *mockQuestRepo) GetGoalActivities(ctx context.Context, userID string, startDate, endDate time.Time) ([]domain.GoalActivity, error) {
	return nil, nil
}

func (m *mockQuestRepo) RecordGoalActivity(ctx context.Context, userID string, activityDate time.Time, activityText string) (bool, error) {
	return false, nil
}

func TestDailyQuestViewAndComplete(t *testing.T) {
	now := time.Date(2026, 6, 15, 7, 0, 0, 0, time.UTC)
	todayStr := domain.GetToday(now).Format("2006-01-02")

	// Pre-populate report
	rep := &domain.Report{
		UserID:         "user1",
		Name:           "Alice",
		JobClass:       "fighter",
		Level:          5,
		TotalPoints:    200,
		LastReportDate: now.AddDate(0, 0, -2), // Has not reported today yet
	}

	repo := &mockQuestRepo{
		report: rep,
		quests: make(map[string]string),
	}

	// Initialize use cases
	questUC := NewDailyQuestUsecase(repo)
	reportUC := NewReportActivityUsecase(repo)

	// 1. View Quest (generates first time)
	viewMsg, err := questUC.ViewQuest(context.Background(), "user1", "Alice", now)
	if err != nil {
		t.Fatalf("unexpected view quest error: %v", err)
	}
	if !strings.Contains(viewMsg, "Side Quest Hari Ini - Alice") {
		t.Errorf("expected view message to contain Alice, got %q", viewMsg)
	}

	// Retrieve generated quest list
	qJSON, err := repo.GetDailyQuest(context.Background(), "user1", todayStr)
	if err != nil {
		t.Fatalf("failed to get stored quest JSON: %v", err)
	}
	var tasks []domain.QuestTask
	if err := json.Unmarshal([]byte(qJSON), &tasks); err != nil {
		t.Fatalf("failed to parse stored quest tasks: %v", err)
	}
	if len(tasks) != 4 {
		t.Fatalf("expected 4 generated tasks, got %d", len(tasks))
	}

	t.Logf("Generated tasks: %+v", tasks)

	// 2. Report below target should be rejected.
	progressMsg, err := questUC.UpdateProgress(context.Background(), "user1", "Alice", []string{"jalan 3999"}, reportUC, now)
	if err != nil {
		t.Fatalf("unexpected update progress error: %v", err)
	}
	if !strings.Contains(progressMsg, "Semangat") {
		t.Errorf("expected below-target side quest to be rejected with motivational message, got %q", progressMsg)
	}

	// 3. Complete walking side quest.
	compMsg, err := questUC.UpdateProgress(context.Background(), "user1", "Alice", []string{"jalan 4000"}, reportUC, now)
	if err != nil {
		t.Fatalf("unexpected complete task error: %v", err)
	}

	if !strings.Contains(compMsg, "SIDE QUEST BERHASIL DISELESAIKAN") {
		t.Errorf("expected success notification, got %q", compMsg)
	}
	if repo.upsertedReport.TotalSideQuests != 1 || repo.upsertedReport.SeasonalSideQuests != 1 {
		t.Errorf("expected side quest counters to increase, got lifetime=%d season=%d", repo.upsertedReport.TotalSideQuests, repo.upsertedReport.SeasonalSideQuests)
	}
	if repo.upsertedReport.TotalPoints != 203 {
		t.Errorf("expected side quest (easy) to add 3 points, got total=%d (expected 203)", repo.upsertedReport.TotalPoints)
	}
}

type fakeMessageClient struct {
	sentJID types.JID
	sentMsg *waE2E.Message
}

func (f *fakeMessageClient) SendMessage(ctx context.Context, to types.JID, msg *waE2E.Message, extra ...whatsmeow.SendRequestExtra) (whatsmeow.SendResponse, error) {
	f.sentJID = to
	f.sentMsg = msg
	return whatsmeow.SendResponse{}, nil
}

func TestSendDailyQuests(t *testing.T) {
	now := time.Date(2026, 6, 15, 7, 0, 0, 0, time.UTC)
	groupJID := types.NewJID("1234567890", types.GroupServer)

	reps := []*domain.Report{
		{
			UserID:         "628123456789",
			Name:           "Alice",
			JobClass:       "fighter",
			Level:          5,
			TotalPoints:    200,
			LastReportDate: now.AddDate(0, 0, -2),
		},
		{
			UserID:         "628987654321",
			Name:           "Bob",
			JobClass:       "", // General
			Level:          1,
			TotalPoints:    10,
			LastReportDate: now.AddDate(0, 0, -2),
		},
		{
			UserID:         "628111222333",
			Name:           "Charlie",
			JobClass:       "fighter",
			Level:          5,
			TotalPoints:    200,
			LastReportDate: now.AddDate(0, 0, -2),
		},
	}

	repo := &mockQuestRepo{
		reports: reps,
		quests:  make(map[string]string),
	}

	questUC := NewDailyQuestUsecase(repo)

	fakeClient := &fakeMessageClient{}
	sender := queue.NewTestSender(fakeClient, context.Background())
	sender.Start()
	defer sender.Shutdown(100 * time.Millisecond)

	err := questUC.SendDailyQuests(context.Background(), now, nil, sender, groupJID.String())
	if err != nil {
		t.Fatalf("unexpected SendDailyQuests error: %v", err)
	}

	// Wait a bit for the sender queue worker to process and call SendMessage
	time.Sleep(50 * time.Millisecond)

	if fakeClient.sentJID.String() != groupJID.String() {
		t.Errorf("expected JID %s, got %s", groupJID.String(), fakeClient.sentJID.String())
	}

	if fakeClient.sentMsg == nil || fakeClient.sentMsg.Conversation == nil {
		t.Fatalf("sent message does not contain conversation text")
	}

	text := *fakeClient.sentMsg.Conversation
	if !strings.Contains(text, "Selamat pagi, Hunters") {
		t.Errorf("expected text to contain title, got: %s", text)
	}
	if !strings.Contains(text, "#mysidequest") {
		t.Errorf("expected text to prompt #mysidequest, got: %s", text)
	}
	if strings.Contains(text, "@") || strings.Contains(text, "Alice") || strings.Contains(text, "Bob") || strings.Contains(text, "Charlie") {
		t.Errorf("daily side quest prompt should not mention or name users, got: %s", text)
	}
	if !strings.Contains(text, "2 hunter") {
		t.Errorf("expected only users with job to be counted, got: %s", text)
	}
}
