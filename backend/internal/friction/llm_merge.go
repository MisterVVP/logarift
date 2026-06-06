package friction

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"strings"

	"github.com/MisterVVP/logarift/backend/internal/domain"
	"github.com/MisterVVP/logarift/backend/internal/llmadapter"
	"github.com/MisterVVP/logarift/backend/internal/ontology"
)

const hybridEngineType = "hybrid_rules_local_llm"

type LLMAdapter interface {
	Enrich(ctx context.Context, req llmadapter.Request) (llmadapter.Response, error)
}

type llmMergeOptions struct {
	minConfidence   float64
	includeMarkdown bool
}

func maybeApplyLLM(ctx context.Context, adapter LLMAdapter, event *domain.FrictionEvent, opts llmMergeOptions) {
	if adapter == nil || event == nil || event.Inference == nil {
		return
	}
	requestID := newRequestID()
	resp, err := adapter.Enrich(ctx, llmadapter.RequestFromEvent(requestID, *event, opts.includeMarkdown))
	if err != nil {
		recordAdapterFailure(event, requestID, err)
		return
	}
	mergeAdapterResponse(event, resp, opts.minConfidence)
}

func mergeAdapterResponse(event *domain.FrictionEvent, resp llmadapter.Response, minConfidence float64) {
	_ = mergeAdapterResponseWithJob(event, resp, minConfidence, "")
}

func mergeAdapterResponseWithJob(event *domain.FrictionEvent, resp llmadapter.Response, minConfidence float64, jobID string) domain.LLMEnrichmentMergeResult {
	if event.Inference == nil {
		return domain.LLMEnrichmentMergeResult{LLMStatus: domain.LLMStatusFailed, FieldDecisions: map[string]domain.LLMFieldDecision{}}
	}
	if resp.SchemaVersion != "llm-adapter-response-v1" {
		event.Inference.LocalLLM = &domain.FrictionAdapterInference{RequestID: resp.RequestID, AdapterVersion: resp.AdapterVersion, ModelRuntime: resp.ModelRuntime, ModelName: resp.ModelName, PromptVersion: resp.PromptVersion, ErrorCode: "invalid_adapter_schema", Warnings: []string{"adapter response schema was invalid; deterministic fallback used"}}
		return domain.LLMEnrichmentMergeResult{LLMStatus: domain.LLMStatusFailed, FieldDecisions: map[string]domain.LLMFieldDecision{}, AdapterResult: adapterSummary(resp), RejectedFieldCount: 1, FallbackFieldCount: 1}
	}
	adapterMeta := &domain.FrictionAdapterInference{RequestID: resp.RequestID, AdapterVersion: resp.AdapterVersion, ModelRuntime: resp.ModelRuntime, ModelName: resp.ModelName, ModelDigest: resp.ModelDigest, PromptVersion: resp.PromptVersion, DurationMS: resp.DurationMS, Warnings: resp.Warnings, AcceptedFields: map[string]domain.FrictionFieldInference{}, RejectedFields: map[string]domain.FrictionRejectedInference{}}
	decisions := map[string]domain.LLMFieldDecision{}
	for name, field := range resp.Fields {
		value, rejection := validateAdapterField(name, field)
		if rejection == "" {
			rejection = acceptanceRejection(name, field.Confidence, minConfidence)
		}
		if rejection != "" {
			adapterMeta.RejectedFields[name] = rejected(field, rejection, resp)
			decisions[name] = domain.LLMFieldDecision{AdapterValue: field.Value, AdapterConfidence: field.Confidence, CanonicalValue: canonicalValue(event, name), Decision: "rejected_fallback_used", Reason: rejection}
			continue
		}
		inference := domain.FrictionFieldInference{Value: value, Confidence: field.Confidence, Source: normalizeSource(field.Source), Explanation: field.Explanation}
		applyAcceptedField(event, name, value, inference)
		adapterMeta.AcceptedFields[name] = inference
		decisions[name] = domain.LLMFieldDecision{AdapterValue: value, AdapterConfidence: field.Confidence, CanonicalValue: canonicalValue(event, name), Decision: "accepted", Reason: "allowed_value_and_confidence_met"}
	}
	if len(adapterMeta.AcceptedFields) > 0 {
		event.Inference.EngineType = hybridEngineType
		event.Inference.EngineVersion = event.Inference.EngineVersion + "+" + resp.AdapterVersion
	}
	if len(adapterMeta.AcceptedFields) > 0 || len(adapterMeta.RejectedFields) > 0 || len(adapterMeta.Warnings) > 0 {
		event.Inference.LocalLLM = adapterMeta
	}
	refreshCanonical(event)
	status := domain.LLMStatusSucceeded
	if len(adapterMeta.AcceptedFields) == 0 && (len(adapterMeta.RejectedFields) > 0 || len(resp.Fields) == 0) {
		status = domain.LLMStatusFailed
	} else if len(adapterMeta.RejectedFields) > 0 || len(adapterMeta.Warnings) > 0 {
		status = domain.LLMStatusPartiallySucceeded
	}
	return domain.LLMEnrichmentMergeResult{LLMStatus: status, AdapterResult: adapterSummary(resp), FieldDecisions: decisions, AcceptedFieldCount: len(adapterMeta.AcceptedFields), RejectedFieldCount: len(adapterMeta.RejectedFields), FallbackFieldCount: len(adapterMeta.RejectedFields)}
}

