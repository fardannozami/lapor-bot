package http

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain/phone"
	"github.com/golang-jwt/jwt/v5"
)

// Claims extends JWT registered claims with the user's phone number.
type Claims struct {
	jwt.RegisteredClaims
	Phone string `json:"sub"`
}

// generateToken creates a signed HS256 JWT for the given phone number.
func (s *Server) generateToken(phone string) (string, time.Time, error) {
	expiry := time.Duration(s.jwtExpiryHours) * time.Hour
	expiresAt := time.Now().Add(expiry)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Phone: phone,
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// HandleLogin authenticates a user by phone number and returns a JWT token.
func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.writeJSON(w, http.StatusNoContent, nil)
		return
	}
	if r.Method != http.MethodPost {
		s.writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
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
		log.Printf("login DB error: %v", err)
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Terjadi kesalahan"})
		return
	}
	if report == nil {
		s.writeJSON(w, http.StatusNotFound, map[string]string{"error": "User tidak ditemukan"})
		return
	}

	tokenString, expiresAt, err := s.generateToken(normalized)
	if err != nil {
		log.Printf("login token error: %v", err)
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Gagal membuat token"})
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]any{
		"token":      tokenString,
		"expires_at": expiresAt.Format(time.RFC3339),
		"user": map[string]string{
			"phone": normalized,
			"name":  report.Name,
		},
	})
}
