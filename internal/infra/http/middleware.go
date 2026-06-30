package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey struct{}

var userIDContextKey = contextKey{}

// AuthMiddleware validates the Authorization Bearer token and injects the
// authenticated user's phone number into the request context.
func (s *Server) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			s.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Authorization header required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			s.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid authorization format"})
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(s.jwtSecret), nil
		})
		if err != nil || !token.Valid {
			s.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid or expired token"})
			return
		}

		if claims.Phone == "" {
			s.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid token claims"})
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, claims.Phone)
		next(w, r.WithContext(ctx))
	}
}

// UserIDFromContext extracts the authenticated user's phone number from context.
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDContextKey).(string)
	return id, ok
}
