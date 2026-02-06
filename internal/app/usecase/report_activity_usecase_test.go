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

func TestReportActivityUsecase_Execute(t *testing.T) {
	repo := &mockRepo{reports: make(map[string]*domain.Report)}
	uc := usecase.NewReportActivityUsecase(repo)
	ctx := context.Background()

	// 1. Initial report
	_, _ = uc.Execute(ctx, "user1", "User One")
	r := repo.reports["user1"]
	if r.Streak != 1 || r.ActivityCount != 1 {
		t.Errorf("Expected streak 1 and count 1, got streak %d and count %d", r.Streak, r.ActivityCount)
	}

	// 2. Report next day (simulated)
	r.LastReportDate = time.Now().AddDate(0, 0, -1)
	_, _ = uc.Execute(ctx, "user1", "User One")
	if r.Streak != 2 || r.ActivityCount != 2 {
		t.Errorf("Expected streak 2 and count 2, got streak %d and count %d", r.Streak, r.ActivityCount)
	}

	// 3. Miss a day and report
	r.LastReportDate = time.Now().AddDate(0, 0, -2)
	_, _ = uc.Execute(ctx, "user1", "User One")
	if r.Streak != 1 || r.ActivityCount != 3 {
		t.Errorf("Expected streak 1 and count 3, got streak %d and count %d", r.Streak, r.ActivityCount)
	}
}
