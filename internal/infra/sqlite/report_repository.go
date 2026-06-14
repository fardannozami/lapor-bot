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

const selectColumns = `user_id, name, COALESCE(job_class, ''), streak, activity_count, last_report_date, 
	COALESCE(max_streak, 0), COALESCE(total_points, 0), COALESCE(level, 0), COALESCE(achievements, ''),
	COALESCE(comeback_streak, 0), COALESCE(inactive_days, 0), COALESCE(centurion_cycles, 0),
	COALESCE(seasonal_points, 0), COALESCE(seasonal_activity_count, 0),
	COALESCE(seasonal_max_streak, 0), COALESCE(seasonal_achievements, ''),
	COALESCE(streak_freezes, 0), COALESCE(goals_completed, 0)`

func scanReport(scanner interface{ Scan(dest ...any) error }) (*domain.Report, error) {
	var report domain.Report
	var lastReportDate string
	err := scanner.Scan(
		&report.UserID, &report.Name, &report.JobClass, &report.Streak, &report.ActivityCount,
		&lastReportDate, &report.MaxStreak, &report.TotalPoints, &report.Level, &report.Achievements,
		&report.ComebackStreak, &report.InactiveDays, &report.CenturionCycles,
		&report.SeasonalPoints, &report.SeasonalActivityCount,
		&report.SeasonalMaxStreak, &report.SeasonalAchievements,
		&report.StreakFreezes, &report.GoalsCompleted,
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
		INSERT INTO user_reports (user_id, name, job_class, streak, activity_count, last_report_date, max_streak, total_points, level, achievements, comeback_streak, inactive_days, centurion_cycles, seasonal_points, seasonal_activity_count, seasonal_max_streak, seasonal_achievements, streak_freezes, goals_completed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			name = excluded.name,
			job_class = excluded.job_class,
			streak = excluded.streak,
			activity_count = excluded.activity_count,
			last_report_date = excluded.last_report_date,
			max_streak = excluded.max_streak,
			total_points = excluded.total_points,
			level = excluded.level,
			achievements = excluded.achievements,
			comeback_streak = excluded.comeback_streak,
			inactive_days = excluded.inactive_days,
			centurion_cycles = excluded.centurion_cycles,
			seasonal_points = excluded.seasonal_points,
			seasonal_activity_count = excluded.seasonal_activity_count,
			seasonal_max_streak = excluded.seasonal_max_streak,
			seasonal_achievements = excluded.seasonal_achievements,
			streak_freezes = excluded.streak_freezes
	`
	_, err := execer.ExecContext(ctx, query,
		report.UserID, report.Name, report.JobClass, report.Streak, report.ActivityCount,
		report.LastReportDate.Format(time.RFC3339), report.MaxStreak, report.TotalPoints,
		report.Level, report.Achievements, report.ComebackStreak, report.InactiveDays, report.CenturionCycles,
		report.SeasonalPoints, report.SeasonalActivityCount, report.SeasonalMaxStreak,
		report.SeasonalAchievements, report.StreakFreezes, report.GoalsCompleted,
	)
	return err
}

func (r *ReportRepository) UpsertReport(ctx context.Context, report *domain.Report) error {
	return upsertReport(ctx, r.db, report)
}

func logActivity(ctx context.Context, execer execContexter, userID string, activityDate time.Time) error {
	query := `
		INSERT INTO activity_logs (user_id, activity_date, created_at, report_count)
		VALUES (?, ?, ?, 1)
		ON CONFLICT(user_id, activity_date) DO UPDATE SET
			report_count = report_count + 1,
			created_at = excluded.created_at
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

// ResetSeasonalCounters clears only seasonal data, preserving lifetime progress.
// This replaces the old destructive reset that deleted all user data.
func (r *ReportRepository) ResetAllReports(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_reports
		SET seasonal_points = 0,
		    seasonal_activity_count = 0,
		    seasonal_max_streak = 0,
		    seasonal_achievements = '',
		    streak_freezes = 1
	`)
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
			report_count INTEGER NOT NULL DEFAULT 1,
			activity_text TEXT DEFAULT '',
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
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN job_class TEXT DEFAULT ''")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN max_streak INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN total_points INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN level INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN achievements TEXT DEFAULT ''")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN comeback_streak INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN inactive_days INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN centurion_cycles INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN seasonal_points INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN seasonal_activity_count INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN seasonal_max_streak INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN seasonal_achievements TEXT DEFAULT ''")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN streak_freezes INTEGER DEFAULT 1")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN goals_completed INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE activity_logs ADD COLUMN report_count INTEGER NOT NULL DEFAULT 1")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE activity_logs ADD COLUMN activity_text TEXT DEFAULT ''")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE strava_accounts ADD COLUMN name TEXT")

	goalQuery := `
		CREATE TABLE IF NOT EXISTS weekly_goals (
			user_id TEXT NOT NULL,
			target_days INTEGER NOT NULL,
			activity TEXT NOT NULL,
			week_start TEXT NOT NULL,
			created_at TEXT NOT NULL,
			completed_at TEXT DEFAULT '',
			PRIMARY KEY (user_id, week_start)
		);
	`
	_, err = r.db.ExecContext(ctx, goalQuery)
	if err != nil {
		return err
	}

	rollingGoalQuery := `
		CREATE TABLE IF NOT EXISTS goals (
			user_id TEXT NOT NULL,
			target_days INTEGER NOT NULL,
			activity TEXT NOT NULL,
			start_at TEXT NOT NULL,
			end_at TEXT NOT NULL,
			created_at TEXT NOT NULL,
			completed_at TEXT DEFAULT '',
			PRIMARY KEY (user_id, start_at)
		);
	`
	_, err = r.db.ExecContext(ctx, rollingGoalQuery)
	if err != nil {
		return err
	}

	goalActivityQuery := `
		CREATE TABLE IF NOT EXISTS goal_activity_logs (
			user_id TEXT NOT NULL,
			goal_start_at TEXT NOT NULL,
			activity_date TEXT NOT NULL,
			activity_text TEXT DEFAULT '',
			created_at TEXT NOT NULL,
			PRIMARY KEY (user_id, goal_start_at, activity_date)
		);
	`
	_, err = r.db.ExecContext(ctx, goalActivityQuery)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_goals_end_at ON goals (end_at)`)
	if err != nil {
		return err
	}

	// Run data migrations
	if err := r.MigrateDayToWeekStreaks(ctx); err != nil {
		return err
	}
	if err := r.MigrateCenturionPrestige(ctx); err != nil {
		return err
	}
	if err := r.MigrateNumericLevels(ctx); err != nil {
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

func (r *ReportRepository) MigrateNumericLevels(ctx context.Context) error {
	var exists int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sys_migrations WHERE name = 'numeric_level_v1'").Scan(&exists)
	if err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}

	rows, err := r.db.QueryContext(ctx, `SELECT user_id, COALESCE(total_points, 0) FROM user_reports`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type levelUpdate struct {
		userID string
		level  int
	}
	var updates []levelUpdate
	for rows.Next() {
		var userID string
		var totalPoints int
		if err := rows.Scan(&userID, &totalPoints); err != nil {
			return err
		}
		updates = append(updates, levelUpdate{userID: userID, level: domain.NumericLevelFromTotalPoints(totalPoints)})
	}
	if err := rows.Err(); err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	for _, update := range updates {
		if _, err := tx.ExecContext(ctx, `UPDATE user_reports SET level = ? WHERE user_id = ?`, update.level, update.userID); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO sys_migrations (name) VALUES ('numeric_level_v1')"); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
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

func (r *ReportRepository) GetUserActivityDates(ctx context.Context, userID string) ([]time.Time, error) {
	query := `
		SELECT activity_date FROM activity_logs
		WHERE user_id = ?
		ORDER BY activity_date ASC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dates []time.Time
	for rows.Next() {
		var dateStr string
		if err := rows.Scan(&dateStr); err != nil {
			return nil, err
		}
		date, err := time.Parse(time.DateOnly, dateStr)
		if err != nil {
			return nil, err
		}
		dates = append(dates, date)
	}
	return dates, rows.Err()
}

func (r *ReportRepository) DeleteActivityLog(ctx context.Context, userID string, activityDate time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM activity_logs WHERE user_id = ? AND activity_date = ?`, userID, activityDate.Format(time.DateOnly)); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := reconcileGoalAfterActivityDelete(ctx, tx, userID, activityDate); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *ReportRepository) DeleteLatestActivityLog(ctx context.Context, userID string, activityDate time.Time) (int, error) {
	date := activityDate.Format(time.DateOnly)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	var count int
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(report_count, 1) FROM activity_logs WHERE user_id = ? AND activity_date = ?`, userID, date).Scan(&count)
	if err == sql.ErrNoRows {
		_ = tx.Rollback()
		return 0, nil
	}
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	remaining := count - 1
	if remaining > 0 {
		if _, err := tx.ExecContext(ctx, `UPDATE activity_logs SET report_count = ? WHERE user_id = ? AND activity_date = ?`, remaining, userID, date); err != nil {
			_ = tx.Rollback()
			return 0, err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `DELETE FROM activity_logs WHERE user_id = ? AND activity_date = ?`, userID, date); err != nil {
			_ = tx.Rollback()
			return 0, err
		}
		if err := reconcileGoalAfterActivityDelete(ctx, tx, userID, activityDate); err != nil {
			_ = tx.Rollback()
			return 0, err
		}
	}

	return remaining, tx.Commit()
}

func reconcileGoalAfterActivityDelete(ctx context.Context, tx *sql.Tx, userID string, activityDate time.Time) error {
	activityDateStr := activityDate.Format(time.DateOnly)
	rows, err := tx.QueryContext(ctx, `
		SELECT g.start_at, g.target_days, COALESCE(g.completed_at, '')
		FROM goals g
		JOIN goal_activity_logs gal ON gal.user_id = g.user_id AND gal.goal_start_at = g.start_at
		WHERE g.user_id = ? AND gal.activity_date = ?
	`, userID, activityDateStr)
	if err != nil {
		return err
	}
	defer rows.Close()

	type affectedGoal struct {
		startAt     string
		targetDays  int
		completedAt string
	}
	var goals []affectedGoal
	for rows.Next() {
		var goal affectedGoal
		if err := rows.Scan(&goal.startAt, &goal.targetDays, &goal.completedAt); err != nil {
			return err
		}
		goals = append(goals, goal)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM goal_activity_logs WHERE user_id = ? AND activity_date = ?`, userID, activityDateStr); err != nil {
		return err
	}

	for _, goal := range goals {
		if goal.completedAt == "" {
			continue
		}
		var activeDays int
		if err := tx.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM goal_activity_logs
			WHERE user_id = ? AND goal_start_at = ?
		`, userID, goal.startAt).Scan(&activeDays); err != nil {
			return err
		}
		if activeDays >= goal.targetDays {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE goals
			SET completed_at = ''
			WHERE user_id = ? AND start_at = ?
		`, userID, goal.startAt); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
		UPDATE user_reports
		SET goals_completed = CASE
			WHEN COALESCE(goals_completed, 0) > 0 THEN goals_completed - 1
			ELSE 0
		END
		WHERE user_id = ?
	`, userID); err != nil {
			return err
		}
	}
	return nil
}

func (r *ReportRepository) DeleteReport(ctx context.Context, userID string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM activity_logs WHERE user_id = ?`, userID); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `DELETE FROM weekly_goals WHERE user_id = ?`, userID); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `DELETE FROM goal_activity_logs WHERE user_id = ?`, userID); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `DELETE FROM goals WHERE user_id = ?`, userID); err != nil {
		return err
	}
	_, err := r.db.ExecContext(ctx, `DELETE FROM user_reports WHERE user_id = ?`, userID)
	return err
}

