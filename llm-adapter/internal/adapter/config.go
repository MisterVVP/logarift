package adapter

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
	AdapterVersion = "llm-adapter-0.1"
	PromptVersion  = "friction-enrichment-prompt-0.1"
	ModelRuntime   = "ollama-compatible"
)

const (
	defaultPort               = "8091"
	defaultBindHost           = "127.0.0.1"
	defaultRuntimeURL         = "http://localhost:11434"
	defaultModel              = "qwen3.6"
	defaultRequestTimeoutMS   = 15000
	defaultMaxInputChars      = 12000
	defaultMaxPromptTokens    = 8192
	defaultTruncationStrategy = "head_tail"
)

type Config struct {
	BindHost           string
	Port               string
	RuntimeURL         string
	AllowRemoteRuntime bool
	Model              string
	RequestTimeout     time.Duration
	MaxInputChars      int
	MaxPromptTokens    int
	TruncationStrategy string
	LogPrompts         bool
	LogResponses       bool
	MockResponse       bool
}

func LoadConfig() (Config, error) {
	cfg := Config{
		BindHost:           getenv("LOGARIFT_LLM_ADAPTER_HOST", defaultBindHost),
		Port:               getenv("LOGARIFT_LLM_ADAPTER_PORT", defaultPort),
		RuntimeURL:         strings.TrimRight(getenv("LOGARIFT_LLM_RUNTIME_URL", defaultRuntimeURL), "/"),
		AllowRemoteRuntime: getenvBool("LOGARIFT_LLM_ADAPTER_ALLOW_REMOTE_RUNTIME", false),
		Model:              getenv("LOGARIFT_LLM_MODEL", defaultModel),
		RequestTimeout:     time.Duration(defaultRequestTimeoutMS) * time.Millisecond,
		MaxInputChars:      defaultMaxInputChars,
		MaxPromptTokens:    defaultMaxPromptTokens,
		TruncationStrategy: defaultTruncationStrategy,
		LogPrompts:         getenvBool("LOGARIFT_LLM_LOG_PROMPTS", false),
		LogResponses:       getenvBool("LOGARIFT_LLM_LOG_RESPONSES", false),
		MockResponse:       getenvBool("LOGARIFT_LLM_MOCK_RESPONSE_ENABLED", false),
	}
	var err error
	if cfg.RequestTimeout, err = getenvDurationMS("LOGARIFT_LLM_REQUEST_TIMEOUT_MS", defaultRequestTimeoutMS); err != nil {
		return Config{}, err
	}
	if cfg.MaxInputChars, err = getenvPositiveInt("LOGARIFT_LLM_MAX_INPUT_CHARS", defaultMaxInputChars); err != nil {
		return Config{}, err
	}
	if cfg.MaxPromptTokens, err = getenvPositiveInt("LOGARIFT_LLM_MAX_PROMPT_TOKENS", defaultMaxPromptTokens); err != nil {
		return Config{}, err
	}
	if strategy := getenv("LOGARIFT_LLM_TRUNCATION_STRATEGY", defaultTruncationStrategy); strategy != "head_tail" {
		return Config{}, fmt.Errorf("LOGARIFT_LLM_TRUNCATION_STRATEGY must be head_tail, got %q", strategy)
	} else {
		cfg.TruncationStrategy = strategy
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Address() string { return net.JoinHostPort(c.BindHost, c.Port) }

func (c Config) Validate() error {
	if c.Port == "" {
		return errors.New("LOGARIFT_LLM_ADAPTER_PORT must not be empty")
	}
	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("LOGARIFT_LLM_ADAPTER_PORT must be a valid TCP port, got %q", c.Port)
	}
	if c.Model == "" {
		return errors.New("LOGARIFT_LLM_MODEL must not be empty")
	}
	parsed, err := url.Parse(c.RuntimeURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("LOGARIFT_LLM_RUNTIME_URL must be an absolute HTTP URL, got %q", c.RuntimeURL)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("LOGARIFT_LLM_RUNTIME_URL must use http or https, got %q", parsed.Scheme)
	}
	if !c.AllowRemoteRuntime && !isAllowedLocalRuntimeHost(parsed.Hostname()) {
		return fmt.Errorf("LOGARIFT_LLM_RUNTIME_URL host %q is not local/private; set LOGARIFT_LLM_ADAPTER_ALLOW_REMOTE_RUNTIME=true to allow it", parsed.Hostname())
	}
	if c.RequestTimeout <= 0 || c.MaxInputChars <= 0 || c.MaxPromptTokens <= 0 {
		return errors.New("adapter timeouts and limits must be greater than zero")
	}
	return nil
}

func isAllowedLocalRuntimeHost(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	switch host {
	case "localhost", "host.docker.internal", "ollama", "ollama-runtime":
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()
}

func getenv(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func getenvBool(name string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
	if value == "" {
		return fallback
	}
	return value == "true" || value == "1" || value == "yes"
}

func getenvDurationMS(name string, fallback int) (time.Duration, error) {
	value, err := getenvPositiveInt(name, fallback)
	if err != nil {
		return 0, err
	}
	return time.Duration(value) * time.Millisecond, nil
}

func getenvPositiveInt(name string, fallback int) (int, error) {
	value := getenv(name, strconv.Itoa(fallback))
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer, got %q", name, value)
	}
	return parsed, nil
}
