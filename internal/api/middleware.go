// internal/api/middleware.go
package api

import (
	"net/http"
	"os"
	"strings"
)

// authMiddleware validates admin token
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			token = r.URL.Query().Get("token")
		}

		expectedToken := os.Getenv("ADMIN_TOKEN")
		if !strings.HasSuffix(token, expectedToken) {
			s.respondError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// apiKeyMiddleware validates API keys for execution
func (s *Server) apiKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			s.respondError(w, http.StatusUnauthorized, "API key required")
			return
		}

		// In production, validate against stored API keys
		// For now, we'll accept any non-empty key
		if apiKey == "" {
			s.respondError(w, http.StatusUnauthorized, "Invalid API key")
			return
		}

		next.ServeHTTP(w, r)
	})
}
