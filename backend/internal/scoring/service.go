package scoring

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/MisterVVP/logarift/backend/internal/domain"
	"github.com/MisterVVP/logarift/backend/internal/serviceerror"
	"github.com/MisterVVP/logarift/backend/internal/store"
	"github.com/MisterVVP/logarift/backend/internal/store/commands"
	"github.com/MisterVVP/logarift/backend/internal/store/cqrs"
	"github.com/MisterVVP/logarift/backend/internal/store/queries"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const defaultLimit int64 = 2000
const snapshotListDefaultLimit int64 = 50
const snapshotListMaxLimit int64 = 200

type Clock interface{ Now() time.Time }
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now().UTC() }

type Service struct {
	dispatcher *cqrs.Dispatcher
	engineURL  string
	httpClient *http.Client
	clock      Clock
}

func NewService(dispatcher *cqrs.Dispatcher, engineURL string, clock Clock) *Service {
	if clock == nil {
		clock = RealClock{}
	}
	return &Service{
		dispatcher: dispatcher,
		engineURL:  strings.TrimRight(engineURL, "/"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
		clock:      clock,
	}
}

type CalculateRequest struct {
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	ScoreType   string    `json:"score_type"`
}

type Filter struct {
	From      *time.Time
	To        *time.Time
	ScoreType string
	Limit     int64
}

type Page struct {
	Snapshots  []domain.ScoreSnapshot `json:"snapshots"`
	NextCursor string                 `json:"next_cursor"`
}

type engineRequest struct {
	ModelVersion string        `json:"model_version"`
	PeriodStart  time.Time     `json:"period_start"`
	PeriodEnd    time.Time     `json:"period_end"`
	Events       []engineEvent `json:"events"`
}

type engineEvent struct {
	ID                string     `json:"id"`
	TimestampStart    time.Time  `json:"timestamp_start"`
	TimestampEnd      *time.Time `json:"timestamp_end,omitempty"`
	FrictionType      string     `json:"friction_type"`
	SeveritySelf      int        `json:"severity_self"`
	CognitiveLoadSelf int        `json:"cognitive_load_self"`
	TimeLostMinutes   int        `json:"time_lost_minutes"`
	ResumeTimeMinutes int        `json:"resume_time_minutes"`
	RecoveryMinutes   int        `json:"recovery_minutes"`
	InterruptionCount int        `json:"interruption_count"`
}

type engineResponse struct {
	ModelVersion    string                 `json:"model_version"`
	PeriodStart     time.Time              `json:"period_start"`
	PeriodEnd       time.Time              `json:"period_end"`
	Scores          map[string]float64     `json:"scores"`
	EventScores     []engineEventScore     `json:"event_scores"`
	TopContributors []engineTopContributor `json:"top_contributors"`
	Error           *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type engineEventScore struct {
	EventID string  `json:"event_id"`
	FCS     float64 `json:"fcs"`
}

type engineTopContributor struct {
	EventID string `json:"event_id"`
	Reason  string `json:"reason"`
}

func (s *Service) Calculate(ctx context.Context, req CalculateRequest) (domain.ScoreSnapshot, error) {
	if s == nil || s.dispatcher == nil {
		return domain.ScoreSnapshot{}, errors.New("scoring service is not configured")
	}
	if s.engineURL == "" {
		return domain.ScoreSnapshot{}, serviceerror.ValidationError{Fields: []serviceerror.FieldError{{Field: "math_engine_url", Message: "is not configured"}}}
	}
	fields := []serviceerror.FieldError{}
	if req.PeriodStart.IsZero() {
		fields = append(fields, serviceerror.FieldError{Field: "period_start", Message: "is required"})
	}
	if req.PeriodEnd.IsZero() {
		fields = append(fields, serviceerror.FieldError{Field: "period_end", Message: "is required"})
	}
	if !req.PeriodEnd.IsZero() && !req.PeriodStart.IsZero() && req.PeriodEnd.Before(req.PeriodStart) {
		fields = append(fields, serviceerror.FieldError{Field: "period_end", Message: "must not be before period_start"})
	}
	if req.ScoreType == "" {
		req.ScoreType = "manual"
	}
	if err := serviceerror.NewValidation(fields); err != nil {
		return domain.ScoreSnapshot{}, err
	}

	modelCfg, err := s.defaultModelConfig(ctx)
	if err != nil {
		return domain.ScoreSnapshot{}, err
	}
	events, err := s.eventsForPeriod(ctx, req.PeriodStart.UTC(), req.PeriodEnd.UTC())
	if err != nil {
		return domain.ScoreSnapshot{}, err
	}
	engineOut, err := s.runEngine(ctx, modelCfg.ModelVersion, req.PeriodStart.UTC(), req.PeriodEnd.UTC(), events)
	if err != nil {
		return domain.ScoreSnapshot{}, err
	}

	snapshot := domain.ScoreSnapshot{
		SchemaVersion:   domain.CurrentSchemaVersion,
		ModelVersion:    engineOut.ModelVersion,
		ModelConfigID:   &modelCfg.ID,
		PeriodStart:     req.PeriodStart.UTC(),
		PeriodEnd:       req.PeriodEnd.UTC(),
		ScoreType:       req.ScoreType,
		Scores:          engineOut.Scores,
		EventScores:     make([]domain.EventScore, 0, len(engineOut.EventScores)),
		TopContributors: make([]domain.TopContributor, 0, len(engineOut.TopContributors)),
		CreatedAt:       s.clock.Now().UTC(),
		UpdatedAt:       s.clock.Now().UTC(),
	}
	if snapshot.ModelVersion == "" {
		snapshot.ModelVersion = modelCfg.ModelVersion
	}
	for _, score := range engineOut.EventScores {
		id, err := bson.ObjectIDFromHex(score.EventID)
		if err != nil {
			continue
		}
		snapshot.EventScores = append(snapshot.EventScores, domain.EventScore{EventID: id, FCS: score.FCS})
	}
	for _, contributor := range engineOut.TopContributors {
		id, err := bson.ObjectIDFromHex(contributor.EventID)
		if err != nil {
			continue
		}
		snapshot.TopContributors = append(snapshot.TopContributors, domain.TopContributor{EventID: id, Reason: contributor.Reason})
	}
	if _, err := s.dispatcher.SendCommand(commands.CreateScoreSnapshot{Context: ctx, Snapshot: &snapshot}); err != nil {
		return domain.ScoreSnapshot{}, mapErr(err)
	}
	return snapshot, nil
}

func (s *Service) List(ctx context.Context, f Filter) (Page, error) {
	fields := []serviceerror.FieldError{}
	if f.Limit < 0 {
		fields = append(fields, serviceerror.FieldError{Field: "limit", Message: "must be >= 0"})
	}
	if err := serviceerror.NewValidation(fields); err != nil {
		return Page{}, err
	}
	from := time.Unix(0, 0).UTC()
	if f.From != nil {
		from = f.From.UTC()
	}
	to := s.clock.Now().UTC().AddDate(100, 0, 0)
	if f.To != nil {
		to = f.To.UTC()
	}
	limit := normalizeSnapshotLimit(f.Limit)
	result, err := s.dispatcher.SendQuery(queries.ListScoreSnapshots{Context: ctx, From: from, To: to, ScoreType: f.ScoreType, Limit: limit})
	if err != nil {
		return Page{}, mapErr(err)
	}
	snapshots, ok := result.([]domain.ScoreSnapshot)
	if !ok {
		return Page{}, errors.New("score snapshot query returned unexpected result")
	}
	return Page{Snapshots: snapshots}, nil
}

func (s *Service) Get(ctx context.Context, id string) (domain.ScoreSnapshot, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return domain.ScoreSnapshot{}, serviceerror.ValidationError{Fields: []serviceerror.FieldError{{Field: "id", Message: "must be a valid ObjectID"}}}
	}
	result, err := s.dispatcher.SendQuery(queries.GetScoreSnapshotByID{Context: ctx, ID: oid})
	if err != nil {
		return domain.ScoreSnapshot{}, mapErr(err)
	}
	snapshot, ok := result.(*domain.ScoreSnapshot)
	if !ok {
		return domain.ScoreSnapshot{}, errors.New("score snapshot query returned unexpected result")
	}
	return *snapshot, nil
}

func (s *Service) defaultModelConfig(ctx context.Context) (domain.ModelConfig, error) {
	result, err := s.dispatcher.SendQuery(queries.GetDefaultModelConfig{Context: ctx, ModelVersion: domain.DefaultModelVersion})
	if err != nil {
		return domain.ModelConfig{}, mapErr(err)
	}
	cfg, ok := result.(*domain.ModelConfig)
	if !ok {
		return domain.ModelConfig{}, errors.New("model config query returned unexpected result")
	}
	return *cfg, nil
}

func (s *Service) eventsForPeriod(ctx context.Context, start, end time.Time) ([]domain.FrictionEvent, error) {
	result, err := s.dispatcher.SendQuery(queries.ListFrictionEvents{Context: ctx, Filter: store.FrictionEventFilter{From: &start, To: &end, Limit: defaultLimit}})
	if err != nil {
		return nil, mapErr(err)
	}
	events, ok := result.([]domain.FrictionEvent)
	if !ok {
		return nil, errors.New("friction event query returned unexpected result")
	}
	return events, nil
}

func (s *Service) runEngine(ctx context.Context, modelVersion string, start, end time.Time, events []domain.FrictionEvent) (engineResponse, error) {
	payload := engineRequest{ModelVersion: modelVersion, PeriodStart: start, PeriodEnd: end, Events: make([]engineEvent, 0, len(events))}
	for _, e := range events {
		payload.Events = append(payload.Events, engineEvent{ID: e.ID.Hex(), TimestampStart: e.TimestampStart, TimestampEnd: e.TimestampEnd, FrictionType: e.FrictionType, SeveritySelf: e.SeveritySelf, CognitiveLoadSelf: e.CognitiveLoadSelf, TimeLostMinutes: e.TimeLostMinutes, ResumeTimeMinutes: e.ResumeTimeMinutes, RecoveryMinutes: e.RecoveryMinutes, InterruptionCount: e.InterruptionCount})
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return engineResponse{}, err
	}
	endpoint := strings.TrimRight(s.engineURL, "/") + "/v1/score"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return engineResponse{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	client := s.httpClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	response, err := client.Do(request)
	if err != nil {
		return engineResponse{}, fmt.Errorf("math engine request failed: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1024*1024))
	if err != nil {
		return engineResponse{}, fmt.Errorf("math engine response read failed: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return engineResponse{}, fmt.Errorf("math engine returned status %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	var out engineResponse
	if err := json.Unmarshal(responseBody, &out); err != nil {
		return engineResponse{}, fmt.Errorf("math engine returned invalid JSON: %w", err)
	}
	if out.Error != nil {
		return engineResponse{}, fmt.Errorf("math engine error %s: %s", out.Error.Code, out.Error.Message)
	}
	if out.Scores == nil {
		out.Scores = map[string]float64{}
	}
	return out, nil
}

func normalizeSnapshotLimit(v int64) int64 {
	if v <= 0 {
		return snapshotListDefaultLimit
	}
	if v > snapshotListMaxLimit {
		return snapshotListMaxLimit
	}
	return v
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, store.ErrNotFound) {
		return serviceerror.ErrNotFound
	}
	if errors.Is(err, store.ErrInvalidInput) {
		return serviceerror.ValidationError{Fields: []serviceerror.FieldError{{Field: "request", Message: "is invalid"}}}
	}
	return err
}
