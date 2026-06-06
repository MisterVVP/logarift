package adapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type fakeRuntime struct{ content string }

func (f fakeRuntime) Chat(context.Context, string, string) (string, error) { return f.content, nil }
func (f fakeRuntime) ModelInfo(context.Context) (ModelInfo, error) {
	return ModelInfo{AdapterVersion: AdapterVersion, ModelRuntime: ModelRuntime, ModelName: "qwen3.6", ModelDigest: "sha256:abc", Available: true}, nil
}

func TestEnrichNormalizesModelOutput(t *testing.T) {
	cfg := Config{BindHost: "127.0.0.1", Port: "8091", RuntimeURL: "http://localhost:11434", Model: "qwen3.6", RequestTimeout: time.Second, MaxInputChars: 12000, MaxPromptTokens: 8192, TruncationStrategy: "head_tail"}
	modelJSON := `{"fields":{"workflow_stage":{"value":"test","confidence":0.9,"source":"local_llm","explanation":"CI validation failed."},"friction_type":{"value":"invalid","confidence":0.9,"source":"local_llm"},"time_lost_minutes":{"value":20,"confidence":0.95,"source":"observed_text"},"tags":{"value":["CI","timeout","ci"],"confidence":0.8,"source":"local_llm"}},"warnings":[]}`
	svc := NewService(cfg, fakeRuntime{content: modelJSON}, nil)
	body := `{"request_id":"req-1","schema_version":"llm-adapter-request-v1","occurred_at":"2026-06-04T19:26:00Z","observed":{"friction_level":"orange","plain_text":"CI failed after 20 min.","links":[],"attachment_metadata":[]},"deterministic_baseline":{"workflow_stage":"test"},"allowed_values":{"workflow_stage":["test"],"friction_layer":["technical"],"friction_type":["failed_feedback"]}}`
	req := httptest.NewRequest(http.MethodPost, "/v1/enrich/friction-event", strings.NewReader(body))
	w := httptest.NewRecorder()
	svc.Router().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	var resp EnrichResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Fields["workflow_stage"].Value != "test" {
		t.Fatalf("workflow stage not accepted: %#v", resp.Fields)
	}
	if _, ok := resp.Fields["friction_type"]; ok {
		t.Fatalf("invalid friction_type should be rejected")
	}
	if len(resp.Warnings) == 0 {
		t.Fatalf("expected rejection warning")
	}
}

func TestTruncateHeadTail(t *testing.T) {
	text := strings.Repeat("a", 100) + "TAIL"
	out, meta := truncateHeadTail(text, 40)
	if !meta.Truncated || !strings.HasSuffix(out, "TAIL") {
		t.Fatalf("unexpected truncation: %#v %q", meta, out)
	}
}
