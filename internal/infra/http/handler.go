package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/app/usecase"
	"github.com/fardannozami/whatsapp-gateway/internal/config"
	"github.com/fardannozami/whatsapp-gateway/internal/domain"
	"github.com/fardannozami/whatsapp-gateway/internal/domain/phone"
	"go.mau.fi/whatsmeow"
)

type Server struct {
	repo        domain.ReportRepository
	linkUC      *usecase.LinkStravaUsecase
	processUC   *usecase.ProcessStravaWebhookUsecase
	waClient    *whatsmeow.Client
	verifyToken string
}

func NewServer(repo domain.ReportRepository, linkUC *usecase.LinkStravaUsecase, processUC *usecase.ProcessStravaWebhookUsecase, waClient *whatsmeow.Client, cfg config.Config) *Server {
	return &Server{
		repo:        repo,
		linkUC:      linkUC,
		processUC:   processUC,
		waClient:    waClient,
		verifyToken: cfg.StravaVerifyToken,
	}
}

func (s *Server) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/health", s.HandleHealth)
	mux.HandleFunc("/strava/link", s.HandleStravaLink)
	mux.HandleFunc("/strava/callback", s.HandleStravaCallback)
	mux.HandleFunc("/strava/webhook", s.HandleStravaWebhook)
	mux.HandleFunc("/api/leaderboard", s.HandleLeaderboard)
	mux.HandleFunc("/api/summary", s.HandleSummary)
	mux.HandleFunc("/api/motivation", s.HandleMotivation)
	mux.HandleFunc("/api/user", s.HandleGetUser)
	mux.HandleFunc("PATCH /api/user/name", s.HandleUpdateName)
	mux.HandleFunc("PATCH /api/user/job", s.HandleSelectJob)
	mux.HandleFunc("PATCH /api/user/goal", s.HandleSetGoal)
	mux.HandleFunc("GET /api/jobs", s.HandleListJobs)
	mux.HandleFunc("POST /api/user", s.HandleGetUserByPhone)
	mux.HandleFunc("/", s.HandleStatic)
}

