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

const selectColumns = `user_id, name, streak, activity_count, last_report_date, 
	COALESCE(max_streak, 0), COALESCE(total_points, 0), COALESCE(achievements, ''),
	COALESCE(comeback_streak, 0), COALESCE(inactive_days, 0)`

func scanReport(scanner interface{ Scan(dest ...any) error }) (*domain.Report, error) {
	var report domain.Report
	var lastReportDate string
	err := scanner.Scan(
		&report.UserID, &report.Name, &report.Streak, &report.ActivityCount,
		&lastReportDate, &report.MaxStreak, &report.TotalPoints, &report.Achievements,
		&report.ComebackStreak, &report.InactiveDays,
	)
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

func (r *ReportRepository) GetReport(ctx context.Context, userID string) (*domain.Report, error) {
	query := `SELECT ` + selectColumns + ` FROM user_reports WHERE user_id = ?`
	row := r.db.QueryRowContext(ctx, query, userID)
	return scanReport(row)
}

func (r *ReportRepository) UpsertReport(ctx context.Context, report *domain.Report) error {
	query := `
		INSERT INTO user_reports (user_id, name, streak, activity_count, last_report_date, max_streak, total_points, achievements, comeback_streak, inactive_days)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			name = excluded.name,
			streak = excluded.streak,
			activity_count = excluded.activity_count,
			last_report_date = excluded.last_report_date,
			max_streak = excluded.max_streak,
			total_points = excluded.total_points,
			achievements = excluded.achievements,
			comeback_streak = excluded.comeback_streak,
			inactive_days = excluded.inactive_days
	`
	_, err := r.db.ExecContext(ctx, query,
		report.UserID, report.Name, report.Streak, report.ActivityCount,
		report.LastReportDate.Format(time.RFC3339), report.MaxStreak, report.TotalPoints,
		report.Achievements, report.ComebackStreak, report.InactiveDays,
	)
	return err
}

func scanReports(rows *sql.Rows) ([]*domain.Report, error) {
	var reports []*domain.Report
	for rows.Next() {
		report, err := scanReport(rows)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}
	return reports, nil
}

func (r *ReportRepository) GetAllReports(ctx context.Context) ([]*domain.Report, error) {
	query := `SELECT ` + selectColumns + ` FROM user_reports ORDER BY activity_count DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanReports(rows)
}

func (r *ReportRepository) GetInactiveUsers(ctx context.Context, days int) ([]*domain.Report, error) {
	query := `
		SELECT ` + selectColumns + `
		FROM user_reports 
		WHERE last_report_date < datetime('now', '-' || ? || ' days')
		ORDER BY last_report_date ASC
	`
	rows, err := r.db.QueryContext(ctx, query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanReports(rows)
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

	// Strava Accounts Table
	stravaQuery := `
		CREATE TABLE IF NOT EXISTS strava_accounts (
			user_id TEXT PRIMARY KEY,
			athlete_id INTEGER UNIQUE,
			access_token TEXT,
			refresh_token TEXT,
			expires_at TEXT,
			FOREIGN KEY (user_id) REFERENCES user_reports(user_id)
		);
	`
	_, err = r.db.ExecContext(ctx, stravaQuery)
	if err != nil {
		return err
	}

	// Simple migration: try to add columns if they don't exist
	// Ignore error if they already exist
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN activity_count INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN max_streak INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN total_points INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN achievements TEXT DEFAULT ''")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN comeback_streak INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN inactive_days INTEGER DEFAULT 0")

	// Run data migrations
	if err := r.MigrateDayToWeekStreaks(ctx); err != nil {
		return err
	}

	return nil
}

func (r *ReportRepository) MigrateDayToWeekStreaks(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS sys_migrations (name TEXT PRIMARY KEY)")
	if err != nil {
		return err
	}

	var exists int
	// Using a new migration name to allow re-running with the refined logic if needed
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sys_migrations WHERE name = 'streak_day_to_week_v2'").Scan(&exists)
	if err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}

	// Perform migration: use total days / 7 for initial weekly streak
	_, err = r.db.ExecContext(ctx, `
		UPDATE user_reports 
		SET streak = activity_count / 7,
			max_streak = activity_count / 7,
			comeback_streak = 0
		WHERE activity_count >= 7
	`)
	if err != nil {
		return err
	}

	// For users with < 7 days, set streak to 1 if they have any activity
	_, err = r.db.ExecContext(ctx, `
		UPDATE user_reports 
		SET streak = 1,
			max_streak = 1
		WHERE activity_count > 0 AND activity_count < 7
	`)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, "INSERT INTO sys_migrations (name) VALUES ('streak_day_to_week_v2')")
	return err
}

// ResolveLIDToPhone looks up a LID in the whatsmeow_lid_map table and returns the phone number.
// If not found or if input is already a phone number, returns the input unchanged.
func (r *ReportRepository) ResolveLIDToPhone(ctx context.Context, lid string) string {
	query := `SELECT pn FROM whatsmeow_lid_map WHERE lid = ?`
	var phone string
	err := r.db.QueryRowContext(ctx, query, lid).Scan(&phone)
	if err == nil && phone != "" {
		return phone
	}
	return lid
}

// Strava Integration

func (r *ReportRepository) UpsertStravaAccount(ctx context.Context, account *domain.StravaAccount) error {
	query := `
		INSERT INTO strava_accounts (user_id, athlete_id, access_token, refresh_token, expires_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			athlete_id = excluded.athlete_id,
			access_token = excluded.access_token,
			refresh_token = excluded.refresh_token,
			expires_at = excluded.expires_at
	`
	_, err := r.db.ExecContext(ctx, query,
		account.UserID, account.AthleteID, account.AccessToken,
		account.RefreshToken, account.ExpiresAt.Format(time.RFC3339),
	)
	return err
}

func (r *ReportRepository) GetStravaAccountByAthleteID(ctx context.Context, athleteID int64) (*domain.StravaAccount, error) {
	query := `SELECT user_id, athlete_id, access_token, refresh_token, expires_at FROM strava_accounts WHERE athlete_id = ?`
	row := r.db.QueryRowContext(ctx, query, athleteID)

	var account domain.StravaAccount
	var expiresAt string
	err := row.Scan(&account.UserID, &account.AthleteID, &account.AccessToken, &account.RefreshToken, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	account.ExpiresAt, err = time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (r *ReportRepository) GetStravaAccountByUserID(ctx context.Context, userID string) (*domain.StravaAccount, error) {
	query := `SELECT user_id, athlete_id, access_token, refresh_token, expires_at FROM strava_accounts WHERE user_id = ?`
	row := r.db.QueryRowContext(ctx, query, userID)

	var account domain.StravaAccount
	var expiresAt string
	err := row.Scan(&account.UserID, &account.AthleteID, &account.AccessToken, &account.RefreshToken, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	account.ExpiresAt, err = time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return nil, err
	}

	return &account, nil
}
