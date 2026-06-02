package domain

import (
	"fmt"
	"strings"
	"time"
)

var allowedWorkflowStages = set("planning", "local_development", "build", "test", "code_review", "merge", "deployment", "operation", "debugging", "documentation", "coordination", "learning")
var allowedFrictionLayers = set("technical", "temporal", "cognitive", "social_process", "motivational", "environmental")
var allowedFrictionTypes = set("slow_feedback", "failed_feedback", "unclear_error", "missing_documentation", "ambiguous_ownership", "interruption", "waiting_for_review", "waiting_for_ci", "context_switch", "rework", "tooling_failure", "environment_setup", "coordination_overhead", "decision_blocked")
var allowedSources = set("manual", "seed", "import")
var allowedGoalStatuses = set("active", "completed", "deferred", "abandoned")
var allowedExportTypes = set("json")
var allowedExportStatuses = set("pending", "completed", "failed")

func set(values ...string) map[string]struct{} {
	m := map[string]struct{}{}
	for _, v := range values {
		m[v] = struct{}{}
	}
	return m
}
func in(v string, m map[string]struct{}) bool { _, ok := m[v]; return ok }
func requiredTime(name string, v time.Time) error {
	if v.IsZero() {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}
func schema(v int) error {
	if v != CurrentSchemaVersion {
		return fmt.Errorf("schema_version must be %d", CurrentSchemaVersion)
	}
	return nil
}
func trimRequired(name, v string) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return v, nil
}
func positive(name string, v float64) error {
	if v <= 0 {
		return fmt.Errorf("%s must be positive", name)
	}
	return nil
}

func (e *FrictionEvent) Validate() error {
	if err := schema(e.SchemaVersion); err != nil {
		return err
	}
	if err := requiredTime("timestamp_start", e.TimestampStart); err != nil {
		return err
	}
	if e.TimestampEnd != nil && e.TimestampEnd.Before(e.TimestampStart) {
		return fmt.Errorf("timestamp_end must be greater than or equal to timestamp_start")
	}
	if !in(e.WorkflowStage, allowedWorkflowStages) {
		return fmt.Errorf("workflow_stage is invalid")
	}
	if !in(e.FrictionLayer, allowedFrictionLayers) {
		return fmt.Errorf("friction_layer is invalid")
	}
	if !in(e.FrictionType, allowedFrictionTypes) {
		return fmt.Errorf("friction_type is invalid")
	}
	if e.SeveritySelf < 1 || e.SeveritySelf > 5 {
		return fmt.Errorf("severity_self must be between 1 and 5")
	}
	if e.CognitiveLoadSelf < 1 || e.CognitiveLoadSelf > 5 {
		return fmt.Errorf("cognitive_load_self must be between 1 and 5")
	}
	if e.EmotionValence < -2 || e.EmotionValence > 2 {
		return fmt.Errorf("emotion_valence must be between -2 and 2")
	}
	if e.TimeLostMinutes < 0 {
		return fmt.Errorf("time_lost_minutes must be >= 0")
	}
	if e.ResumeTimeMinutes < 0 {
		return fmt.Errorf("resume_time_minutes must be >= 0")
	}
	if e.RecoveryMinutes < 0 {
		return fmt.Errorf("recovery_minutes must be >= 0")
	}
	if e.InterruptionCount < 0 {
		return fmt.Errorf("interruption_count must be >= 0")
	}
	if !in(e.Source, allowedSources) {
		return fmt.Errorf("source is invalid")
	}
	e.Tags = NormalizeTags(e.Tags)
	return nil
}
func NormalizeTags(tags []string) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}
func (g *WorkGoal) Validate() error {
	if err := schema(g.SchemaVersion); err != nil {
		return err
	}
	title, err := trimRequired("title", g.Title)
	if err != nil {
		return err
	}
	g.Title = title
	if !in(g.Status, allowedGoalStatuses) {
		return fmt.Errorf("status is invalid")
	}
	return nil
}
func (s *WorkSession) Validate() error {
	if err := schema(s.SchemaVersion); err != nil {
		return err
	}
	title, err := trimRequired("title", s.Title)
	if err != nil {
		return err
	}
	s.Title = title
	if err := requiredTime("started_at", s.StartedAt); err != nil {
		return err
	}
	if s.EndedAt != nil && s.EndedAt.Before(s.StartedAt) {
		return fmt.Errorf("ended_at must be greater than or equal to started_at")
	}
	return nil
}
func (s *ScoreSnapshot) Validate() error {
	if err := schema(s.SchemaVersion); err != nil {
		return err
	}
	mv, err := trimRequired("model_version", s.ModelVersion)
	if err != nil {
		return err
	}
	s.ModelVersion = mv
	if err := requiredTime("period_start", s.PeriodStart); err != nil {
		return err
	}
	if err := requiredTime("period_end", s.PeriodEnd); err != nil {
		return err
	}
	if s.PeriodEnd.Before(s.PeriodStart) {
		return fmt.Errorf("period_end must be greater than or equal to period_start")
	}
	st, err := trimRequired("score_type", s.ScoreType)
	if err != nil {
		return err
	}
	s.ScoreType = st
	return nil
}
func (m *ModelConfig) Validate() error {
	if err := schema(m.SchemaVersion); err != nil {
		return err
	}
	mv, err := trimRequired("model_version", m.ModelVersion)
	if err != nil {
		return err
	}
	m.ModelVersion = mv
	name, err := trimRequired("name", m.Name)
	if err != nil {
		return err
	}
	m.Name = name
	p := m.Parameters
	for n, v := range map[string]float64{"cla_decay": p.CLADecay, "severity_multiplier": p.SeverityMultiplier, "cognitive_load_multiplier": p.CognitiveLoadMultiplier, "interruption_multiplier": p.InterruptionMultiplier, "recovery_multiplier": p.RecoveryMultiplier, "fci_half_life_minutes": p.FCIHalfLifeMinutes} {
		if err := positive(n, v); err != nil {
			return err
		}
	}
	return nil
}
func (e *ExportRecord) Validate() error {
	if err := schema(e.SchemaVersion); err != nil {
		return err
	}
	if !in(e.ExportType, allowedExportTypes) {
		return fmt.Errorf("export_type is invalid")
	}
	if !in(e.Status, allowedExportStatuses) {
		return fmt.Errorf("status is invalid")
	}
	if e.PeriodStart != nil && e.PeriodEnd != nil && e.PeriodEnd.Before(*e.PeriodStart) {
		return fmt.Errorf("period_end must be greater than or equal to period_start")
	}
	return nil
}
