package sessions

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/MisterVVP/logarift/backend/internal/domain"
	"github.com/MisterVVP/logarift/backend/internal/serviceerror"
	"github.com/MisterVVP/logarift/backend/internal/store"
	"github.com/MisterVVP/logarift/backend/internal/store/commands"
	"github.com/MisterVVP/logarift/backend/internal/store/cqrs"
	"github.com/MisterVVP/logarift/backend/internal/store/queries"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const defaultLimit int64 = 50
const maxLimit int64 = 200

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

type Session = domain.WorkSession
type Request struct {
	Title     string     `json:"title"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at"`
	GoalIDs   []string   `json:"goal_ids"`
	Notes     string     `json:"notes"`
}
type Filter struct {
	From, To       *time.Time
	GoalID, Cursor string
	Limit          int64
}
type Page struct {
	Sessions   []Session
	NextCursor string
}

func (s *Service) Create(ctx context.Context, req Request) (Session, error) {
	sess, err := sessionFromRequest(req, Session{}, s.clock.Now().UTC(), true)
	if err != nil {
		return Session{}, err
	}
	if _, err := s.dispatcher.SendCommand(commands.CreateWorkSession{Context: ctx, Session: &sess}); err != nil {
		return Session{}, mapErr(err)
	}
	return sess, nil
}
func (s *Service) List(ctx context.Context, f Filter) (Page, error) {
	fields := []serviceerror.FieldError{}
	var gid *bson.ObjectID
	if f.GoalID != "" {
		id, err := bson.ObjectIDFromHex(f.GoalID)
		if err != nil {
			fields = append(fields, fe("goal_id", "must be a valid ObjectID"))
		} else {
			gid = &id
		}
	}
	if f.Limit < 0 {
		fields = append(fields, fe("limit", "must be >= 0"))
	}
	if err := serviceerror.NewValidation(fields); err != nil {
		return Page{}, err
	}
	result, err := s.dispatcher.SendQuery(queries.ListWorkSessions{Context: ctx, From: f.From, To: f.To, GoalID: gid, Limit: normalizeLimit(f.Limit)})
	if err != nil {
		return Page{}, mapErr(err)
	}
	sessions, ok := result.([]domain.WorkSession)
	if !ok {
		return Page{}, errors.New("work session query returned unexpected result")
	}
	return Page{Sessions: sessions}, nil
}
func (s *Service) Get(ctx context.Context, id string) (Session, error) {
	oid, err := parseID(id, "id")
	if err != nil {
		return Session{}, err
	}
	result, err := s.dispatcher.SendQuery(queries.GetWorkSessionByID{Context: ctx, ID: oid})
	if err != nil {
		return Session{}, mapErr(err)
	}
	sess, ok := result.(*domain.WorkSession)
	if !ok {
		return Session{}, errors.New("work session query returned unexpected result")
	}
	return *sess, nil
}
func (s *Service) Update(ctx context.Context, id string, req Request) (Session, error) {
	oid, err := parseID(id, "id")
	if err != nil {
		return Session{}, err
	}
	existing, err := s.getExisting(ctx, oid)
	if err != nil {
		return Session{}, err
	}
	sess, err := sessionFromRequest(req, *existing, s.clock.Now().UTC(), false)
	if err != nil {
		return Session{}, err
	}
	sess.ID = oid
	sess.SchemaVersion = existing.SchemaVersion
	sess.CreatedAt = existing.CreatedAt
	if _, err := s.dispatcher.SendCommand(commands.UpdateWorkSession{Context: ctx, Session: &sess}); err != nil {
		return Session{}, mapErr(err)
	}
	return sess, nil
}
func (s *Service) Delete(ctx context.Context, id string) error {
	oid, err := parseID(id, "id")
	if err != nil {
		return err
	}
	_, err = s.dispatcher.SendCommand(commands.DeleteWorkSession{Context: ctx, ID: oid})
	return mapErr(err)
}

func (s *Service) getExisting(ctx context.Context, id bson.ObjectID) (*domain.WorkSession, error) {
	result, err := s.dispatcher.SendQuery(queries.GetWorkSessionByID{Context: ctx, ID: id})
	if err != nil {
		return nil, mapErr(err)
	}
	sess, ok := result.(*domain.WorkSession)
	if !ok {
		return nil, errors.New("work session query returned unexpected result")
	}
	return sess, nil
}

func sessionFromRequest(req Request, base Session, now time.Time, create bool) (Session, error) {
	fields := []serviceerror.FieldError{}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		fields = append(fields, fe("title", "is required"))
	}
	if utf8.RuneCountInString(title) > 200 {
		fields = append(fields, fe("title", "must be at most 200 characters"))
	}
	if req.StartedAt.IsZero() {
		fields = append(fields, fe("started_at", "is required"))
	}
	if req.EndedAt != nil && req.EndedAt.Before(req.StartedAt) {
		fields = append(fields, fe("ended_at", "must not be before started_at"))
	}
	if utf8.RuneCountInString(req.Notes) > 4000 {
		fields = append(fields, fe("notes", "must be at most 4000 characters"))
	}
	goalIDs := []bson.ObjectID{}
	seen := map[bson.ObjectID]struct{}{}
	for _, raw := range req.GoalIDs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		id, err := bson.ObjectIDFromHex(raw)
		if err != nil {
			fields = append(fields, fe("goal_ids", "must contain valid ObjectIDs"))
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		goalIDs = append(goalIDs, id)
	}
	if err := serviceerror.NewValidation(fields); err != nil {
		return Session{}, err
	}
	if create {
		base.SchemaVersion = domain.CurrentSchemaVersion
		base.CreatedAt = now
	}
	base.Title = title
	base.StartedAt = req.StartedAt.UTC()
	base.EndedAt = req.EndedAt
	if base.EndedAt != nil {
		t := base.EndedAt.UTC()
		base.EndedAt = &t
	}
	base.GoalIDs = goalIDs
	base.Notes = req.Notes
	base.UpdatedAt = now
	return base, nil
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
