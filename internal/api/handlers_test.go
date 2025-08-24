package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go_runner/internal/config"
	"go_runner/internal/models"
	"go_runner/internal/storage"
)

// MockStorage is a mock implementation of the Storage interface
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Init() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStorage) SaveBinary(binary *models.Binary) error {
	args := m.Called(binary)
	return args.Error(0)
}

func (m *MockStorage) GetBinary(id string) (*models.Binary, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Binary), args.Error(1)
}

func (m *MockStorage) ListBinaries() ([]*models.Binary, error) {
	args := m.Called()
	return args.Get(0).([]*models.Binary), args.Error(1)
}

func (m *MockStorage) UpdateBinary(binary *models.Binary) error {
	args := m.Called(binary)
	return args.Error(0)
}

func (m *MockStorage) DeleteBinary(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockStorage) SaveExecution(result *models.ExecutionResult) error {
	args := m.Called(result)
	return args.Error(0)
}

func (m *MockStorage) GetExecution(id string) (*models.ExecutionResult, error) {
	args := m.Called(id)
	return args.Get(0).(*models.ExecutionResult), args.Error(1)
}

// MockGitManager is a mock implementation of the GitManager interface
type MockGitManager struct {
	mock.Mock
}

func (m *MockGitManager) CloneOrUpdate(repoURL, branch, targetPath string) error {
	args := m.Called(repoURL, branch, targetPath)
	return args.Error(0)
}

func (m *MockGitManager) GetCommitHash(repoPath string) (string, error) {
	args := m.Called(repoPath)
	return args.String(0), args.Error(1)
}

func (m *MockGitManager) BuildGoBinary(repoPath, buildPath, outputPath string) error {
	args := m.Called(repoPath, buildPath, outputPath)
	return args.Error(0)
}

// MockExecutor is a mock implementation of the Executor interface
type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) Execute(ctx context.Context, binaryPath string, req *models.ExecutionRequest, started chan<- string) (*models.ExecutionResult, error) {
	args := m.Called(ctx, binaryPath, req, started)
	return args.Get(0).(*models.ExecutionResult), args.Error(1)
}

func (m *MockExecutor) StopExecution(executionID string) error {
	args := m.Called(executionID)
	return args.Error(0)
}

func TestHealthHandler(t *testing.T) {
	server := NewServer(config.ServerConfig{}, nil, nil, nil)
	req, err := http.NewRequest("GET", "/api/v1/health", nil)
	assert.NoError(t, err)
	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

// getAdminCookie simulates a login and returns the admin_token cookie
func getAdminCookie(t *testing.T, server *Server) *http.Cookie {
	// Set a temporary ADMIN_TOKEN for the test
	t.Setenv("ADMIN_TOKEN", "test-admin-token")

	// Simulate login request
	form := bytes.NewBufferString("token=test-admin-token")
	req, _ := http.NewRequest("POST", "/login", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusSeeOther, rr.Code) // Should redirect after successful login

	// Find the admin_token cookie
	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == "admin_token" {
			return cookie
		}
	}
	t.Fatal("admin_token cookie not found after login")
	return nil
}

func TestListBinariesHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	server := NewServer(config.ServerConfig{}, mockStorage, nil, nil)

	// Get admin cookie
	adminCookie := getAdminCookie(t, server)

	binaries := []*models.Binary{{ID: "1", Name: "test"}}
	mockStorage.On("ListBinaries").Return(binaries, nil)

	req, _ := http.NewRequest("GET", "/api/v1/binaries", nil)
	req.AddCookie(adminCookie) // Add the admin cookie to the request
	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "test")
	mockStorage.AssertExpectations(t)
}

func TestCreateBinaryHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	server := NewServer(config.ServerConfig{}, mockStorage, nil, nil)

	adminCookie := getAdminCookie(t, server)

	binary := &models.Binary{Name: "test"}
	mockStorage.On("SaveBinary", mock.AnythingOfType("*models.Binary")).Return(nil)

	body, _ := json.Marshal(binary)
	req, _ := http.NewRequest("POST", "/api/v1/binaries", bytes.NewBuffer(body))
	req.AddCookie(adminCookie)
	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Contains(t, rr.Body.String(), "test")
	mockStorage.AssertExpectations(t)
}

func TestGetBinaryHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	server := NewServer(config.ServerConfig{}, mockStorage, nil, nil)

	adminCookie := getAdminCookie(t, server)

	binary := &models.Binary{ID: "1", Name: "test"}
	mockStorage.On("GetBinary", "1").Return(binary, nil)

	req, _ := http.NewRequest("GET", "/api/v1/binaries/1", nil)
	req.AddCookie(adminCookie)
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "test")
	mockStorage.AssertExpectations(t)
}

