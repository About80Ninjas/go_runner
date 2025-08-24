// internal/config/config.go
package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for our application
type Config struct {
	Server   ServerConfig
	Storage  StorageConfig
	Executor ExecutorConfig
	Auth     AuthConfig
}

type ServerConfig struct {
	Port         int           `json:"port"`
	Host         string        `json:"host"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
}

type StorageConfig struct {
	Path       string `json:"path"`
	RepoPath   string `json:"repo_path"`
	BinaryPath string `json:"binary_path"`
}

type ExecutorConfig struct {
	MaxConcurrent int           `json:"max_concurrent"`
	Timeout       time.Duration `json:"timeout"`
	MaxMemoryMB   int           `json:"max_memory_mb"`
}

type AuthConfig struct {
	AdminToken string `json:"admin_token"`
	APIKeys    bool   `json:"api_keys_enabled"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{}

	// Server configuration
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	} else {
		config.Server.Port = 8080
	}

	config.Server.Host = getEnvOrDefault("SERVER_HOST", "localhost")
	config.Server.ReadTimeout = getDurationOrDefault("SERVER_READ_TIMEOUT", 10*time.Second)
	config.Server.WriteTimeout = getDurationOrDefault("SERVER_WRITE_TIMEOUT", 10*time.Second)

	// Storage configuration
	config.Storage.Path = getEnvOrDefault("STORAGE_PATH", "./data")
	config.Storage.RepoPath = getEnvOrDefault("REPO_PATH", "./data/repos")
	config.Storage.BinaryPath = getEnvOrDefault("BINARY_PATH", "./data/binaries")

	// Executor configuration
	config.Executor.MaxConcurrent = getIntOrDefault("EXECUTOR_MAX_CONCURRENT", 10)
	config.Executor.Timeout = getDurationOrDefault("EXECUTOR_TIMEOUT", 5*time.Minute)
	config.Executor.MaxMemoryMB = getIntOrDefault("EXECUTOR_MAX_MEMORY_MB", 512)

	// Auth configuration
	config.Auth.AdminToken = os.Getenv("ADMIN_TOKEN")
	if config.Auth.AdminToken == "" {
		return nil, errors.New("ADMIN_TOKEN is required")
	}
	config.Auth.APIKeys = getBoolOrDefault("API_KEYS_ENABLED", true)

	return config, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

func getBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}
