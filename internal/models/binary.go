// internal/models/binary.go
package models

import (
	"time"
)

// Binary represents a managed Go binary
type Binary struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name" validate:"required,min=1,max=255"`
	Description string    `json:"description" db:"description" validate:"max=1000"`
	RepoURL     string    `json:"repo_url" db:"repo_url" validate:"required,url"`
	Branch      string    `json:"branch" db:"branch" validate:"required"`
	BuildPath   string    `json:"build_path" db:"build_path"` // Path within repo to build
	BinaryPath  string    `json:"binary_path" db:"binary_path"`
	Version     string    `json:"version" db:"version"`
	Status      string    `json:"status" db:"status"` // pending, building, ready, failed
	LastBuilt   time.Time `json:"last_built" db:"last_built"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ExecutionRequest represents a request to execute a binary
type ExecutionRequest struct {
	BinaryID string   `json:"binary_id" validate:"required"`
	Args     []string `json:"args"`
	Env      []string `json:"env"`
	Stdin    string   `json:"stdin"`
	Timeout  int      `json:"timeout"` // seconds
}

// ExecutionResult represents the result of executing a binary
type ExecutionResult struct {
	ID         string    `json:"id"`
	BinaryID   string    `json:"binary_id"`
	Status     string    `json:"status"` // running, completed, failed, timeout
	ExitCode   int       `json:"exit_code"`
	Stdout     string    `json:"stdout"`
	Stderr     string    `json:"stderr"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	Duration   int64     `json:"duration_ms"`
}
