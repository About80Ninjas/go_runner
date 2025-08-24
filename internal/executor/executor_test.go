
package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go_runner/internal/config"
	"go_runner/internal/models"
)

var testBinPath string

func TestMain(m *testing.M) {
	// Create a dummy binary for testing
	binDir, err := os.MkdirTemp("", "test-bin")
	if err != nil {
		fmt.Println("Error creating temp dir:", err)
		os.Exit(1)
	}
	defer os.RemoveAll(binDir)

	testBinPath = filepath.Join(binDir, "test_binary")
	cmd := exec.Command("go", "build", "-o", testBinPath, "./testdata/main.go")
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error building test binary:", err)
		os.Exit(1)
	}

	// Run tests
	exitCode := m.Run()

	os.Exit(exitCode)
}

func TestExecutor_Execute_Success(t *testing.T) {
	t.Parallel()
	executor := NewExecutor(testBinPath, config.ExecutorConfig{Timeout: 5 * time.Second})
	req := &models.ExecutionRequest{
		BinaryID: "test-binary",
		Args:     []string{"hello"},
	}

	result, err := executor.Execute(context.Background(), testBinPath, req, nil)
	assert.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, "hello", result.Stdout)
}

func TestExecutor_Execute_Timeout(t *testing.T) {
	t.Parallel()
	executor := NewExecutor(testBinPath, config.ExecutorConfig{Timeout: 1 * time.Second})
	req := &models.ExecutionRequest{
		BinaryID: "test-binary",
		Args:     []string{"sleep"},
		Timeout:  1,
	}

	result, err := executor.Execute(context.Background(), testBinPath, req, nil)
	assert.NoError(t, err)
	assert.Equal(t, "timeout", result.Status)
	assert.Equal(t, -1, result.ExitCode)
}

func TestExecutor_StopExecution(t *testing.T) {
	t.Parallel()
	executor := NewExecutor(testBinPath, config.ExecutorConfig{Timeout: 5 * time.Second})
	req := &models.ExecutionRequest{
		BinaryID: "test-binary",
		Args:     []string{"sleep"},
	}

	var result *models.ExecutionResult
	var err error
	done := make(chan struct{})
	started := make(chan string)

	go func() {
		result, err = executor.Execute(context.Background(), testBinPath, req, started)
		close(done)
	}()

	// Wait for the execution to start and get the ID
	var executionID string
	select {
	case executionID = <-started:
		time.Sleep(100 * time.Millisecond)
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for execution to start")
	}

	// Stop the execution
	err = executor.StopExecution(executionID)
	assert.NoError(t, err)

	<-done

	assert.Equal(t, "failed", result.Status)
}
