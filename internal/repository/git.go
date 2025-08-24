// internal/repository/git.go
package repository

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitManager handles Git repository operations
type GitManager struct {
	basePath string
}

// NewGitManager creates a new Git manager
func NewGitManager(basePath string) *GitManager {
	return &GitManager{
		basePath: basePath,
	}
}

// CloneOrUpdate clones a repository or updates it if it exists
func (gm *GitManager) CloneOrUpdate(repoURL, branch, targetPath string) error {
	fullPath := filepath.Join(gm.basePath, targetPath)

	// Check if repo exists
	if _, err := os.Stat(filepath.Join(fullPath, ".git")); err == nil {
		// Repository exists, update it
		return gm.updateRepo(fullPath, branch)
	}

	// Clone the repository
	return gm.cloneRepo(repoURL, branch, fullPath)
}

func (gm *GitManager) cloneRepo(repoURL, branch, targetPath string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	cmd := exec.Command("git", "clone", "-b", branch, repoURL, targetPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (gm *GitManager) updateRepo(repoPath, branch string) error {
	// Fetch latest changes
	cmd := exec.Command("git", "fetch", "origin")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git fetch failed: %w\nOutput: %s", err, string(output))
	}

	// Checkout branch
	cmd = exec.Command("git", "checkout", branch)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout failed: %w\nOutput: %s", err, string(output))
	}

	// Pull latest changes
	cmd = exec.Command("git", "pull", "origin", branch)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git pull failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// GetCommitHash returns the current commit hash
func (gm *GitManager) GetCommitHash(repoPath string) (string, error) {
	fullPath := filepath.Join(gm.basePath, repoPath)

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = fullPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// BuildGoBinary builds a Go binary from the repository
func (gm *GitManager) BuildGoBinary(repoPath, buildPath, outputPath string) error {
	fullRepoPath := filepath.Join(gm.basePath, repoPath)
	fullBuildPath := filepath.Join(fullRepoPath, buildPath)

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build the binary
	cmd := exec.Command("go", "build", "-o", outputPath, ".")
	cmd.Dir = fullBuildPath
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go build failed: %w\nOutput: %s", err, string(output))
	}

	// Make binary executable
	if err := os.Chmod(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}

	return nil
}