func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if s.waClient == nil || !s.waClient.IsConnected() {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "Disconnected")
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func (s *Server) HandleStravaLink(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		name = "User"
	}
	authURL := s.linkUC.GetAuthURL(userID, name)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (s *Server) HandleStravaCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	userID := r.URL.Query().Get("state") // we passed userID in 'state'

	if code == "" || userID == "" {
		http.Error(w, "Invalid callback data", http.StatusBadRequest)
		return
	}

	if err := s.linkUC.HandleCallback(r.Context(), code, userID); err != nil {
		log.Printf("Strava callback error: %v", err)
		http.Error(w, fmt.Sprintf("Failed to link Strava: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h2>Success!</h2><p>Akun Strava kamu sudah terhubung dengan Bot Lapor. Kamu bisa menutup halaman ini.</p>")
}

func (s *Server) HandleStravaWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Verification (Challenge)
		challenge := r.URL.Query().Get("hub.challenge")
		verify := r.URL.Query().Get("hub.verify_token")

		if verify != s.verifyToken {
			http.Error(w, "Invalid verify token", http.StatusForbidden)
			return
		}

		log.Printf("Strava webhook challenge verification successful")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"hub.challenge": challenge})
		return
	}

	if r.Method == http.MethodPost {
		// Event processing
		var event usecase.StravaWebhookEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		log.Printf("Received Strava webhook: %+v", event)

		go func() {
			if err := s.processUC.Execute(context.Background(), s.waClient, event); err != nil {
				log.Printf("Strava webhook execution error: %v", err)
			}
		}()

		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// EnrichedReport holds parsed and visual RPG stats for rendering in the FE.
type EnrichedReport struct {
	UserID                string                      `json:"user_id"`
	Name                  string                      `json:"name"`
	JobClass              string                      `json:"job_class"`
	JobName               string                      `json:"job_name"`
	JobIcon               string                      `json:"job_icon"`
	JobDescription        string                      `json:"job_description"`
	JobTrait              string                      `json:"job_trait"`
	Streak                int                         `json:"streak"`
	ActivityCount         int                         `json:"activity_count"`
	TotalActiveDays       int                         `json:"total_active_days"`
	LastReportDate        string                      `json:"last_report_date"`
	MaxStreak             int                         `json:"max_streak"`
	TotalPoints           int                         `json:"total_points"`
	Level                 int                         `json:"level"`
	LevelName             string                      `json:"level_name"`
	LevelIcon             string                      `json:"level_icon"`
	XPProgress            domain.NumericLevelProgress `json:"xp_progress"`
	LevelTierProgress     TierProgress                `json:"level_tier_progress"`
	Achievements          []string                    `json:"achievements"`
	ComebackStreak        int                         `json:"comeback_streak"`
	InactiveDays          int                         `json:"inactive_days"`
	DaysSinceLastReport   int                         `json:"days_since_last_report"`
	CenturionCycles       int                         `json:"centurion_cycles"`
	SeasonalPoints        int                         `json:"seasonal_points"`
	SeasonalActivityCount int                         `json:"seasonal_activity_count"`
	SeasonalMaxStreak     int                         `json:"seasonal_max_streak"`
	SeasonalAchievements  []string                    `json:"seasonal_achievements"`
	StreakFreezes         int                         `json:"streak_freezes"`
	GoalsCompleted        int                         `json:"goals_completed"`
	TotalSideQuests       int                         `json:"total_side_quests"`
	SeasonalSideQuests    int                         `json:"seasonal_side_quests"`
	Str                   int                         `json:"str"`
	Sta                   int                         `json:"sta"`
	Agi                   int                         `json:"agi"`
	Vit                   int                         `json:"vit"`
	RankName              string                      `json:"rank_name"`
	RankIcon              string                      `json:"rank_icon"`
	SeasonRankProgress    TierProgress                `json:"season_rank_progress"`
	WeekActiveDays        int                         `json:"week_active_days"`
	WeekActivity          []bool                      `json:"week_activity"`
	EstimatedWeeklyPoints int                         `json:"estimated_weekly_points"`
	IsActiveToday         bool                        `json:"is_active_today"`
	DailyActivity         []DailyActivity             `json:"daily_activity,omitempty"`
	CurrentDailyStreak    int                         `json:"current_daily_streak,omitempty"`
	LongestDailyStreak    int                         `json:"longest_daily_streak,omitempty"`
	ActiveDaysInWindow    int                         `json:"active_days_in_window,omitempty"`
	ActiveGoal            *PersonalGoal               `json:"active_goal,omitempty"`
	TodaySideQuests       []domain.QuestTask          `json:"today_side_quests,omitempty"`
}

// TierProgress is precomputed for the web UI so templates/components only render it.
type TierProgress struct {
	CurrentMin int    `json:"current_min"`
	NextMin    int    `json:"next_min"`
	Value      int    `json:"value"`
	Percent    int    `json:"percent"`
	Remaining  int    `json:"remaining"`
	NextName   string `json:"next_name"`
	NextIcon   string `json:"next_icon"`
	IsMax      bool   `json:"is_max"`
}

// DailyActivity is a compact contribution-calendar cell for a personal profile.
type DailyActivity struct {
	Date   string `json:"date"`
	Count  int    `json:"count"`
	Active bool   `json:"active"`
}

type GoalDay struct {
	Date     string `json:"date"`
	DayLabel string `json:"day_label"`
	Activity string `json:"activity"`
	Active   bool   `json:"active"`
}

type PersonalGoal struct {
	TargetDays    int       `json:"target_days"`
	Activity      string    `json:"activity"`
	StartAt       string    `json:"start_at"`
	EndAt         string    `json:"end_at"`
	CompletedDays int       `json:"completed_days"`
	RemainingDays int       `json:"remaining_days"`
	Percent       int       `json:"percent"`
	IsCompleted   bool      `json:"is_completed"`
	CompletedAt   string   `json:"completed_at,omitempty"`
	Days          []GoalDay `json:"days"`
}

// GlobalSummary holds aggregated dashboard data.
type GlobalSummary struct {
	TotalParticipants   int            `json:"total_participants"`
	ActiveStreakCount   int            `json:"active_streak_count"`
	TotalWorkoutsLogged int            `json:"total_workouts_logged"`
	ActiveJobs          map[string]int `json:"active_jobs"`
	CurrentSeason       int            `json:"current_season"`
	CurrentDay          int            `json:"current_day"`
}

// maskPhone hides all but the last 2 digits of a phone number so the public
// leaderboard cannot be used to reconstruct real numbers.
// ponytail: fixed-width mask (9 stars + 2 suffix) — no information about
// prefix length leaks; ceiling: 2-digit suffix means 100 candidates per masked
// value, acceptable for a public gamification dashboard.
func maskPhone(phone string) string {
	if len(phone) <= 2 {
		return "*********"
	}
	return "*********" + phone[len(phone)-2:]
}

func buildTierProgress(value, currentMin int, nextMin int, nextName, nextIcon string, isMax bool) TierProgress {
	progress := TierProgress{
		CurrentMin: currentMin,
		NextMin:    nextMin,
		Value:      value,
		Percent:    100,
		NextName:   nextName,
		NextIcon:   nextIcon,
		IsMax:      isMax,
	}
	if isMax {
		return progress
	}

	progress.Remaining = nextMin - value
	if progress.Remaining < 0 {
		progress.Remaining = 0
	}

	rangeTotal := nextMin - currentMin
	if rangeTotal <= 0 {
		return progress
	}
	percent := ((value - currentMin) * 100) / rangeTotal
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	progress.Percent = percent
	return progress
}

func buildLevelTierProgress(totalPoints int, level domain.Level) TierProgress {
	next, _ := domain.GetNextLevel(totalPoints)
	if next == nil {
		return buildTierProgress(totalPoints, level.MinPoints, level.MinPoints, "", "", true)
	}
	return buildTierProgress(totalPoints, level.MinPoints, next.MinPoints, next.Name, next.Icon, false)
}

func buildSeasonRankProgress(seasonalPoints int, rank domain.Rank) TierProgress {
	next, _ := domain.GetNextSeasonRank(seasonalPoints)
	if next == nil {
		return buildTierProgress(seasonalPoints, rank.MinPoints, rank.MinPoints, "", "", true)
	}
	return buildTierProgress(seasonalPoints, rank.MinPoints, next.MinPoints, next.Name, next.Icon, false)
}

func buildWeekActivity(activityDates []time.Time, weekStart time.Time) ([]bool, int) {
	activeDates := make(map[string]bool, len(activityDates))
	for _, date := range activityDates {
		activeDates[date.Format(time.DateOnly)] = true
	}

	activity := make([]bool, 7)
	activeDays := 0
	for i := range activity {
		date := weekStart.AddDate(0, 0, i)
		active := activeDates[date.Format(time.DateOnly)]
		activity[i] = active
		if active {
			activeDays++
		}
	}
	return activity, activeDays
}

func buildDailyActivity(activityDates []time.Time, today time.Time, days int) ([]DailyActivity, int, int, int) {
	activeDates := make(map[string]bool, len(activityDates))
	for _, date := range activityDates {
		activeDates[date.Format(time.DateOnly)] = true
	}

	start := today.AddDate(0, 0, -(days - 1))
	activity := make([]DailyActivity, 0, days)
	activeDays := 0
	for i := 0; i < days; i++ {
		date := start.AddDate(0, 0, i)
		active := activeDates[date.Format(time.DateOnly)]
		count := 0
		if active {
			count = 1
			activeDays++
		}
		activity = append(activity, DailyActivity{
			Date:   date.Format(time.DateOnly),
			Count:  count,
			Active: active,
		})
	}

	currentStreak := 0
	for date := today; activeDates[date.Format(time.DateOnly)]; date = date.AddDate(0, 0, -1) {
		currentStreak++
	}

	longestStreak := 0
	runningStreak := 0
	var previousActiveDay time.Time
	for _, day := range activityDates {
		day = domain.GetToday(day)
		if !previousActiveDay.IsZero() && day.Equal(previousActiveDay) {
			continue
		}
		if !previousActiveDay.IsZero() && day.Equal(previousActiveDay.AddDate(0, 0, 1)) {
			runningStreak++
		} else {
			runningStreak = 1
		}
		if runningStreak > longestStreak {
			longestStreak = runningStreak
		}
		previousActiveDay = day
	}

	return activity, activeDays, currentStreak, longestStreak
}

var profileDayLabels = []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}

