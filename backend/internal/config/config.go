package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	defaultAPIHost            = "0.0.0.0"
	defaultAPIPort            = "8080"
	defaultMongoDBURI         = "mongodb://localhost:27017"
	defaultMongoDBDatabase    = "logarift"
	defaultMathEnginePath     = "./bin/friction-math"
	defaultExportDir          = "./exports"
	defaultReadinessTimeoutMS = 2000
	defaultShutdownTimeoutMS  = 5000
)

// Config contains runtime settings for the local-first backend.
type Config struct {
	APIHost          string
	APIPort          string
	MongoDBURI       string
	MongoDBDatabase  string
	MathEnginePath   string
	ExportDir        string
	ReadinessTimeout time.Duration
	ShutdownTimeout  time.Duration
}

// Load reads configuration from environment variables and applies local-first
// defaults that work for direct local execution. Docker Compose overrides the
// MongoDB URI to use the container service name.
func Load() (Config, error) {
	cfg := Config{
		APIHost:          getenv("LOGARIFT_API_HOST", defaultAPIHost),
		APIPort:          getenv("LOGARIFT_API_PORT", defaultAPIPort),
		MongoDBURI:       getenv("LOGARIFT_MONGODB_URI", defaultMongoDBURI),
		MongoDBDatabase:  getenv("LOGARIFT_MONGODB_DATABASE", defaultMongoDBDatabase),
		MathEnginePath:   getenv("LOGARIFT_MATH_ENGINE_PATH", defaultMathEnginePath),
		ExportDir:        getenv("LOGARIFT_EXPORT_DIR", defaultExportDir),
		ReadinessTimeout: time.Duration(defaultReadinessTimeoutMS) * time.Millisecond,
		ShutdownTimeout:  time.Duration(defaultShutdownTimeoutMS) * time.Millisecond,
	}

	readinessTimeout, err := getenvDurationMS("LOGARIFT_READINESS_TIMEOUT_MS", defaultReadinessTimeoutMS)
	if err != nil {
		return Config{}, err
	}
	cfg.ReadinessTimeout = readinessTimeout

	shutdownTimeout, err := getenvDurationMS("LOGARIFT_SHUTDOWN_TIMEOUT_MS", defaultShutdownTimeoutMS)
	if err != nil {
		return Config{}, err
	}
	cfg.ShutdownTimeout = shutdownTimeout

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if c.APIHost == "" {
		return errors.New("LOGARIFT_API_HOST must not be empty")
	}
	if c.APIPort == "" {
		return errors.New("LOGARIFT_API_PORT must not be empty")
	}
	port, err := strconv.Atoi(c.APIPort)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("LOGARIFT_API_PORT must be a valid TCP port, got %q", c.APIPort)
	}
	if c.MongoDBURI == "" {
		return errors.New("LOGARIFT_MONGODB_URI must not be empty")
	}
	if c.MongoDBDatabase == "" {
		return errors.New("LOGARIFT_MONGODB_DATABASE must not be empty")
	}
	if c.MathEnginePath == "" {
		return errors.New("LOGARIFT_MATH_ENGINE_PATH must not be empty")
	}
	if c.ExportDir == "" {
		return errors.New("LOGARIFT_EXPORT_DIR must not be empty")
	}
	if c.ReadinessTimeout <= 0 {
		return errors.New("LOGARIFT_READINESS_TIMEOUT_MS must be greater than zero")
	}
	if c.ShutdownTimeout <= 0 {
		return errors.New("LOGARIFT_SHUTDOWN_TIMEOUT_MS must be greater than zero")
	}
	return nil
}

func (c Config) Address() string {
	return net.JoinHostPort(c.APIHost, c.APIPort)
}

// PublicStatus returns a sanitized configuration view for status endpoints.
func (c Config) PublicStatus() map[string]any {
	return map[string]any{
		"api_host":               c.APIHost,
		"api_port":               c.APIPort,
		"mongodb_database":       c.MongoDBDatabase,
		"mongodb_uri_configured": c.MongoDBURI != "",
		"math_engine_path":       c.MathEnginePath,
		"export_dir":             c.ExportDir,
	}
}

func getenv(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}

func getenvDurationMS(name string, fallback int) (time.Duration, error) {
	value := getenv(name, strconv.Itoa(fallback))
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer number of milliseconds: %w", name, err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", name)
	}
	return time.Duration(parsed) * time.Millisecond, nil
}
