package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"time"
)

type Runtime interface {
	Chat(ctx context.Context, system, user string) (string, error)
	ModelInfo(ctx context.Context) (ModelInfo, error)
}

type Service struct {
	cfg     Config
	runtime Runtime
	logger  *slog.Logger
	now     func() time.Time
}

func NewService(cfg Config, runtime Runtime, logger *slog.Logger) *Service {
	if runtime == nil {
		runtime = NewOllamaClient(cfg)
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{cfg: cfg, runtime: runtime, logger: logger, now: time.Now}
}

func (s *Service) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", s.live)
	mux.HandleFunc("GET /health/ready", s.ready)
	mux.HandleFunc("GET /v1/models/current", s.modelCurrent)
	mux.HandleFunc("POST /v1/enrich/friction-event", s.enrich)
	return mux
}

func (s *Service) live(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "adapter_version": AdapterVersion})
}

func (s *Service) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.RequestTimeout)
	defer cancel()
	info, err := s.runtime.ModelInfo(ctx)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "unavailable", "error_code": "runtime_unavailable", "model": info})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "model": info})
}

func (s *Service) modelCurrent(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.RequestTimeout)
	defer cancel()
	info, err := s.runtime.ModelInfo(ctx)
	status := http.StatusOK
	if err != nil {
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, info)
}

func (s *Service) enrich(w http.ResponseWriter, r *http.Request) {
	started := s.now()
	var req EnrichRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 256*1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error_code": "invalid_json", "message": err.Error()})
		return
	}
	if err := validateRequest(req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error_code": "invalid_request", "message": err.Error()})
		return
	}
	text, trunc := truncateHeadTail(req.Observed.PlainText, s.cfg.MaxInputChars)
	prompt, err := buildPrompt(req, text, trunc)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error_code": "prompt_build_failed"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.RequestTimeout)
	defer cancel()
	content, err := s.runtime.Chat(ctx, systemPrompt, prompt)
	if err != nil {
		s.log(req.RequestID, "failed", started, 0, 0, "runtime_error")
		writeJSON(w, http.StatusBadGateway, map[string]any{"error_code": "runtime_error", "message": "local model runtime call failed"})
		return
	}
	if s.cfg.LogResponses {
		s.logger.Debug("llm response", "request_id", req.RequestID, "response", content)
	}
	resp, warnings, err := normalizeResponse(req, content, s.cfg.Model, trunc, s.now().Sub(started))
	if err != nil {
		s.log(req.RequestID, "failed", started, 0, 0, "invalid_model_json")
		writeJSON(w, http.StatusBadGateway, map[string]any{"error_code": "invalid_model_json", "message": err.Error()})
		return
	}
	resp.Warnings = append(resp.Warnings, warnings...)
	if info, err := s.runtime.ModelInfo(ctx); err == nil {
		resp.ModelDigest = info.ModelDigest
	}
	s.log(req.RequestID, "ok", started, len(resp.Fields), len(resp.Warnings), "")
	writeJSON(w, http.StatusOK, resp)
}

func validateRequest(req EnrichRequest) error {
	if req.SchemaVersion != "llm-adapter-request-v1" {
		return fmt.Errorf("schema_version must be llm-adapter-request-v1")
	}
	if strings.TrimSpace(req.RequestID) == "" {
		return errors.New("request_id is required")
	}
	if req.OccurredAt.IsZero() {
		return errors.New("occurred_at is required")
	}
	if strings.TrimSpace(req.Observed.PlainText) == "" {
		return errors.New("observed.plain_text is required")
	}
	if len(req.AllowedValues.WorkflowStage) == 0 || len(req.AllowedValues.FrictionLayer) == 0 || len(req.AllowedValues.FrictionType) == 0 {
		return errors.New("allowed_values are required")
	}
	return nil
}

