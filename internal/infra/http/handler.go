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
	"strings"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/app/usecase"
	"github.com/fardannozami/whatsapp-gateway/internal/config"
	"github.com/fardannozami/whatsapp-gateway/internal/domain"
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
	LastReportDate        string                      `json:"last_report_date"`
	MaxStreak             int                         `json:"max_streak"`
	TotalPoints           int                         `json:"total_points"`
	Level                 int                         `json:"level"`
	LevelName             string                      `json:"level_name"`
	LevelIcon             string                      `json:"level_icon"`
	XPProgress            domain.NumericLevelProgress `json:"xp_progress"`
	Achievements          []string                    `json:"achievements"`
	ComebackStreak        int                         `json:"comeback_streak"`
	InactiveDays          int                         `json:"inactive_days"`
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
	IsActiveToday         bool                        `json:"is_active_today"`
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

func maskPhone(phone string) string {
	if len(phone) < 7 {
		return "****"
	}
	return phone[:5] + "****" + phone[len(phone)-3:]
}

func enrichReport(r *domain.Report) EnrichedReport {
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

	now := time.Now()
	today := domain.GetToday(now)
	lastReportDay := domain.GetToday(r.LastReportDate)
	isActiveToday := today.Equal(lastReportDay)

	return EnrichedReport{
		UserID:                maskPhone(r.UserID),
		Name:                  r.Name,
		JobClass:              r.JobClass,
		JobName:               jobName,
		JobIcon:               jobIcon,
		JobDescription:        jobDesc,
		JobTrait:              jobTrait,
		Streak:                r.Streak,
		ActivityCount:         r.ActivityCount,
		LastReportDate:        r.LastReportDate.Format(time.RFC3339),
		MaxStreak:             r.MaxStreak,
		TotalPoints:           r.TotalPoints,
		Level:                 xpProg.Level,
		LevelName:             lvl.Name,
		LevelIcon:             lvl.Icon,
		XPProgress:            xpProg,
		Achievements:          achs,
		ComebackStreak:        r.ComebackStreak,
		InactiveDays:          r.InactiveDays,
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

func (s *Server) HandleLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.writeJSON(w, http.StatusNoContent, nil)
		return
	}

	reports, err := s.repo.GetAllReports(r.Context())
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	var enriched []EnrichedReport
	for _, rep := range reports {
		enriched = append(enriched, enrichReport(rep))
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

		totalWorkoutsLogged += rep.CenturionCycles*100 + rep.ActivityCount

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

