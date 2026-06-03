package goals

import (
	"context"
	"testing"
	"time"

	"github.com/MisterVVP/logarift/backend/internal/domain"
	"github.com/MisterVVP/logarift/backend/internal/store"
	"github.com/MisterVVP/logarift/backend/internal/store/commands"
	"github.com/MisterVVP/logarift/backend/internal/store/cqrs"
	"github.com/MisterVVP/logarift/backend/internal/store/queries"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type fakeRepo struct{ goal *domain.WorkGoal }

func dispatcherForRepo(t *testing.T, repo *fakeRepo) *cqrs.Dispatcher {
	t.Helper()
	d := cqrs.NewDispatcher()
	must := func(err error) {
		if err != nil {
			t.Fatalf("register handler: %v", err)
		}
	}
	must(cqrs.RegisterCommand(d, func(ctx context.Context, command commands.CreateWorkGoal) (commands.IDResult, error) {
		command.Goal.ID = bson.NewObjectID()
		cp := *command.Goal
		repo.goal = &cp
		return commands.IDResult{ID: command.Goal.ID}, nil
	}))
	must(cqrs.RegisterCommand(d, func(ctx context.Context, command commands.UpdateWorkGoal) (cqrs.Empty, error) {
		cp := *command.Goal
		repo.goal = &cp
		return cqrs.Empty{}, nil
	}))
	must(cqrs.RegisterCommand(d, func(ctx context.Context, command commands.DeleteWorkGoal) (cqrs.Empty, error) {
		repo.goal = nil
		return cqrs.Empty{}, nil
	}))
	must(cqrs.RegisterQuery(d, func(ctx context.Context, query queries.GetWorkGoalByID) (*domain.WorkGoal, error) {
		if repo.goal == nil {
			return nil, store.ErrNotFound
		}
		cp := *repo.goal
		cp.ID = query.ID
		return &cp, nil
	}))
	must(cqrs.RegisterQuery(d, func(ctx context.Context, query queries.ListWorkGoals) ([]domain.WorkGoal, error) {
		if repo.goal == nil {
			return nil, nil
		}
		return []domain.WorkGoal{*repo.goal}, nil
	}))
	return d
}

func TestCreateDefaultsActiveAndTrimsTitle(t *testing.T) {
	now := time.Date(2026, 6, 1, 8, 0, 0, 0, time.UTC)
	svc := NewService(dispatcherForRepo(t, &fakeRepo{}), fixedClock{now})
	got, err := svc.Create(context.Background(), Request{Title: "  Implement API  "})
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Implement API" || got.Status != "active" || !got.CreatedAt.Equal(now) {
		t.Fatalf("unexpected goal: %#v", got)
	}
}
