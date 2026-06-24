package sqlite

import (
	"context"
	"database/sql"
	"fmt"
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
	COALESCE(streak_freezes, 0), COALESCE(goals_completed, 0),
	COALESCE(total_side_quests, 0), COALESCE(seasonal_side_quests, 0),
	COALESCE(str, 0), COALESCE(sta, 0), COALESCE(agi, 0), COALESCE(vit, 0)`

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
		&report.TotalSideQuests, &report.SeasonalSideQuests,
		&report.Str, &report.Sta, &report.Agi, &report.Vit,
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
		INSERT INTO user_reports (user_id, name, job_class, streak, activity_count, last_report_date, max_streak, total_points, level, achievements, comeback_streak, inactive_days, centurion_cycles, seasonal_points, seasonal_activity_count, seasonal_max_streak, seasonal_achievements, streak_freezes, goals_completed, total_side_quests, seasonal_side_quests, str, sta, agi, vit)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
			streak_freezes = excluded.streak_freezes,
			total_side_quests = excluded.total_side_quests,
			seasonal_side_quests = excluded.seasonal_side_quests,
			str = excluded.str,
			sta = excluded.sta,
			agi = excluded.agi,
			vit = excluded.vit
	`
	_, err := execer.ExecContext(ctx, query,
		report.UserID, report.Name, report.JobClass, report.Streak, report.ActivityCount,
		report.LastReportDate.Format(time.RFC3339), report.MaxStreak, report.TotalPoints,
		report.Level, report.Achievements, report.ComebackStreak, report.InactiveDays, report.CenturionCycles,
		report.SeasonalPoints, report.SeasonalActivityCount, report.SeasonalMaxStreak,
		report.SeasonalAchievements, report.StreakFreezes, report.GoalsCompleted,
		report.TotalSideQuests, report.SeasonalSideQuests,
		report.Str, report.Sta, report.Agi, report.Vit,
	)
	return err
}

func (r *ReportRepository) UpsertReport(ctx context.Context, report *domain.Report) error {
	return upsertReport(ctx, r.db, report)
}

func logActivity(ctx context.Context, execer execContexter, userID string, activityDate time.Time) error {
	return logActivityKind(ctx, execer, userID, activityDate, domain.ActivityKindRegularReport)
}

func logActivityEvent(ctx context.Context, execer execContexter, event domain.ReportActivityEvent) error {
	regularIncrement := event.RegularCountDelta
	sideQuestIncrement := event.SideQuestCountDelta
	if regularIncrement < 0 {
		regularIncrement = 0
	}
	if sideQuestIncrement < 0 {
		sideQuestIncrement = 0
	}
	totalIncrement := regularIncrement + sideQuestIncrement
	if totalIncrement == 0 {
		totalIncrement = 1
	}

	query := `
		INSERT INTO activity_logs (user_id, activity_date, created_at, report_count, regular_report_count, sidequest_count, activity_text)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, activity_date) DO UPDATE SET
			report_count = report_count + excluded.report_count,
			regular_report_count = COALESCE(regular_report_count, 0) + excluded.regular_report_count,
			sidequest_count = COALESCE(sidequest_count, 0) + excluded.sidequest_count,
			activity_text = CASE
				WHEN excluded.activity_text = '' THEN activity_text
				WHEN activity_text = '' THEN excluded.activity_text
				ELSE activity_text || '\n' || excluded.activity_text
			END,
			created_at = excluded.created_at
	`
	_, err := execer.ExecContext(ctx, query,
		event.UserID,
		event.ActivityDate.Format(time.DateOnly),
		event.OccurredAt.UTC().Format(time.RFC3339),
		totalIncrement,
		regularIncrement,
		sideQuestIncrement,
		event.ActivityText,
	)
	return err
}

