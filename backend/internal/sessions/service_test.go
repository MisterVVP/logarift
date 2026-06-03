package sessions

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

type fakeRepo struct{ session *domain.WorkSession }

func dispatcherForRepo(t *testing.T, repo *fakeRepo) *cqrs.Dispatcher {
	t.Helper()
	d := cqrs.NewDispatcher()
	must := func(err error) {
		if err != nil {
			t.Fatalf("register handler: %v", err)
		}
	}
	must(cqrs.RegisterCommand(d, func(ctx context.Context, command commands.CreateWorkSession) (commands.IDResult, error) {
		command.Session.ID = bson.NewObjectID()
		cp := *command.Session
		repo.session = &cp
		return commands.IDResult{ID: command.Session.ID}, nil
	}))
	must(cqrs.RegisterCommand(d, func(ctx context.Context, command commands.UpdateWorkSession) (cqrs.Empty, error) {
		cp := *command.Session
		repo.session = &cp
		return cqrs.Empty{}, nil
	}))
	must(cqrs.RegisterCommand(d, func(ctx context.Context, command commands.DeleteWorkSession) (cqrs.Empty, error) {
		repo.session = nil
		return cqrs.Empty{}, nil
	}))
	must(cqrs.RegisterQuery(d, func(ctx context.Context, query queries.GetWorkSessionByID) (*domain.WorkSession, error) {
		if repo.session == nil {
			return nil, store.ErrNotFound
		}
		cp := *repo.session
		cp.ID = query.ID
		return &cp, nil
	}))
	must(cqrs.RegisterQuery(d, func(ctx context.Context, query queries.ListWorkSessions) ([]domain.WorkSession, error) {
		if repo.session == nil {
			return nil, nil
		}
		return []domain.WorkSession{*repo.session}, nil
	}))
	return d
}

func TestCreateDeduplicatesGoalIDsAndSetsTimestamps(t *testing.T) {
	now := time.Date(2026, 6, 1, 8, 30, 0, 0, time.UTC)
	gid := bson.NewObjectID().Hex()
	svc := NewService(dispatcherForRepo(t, &fakeRepo{}), fixedClock{now})
	got, err := svc.Create(context.Background(), Request{Title: " Morning ", StartedAt: now, GoalIDs: []string{gid, gid}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Morning" || len(got.GoalIDs) != 1 || !got.UpdatedAt.Equal(now) {
		t.Fatalf("unexpected session: %#v", got)
	}
}
func TestCreateRejectsEndedBeforeStarted(t *testing.T) {
	now := time.Date(2026, 6, 1, 8, 30, 0, 0, time.UTC)
	ended := now.Add(-time.Minute)
	svc := NewService(dispatcherForRepo(t, &fakeRepo{}), fixedClock{now})
	_, err := svc.Create(context.Background(), Request{Title: "Morning", StartedAt: now, EndedAt: &ended})
	var validation serviceerror.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}