func adapterSummary(resp llmadapter.Response) domain.LLMAdapterResultSummary {
	return domain.LLMAdapterResultSummary{AdapterVersion: resp.AdapterVersion, ModelRuntime: resp.ModelRuntime, ModelName: resp.ModelName, PromptVersion: resp.PromptVersion, DurationMS: resp.DurationMS, WarningCount: len(resp.Warnings)}
}

func canonicalValue(event *domain.FrictionEvent, name string) any {
	switch name {
	case "workflow_stage":
		return event.WorkflowStage
	case "friction_layer":
		return event.FrictionLayer
	case "friction_type":
		return event.FrictionType
	case "time_lost_minutes":
		return event.TimeLostMinutes
	case "resume_time_minutes":
		return event.ResumeTimeMinutes
	case "interruption_count":
		return event.InterruptionCount
	case "tags":
		return event.Tags
	default:
		return nil
	}
}

func validateAdapterField(name string, field llmadapter.Field) (any, string) {
	if math.IsNaN(field.Confidence) || field.Confidence < 0 || field.Confidence > 1 {
		return nil, "confidence_out_of_range"
	}
	switch name {
	case "workflow_stage":
		value, ok := field.Value.(string)
		if !ok || !ontology.IsWorkflowStage(value) {
			return nil, "invalid_workflow_stage"
		}
		return value, ""
	case "friction_layer":
		value, ok := field.Value.(string)
		if !ok || !ontology.IsFrictionLayer(value) {
			return nil, "invalid_friction_layer"
		}
		return value, ""
	case "friction_type":
		value, ok := field.Value.(string)
		if !ok || !ontology.IsFrictionType(value) {
			return nil, "invalid_friction_type"
		}
		return value, ""
	case "time_lost_minutes", "resume_time_minutes", "interruption_count":
		value, ok := numericValue(field.Value)
		if !ok || value < 0 {
			return nil, "invalid_numeric_value"
		}
		return value, ""
	case "tags":
		value, ok := stringSliceValue(field.Value)
		if !ok {
			return nil, "invalid_tags"
		}
		return normalizeTagsForLLM(value), ""
	default:
		return nil, "unknown_field"
	}
}

func acceptanceRejection(name string, confidence, minConfidence float64) string {
	threshold := minConfidence
	switch name {
	case "friction_type":
		threshold = maxFloat(threshold, 0.75)
	case "time_lost_minutes", "resume_time_minutes", "interruption_count":
		threshold = maxFloat(threshold, 0.85)
	case "workflow_stage", "friction_layer", "tags":
		threshold = maxFloat(threshold, 0.70)
	}
	if confidence < threshold {
		return fmt.Sprintf("confidence_below_threshold_%.2f", threshold)
	}
	return ""
}