func logActivityKind(ctx context.Context, execer execContexter, userID string, activityDate time.Time, kind string) error {
	regularIncrement := 0
	sideQuestIncrement := 0
	if kind == domain.ActivityKindSideQuest {
		sideQuestIncrement = 1
	} else {
		regularIncrement = 1
	}

	query := `
		INSERT INTO activity_logs (user_id, activity_date, created_at, report_count, regular_report_count, sidequest_count)
		VALUES (?, ?, ?, 1, ?, ?)
		ON CONFLICT(user_id, activity_date) DO UPDATE SET
			report_count = report_count + 1,
			regular_report_count = COALESCE(regular_report_count, 0) + excluded.regular_report_count,
			sidequest_count = COALESCE(sidequest_count, 0) + excluded.sidequest_count,
			created_at = excluded.created_at
	`
	_, err := execer.ExecContext(ctx, query, userID, activityDate.Format(time.DateOnly), time.Now().UTC().Format(time.RFC3339), regularIncrement, sideQuestIncrement)
	return err
}

func (r *ReportRepository) UpsertReportWithActivity(ctx context.Context, report *domain.Report, activityDate time.Time) error {
	return r.UpsertReportWithActivityKind(ctx, report, activityDate, domain.ActivityKindRegularReport)
}

func (r *ReportRepository) UpsertReportWithActivityKind(ctx context.Context, report *domain.Report, activityDate time.Time, kind string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := upsertReport(ctx, tx, report); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := logActivityKind(ctx, tx, report.UserID, activityDate, kind); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *ReportRepository) UpsertReportWithActivityEvent(ctx context.Context, report *domain.Report, event domain.ReportActivityEvent) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if event.UserID == "" {
		event.UserID = report.UserID
	}
	if event.Kind == "" {
		event.Kind = domain.ActivityKindRegularReport
	}
	if event.ActivityDate.IsZero() {
		event.ActivityDate = domain.GetToday(report.LastReportDate)
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now().UTC()
	}
	if event.RuleVersion == 0 {
		event.RuleVersion = 1
	}
	if event.Source == "" {
		event.Source = "whatsapp"
	}
	if event.MetadataJSON == "" {
		event.MetadataJSON = "{}"
	}

	if err := upsertReport(ctx, tx, report); err != nil {
		_ = tx.Rollback()
		return err
	}
	inserted, err := insertReportEvent(ctx, tx, event)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if inserted {
		if err := logActivityEvent(ctx, tx, event); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := upsertDailyActivityProjection(ctx, tx, event); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := upsertSeasonStatsProjection(ctx, tx, event); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func insertReportEvent(ctx context.Context, execer execContexter, event domain.ReportActivityEvent) (bool, error) {
	query := `
		INSERT OR IGNORE INTO report_events (
			event_id, user_id, season_number, kind, activity_date,
			occurred_at_utc, recorded_at_utc, points_delta,
			regular_count_delta, sidequest_count_delta, rule_version,
			source, activity_text, metadata_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := execer.ExecContext(ctx, query,
		event.EventID,
		event.UserID,
		event.SeasonNumber,
		event.Kind,
		event.ActivityDate.Format(time.DateOnly),
		event.OccurredAt.UTC().Format(time.RFC3339),
		time.Now().UTC().Format(time.RFC3339),
		event.PointsDelta,
		event.RegularCountDelta,
		event.SideQuestCountDelta,
		event.RuleVersion,
		event.Source,
		event.ActivityText,
		event.MetadataJSON,
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func upsertDailyActivityProjection(ctx context.Context, execer execContexter, event domain.ReportActivityEvent) error {
	query := `
		INSERT INTO user_daily_activity (
			user_id, season_number, activity_date,
			regular_count, sidequest_count, total_points,
			first_event_id, last_event_id, first_reported_at_utc, last_reported_at_utc
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, season_number, activity_date) DO UPDATE SET
			regular_count = regular_count + excluded.regular_count,
			sidequest_count = sidequest_count + excluded.sidequest_count,
			total_points = total_points + excluded.total_points,
			last_event_id = excluded.last_event_id,
			last_reported_at_utc = excluded.last_reported_at_utc
	`
	_, err := execer.ExecContext(ctx, query,
		event.UserID,
		event.SeasonNumber,
		event.ActivityDate.Format(time.DateOnly),
		event.RegularCountDelta,
		event.SideQuestCountDelta,
		event.PointsDelta,
		event.EventID,
		event.EventID,
		event.OccurredAt.UTC().Format(time.RFC3339),
		event.OccurredAt.UTC().Format(time.RFC3339),
	)
	return err
}

func upsertSeasonStatsProjection(ctx context.Context, execer execContexter, event domain.ReportActivityEvent) error {
	activityDate := event.ActivityDate.Format(time.DateOnly)
	occurredAt := event.OccurredAt.UTC().Format(time.RFC3339)
	query := `
		INSERT INTO user_season_stats (
			user_id, season_number, total_points,
			regular_reports, sidequest_reports, active_days,
			first_activity_date, last_activity_date,
			first_reported_at_utc, last_reported_at_utc
		) VALUES (?, ?, ?, ?, ?, 1, ?, ?, ?, ?)
		ON CONFLICT(user_id, season_number) DO UPDATE SET
			total_points = total_points + excluded.total_points,
			regular_reports = regular_reports + excluded.regular_reports,
			sidequest_reports = sidequest_reports + excluded.sidequest_reports,
			active_days = active_days + CASE
				WHEN EXISTS (
					SELECT 1 FROM user_daily_activity uda
					WHERE uda.user_id = excluded.user_id
					  AND uda.season_number = excluded.season_number
					  AND uda.activity_date = excluded.last_activity_date
					  AND (uda.regular_count + uda.sidequest_count) > (excluded.regular_reports + excluded.sidequest_reports)
				) THEN 0 ELSE 1
			END,
			first_activity_date = MIN(first_activity_date, excluded.first_activity_date),
			last_activity_date = MAX(last_activity_date, excluded.last_activity_date),
			last_reported_at_utc = excluded.last_reported_at_utc
	`
	_, err := execer.ExecContext(ctx, query,
		event.UserID,
		event.SeasonNumber,
		event.PointsDelta,
		event.RegularCountDelta,
		event.SideQuestCountDelta,
		activityDate,
		activityDate,
		occurredAt,
		occurredAt,
	)
	return err
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
		    seasonal_side_quests = 0,
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
			regular_report_count INTEGER NOT NULL DEFAULT 0,
			sidequest_count INTEGER NOT NULL DEFAULT 0,
			activity_text TEXT DEFAULT '',
			PRIMARY KEY (user_id, activity_date),
			FOREIGN KEY (user_id) REFERENCES user_reports(user_id) ON DELETE CASCADE
		);
	`
	_, err = r.db.ExecContext(ctx, activityLogQuery)
	if err != nil {
		return err
	}

	reportEventsQuery := `
		CREATE TABLE IF NOT EXISTS report_events (
			event_id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			season_number INTEGER NOT NULL,
			kind TEXT NOT NULL CHECK (kind IN ('regular_report', 'sidequest')),
			activity_date TEXT NOT NULL,
			occurred_at_utc TEXT NOT NULL,
			recorded_at_utc TEXT NOT NULL,
			points_delta INTEGER NOT NULL,
			regular_count_delta INTEGER NOT NULL DEFAULT 0,
			sidequest_count_delta INTEGER NOT NULL DEFAULT 0,
			rule_version INTEGER NOT NULL DEFAULT 1,
			source TEXT NOT NULL DEFAULT 'whatsapp',
			activity_text TEXT NOT NULL DEFAULT '',
			metadata_json TEXT NOT NULL DEFAULT '{}',
			FOREIGN KEY (user_id) REFERENCES user_reports(user_id) ON DELETE CASCADE
		);
	`
	_, err = r.db.ExecContext(ctx, reportEventsQuery)
	if err != nil {
		return err
	}

	dailyActivityQuery := `
		CREATE TABLE IF NOT EXISTS user_daily_activity (
			user_id TEXT NOT NULL,
			season_number INTEGER NOT NULL,
			activity_date TEXT NOT NULL,
			regular_count INTEGER NOT NULL DEFAULT 0,
			sidequest_count INTEGER NOT NULL DEFAULT 0,
			total_points INTEGER NOT NULL DEFAULT 0,
			first_event_id TEXT,
			last_event_id TEXT,
			first_reported_at_utc TEXT,
			last_reported_at_utc TEXT,
			PRIMARY KEY (user_id, season_number, activity_date),
			FOREIGN KEY (user_id) REFERENCES user_reports(user_id) ON DELETE CASCADE
		);
	`
	_, err = r.db.ExecContext(ctx, dailyActivityQuery)
	if err != nil {
		return err
	}

	seasonStatsQuery := `
		CREATE TABLE IF NOT EXISTS user_season_stats (
			user_id TEXT NOT NULL,
			season_number INTEGER NOT NULL,
			total_points INTEGER NOT NULL DEFAULT 0,
			regular_reports INTEGER NOT NULL DEFAULT 0,
			sidequest_reports INTEGER NOT NULL DEFAULT 0,
			active_days INTEGER NOT NULL DEFAULT 0,
			first_activity_date TEXT,
			last_activity_date TEXT,
			first_reported_at_utc TEXT,
			last_reported_at_utc TEXT,
			PRIMARY KEY (user_id, season_number),
			FOREIGN KEY (user_id) REFERENCES user_reports(user_id) ON DELETE CASCADE
		);
	`
	_, err = r.db.ExecContext(ctx, seasonStatsQuery)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_activity_logs_date ON activity_logs (activity_date)`)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_report_events_user_season_date ON report_events (user_id, season_number, activity_date)`)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_report_events_season_date ON report_events (season_number, activity_date)`)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_user_daily_activity_season_date ON user_daily_activity (season_number, activity_date)`)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_user_season_stats_leaderboard ON user_season_stats (season_number, total_points DESC, regular_reports DESC, sidequest_reports DESC, user_id ASC)`)
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
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN total_side_quests INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN seasonal_side_quests INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN str INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN sta INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN agi INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE user_reports ADD COLUMN vit INTEGER DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE activity_logs ADD COLUMN report_count INTEGER NOT NULL DEFAULT 1")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE activity_logs ADD COLUMN regular_report_count INTEGER NOT NULL DEFAULT 0")
	_, _ = r.db.ExecContext(ctx, "ALTER TABLE activity_logs ADD COLUMN sidequest_count INTEGER NOT NULL DEFAULT 0")
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

	dailyQuestTableQuery := `
		CREATE TABLE IF NOT EXISTS daily_quests (
			user_id TEXT NOT NULL,
			quest_date TEXT NOT NULL,
			tasks_json TEXT NOT NULL,
			PRIMARY KEY (user_id, quest_date)
		);
	`
	_, err = r.db.ExecContext(ctx, dailyQuestTableQuery)
	if err != nil {
		return err
	}

	jobClassesQuery := `
		CREATE TABLE IF NOT EXISTS job_classes (
			id          TEXT PRIMARY KEY,
			name        TEXT NOT NULL,
			icon        TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			trait       TEXT NOT NULL DEFAULT '',
			sort_order  INTEGER NOT NULL DEFAULT 0
		);
		INSERT OR IGNORE INTO job_classes (id, name, icon, description, trait, sort_order) VALUES
		  ('fighter','Fighter','⚔️','Melee hunter yang mengandalkan disiplin, stamina, dan daya tahan.','cocok untuk yang suka latihan strength/functional',0),
		  ('tank','Tanker','🛡️','Frontliner yang kuat bertahan dan konsisten menjaga formasi.','cocok untuk yang fokus konsistensi dan habit jangka panjang',1),
		  ('assassin','Assassin','🗡️','Hunter cepat, gesit, dan tajam mengeksekusi sesi singkat tapi intens.','cocok untuk HIIT, sprint, atau workout cepat',2),
		  ('mage','Mage','🔥','Damage dealer jarak jauh dengan energi eksplosif dan variasi latihan.','cocok untuk yang suka eksplor banyak jenis olahraga',3),
		  ('ranger','Ranger','🏹','Hunter presisi yang unggul di endurance, pace, dan jarak.','cocok untuk lari, sepeda, jalan jauh, hiking',4),
		  ('healer','Healer','💚','Support hunter yang menjaga recovery, mobilitas, dan kesehatan jangka panjang.','cocok untuk yoga, mobility, recovery, pola hidup sehat',5),
		  ('necromancer','Necromancer','🌑','Hidden job yang bangkit dari kegagalan dan mengubah comeback jadi kekuatan.','cocok untuk comeback setelah absen dan bangun sistem baru',6);
	`
	_, err = r.db.ExecContext(ctx, jobClassesQuery)
	if err != nil {
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

func (r *ReportRepository) GetUserActivityDatesByKind(ctx context.Context, userID string, kind string) ([]time.Time, error) {
	condition := "COALESCE(regular_report_count, 0) > 0"
	if kind == domain.ActivityKindSideQuest {
		condition = "COALESCE(sidequest_count, 0) > 0"
	}

	query := fmt.Sprintf(`
		SELECT activity_date FROM activity_logs
		WHERE user_id = ? AND %s
		ORDER BY activity_date ASC
	`, condition)
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

func (r *ReportRepository) DeleteActivityLogByKind(ctx context.Context, userID string, activityDate time.Time, kind string) error {
	date := activityDate.Format(time.DateOnly)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	regularCount, sideQuestCount, err := getActivityKindCounts(ctx, tx, userID, date)
	if err == sql.ErrNoRows {
		_ = tx.Rollback()
		return nil
	}
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if kind == domain.ActivityKindSideQuest {
		sideQuestCount = 0
	} else {
		regularCount = 0
	}

	if err := saveActivityKindCounts(ctx, tx, userID, date, regularCount, sideQuestCount); err != nil {
		_ = tx.Rollback()
		return err
	}
	if kind != domain.ActivityKindSideQuest {
		if err := reconcileGoalAfterActivityDelete(ctx, tx, userID, activityDate); err != nil {
			_ = tx.Rollback()
			return err
		}
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

func (r *ReportRepository) DeleteLatestActivityLogByKind(ctx context.Context, userID string, activityDate time.Time, kind string) (int, error) {
	date := activityDate.Format(time.DateOnly)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	regularCount, sideQuestCount, err := getActivityKindCounts(ctx, tx, userID, date)
	if err == sql.ErrNoRows {
		_ = tx.Rollback()
		return 0, nil
	}
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	remaining := 0
	if kind == domain.ActivityKindSideQuest {
		if sideQuestCount == 0 {
			_ = tx.Rollback()
			return 0, nil
		}
		sideQuestCount--
		remaining = sideQuestCount
	} else {
		if regularCount == 0 {
			_ = tx.Rollback()
			return 0, nil
		}
		regularCount--
		remaining = regularCount
	}

	if err := saveActivityKindCounts(ctx, tx, userID, date, regularCount, sideQuestCount); err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if kind != domain.ActivityKindSideQuest && remaining == 0 {
		if err := reconcileGoalAfterActivityDelete(ctx, tx, userID, activityDate); err != nil {
			_ = tx.Rollback()
			return 0, err
		}
	}

	return remaining, tx.Commit()
}

func getActivityKindCounts(ctx context.Context, tx *sql.Tx, userID, date string) (regularCount, sideQuestCount int, err error) {
	var totalCount int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(report_count, 1), COALESCE(regular_report_count, 0), COALESCE(sidequest_count, 0)
		FROM activity_logs
		WHERE user_id = ? AND activity_date = ?
	`, userID, date).Scan(&totalCount, &regularCount, &sideQuestCount)
	if err != nil {
		return 0, 0, err
	}
	if regularCount == 0 && sideQuestCount == 0 && totalCount > 0 {
		regularCount = totalCount
	}
	return regularCount, sideQuestCount, nil
}