func normalizeResponse(req EnrichRequest, raw, model string, trunc TruncationMetadata, duration time.Duration) (EnrichResponse, []string, error) {
	var modelOut struct {
		Fields   map[string]Field `json:"fields"`
		Warnings []string         `json:"warnings"`
	}
	if err := json.Unmarshal([]byte(raw), &modelOut); err != nil {
		return EnrichResponse{}, nil, err
	}
	allowed := allowedMaps(req.AllowedValues)
	fields := map[string]Field{}
	warnings := []string{}
	for name, field := range modelOut.Fields {
		clean, warning, ok := normalizeField(name, field, allowed)
		if warning != "" {
			warnings = append(warnings, warning)
		}
		if ok {
			fields[name] = clean
		}
	}
	return EnrichResponse{SchemaVersion: "llm-adapter-response-v1", RequestID: req.RequestID, AdapterVersion: AdapterVersion, ModelRuntime: ModelRuntime, ModelName: model, PromptVersion: PromptVersion, DurationMS: duration.Milliseconds(), Fields: fields, Warnings: append(modelOut.Warnings, warnings...), TruncationMetadata: &trunc}, nil, nil
}

func allowedMaps(values AllowedValues) map[string]map[string]struct{} {
	toMap := func(items []string) map[string]struct{} {
		out := map[string]struct{}{}
		for _, item := range items {
			out[item] = struct{}{}
		}
		return out
	}
	return map[string]map[string]struct{}{"workflow_stage": toMap(values.WorkflowStage), "friction_layer": toMap(values.FrictionLayer), "friction_type": toMap(values.FrictionType)}
}

func normalizeField(name string, field Field, allowed map[string]map[string]struct{}) (Field, string, bool) {
	known := map[string]struct{}{"workflow_stage": {}, "friction_layer": {}, "friction_type": {}, "time_lost_minutes": {}, "resume_time_minutes": {}, "interruption_count": {}, "tags": {}}
	if _, ok := known[name]; !ok {
		return Field{}, "rejected unknown field: " + name, false
	}
	if math.IsNaN(field.Confidence) || field.Confidence < 0 || field.Confidence > 1 {
		return Field{}, "rejected invalid confidence: " + name, false
	}
	field.Source = strings.TrimSpace(field.Source)
	if field.Source == "" {
		field.Source = "local_llm"
	}
	field.Explanation = limitRunes(strings.TrimSpace(field.Explanation), 240)
	if set, ontology := allowed[name]; ontology {
		value, ok := field.Value.(string)
		if !ok {
			return Field{}, "rejected non-string ontology field: " + name, false
		}
		if _, ok := set[value]; !ok {
			return Field{}, "rejected ontology value outside allowed list: " + name, false
		}
		field.Value = value
		return field, "", true
	}
	switch name {
	case "time_lost_minutes", "resume_time_minutes", "interruption_count":
		value, ok := numberToInt(field.Value)
		if !ok || value < 0 {
			return Field{}, "rejected invalid numeric field: " + name, false
		}
		field.Value = value
	case "tags":
		tags, ok := normalizeTags(field.Value)
		if !ok || len(tags) == 0 {
			return Field{}, "rejected invalid tags field", false
		}
		field.Value = tags
	}
	return field, "", true
}

func numberToInt(value any) (int, bool) {
	switch v := value.(type) {
	case float64:
		if math.Trunc(v) != v {
			return 0, false
		}
		return int(v), true
	case int:
		return v, true
	default:
		return 0, false
	}
}

func normalizeTags(value any) ([]string, bool) {
	items, ok := value.([]any)
	if !ok {
		return nil, false
	}
	seen := map[string]struct{}{}
	out := []string{}
	for _, item := range items {
		tag, ok := item.(string)
		if !ok {
			continue
		}
		tag = strings.ToLower(strings.TrimSpace(tag))
		tag = strings.Trim(tag, "#")
		if tag == "" || len([]rune(tag)) > 64 {
			continue
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
		if len(out) >= 10 {
			break
		}
	}
	return out, true
}

func limitRunes(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}

func (s *Service) log(requestID, status string, started time.Time, accepted, warnings int, errorCode string) {
	s.logger.Info("llm adapter request", "request_id", requestID, "status", status, "adapter_version", AdapterVersion, "model_runtime", ModelRuntime, "model_name", s.cfg.Model, "duration_ms", s.now().Sub(started).Milliseconds(), "accepted_field_count", accepted, "warning_count", warnings, "error_code", errorCode)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