func TestGetBinaryHandler_NotFound(t *testing.T) {
	mockStorage := new(MockStorage)
	server := NewServer(config.ServerConfig{}, mockStorage, nil, nil)

	adminCookie := getAdminCookie(t, server)

	mockStorage.On("GetBinary", "nonexistent").Return(&models.Binary{}, storage.ErrBinaryNotFound)

	req, _ := http.NewRequest("GET", "/api/v1/binaries/nonexistent", nil)
	req.AddCookie(adminCookie)
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "Binary not found")
	mockStorage.AssertExpectations(t)
}

func TestUpdateBinaryHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	server := NewServer(config.ServerConfig{}, mockStorage, nil, nil)

	adminCookie := getAdminCookie(t, server)

	binary := &models.Binary{ID: "1", Name: "updated"}
	mockStorage.On("UpdateBinary", mock.AnythingOfType("*models.Binary")).Return(nil)

	body, _ := json.Marshal(binary)
	req, _ := http.NewRequest("PUT", "/api/v1/binaries/1", bytes.NewBuffer(body))
	req.AddCookie(adminCookie)
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockStorage.AssertExpectations(t)
}

func TestDeleteBinaryHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	server := NewServer(config.ServerConfig{}, mockStorage, nil, nil)

	adminCookie := getAdminCookie(t, server)

	mockStorage.On("DeleteBinary", "1").Return(nil)

	req, _ := http.NewRequest("DELETE", "/api/v1/binaries/1", nil)
	req.AddCookie(adminCookie)
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockStorage.AssertExpectations(t)
}

func TestBuildBinaryHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	mockGit := new(MockGitManager)
	server := NewServer(config.ServerConfig{}, mockStorage, mockGit, nil)

	adminCookie := getAdminCookie(t, server)

	binary := &models.Binary{
		ID:        "1",
		RepoURL:   "http://example.com/repo.git",
		Branch:    "main",
		BuildPath: ".",
		Status:    "pending",
	}
	mockStorage.On("GetBinary", "1").Return(binary, nil).Once()
	mockStorage.On("UpdateBinary", mock.AnythingOfType("*models.Binary")).Return(nil).Twice()
	mockGit.On("CloneOrUpdate", binary.RepoURL, binary.Branch, mock.AnythingOfType("string")).Return(nil).Once()
	mockGit.On("GetCommitHash", mock.AnythingOfType("string")).Return("abcdef123456", nil).Once()
	mockGit.On("BuildGoBinary", mock.AnythingOfType("string"), binary.BuildPath, mock.AnythingOfType("string")).Return(nil).Once()

	req, _ := http.NewRequest("POST", "/api/v1/binaries/1/build", nil)
	req.AddCookie(adminCookie)
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	// Give the goroutine a moment to execute
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, http.StatusAccepted, rr.Code)
	mockStorage.AssertExpectations(t)
	mockGit.AssertExpectations(t)
}

func TestExecuteBinaryHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	mockExecutor := new(MockExecutor)
	server := NewServer(config.ServerConfig{}, mockStorage, nil, mockExecutor)

	binary := &models.Binary{
		ID:         "1",
		BinaryPath: "/path/to/binary",
		Status:     "ready",
	}
	executionReq := &models.ExecutionRequest{BinaryID: "1", Args: []string{"arg1"}}
	executionResult := &models.ExecutionResult{ID: "exec1", Status: "completed"}

	mockStorage.On("GetBinary", "1").Return(binary, nil).Once()
	mockExecutor.On("Execute", mock.Anything, binary.BinaryPath, executionReq, mock.Anything).Return(executionResult, nil).Once()
	mockStorage.On("SaveExecution", executionResult).Return(nil).Once()

	body, _ := json.Marshal(executionReq)
	req, _ := http.NewRequest("POST", "/api/v1/execute", bytes.NewBuffer(body))
	req.Header.Set("X-API-Key", "test-api-key") // Add API key
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "completed")
	mockStorage.AssertExpectations(t)
	mockExecutor.AssertExpectations(t)
}

func TestGetExecutionHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	server := NewServer(config.ServerConfig{}, mockStorage, nil, nil)

	executionResult := &models.ExecutionResult{ID: "exec1", Status: "completed"}
	mockStorage.On("GetExecution", "exec1").Return(executionResult, nil).Once()

	req, _ := http.NewRequest("GET", "/api/v1/execute/exec1", nil)
	req.Header.Set("X-API-Key", "test-api-key") // Add API key
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "completed")
	mockStorage.AssertExpectations(t)
}

func TestStopExecutionHandler(t *testing.T) {
	mockExecutor := new(MockExecutor)
	server := NewServer(config.ServerConfig{}, nil, nil, mockExecutor)

	mockExecutor.On("StopExecution", "exec1").Return(nil).Once()

	req, _ := http.NewRequest("DELETE", "/api/v1/execute/exec1", nil)
	req.Header.Set("X-API-Key", "test-api-key") // Add API key
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Execution stopped")
	mockExecutor.AssertExpectations(t)
}