package api

import (
	"net/http"
	"time"
)

// logoutHandler clears the admin cookie and redirects to login
func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Overwrite the cookie with an expired one
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // set true if running HTTPS
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})

	// If the request came from the UI form, redirect user
	if r.Header.Get("Accept") == "" || r.Header.Get("Accept") == "text/html" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// For API clients, just send JSON
	s.respondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "logged out",
	})
}
