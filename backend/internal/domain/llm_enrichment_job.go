package domain

import (
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	LLMStatusNotRequested       = "not_requested"
	LLMStatusQueued             = "queued"
	LLMStatusRunning            = "running"
	LLMStatusSucceeded          = "succeeded"
	LLMStatusPartiallySucceeded = "partially_succeeded"
	LLMStatusFailed             = "failed"
	LLMStatusTimedOut           = "timed_out"
	LLMStatusCancelled          = "cancelled"
	LLMStatusDisabled           = "disabled"
	LLMStatusNotQueued          = "not_queued"
)

var allowedLLMStatuses = set(LLMStatusNotRequested, LLMStatusQueued, LLMStatusRunning, LLMStatusSucceeded, LLMStatusPartiallySucceeded, LLMStatusFailed, LLMStatusTimedOut, LLMStatusCancelled, LLMStatusDisabled, LLMStatusNotQueued)

type FrictionEnrichment struct {
	LLMStatus           string                    `bson:"llm_status" json:"llm_status"`
	JobID               string                    `bson:"job_id,omitempty" json:"job_id,omitempty"`
	TraceID             string                    `bson:"trace_id,omitempty" json:"trace_id,omitempty"`
	DeterministicStatus string                    `bson:"deterministic_status" json:"deterministic_status"`
	UserMessage         string                    `bson:"user_message,omitempty" json:"user_message,omitempty"`
	UpdatedAt           time.Time                 `bson:"updated_at" json:"updated_at"`
	MergeSummary        *LLMEnrichmentMergeResult `bson:"merge_summary,omitempty" json:"merge_summary,omitempty"`
}

type LLMEnrichmentJob struct {
	ID            bson.ObjectID             `bson:"_id,omitempty" json:"id"`
	SchemaVersion int                       `bson:"schema_version" json:"schema_version"`
	EventID       bson.ObjectID             `bson:"event_id" json:"event_id"`
	RequestID     string                    `bson:"request_id" json:"request_id"`
	TraceID       string                    `bson:"trace_id" json:"trace_id"`
	Status        string                    `bson:"status" json:"status"`
	Attempt       int                       `bson:"attempt" json:"attempt"`
	MaxAttempts   int                       `bson:"max_attempts" json:"max_attempts"`
	CreatedAt     time.Time                 `bson:"created_at" json:"created_at"`
	ClaimedAt     *time.Time                `bson:"claimed_at,omitempty" json:"claimed_at,omitempty"`
	CompletedAt   *time.Time                `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	AdapterURL    string                    `bson:"adapter_url,omitempty" json:"adapter_url,omitempty"`
	ModelName     string                    `bson:"model_name,omitempty" json:"model_name,omitempty"`
	PromptVersion string                    `bson:"prompt_version,omitempty" json:"prompt_version,omitempty"`
	ErrorCode     string                    `bson:"error_code,omitempty" json:"error_code,omitempty"`
	WarningCount  int                       `bson:"warning_count" json:"warning_count"`
	LastError     string                    `bson:"last_error,omitempty" json:"last_error,omitempty"`
	UpdatedAt     time.Time                 `bson:"updated_at" json:"updated_at"`
	MergeSummary  *LLMEnrichmentMergeResult `bson:"merge_summary,omitempty" json:"merge_summary,omitempty"`
}

type LLMEnrichmentMergeResult struct {
	EventID            string                      `bson:"event_id" json:"event_id"`
	JobID              string                      `bson:"job_id" json:"job_id"`
	LLMStatus          string                      `bson:"llm_status" json:"llm_status"`
	AdapterResult      LLMAdapterResultSummary     `bson:"adapter_result" json:"adapter_result"`
	FieldDecisions     map[string]LLMFieldDecision `bson:"field_decisions" json:"field_decisions"`
	AcceptedFieldCount int                         `bson:"accepted_field_count" json:"accepted_field_count"`
	RejectedFieldCount int                         `bson:"rejected_field_count" json:"rejected_field_count"`
	FallbackFieldCount int                         `bson:"fallback_field_count" json:"fallback_field_count"`
}

type LLMAdapterResultSummary struct {
	AdapterVersion string `bson:"adapter_version,omitempty" json:"adapter_version,omitempty"`
	ModelRuntime   string `bson:"model_runtime,omitempty" json:"model_runtime,omitempty"`
	ModelName      string `bson:"model_name,omitempty" json:"model_name,omitempty"`
	PromptVersion  string `bson:"prompt_version,omitempty" json:"prompt_version,omitempty"`
	DurationMS     int64  `bson:"duration_ms,omitempty" json:"duration_ms,omitempty"`
	WarningCount   int    `bson:"warning_count" json:"warning_count"`
}

type LLMFieldDecision struct {
	AdapterValue      any     `bson:"adapter_value,omitempty" json:"adapter_value,omitempty"`
	AdapterConfidence float64 `bson:"adapter_confidence,omitempty" json:"adapter_confidence,omitempty"`
	CanonicalValue    any     `bson:"canonical_value,omitempty" json:"canonical_value,omitempty"`
	Decision          string  `bson:"decision" json:"decision"`
	Reason            string  `bson:"reason" json:"reason"`
}

func ValidLLMStatus(status string) bool { return in(status, allowedLLMStatuses) }

func (j *LLMEnrichmentJob) Validate() error {
	if err := schema(j.SchemaVersion); err != nil {
		return err
	}
	if j.EventID.IsZero() {
		return fmt.Errorf("event_id is required")
	}
	if strings.TrimSpace(j.RequestID) == "" {
		return fmt.Errorf("request_id is required")
	}
	if strings.TrimSpace(j.TraceID) == "" {
		return fmt.Errorf("trace_id is required")
	}
	if !ValidLLMStatus(j.Status) {
		return fmt.Errorf("status is invalid")
	}
	if j.MaxAttempts < 1 {
		return fmt.Errorf("max_attempts must be >= 1")
	}
	if j.Attempt < 0 {
		return fmt.Errorf("attempt must be >= 0")
	}
	if j.CreatedAt.IsZero() {
		return fmt.Errorf("created_at is required")
	}
	return nil
}