func buildPersonalGoal(goal *domain.WeeklyGoal, activities []domain.GoalActivity) *PersonalGoal {
	activityByDate := make(map[string]string, len(activities))
	for _, activity := range activities {
		activityByDate[activity.Date.Format(time.DateOnly)] = strings.TrimSpace(activity.Activity)
	}

	startDate := domain.GetToday(goal.StartAt.UTC())
	endDate := domain.GetToday(goal.EndAt.Add(-time.Nanosecond).UTC())
	days := make([]GoalDay, 0, 7)
	for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
		activity, active := activityByDate[date.Format(time.DateOnly)]
		if activity == "" {
			activity = "—"
		}
		days = append(days, GoalDay{
			Date:     date.Format(time.DateOnly),
			DayLabel: profileDayLabels[date.Weekday()],
			Activity: activity,
			Active:   active,
		})
	}

	completedDays := len(activityByDate)
	if completedDays > goal.TargetDays {
		completedDays = goal.TargetDays
	}
	remainingDays := goal.TargetDays - completedDays
	if remainingDays < 0 {
		remainingDays = 0
	}
	percent := 0
	if goal.TargetDays > 0 {
		percent = (completedDays * 100) / goal.TargetDays
	}
	if percent > 100 {
		percent = 100
	}

	isCompleted := goal.CompletedAt != nil
	completedAtStr := ""
	if isCompleted {
		completedAtStr = goal.CompletedAt.Format(time.RFC3339)
	}

	return &PersonalGoal{
		TargetDays:    goal.TargetDays,
		Activity:      goal.Activity,
		StartAt:       goal.StartAt.Format(time.RFC3339),
		EndAt:         goal.EndAt.Format(time.RFC3339),
		CompletedDays: completedDays,
		RemainingDays: remainingDays,
		Percent:       percent,
		IsCompleted:   isCompleted,
		CompletedAt:   completedAtStr,
		Days:          days,
	}
}

