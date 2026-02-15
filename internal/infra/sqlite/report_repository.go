package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type ReportRepository struct {
	db *sql.DB
}

func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) GetReport(ctx context.Context, userID string) (*domain.Report, error) {
	query := `SELECT user_id, name, streak, activity_count, last_report_date, COALESCE(max_streak, 0), COALESCE(total_points, 0), COALESCE(achievements, '') FROM user_reports WHERE user_id = ?`
	row := r.db.QueryRowContext(ctx, query, userID)

	var report domain.Report
	var lastReportDate string
	err := row.Scan(&report.UserID, &report.Name, &report.Streak, &report.ActivityCount, &lastReportDate, &report.MaxStreak, &report.TotalPoints, &report.Achievements)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	report.LastReportDate, err = time.Parse(time.RFC3339, lastReportDate)
	if err != nil {
		return nil, err
	}

	return &report, nil
}

func (r *ReportRepository) UpsertReport(ctx context.Context, report *domain.Report) error {
	query := `
		INSERT INTO user_reports (user_id, name, streak, activity_count, last_report_date, max_streak, total_points, achievements)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			name = excluded.name,
			streak = excluded.streak,
			activity_count = excluded.activity_count,
			last_report_date = excluded.last_report_date,
			max_streak = excluded.max_streak,
			total_points = excluded.total_points,
			achievements = excluded.achievements
	`
	_, err := r.db.ExecContext(ctx, query, report.UserID, report.Name, report.Streak, report.ActivityCount, report.LastReportDate.Format(time.RFC3339), report.MaxStreak, report.TotalPoints, report.Achievements)
	return err
}

func (r *ReportRepository) GetAllReports(ctx context.Context) ([]*domain.Report, error) {
	query := `SELECT user_id, name, streak, activity_count, last_report_date, COALESCE(max_streak, 0), COALESCE(total_points, 0), COALESCE(achievements, '') FROM user_reports ORDER BY activity_count DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*domain.Report
	for rows.Next() {
		var report domain.Report
		var lastReportDate string
		if err := rows.Scan(&report.UserID, &report.Name, &report.Streak, &report.ActivityCount, &lastReportDate, &report.MaxStreak, &report.TotalPoints, &report.Achievements); err != nil {
			return nil, err
		}
		report.LastReportDate, err = time.Parse(time.RFC3339, lastReportDate)
		if err != nil {
			return nil, err
		}
		reports = append(reports, &report)
	}
	return reports, nil
}

func (r *ReportRepository) InitTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS user_reports (
			user_id TEXT PRIMARY KEY,
			name TEXT,
			streak INTEGER,
			activity_count INTEGER DEFAULT 0,
			last_report_date TEXT
		);
	`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	// Simple migration: try to add columns if they don't exist
	// Ignore error if they already exist
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN activity_count INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN max_streak INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN total_points INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN achievements TEXT DEFAULT ''")

	return nil
}

// ResolveLIDToPhone looks up a LID in the whatsmeow_lid_map table and returns the phone number.
// If not found or if input is already a phone number, returns the input unchanged.
func (r *ReportRepository) ResolveLIDToPhone(ctx context.Context, lid string) string {
	// Query the lid_map table
	query := `SELECT pn FROM whatsmeow_lid_map WHERE lid = ?`
	var phone string
	err := r.db.QueryRowContext(ctx, query, lid).Scan(&phone)
	if err == nil && phone != "" {
		return phone
	}
	// Not found or error, return original
	return lid
}
