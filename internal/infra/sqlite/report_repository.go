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

type execContexter interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

const selectColumns = `user_id, name, streak, activity_count, last_report_date, 
	COALESCE(max_streak, 0), COALESCE(total_points, 0), COALESCE(achievements, ''),
	COALESCE(comeback_streak, 0), COALESCE(inactive_days, 0), COALESCE(centurion_cycles, 0)`

func scanReport(scanner interface{ Scan(dest ...any) error }) (*domain.Report, error) {
	var report domain.Report
	var lastReportDate string
	err := scanner.Scan(
		&report.UserID, &report.Name, &report.Streak, &report.ActivityCount,
		&lastReportDate, &report.MaxStreak, &report.TotalPoints, &report.Achievements,
		&report.ComebackStreak, &report.InactiveDays, &report.CenturionCycles,
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

func upsertReport(ctx context.Context, execer execContexter, report *domain.Report) error {
	query := `
		INSERT INTO user_reports (user_id, name, streak, activity_count, last_report_date, max_streak, total_points, achievements, comeback_streak, inactive_days, centurion_cycles)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			name = excluded.name,
			streak = excluded.streak,
			activity_count = excluded.activity_count,
			last_report_date = excluded.last_report_date,
			max_streak = excluded.max_streak,
			total_points = excluded.total_points,
			achievements = excluded.achievements,
			comeback_streak = excluded.comeback_streak,
			inactive_days = excluded.inactive_days,
			centurion_cycles = excluded.centurion_cycles
	`
	_, err := execer.ExecContext(ctx, query,
		report.UserID, report.Name, report.Streak, report.ActivityCount,
		report.LastReportDate.Format(time.RFC3339), report.MaxStreak, report.TotalPoints,
		report.Achievements, report.ComebackStreak, report.InactiveDays, report.CenturionCycles,
	)
	return err
}

func (r *ReportRepository) UpsertReport(ctx context.Context, report *domain.Report) error {
	return upsertReport(ctx, r.db, report)
}

func logActivity(ctx context.Context, execer execContexter, userID string, activityDate time.Time) error {
	query := `
		INSERT OR IGNORE INTO activity_logs (user_id, activity_date, created_at)
		VALUES (?, ?, ?)
	`
	_, err := execer.ExecContext(ctx, query, userID, activityDate.Format(time.DateOnly), time.Now().UTC().Format(time.RFC3339))
	return err
}

func (r *ReportRepository) UpsertReportWithActivity(ctx context.Context, report *domain.Report, activityDate time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := upsertReport(ctx, tx, report); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := logActivity(ctx, tx, report.UserID, activityDate); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *ReportRepository) LogActivity(ctx context.Context, userID string, activityDate time.Time) error {
	return logActivity(ctx, r.db, userID, activityDate)
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

func (r *ReportRepository) GetActivityCountsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]domain.ActivityLeaderboardEntry, error) {
	query := `
		SELECT al.user_id,
		       COALESCE(NULLIF(ur.name, ''), al.user_id) AS name,
		       COUNT(*) AS activity_count
		FROM activity_logs al
		LEFT JOIN user_reports ur ON ur.user_id = al.user_id
		WHERE al.activity_date >= ? AND al.activity_date < ?
		GROUP BY al.user_id, COALESCE(NULLIF(ur.name, ''), al.user_id)
		ORDER BY activity_count DESC, name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, startDate.Format(time.DateOnly), endDate.Format(time.DateOnly))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.ActivityLeaderboardEntry
	for rows.Next() {
		var entry domain.ActivityLeaderboardEntry
		if err := rows.Scan(&entry.UserID, &entry.Name, &entry.ActivityCount); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
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

func (r *ReportRepository) ResetAllReports(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM activity_logs`); err != nil {
		return err
	}
	_, err := r.db.ExecContext(ctx, `DELETE FROM user_reports`)
	return err
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

	activityLogQuery := `
		CREATE TABLE IF NOT EXISTS activity_logs (
			user_id TEXT NOT NULL,
			activity_date TEXT NOT NULL,
			created_at TEXT NOT NULL,
			PRIMARY KEY (user_id, activity_date),
			FOREIGN KEY (user_id) REFERENCES user_reports(user_id) ON DELETE CASCADE
		);
	`
	_, err = r.db.ExecContext(ctx, activityLogQuery)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_activity_logs_date ON activity_logs (activity_date)`)
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
			name TEXT,
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
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN centurion_cycles INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE strava_accounts ADD COLUMN name TEXT")

	// Run data migrations
	if err := r.MigrateDayToWeekStreaks(ctx); err != nil {
		return err
	}
	if err := r.MigrateCenturionPrestige(ctx); err != nil {
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

func (r *ReportRepository) MigrateCenturionPrestige(ctx context.Context) error {
	var exists int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sys_migrations WHERE name = 'centurion_prestige_v1'").Scan(&exists)
	if err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}

	// For users with > 100 days:
	// cycles = (activity_count - 1) / 100
	// new_count = ((activity_count - 1) % 100) + 1
	// SQLite doesn't have integer division like Go, but we can use CAST or floor
	_, err = r.db.ExecContext(ctx, `
		UPDATE user_reports 
		SET centurion_cycles = (activity_count - 1) / 100,
		    activity_count = ((activity_count - 1) % 100) + 1
		WHERE activity_count > 100
	`)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, "INSERT INTO sys_migrations (name) VALUES ('centurion_prestige_v1')")
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
		INSERT INTO strava_accounts (user_id, athlete_id, access_token, refresh_token, expires_at, name)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			athlete_id = excluded.athlete_id,
			access_token = excluded.access_token,
			refresh_token = excluded.refresh_token,
			expires_at = excluded.expires_at,
			name = excluded.name
	`
	_, err := r.db.ExecContext(ctx, query,
		account.UserID, account.AthleteID, account.AccessToken,
		account.RefreshToken, account.ExpiresAt.Format(time.RFC3339), account.Name,
	)
	return err
}

func (r *ReportRepository) GetStravaAccountByAthleteID(ctx context.Context, athleteID int64) (*domain.StravaAccount, error) {
	query := `SELECT user_id, athlete_id, access_token, refresh_token, expires_at, COALESCE(name, '') FROM strava_accounts WHERE athlete_id = ?`
	row := r.db.QueryRowContext(ctx, query, athleteID)

	var account domain.StravaAccount
	var expiresAt string
	err := row.Scan(&account.UserID, &account.AthleteID, &account.AccessToken, &account.RefreshToken, &expiresAt, &account.Name)
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
	query := `SELECT user_id, athlete_id, access_token, refresh_token, expires_at, COALESCE(name, '') FROM strava_accounts WHERE user_id = ?`
	row := r.db.QueryRowContext(ctx, query, userID)

	var account domain.StravaAccount
	var expiresAt string
	err := row.Scan(&account.UserID, &account.AthleteID, &account.AccessToken, &account.RefreshToken, &expiresAt, &account.Name)
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