func totalActiveDays(r *domain.Report) int {
	return r.CenturionCycles*100 + r.ActivityCount
}

func sortEnrichedReportsBySeason(reports []EnrichedReport) {
	sort.Slice(reports, func(i, j int) bool {
		if reports[i].SeasonalPoints == reports[j].SeasonalPoints {
			if reports[i].SeasonalActivityCount == reports[j].SeasonalActivityCount {
				return reports[i].Name < reports[j].Name
			}
			return reports[i].SeasonalActivityCount > reports[j].SeasonalActivityCount
		}
		return reports[i].SeasonalPoints > reports[j].SeasonalPoints
	})
}

func enrichReport(r *domain.Report, today time.Time, weekActivity []bool, weekActiveDays int) EnrichedReport {
	return enrichReportWithMasking(r, today, weekActivity, weekActiveDays, true)
}

func enrichReportWithMasking(r *domain.Report, today time.Time, weekActivity []bool, weekActiveDays int, mask bool) EnrichedReport {
	xpProg := domain.GetNumericLevelProgress(r.TotalPoints)
	lvl := domain.GetLevel(r.TotalPoints)
	rank := domain.GetSeasonRank(r.SeasonalPoints)

	jobName := "No Job"
	jobIcon := "🌱"
	jobDesc := "Belum memilih job class."
	jobTrait := ""
	if job, ok := domain.GetJobClass(r.JobClass); ok {
		jobName = job.Name
		jobIcon = job.Icon
		jobDesc = job.Description
		jobTrait = job.Trait
	}

	var achs []string
	if r.Achievements != "" {
		for _, a := range strings.Split(r.Achievements, ",") {
			a = strings.TrimSpace(a)
			if a != "" {
				achs = append(achs, a)
			}
		}
	}

	var seasonalAchs []string
	if r.SeasonalAchievements != "" {
		for _, a := range strings.Split(r.SeasonalAchievements, ",") {
			a = strings.TrimSpace(a)
			if a != "" {
				seasonalAchs = append(seasonalAchs, a)
			}
		}
	}

	lastReportDay := domain.GetToday(r.LastReportDate)
	isActiveToday := today.Equal(lastReportDay)
	daysSinceLastReport := int(today.Sub(lastReportDay).Hours() / 24)
	if daysSinceLastReport < 0 {
		daysSinceLastReport = 0
	}

	userID := r.UserID
	if mask {
		userID = maskPhone(r.UserID)
	}

	return EnrichedReport{
		UserID:                userID,
		Name:                  r.Name,
		JobClass:              r.JobClass,
		JobName:               jobName,
		JobIcon:               jobIcon,
		JobDescription:        jobDesc,
		JobTrait:              jobTrait,
		Streak:                r.Streak,
		ActivityCount:         r.ActivityCount,
		TotalActiveDays:       totalActiveDays(r),
		LastReportDate:        r.LastReportDate.Format(time.RFC3339),
		MaxStreak:             r.MaxStreak,
		TotalPoints:           r.TotalPoints,
		Level:                 xpProg.Level,
		LevelName:             lvl.Name,
		LevelIcon:             lvl.Icon,
		XPProgress:            xpProg,
		LevelTierProgress:     buildLevelTierProgress(r.TotalPoints, lvl),
		Achievements:          achs,
		ComebackStreak:        r.ComebackStreak,
		InactiveDays:          r.InactiveDays,
		DaysSinceLastReport:   daysSinceLastReport,
		CenturionCycles:       r.CenturionCycles,
		SeasonalPoints:        r.SeasonalPoints,
		SeasonalActivityCount: r.SeasonalActivityCount,
		SeasonalMaxStreak:     r.SeasonalMaxStreak,
		SeasonalAchievements:  seasonalAchs,
		StreakFreezes:         r.StreakFreezes,
		GoalsCompleted:        r.GoalsCompleted,
		TotalSideQuests:       r.TotalSideQuests,
		SeasonalSideQuests:    r.SeasonalSideQuests,
		Str:                   r.Str,
		Sta:                   r.Sta,
		Agi:                   r.Agi,
		Vit:                   r.Vit,
		RankName:              rank.Name,
		RankIcon:              rank.Icon,
		SeasonRankProgress:    buildSeasonRankProgress(r.SeasonalPoints, rank),
		WeekActiveDays:        weekActiveDays,
		WeekActivity:          weekActivity,
		EstimatedWeeklyPoints: weekActiveDays * 10,
		IsActiveToday:         isActiveToday,
	}
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

// AssemblePublicLeaderboard builds the EnrichedReport list for the public web leaderboard (klasemen publik).
// All tab-specific sorting, filtering and pagination is intentionally client-side (React/TS) for low-latency metric switching.
// Backend only guarantees canonical enrichment (week activity dots computed from activity_logs, RPG fields, masking).
//
// Clean Go: this assembler isolates data transformation and N queries from the HTTP handler.
// Note (ponytail): the per-user GetUserActivityDates loop is N+1. Acceptable ceiling for current scale (<100 users);
// upgrade path: pre-load activity by week range or pre-join + group.
func (s *Server) AssemblePublicLeaderboard(ctx context.Context) ([]EnrichedReport, error) {
	reports, err := s.repo.GetAllReports(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	today := domain.GetToday(now)
	// Attribution-aware ISO week (derives from GetToday, matching the
	// activity_date keys produced by logActivity/UpsertReportWithActivity).
	// All entry points (main /lapor, /lapor sidequest via ExecuteSideQuest,
	// and Strava auto-reports) use GetToday for the stored date.
	// Using matching week start keeps week_active_days and week_activity
	// dots consistent with what is actually logged (no cutoff <-> strict drift).
	// Manual user counts and the week tab will now align.
	weekStart := domain.GetStartOfISOWeek(now)

	enriched := make([]EnrichedReport, 0, len(reports))
	for _, rep := range reports {
		dates, err := s.repo.GetUserActivityDates(ctx, rep.UserID)
		if err != nil {
			return nil, err
		}
		weekAct, weekDays := buildWeekActivity(dates, weekStart)
		enriched = append(enriched, enrichReport(rep, today, weekAct, weekDays))
	}

	// Season sort is our canonical order (mirrors WA leaderboard ordering).
	// Client re-sorts for its activeTab.
	sortEnrichedReportsBySeason(enriched)
	return enriched, nil
}

func (s *Server) HandleLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.writeJSON(w, http.StatusNoContent, nil)
		return
	}

	enriched, err := s.AssemblePublicLeaderboard(r.Context())
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	s.writeJSON(w, http.StatusOK, enriched)
}

