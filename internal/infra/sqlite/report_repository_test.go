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
