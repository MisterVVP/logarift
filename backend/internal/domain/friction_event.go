package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type FrictionEvent struct {
	ID                bson.ObjectID       `bson:"_id,omitempty" json:"id"`
	SchemaVersion     int                 `bson:"schema_version" json:"schema_version"`
	InputMode         string              `bson:"input_mode,omitempty" json:"input_mode,omitempty"`
	Observed          *FrictionObserved   `bson:"observed,omitempty" json:"observed,omitempty"`
	Inference         *FrictionInference  `bson:"inference,omitempty" json:"inference,omitempty"`
	Canonical         *FrictionCanonical  `bson:"canonical,omitempty" json:"canonical,omitempty"`
	Enrichment        *FrictionEnrichment `bson:"enrichment,omitempty" json:"enrichment,omitempty"`
	TimestampStart    time.Time           `bson:"timestamp_start" json:"timestamp_start"`
	TimestampEnd      *time.Time          `bson:"timestamp_end,omitempty" json:"timestamp_end,omitempty"`
	WorkflowStage     string              `bson:"workflow_stage" json:"workflow_stage"`
	FrictionLayer     string              `bson:"friction_layer" json:"friction_layer"`
	FrictionType      string              `bson:"friction_type" json:"friction_type"`
	SeveritySelf      int                 `bson:"severity_self" json:"severity_self"`
	CognitiveLoadSelf int                 `bson:"cognitive_load_self" json:"cognitive_load_self"`
	EmotionValence    int                 `bson:"emotion_valence" json:"emotion_valence"`
	TimeLostMinutes   int                 `bson:"time_lost_minutes" json:"time_lost_minutes"`
	ResumeTimeMinutes int                 `bson:"resume_time_minutes" json:"resume_time_minutes"`
	RecoveryMinutes   int                 `bson:"recovery_minutes" json:"recovery_minutes"`
	InterruptionCount int                 `bson:"interruption_count" json:"interruption_count"`
	GoalID            *bson.ObjectID      `bson:"goal_id,omitempty" json:"goal_id,omitempty"`
	SessionID         *bson.ObjectID      `bson:"session_id,omitempty" json:"session_id,omitempty"`
	Tags              []string            `bson:"tags,omitempty" json:"tags,omitempty"`
	Notes             string              `bson:"notes,omitempty" json:"notes,omitempty"`
	Source            string              `bson:"source" json:"source"`
	CreatedAt         time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time           `bson:"updated_at" json:"updated_at"`
}

type FrictionObserved struct {
	OccurredAt    time.Time            `bson:"occurred_at" json:"occurred_at"`
	FrictionLevel string               `bson:"friction_level" json:"friction_level"`
	NotesMarkdown string               `bson:"notes_markdown" json:"notes_markdown"`
	PlainText     string               `bson:"plain_text" json:"plain_text"`
	Links         []FrictionLink       `bson:"links,omitempty" json:"links,omitempty"`
	Attachments   []FrictionAttachment `bson:"attachments,omitempty" json:"attachments,omitempty"`
}

type FrictionLink struct {
	URL    string `bson:"url" json:"url"`
	Source string `bson:"source" json:"source"`
}

type FrictionAttachment struct {
	ID          string `bson:"id,omitempty" json:"id,omitempty"`
	Filename    string `bson:"filename" json:"filename"`
	ContentType string `bson:"content_type" json:"content_type"`
	LocalPath   string `bson:"local_path,omitempty" json:"local_path,omitempty"`
}

type FrictionInference struct {
	EngineVersion string                            `bson:"engine_version" json:"engine_version"`
	EngineType    string                            `bson:"engine_type" json:"engine_type"`
	CreatedAt     time.Time                         `bson:"created_at" json:"created_at"`
	Fields        map[string]FrictionFieldInference `bson:"fields" json:"fields"`
	LocalLLM      *FrictionAdapterInference         `bson:"local_llm,omitempty" json:"local_llm,omitempty"`
}

type FrictionAdapterInference struct {
	RequestID      string                               `bson:"request_id,omitempty" json:"request_id,omitempty"`
	AdapterVersion string                               `bson:"adapter_version,omitempty" json:"adapter_version,omitempty"`
	ModelRuntime   string                               `bson:"model_runtime,omitempty" json:"model_runtime,omitempty"`
	ModelName      string                               `bson:"model_name,omitempty" json:"model_name,omitempty"`
	ModelDigest    string                               `bson:"model_digest,omitempty" json:"model_digest,omitempty"`
	PromptVersion  string                               `bson:"prompt_version,omitempty" json:"prompt_version,omitempty"`
	DurationMS     int64                                `bson:"duration_ms,omitempty" json:"duration_ms,omitempty"`
	Warnings       []string                             `bson:"warnings,omitempty" json:"warnings,omitempty"`
	ErrorCode      string                               `bson:"error_code,omitempty" json:"error_code,omitempty"`
	AcceptedFields map[string]FrictionFieldInference    `bson:"accepted_fields,omitempty" json:"accepted_fields,omitempty"`
	RejectedFields map[string]FrictionRejectedInference `bson:"rejected_fields,omitempty" json:"rejected_fields,omitempty"`
}

type FrictionRejectedInference struct {
	SuggestedValue  any     `bson:"suggested_value,omitempty" json:"suggested_value,omitempty"`
	Confidence      float64 `bson:"confidence,omitempty" json:"confidence,omitempty"`
	Source          string  `bson:"source,omitempty" json:"source,omitempty"`
	Explanation     string  `bson:"explanation,omitempty" json:"explanation,omitempty"`
	RejectionReason string  `bson:"rejection_reason" json:"rejection_reason"`
	AdapterVersion  string  `bson:"adapter_version,omitempty" json:"adapter_version,omitempty"`
	ModelName       string  `bson:"model_name,omitempty" json:"model_name,omitempty"`
	PromptVersion   string  `bson:"prompt_version,omitempty" json:"prompt_version,omitempty"`
}

type FrictionFieldInference struct {
	Value       any     `bson:"value" json:"value"`
	Confidence  float64 `bson:"confidence" json:"confidence"`
	Source      string  `bson:"source" json:"source"`
	Explanation string  `bson:"explanation,omitempty" json:"explanation,omitempty"`
}

type FrictionCanonical struct {
	WorkflowStage     string   `bson:"workflow_stage" json:"workflow_stage"`
	FrictionLayer     string   `bson:"friction_layer" json:"friction_layer"`
	FrictionType      string   `bson:"friction_type" json:"friction_type"`
	SeveritySelf      int      `bson:"severity_self" json:"severity_self"`
	CognitiveLoadSelf int      `bson:"cognitive_load_self" json:"cognitive_load_self"`
	EmotionValence    int      `bson:"emotion_valence" json:"emotion_valence"`
	TimeLostMinutes   int      `bson:"time_lost_minutes" json:"time_lost_minutes"`
	ResumeTimeMinutes int      `bson:"resume_time_minutes" json:"resume_time_minutes"`
	RecoveryMinutes   int      `bson:"recovery_minutes" json:"recovery_minutes"`
	InterruptionCount int      `bson:"interruption_count" json:"interruption_count"`
	Tags              []string `bson:"tags,omitempty" json:"tags,omitempty"`
}
