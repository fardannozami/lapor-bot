package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/fardannozami/whatsapp-gateway/internal/app/usecase"
	"github.com/fardannozami/whatsapp-gateway/internal/config"
	"go.mau.fi/whatsmeow"
)

type Server struct {
	linkUC     *usecase.LinkStravaUsecase
	processUC  *usecase.ProcessStravaWebhookUsecase
	waClient   *whatsmeow.Client
	verifyToken string
}

func NewServer(linkUC *usecase.LinkStravaUsecase, processUC *usecase.ProcessStravaWebhookUsecase, waClient *whatsmeow.Client, cfg config.Config) *Server {
	return &Server{
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
