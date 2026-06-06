package friction

import (
	"context"
	"testing"
	"time"

	"github.com/MisterVVP/logarift/backend/internal/llmadapter"
)

type fakeLLM struct{ resp llmadapter.Response }

func (f fakeLLM) Enrich(context.Context, llmadapter.Request) (llmadapter.Response, error) {
	return f.resp, nil
}

func TestCreateQuickMergesAcceptedLLMFieldsAndPersistsRejected(t *testing.T) {
	now := time.Date(2026, 6, 4, 22, 0, 0, 0, time.UTC)
	repo := &fakeRepo{}
	llm := fakeLLM{resp: llmadapter.Response{SchemaVersion: "llm-adapter-response-v1", RequestID: "req-1", AdapterVersion: "llm-adapter-0.1", ModelRuntime: "ollama-compatible", ModelName: "qwen3.6", PromptVersion: "friction-enrichment-prompt-0.1", Fields: map[string]llmadapter.Field{
		"workflow_stage": {Value: "debugging", Confidence: 0.91, Source: "local_llm", Explanation: "Debugging a runtime issue."},
		"friction_type":  {Value: "tooling_failure", Confidence: 0.4, Source: "local_llm", Explanation: "Too uncertain."},
		"tags":           {Value: []any{"debug", "timeout"}, Confidence: 0.8, Source: "local_llm"},
	}}}
	svc := NewServiceWithLLM(dispatcherForRepo(t, repo), fixedClock{now}, llm, 0.70, false)
	got, err := svc.CreateQuick(context.Background(), QuickRequest{OccurredAt: now.Add(-10 * time.Minute), FrictionLevel: "orange", NotesMarkdown: "Debug timeout in local tool for 10 min."})
	if err != nil {
		t.Fatalf("CreateQuick() error: %v", err)
	}
	if got.WorkflowStage != "debugging" {
		t.Fatalf("expected llm workflow stage, got %q", got.WorkflowStage)
	}
	if got.Inference == nil || got.Inference.LocalLLM == nil {
		t.Fatalf("missing llm metadata: %#v", got.Inference)
	}
	if _, ok := got.Inference.LocalLLM.RejectedFields["friction_type"]; !ok {
		t.Fatalf("expected rejected low-confidence field")
	}
	if got.Inference.EngineType != hybridEngineType {
		t.Fatalf("expected hybrid engine type, got %q", got.Inference.EngineType)
	}
}
