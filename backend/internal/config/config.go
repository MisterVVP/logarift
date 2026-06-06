package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAPIHost                 = "0.0.0.0"
	defaultAPIPort                 = "8080"
	defaultMongoDBURI              = "mongodb://localhost:27017"
	defaultMongoDBDatabase         = "logarift"
	defaultMathEngineURL           = "http://localhost:8090"
	defaultExportDir               = "./exports"
	defaultUploadDir               = "./data/uploads"
	defaultReadinessTimeoutMS      = 2000
	defaultShutdownTimeoutMS       = 5000
	defaultMongoDBConnectTimeoutMS = 5000
	defaultLLMAdapterURL           = "http://localhost:8091"
	defaultLLMAdapterTimeoutMS     = 1500
	defaultLLMAdapterMinConfidence = 0.70
	defaultLLMAdapterPrivacyMode   = "text_only"
)

// Config contains runtime settings for the local-first backend.
type Config struct {
	APIHost                      string
	APIPort                      string
	MongoDBURI                   string
	MongoDBDatabase              string
	MathEngineURL                string
	ExportDir                    string
	UploadDir                    string
	ReadinessTimeout             time.Duration
	ShutdownTimeout              time.Duration
	MongoDBConnectTimeout        time.Duration
	LLMAdapterEnabled            bool
	LLMAdapterURL                string
	LLMAdapterTimeout            time.Duration
	LLMAdapterMinConfidence      float64
	LLMAdapterPromptPrivacyMode  string
	LLMAdapterAllowRemoteRuntime bool
}

