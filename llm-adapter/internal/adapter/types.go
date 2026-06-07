package adapter

import "time"

type EnrichRequest struct {
	RequestID             string         `json:"request_id"`
	SchemaVersion         string         `json:"schema_version"`
	OccurredAt            time.Time      `json:"occurred_at"`
	Observed              ObservedInput  `json:"observed"`
	DeterministicBaseline map[string]any `json:"deterministic_baseline"`
	AllowedValues         AllowedValues  `json:"allowed_values"`
}

type ObservedInput struct {
	FrictionLevel      string               `json:"friction_level"`
	NotesMarkdown      string               `json:"notes_markdown,omitempty"`
	PlainText          string               `json:"plain_text"`
	Links              []LinkInput          `json:"links"`
	AttachmentMetadata []AttachmentMetadata `json:"attachment_metadata"`
}

type LinkInput struct {
	URL    string `json:"url"`
	Source string `json:"source,omitempty"`
}

type AttachmentMetadata struct {
	Filename    string `json:"filename,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

type AllowedValues struct {
	WorkflowStage stringSlice `json:"workflow_stage"`
	FrictionLayer stringSlice `json:"friction_layer"`
	FrictionType  stringSlice `json:"friction_type"`
}

type stringSlice []string

type EnrichResponse struct {
	SchemaVersion      string              `json:"schema_version"`
	RequestID          string              `json:"request_id"`
	TraceID            string              `json:"trace_id,omitempty"`
	AdapterVersion     string              `json:"adapter_version"`
	ModelRuntime       string              `json:"model_runtime"`
	ModelName          string              `json:"model_name"`
	ModelDigest        string              `json:"model_digest,omitempty"`
	PromptVersion      string              `json:"prompt_version"`
	DurationMS         int64               `json:"duration_ms"`
	Fields             map[string]Field    `json:"fields"`
	Warnings           []string            `json:"warnings"`
	TruncationMetadata *TruncationMetadata `json:"truncation,omitempty"`
}

type Field struct {
	Value       any     `json:"value"`
	Confidence  float64 `json:"confidence"`
	Source      string  `json:"source"`
	Explanation string  `json:"explanation,omitempty"`
}

type TruncationMetadata struct {
	Truncated              bool   `json:"truncated"`
	OriginalCharacterCount int    `json:"original_character_count"`
	RetainedCharacterCount int    `json:"retained_character_count"`
	Strategy               string `json:"strategy"`
}

type ModelInfo struct {
	AdapterVersion string `json:"adapter_version"`
	ModelRuntime   string `json:"model_runtime"`
	ModelName      string `json:"model_name"`
	ModelDigest    string `json:"model_digest,omitempty"`
	RuntimeURL     string `json:"runtime_url"`
	Available      bool   `json:"available"`
}
