// internal/executor/executor.go
package executor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"go_runner/internal/config"
	"go_runner/internal/models"
)

// Executor handles binary execution
type Executor struct {
	binaryPath  string
	config      config.ExecutorConfig
	runningJobs map[string]*exec.Cmd
	mu          sync.RWMutex
}

// NewExecutor creates a new executor
func NewExecutor(binaryPath string, config config.ExecutorConfig) *Executor {
	return &Executor{
		binaryPath:  binaryPath,
		config:      config,
		runningJobs: make(map[string]*exec.Cmd),
	}
}

// Execute runs a binary with the given parameters
func (e *Executor) Execute(ctx context.Context, binaryPath string, req *models.ExecutionRequest) (*models.ExecutionResult, error) {
	// Create execution result
	result := &models.ExecutionResult{
		ID:        generateID(),
		BinaryID:  req.BinaryID,
		Status:    "running",
		StartedAt: time.Now(),
	}

	// Set timeout
	timeout := time.Duration(req.Timeout) * time.Second
	if timeout <= 0 {
		timeout = e.config.Timeout
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(execCtx, binaryPath, req.Args...)

	// Set environment variables
	if len(req.Env) > 0 {
		cmd.Env = append(cmd.Env, req.Env...)
	}

	// Set stdin if provided
	if req.Stdin != "" {
		cmd.Stdin = strings.NewReader(req.Stdin)
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Track running job
	e.mu.Lock()
	e.runningJobs[result.ID] = cmd
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		delete(e.runningJobs, result.ID)
		e.mu.Unlock()
	}()

	// Execute command
	err := cmd.Run()

	result.FinishedAt = time.Now()
	result.Duration = result.FinishedAt.Sub(result.StartedAt).Milliseconds()
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			result.Status = "timeout"
			result.ExitCode = -1
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			result.Status = "failed"
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.Status = "failed"
			result.ExitCode = -1
		}
	} else {
		result.Status = "completed"
		result.ExitCode = 0
	}

	return result, nil
}

// StopExecution stops a running execution
func (e *Executor) StopExecution(executionID string) error {
	e.mu.RLock()
	cmd, exists := e.runningJobs[executionID]
	e.mu.RUnlock()

	if !exists {
		return fmt.Errorf("execution %s not found", executionID)
	}

	if cmd.Process != nil {
		return cmd.Process.Kill()
	}

	return nil
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
