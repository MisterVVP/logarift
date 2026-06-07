package llmadapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientAcceptsAdapterResponseWithTruncationMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/enrich/friction-event" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"schema_version":"llm-adapter-response-v1",
			"request_id":"llm-test",
			"trace_id":"9966136fa3e6b91131fa8a0bcd217136",
			"adapter_version":"llm-adapter-0.1",
			"model_runtime":"ollama-compatible",
			"model_name":"logarift-enricher-qwen3-8b",
			"prompt_version":"friction-enrichment-prompt-0.1",
			"duration_ms":8964,
			"fields":{"workflow_stage":{"value":"test","confidence":0.9,"source":"local_llm"}},
			"warnings":[],
			"truncation":{"truncated":false,"original_character_count":54,"retained_character_count":54,"strategy":"head_tail"}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, time.Second)
	resp, err := client.Enrich(context.Background(), Request{RequestID: "llm-test", SchemaVersion: RequestSchemaVersion})
	if err != nil {
		t.Fatalf("Enrich() error: %v", err)
	}
	if resp.TruncationMetadata == nil || resp.TruncationMetadata.Strategy != "head_tail" {
		t.Fatalf("expected truncation metadata, got %#v", resp.TruncationMetadata)
	}
	if resp.Fields["workflow_stage"].Value != "test" {
		t.Fatalf("expected workflow_stage field, got %#v", resp.Fields)
	}
}
