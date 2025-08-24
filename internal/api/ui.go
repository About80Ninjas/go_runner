package api

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed templates/*
var templates embed.FS

//go:embed static/*
var static embed.FS

// adminUIHandler renders admin.html
func (s *Server) adminUIHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFS(templates, "templates/admin.html")
	if err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_ = t.Execute(w, nil)
}

// loginUIHandler renders login.html
func (s *Server) loginUIHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFS(templates, "templates/login.html")
	if err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_ = t.Execute(w, nil)
}
