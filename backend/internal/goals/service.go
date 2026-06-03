package goals

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

type Goal = domain.WorkGoal
type Request struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}
type Filter struct {
	Status, Cursor string
	Limit          int64
}
type Page struct {
	Goals      []Goal
	NextCursor string
}

func (s *Service) Create(ctx context.Context, req Request) (Goal, error) {
	g, err := goalFromRequest(req, Goal{}, s.clock.Now().UTC(), true)
	if err != nil {
		return Goal{}, err
	}
	if _, err := s.dispatcher.SendCommand(commands.CreateWorkGoal{Context: ctx, Goal: &g}); err != nil {
		return Goal{}, mapErr(err)
	}
	return g, nil
}
func (s *Service) List(ctx context.Context, f Filter) (Page, error) {
	fields := []serviceerror.FieldError{}
	if f.Status != "" && !ontology.IsGoalStatus(f.Status) {
		fields = append(fields, fe("status", "is invalid"))
	}
	if f.Limit < 0 {
		fields = append(fields, fe("limit", "must be >= 0"))
	}
	if err := serviceerror.NewValidation(fields); err != nil {
		return Page{}, err
	}
	result, err := s.dispatcher.SendQuery(queries.ListWorkGoals{Context: ctx, Status: f.Status, Limit: normalizeLimit(f.Limit)})
	if err != nil {
		return Page{}, mapErr(err)
	}
	goals, ok := result.([]domain.WorkGoal)
	if !ok {
		return Page{}, errors.New("work goal query returned unexpected result")
	}
	return Page{Goals: goals}, nil
}
func (s *Service) Get(ctx context.Context, id string) (Goal, error) {
	oid, err := parseID(id, "id")
	if err != nil {
		return Goal{}, err
	}
	result, err := s.dispatcher.SendQuery(queries.GetWorkGoalByID{Context: ctx, ID: oid})
	if err != nil {
		return Goal{}, mapErr(err)
	}
	g, ok := result.(*domain.WorkGoal)
	if !ok {
		return Goal{}, errors.New("work goal query returned unexpected result")
	}
	return *g, nil
}
func (s *Service) Update(ctx context.Context, id string, req Request) (Goal, error) {
	oid, err := parseID(id, "id")
	if err != nil {
		return Goal{}, err
	}
	existing, err := s.getExisting(ctx, oid)
	if err != nil {
		return Goal{}, err
	}
	g, err := goalFromRequest(req, *existing, s.clock.Now().UTC(), false)
	if err != nil {
		return Goal{}, err
	}
	g.ID = oid
	g.SchemaVersion = existing.SchemaVersion
	g.CreatedAt = existing.CreatedAt
	if _, err := s.dispatcher.SendCommand(commands.UpdateWorkGoal{Context: ctx, Goal: &g}); err != nil {
		return Goal{}, mapErr(err)
	}
	return g, nil
}
func (s *Service) Delete(ctx context.Context, id string) error {
	oid, err := parseID(id, "id")
	if err != nil {
		return err
	}
	_, err = s.dispatcher.SendCommand(commands.DeleteWorkGoal{Context: ctx, ID: oid})
	return mapErr(err)
}

func (s *Service) getExisting(ctx context.Context, id bson.ObjectID) (*domain.WorkGoal, error) {
	result, err := s.dispatcher.SendQuery(queries.GetWorkGoalByID{Context: ctx, ID: id})
	if err != nil {
		return nil, mapErr(err)
	}
	g, ok := result.(*domain.WorkGoal)
	if !ok {
		return nil, errors.New("work goal query returned unexpected result")
	}
	return g, nil
}

func goalFromRequest(req Request, base Goal, now time.Time, create bool) (Goal, error) {
	fields := []serviceerror.FieldError{}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		fields = append(fields, fe("title", "is required"))
	}
	if utf8.RuneCountInString(title) > 200 {
		fields = append(fields, fe("title", "must be at most 200 characters"))
	}
	if utf8.RuneCountInString(req.Description) > 4000 {
		fields = append(fields, fe("description", "must be at most 4000 characters"))
	}
	status := req.Status
	if status == "" {
		status = ontology.GoalStatusActive
	}
	if !ontology.IsGoalStatus(status) {
		fields = append(fields, fe("status", "is invalid"))
	}
	if err := serviceerror.NewValidation(fields); err != nil {
		return Goal{}, err
	}
	if create {
		base.SchemaVersion = domain.CurrentSchemaVersion
		base.CreatedAt = now
	}
	base.Title = title
	base.Description = req.Description
	base.Status = status
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