// Load reads configuration from environment variables and applies local-first
// defaults that work for direct local execution. Docker Compose overrides the
// MongoDB URI to use the container service name.
func Load() (Config, error) {
	cfg := Config{
		APIHost:                      getenv("LOGARIFT_API_HOST", defaultAPIHost),
		APIPort:                      getenv("LOGARIFT_API_PORT", defaultAPIPort),
		MongoDBURI:                   getenv("LOGARIFT_MONGODB_URI", defaultMongoDBURI),
		MongoDBDatabase:              getenv("LOGARIFT_MONGODB_DATABASE", defaultMongoDBDatabase),
		MathEngineURL:                getenv("LOGARIFT_MATH_ENGINE_URL", defaultMathEngineURL),
		ExportDir:                    getenv("LOGARIFT_EXPORT_DIR", defaultExportDir),
		UploadDir:                    getenv("LOGARIFT_UPLOAD_DIR", defaultUploadDir),
		ReadinessTimeout:             time.Duration(defaultReadinessTimeoutMS) * time.Millisecond,
		ShutdownTimeout:              time.Duration(defaultShutdownTimeoutMS) * time.Millisecond,
		MongoDBConnectTimeout:        time.Duration(defaultMongoDBConnectTimeoutMS) * time.Millisecond,
		LLMAdapterEnabled:            getenvBool("LOGARIFT_LLM_ADAPTER_ENABLED", false),
		LLMAdapterURL:                getenv("LOGARIFT_LLM_ADAPTER_URL", defaultLLMAdapterURL),
		LLMAdapterTimeout:            time.Duration(defaultLLMAdapterTimeoutMS) * time.Millisecond,
		LLMAdapterMinConfidence:      defaultLLMAdapterMinConfidence,
		LLMAdapterPromptPrivacyMode:  getenv("LOGARIFT_LLM_ADAPTER_PROMPT_PRIVACY_MODE", defaultLLMAdapterPrivacyMode),
		LLMAdapterAllowRemoteRuntime: getenvBool("LOGARIFT_LLM_ADAPTER_ALLOW_REMOTE_RUNTIME", false),
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

	connectTimeout, err := getenvDurationMS("LOGARIFT_MONGODB_CONNECT_TIMEOUT_MS", defaultMongoDBConnectTimeoutMS)
	if err != nil {
		return Config{}, err
	}
	cfg.MongoDBConnectTimeout = connectTimeout

	adapterTimeout, err := getenvDurationMS("LOGARIFT_LLM_ADAPTER_TIMEOUT_MS", defaultLLMAdapterTimeoutMS)
	if err != nil {
		return Config{}, err
	}
	cfg.LLMAdapterTimeout = adapterTimeout

	adapterMinConfidence, err := getenvFloat("LOGARIFT_LLM_ADAPTER_MIN_CONFIDENCE", defaultLLMAdapterMinConfidence)
	if err != nil {
		return Config{}, err
	}
	cfg.LLMAdapterMinConfidence = adapterMinConfidence

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
	if c.MathEngineURL == "" {
		return errors.New("LOGARIFT_MATH_ENGINE_URL must not be empty")
	}
	parsedMathURL, err := url.Parse(c.MathEngineURL)
	if err != nil || parsedMathURL.Scheme == "" || parsedMathURL.Host == "" {
		return fmt.Errorf("LOGARIFT_MATH_ENGINE_URL must be an absolute HTTP URL, got %q", c.MathEngineURL)
	}
	if parsedMathURL.Scheme != "http" && parsedMathURL.Scheme != "https" {
		return fmt.Errorf("LOGARIFT_MATH_ENGINE_URL must use http or https, got %q", parsedMathURL.Scheme)
	}
	if c.ExportDir == "" {
		return errors.New("LOGARIFT_EXPORT_DIR must not be empty")
	}
	if c.UploadDir == "" {
		return errors.New("LOGARIFT_UPLOAD_DIR must not be empty")
	}
	if c.ReadinessTimeout <= 0 {
		return errors.New("LOGARIFT_READINESS_TIMEOUT_MS must be greater than zero")
	}
	if c.ShutdownTimeout <= 0 {
		return errors.New("LOGARIFT_SHUTDOWN_TIMEOUT_MS must be greater than zero")
	}
	if c.MongoDBConnectTimeout <= 0 {
		return errors.New("LOGARIFT_MONGODB_CONNECT_TIMEOUT_MS must be greater than zero")
	}
	if c.LLMAdapterURL == "" {
		return errors.New("LOGARIFT_LLM_ADAPTER_URL must not be empty")
	}
	parsedAdapterURL, err := url.Parse(c.LLMAdapterURL)
	if err != nil || parsedAdapterURL.Scheme == "" || parsedAdapterURL.Host == "" {
		return fmt.Errorf("LOGARIFT_LLM_ADAPTER_URL must be an absolute HTTP URL, got %q", c.LLMAdapterURL)
	}
	if parsedAdapterURL.Scheme != "http" && parsedAdapterURL.Scheme != "https" {
		return fmt.Errorf("LOGARIFT_LLM_ADAPTER_URL must use http or https, got %q", parsedAdapterURL.Scheme)
	}
	if c.LLMAdapterTimeout <= 0 {
		return errors.New("LOGARIFT_LLM_ADAPTER_TIMEOUT_MS must be greater than zero")
	}
	if c.LLMAdapterMinConfidence < 0 || c.LLMAdapterMinConfidence > 1 {
		return errors.New("LOGARIFT_LLM_ADAPTER_MIN_CONFIDENCE must be between 0 and 1")
	}
	if c.LLMAdapterPromptPrivacyMode != "text_only" && c.LLMAdapterPromptPrivacyMode != "markdown" {
		return errors.New("LOGARIFT_LLM_ADAPTER_PROMPT_PRIVACY_MODE must be text_only or markdown")
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
		"math_engine_url":        c.MathEngineURL,
		"export_dir":             c.ExportDir,
		"upload_dir":             c.UploadDir,
		"llm_adapter_enabled":    c.LLMAdapterEnabled,
		"llm_adapter_url":        c.LLMAdapterURL,
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

func getenvBool(name string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
	if value == "" {
		return fallback
	}
	return value == "true" || value == "1" || value == "yes"
}

func getenvFloat(name string, fallback float64) (float64, error) {
	value := getenv(name, strconv.FormatFloat(fallback, 'f', -1, 64))
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number: %w", name, err)
	}
	return parsed, nil
}
