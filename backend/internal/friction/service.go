package friction

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/MisterVVP/logarift/backend/internal/domain"
	"github.com/MisterVVP/logarift/backend/internal/ontology"
	"github.com/MisterVVP/logarift/backend/internal/serviceerror"
	"github.com/MisterVVP/logarift/backend/internal/store"
	"github.com/MisterVVP/logarift/backend/internal/store/commands"
	"github.com/MisterVVP/logarift/backend/internal/store/cqrs"
	"github.com/MisterVVP/logarift/backend/internal/store/queries"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const defaultLimit int64 = 50
const maxLimit int64 = 200
const maxNotesRunes = 4000
const maxTags = 20
const maxTagRunes = 64

type Clock interface{ Now() time.Time }
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now().UTC() }

type Service struct {
	dispatcher *cqrs.Dispatcher
	clock      Clock
}

func NewService(dispatcher *cqrs.Dispatcher, clock Clock) *Service {
	if clock == nil {
		clock = RealClock{}
	}
	return &Service{dispatcher: dispatcher, clock: clock}
}

type Event = domain.FrictionEvent

type Request struct {
	TimestampStart    time.Time  `json:"timestamp_start"`
	TimestampEnd      *time.Time `json:"timestamp_end"`
	WorkflowStage     string     `json:"workflow_stage"`
	FrictionLayer     string     `json:"friction_layer"`
	FrictionType      string     `json:"friction_type"`
	SeveritySelf      int        `json:"severity_self"`
	CognitiveLoadSelf int        `json:"cognitive_load_self"`
	EmotionValence    int        `json:"emotion_valence"`
	TimeLostMinutes   int        `json:"time_lost_minutes"`
	ResumeTimeMinutes int        `json:"resume_time_minutes"`
	RecoveryMinutes   int        `json:"recovery_minutes"`
	InterruptionCount int        `json:"interruption_count"`
	GoalID            string     `json:"goal_id"`
	SessionID         string     `json:"session_id"`
	Tags              []string   `json:"tags"`
	Notes             string     `json:"notes"`
	Source            string     `json:"source"`
}

type Filter struct {
	From, To                                                                      *time.Time
	WorkflowStage, FrictionLayer, FrictionType, GoalID, SessionID, Source, Cursor string
	Limit                                                                         int64
}
type Page struct {
	Events     []Event
	NextCursor string
}

