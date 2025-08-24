// internal/storage/storage.go
package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go_runner/internal/models"
)

var (
	ErrBinaryNotFound  = errors.New("binary not found")
	ErrStorageInit     = errors.New("storage initialization failed")
	ErrInvalidID       = errors.New("invalid ID supplied")
)

// isValidID checks that the id does not contain path separators or directory traversal
func isValidID(id string) bool {
	if len(id) == 0 ||
		strings.Contains(id, "/") ||
		strings.Contains(id, "\\") ||
		strings.Contains(id, "..") {
		return false
	}
	return true
}

// Storage interface for binary management
type Storage interface {
	Init() error
	SaveBinary(binary *models.Binary) error
	GetBinary(id string) (*models.Binary, error)
	ListBinaries() ([]*models.Binary, error)
	UpdateBinary(binary *models.Binary) error
	DeleteBinary(id string) error
	SaveExecution(result *models.ExecutionResult) error
	GetExecution(id string) (*models.ExecutionResult, error)
}

// FileStorage implements Storage using filesystem
type FileStorage struct {
	basePath   string
	mu         sync.RWMutex
	binaries   map[string]*models.Binary
	executions map[string]*models.ExecutionResult
}

// NewFileStorage creates a new file-based storage
func NewFileStorage(basePath string) *FileStorage {
	return &FileStorage{
		basePath:   basePath,
		binaries:   make(map[string]*models.Binary),
		executions: make(map[string]*models.ExecutionResult),
	}
}

// Init initializes the storage
func (fs *FileStorage) Init() error {
	// Create necessary directories
	dirs := []string{
		fs.basePath,
		filepath.Join(fs.basePath, "binaries"),
		filepath.Join(fs.basePath, "executions"),
		filepath.Join(fs.basePath, "metadata"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("%w: %v", ErrStorageInit, err)
		}
	}

	// Load existing metadata
	return fs.loadMetadata()
}

func (fs *FileStorage) loadMetadata() error {
	metadataPath := filepath.Join(fs.basePath, "metadata", "binaries.json")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No existing data
		}
		return err
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	return json.Unmarshal(data, &fs.binaries)
}

func (fs *FileStorage) saveMetadata() error {
	fs.mu.RLock()
	data, err := json.MarshalIndent(fs.binaries, "", "  ")
	fs.mu.RUnlock()

	if err != nil {
		return err
	}

	metadataPath := filepath.Join(fs.basePath, "metadata", "binaries.json")
	return os.WriteFile(metadataPath, data, 0644)
}

// SaveBinary saves a binary to storage
func (fs *FileStorage) SaveBinary(binary *models.Binary) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	binary.CreatedAt = time.Now()
	binary.UpdatedAt = time.Now()
	fs.binaries[binary.ID] = binary

	return fs.saveMetadata()
}

// GetBinary retrieves a binary by ID
func (fs *FileStorage) GetBinary(id string) (*models.Binary, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	binary, exists := fs.binaries[id]
	if !exists {
		return nil, ErrBinaryNotFound
	}

	return binary, nil
}

// ListBinaries returns all binaries
func (fs *FileStorage) ListBinaries() ([]*models.Binary, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	binaries := make([]*models.Binary, 0, len(fs.binaries))
	for _, binary := range fs.binaries {
		binaries = append(binaries, binary)
	}

	return binaries, nil
}

// UpdateBinary updates a binary
func (fs *FileStorage) UpdateBinary(binary *models.Binary) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if _, exists := fs.binaries[binary.ID]; !exists {
		return ErrBinaryNotFound
	}

	binary.UpdatedAt = time.Now()
	fs.binaries[binary.ID] = binary

	return fs.saveMetadata()
}

	if !isValidID(id) {
		return ErrInvalidID
	}
// DeleteBinary deletes a binary
func (fs *FileStorage) DeleteBinary(id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if _, exists := fs.binaries[id]; !exists {
		return ErrBinaryNotFound
	}

	delete(fs.binaries, id)

	// Remove binary file
	binaryPath := filepath.Join(fs.basePath, "binaries", id)
	os.Remove(binaryPath)

	return fs.saveMetadata()
}

	if !isValidID(result.ID) {
		return ErrInvalidID
	}
// SaveExecution saves an execution result
func (fs *FileStorage) SaveExecution(result *models.ExecutionResult) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.executions[result.ID] = result

	// Also save to file for persistence
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	execPath := filepath.Join(fs.basePath, "executions", result.ID+".json")
	return os.WriteFile(execPath, data, 0644)
}

	if !isValidID(id) {
		return nil, ErrInvalidID
	}
// GetExecution retrieves an execution result
func (fs *FileStorage) GetExecution(id string) (*models.ExecutionResult, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	if result, exists := fs.executions[id]; exists {
		return result, nil
	}

	// Try loading from file
	execPath := filepath.Join(fs.basePath, "executions", id+".json")
	data, err := os.ReadFile(execPath)
	if err != nil {
		return nil, err
	}

	var result models.ExecutionResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
