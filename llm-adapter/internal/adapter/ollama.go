package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
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
	body := ollamaChatRequest{
		Model:    c.model,
		Messages: []ollamaMessage{{Role: "system", Content: system}, {Role: "user", Content: user}},
		Stream:   false,
		Format:   "json",
		Options:  map[string]any{"temperature": 0, "num_predict": 1024},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		limited, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("ollama chat returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(limited)))
	}
	var out ollamaChatResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2*1024*1024)).Decode(&out); err != nil {
		return "", err
	}
	return out.Message.Content, nil
}

func (c *OllamaClient) ModelInfo(ctx context.Context) (ModelInfo, error) {
	info := ModelInfo{AdapterVersion: AdapterVersion, ModelRuntime: ModelRuntime, ModelName: c.model, RuntimeURL: c.baseURL, Available: false}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/tags", nil)
	if err != nil {
		return info, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return info, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return info, fmt.Errorf("ollama tags returned HTTP %d", resp.StatusCode)
	}
	var tags ollamaTagsResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1024*1024)).Decode(&tags); err != nil {
		return info, err
	}
	for _, model := range tags.Models {
		if model.Name == c.model || model.Model == c.model || strings.TrimSuffix(model.Name, ":latest") == c.model {
			info.Available = true
			info.ModelDigest = model.Digest
			return info, nil
		}
	}
	return info, fmt.Errorf("configured model %q is not available", c.model)
}
