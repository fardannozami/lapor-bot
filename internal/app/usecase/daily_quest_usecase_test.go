package usecase

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type mockQuestRepo struct {
	domain.ReportRepository
	report         *domain.Report
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
