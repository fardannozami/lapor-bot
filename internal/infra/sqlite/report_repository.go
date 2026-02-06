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
	query := `SELECT user_id, name, streak, last_report_date FROM user_reports WHERE user_id = ?`
	row := r.db.QueryRowContext(ctx, query, userID)

	var report domain.Report
	var lastReportDate string
	err := row.Scan(&report.UserID, &report.Name, &report.Streak, &lastReportDate)
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
		INSERT INTO user_reports (user_id, name, streak, last_report_date)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			name = excluded.name,
			streak = excluded.streak,
			last_report_date = excluded.last_report_date
	`
	_, err := r.db.ExecContext(ctx, query, report.UserID, report.Name, report.Streak, report.LastReportDate.Format(time.RFC3339))
	return err
}

func (r *ReportRepository) GetAllReports(ctx context.Context) ([]*domain.Report, error) {
	query := `SELECT user_id, name, streak, last_report_date FROM user_reports ORDER BY streak DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*domain.Report
	for rows.Next() {
		var report domain.Report
		var lastReportDate string
		if err := rows.Scan(&report.UserID, &report.Name, &report.Streak, &lastReportDate); err != nil {
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
			last_report_date TEXT
		);
	`
	_, err := r.db.ExecContext(ctx, query)
	return err
}
