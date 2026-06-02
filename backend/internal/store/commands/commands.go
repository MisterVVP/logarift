package commands

import (
	"context"
	"errors"
	"fmt"

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

type IDResult struct {
	ID bson.ObjectID
}

type CreateFrictionEvent struct{ Event *domain.FrictionEvent }
type UpdateFrictionEvent struct{ Event *domain.FrictionEvent }
type DeleteFrictionEvent struct{ ID bson.ObjectID }

type CreateWorkGoal struct{ Goal *domain.WorkGoal }
type UpdateWorkGoal struct{ Goal *domain.WorkGoal }
type DeleteWorkGoal struct{ ID bson.ObjectID }

type CreateWorkSession struct{ Session *domain.WorkSession }
type UpdateWorkSession struct{ Session *domain.WorkSession }
type DeleteWorkSession struct{ ID bson.ObjectID }

type CreateScoreSnapshot struct{ Snapshot *domain.ScoreSnapshot }
type DeleteScoreSnapshot struct{ ID bson.ObjectID }

type CreateModelConfig struct{ Config *domain.ModelConfig }
type SetDefaultModelConfig struct{ ID bson.ObjectID }
type EnsureDefaultModelConfig struct{}

type CreateExport struct{ Export *domain.ExportRecord }
type UpdateExportStatus struct {
	ID       bson.ObjectID
	Status   string
	FilePath string
}
type DeleteExport struct{ ID bson.ObjectID }

func Register(orchestrator *cqrs.Orchestrator, handlers Handlers) error {
	registrations := []func() error{
		func() error {
			return cqrs.RegisterCommand[CreateFrictionEvent, IDResult](orchestrator, handlers.createFrictionEvent)
		},
		func() error {
			return cqrs.RegisterCommand[UpdateFrictionEvent, cqrs.Empty](orchestrator, handlers.updateFrictionEvent)
		},
		func() error {
			return cqrs.RegisterCommand[DeleteFrictionEvent, cqrs.Empty](orchestrator, handlers.deleteFrictionEvent)
		},
		func() error {
			return cqrs.RegisterCommand[CreateWorkGoal, IDResult](orchestrator, handlers.createWorkGoal)
		},
		func() error {
			return cqrs.RegisterCommand[UpdateWorkGoal, cqrs.Empty](orchestrator, handlers.updateWorkGoal)
		},
		func() error {
			return cqrs.RegisterCommand[DeleteWorkGoal, cqrs.Empty](orchestrator, handlers.deleteWorkGoal)
		},
		func() error {
			return cqrs.RegisterCommand[CreateWorkSession, IDResult](orchestrator, handlers.createWorkSession)
		},
		func() error {
			return cqrs.RegisterCommand[UpdateWorkSession, cqrs.Empty](orchestrator, handlers.updateWorkSession)
		},
		func() error {
			return cqrs.RegisterCommand[DeleteWorkSession, cqrs.Empty](orchestrator, handlers.deleteWorkSession)
		},
		func() error {
			return cqrs.RegisterCommand[CreateScoreSnapshot, IDResult](orchestrator, handlers.createScoreSnapshot)
		},
		func() error {
			return cqrs.RegisterCommand[DeleteScoreSnapshot, cqrs.Empty](orchestrator, handlers.deleteScoreSnapshot)
		},
		func() error {
			return cqrs.RegisterCommand[CreateModelConfig, IDResult](orchestrator, handlers.createModelConfig)
		},
		func() error {
			return cqrs.RegisterCommand[SetDefaultModelConfig, cqrs.Empty](orchestrator, handlers.setDefaultModelConfig)
		},
		func() error {
			return cqrs.RegisterCommand[EnsureDefaultModelConfig, cqrs.Empty](orchestrator, handlers.ensureDefaultModelConfig)
		},
		func() error { return cqrs.RegisterCommand[CreateExport, IDResult](orchestrator, handlers.createExport) },
		func() error {
			return cqrs.RegisterCommand[UpdateExportStatus, cqrs.Empty](orchestrator, handlers.updateExportStatus)
		},
		func() error {
			return cqrs.RegisterCommand[DeleteExport, cqrs.Empty](orchestrator, handlers.deleteExport)
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

func (h Handlers) createFrictionEvent(ctx context.Context, command CreateFrictionEvent) (IDResult, error) {
	if h.FrictionEvents == nil {
		return IDResult{}, missingDependency("friction event command handler")
	}
	if err := h.FrictionEvents.Create(ctx, command.Event); err != nil {
		return IDResult{}, err
	}
	return IDResult{ID: command.Event.ID}, nil
}

func (h Handlers) updateFrictionEvent(ctx context.Context, command UpdateFrictionEvent) (cqrs.Empty, error) {
	if h.FrictionEvents == nil {
		return cqrs.Empty{}, missingDependency("friction event command handler")
	}
	return cqrs.Empty{}, h.FrictionEvents.Update(ctx, command.Event)
}

func (h Handlers) deleteFrictionEvent(ctx context.Context, command DeleteFrictionEvent) (cqrs.Empty, error) {
	if h.FrictionEvents == nil {
		return cqrs.Empty{}, missingDependency("friction event command handler")
	}
	return cqrs.Empty{}, h.FrictionEvents.Delete(ctx, command.ID)
}

func (h Handlers) createWorkGoal(ctx context.Context, command CreateWorkGoal) (IDResult, error) {
	if h.WorkGoals == nil {
		return IDResult{}, missingDependency("work goal command handler")
	}
	if err := h.WorkGoals.Create(ctx, command.Goal); err != nil {
		return IDResult{}, err
	}
	return IDResult{ID: command.Goal.ID}, nil
}

func (h Handlers) updateWorkGoal(ctx context.Context, command UpdateWorkGoal) (cqrs.Empty, error) {
	if h.WorkGoals == nil {
		return cqrs.Empty{}, missingDependency("work goal command handler")
	}
	return cqrs.Empty{}, h.WorkGoals.Update(ctx, command.Goal)
}

func (h Handlers) deleteWorkGoal(ctx context.Context, command DeleteWorkGoal) (cqrs.Empty, error) {
	if h.WorkGoals == nil {
		return cqrs.Empty{}, missingDependency("work goal command handler")
	}
	return cqrs.Empty{}, h.WorkGoals.Delete(ctx, command.ID)
}

func (h Handlers) createWorkSession(ctx context.Context, command CreateWorkSession) (IDResult, error) {
	if h.WorkSessions == nil {
		return IDResult{}, missingDependency("work session command handler")
	}
	if err := h.WorkSessions.Create(ctx, command.Session); err != nil {
		return IDResult{}, err
	}
	return IDResult{ID: command.Session.ID}, nil
}

func (h Handlers) updateWorkSession(ctx context.Context, command UpdateWorkSession) (cqrs.Empty, error) {
	if h.WorkSessions == nil {
		return cqrs.Empty{}, missingDependency("work session command handler")
	}
	return cqrs.Empty{}, h.WorkSessions.Update(ctx, command.Session)
}

func (h Handlers) deleteWorkSession(ctx context.Context, command DeleteWorkSession) (cqrs.Empty, error) {
	if h.WorkSessions == nil {
		return cqrs.Empty{}, missingDependency("work session command handler")
	}
	return cqrs.Empty{}, h.WorkSessions.Delete(ctx, command.ID)
}

func (h Handlers) createScoreSnapshot(ctx context.Context, command CreateScoreSnapshot) (IDResult, error) {
	if h.ScoreSnapshots == nil {
		return IDResult{}, missingDependency("score snapshot command handler")
	}
	if err := h.ScoreSnapshots.Create(ctx, command.Snapshot); err != nil {
		return IDResult{}, err
	}
	return IDResult{ID: command.Snapshot.ID}, nil
}

func (h Handlers) deleteScoreSnapshot(ctx context.Context, command DeleteScoreSnapshot) (cqrs.Empty, error) {
	if h.ScoreSnapshots == nil {
		return cqrs.Empty{}, missingDependency("score snapshot command handler")
	}
	return cqrs.Empty{}, h.ScoreSnapshots.Delete(ctx, command.ID)
}

func (h Handlers) createModelConfig(ctx context.Context, command CreateModelConfig) (IDResult, error) {
	if h.ModelConfigs == nil {
		return IDResult{}, missingDependency("model config command handler")
	}
	if err := h.ModelConfigs.Create(ctx, command.Config); err != nil {
		return IDResult{}, err
	}
	return IDResult{ID: command.Config.ID}, nil
}

func (h Handlers) setDefaultModelConfig(ctx context.Context, command SetDefaultModelConfig) (cqrs.Empty, error) {
	if h.ModelConfigs == nil {
		return cqrs.Empty{}, missingDependency("model config command handler")
	}
	return cqrs.Empty{}, h.ModelConfigs.SetDefault(ctx, command.ID)
}

func (h Handlers) ensureDefaultModelConfig(ctx context.Context, command EnsureDefaultModelConfig) (cqrs.Empty, error) {
	if h.ModelConfigs == nil {
		return cqrs.Empty{}, missingDependency("model config command handler")
	}
	_, err := h.ModelConfigs.GetDefault(ctx, domain.DefaultModelVersion)
	if err == nil {
		return cqrs.Empty{}, nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return cqrs.Empty{}, err
	}
	cfg := domain.DefaultModelConfig()
	return cqrs.Empty{}, h.ModelConfigs.Create(ctx, &cfg)
}

func (h Handlers) createExport(ctx context.Context, command CreateExport) (IDResult, error) {
	if h.Exports == nil {
		return IDResult{}, missingDependency("export command handler")
	}
	if err := h.Exports.Create(ctx, command.Export); err != nil {
		return IDResult{}, err
	}
	return IDResult{ID: command.Export.ID}, nil
}

func (h Handlers) updateExportStatus(ctx context.Context, command UpdateExportStatus) (cqrs.Empty, error) {
	if h.Exports == nil {
		return cqrs.Empty{}, missingDependency("export command handler")
	}
	return cqrs.Empty{}, h.Exports.UpdateStatus(ctx, command.ID, command.Status, command.FilePath)
}

func (h Handlers) deleteExport(ctx context.Context, command DeleteExport) (cqrs.Empty, error) {
	if h.Exports == nil {
		return cqrs.Empty{}, missingDependency("export command handler")
	}
	return cqrs.Empty{}, h.Exports.Delete(ctx, command.ID)
}

func missingDependency(name string) error {
	return fmt.Errorf("%w: missing %s", store.ErrInvalidInput, name)
}
