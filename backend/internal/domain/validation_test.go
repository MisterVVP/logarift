package domain

import (
	"strings"
	"testing"
	"time"
)

func validEvent() FrictionEvent {
	return FrictionEvent{SchemaVersion: CurrentSchemaVersion, TimestampStart: time.Now().UTC(), WorkflowStage: "build", FrictionLayer: "technical", FrictionType: "slow_feedback", SeveritySelf: 3, CognitiveLoadSelf: 4, EmotionValence: 0, TimeLostMinutes: 1, ResumeTimeMinutes: 0, RecoveryMinutes: 0, InterruptionCount: 0, Source: "manual", Tags: []string{" ci ", "ci", ""}}
}
func TestFrictionEventValidationAndTagNormalization(t *testing.T) {
	e := validEvent()
	if err := e.Validate(); err != nil {
		t.Fatalf("Validate() error: %v", err)
	}
	if len(e.Tags) != 1 || e.Tags[0] != "ci" {
		t.Fatalf("tags not normalized: %#v", e.Tags)
	}
}
func TestFrictionEventRejectsEnumsAndRanges(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*FrictionEvent)
		want   string
	}{{"stage", func(e *FrictionEvent) { e.WorkflowStage = "bad" }, "workflow_stage"}, {"severity", func(e *FrictionEvent) { e.SeveritySelf = 6 }, "severity_self"}, {"valence", func(e *FrictionEvent) { e.EmotionValence = -3 }, "emotion_valence"}, {"negative", func(e *FrictionEvent) { e.TimeLostMinutes = -1 }, "time_lost_minutes"}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := validEvent()
			tc.mutate(&e)
			err := e.Validate()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q error, got %v", tc.want, err)
			}
		})
	}
}
func TestTimestampValidation(t *testing.T) {
	start := time.Now().UTC()
	end := start.Add(-time.Minute)
	e := validEvent()
	e.TimestampEnd = &end
	if err := e.Validate(); err == nil {
		t.Fatal("expected end before start error")
	}
	s := WorkSession{SchemaVersion: CurrentSchemaVersion, Title: "test", StartedAt: start, EndedAt: &end}
	if err := s.Validate(); err == nil {
		t.Fatal("expected session end before start error")
	}
}
func TestOtherModelValidation(t *testing.T) {
	now := time.Now().UTC()
	goal := WorkGoal{SchemaVersion: CurrentSchemaVersion, Title: " goal ", Status: "active"}
	if err := goal.Validate(); err != nil {
		t.Fatalf("goal valid: %v", err)
	}
	session := WorkSession{SchemaVersion: CurrentSchemaVersion, Title: "session", StartedAt: now}
	if err := session.Validate(); err != nil {
		t.Fatalf("session valid: %v", err)
	}
	snap := ScoreSnapshot{SchemaVersion: CurrentSchemaVersion, ModelVersion: "mvp-0.1", PeriodStart: now, PeriodEnd: now, ScoreType: "daily"}
	if err := snap.Validate(); err != nil {
		t.Fatalf("snapshot valid: %v", err)
	}
	cfg := DefaultModelConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("config valid: %v", err)
	}
	exp := ExportRecord{SchemaVersion: CurrentSchemaVersion, ExportType: "json", Status: "pending"}
	if err := exp.Validate(); err != nil {
		t.Fatalf("export valid: %v", err)
	}
}
