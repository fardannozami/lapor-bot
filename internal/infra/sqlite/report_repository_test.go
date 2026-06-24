package sqlite_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
	"github.com/fardannozami/whatsapp-gateway/internal/infra/sqlite"
)

// =============================================================================
// SQLITE REPORT REPOSITORY TESTS
// =============================================================================
//
// Tests the SQLite implementation of ReportRepository using in-memory database.
// Each test gets a fresh database to ensure isolation.
//
// =============================================================================

func setupTestDB(t *testing.T) (*sql.DB, *sqlite.ReportRepository, func()) {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	repo := sqlite.NewReportRepository(db)
	if err := repo.InitTable(context.Background()); err != nil {
		t.Fatalf("Failed to initialize table: %v", err)
	}

	cleanup := func() {
		db.Close()
	}

	return db, repo, cleanup
}

func TestReportRepository_GetReport_NotFound(t *testing.T) {
	_, repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	report, err := repo.GetReport(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if report != nil {
		t.Errorf("Expected nil for nonexistent user, got %+v", report)
	}
}

func TestReportRepository_UpsertReport_Insert(t *testing.T) {
	_, repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	report := &domain.Report{
		UserID:         "user123",
		Name:           "Alice",
		Streak:         5,
		ActivityCount:  10,
		LastReportDate: now,
	}

	// Insert new report
	err := repo.UpsertReport(ctx, report)
	if err != nil {
		t.Fatalf("Failed to insert report: %v", err)
	}

	// Verify it was inserted
	got, err := repo.GetReport(ctx, "user123")
	if err != nil {
		t.Fatalf("Failed to get report: %v", err)
	}
	if got == nil {
		t.Fatal("Expected report to be found")
	}

	if got.UserID != "user123" {
		t.Errorf("UserID: expected 'user123', got '%s'", got.UserID)
	}
	if got.Name != "Alice" {
		t.Errorf("Name: expected 'Alice', got '%s'", got.Name)
	}
	if got.Streak != 5 {
		t.Errorf("Streak: expected 5, got %d", got.Streak)
	}
	if got.ActivityCount != 10 {
		t.Errorf("ActivityCount: expected 10, got %d", got.ActivityCount)
	}
}

func TestReportRepository_UpsertReportWithActivityEvent_WritesLedgerAndProjections(t *testing.T) {
	db, repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Date(2026, time.September, 2, 10, 0, 0, 0, time.UTC)
	report := &domain.Report{
		UserID:                "user123",
		Name:                  "Alice",
		Streak:                1,
		ActivityCount:         1,
		SeasonalActivityCount: 1,
		LastReportDate:        now,
		TotalPoints:           7,
		SeasonalPoints:        7,
	}
	event := domain.ReportActivityEvent{
		EventID:             "event-1",
		UserID:              "user123",
		SeasonNumber:        2,
		Kind:                domain.ActivityKindSideQuest,
		ActivityDate:        domain.GetToday(now),
		OccurredAt:          now,
		PointsDelta:         7,
		RegularCountDelta:   0,
		SideQuestCountDelta: 2,
		RuleVersion:         1,
		Source:              "whatsapp",
		ActivityText:        "Side quest: pushup, walk",
		MetadataJSON:        "{}",
	}

	if err := repo.UpsertReportWithActivityEvent(ctx, report, event); err != nil {
		t.Fatalf("UpsertReportWithActivityEvent() error = %v", err)
	}

	var ledgerCount, pointsDelta, sideQuestDelta int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(points_delta), 0), COALESCE(SUM(sidequest_count_delta), 0)
		FROM report_events
		WHERE user_id = ? AND season_number = ?
	`, "user123", 2).Scan(&ledgerCount, &pointsDelta, &sideQuestDelta); err != nil {
		t.Fatalf("query report_events: %v", err)
	}
	if ledgerCount != 1 || pointsDelta != 7 || sideQuestDelta != 2 {
		t.Fatalf("unexpected ledger values count=%d points=%d sidequests=%d", ledgerCount, pointsDelta, sideQuestDelta)
	}

	var activitySideQuests int
	if err := db.QueryRowContext(ctx, `
		SELECT sidequest_count
		FROM activity_logs
		WHERE user_id = ? AND activity_date = ?
	`, "user123", event.ActivityDate.Format(time.DateOnly)).Scan(&activitySideQuests); err != nil {
		t.Fatalf("query activity_logs: %v", err)
	}
	if activitySideQuests != 2 {
		t.Fatalf("expected activity_logs sidequest_count=2, got %d", activitySideQuests)
	}

	var dailyPoints, dailySideQuests int
	if err := db.QueryRowContext(ctx, `
		SELECT total_points, sidequest_count
		FROM user_daily_activity
		WHERE user_id = ? AND season_number = ? AND activity_date = ?
	`, "user123", 2, event.ActivityDate.Format(time.DateOnly)).Scan(&dailyPoints, &dailySideQuests); err != nil {
		t.Fatalf("query user_daily_activity: %v", err)
	}
	if dailyPoints != 7 || dailySideQuests != 2 {
		t.Fatalf("unexpected daily projection points=%d sidequests=%d", dailyPoints, dailySideQuests)
	}

	var seasonPoints, seasonSideQuests, activeDays int
	if err := db.QueryRowContext(ctx, `
		SELECT total_points, sidequest_reports, active_days
		FROM user_season_stats
		WHERE user_id = ? AND season_number = ?
	`, "user123", 2).Scan(&seasonPoints, &seasonSideQuests, &activeDays); err != nil {
		t.Fatalf("query user_season_stats: %v", err)
	}
	if seasonPoints != 7 || seasonSideQuests != 2 || activeDays != 1 {
		t.Fatalf("unexpected season projection points=%d sidequests=%d activeDays=%d", seasonPoints, seasonSideQuests, activeDays)
	}

	if err := repo.UpsertReportWithActivityEvent(ctx, report, event); err != nil {
		t.Fatalf("duplicate UpsertReportWithActivityEvent() error = %v", err)
	}

	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(points_delta), 0), COALESCE(SUM(sidequest_count_delta), 0)
		FROM report_events
		WHERE user_id = ? AND season_number = ?
	`, "user123", 2).Scan(&ledgerCount, &pointsDelta, &sideQuestDelta); err != nil {
		t.Fatalf("query duplicate report_events: %v", err)
	}
	if ledgerCount != 1 || pointsDelta != 7 || sideQuestDelta != 2 {
		t.Fatalf("duplicate event should not change ledger count=%d points=%d sidequests=%d", ledgerCount, pointsDelta, sideQuestDelta)
	}

	if err := db.QueryRowContext(ctx, `
		SELECT sidequest_count
		FROM activity_logs
		WHERE user_id = ? AND activity_date = ?
	`, "user123", event.ActivityDate.Format(time.DateOnly)).Scan(&activitySideQuests); err != nil {
		t.Fatalf("query duplicate activity_logs: %v", err)
	}
	if activitySideQuests != 2 {
		t.Fatalf("duplicate event should not inflate activity_logs sidequest_count, got %d", activitySideQuests)
	}

	if err := db.QueryRowContext(ctx, `
		SELECT total_points, sidequest_count
		FROM user_daily_activity
		WHERE user_id = ? AND season_number = ? AND activity_date = ?
	`, "user123", 2, event.ActivityDate.Format(time.DateOnly)).Scan(&dailyPoints, &dailySideQuests); err != nil {
		t.Fatalf("query duplicate user_daily_activity: %v", err)
	}
	if dailyPoints != 7 || dailySideQuests != 2 {
		t.Fatalf("duplicate event should not inflate daily projection points=%d sidequests=%d", dailyPoints, dailySideQuests)
	}
}