func (s *Server) HandleSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.writeJSON(w, http.StatusNoContent, nil)
		return
	}

	reports, err := s.repo.GetAllReports(r.Context())
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	totalParticipants := 0
	activeStreakCount := 0
	totalWorkoutsLogged := 0
	activeJobs := make(map[string]int)

	now := time.Now()
	currentWeekStart := domain.GetStartOfISOWeek(now)

	for _, rep := range reports {
		if rep.ActivityCount > 0 || rep.SeasonalActivityCount > 0 {
			totalParticipants++
		}

		lastWeekStart := domain.GetStartOfISOWeek(rep.LastReportDate)
		weeksSinceLastReport := int(math.Round(currentWeekStart.Sub(lastWeekStart).Hours() / (24 * 7)))
		if weeksSinceLastReport <= 1 && rep.Streak > 0 {
			activeStreakCount++
		}

		totalWorkoutsLogged += totalActiveDays(rep)

		jc := rep.JobClass
		if jc == "" {
			jc = "none"
		}
		activeJobs[jc]++
	}

	sessionNumber, sessionStart := usecase.GetCurrentSessionInfo(now)
	challengeDay := int(now.Sub(sessionStart).Hours()/24) + 1

	summary := GlobalSummary{
		TotalParticipants:   totalParticipants,
		ActiveStreakCount:   activeStreakCount,
		TotalWorkoutsLogged: totalWorkoutsLogged,
		ActiveJobs:          activeJobs,
		CurrentSeason:       sessionNumber,
		CurrentDay:          challengeDay,
	}

	s.writeJSON(w, http.StatusOK, summary)
}

