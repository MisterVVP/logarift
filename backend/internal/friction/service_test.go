package friction

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MisterVVP/logarift/backend/internal/domain"
	"github.com/MisterVVP/logarift/backend/internal/serviceerror"
	"github.com/MisterVVP/logarift/backend/internal/store"
	"github.com/MisterVVP/logarift/backend/internal/store/commands"
	"github.com/MisterVVP/logarift/backend/internal/store/cqrs"
	"github.com/MisterVVP/logarift/backend/internal/store/queries"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type fakeRepo struct {
	event  *domain.FrictionEvent
	job    *domain.LLMEnrichmentJob
	getErr error
}

func dispatcherForRepo(t *testing.T, repo *fakeRepo) *cqrs.Dispatcher {
	t.Helper()
	d := cqrs.NewDispatcher()
	must := func(err error) {
		if err != nil {
			t.Fatalf("register handler: %v", err)
		}
	}
	must(cqrs.RegisterCommand(d, func(ctx context.Context, command commands.CreateFrictionEvent) (commands.IDResult, error) {
		command.Event.ID = bson.NewObjectID()
		cp := *command.Event
		repo.event = &cp
		return commands.IDResult{ID: command.Event.ID}, nil
	}))
	must(cqrs.RegisterCommand(d, func(ctx context.Context, command commands.UpdateFrictionEvent) (cqrs.Empty, error) {
		cp := *command.Event
		repo.event = &cp
		return cqrs.Empty{}, nil
	}))
	must(cqrs.RegisterCommand(d, func(ctx context.Context, command commands.CreateLLMEnrichmentJob) (commands.IDResult, error) {
		command.Job.ID = bson.NewObjectID()
		cp := *command.Job
		repo.job = &cp
		return commands.IDResult{ID: command.Job.ID}, nil
	}))
	must(cqrs.RegisterCommand(d, func(ctx context.Context, command commands.UpdateLLMEnrichmentJob) (cqrs.Empty, error) {
		cp := *command.Job
		repo.job = &cp
		return cqrs.Empty{}, nil
	}))
	must(cqrs.RegisterCommand(d, func(ctx context.Context, command commands.DeleteFrictionEvent) (cqrs.Empty, error) {
		if repo.getErr != nil {
			return cqrs.Empty{}, repo.getErr
		}
		repo.event = nil
		return cqrs.Empty{}, nil
	}))
	must(cqrs.RegisterQuery(d, func(ctx context.Context, query queries.GetFrictionEventByID) (*domain.FrictionEvent, error) {
		if repo.getErr != nil {
			return nil, repo.getErr
		}
		if repo.event == nil {
			return nil, store.ErrNotFound
		}
		cp := *repo.event
		cp.ID = query.ID
		return &cp, nil
	}))
	must(cqrs.RegisterQuery(d, func(ctx context.Context, query queries.GetLLMEnrichmentJobByID) (*domain.LLMEnrichmentJob, error) {
		if repo.getErr != nil {
			return nil, repo.getErr
		}
		if repo.job == nil {
			return nil, store.ErrNotFound
		}
		cp := *repo.job
		cp.ID = query.ID
		return &cp, nil
	}))
	must(cqrs.RegisterQuery(d, func(ctx context.Context, query queries.ListFrictionEvents) ([]domain.FrictionEvent, error) {
		if repo.event == nil {
			return nil, nil
		}
		return []domain.FrictionEvent{*repo.event}, nil
	}))
	return d
}

func validRequest() Request {
	return Request{TimestampStart: time.Date(2026, 6, 1, 9, 15, 0, 0, time.UTC), WorkflowStage: "test", FrictionLayer: "technical", FrictionType: "failed_feedback", SeveritySelf: 4, CognitiveLoadSelf: 3, EmotionValence: -1, TimeLostMinutes: 20, ResumeTimeMinutes: 8, InterruptionCount: 1, Tags: []string{" ci ", "ci", "flaky-test"}, Notes: "unchanged"}
}

func TestCreateDefaultsAndNormalizes(t *testing.T) {
	now := time.Date(2026, 6, 1, 9, 36, 0, 0, time.UTC)
	repo := &fakeRepo{}
	svc := NewService(dispatcherForRepo(t, repo), fixedClock{now})
	got, err := svc.Create(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if got.SchemaVersion != domain.CurrentSchemaVersion || !got.CreatedAt.Equal(now) || !got.UpdatedAt.Equal(now) {
		t.Fatalf("server fields not set: %#v", got)
	}
	if got.Source != "manual" {
		t.Fatalf("expected manual source, got %q", got.Source)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "ci" || got.Tags[1] != "flaky-test" {
		t.Fatalf("tags not normalized: %#v", got.Tags)
	}
}
func TestCreateValidationErrors(t *testing.T) {
	svc := NewService(dispatcherForRepo(t, &fakeRepo{}), fixedClock{time.Now()})
	req := validRequest()
	req.SeveritySelf = 6
	req.WorkflowStage = "bad"
	_, err := svc.Create(context.Background(), req)
	var validation serviceerror.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if len(validation.Fields) < 2 {
		t.Fatalf("expected field errors, got %#v", validation.Fields)
	}
}
func TestGetMapsNotFound(t *testing.T) {
	svc := NewService(dispatcherForRepo(t, &fakeRepo{getErr: store.ErrNotFound}), fixedClock{time.Now()})
	id := bson.NewObjectID().Hex()
	_, err := svc.Get(context.Background(), id)
	if !errors.Is(err, serviceerror.ErrNotFound) {
		t.Fatalf("expected service not found, got %v", err)
	}
}

func TestCreateQuickEnrichesAndPersistsCanonicalFields(t *testing.T) {
	now := time.Date(2026, 6, 4, 22, 0, 0, 0, time.UTC)
	repo := &fakeRepo{}
	svc := NewService(dispatcherForRepo(t, repo), fixedClock{now})
	got, err := svc.CreateQuick(context.Background(), QuickRequest{
		OccurredAt:    now.Add(-30 * time.Minute),
		FrictionLevel: "red",
		NotesMarkdown: "PR review blocked me for 1h while waiting for approval.",
	})
	if err != nil {
		t.Fatalf("CreateQuick() error: %v", err)
	}
	if got.InputMode != "quick" || got.Observed == nil || got.Inference == nil || got.Canonical == nil {
		t.Fatalf("quick metadata not populated: %#v", got)
	}
	if got.FrictionType != "waiting_for_review" || got.WorkflowStage != "code_review" {
		t.Fatalf("unexpected inference: %s %s", got.WorkflowStage, got.FrictionType)
	}
	if got.TimeLostMinutes != 60 {
		t.Fatalf("expected 60 minutes from 1h note, got %d", got.TimeLostMinutes)
	}
}
