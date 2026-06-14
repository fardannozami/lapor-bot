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
	if !strings.Contains(viewMsg, "Quest Harian - Alice") {
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
	if len(tasks) != 3 {
		t.Fatalf("expected 3 generated tasks, got %d", len(tasks))
	}

	// First task is Push-up (Target: 10 + 5 = 15 x)
	t.Logf("Generated tasks: %+v", tasks)

	// 2. Report progress for Push-up
	progressMsg, err := questUC.UpdateProgress(context.Background(), "user1", "Alice", []string{"pushup 5"}, reportUC, now)
	if err != nil {
		t.Fatalf("unexpected update progress error: %v", err)
	}

	// Verify progress was updated
	if !strings.Contains(progressMsg, "Push-up: 5/") {
		t.Errorf("expected progress message to show updated reps, got %q", progressMsg)
	}

	// Verify auto-#lapor was triggered (as they hadn't reported yet today)
	if !strings.Contains(progressMsg, "AUTO #LAPOR DETECTED") {
		t.Errorf("expected response to show auto-#lapor trigger, got %q", progressMsg)
	}

	// Verify streak increased (last report date updated)
	if repo.upsertedReport.LastReportDate.Sub(now) > time.Second {
		t.Errorf("expected last report date to be now, got %v", repo.upsertedReport.LastReportDate)
	}

	// 3. Complete pushup task
	compMsg, err := questUC.UpdateProgress(context.Background(), "user1", "Alice", []string{"pushup 15"}, reportUC, now)
	if err != nil {
		t.Fatalf("unexpected complete task error: %v", err)
	}

	// Verify points awarded
	if !strings.Contains(compMsg, "QUEST BERHASIL DISELESAIKAN") {
		t.Errorf("expected success notification, got %q", compMsg)
	}

	// Pushup reward is 3 + 5/5 = 4 points
	if repo.upsertedReport.TotalPoints <= 200 {
		t.Errorf("expected points to increase, got %d", repo.upsertedReport.TotalPoints)
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

	if fakeClient.sentMsg == nil || fakeClient.sentMsg.ExtendedTextMessage == nil || fakeClient.sentMsg.ExtendedTextMessage.Text == nil {
		t.Fatalf("sent message does not contain ExtendedTextMessage text")
	}

	text := *fakeClient.sentMsg.ExtendedTextMessage.Text
	if !strings.Contains(text, "⚔️ *DAILY QUEST HARIAN HUNTER* ⚔️") {
		t.Errorf("expected text to contain title, got: %s", text)
	}
	if !strings.Contains(text, "👤 @628123456789") {
		t.Errorf("expected text to mention Alice, got: %s", text)
	}
	if !strings.Contains(text, "👤 @628987654321") {
		t.Errorf("expected text to mention Bob, got: %s", text)
	}
	if !strings.Contains(text, "Job: Fighter ⚔️ (Lv.5)") {
		t.Errorf("expected text to show Alice job and level, got: %s", text)
	}
	if !strings.Contains(text, "Job: - (General Quest)") {
		t.Errorf("expected text to show Bob general quest, got: %s", text)
	}

	// Verify mentions are populated correctly in ContextInfo
	contextInfo := fakeClient.sentMsg.ExtendedTextMessage.ContextInfo
	if contextInfo == nil {
		t.Fatalf("expected ContextInfo in ExtendedTextMessage to be set")
	}
	expectedMentions := []string{"628123456789@s.whatsapp.net", "628987654321@s.whatsapp.net"}
	if len(contextInfo.MentionedJID) != len(expectedMentions) {
		t.Errorf("expected %d mentions, got %d", len(expectedMentions), len(contextInfo.MentionedJID))
	} else {
		for i, m := range contextInfo.MentionedJID {
			if m != expectedMentions[i] {
				t.Errorf("expected mention %q at index %d, got %q", expectedMentions[i], i, m)
			}
		}
	}
}