func applyAcceptedField(event *domain.FrictionEvent, name string, value any, inference domain.FrictionFieldInference) {
	switch name {
	case "workflow_stage":
		event.WorkflowStage = value.(string)
	case "friction_layer":
		event.FrictionLayer = value.(string)
	case "friction_type":
		event.FrictionType = value.(string)
	case "time_lost_minutes":
		event.TimeLostMinutes = value.(int)
	case "resume_time_minutes":
		event.ResumeTimeMinutes = value.(int)
	case "interruption_count":
		event.InterruptionCount = value.(int)
	case "tags":
		event.Tags = mergeTags(event.Tags, value.([]string))
		inference.Value = event.Tags
	}
	event.Inference.Fields[name] = inference
}

func refreshCanonical(event *domain.FrictionEvent) {
	if event.Canonical == nil {
		event.Canonical = &domain.FrictionCanonical{}
	}
	event.Canonical.WorkflowStage = event.WorkflowStage
	event.Canonical.FrictionLayer = event.FrictionLayer
	event.Canonical.FrictionType = event.FrictionType
	event.Canonical.SeveritySelf = event.SeveritySelf
	event.Canonical.CognitiveLoadSelf = event.CognitiveLoadSelf
	event.Canonical.EmotionValence = event.EmotionValence
	event.Canonical.TimeLostMinutes = event.TimeLostMinutes
	event.Canonical.ResumeTimeMinutes = event.ResumeTimeMinutes
	event.Canonical.RecoveryMinutes = event.RecoveryMinutes
	event.Canonical.InterruptionCount = event.InterruptionCount
	event.Canonical.Tags = event.Tags
}

func rejected(field llmadapter.Field, reason string, resp llmadapter.Response) domain.FrictionRejectedInference {
	return domain.FrictionRejectedInference{SuggestedValue: field.Value, Confidence: field.Confidence, Source: normalizeSource(field.Source), Explanation: field.Explanation, RejectionReason: reason, AdapterVersion: resp.AdapterVersion, ModelName: resp.ModelName, PromptVersion: resp.PromptVersion}
}

func recordAdapterFailure(event *domain.FrictionEvent, requestID string, err error) {
	if event.Inference == nil {
		return
	}
	event.Inference.LocalLLM = &domain.FrictionAdapterInference{RequestID: requestID, Warnings: []string{"adapter unavailable; deterministic fallback used"}, ErrorCode: "adapter_unavailable", RejectedFields: map[string]domain.FrictionRejectedInference{"adapter_response": {RejectionReason: err.Error()}}}
}

func numericValue(value any) (int, bool) {
	switch v := value.(type) {
	case float64:
		if math.Trunc(v) != v {
			return 0, false
		}
		return int(v), true
	case int:
		return v, true
	case int64:
		return int(v), true
	default:
		return 0, false
	}
}

func stringSliceValue(value any) ([]string, bool) {
	switch v := value.(type) {
	case []string:
		return v, true
	case []any:
		out := []string{}
		for _, item := range v {
			text, ok := item.(string)
			if !ok {
				return nil, false
			}
			out = append(out, text)
		}
		return out, true
	default:
		return nil, false
	}
}

func normalizeTagsForLLM(tags []string) []string {
	out := []string{}
	for _, tag := range tags {
		tag = strings.ToLower(strings.Trim(strings.TrimSpace(tag), "#"))
		if tag != "" {
			out = append(out, tag)
		}
	}
	return out
}

func mergeTags(existing, suggested []string) []string {
	tags := domain.NormalizeTags(append(existing, suggested...))
	if len(tags) > maxTags {
		return tags[:maxTags]
	}
	return tags
}
func normalizeSource(source string) string {
	if strings.TrimSpace(source) == "" {
		return "local_llm"
	}
	return strings.TrimSpace(source)
}
func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func newRequestID() string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "llm-adapter-request"
	}
	return "llm-" + hex.EncodeToString(buf)
}

func newTraceID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "trace-local"
	}
	return hex.EncodeToString(buf)
}
