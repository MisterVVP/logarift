package httpserver

import (
	"testing"

	"github.com/MisterVVP/logarift/backend/internal/domain"
)

func TestFrictionEventPayloadIncludesLLMMergeSummary(t *testing.T) {
	event := domain.FrictionEvent{Inference: &domain.FrictionInference{LocalLLM: &domain.FrictionAdapterInference{
		AcceptedFields: map[string]domain.FrictionFieldInference{
			"workflow_stage": {Value: "debugging", Confidence: 0.91, Source: "local_llm"},
		},
		RejectedFields: map[string]domain.FrictionRejectedInference{
			"friction_type": {RejectionReason: "confidence_below_threshold"},
		},
		Warnings: []string{"deterministic fallback used for rejected fields"},
	}}}

	payload := frictionEventPayload(event)
	if payload.LLMMergeSummary == nil {
		t.Fatal("expected LLM merge summary")
	}
	if payload.LLMMergeSummary.AcceptedFieldsCount != 1 || payload.LLMMergeSummary.RejectedFieldsCount != 1 || payload.LLMMergeSummary.WarningCount != 1 {
		t.Fatalf("unexpected summary counts: %#v", payload.LLMMergeSummary)
	}
	if got := payload.LLMMergeSummary.RejectedFieldNames; len(got) != 1 || got[0] != "friction_type" {
		t.Fatalf("unexpected rejected field names: %#v", got)
	}
}