func saveActivityKindCounts(ctx context.Context, tx *sql.Tx, userID, date string, regularCount, sideQuestCount int) error {
	if regularCount < 0 {
		regularCount = 0
	}
	if sideQuestCount < 0 {
		sideQuestCount = 0
	}
	totalCount := regularCount + sideQuestCount
	if totalCount == 0 {
		_, err := tx.ExecContext(ctx, `DELETE FROM activity_logs WHERE user_id = ? AND activity_date = ?`, userID, date)
		return err
	}
	_, err := tx.ExecContext(ctx, `
		UPDATE activity_logs
		SET report_count = ?, regular_report_count = ?, sidequest_count = ?
		WHERE user_id = ? AND activity_date = ?
	`, totalCount, regularCount, sideQuestCount, userID, date)
	return err
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
	if _, err := r.db.ExecContext(ctx, `DELETE FROM report_events WHERE user_id = ?`, userID); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `DELETE FROM user_daily_activity WHERE user_id = ?`, userID); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `DELETE FROM user_season_stats WHERE user_id = ?`, userID); err != nil {
		return err
	}
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

func (r *ReportRepository) GetDailyActivityCountByKind(ctx context.Context, userID string, date time.Time, kind string) (int, error) {
	column := "regular_report_count"
	if kind == domain.ActivityKindSideQuest {
		column = "sidequest_count"
	}

	query := `SELECT COALESCE(report_count, 1), COALESCE(regular_report_count, 0), COALESCE(sidequest_count, 0) FROM activity_logs WHERE user_id = ? AND activity_date = ?`
	var totalCount, regularCount, sideQuestCount int
	err := r.db.QueryRowContext(ctx, query, userID, date.Format(time.DateOnly)).Scan(&totalCount, &regularCount, &sideQuestCount)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if column == "sidequest_count" {
		return sideQuestCount, nil
	}
	if regularCount == 0 && sideQuestCount == 0 && totalCount > 0 {
		return totalCount, nil
	}
	return regularCount, nil
}