func (s *Server) HandleMotivation(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.writeJSON(w, http.StatusNoContent, nil)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"quote": usecase.RandomQuote()})
}

func (s *Server) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.writeJSON(w, http.StatusNoContent, nil)
		return
	}

	rawPhone := r.URL.Query().Get("phone")
	normalized, err := phone.Normalize(rawPhone)
	if err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Nomor telepon tidak valid"})
		return
	}

	report, err := s.repo.GetReport(r.Context(), normalized)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if report == nil {
		s.writeJSON(w, http.StatusNotFound, map[string]string{"error": "User tidak ditemukan"})
		return
	}

	now := time.Now()
	today := domain.GetToday(now)
	weekStart := domain.GetStartOfISOWeekStrict(now)
	activityDates, err := s.repo.GetUserActivityDates(r.Context(), report.UserID)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	weekActivity, weekActiveDays := buildWeekActivity(activityDates, weekStart)
	enriched := enrichReportWithMasking(report, today, weekActivity, weekActiveDays, false)
	dailyActivity, activeDaysInWindow, currentDailyStreak, longestDailyStreak := buildDailyActivity(activityDates, today, 35)
	enriched.DailyActivity = dailyActivity
	enriched.ActiveDaysInWindow = activeDaysInWindow
	enriched.CurrentDailyStreak = currentDailyStreak
	enriched.LongestDailyStreak = longestDailyStreak

	goal, err := s.repo.GetActiveGoal(r.Context(), report.UserID, now)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if goal != nil {
		activities, err := s.repo.GetGoalActivities(r.Context(), report.UserID, goal.StartAt, goal.EndAt)
		if err != nil {
			s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		enriched.ActiveGoal = buildPersonalGoal(goal, activities)
	}

	if strings.TrimSpace(report.JobClass) != "" {
		questUC := usecase.NewDailyQuestUsecase(s.repo)
		tasks, err := questUC.GetOrGenerateQuestList(r.Context(), report.UserID, report.JobClass, report.Level, now)
		if err != nil {
			s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		enriched.TodaySideQuests = tasks
	}
	s.writeJSON(w, http.StatusOK, enriched)
}

func (s *Server) HandleUpdateName(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.writeJSON(w, http.StatusNoContent, nil)
		return
	}

	var body struct {
		Phone string `json:"phone"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Body tidak valid"})
		return
	}

	normalized, err := phone.Normalize(body.Phone)
	if err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Nomor telepon tidak valid"})
		return
	}

	msg, err := usecase.NewUpdateNameUsecase(s.repo).Execute(r.Context(), normalized, body.Name)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": msg})
}

func (s *Server) HandleSelectJob(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.writeJSON(w, http.StatusNoContent, nil)
		return
	}

	var body struct {
		Phone string `json:"phone"`
		JobID string `json:"job_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Body tidak valid"})
		return
	}

	normalized, err := phone.Normalize(body.Phone)
	if err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Nomor telepon tidak valid"})
		return
	}

	msg, err := usecase.NewJobUsecase(s.repo).Select(r.Context(), normalized, "", body.JobID)
	if err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": msg})
}

