package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type OllamaClient struct {
	baseURL string
	model   string
	http    *http.Client
}

func NewOllamaClient(cfg Config) *OllamaClient {
	return &OllamaClient{baseURL: strings.TrimRight(cfg.RuntimeURL, "/"), model: cfg.Model, http: &http.Client{Timeout: cfg.RequestTimeout}}
}

type RuntimeError struct {
	Code       string
	Message    string
	HTTPStatus int
	Endpoint   string
	Hint       string
}

func (e RuntimeError) Error() string {
	parts := []string{e.Code}
	if e.Message != "" {
		parts = append(parts, e.Message)
	}
	if e.HTTPStatus > 0 {
		parts = append(parts, fmt.Sprintf("http_status=%d", e.HTTPStatus))
	}
	if e.Hint != "" {
		parts = append(parts, "hint="+e.Hint)
	}
	return strings.Join(parts, "; ")
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Think    *bool           `json:"think,omitempty"`
	Format   string          `json:"format"`
	Options  map[string]any  `json:"options,omitempty"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatResponse struct {
	Model     string        `json:"model"`
	Message   ollamaMessage `json:"message"`
	Done      bool          `json:"done"`
	Total     int64         `json:"total_duration"`
	CreatedAt time.Time     `json:"created_at"`
}

type ollamaTagsResponse struct {
	Models []ollamaModel `json:"models"`
}

type ollamaModel struct {
	Name   string `json:"name"`
	Model  string `json:"model"`
	Digest string `json:"digest"`
}

func (c *OllamaClient) Chat(ctx context.Context, system, user string) (string, error) {
	think := false
	body := ollamaChatRequest{
		Model:    c.model,
		Messages: []ollamaMessage{{Role: "system", Content: system}, {Role: "user", Content: user}},
		Stream:   false,
		Think:    &think,
		Format:   "json",
		Options:  map[string]any{"temperature": 0, "num_predict": 1024},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	endpoint := c.baseURL + "/api/chat"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", RuntimeError{Code: "ollama_request_build_failed", Message: err.Error(), Endpoint: endpoint}
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", runtimeHTTPError("ollama_chat_http_error", endpoint, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		limited, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", RuntimeError{Code: "ollama_chat_http_status", Message: strings.TrimSpace(string(limited)), HTTPStatus: resp.StatusCode, Endpoint: endpoint, Hint: ollamaStatusHint(resp.StatusCode)}
	}
	var out ollamaChatResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2*1024*1024)).Decode(&out); err != nil {
		return "", RuntimeError{Code: "ollama_chat_decode_failed", Message: err.Error(), Endpoint: endpoint}
	}
	return out.Message.Content, nil
}

func (c *OllamaClient) ModelInfo(ctx context.Context) (ModelInfo, error) {
	info := ModelInfo{AdapterVersion: AdapterVersion, ModelRuntime: ModelRuntime, ModelName: c.model, RuntimeURL: c.baseURL, Available: false}
	endpoint := c.baseURL + "/api/tags"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return info, RuntimeError{Code: "ollama_request_build_failed", Message: err.Error(), Endpoint: endpoint}
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return info, runtimeHTTPError("ollama_tags_http_error", endpoint, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		limited, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return info, RuntimeError{Code: "ollama_tags_http_status", Message: strings.TrimSpace(string(limited)), HTTPStatus: resp.StatusCode, Endpoint: endpoint, Hint: ollamaStatusHint(resp.StatusCode)}
	}
	var tags ollamaTagsResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1024*1024)).Decode(&tags); err != nil {
		return info, RuntimeError{Code: "ollama_tags_decode_failed", Message: err.Error(), Endpoint: endpoint}
	}
	available := make([]string, 0, len(tags.Models))
	for _, model := range tags.Models {
		name := model.Name
		if name == "" {
			name = model.Model
		}
		if name != "" {
			available = append(available, name)
		}
		if model.Name == c.model || model.Model == c.model || strings.TrimSuffix(model.Name, ":latest") == c.model {
			info.Available = true
			info.ModelDigest = model.Digest
			return info, nil
		}
	}
	return info, RuntimeError{Code: "ollama_model_not_found", Message: fmt.Sprintf("configured model %q is not available; available models: %s", c.model, strings.Join(available, ", ")), Endpoint: endpoint, Hint: "run `ollama list` on the host and set LOGARIFT_LLM_MODEL to the exact model name, or create the alias with `ollama create`"}
}

func runtimeHTTPError(code, endpoint string, err error) RuntimeError {
	message := err.Error()
	hint := ""
	if errors.Is(err, context.DeadlineExceeded) {
		hint = "increase LOGARIFT_LLM_REQUEST_TIMEOUT_MS or verify the model is already loaded"
	} else if parsed, parseErr := url.Parse(endpoint); parseErr == nil && parsed.Hostname() == "host.docker.internal" {
		hint = "verify host Ollama is reachable from the container; on Linux, Ollama may need OLLAMA_HOST=0.0.0.0:11434 and a service restart"
	}
	return RuntimeError{Code: code, Message: message, Endpoint: endpoint, Hint: hint}
}

func ollamaStatusHint(status int) string {
	switch status {
	case http.StatusNotFound:
		return "verify LOGARIFT_LLM_MODEL matches `ollama list` or create the model alias with `ollama create`"
	case http.StatusBadRequest:
		return "verify the installed Ollama version supports /api/chat JSON format and think=false"
	default:
		return "check Ollama host logs and model availability"
	}
}
