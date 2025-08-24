// internal/api/handlers.go
package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"go_runner/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// healthHandler returns the health status
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now(),
		"services": map[string]string{
			"storage":  "healthy",
			"executor": "healthy",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// listBinariesHandler returns all binaries
func (s *Server) listBinariesHandler(w http.ResponseWriter, r *http.Request) {
	binaries, err := s.storage.ListBinaries()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to list binaries")
		return
	}

	s.respondJSON(w, http.StatusOK, binaries)
}

// createBinaryHandler creates a new binary
func (s *Server) createBinaryHandler(w http.ResponseWriter, r *http.Request) {
	var binary models.Binary
	if err := json.NewDecoder(r.Body).Decode(&binary); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Generate ID
	binary.ID = uuid.New().String()
	binary.Status = "pending"

	// Save binary
	if err := s.storage.SaveBinary(&binary); err != nil {
		slog.Error("Failed to save binary", slog.String("error", err.Error()))
		s.respondError(w, http.StatusInternalServerError, "Failed to save binary")
		return
	}

	s.respondJSON(w, http.StatusCreated, binary)
}

// getBinaryHandler returns a specific binary
func (s *Server) getBinaryHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	binary, err := s.storage.GetBinary(id)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "Binary not found")
		return
	}

	s.respondJSON(w, http.StatusOK, binary)
}

// updateBinaryHandler updates a binary
func (s *Server) updateBinaryHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var binary models.Binary
	if err := json.NewDecoder(r.Body).Decode(&binary); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	binary.ID = id
	if err := s.storage.UpdateBinary(&binary); err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to update binary")
		return
	}

	s.respondJSON(w, http.StatusOK, binary)
}

// deleteBinaryHandler deletes a binary
func (s *Server) deleteBinaryHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := s.storage.DeleteBinary(id); err != nil {
		s.respondError(w, http.StatusNotFound, "Binary not found")
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]string{"message": "Binary deleted successfully"})
}

// buildBinaryHandler triggers a build for a binary
func (s *Server) buildBinaryHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	binary, err := s.storage.GetBinary(id)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "Binary not found")
		return
	}

	// Update status to building
	binary.Status = "building"
	s.storage.UpdateBinary(binary)

	// Build in background
	go func() {
		if err := s.buildBinary(binary); err != nil {
			slog.Error("Failed to build binary",
				slog.String("id", binary.ID),
				slog.String("error", err.Error()))
			binary.Status = "failed"
		} else {
			binary.Status = "ready"
			binary.LastBuilt = time.Now()
		}
		s.storage.UpdateBinary(binary)
	}()

	s.respondJSON(w, http.StatusAccepted, map[string]string{
		"message": "Build started",
		"id":      binary.ID,
	})
}

func (s *Server) buildBinary(binary *models.Binary) error {
	// Clone or update repository
	repoPath := fmt.Sprintf("repo_%s", binary.ID)
	if err := s.git.CloneOrUpdate(binary.RepoURL, binary.Branch, repoPath); err != nil {
		return fmt.Errorf("failed to clone/update repo: %w", err)
	}

	// Get commit hash for version
	commitHash, err := s.git.GetCommitHash(repoPath)
	if err != nil {
		return fmt.Errorf("failed to get commit hash: %w", err)
	}
	binary.Version = commitHash[:8]

	// Build the binary
	outputPath := filepath.Join("./data/binaries", binary.ID)
	if err := s.git.BuildGoBinary(repoPath, binary.BuildPath, outputPath); err != nil {
		return fmt.Errorf("failed to build binary: %w", err)
	}

	binary.BinaryPath = outputPath
	return nil
}

// executeBinaryHandler executes a binary
func (s *Server) executeBinaryHandler(w http.ResponseWriter, r *http.Request) {
	var req models.ExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get binary
	binary, err := s.storage.GetBinary(req.BinaryID)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "Binary not found")
		return
	}

	if binary.Status != "ready" {
		s.respondError(w, http.StatusBadRequest, "Binary is not ready for execution")
		return
	}

	// Execute binary
	result, err := s.executor.Execute(r.Context(), binary.BinaryPath, &req, nil)
	if err != nil {
		slog.Error("Failed to execute binary", slog.String("error", err.Error()))
		s.respondError(w, http.StatusInternalServerError, "Failed to execute binary")
		return
	}

	// Save execution result
	if err := s.storage.SaveExecution(result); err != nil {
		slog.Error("Failed to save execution result", slog.String("error", err.Error()))
	}

	s.respondJSON(w, http.StatusOK, result)
}

// getExecutionHandler returns execution details
func (s *Server) getExecutionHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := s.storage.GetExecution(id)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "Execution not found")
		return
	}

	s.respondJSON(w, http.StatusOK, result)
}

// stopExecutionHandler stops a running execution
func (s *Server) stopExecutionHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := s.executor.StopExecution(id); err != nil {
		s.respondError(w, http.StatusNotFound, "Execution not found or already stopped")
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]string{"message": "Execution stopped"})
}

// Helper functions
func (s *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) respondError(w http.ResponseWriter, status int, message string) {
	s.respondJSON(w, status, map[string]string{"error": message})
}
