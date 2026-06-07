package llmadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/MisterVVP/logarift/backend/internal/domain"
	"github.com/MisterVVP/logarift/backend/internal/ontology"
)

const RequestSchemaVersion = "llm-adapter-request-v1"

type Client struct {
	baseURL       string
	timeout       time.Duration
	http          *http.Client
	failureMu     sync.Mutex
	failures      int
	cooldownUntil time.Time
	now           func() time.Time
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), timeout: timeout, http: &http.Client{Timeout: timeout}, now: time.Now}
}

type Request struct {
	RequestID             string         `json:"request_id"`
	SchemaVersion         string         `json:"schema_version"`
	OccurredAt            time.Time      `json:"occurred_at"`
	Observed              Observed       `json:"observed"`
	DeterministicBaseline map[string]any `json:"deterministic_baseline"`
	AllowedValues         AllowedValues  `json:"allowed_values"`
}

type Observed struct {
	FrictionLevel      string                `json:"friction_level"`
	NotesMarkdown      string                `json:"notes_markdown,omitempty"`
	PlainText          string                `json:"plain_text"`
	Links              []domain.FrictionLink `json:"links"`
	AttachmentMetadata []AttachmentMetadata  `json:"attachment_metadata"`
}

type AttachmentMetadata struct {
	Filename    string `json:"filename,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

type AllowedValues struct {
	WorkflowStage []string `json:"workflow_stage"`
	FrictionLayer []string `json:"friction_layer"`
	FrictionType  []string `json:"friction_type"`
}

type Response struct {
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

type TruncationMetadata struct {
	Truncated              bool   `json:"truncated"`
	OriginalCharacterCount int    `json:"original_character_count"`
	RetainedCharacterCount int    `json:"retained_character_count"`
	Strategy               string `json:"strategy"`
}

type Field struct {
	Value       any     `json:"value"`
	Confidence  float64 `json:"confidence"`
	Source      string  `json:"source"`
	Explanation string  `json:"explanation,omitempty"`
}

func RequestFromEvent(requestID string, event domain.FrictionEvent, includeMarkdown bool) Request {
	attachments := make([]AttachmentMetadata, 0)
	if event.Observed != nil {
		for _, item := range event.Observed.Attachments {
			attachments = append(attachments, AttachmentMetadata{Filename: item.Filename, ContentType: item.ContentType})
		}
	}
	observed := Observed{}
	if event.Observed != nil {
		observed.FrictionLevel = event.Observed.FrictionLevel
		observed.PlainText = event.Observed.PlainText
		observed.Links = event.Observed.Links
		observed.AttachmentMetadata = attachments
		if includeMarkdown {
			observed.NotesMarkdown = event.Observed.NotesMarkdown
		}
	}
	return Request{RequestID: requestID, SchemaVersion: RequestSchemaVersion, OccurredAt: event.TimestampStart.UTC(), Observed: observed, DeterministicBaseline: baseline(event), AllowedValues: AllowedValues{WorkflowStage: ontology.WorkflowStages, FrictionLayer: ontology.FrictionLayers, FrictionType: ontology.FrictionTypes}}
}

func baseline(event domain.FrictionEvent) map[string]any {
	return map[string]any{"workflow_stage": event.WorkflowStage, "friction_layer": event.FrictionLayer, "friction_type": event.FrictionType, "severity_self": event.SeveritySelf, "cognitive_load_self": event.CognitiveLoadSelf, "emotion_valence": event.EmotionValence, "time_lost_minutes": event.TimeLostMinutes, "resume_time_minutes": event.ResumeTimeMinutes, "interruption_count": event.InterruptionCount, "tags": event.Tags}
}

func (c *Client) Enrich(ctx context.Context, req Request) (Response, error) {
	if c.inCooldown() {
		return Response{}, fmt.Errorf("llm adapter cooldown is active")
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return Response{}, err
	}
	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	httpReq, err := http.NewRequestWithContext(callCtx, http.MethodPost, c.baseURL+"/v1/enrich/friction-event", bytes.NewReader(payload))
	if err != nil {
		return Response{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	metadata := metadataFromContext(ctx)
	if metadata.TraceID != "" {
		httpReq.Header.Set("traceparent", "00-"+metadata.TraceID+"-0000000000000001-01")
	}
	if metadata.RequestID != "" {
		httpReq.Header.Set("x-logarift-request-id", metadata.RequestID)
	}
	if metadata.EventID != "" {
		httpReq.Header.Set("x-logarift-event-id", metadata.EventID)
	}
	if metadata.JobID != "" {
		httpReq.Header.Set("x-logarift-job-id", metadata.JobID)
	}
	resp, err := c.http.Do(httpReq)
	if err != nil {
		c.recordFailure()
		return Response{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.recordFailure()
		limited, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return Response{}, fmt.Errorf("llm adapter HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(limited)))
	}
	var out Response
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 512*1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&out); err != nil {
		c.recordFailure()
		return Response{}, err
	}
	c.recordSuccess()
	return out, nil
}

func (c *Client) inCooldown() bool {
	c.failureMu.Lock()
	defer c.failureMu.Unlock()
	return c.now().Before(c.cooldownUntil)
}

func (c *Client) recordFailure() {
	c.failureMu.Lock()
	defer c.failureMu.Unlock()
	c.failures++
	if c.failures >= 3 {
		c.cooldownUntil = c.now().Add(30 * time.Second)
	}
}

func (c *Client) recordSuccess() {
	c.failureMu.Lock()
	defer c.failureMu.Unlock()
	c.failures = 0
	c.cooldownUntil = time.Time{}
}
