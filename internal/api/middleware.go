// internal/api/middleware.go
package api

import (
	"net/http"
	"os"
	"strings"
)

// authMiddleware secures JSON API endpoints under /api/v1 that expect JSON error responses.
// Accepts any of:
//   - Authorization: Bearer <ADMIN_TOKEN>
//   - ?token=<ADMIN_TOKEN>
//   - Cookie: admin_token=<ADMIN_TOKEN>
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isAuthenticated(r) {
			next.ServeHTTP(w, r)
			return
		}
		s.respondError(w, http.StatusUnauthorized, "Unauthorized")
	})
}

// requireAdminUI secures the HTML admin UI. If unauthenticated, it redirects to /login.
func (s *Server) requireAdminUI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isAuthenticated(r) {
			next.ServeHTTP(w, r)
			return
		}
		// 303 = See Other (safer redirect after POST/other verbs)
		http.Redirect(w, r, "/login?next="+r.URL.Path, http.StatusSeeOther)
	})
}

// isAuthenticated checks Authorization header, query token, or admin_token cookie.
func isAuthenticated(r *http.Request) bool {
	expectedToken := os.Getenv("ADMIN_TOKEN")
	if expectedToken == "" {
		// For safety, if unset treat as not authenticated.
		return false
	}

	// 1) Authorization: Bearer <token>
	if h := r.Header.Get("Authorization"); h != "" {
		parts := strings.SplitN(h, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && parts[1] == expectedToken {
			return true
		}
	}

	// 2) ?token=<token>
	if q := r.URL.Query().Get("token"); q == expectedToken {
		return true
	}

	// 3) Cookie: admin_token=<token>
	if c, err := r.Cookie("admin_token"); err == nil && c.Value == expectedToken {
		return true
	}

	return false
}

// apiKeyMiddleware validates API keys for execution endpoints (unchanged)
func (s *Server) apiKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			s.respondError(w, http.StatusUnauthorized, "API key required")
			return
		}
		// In production you'd validate the key properly.
		next.ServeHTTP(w, r)
	})
}