func (r *ReportRepository) GetDailyActivityCount(ctx context.Context, userID string, date time.Time) (int, error) {
	query := `SELECT COALESCE(report_count, 1) FROM activity_logs WHERE user_id = ? AND activity_date = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID, date.Format(time.DateOnly)).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ReportRepository) SetGoal(ctx context.Context, goal *domain.WeeklyGoal) error {
	query := `
		INSERT INTO goals (user_id, target_days, activity, start_at, end_at, created_at, completed_at)
		SELECT ?, ?, ?, ?, ?, ?, ''
		WHERE NOT EXISTS (
			SELECT 1 FROM goals
			WHERE user_id = ? AND start_at <= ? AND end_at > ?
		)
	`
	startAt := goal.StartAt.UTC().Format(time.RFC3339)
	endAt := goal.EndAt.UTC().Format(time.RFC3339)
	createdAt := goal.CreatedAt.UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, query,
		goal.UserID,
		goal.TargetDays,
		goal.Activity,
		startAt,
		endAt,
		createdAt,
		goal.UserID,
		startAt,
		startAt,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrActiveGoalExists
	}
	return nil
}

func (r *ReportRepository) GetActiveGoal(ctx context.Context, userID string, now time.Time) (*domain.WeeklyGoal, error) {
	query := `
		SELECT user_id, target_days, activity, start_at, end_at, created_at, COALESCE(completed_at, '')
		FROM goals
		WHERE user_id = ? AND start_at <= ? AND end_at > ?
		ORDER BY start_at DESC
		LIMIT 1
	`
	nowStr := now.UTC().Format(time.RFC3339)
	row := r.db.QueryRowContext(ctx, query, userID, nowStr, nowStr)
	return scanGoal(row)
}

func scanGoal(scanner interface{ Scan(dest ...any) error }) (*domain.WeeklyGoal, error) {
	var goal domain.WeeklyGoal
	var startAtStr, endAtStr, createdAtStr, completedAtStr string
	err := scanner.Scan(&goal.UserID, &goal.TargetDays, &goal.Activity, &startAtStr, &endAtStr, &createdAtStr, &completedAtStr)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	goal.StartAt, err = time.Parse(time.RFC3339, startAtStr)
	if err != nil {
		return nil, err
	}
	goal.EndAt, err = time.Parse(time.RFC3339, endAtStr)
	if err != nil {
		return nil, err
	}
	goal.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	if completedAtStr != "" {
		completedAt, err := time.Parse(time.RFC3339, completedAtStr)
		if err != nil {
			return nil, err
		}
		goal.CompletedAt = &completedAt
	}
	return &goal, nil
}

func (r *ReportRepository) DeleteActiveGoal(ctx context.Context, userID string, now time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	nowStr := now.UTC().Format(time.RFC3339)
	var startAt, completedAt string
	err = tx.QueryRowContext(ctx, `
		SELECT start_at, COALESCE(completed_at, '')
		FROM goals
		WHERE user_id = ? AND start_at <= ? AND end_at > ?
		ORDER BY start_at DESC
		LIMIT 1
	`, userID, nowStr, nowStr).Scan(&startAt, &completedAt)
	if err == sql.ErrNoRows {
		return tx.Commit()
	}
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM goal_activity_logs WHERE user_id = ? AND goal_start_at = ?`, userID, startAt); err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM goals WHERE user_id = ? AND start_at = ?`, userID, startAt); err != nil {
		_ = tx.Rollback()
		return err
	}
	if completedAt != "" {
		if _, err := tx.ExecContext(ctx, `
			UPDATE user_reports
			SET goals_completed = CASE
				WHEN COALESCE(goals_completed, 0) > 0 THEN goals_completed - 1
				ELSE 0
			END
			WHERE user_id = ?
		`, userID); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *ReportRepository) DeleteExpiredGoals(ctx context.Context, now time.Time) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	nowStr := now.UTC().Format(time.RFC3339)
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM goal_activity_logs
		WHERE (user_id, goal_start_at) IN (
			SELECT user_id, start_at FROM goals WHERE end_at <= ?
		)
	`, nowStr); err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM goals WHERE end_at <= ?`, nowStr)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	deleted, err := res.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	return deleted, tx.Commit()
}

func (r *ReportRepository) GetGoalActivities(ctx context.Context, userID string, startAt, endAt time.Time) ([]domain.GoalActivity, error) {
	query := `
		SELECT activity_date, COALESCE(activity_text, '')
		FROM goal_activity_logs
		WHERE user_id = ? AND goal_start_at = ?
		ORDER BY activity_date ASC
	`
	rows, err := r.db.QueryContext(ctx, query, userID, startAt.UTC().Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	activities := []domain.GoalActivity{}
	for rows.Next() {
		var activity domain.GoalActivity
		var dateStr string
		if err := rows.Scan(&dateStr, &activity.Activity); err != nil {
			return nil, err
		}
		activity.Date, err = time.Parse(time.DateOnly, dateStr)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}
	return activities, rows.Err()
}

func (r *ReportRepository) RecordGoalActivity(ctx context.Context, userID string, activityAt time.Time, activityText string) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}

	activityAtUTC := activityAt.UTC()
	activityAtStr := activityAtUTC.Format(time.RFC3339)
	var goalStartAt, completedAt string
	var targetDays int
	err = tx.QueryRowContext(ctx, `
		SELECT start_at, target_days, COALESCE(completed_at, '')
		FROM goals
		WHERE user_id = ? AND start_at <= ? AND end_at > ?
		ORDER BY start_at DESC
		LIMIT 1
	`, userID, activityAtStr, activityAtStr).Scan(&goalStartAt, &targetDays, &completedAt)
	if err == sql.ErrNoRows {
		return false, tx.Commit()
	}
	if err != nil {
		_ = tx.Rollback()
		return false, err
	}

	activityDate := domain.GetToday(activityAtUTC)
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO goal_activity_logs (user_id, goal_start_at, activity_date, activity_text, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(user_id, goal_start_at, activity_date) DO UPDATE SET
			activity_text = CASE WHEN COALESCE(activity_text, '') = '' THEN excluded.activity_text ELSE activity_text END
	`, userID, goalStartAt, activityDate.Format(time.DateOnly), activityText, activityAtStr); err != nil {
		_ = tx.Rollback()
		return false, err
	}

	if completedAt != "" {
		return false, tx.Commit()
	}

	var activeDays int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM goal_activity_logs
		WHERE user_id = ? AND goal_start_at = ?
	`, userID, goalStartAt).Scan(&activeDays); err != nil {
		_ = tx.Rollback()
		return false, err
	}
	if activeDays < targetDays {
		return false, tx.Commit()
	}

	now := time.Now().UTC().Format(time.RFC3339)
	res, err := tx.ExecContext(ctx, `
		UPDATE goals
		SET completed_at = ?
		WHERE user_id = ? AND start_at = ? AND COALESCE(completed_at, '') = ''
	`, now, userID, goalStartAt)
	if err != nil {
		_ = tx.Rollback()
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return false, err
	}
	if affected == 0 {
		return false, tx.Commit()
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE user_reports
		SET goals_completed = COALESCE(goals_completed, 0) + 1
		WHERE user_id = ?
	`, userID); err != nil {
		_ = tx.Rollback()
		return false, err
	}

	return true, tx.Commit()
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