func (s *Server) HandleSetGoal(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.writeJSON(w, http.StatusNoContent, nil)
		return
	}

	var body struct {
		Phone      string `json:"phone"`
		TargetDays string `json:"target_days"`
		Activity   string `json:"activity"`
		Action     string `json:"action,omitempty"` // "reset"
		// Optional custom start for web dashboard: either ISO start_at or date+hour
		StartAt   string `json:"start_at,omitempty"`
		StartDate string `json:"start_date,omitempty"` // YYYY-MM-DD
		StartHour *int   `json:"start_hour,omitempty"` // 0-23
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Body tidak valid"})
		return
	}

	normalized, err := phone.Normalize(body.Phone)
	if err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Nomor telepon tidak valid"})
		return
	}

	uc := usecase.NewGoalUsecase(s.repo)

	// Reset support (web dashboard)
	if strings.EqualFold(strings.TrimSpace(body.Action), "reset") || strings.EqualFold(strings.TrimSpace(body.TargetDays), "reset") {
		msg, err := uc.Execute(r.Context(), normalized, "", "#goal reset")
		if err != nil {
			s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		s.writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": msg})
		return
	}

	// If explicit start provided (from web), use SetWithStart; else legacy path (WA-style)
	if body.StartAt != "" || body.StartDate != "" || body.StartHour != nil {
		start, err := resolveGoalStart(body.StartAt, body.StartDate, body.StartHour)
		if err != nil {
			s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		td, err := strconv.Atoi(strings.TrimSpace(body.TargetDays))
		if err != nil || td < 1 || td > 7 {
			s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Target goal harus angka 1 sampai 7"})
			return
		}
		act := strings.TrimSpace(body.Activity)
		if act == "" {
			act = "Olahraga"
		}
		msg, err := uc.SetWithStart(r.Context(), normalized, td, act, start)
		if err != nil {
			s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		s.writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": msg})
		return
	}

	// Legacy path (ProfileSetup + WA commands) — keeps using server "now"
	command := "goal set " + body.TargetDays + " " + body.Activity
	msg, err := uc.Execute(r.Context(), normalized, "", command)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": msg})
}

func resolveGoalStart(startAt, startDate string, startHour *int) (time.Time, error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	if startAt != "" {
		t, err := time.Parse(time.RFC3339, startAt)
		if err != nil {
			return time.Time{}, fmt.Errorf("start_at tidak valid (gunakan RFC3339)")
		}
		// Normalize to date at chosen hour if hour present? Keep as provided but ensure WIB semantics
		return t, nil
	}
	if startDate == "" {
		// default to today at 00:00 WIB
		now := time.Now().In(loc)
		y, m, d := now.Date()
		return time.Date(y, m, d, 0, 0, 0, 0, loc), nil
	}
	// parse YYYY-MM-DD
	parts := strings.Split(startDate, "-")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("start_date harus format YYYY-MM-DD")
	}
	y, _ := strconv.Atoi(parts[0])
	mi, _ := strconv.Atoi(parts[1])
	d, _ := strconv.Atoi(parts[2])
	hour := 0
	if startHour != nil {
		if *startHour < 0 || *startHour > 23 {
			return time.Time{}, fmt.Errorf("start_hour harus 0-23")
		}
		hour = *startHour
	}
	return time.Date(y, time.Month(mi), d, hour, 0, 0, 0, loc), nil
}

func (s *Server) HandleListJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.writeJSON(w, http.StatusNoContent, nil)
		return
	}

	jobs, err := s.repo.GetAllJobClasses(r.Context())
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	s.writeJSON(w, http.StatusOK, jobs)
}

