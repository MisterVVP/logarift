package friction

import (
	"context"
	"testing"
	"time"

	"github.com/MisterVVP/logarift/backend/internal/domain"
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
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if repo.event != nil && repo.event.Inference != nil && repo.event.Inference.LocalLLM != nil {
			got.Event = *repo.event
			got.Enrichment = *repo.event.Enrichment
			break
		}
		time.Sleep(10 * time.Millisecond)
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

func TestCreateQuickTreatsEmptyAdapterFieldsAsPartialFallback(t *testing.T) {
	now := time.Date(2026, 6, 6, 20, 10, 0, 0, time.UTC)
	repo := &fakeRepo{}
	llm := fakeLLM{resp: llmadapter.Response{SchemaVersion: "llm-adapter-response-v1", RequestID: "req-empty", AdapterVersion: "llm-adapter-0.1", ModelRuntime: "ollama-compatible", ModelName: "qwen3.6", PromptVersion: "friction-enrichment-prompt-0.1", Fields: map[string]llmadapter.Field{}, Warnings: []string{"no_fields_accepted"}}}
	svc := NewServiceWithLLM(dispatcherForRepo(t, repo), fixedClock{now}, llm, 0.70, false)
	got, err := svc.CreateQuick(context.Background(), QuickRequest{OccurredAt: now.Add(-20 * time.Minute), FrictionLevel: "orange", NotesMarkdown: "CI failed again after 20 min with an unclear timeout."})
	if err != nil {
		t.Fatalf("CreateQuick() error: %v", err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if repo.event != nil && repo.event.Enrichment != nil && repo.event.Enrichment.LLMStatus == domain.LLMStatusPartiallySucceeded {
			got.Event = *repo.event
			got.Enrichment = *repo.event.Enrichment
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if got.Enrichment.LLMStatus != domain.LLMStatusPartiallySucceeded {
		t.Fatalf("expected partial LLM fallback status, got %#v", got.Enrichment)
	}
	if got.Inference == nil || got.Inference.LocalLLM == nil {
		t.Fatalf("expected local LLM metadata for empty adapter response")
	}
	if got.Enrichment.MergeSummary == nil || got.Enrichment.MergeSummary.FallbackFieldCount != 1 {
		t.Fatalf("expected fallback merge summary, got %#v", got.Enrichment.MergeSummary)
	}
	if got.Enrichment.MergeSummary.FieldDecisions["adapter_response"].Reason != "no_fields_returned" {
		t.Fatalf("expected no_fields_returned decision, got %#v", got.Enrichment.MergeSummary.FieldDecisions)
	}
}