func (s *Service) Create(ctx context.Context, req Request) (Event, error) {
	e, err := eventFromRequest(req, domain.FrictionEvent{}, s.clock.Now().UTC(), true)
	if err != nil {
		return Event{}, err
	}
	if _, err := s.dispatcher.SendCommand(commands.CreateFrictionEvent{Context: ctx, Event: &e}); err != nil {
		return Event{}, mapErr(err)
	}
	return e, nil
}
func (s *Service) List(ctx context.Context, filter Filter) (Page, error) {
	fields := []serviceerror.FieldError{}
	if filter.WorkflowStage != "" && !ontology.IsWorkflowStage(filter.WorkflowStage) {
		fields = append(fields, fe("workflow_stage", "is invalid"))
	}
	if filter.FrictionLayer != "" && !ontology.IsFrictionLayer(filter.FrictionLayer) {
		fields = append(fields, fe("friction_layer", "is invalid"))
	}
	if filter.FrictionType != "" && !ontology.IsFrictionType(filter.FrictionType) {
		fields = append(fields, fe("friction_type", "is invalid"))
	}
	if filter.Source != "" && !ontology.IsEventSource(filter.Source) {
		fields = append(fields, fe("source", "is invalid"))
	}
	var gid, sid *bson.ObjectID
	if filter.GoalID != "" {
		id, err := bson.ObjectIDFromHex(filter.GoalID)
		if err != nil {
			fields = append(fields, fe("goal_id", "must be a valid ObjectID"))
		} else {
			gid = &id
		}
	}
	if filter.SessionID != "" {
		id, err := bson.ObjectIDFromHex(filter.SessionID)
		if err != nil {
			fields = append(fields, fe("session_id", "must be a valid ObjectID"))
		} else {
			sid = &id
		}
	}
	if filter.Limit < 0 {
		fields = append(fields, fe("limit", "must be >= 0"))
	}
	if err := serviceerror.NewValidation(fields); err != nil {
		return Page{}, err
	}
	limit := normalizeLimit(filter.Limit)
	result, err := s.dispatcher.SendQuery(queries.ListFrictionEvents{Context: ctx, Filter: store.FrictionEventFilter{From: filter.From, To: filter.To, WorkflowStage: filter.WorkflowStage, FrictionLayer: filter.FrictionLayer, FrictionType: filter.FrictionType, GoalID: gid, SessionID: sid, Source: filter.Source, Limit: limit}})
	if err != nil {
		return Page{}, mapErr(err)
	}
	events, ok := result.([]domain.FrictionEvent)
	if !ok {
		return Page{}, errors.New("friction event query returned unexpected result")
	}
	return Page{Events: events}, nil
}
func (s *Service) Get(ctx context.Context, id string) (Event, error) {
	oid, err := parseID(id, "id")
	if err != nil {
		return Event{}, err
	}
	result, err := s.dispatcher.SendQuery(queries.GetFrictionEventByID{Context: ctx, ID: oid})
	if err != nil {
		return Event{}, mapErr(err)
	}
	e, ok := result.(*domain.FrictionEvent)
	if !ok {
		return Event{}, errors.New("friction event query returned unexpected result")
	}
	return *e, nil
}
func (s *Service) Update(ctx context.Context, id string, req Request) (Event, error) {
	oid, err := parseID(id, "id")
	if err != nil {
		return Event{}, err
	}
	existing, err := s.getExisting(ctx, oid)
	if err != nil {
		return Event{}, err
	}
	e, err := eventFromRequest(req, *existing, s.clock.Now().UTC(), false)
	if err != nil {
		return Event{}, err
	}
	e.ID = oid
	e.SchemaVersion = existing.SchemaVersion
	e.CreatedAt = existing.CreatedAt
	if _, err := s.dispatcher.SendCommand(commands.UpdateFrictionEvent{Context: ctx, Event: &e}); err != nil {
		return Event{}, mapErr(err)
	}
	return e, nil
}
func (s *Service) Delete(ctx context.Context, id string) error {
	oid, err := parseID(id, "id")
	if err != nil {
		return err
	}
	_, err = s.dispatcher.SendCommand(commands.DeleteFrictionEvent{Context: ctx, ID: oid})
	return mapErr(err)
}

func (s *Service) getExisting(ctx context.Context, id bson.ObjectID) (*domain.FrictionEvent, error) {
	result, err := s.dispatcher.SendQuery(queries.GetFrictionEventByID{Context: ctx, ID: id})
	if err != nil {
		return nil, mapErr(err)
	}
	e, ok := result.(*domain.FrictionEvent)
	if !ok {
		return nil, errors.New("friction event query returned unexpected result")
	}
	return e, nil
}

