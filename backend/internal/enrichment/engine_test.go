package enrichment

import (
	"testing"
	"time"
)

func TestEngineEnrichesCIFailureFromQuickInput(t *testing.T) {
	now := time.Date(2026, 6, 4, 20, 0, 0, 0, time.UTC)
	event := NewEngine().Enrich(Input{
		OccurredAt:    now.Add(-time.Hour),
		FrictionLevel: LevelOrange,
		NotesMarkdown: "CI failed again after 20 min with an unclear timeout. https://github.com/org/repo/actions/runs/123",
	}, now)

	if event.InputMode != "quick" {
		t.Fatalf("expected quick input mode, got %q", event.InputMode)
	}
	if event.WorkflowStage != "test" || event.FrictionLayer != "technical" || event.FrictionType != "failed_feedback" {
		t.Fatalf("unexpected classification: %s %s %s", event.WorkflowStage, event.FrictionLayer, event.FrictionType)
	}
	if event.TimeLostMinutes != 20 {
		t.Fatalf("expected parsed 20 minute time loss, got %d", event.TimeLostMinutes)
	}
	if event.Observed == nil || len(event.Observed.Links) != 1 {
		t.Fatalf("expected observed link metadata, got %#v", event.Observed)
	}
	if event.Inference == nil || event.Inference.EngineVersion != EngineVersion {
		t.Fatalf("expected inference metadata, got %#v", event.Inference)
	}
	if event.Canonical == nil || event.Canonical.WorkflowStage != event.WorkflowStage {
		t.Fatalf("canonical fields not populated: %#v", event.Canonical)
	}
}