func (r *ReportRepository) SetGoal(ctx context.Context, goal *domain.WeeklyGoal) error {
	query := `
		INSERT INTO goals (user_id, target_days, activity, start_at, end_at, created_at, completed_at)
		SELECT ?, ?, ?, ?, ?, ?, ''
		WHERE NOT EXISTS (
			SELECT 1 FROM goals
			WHERE user_id = ? AND start_at <= ? AND end_at > ?
			AND COALESCE(completed_at, '') = ''
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

func (r *ReportRepository) SaveDailyQuest(ctx context.Context, userID, questDate, tasksJSON string) error {
	query := `
		INSERT INTO daily_quests (user_id, quest_date, tasks_json)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id, quest_date) DO UPDATE SET
			tasks_json = excluded.tasks_json
	`
	_, err := r.db.ExecContext(ctx, query, userID, questDate, tasksJSON)
	return err
}

func (r *ReportRepository) GetDailyQuest(ctx context.Context, userID, questDate string) (string, error) {
	query := `SELECT tasks_json FROM daily_quests WHERE user_id = ? AND quest_date = ?`
	var tasksJSON string
	err := r.db.QueryRowContext(ctx, query, userID, questDate).Scan(&tasksJSON)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return tasksJSON, err
}

func (r *ReportRepository) GetAllJobClasses(ctx context.Context) ([]domain.JobClass, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, icon, description, trait FROM job_classes ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []domain.JobClass
	for rows.Next() {
		var j domain.JobClass
		if err := rows.Scan(&j.ID, &j.Name, &j.Icon, &j.Description, &j.Trait); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

func (r *ReportRepository) GetJobClass(ctx context.Context, id string) (*domain.JobClass, error) {
	var j domain.JobClass
	err := r.db.QueryRowContext(ctx, `SELECT id, name, icon, description, trait FROM job_classes WHERE id = ?`, id).
		Scan(&j.ID, &j.Name, &j.Icon, &j.Description, &j.Trait)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &j, nil
}
