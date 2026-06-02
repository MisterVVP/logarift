package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/MisterVVP/logarift/backend/internal/domain"
	"github.com/MisterVVP/logarift/backend/internal/store"
	"github.com/MisterVVP/logarift/backend/internal/store/cqrs"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Handlers struct {
	FrictionEvents store.FrictionEventRepository
	WorkGoals      store.WorkGoalRepository
	WorkSessions   store.WorkSessionRepository
	ScoreSnapshots store.ScoreSnapshotRepository
	ModelConfigs   store.ModelConfigRepository
	Exports        store.ExportRepository
}

type GetFrictionEventByID struct{ ID bson.ObjectID }
type ListFrictionEvents struct{ Filter store.FrictionEventFilter }

type GetWorkGoalByID struct{ ID bson.ObjectID }
type ListWorkGoals struct {
	Status string
	Limit  int64
}

type GetWorkSessionByID struct{ ID bson.ObjectID }
type ListWorkSessions struct {
	From  *time.Time
	To    *time.Time
	Limit int64
}

type GetScoreSnapshotByID struct{ ID bson.ObjectID }
type ListScoreSnapshots struct {
	From      time.Time
	To        time.Time
	ScoreType string
	Limit     int64
}

type GetModelConfigByID struct{ ID bson.ObjectID }
type GetDefaultModelConfig struct{ ModelVersion string }
type ListModelConfigs struct{}

type GetExportByID struct{ ID bson.ObjectID }
type ListExports struct {
	Status string
	Limit  int64
}

func Register(orchestrator *cqrs.Orchestrator, handlers Handlers) error {
	registrations := []func() error{
		func() error {
			return cqrs.RegisterQuery[GetFrictionEventByID, *domain.FrictionEvent](orchestrator, handlers.getFrictionEventByID)
		},
		func() error {
			return cqrs.RegisterQuery[ListFrictionEvents, []domain.FrictionEvent](orchestrator, handlers.listFrictionEvents)
		},
		func() error {
			return cqrs.RegisterQuery[GetWorkGoalByID, *domain.WorkGoal](orchestrator, handlers.getWorkGoalByID)
		},
		func() error {
			return cqrs.RegisterQuery[ListWorkGoals, []domain.WorkGoal](orchestrator, handlers.listWorkGoals)
		},
		func() error {
			return cqrs.RegisterQuery[GetWorkSessionByID, *domain.WorkSession](orchestrator, handlers.getWorkSessionByID)
		},
		func() error {
			return cqrs.RegisterQuery[ListWorkSessions, []domain.WorkSession](orchestrator, handlers.listWorkSessions)
		},
		func() error {
			return cqrs.RegisterQuery[GetScoreSnapshotByID, *domain.ScoreSnapshot](orchestrator, handlers.getScoreSnapshotByID)
		},
		func() error {
			return cqrs.RegisterQuery[ListScoreSnapshots, []domain.ScoreSnapshot](orchestrator, handlers.listScoreSnapshots)
		},
		func() error {
			return cqrs.RegisterQuery[GetModelConfigByID, *domain.ModelConfig](orchestrator, handlers.getModelConfigByID)
		},
		func() error {
			return cqrs.RegisterQuery[GetDefaultModelConfig, *domain.ModelConfig](orchestrator, handlers.getDefaultModelConfig)
		},
		func() error {
			return cqrs.RegisterQuery[ListModelConfigs, []domain.ModelConfig](orchestrator, handlers.listModelConfigs)
		},
		func() error {
			return cqrs.RegisterQuery[GetExportByID, *domain.ExportRecord](orchestrator, handlers.getExportByID)
		},
		func() error {
			return cqrs.RegisterQuery[ListExports, []domain.ExportRecord](orchestrator, handlers.listExports)
		},
	}
	for _, register := range registrations {
		if err := register(); err != nil {
			return err
		}
	}
	return nil
}

func FromStore(s interface {
	Repositories() (store.FrictionEventRepository, store.WorkGoalRepository, store.WorkSessionRepository, store.ScoreSnapshotRepository, store.ModelConfigRepository, store.ExportRepository)
}) Handlers {
	frictionEvents, workGoals, workSessions, scoreSnapshots, modelConfigs, exports := s.Repositories()
	return Handlers{FrictionEvents: frictionEvents, WorkGoals: workGoals, WorkSessions: workSessions, ScoreSnapshots: scoreSnapshots, ModelConfigs: modelConfigs, Exports: exports}
}

