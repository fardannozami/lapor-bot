package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Mirrors domain.NumericLevelFromTotalPoints
func numericLevelFromTotalPoints(totalPoints int) int {
	if totalPoints < 0 {
		totalPoints = 0
	}
	level := 0
	remaining := totalPoints
	for {
		required := 5*level*level + 50*level + 100
		if remaining < required {
			return level
		}
		remaining -= required
		level++
	}
}

func main() {
	dbPath := flag.String("db", "./data/whatsapp.db", "Path to SQLite database")
	userID := flag.String("user-id", "hunter-e1ec69226d-72", "User ID to recover")
	dryRun := flag.Bool("dry-run", false, "Preview changes without committing")
	flag.Parse()

	if _, err := os.Stat(*dbPath); os.IsNotExist(err) {
		log.Fatalf("Database not found at %s", *dbPath)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", *dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	fmt.Println("========================================")
	fmt.Printf("🔧 Recovery Tool for user: %s\n", *userID)
	fmt.Printf("📁 Database: %s\n", *dbPath)
	if *dryRun {
		fmt.Println("🔍 Mode: DRY RUN (no changes will be made)")
	} else {
		fmt.Println("⚡ Mode: LIVE (changes will be committed)")
	}
	fmt.Println("========================================")
	fmt.Println()

	// ─── Step 1: Read current (corrupted) state ───
	var (
		name            string
		jobClass        sql.NullString
		streak          int
		activityCount   int
		lastReportDate  string
		maxStreak       int
		totalPoints     int
		level           int
		achievements    string
		comebackStreak  int
		inactiveDays    int
		centurionCycles int
		seasonalPoints  int
		seasonalActCnt  int
		seasonalMaxStr  int
		seasonalAch     string
		streakFreezes   int
		goalsCompleted  int
		totalSQ         int
		seasonalSQ      int
		str_, sta, agi, vit int
	)

	err = db.QueryRowContext(ctx, `
		SELECT name, COALESCE(job_class, ''), streak, activity_count, last_report_date,
		       COALESCE(max_streak,0), COALESCE(total_points,0), COALESCE(level,0),
		       COALESCE(achievements,''), COALESCE(comeback_streak,0), COALESCE(inactive_days,0),
		       COALESCE(centurion_cycles,0), COALESCE(seasonal_points,0),
		       COALESCE(seasonal_activity_count,0), COALESCE(seasonal_max_streak,0),
		       COALESCE(seasonal_achievements,''), COALESCE(streak_freezes,0),
		       COALESCE(goals_completed,0), COALESCE(total_side_quests,0),
		       COALESCE(seasonal_side_quests,0),
		       COALESCE(str,0), COALESCE(sta,0), COALESCE(agi,0), COALESCE(vit,0)
		FROM user_reports WHERE user_id = ?`, *userID).Scan(
		&name, &jobClass, &streak, &activityCount, &lastReportDate,
		&maxStreak, &totalPoints, &level, &achievements, &comebackStreak,
		&inactiveDays, &centurionCycles, &seasonalPoints,
		&seasonalActCnt, &seasonalMaxStr, &seasonalAch, &streakFreezes,
		&goalsCompleted, &totalSQ, &seasonalSQ,
		&str_, &sta, &agi, &vit,
	)
	if err == sql.ErrNoRows {
		log.Fatalf("User %s not found in user_reports", *userID)
	}
	if err != nil {
		log.Fatalf("Failed to read current state: %v", err)
	}

	fmt.Println("📊 CURRENT (corrupted) state:")
	fmt.Printf("   Name:                 %s\n", name)
	fmt.Printf("   Job Class:            %s\n", jobClass.String)
	fmt.Printf("   Streak:               %d\n", streak)
	fmt.Printf("   Max Streak:           %d\n", maxStreak)
	fmt.Printf("   Activity Count:       %d\n", activityCount)
	fmt.Printf("   Last Report Date:     %s\n", lastReportDate)
	fmt.Printf("   Total Points:         %d\n", totalPoints)
	fmt.Printf("   Level:                %d\n", level)
	fmt.Printf("   Achievements:         %s\n", achievements)
	fmt.Printf("   Comeback Streak:      %d\n", comebackStreak)
	fmt.Printf("   Inactive Days:        %d\n", inactiveDays)
	fmt.Printf("   Centurion Cycles:     %d\n", centurionCycles)
	fmt.Printf("   Seasonal Points:      %d\n", seasonalPoints)
	fmt.Printf("   Seasonal Activity:    %d\n", seasonalActCnt)
	fmt.Printf("   Seasonal Max Streak:  %d\n", seasonalMaxStr)
	fmt.Printf("   Seasonal Achievements:%s\n", seasonalAch)
	fmt.Printf("   Streak Freezes:       %d\n", streakFreezes)
	fmt.Printf("   Goals Completed:      %d\n", goalsCompleted)
	fmt.Printf("   Total Side Quests:    %d\n", totalSQ)
	fmt.Printf("   Seasonal Side Quests: %d\n", seasonalSQ)
	fmt.Printf("   STR/STA/AGI/VIT:      %d/%d/%d/%d\n", str_, sta, agi, vit)
	fmt.Println()

	// ─── Step 2: Check if report_events table exists ───
	var hasReportEvents bool
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='report_events'`).Scan(&hasReportEvents)
	if err != nil {
		log.Fatalf("Failed to check for report_events table: %v", err)
	}

	var hasUserSeasonStats bool
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='user_season_stats'`).Scan(&hasUserSeasonStats)
	if err != nil {
		log.Fatalf("Failed to check for user_season_stats table: %v", err)
	}

	var hasUserDailyActivity bool
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='user_daily_activity'`).Scan(&hasUserDailyActivity)
	if err != nil {
		log.Fatalf("Failed to check for user_daily_activity table: %v", err)
	}

	// ─── Step 3: Calculate true lifetime points from report_events ───
	var trueLifetimePoints int
	var eventCount int

	if hasReportEvents {
		err = db.QueryRowContext(ctx, `
			SELECT COALESCE(SUM(points_delta), 0), COUNT(*)
			FROM report_events
			WHERE user_id = ?`, *userID).Scan(&trueLifetimePoints, &eventCount)
		if err != nil {
			log.Fatalf("Failed to sum report_events: %v", err)
		}
		fmt.Printf("📋 report_events: Found %d events, total points_delta = %d\n", eventCount, trueLifetimePoints)
	} else {
		fmt.Println("⚠️  report_events table does not exist. Cannot recover points from events.")
		fmt.Println("    Will keep current total_points value.")
		trueLifetimePoints = totalPoints
	}

	// ─── Step 4: Get all activity dates from activity_logs ───
	type activityLogEntry struct {
		date         time.Time
		regularCount int
		sqCount      int
	}

	var activityEntries []activityLogEntry
	rows, err := db.QueryContext(ctx, `
		SELECT activity_date, COALESCE(regular_report_count, report_count), COALESCE(sidequest_count, 0)
		FROM activity_logs
		WHERE user_id = ?
		ORDER BY activity_date ASC`, *userID)
	if err != nil {
		log.Fatalf("Failed to query activity_logs: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var dateStr string
		var rc, sc int
		if err := rows.Scan(&dateStr, &rc, &sc); err != nil {
			log.Fatalf("Failed to scan activity_log: %v", err)
		}
		loc := time.FixedZone("WIB", 7*3600)
		d, err := time.ParseInLocation(time.DateOnly, dateStr, loc)
		if err != nil {
			log.Printf("Warning: skipping unparseable date %q: %v", dateStr, err)
			continue
		}
		activityEntries = append(activityEntries, activityLogEntry{date: d, regularCount: rc, sqCount: sc})
	}

	fmt.Printf("📋 activity_logs: Found %d day entries for user\n", len(activityEntries))

	// ─── Step 5: Calculate true seasonal stats from user_season_stats or report_events ───
	var trueSeasonalPoints int
	var trueSeasonalActivityCount int
	var trueSeasonalSideQuests int

	if hasUserSeasonStats {
		err = db.QueryRowContext(ctx, `
			SELECT COALESCE(total_points, 0), COALESCE(active_days, 0),
			       COALESCE(sidequest_reports, 0)
			FROM user_season_stats
			WHERE user_id = ? AND season_number = 1`, *userID).Scan(
			&trueSeasonalPoints, &trueSeasonalActivityCount, &trueSeasonalSideQuests)
		if err == sql.ErrNoRows {
			fmt.Println("⚠️  No user_season_stats record for season 1.")
			trueSeasonalPoints = seasonalPoints
			trueSeasonalActivityCount = seasonalActCnt
			trueSeasonalSideQuests = seasonalSQ
		} else if err != nil {
			log.Fatalf("Failed to query user_season_stats: %v", err)
		} else {
			fmt.Printf("📋 user_season_stats (season 1): points=%d, active_days=%d, sidequests=%d\n",
				trueSeasonalPoints, trueSeasonalActivityCount, trueSeasonalSideQuests)
		}
	} else if hasReportEvents {
		err = db.QueryRowContext(ctx, `
			SELECT COALESCE(SUM(points_delta), 0),
			       COUNT(DISTINCT activity_date),
			       COALESCE(SUM(sidequest_count_delta), 0)
			FROM report_events
			WHERE user_id = ? AND season_number = 1`, *userID).Scan(
			&trueSeasonalPoints, &trueSeasonalActivityCount, &trueSeasonalSideQuests)
		if err != nil {
			log.Fatalf("Failed to calculate seasonal stats from report_events: %v", err)
		}
		fmt.Printf("📋 report_events (season 1): points=%d, active_days=%d, sidequests=%d\n",
			trueSeasonalPoints, trueSeasonalActivityCount, trueSeasonalSideQuests)
	} else {
		fmt.Println("⚠️  Cannot recover seasonal stats. Keeping current values.")
		trueSeasonalPoints = seasonalPoints
		trueSeasonalActivityCount = seasonalActCnt
		trueSeasonalSideQuests = seasonalSQ
	}

	// ─── Step 6: Calculate true total side quests from report_events ───
	var trueTotalSideQuests int
	if hasReportEvents {
		err = db.QueryRowContext(ctx, `
			SELECT COALESCE(SUM(sidequest_count_delta), 0)
			FROM report_events
			WHERE user_id = ?`, *userID).Scan(&trueTotalSideQuests)
		if err != nil {
			log.Printf("Warning: failed to calculate total side quests: %v", err)
			trueTotalSideQuests = totalSQ
		}
	} else {
		trueTotalSideQuests = totalSQ
	}

	// ─── Step 7: Calculate true streak from activity_logs ───
	var regularDates []time.Time
	for _, e := range activityEntries {
		if e.regularCount > 0 {
			regularDates = append(regularDates, e.date)
		}
	}

	sort.Slice(regularDates, func(i, j int) bool {
		return regularDates[i].Before(regularDates[j])
	})

	trueActivityCount := len(regularDates)
	trueCenturionCycles := 0
	if trueActivityCount > 100 {
		trueCenturionCycles = (trueActivityCount - 1) / 100
		trueActivityCount = ((trueActivityCount - 1) % 100) + 1
	}

	// Calculate streaks from ISO weeks
	seenWeeks := make(map[string]bool)
	var weeks []time.Time
	for _, d := range regularDates {
		ws := getStartOfISOWeek(d)
		key := ws.Format(time.DateOnly)
		if !seenWeeks[key] {
			seenWeeks[key] = true
			weeks = append(weeks, ws)
		}
	}

	trueStreak, trueMaxStreak, trueComebackStreak, trueInactiveDays := calculateStreaksFromWeeks(weeks)

	// Last report date: keep the existing one (has precise time info from the actual report)
	trueLastReportDate := lastReportDate

	// ─── Step 8: Calculate seasonal max streak ───
	loc := time.FixedZone("WIB", 7*3600)
	seasonStart := time.Date(2026, time.May, 1, 0, 0, 0, 0, loc)
	var seasonalWeeks []time.Time
	seasonalWeeksSeen := make(map[string]bool)
	for _, d := range regularDates {
		if d.Before(seasonStart) {
			continue
		}
		ws := getStartOfISOWeek(d)
		key := ws.Format(time.DateOnly)
		if !seasonalWeeksSeen[key] {
			seasonalWeeksSeen[key] = true
			seasonalWeeks = append(seasonalWeeks, ws)
		}
	}
	_, trueSeasonalMaxStreak, _, _ := calculateStreaksFromWeeks(seasonalWeeks)

	// ─── Step 9: Keep existing achievements ───
	trueAchievements := achievements
	trueSeasonalAchievements := seasonalAch

	// ─── Step 10: Recalculate level ───
	trueLevel := numericLevelFromTotalPoints(trueLifetimePoints)

	// ─── Summary ───
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("📊 RECOVERED state:")
	fmt.Printf("   Name:                 %s\n", name)
	fmt.Printf("   Job Class:            %s (unchanged)\n", jobClass.String)
	fmt.Printf("   Streak:               %d → %d\n", streak, trueStreak)
	fmt.Printf("   Max Streak:           %d → %d\n", maxStreak, trueMaxStreak)
	fmt.Printf("   Activity Count:       %d → %d\n", activityCount, trueActivityCount)
	fmt.Printf("   Last Report Date:     %s (unchanged)\n", trueLastReportDate)
	fmt.Printf("   Total Points:         %d → %d\n", totalPoints, trueLifetimePoints)
	fmt.Printf("   Level:                %d → %d\n", level, trueLevel)
	fmt.Printf("   Achievements:         %s (unchanged)\n", trueAchievements)
	fmt.Printf("   Comeback Streak:      %d → %d\n", comebackStreak, trueComebackStreak)
	fmt.Printf("   Inactive Days:        %d → %d\n", inactiveDays, trueInactiveDays)
	fmt.Printf("   Centurion Cycles:     %d → %d\n", centurionCycles, trueCenturionCycles)
	fmt.Printf("   Seasonal Points:      %d → %d\n", seasonalPoints, trueSeasonalPoints)
	fmt.Printf("   Seasonal Activity:    %d → %d\n", seasonalActCnt, trueSeasonalActivityCount)
	fmt.Printf("   Seasonal Max Streak:  %d → %d\n", seasonalMaxStr, trueSeasonalMaxStreak)
	fmt.Printf("   Seasonal Achievements:%s (unchanged)\n", trueSeasonalAchievements)
	fmt.Printf("   Streak Freezes:       %d (unchanged)\n", streakFreezes)
	fmt.Printf("   Goals Completed:      %d (unchanged)\n", goalsCompleted)
	fmt.Printf("   Total Side Quests:    %d → %d\n", totalSQ, trueTotalSideQuests)
	fmt.Printf("   Seasonal Side Quests: %d → %d\n", seasonalSQ, trueSeasonalSideQuests)
	fmt.Printf("   STR/STA/AGI/VIT:      %d/%d/%d/%d (unchanged)\n", str_, sta, agi, vit)
	fmt.Println("========================================")
	fmt.Println()

	if *dryRun {
		fmt.Println("🔍 DRY RUN complete. No changes were made.")
		fmt.Println("   Run without -dry-run to apply the recovery.")
		return
	}

	// ─── Step 11: Apply recovery (ONLY for this user) ───
	result, err := db.ExecContext(ctx, `
		UPDATE user_reports SET
			streak = ?,
			max_streak = ?,
			activity_count = ?,
			total_points = ?,
			level = ?,
			achievements = ?,
			comeback_streak = ?,
			inactive_days = ?,
			centurion_cycles = ?,
			seasonal_points = ?,
			seasonal_activity_count = ?,
			seasonal_max_streak = ?,
			seasonal_achievements = ?,
			total_side_quests = ?,
			seasonal_side_quests = ?
		WHERE user_id = ?`,
		trueStreak,
		trueMaxStreak,
		trueActivityCount,
		trueLifetimePoints,
		trueLevel,
		trueAchievements,
		trueComebackStreak,
		trueInactiveDays,
		trueCenturionCycles,
		trueSeasonalPoints,
		trueSeasonalActivityCount,
		trueSeasonalMaxStreak,
		trueSeasonalAchievements,
		trueTotalSideQuests,
		trueSeasonalSideQuests,
		*userID,
	)
	if err != nil {
		log.Fatalf("❌ Failed to update user_reports: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		fmt.Println("⚠️  No rows updated. User may not exist in user_reports.")
	} else {
		fmt.Printf("✅ Successfully recovered %s's progress! (%d row updated)\n", name, rowsAffected)
		fmt.Println("   Only hunter-e1ec69226d-72 was modified. No other data was touched.")
	}
}

func getStartOfISOWeek(t time.Time) time.Time {
	loc := time.FixedZone("WIB", 7*3600)
	t = t.In(loc)
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	monday := t.AddDate(0, 0, -int(weekday-time.Monday))
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, loc)
}

func calculateStreaksFromWeeks(weeks []time.Time) (streak, maxStreak, comebackStreak, inactiveDays int) {
	if len(weeks) == 0 {
		return 0, 0, 0, 0
	}

	streak = 1
	maxStreak = 1
	comebackStreak = 1
	inactiveDays = 0

	var lastGapDays int

	for i := 1; i < len(weeks); i++ {
		prevWeek := weeks[i-1]
		currWeek := weeks[i]
		weeksDiff := int(math.Round(currWeek.Sub(prevWeek).Hours() / (24 * 7)))

		if weeksDiff == 1 {
			streak++
			comebackStreak++
		} else {
			if streak > maxStreak {
				maxStreak = streak
			}
			lastGapDays = int(math.Round(currWeek.Sub(prevWeek.AddDate(0, 0, 7)).Hours() / 24))
			streak = 1
			comebackStreak = 1
		}
	}

	if streak > maxStreak {
		maxStreak = streak
	}

	if len(weeks) >= 2 {
		lastTwoDiff := int(math.Round(weeks[len(weeks)-1].Sub(weeks[len(weeks)-2]).Hours() / (24 * 7)))
		if lastTwoDiff > 1 {
			inactiveDays = int(math.Round(weeks[len(weeks)-1].Sub(weeks[len(weeks)-2].AddDate(0, 0, 7)).Hours() / 24))
		} else if lastGapDays > 0 {
			inactiveDays = lastGapDays
		}
	}

	_ = lastGapDays
	return streak, maxStreak, comebackStreak, inactiveDays
}

// strJoin is a small helper for printing (unused but kept for potential debug).
func strJoin(ss []string, sep string) string {
	return strings.Join(ss, sep)
}
