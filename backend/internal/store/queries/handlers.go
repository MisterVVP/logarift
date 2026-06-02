package queries

import (
	"context"
	"fmt"

	"github.com/MisterVVP/logarift/backend/internal/domain"
	"github.com/MisterVVP/logarift/backend/internal/store"
)

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
