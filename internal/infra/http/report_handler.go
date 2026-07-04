package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type CreateReportRequest struct {
	Message string `json:"message"`
}

type CreateReportResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type CreateSidequestRequest struct {
	Activities []string `json:"activities"`
}

type CreateSidequestResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func (s *Server) HandleCreateReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	phone, ok := UserIDFromContext(r.Context())
	if !ok {
		s.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	var req CreateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Message == "" {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Message is required"})
		return
	}

	report, err := s.repo.GetReport(r.Context(), phone)
	if err != nil || report == nil {
		s.writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found. Please register via WhatsApp first."})
		return
	}

	fullMessage := "/lapor " + req.Message
	workout := domain.ParseHevy(req.Message)

	responseText, err := s.reportUC.ExecuteWithMessage(
		r.Context(),
		phone,
		report.Name,
		fullMessage,
		workout,
	)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	s.writeJSON(w, http.StatusOK, CreateReportResponse{
		Message: responseText,
		Success: true,
	})
}

func (s *Server) HandleCreateSidequestReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	phone, ok := UserIDFromContext(r.Context())
	if !ok {
		s.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	var req CreateSidequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if len(req.Activities) == 0 {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "At least one activity is required"})
		return
	}

	report, err := s.repo.GetReport(r.Context(), phone)
	if err != nil || report == nil {
		s.writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found. Please register via WhatsApp first."})
		return
	}

	responseText, err := s.dailyQuestUC.UpdateProgress(
		r.Context(),
		phone,
		report.Name,
		req.Activities,
		s.reportUC,
		time.Now(),
	)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	s.writeJSON(w, http.StatusOK, CreateSidequestResponse{
		Message: responseText,
		Success: true,
	})
}
