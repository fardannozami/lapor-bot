package usecase_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/fardannozami/whatsapp-gateway/internal/app/usecase"
	"github.com/fardannozami/whatsapp-gateway/internal/domain"
	"github.com/fardannozami/whatsapp-gateway/internal/infra/sqlite"
)

// Integration test (real SQLite, real event-ledger write path) for the
// same-day fairness invariant: multiple accepted reports on one calendar day
// must yield exactly one active day, with repeat reports earning half XP.

func setupSameDayDB(t *testing.T) (*sql.DB, *sqlite.ReportRepository) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	repo := sqlite.NewReportRepository(db)
	if err := repo.InitTable(context.Background()); err != nil {
		t.Fatalf("init: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, repo
}

func activityLogCounts(t *testing.T, ctx context.Context, db *sql.DB, userID string, date time.Time) (reportCount, regular, sidequest int) {
	t.Helper()
	err := db.QueryRowContext(ctx, `
		SELECT report_count, regular_report_count, sidequest_count
		FROM activity_logs WHERE user_id = ? AND activity_date = ?
	`, userID, date.Format(time.DateOnly)).Scan(&reportCount, &regular, &sidequest)
	if err == sql.ErrNoRows {
		return 0, 0, 0
	}
	if err != nil {
		t.Fatalf("query activity_logs: %v", err)
	}
	return reportCount, regular, sidequest
}

func TestIntegration_SameDayTripleReport_OneActiveDay(t *testing.T) {
	ctx := context.Background()
	db, repo := setupSameDayDB(t)
	uc := usecase.NewReportActivityUsecase(repo)

	for i := 1; i <= 3; i++ {
		if _, err := uc.ExecuteWithMessage(ctx, "u1", "Alice", "lari pagi", nil); err != nil {
			t.Fatalf("report %d: %v", i, err)
		}
	}

	r, err := repo.GetReport(ctx, "u1")
	if err != nil {
		t.Fatalf("get report: %v", err)
	}
	if r.ActivityCount != 1 || r.SeasonalActivityCount != 1 {
		t.Errorf("same-day repeats must yield one active day, got ActivityCount=%d Seasonal=%d", r.ActivityCount, r.SeasonalActivityCount)
	}
	if r.Streak != 1 {
		t.Errorf("same-day repeats must not change streak, got %d", r.Streak)
	}

	today := domain.GetToday(time.Now())
	reportCount, regular, sidequest := activityLogCounts(t, ctx, db, "u1", today)
	if reportCount != 3 || regular != 3 || sidequest != 0 {
		t.Errorf("expected one activity_logs row with report_count=3 regular=3, got rc=%d reg=%d sq=%d", reportCount, regular, sidequest)
	}

	// Weekly leaderboard counts distinct active days, not reports.
	weekStart := domain.GetStartOfISOWeekStrict(time.Now())
	entries, err := repo.GetActivityCountsByDateRange(ctx, weekStart, weekStart.AddDate(0, 0, 7))
	if err != nil {
		t.Fatalf("weekly counts: %v", err)
	}
	if len(entries) != 1 || entries[0].ActivityCount != 1 {
		t.Errorf("weekly leaderboard must count 1 active day, got %+v", entries)
	}
}

func TestIntegration_SameDayReportAndSideQuests_FairPoints(t *testing.T) {
	ctx := context.Background()
	db, repo := setupSameDayDB(t)
	uc := usecase.NewReportActivityUsecase(repo)

	if _, err := uc.ExecuteWithMessage(ctx, "u2", "Bob", "gym", nil); err != nil {
		t.Fatalf("regular report: %v", err)
	}
	first, err := repo.GetReport(ctx, "u2")
	if err != nil {
		t.Fatalf("get report: %v", err)
	}

	now := time.Now()
	if _, err := uc.ExecuteSideQuest(ctx, "u2", "Bob", "Side quest: Jalan Kaki", 1, 2, now); err != nil {
		t.Fatalf("side quest 1: %v", err)
	}
	if _, err := uc.ExecuteSideQuest(ctx, "u2", "Bob", "Side quest: Chair Squat", 1, 3, now); err != nil {
		t.Fatalf("side quest 2: %v", err)
	}

	r, err := repo.GetReport(ctx, "u2")
	if err != nil {
		t.Fatalf("get report: %v", err)
	}
	if r.ActivityCount != 1 || r.SeasonalActivityCount != 1 {
		t.Errorf("side quests must not add active days, got ActivityCount=%d Seasonal=%d", r.ActivityCount, r.SeasonalActivityCount)
	}
	if got := r.SeasonalPoints - first.SeasonalPoints; got != 5 {
		t.Errorf("side quests award difficulty points only (2+3=5), got +%d", got)
	}
	if r.TotalSideQuests != 2 || r.SeasonalSideQuests != 2 {
		t.Errorf("expected 2 side quests, got %d/%d", r.TotalSideQuests, r.SeasonalSideQuests)
	}

	today := domain.GetToday(now)
	reportCount, regular, sidequest := activityLogCounts(t, ctx, db, "u2", today)
	if reportCount != 3 || regular != 1 || sidequest != 2 {
		t.Errorf("expected rc=3 reg=1 sq=2, got rc=%d reg=%d sq=%d", reportCount, regular, sidequest)
	}
}
