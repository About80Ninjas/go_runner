// internal/api/login.go
package api

import (
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// loginPageHandler renders the login page
func (s *Server) loginPageHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFS(templates, "templates/login.html")
	if err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_ = t.Execute(w, map[string]any{
		"Next": r.URL.Query().Get("next"),
	})
}

// loginHandler validates the admin token and sets a cookie
func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	token := r.FormValue("token")
	expected := os.Getenv("ADMIN_TOKEN")
	if expected == "" || token != expected {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	// ✅ set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // set true if using HTTPS
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	// ✅ safe redirect
	next := r.FormValue("next")
	// Define list of allowed redirect paths
	allowedNext := map[string]struct{}{
		"/admin": {},
		"/dashboard": {},
	}
	next = strings.ReplaceAll(next, "\\", "/")
	u, err := url.Parse(next)
	if err != nil || u.Hostname() != "" || u.Path == "" {
		// fallback if next is malformed or external
		next = "/admin"
	} else if _, ok := allowedNext[u.Path]; !ok {
		next = "/admin"
	} else {
		// use normalized and path-only form
		next = u.Path
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}
