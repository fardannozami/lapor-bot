package http

import (
	"net/http"
	"time"
)

type stravaStatusResponse struct {
	Linked    bool   `json:"linked"`
	AthleteID int64  `json:"athlete_id,omitempty"`
	Name      string `json:"name,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

func (s *Server) HandleGetStravaStatus(w http.ResponseWriter, r *http.Request) {
	normalized, ok := UserIDFromContext(r.Context())
	if !ok {
		s.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	account, err := s.repo.GetStravaAccountByUserID(r.Context(), normalized)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if account == nil {
		s.writeJSON(w, http.StatusOK, stravaStatusResponse{Linked: false})
		return
	}

	s.writeJSON(w, http.StatusOK, stravaStatusResponse{
		Linked:    true,
		AthleteID: account.AthleteID,
		Name:      account.Name,
		ExpiresAt: account.ExpiresAt.Format(time.RFC3339),
	})
}

func (s *Server) HandleCreateStravaLink(w http.ResponseWriter, r *http.Request) {
	normalized, ok := UserIDFromContext(r.Context())
	if !ok {
		s.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
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

	s.writeJSON(w, http.StatusOK, map[string]string{
		"auth_url": s.linkUC.GetAuthURL(normalized, report.Name),
	})
}

func (s *Server) HandleDisconnectStrava(w http.ResponseWriter, r *http.Request) {
	normalized, ok := UserIDFromContext(r.Context())
	if !ok {
		s.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}
	if s.disconnectUC == nil {
		s.writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "Strava disconnect belum tersedia"})
		return
	}

	message, err := s.disconnectUC.Execute(r.Context(), normalized)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": message})
}