func eventFromRequest(req Request, base Event, now time.Time, create bool) (Event, error) {
	fields := []serviceerror.FieldError{}
	if req.TimestampStart.IsZero() {
		fields = append(fields, fe("timestamp_start", "is required"))
	}
	if req.TimestampEnd != nil && req.TimestampEnd.Before(req.TimestampStart) {
		fields = append(fields, fe("timestamp_end", "must not be before timestamp_start"))
	}
	if !ontology.IsWorkflowStage(req.WorkflowStage) {
		fields = append(fields, fe("workflow_stage", "is invalid"))
	}
	if !ontology.IsFrictionLayer(req.FrictionLayer) {
		fields = append(fields, fe("friction_layer", "is invalid"))
	}
	if !ontology.IsFrictionType(req.FrictionType) {
		fields = append(fields, fe("friction_type", "is invalid"))
	}
	if req.SeveritySelf < 1 || req.SeveritySelf > 5 {
		fields = append(fields, fe("severity_self", "must be between 1 and 5"))
	}
	if req.CognitiveLoadSelf < 1 || req.CognitiveLoadSelf > 5 {
		fields = append(fields, fe("cognitive_load_self", "must be between 1 and 5"))
	}
	if req.EmotionValence < -2 || req.EmotionValence > 2 {
		fields = append(fields, fe("emotion_valence", "must be between -2 and 2"))
	}
	if req.TimeLostMinutes < 0 {
		fields = append(fields, fe("time_lost_minutes", "must be >= 0"))
	}
	if req.ResumeTimeMinutes < 0 {
		fields = append(fields, fe("resume_time_minutes", "must be >= 0"))
	}
	if req.RecoveryMinutes < 0 {
		fields = append(fields, fe("recovery_minutes", "must be >= 0"))
	}
	if req.InterruptionCount < 0 {
		fields = append(fields, fe("interruption_count", "must be >= 0"))
	}
	source := req.Source
	if source == "" {
		source = ontology.SourceManual
	}
	if source != ontology.SourceManual {
		fields = append(fields, fe("source", "must be manual"))
	}
	var gid, sid *bson.ObjectID
	if strings.TrimSpace(req.GoalID) != "" {
		id, err := bson.ObjectIDFromHex(strings.TrimSpace(req.GoalID))
		if err != nil {
			fields = append(fields, fe("goal_id", "must be a valid ObjectID"))
		} else {
			gid = &id
		}
	}
	if strings.TrimSpace(req.SessionID) != "" {
		id, err := bson.ObjectIDFromHex(strings.TrimSpace(req.SessionID))
		if err != nil {
			fields = append(fields, fe("session_id", "must be a valid ObjectID"))
		} else {
			sid = &id
		}
	}
	tags, tagFields := normalizeTags(req.Tags)
	fields = append(fields, tagFields...)
	if utf8.RuneCountInString(req.Notes) > maxNotesRunes {
		fields = append(fields, fe("notes", "must be at most 4000 characters"))
	}
	if err := serviceerror.NewValidation(fields); err != nil {
		return Event{}, err
	}
	if create {
		base.SchemaVersion = domain.CurrentSchemaVersion
		base.CreatedAt = now
	}
	base.TimestampStart = req.TimestampStart.UTC()
	base.TimestampEnd = req.TimestampEnd
	if base.TimestampEnd != nil {
		t := base.TimestampEnd.UTC()
		base.TimestampEnd = &t
	}
	base.WorkflowStage = req.WorkflowStage
	base.FrictionLayer = req.FrictionLayer
	base.FrictionType = req.FrictionType
	base.SeveritySelf = req.SeveritySelf
	base.CognitiveLoadSelf = req.CognitiveLoadSelf
	base.EmotionValence = req.EmotionValence
	base.TimeLostMinutes = req.TimeLostMinutes
	base.ResumeTimeMinutes = req.ResumeTimeMinutes
	base.RecoveryMinutes = req.RecoveryMinutes
	base.InterruptionCount = req.InterruptionCount
	base.GoalID = gid
	base.SessionID = sid
	base.Tags = tags
	base.Notes = req.Notes
	base.Source = ontology.SourceManual
	base.UpdatedAt = now
	return base, nil
}
func normalizeTags(tags []string) ([]string, []serviceerror.FieldError) {
	fields := []serviceerror.FieldError{}
	seen := map[string]struct{}{}
	out := []string{}
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if utf8.RuneCountInString(tag) > maxTagRunes {
			fields = append(fields, fe("tags", "each tag must be at most 64 characters"))
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	if len(out) > maxTags {
		fields = append(fields, fe("tags", "must contain at most 20 tags"))
	}
	if len(out) > maxTags {
		out = out[:maxTags]
	}
	return out, fields
}
func normalizeLimit(v int64) int64 {
	if v <= 0 {
		return defaultLimit
	}
	if v > maxLimit {
		return maxLimit
	}
	return v
}
func parseID(value, field string) (bson.ObjectID, error) {
	id, err := bson.ObjectIDFromHex(value)
	if err != nil {
		return bson.NilObjectID, serviceerror.ValidationError{Fields: []serviceerror.FieldError{fe(field, "must be a valid ObjectID")}}
	}
	return id, nil
}
func fe(field, msg string) serviceerror.FieldError {
	return serviceerror.FieldError{Field: field, Message: msg}
}
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, store.ErrNotFound) {
		return serviceerror.ErrNotFound
	}
	if errors.Is(err, cqrs.ErrHandlerNotFound) {
		return err
	}
	if errors.Is(err, store.ErrInvalidInput) {
		return serviceerror.ValidationError{Fields: []serviceerror.FieldError{fe("request", "is invalid")}}
	}
	return err
}