func (s *Server) HandleGetUserByPhone(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.writeJSON(w, http.StatusNoContent, nil)
		return
	}

	var body struct {
		Phone string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Body tidak valid"})
		return
	}

	normalized, err := phone.Normalize(body.Phone)
	if err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Nomor telepon tidak valid"})
		return
	}

	report, err := s.repo.GetReport(r.Context(), normalized)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if report == nil {
		s.writeJSON(w, http.StatusNotFound, map[string]string{"error": "User tidak ditemukan"})
		return
	}

	now := time.Now()
	today := domain.GetToday(now)
	weekStart := domain.GetStartOfISOWeekStrict(now)
	activityDates, err := s.repo.GetUserActivityDates(r.Context(), report.UserID)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	weekActivity, weekActiveDays := buildWeekActivity(activityDates, weekStart)
	enriched := enrichReportWithMasking(report, today, weekActivity, weekActiveDays, false)
	dailyActivity, activeDaysInWindow, currentDailyStreak, longestDailyStreak := buildDailyActivity(activityDates, today, 35)
	enriched.DailyActivity = dailyActivity
	enriched.ActiveDaysInWindow = activeDaysInWindow
	enriched.CurrentDailyStreak = currentDailyStreak
	enriched.LongestDailyStreak = longestDailyStreak

	goal, err := s.repo.GetActiveGoal(r.Context(), report.UserID, now)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if goal != nil {
		activities, err := s.repo.GetGoalActivities(r.Context(), report.UserID, goal.StartAt, goal.EndAt)
		if err != nil {
			s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		enriched.ActiveGoal = buildPersonalGoal(goal, activities)
	}

	if strings.TrimSpace(report.JobClass) != "" {
		questUC := usecase.NewDailyQuestUsecase(s.repo)
		tasks, err := questUC.GetOrGenerateQuestList(r.Context(), report.UserID, report.JobClass, report.Level, now)
		if err != nil {
			s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		enriched.TodaySideQuests = tasks
	}
	s.writeJSON(w, http.StatusOK, enriched)
}

func (s *Server) HandleStatic(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	path := filepath.Clean(r.URL.Path)
	distDir := "./frontend/dist"
	fullPath := filepath.Join(distDir, path)

	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		// Fallback to index.html for SPA client-side routing
		http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
		return
	}

	http.ServeFile(w, r, fullPath)
}
