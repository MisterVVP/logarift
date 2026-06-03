package ontology

import "testing"

func TestMVP3OntologyMembership(t *testing.T) {
	if !IsWorkflowStage("test") || IsWorkflowStage("qa") {
		t.Fatalf("workflow stage validation mismatch")
	}
	if !IsFrictionLayer("technical") || IsFrictionLayer("product") {
		t.Fatalf("friction layer validation mismatch")
	}
	if !IsFrictionType("failed_feedback") || IsFrictionType("meeting") {
		t.Fatalf("friction type validation mismatch")
	}
	if !IsEventSource("manual") || IsEventSource("telemetry") {
		t.Fatalf("event source validation mismatch")
	}
	if !IsGoalStatus("active") || IsGoalStatus("blocked") {
		t.Fatalf("goal status validation mismatch")
	}
}