func TestReportRepository_UpsertReport_Update(t *testing.T) {
	_, repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	// Insert initial report
	report := &domain.Report{
		UserID:         "user123",
		Name:           "Alice",
		Streak:         5,
		ActivityCount:  10,
		LastReportDate: now,
	}
	if err := repo.UpsertReport(ctx, report); err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	// Update the report
	report.Streak = 6
	report.ActivityCount = 11
	report.Name = "Alice Updated"
	if err := repo.UpsertReport(ctx, report); err != nil {
		t.Fatalf("Failed to update: %v", err)
	}

	// Verify update
	got, err := repo.GetReport(ctx, "user123")
	if err != nil {
		t.Fatalf("Failed to get report: %v", err)
	}

	if got.Streak != 6 {
		t.Errorf("Streak: expected 6, got %d", got.Streak)
	}
	if got.ActivityCount != 11 {
		t.Errorf("ActivityCount: expected 11, got %d", got.ActivityCount)
	}
	if got.Name != "Alice Updated" {
		t.Errorf("Name: expected 'Alice Updated', got '%s'", got.Name)
	}
}

func TestReportRepository_GetAllReports_Empty(t *testing.T) {
	_, repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	reports, err := repo.GetAllReports(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(reports) != 0 {
		t.Errorf("Expected empty slice, got %d reports", len(reports))
	}
}

func TestReportRepository_GetAllReports_OrderByActivityCount(t *testing.T) {
	_, repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	// Insert reports with different activity counts (out of order)
	reports := []*domain.Report{
		{UserID: "user1", Name: "Low", Streak: 1, ActivityCount: 5, LastReportDate: now},
		{UserID: "user2", Name: "High", Streak: 3, ActivityCount: 30, LastReportDate: now},
		{UserID: "user3", Name: "Medium", Streak: 2, ActivityCount: 15, LastReportDate: now},
	}

	for _, r := range reports {
		if err := repo.UpsertReport(ctx, r); err != nil {
			t.Fatalf("Failed to insert: %v", err)
		}
	}

	// Get all reports (should be ordered by activity_count DESC)
	got, err := repo.GetAllReports(ctx)
	if err != nil {
		t.Fatalf("Failed to get all reports: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("Expected 3 reports, got %d", len(got))
	}

	// Verify order: High (30) > Medium (15) > Low (5)
	if got[0].Name != "High" || got[0].ActivityCount != 30 {
		t.Errorf("First should be 'High' with 30, got '%s' with %d", got[0].Name, got[0].ActivityCount)
	}
	if got[1].Name != "Medium" || got[1].ActivityCount != 15 {
		t.Errorf("Second should be 'Medium' with 15, got '%s' with %d", got[1].Name, got[1].ActivityCount)
	}
	if got[2].Name != "Low" || got[2].ActivityCount != 5 {
		t.Errorf("Third should be 'Low' with 5, got '%s' with %d", got[2].Name, got[2].ActivityCount)
	}
}

func TestReportRepository_GetActivityCountsByDateRange(t *testing.T) {
	_, repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	base := time.Date(2026, time.April, 26, 0, 0, 0, 0, time.UTC)

	reports := []*domain.Report{
		{UserID: "user1", Name: "Alice", LastReportDate: base},
		{UserID: "user2", Name: "Budi", LastReportDate: base},
	}
	for _, report := range reports {
		if err := repo.UpsertReport(ctx, report); err != nil {
			t.Fatalf("Failed to insert report: %v", err)
		}
	}

	activityDates := []struct {
		userID string
		date   time.Time
	}{
		{userID: "user1", date: base},
		{userID: "user1", date: base.AddDate(0, 0, 2)},
		{userID: "user2", date: base.AddDate(0, 0, 3)},
		{userID: "user2", date: base.AddDate(0, 0, 7)},
	}
	for _, activity := range activityDates {
		if err := repo.LogActivity(ctx, activity.userID, activity.date); err != nil {
			t.Fatalf("Failed to log activity: %v", err)
		}
	}

	entries, err := repo.GetActivityCountsByDateRange(ctx, base, base.AddDate(0, 0, 7))
	if err != nil {
		t.Fatalf("Failed to get activity counts: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(entries))
	}
	if entries[0].Name != "Alice" || entries[0].ActivityCount != 2 {
		t.Fatalf("Expected Alice with 2 activities first, got %+v", entries[0])
	}
	if entries[1].Name != "Budi" || entries[1].ActivityCount != 1 {
		t.Fatalf("Expected Budi with 1 activity second, got %+v", entries[1])
	}
}

func TestReportRepository_InitTable_Idempotent(t *testing.T) {
	db, repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// InitTable was already called in setup, call it again
	if err := repo.InitTable(ctx); err != nil {
		t.Fatalf("Second InitTable should not fail: %v", err)
	}

	// Insert a report and verify table still works
	report := &domain.Report{
		UserID:         "user1",
		Name:           "Test",
		Streak:         1,
		ActivityCount:  1,
		LastReportDate: time.Now(),
	}
	if err := repo.UpsertReport(ctx, report); err != nil {
		t.Fatalf("Insert after double InitTable failed: %v", err)
	}

	// Check that we can use a fresh repo with the same db
	repo2 := sqlite.NewReportRepository(db)
	got, err := repo2.GetReport(ctx, "user1")
	if err != nil {
		t.Fatalf("Get with new repo failed: %v", err)
	}
	if got == nil {
		t.Error("Report should exist")
	}
}

func TestReportRepository_DatePersistence(t *testing.T) {
	_, repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Use a specific date to test RFC3339 parsing
	testDate := time.Date(2026, 2, 6, 15, 30, 45, 0, time.UTC)

	report := &domain.Report{
		UserID:         "user1",
		Name:           "Test",
		Streak:         1,
		ActivityCount:  1,
		LastReportDate: testDate,
	}
	if err := repo.UpsertReport(ctx, report); err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	got, err := repo.GetReport(ctx, "user1")
	if err != nil {
		t.Fatalf("Failed to get: %v", err)
	}

	// Compare times (allow for timezone normalization)
	if !got.LastReportDate.Equal(testDate) {
		t.Errorf("Date not preserved: expected %v, got %v", testDate, got.LastReportDate)
	}
}

func TestReportRepository_ResolveLIDToPhone_NotFound(t *testing.T) {
	_, repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// LID not in mapping should return the input unchanged
	result := repo.ResolveLIDToPhone(ctx, "some_lid_12345")
	if result != "some_lid_12345" {
		t.Errorf("Expected input returned unchanged, got '%s'", result)
	}
}

func TestReportRepository_ResolveLIDToPhone_Found(t *testing.T) {
	db, repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create the lid_map table and insert a mapping
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS whatsmeow_lid_map (
			lid TEXT PRIMARY KEY,
			pn TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create lid_map table: %v", err)
	}

	_, err = db.ExecContext(ctx, `INSERT INTO whatsmeow_lid_map (lid, pn) VALUES (?, ?)`, "lid123", "628123456789")
	if err != nil {
		t.Fatalf("Failed to insert mapping: %v", err)
	}

	// Now resolve should return the phone number
	result := repo.ResolveLIDToPhone(ctx, "lid123")
	if result != "628123456789" {
		t.Errorf("Expected '628123456789', got '%s'", result)
	}
}

func TestReportRepository_ConcurrentAccess(t *testing.T) {
	_, repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	// Insert initial report
	report := &domain.Report{
		UserID:         "user1",
		Name:           "Test",
		Streak:         0,
		ActivityCount:  0,
		LastReportDate: now,
	}
	if err := repo.UpsertReport(ctx, report); err != nil {
		t.Fatalf("Initial insert failed: %v", err)
	}

	// Simulate concurrent increments (in-memory SQLite is not truly concurrent,
	// but this tests the upsert behavior)
	for i := 0; i < 10; i++ {
		got, err := repo.GetReport(ctx, "user1")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		got.ActivityCount++
		if err := repo.UpsertReport(ctx, got); err != nil {
			t.Fatalf("Update failed: %v", err)
		}
	}

	final, err := repo.GetReport(ctx, "user1")
	if err != nil {
		t.Fatalf("Final get failed: %v", err)
	}
	if final.ActivityCount != 10 {
		t.Errorf("Expected ActivityCount=10, got %d", final.ActivityCount)
	}
}

func TestReportRepository_RecordGoalActivity_CompletesOncePerRollingWindow(t *testing.T) {
	_, repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	startAt := time.Date(2026, time.June, 8, 16, 30, 0, 0, time.UTC)
	endAt := startAt.AddDate(0, 0, 7)
	report := &domain.Report{
		UserID:         "user1",
		Name:           "Budi",
		LastReportDate: startAt,
		StreakFreezes:  1,
	}
	if err := repo.UpsertReport(ctx, report); err != nil {
		t.Fatalf("upsert report: %v", err)
	}
	if err := repo.SetGoal(ctx, &domain.WeeklyGoal{
		UserID:     "user1",
		TargetDays: 2,
		Activity:   "Olahraga",
		StartAt:    startAt,
		EndAt:      endAt,
		CreatedAt:  startAt,
	}); err != nil {
		t.Fatalf("set goal: %v", err)
	}

	if err := repo.LogActivity(ctx, "user1", domain.GetToday(startAt)); err != nil {
		t.Fatalf("log day 1: %v", err)
	}
	completed, err := repo.RecordGoalActivity(ctx, "user1", startAt, "Lari 5km")
	if err != nil {
		t.Fatalf("record day 1: %v", err)
	}
	if completed {
		t.Fatal("goal should not complete after one active day")
	}

	secondDay := startAt.AddDate(0, 0, 1)
	if err := repo.LogActivity(ctx, "user1", domain.GetToday(secondDay)); err != nil {
		t.Fatalf("log day 2: %v", err)
	}
	completed, err = repo.RecordGoalActivity(ctx, "user1", secondDay, "Gym 45min")
	if err != nil {
		t.Fatalf("record day 2: %v", err)
	}
	if !completed {
		t.Fatal("goal should complete after two unique active days")
	}

	completed, err = repo.RecordGoalActivity(ctx, "user1", secondDay.Add(10*time.Minute), "Laporan kedua")
	if err != nil {
		t.Fatalf("record repeat: %v", err)
	}
	if completed {
		t.Fatal("goal completion should only be counted once")
	}

	updated, err := repo.GetReport(ctx, "user1")
	if err != nil {
		t.Fatalf("get report: %v", err)
	}
	if updated.GoalsCompleted != 1 {
		t.Fatalf("expected GoalsCompleted=1, got %d", updated.GoalsCompleted)
	}

	activities, err := repo.GetGoalActivities(ctx, "user1", startAt, endAt)
	if err != nil {
		t.Fatalf("get goal activities: %v", err)
	}
	if len(activities) != 2 {
		t.Fatalf("expected 2 goal activity days, got %d", len(activities))
	}
	if activities[0].Activity != "Lari 5km" || activities[1].Activity != "Gym 45min" {
		t.Fatalf("unexpected activities: %+v", activities)
	}
}

func TestReportRepository_SetGoal_RejectsSecondActiveGoal(t *testing.T) {
	_, repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	startAt := time.Date(2026, time.June, 8, 16, 30, 0, 0, time.UTC)
	first := &domain.WeeklyGoal{
		UserID:     "user1",
		TargetDays: 2,
		Activity:   "Olahraga",
		StartAt:    startAt,
		EndAt:      startAt.AddDate(0, 0, 7),
		CreatedAt:  startAt,
	}
	if err := repo.SetGoal(ctx, first); err != nil {
		t.Fatalf("set first goal: %v", err)
	}

	secondStart := startAt.Add(time.Minute)
	err := repo.SetGoal(ctx, &domain.WeeklyGoal{
		UserID:     "user1",
		TargetDays: 3,
		Activity:   "Gym",
		StartAt:    secondStart,
		EndAt:      secondStart.AddDate(0, 0, 7),
		CreatedAt:  secondStart,
	})
	if err != domain.ErrActiveGoalExists {
		t.Fatalf("expected ErrActiveGoalExists, got %v", err)
	}
}

func TestReportRepository_RecordGoalActivity_UsesUTCReportDate(t *testing.T) {
	_, repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	wib := time.FixedZone("WIB", 7*60*60)
	startAt := time.Date(2026, time.June, 8, 20, 0, 0, 0, wib)
	activityAt := time.Date(2026, time.June, 9, 1, 0, 0, 0, wib)
	if err := repo.SetGoal(ctx, &domain.WeeklyGoal{
		UserID:     "user1",
		TargetDays: 1,
		Activity:   "Olahraga",
		StartAt:    startAt,
		EndAt:      startAt.AddDate(0, 0, 7),
		CreatedAt:  startAt,
	}); err != nil {
		t.Fatalf("set goal: %v", err)
	}
	completed, err := repo.RecordGoalActivity(ctx, "user1", activityAt, "Lari malam")
	if err != nil || !completed {
		t.Fatalf("expected completion, completed=%v err=%v", completed, err)
	}

	activities, err := repo.GetGoalActivities(ctx, "user1", startAt, startAt.AddDate(0, 0, 7))
	if err != nil {
		t.Fatalf("get goal activities: %v", err)
	}
	if len(activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(activities))
	}
	want := time.Date(2026, time.June, 8, 0, 0, 0, 0, time.UTC)
	if !activities[0].Date.Equal(want) {
		t.Fatalf("expected UTC report date %v, got %v", want, activities[0].Date)
	}
}

func TestReportRepository_RecordGoalActivity_CountsFinalPartialDayBeforeEnd(t *testing.T) {
	_, repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	wib := time.FixedZone("WIB", 7*60*60)
	startAt := time.Date(2026, time.June, 8, 16, 30, 0, 0, wib)
	endAt := startAt.AddDate(0, 0, 7)
	if err := repo.SetGoal(ctx, &domain.WeeklyGoal{
		UserID:     "user1",
		TargetDays: 1,
		Activity:   "Olahraga",
		StartAt:    startAt,
		EndAt:      endAt,
		CreatedAt:  startAt,
	}); err != nil {
		t.Fatalf("set goal: %v", err)
	}

	completed, err := repo.RecordGoalActivity(ctx, "user1", endAt.Add(-time.Hour), "Lari final")
	if err != nil || !completed {
		t.Fatalf("expected final partial day to count, completed=%v err=%v", completed, err)
	}
	completed, err = repo.RecordGoalActivity(ctx, "user1", endAt, "Terlambat")
	if err != nil {
		t.Fatalf("record at end: %v", err)
	}
	if completed {
		t.Fatal("activity exactly at end_at should not count as a new completion")
	}
}

func TestReportRepository_DeleteActiveGoal_DecrementsCompletedGoal(t *testing.T) {
	_, repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	startAt := time.Date(2026, time.June, 8, 16, 30, 0, 0, time.UTC)
	endAt := startAt.AddDate(0, 0, 7)
	report := &domain.Report{
		UserID:         "user1",
		Name:           "Budi",
		LastReportDate: startAt,
		StreakFreezes:  1,
	}
	if err := repo.UpsertReport(ctx, report); err != nil {
		t.Fatalf("upsert report: %v", err)
	}
	if err := repo.SetGoal(ctx, &domain.WeeklyGoal{
		UserID:     "user1",
		TargetDays: 1,
		Activity:   "Olahraga",
		StartAt:    startAt,
		EndAt:      endAt,
		CreatedAt:  startAt,
	}); err != nil {
		t.Fatalf("set goal: %v", err)
	}
	completed, err := repo.RecordGoalActivity(ctx, "user1", startAt.Add(time.Minute), "Lari")
	if err != nil || !completed {
		t.Fatalf("expected completion, completed=%v err=%v", completed, err)
	}

	if err := repo.DeleteActiveGoal(ctx, "user1", startAt.Add(time.Hour)); err != nil {
		t.Fatalf("delete active goal: %v", err)
	}
	goal, err := repo.GetActiveGoal(ctx, "user1", startAt.Add(time.Hour))
	if err != nil {
		t.Fatalf("get active goal: %v", err)
	}
	if goal != nil {
		t.Fatalf("expected goal deleted, got %+v", goal)
	}
	updated, err := repo.GetReport(ctx, "user1")
	if err != nil {
		t.Fatalf("get report: %v", err)
	}
	if updated.GoalsCompleted != 0 {
		t.Fatalf("expected GoalsCompleted=0 after reset, got %d", updated.GoalsCompleted)
	}
}

func TestReportRepository_DeleteActivityLog_ReopensCompletedGoalWhenBelowTarget(t *testing.T) {
	_, repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	startAt := time.Date(2026, time.June, 8, 16, 30, 0, 0, time.UTC)
	endAt := startAt.AddDate(0, 0, 7)
	report := &domain.Report{
		UserID:         "user1",
		Name:           "Budi",
		LastReportDate: startAt,
		StreakFreezes:  1,
	}
	if err := repo.UpsertReport(ctx, report); err != nil {
		t.Fatalf("upsert report: %v", err)
	}
	if err := repo.SetGoal(ctx, &domain.WeeklyGoal{
		UserID:     "user1",
		TargetDays: 2,
		Activity:   "Olahraga",
		StartAt:    startAt,
		EndAt:      endAt,
		CreatedAt:  startAt,
	}); err != nil {
		t.Fatalf("set goal: %v", err)
	}
	if err := repo.LogActivity(ctx, "user1", domain.GetToday(startAt)); err != nil {
		t.Fatalf("log day 1: %v", err)
	}
	if _, err := repo.RecordGoalActivity(ctx, "user1", startAt, "Lari"); err != nil {
		t.Fatalf("record day 1: %v", err)
	}
	secondDay := startAt.AddDate(0, 0, 1)
	if err := repo.LogActivity(ctx, "user1", domain.GetToday(secondDay)); err != nil {
		t.Fatalf("log day 2: %v", err)
	}
	completed, err := repo.RecordGoalActivity(ctx, "user1", secondDay, "Gym")
	if err != nil || !completed {
		t.Fatalf("expected goal completion, completed=%v err=%v", completed, err)
	}

	if err := repo.DeleteActivityLog(ctx, "user1", domain.GetToday(secondDay)); err != nil {
		t.Fatalf("delete activity: %v", err)
	}
	goal, err := repo.GetActiveGoal(ctx, "user1", startAt.Add(time.Hour))
	if err != nil {
		t.Fatalf("get active goal: %v", err)
	}
	if goal == nil || goal.CompletedAt != nil {
		t.Fatalf("expected goal reopened after falling below target, got %+v", goal)
	}
	updated, err := repo.GetReport(ctx, "user1")
	if err != nil {
		t.Fatalf("get report: %v", err)
	}
	if updated.GoalsCompleted != 0 {
		t.Fatalf("expected GoalsCompleted=0 after cancel, got %d", updated.GoalsCompleted)
	}
}

func TestReportRepository_DeleteActivityLogByKind_KeepsOtherKind(t *testing.T) {
	db, repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	today := domain.GetToday(time.Now())
	report := &domain.Report{UserID: "user1", Name: "Budi", LastReportDate: today, StreakFreezes: 1}
	if err := repo.UpsertReport(ctx, report); err != nil {
		t.Fatalf("upsert report: %v", err)
	}
	if err := repo.UpsertReportWithActivityKind(ctx, report, today, domain.ActivityKindRegularReport); err != nil {
		t.Fatalf("log regular: %v", err)
	}
	if err := repo.UpsertReportWithActivityKind(ctx, report, today, domain.ActivityKindSideQuest); err != nil {
		t.Fatalf("log sidequest: %v", err)
	}

	if err := repo.DeleteActivityLogByKind(ctx, "user1", today, domain.ActivityKindRegularReport); err != nil {
		t.Fatalf("delete regular: %v", err)
	}

	var totalCount, regularCount, sideQuestCount int
	err := db.QueryRowContext(ctx, `
		SELECT report_count, regular_report_count, sidequest_count
		FROM activity_logs
		WHERE user_id = ? AND activity_date = ?
	`, "user1", today.Format(time.DateOnly)).Scan(&totalCount, &regularCount, &sideQuestCount)
	if err != nil {
		t.Fatalf("query activity log: %v", err)
	}
	if totalCount != 1 || regularCount != 0 || sideQuestCount != 1 {
		t.Fatalf("expected only sidequest remaining, got total=%d regular=%d sidequest=%d", totalCount, regularCount, sideQuestCount)
	}

	if err := repo.DeleteActivityLogByKind(ctx, "user1", today, domain.ActivityKindSideQuest); err != nil {
		t.Fatalf("delete sidequest: %v", err)
	}
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM activity_logs WHERE user_id = ? AND activity_date = ?`, "user1", today.Format(time.DateOnly)).Scan(&totalCount)
	if err != nil {
		t.Fatalf("count activity log: %v", err)
	}
	if totalCount != 0 {
		t.Fatalf("expected activity log deleted after both kinds removed, got %d rows", totalCount)
	}
}

func TestReportRepository_DeleteLatestActivityLogByKind_DecrementsOnlyRequestedKind(t *testing.T) {
	db, repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	today := domain.GetToday(time.Now())
	report := &domain.Report{UserID: "user1", Name: "Budi", LastReportDate: today, StreakFreezes: 1}
	if err := repo.UpsertReport(ctx, report); err != nil {
		t.Fatalf("upsert report: %v", err)
	}
	if err := repo.UpsertReportWithActivityKind(ctx, report, today, domain.ActivityKindRegularReport); err != nil {
		t.Fatalf("log regular: %v", err)
	}
	for i := 0; i < 2; i++ {
		if err := repo.UpsertReportWithActivityKind(ctx, report, today, domain.ActivityKindSideQuest); err != nil {
			t.Fatalf("log sidequest %d: %v", i+1, err)
		}
	}

	remaining, err := repo.DeleteLatestActivityLogByKind(ctx, "user1", today, domain.ActivityKindSideQuest)
	if err != nil {
		t.Fatalf("delete latest sidequest: %v", err)
	}
	if remaining != 1 {
		t.Fatalf("expected 1 sidequest remaining, got %d", remaining)
	}

	var totalCount, regularCount, sideQuestCount int
	err = db.QueryRowContext(ctx, `
		SELECT report_count, regular_report_count, sidequest_count
		FROM activity_logs
		WHERE user_id = ? AND activity_date = ?
	`, "user1", today.Format(time.DateOnly)).Scan(&totalCount, &regularCount, &sideQuestCount)
	if err != nil {
		t.Fatalf("query activity log: %v", err)
	}
	if totalCount != 2 || regularCount != 1 || sideQuestCount != 1 {
		t.Fatalf("expected one regular and one sidequest remaining, got total=%d regular=%d sidequest=%d", totalCount, regularCount, sideQuestCount)
	}
}
