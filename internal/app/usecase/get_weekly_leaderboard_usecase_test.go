package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type weeklyLeaderboardRepoStub struct {
	entries []domain.ActivityLeaderboardEntry
	start   time.Time
	end     time.Time
}

func (r *weeklyLeaderboardRepoStub) GetReport(ctx context.Context, userID string) (*domain.Report, error) {
	return nil, nil
}

func (r *weeklyLeaderboardRepoStub) UpsertReport(ctx context.Context, report *domain.Report) error {
	return nil
}

func (r *weeklyLeaderboardRepoStub) UpsertReportWithActivity(ctx context.Context, report *domain.Report, activityDate time.Time) error {
	return nil
}

func (r *weeklyLeaderboardRepoStub) GetAllReports(ctx context.Context) ([]*domain.Report, error) {
	return nil, nil
}

func (r *weeklyLeaderboardRepoStub) LogActivity(ctx context.Context, userID string, activityDate time.Time) error {
	return nil
}

func (r *weeklyLeaderboardRepoStub) GetActivityCountsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]domain.ActivityLeaderboardEntry, error) {
	r.start = startDate
	r.end = endDate
	return r.entries, nil
}

func (r *weeklyLeaderboardRepoStub) GetInactiveUsers(ctx context.Context, days int) ([]*domain.Report, error) {
	return nil, nil
}

func (r *weeklyLeaderboardRepoStub) ResetAllReports(ctx context.Context) error {
	return nil
}

func (r *weeklyLeaderboardRepoStub) InitTable(ctx context.Context) error {
	return nil
}

func (r *weeklyLeaderboardRepoStub) ResolveLIDToPhone(ctx context.Context, lid string) string {
	return lid
}

func (r *weeklyLeaderboardRepoStub) UpsertStravaAccount(ctx context.Context, account *domain.StravaAccount) error {
	return nil
}

func (r *weeklyLeaderboardRepoStub) GetStravaAccountByAthleteID(ctx context.Context, athleteID int64) (*domain.StravaAccount, error) {
	return nil, nil
}

func (r *weeklyLeaderboardRepoStub) GetStravaAccountByUserID(ctx context.Context, userID string) (*domain.StravaAccount, error) {
	return nil, nil
}

func TestGetWeeklyLeaderboardUsecase_UsesSundayWeekRange(t *testing.T) {
	repo := &weeklyLeaderboardRepoStub{
		entries: []domain.ActivityLeaderboardEntry{
			{Name: "Cici", ActivityCount: 2},
			{Name: "Ayu", ActivityCount: 2},
			{Name: "Budi", ActivityCount: 4},
		},
	}
	uc := NewGetWeeklyLeaderboardUsecase(repo)
	fixedNow := time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC)
	uc.now = func() time.Time { return fixedNow }

	result, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedStart := time.Date(2026, time.April, 26, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2026, time.May, 3, 0, 0, 0, 0, time.UTC)
	if !repo.start.Equal(expectedStart) {
		t.Fatalf("expected start %v, got %v", expectedStart, repo.start)
	}
	if !repo.end.Equal(expectedEnd) {
		t.Fatalf("expected end %v, got %v", expectedEnd, repo.end)
	}

	if !contains(result, "Periode: 26-04-2026 s/d sebelum 03-05-2026") {
		t.Fatalf("expected weekly period in response, got %q", result)
	}
	if !contains(result, "1. Budi: 4 hari") {
		t.Fatalf("expected Budi to rank first, got %q", result)
	}
	if !contains(result, "2. Ayu: 2 hari") || !contains(result, "3. Cici: 2 hari") {
		t.Fatalf("expected ties to be sorted by name, got %q", result)
	}
	if !contains(result, "Berikut posisi sementara mingguan") {
		t.Fatalf("expected weekly leaderboard intro, got %q", result)
	}
}

func contains(s, substr string) bool {
	return len(substr) == 0 || len(s) >= len(substr) && index(s, substr) >= 0
}

func index(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
