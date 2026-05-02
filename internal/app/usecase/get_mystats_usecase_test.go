package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type myStatsRepoStub struct {
	report  *domain.Report
	entries map[string][]domain.ActivityLeaderboardEntry
	calls   [][2]time.Time
}

func (r *myStatsRepoStub) GetReport(ctx context.Context, userID string) (*domain.Report, error) {
	return r.report, nil
}

func (r *myStatsRepoStub) UpsertReport(ctx context.Context, report *domain.Report) error {
	return nil
}

func (r *myStatsRepoStub) UpsertReportWithActivity(ctx context.Context, report *domain.Report, activityDate time.Time) error {
	return nil
}

func (r *myStatsRepoStub) GetAllReports(ctx context.Context) ([]*domain.Report, error) {
	return nil, nil
}

func (r *myStatsRepoStub) LogActivity(ctx context.Context, userID string, activityDate time.Time) error {
	return nil
}

func (r *myStatsRepoStub) GetActivityCountsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]domain.ActivityLeaderboardEntry, error) {
	r.calls = append(r.calls, [2]time.Time{startDate, endDate})
	key := startDate.Format(time.DateOnly) + "|" + endDate.Format(time.DateOnly)
	return r.entries[key], nil
}

func (r *myStatsRepoStub) GetInactiveUsers(ctx context.Context, days int) ([]*domain.Report, error) {
	return nil, nil
}

func (r *myStatsRepoStub) ResetAllReports(ctx context.Context) error {
	return nil
}

func (r *myStatsRepoStub) InitTable(ctx context.Context) error {
	return nil
}

func (r *myStatsRepoStub) ResolveLIDToPhone(ctx context.Context, lid string) string {
	return lid
}

func (r *myStatsRepoStub) UpsertStravaAccount(ctx context.Context, account *domain.StravaAccount) error {
	return nil
}

func (r *myStatsRepoStub) GetStravaAccountByAthleteID(ctx context.Context, athleteID int64) (*domain.StravaAccount, error) {
	return nil, nil
}

func (r *myStatsRepoStub) GetStravaAccountByUserID(ctx context.Context, userID string) (*domain.StravaAccount, error) {
	return nil, nil
}

func TestGetMyStatsUsecase_IncludesSeasonAndWeeklyCounts(t *testing.T) {
	fixedNow := time.Date(2026, time.June, 10, 9, 0, 0, 0, time.UTC)
	weekStart := time.Date(2026, time.June, 7, 0, 0, 0, 0, time.UTC)
	weekEnd := time.Date(2026, time.June, 14, 0, 0, 0, 0, time.UTC)
	seasonStart := time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC)
	seasonEnd := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.FixedZone("WIB", 7*3600))

	repo := &myStatsRepoStub{
		report: &domain.Report{
			UserID:        "user1",
			Name:          "Gamer",
			Streak:        10,
			ActivityCount: 10,
			MaxStreak:     12,
			TotalPoints:   50,
		},
		entries: map[string][]domain.ActivityLeaderboardEntry{
			weekStart.Format(time.DateOnly) + "|" + weekEnd.Format(time.DateOnly): {
				{UserID: "user1", Name: "Gamer", ActivityCount: 3},
			},
			seasonStart.Format(time.DateOnly) + "|" + seasonEnd.Format(time.DateOnly): {
				{UserID: "user1", Name: "Gamer", ActivityCount: 8},
			},
		},
	}

	uc := NewGetMyStatsUsecase(repo)
	uc.now = func() time.Time { return fixedNow }

	result, err := uc.Execute(context.Background(), "user1", "Gamer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.calls) != 2 {
		t.Fatalf("expected 2 activity range calls, got %d", len(repo.calls))
	}
	if !repo.calls[0][0].Equal(weekStart) || !repo.calls[0][1].Equal(weekEnd) {
		t.Fatalf("expected weekly range %v - %v, got %v - %v", weekStart, weekEnd, repo.calls[0][0], repo.calls[0][1])
	}
	if !repo.calls[1][0].Equal(seasonStart) || !repo.calls[1][1].Equal(seasonEnd) {
		t.Fatalf("expected season range %v - %v, got %v - %v", seasonStart, seasonEnd, repo.calls[1][0], repo.calls[1][1])
	}
	if !contains(result, "🗓️ Total hari season: 8") {
		t.Fatalf("expected season count in response, got %q", result)
	}
	if !contains(result, "📆 Total mingguan: 3") {
		t.Fatalf("expected weekly count in response, got %q", result)
	}
}