func (h Handlers) getFrictionEventByID(ctx context.Context, query GetFrictionEventByID) (*domain.FrictionEvent, error) {
	if h.FrictionEvents == nil {
		return nil, missingDependency("friction event query handler")
	}
	return h.FrictionEvents.GetByID(ctx, query.ID)
}

func (h Handlers) listFrictionEvents(ctx context.Context, query ListFrictionEvents) ([]domain.FrictionEvent, error) {
	if h.FrictionEvents == nil {
		return nil, missingDependency("friction event query handler")
	}
	return h.FrictionEvents.List(ctx, query.Filter)
}

func (h Handlers) getWorkGoalByID(ctx context.Context, query GetWorkGoalByID) (*domain.WorkGoal, error) {
	if h.WorkGoals == nil {
		return nil, missingDependency("work goal query handler")
	}
	return h.WorkGoals.GetByID(ctx, query.ID)
}

func (h Handlers) listWorkGoals(ctx context.Context, query ListWorkGoals) ([]domain.WorkGoal, error) {
	if h.WorkGoals == nil {
		return nil, missingDependency("work goal query handler")
	}
	return h.WorkGoals.List(ctx, query.Status, query.Limit)
}

func (h Handlers) getWorkSessionByID(ctx context.Context, query GetWorkSessionByID) (*domain.WorkSession, error) {
	if h.WorkSessions == nil {
		return nil, missingDependency("work session query handler")
	}
	return h.WorkSessions.GetByID(ctx, query.ID)
}

func (h Handlers) listWorkSessions(ctx context.Context, query ListWorkSessions) ([]domain.WorkSession, error) {
	if h.WorkSessions == nil {
		return nil, missingDependency("work session query handler")
	}
	return h.WorkSessions.List(ctx, query.From, query.To, query.Limit)
}

func (h Handlers) getScoreSnapshotByID(ctx context.Context, query GetScoreSnapshotByID) (*domain.ScoreSnapshot, error) {
	if h.ScoreSnapshots == nil {
		return nil, missingDependency("score snapshot query handler")
	}
	return h.ScoreSnapshots.GetByID(ctx, query.ID)
}

func (h Handlers) listScoreSnapshots(ctx context.Context, query ListScoreSnapshots) ([]domain.ScoreSnapshot, error) {
	if h.ScoreSnapshots == nil {
		return nil, missingDependency("score snapshot query handler")
	}
	return h.ScoreSnapshots.List(ctx, query.From, query.To, query.ScoreType, query.Limit)
}

func (h Handlers) getModelConfigByID(ctx context.Context, query GetModelConfigByID) (*domain.ModelConfig, error) {
	if h.ModelConfigs == nil {
		return nil, missingDependency("model config query handler")
	}
	return h.ModelConfigs.GetByID(ctx, query.ID)
}

func (h Handlers) getDefaultModelConfig(ctx context.Context, query GetDefaultModelConfig) (*domain.ModelConfig, error) {
	if h.ModelConfigs == nil {
		return nil, missingDependency("model config query handler")
	}
	return h.ModelConfigs.GetDefault(ctx, query.ModelVersion)
}

func (h Handlers) listModelConfigs(ctx context.Context, query ListModelConfigs) ([]domain.ModelConfig, error) {
	if h.ModelConfigs == nil {
		return nil, missingDependency("model config query handler")
	}
	return h.ModelConfigs.List(ctx)
}

func (h Handlers) getExportByID(ctx context.Context, query GetExportByID) (*domain.ExportRecord, error) {
	if h.Exports == nil {
		return nil, missingDependency("export query handler")
	}
	return h.Exports.GetByID(ctx, query.ID)
}

func (h Handlers) listExports(ctx context.Context, query ListExports) ([]domain.ExportRecord, error) {
	if h.Exports == nil {
		return nil, missingDependency("export query handler")
	}
	return h.Exports.List(ctx, query.Status, query.Limit)
}

func missingDependency(name string) error {
	return fmt.Errorf("%w: missing %s", store.ErrInvalidInput, name)
}
