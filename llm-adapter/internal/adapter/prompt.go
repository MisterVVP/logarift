package adapter

import (
	"encoding/json"
	"fmt"
	"strings"
)

const systemPrompt = `You enrich developer friction event notes into strict JSON fields.
Rules:
- Return JSON only. No markdown, advice, productivity judgement, coaching, or performance rating.
- Choose values only from allowed ontology values.
- If you cannot choose safely, omit the field instead of inventing a value.
- Keep explanations short, local, non-prescriptive, and based only on the provided event text and metadata.
- Use confidence between 0.0 and 1.0.`

func buildPrompt(req EnrichRequest, input string, trunc TruncationMetadata) (string, error) {
	payload := map[string]any{
		"task":                   "Return candidate friction enrichment fields as JSON matching the requested schema.",
		"allowed_values":         req.AllowedValues,
		"deterministic_baseline": req.DeterministicBaseline,
		"observed": map[string]any{
			"occurred_at":         req.OccurredAt,
			"friction_level":      req.Observed.FrictionLevel,
			"plain_text":          input,
			"links":               req.Observed.Links,
			"attachment_metadata": req.Observed.AttachmentMetadata,
			"truncation":          trunc,
		},
		"requested_json_schema": map[string]any{
			"fields": map[string]any{
				"workflow_stage":      "optional object: {value: allowed workflow_stage string, confidence: number, source: local_llm, explanation: short string}",
				"friction_layer":      "optional object: {value: allowed friction_layer string, confidence: number, source: local_llm, explanation: short string}",
				"friction_type":       "optional object: {value: allowed friction_type string, confidence: number, source: local_llm, explanation: short string}",
				"time_lost_minutes":   "optional object: {value: non-negative integer, confidence: number, source: observed_text or local_llm, explanation: short string}",
				"resume_time_minutes": "optional object: {value: non-negative integer, confidence: number, source: local_llm, explanation: short string}",
				"interruption_count":  "optional object: {value: non-negative integer, confidence: number, source: local_llm, explanation: short string}",
				"tags":                "optional object: {value: array of short lowercase strings, confidence: number, source: local_llm, explanation: short string}",
			},
			"warnings": "array of short strings",
		},
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func truncateHeadTail(text string, maxChars int) (string, TruncationMetadata) {
	runes := []rune(strings.TrimSpace(text))
	meta := TruncationMetadata{Truncated: false, OriginalCharacterCount: len(runes), RetainedCharacterCount: len(runes), Strategy: "head_tail"}
	if len(runes) <= maxChars {
		return string(runes), meta
	}
	marker := fmt.Sprintf("\n[... truncated middle of note: original_characters=%d strategy=head_tail ...]\n", len(runes))
	markerRunes := []rune(marker)
	keep := maxChars - len(markerRunes)
	if keep < 2 {
		keep = maxChars
		markerRunes = nil
	}
	head := keep / 2
	tail := keep - head
	out := string(runes[:head]) + string(markerRunes) + string(runes[len(runes)-tail:])
	meta.Truncated = true
	meta.RetainedCharacterCount = len([]rune(out))
	return out, meta
}
