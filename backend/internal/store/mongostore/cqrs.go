package mongostore

import (
	"context"
	"errors"
	"fmt"

	"github.com/MisterVVP/logarift/backend/internal/domain"
	"github.com/MisterVVP/logarift/backend/internal/store"
	"github.com/MisterVVP/logarift/backend/internal/store/commands"
	"github.com/MisterVVP/logarift/backend/internal/store/cqrs"
	"github.com/MisterVVP/logarift/backend/internal/store/queries"
)

// RegisterCQRS registers this store's repository-backed command and query
// handlers on the supplied singleton dispatcher. HTTP and service packages must
// depend on the dispatcher rather than reaching into repositories directly.
func (s *Store) RegisterCQRS(dispatcher *cqrs.Dispatcher) error {
	if err := s.validateCQRSDependencies(); err != nil {
		return err
	}
	return cqrs.RegisterHandlers(
		dispatcher,
		s.createFrictionEvent,
		s.updateFrictionEvent,
		s.deleteFrictionEvent,
		s.createLLMEnrichmentJob,
		s.updateLLMEnrichmentJob,
		s.createWorkGoal,
		s.updateWorkGoal,
		s.deleteWorkGoal,
		s.createWorkSession,
		s.updateWorkSession,
		s.deleteWorkSession,
		s.createScoreSnapshot,
		s.deleteScoreSnapshot,
		s.createModelConfig,
		s.setDefaultModelConfig,
		s.ensureDefaultModelConfig,
		s.createExport,
		s.updateExportStatus,
		s.deleteExport,
		s.getFrictionEventByID,
		s.listFrictionEvents,
		s.getLLMEnrichmentJobByID,
		s.getWorkGoalByID,
		s.listWorkGoals,
		s.getWorkSessionByID,
		s.listWorkSessions,
		s.getScoreSnapshotByID,
		s.listScoreSnapshots,
		s.getModelConfigByID,
		s.getDefaultModelConfig,
		s.listModelConfigs,
		s.getExportByID,
		s.listExports,
	)
}

func (s *Store) validateCQRSDependencies() error {
	if s == nil {
		return missingCQRSDependency("store")
	}
	dependencies := []struct {
		name  string
		value any
	}{
		{name: "friction event repository", value: s.frictionEvents},
		{name: "work goal repository", value: s.workGoals},
		{name: "work session repository", value: s.workSessions},
		{name: "score snapshot repository", value: s.scoreSnapshots},
		{name: "model config repository", value: s.modelConfigs},
		{name: "export repository", value: s.exports},
		{name: "llm enrichment job repository", value: s.llmJobs},
	}
	for _, dependency := range dependencies {
		if dependency.value == nil {
			return missingCQRSDependency(dependency.name)
		}
	}
	return nil
}

func (s *Store) createFrictionEvent(ctx context.Context, command commands.CreateFrictionEvent) (commands.IDResult, error) {
	if err := s.frictionEvents.create(ctx, command.Event); err != nil {
		return commands.IDResult{}, err
	}
	return commands.IDResult{ID: command.Event.ID}, nil
}

func (s *Store) updateFrictionEvent(ctx context.Context, command commands.UpdateFrictionEvent) (cqrs.Empty, error) {
	return cqrs.Empty{}, s.frictionEvents.update(ctx, command.Event)
}

func (s *Store) deleteFrictionEvent(ctx context.Context, command commands.DeleteFrictionEvent) (cqrs.Empty, error) {
	return cqrs.Empty{}, s.frictionEvents.delete(ctx, command.ID)
}

func (s *Store) createLLMEnrichmentJob(ctx context.Context, command commands.CreateLLMEnrichmentJob) (commands.IDResult, error) {
	if err := s.llmJobs.create(ctx, command.Job); err != nil {
		return commands.IDResult{}, err
	}
	return commands.IDResult{ID: command.Job.ID}, nil
}

func (s *Store) updateLLMEnrichmentJob(ctx context.Context, command commands.UpdateLLMEnrichmentJob) (cqrs.Empty, error) {
	return cqrs.Empty{}, s.llmJobs.update(ctx, command.Job)
}

func (s *Store) createWorkGoal(ctx context.Context, command commands.CreateWorkGoal) (commands.IDResult, error) {
	if err := s.workGoals.create(ctx, command.Goal); err != nil {
		return commands.IDResult{}, err
	}
	return commands.IDResult{ID: command.Goal.ID}, nil
}

func (s *Store) updateWorkGoal(ctx context.Context, command commands.UpdateWorkGoal) (cqrs.Empty, error) {
	return cqrs.Empty{}, s.workGoals.update(ctx, command.Goal)
}

func (s *Store) deleteWorkGoal(ctx context.Context, command commands.DeleteWorkGoal) (cqrs.Empty, error) {
	return cqrs.Empty{}, s.workGoals.delete(ctx, command.ID)
}

func (s *Store) createWorkSession(ctx context.Context, command commands.CreateWorkSession) (commands.IDResult, error) {
	if err := s.workSessions.create(ctx, command.Session); err != nil {
		return commands.IDResult{}, err
	}
	return commands.IDResult{ID: command.Session.ID}, nil
}

func (s *Store) updateWorkSession(ctx context.Context, command commands.UpdateWorkSession) (cqrs.Empty, error) {
	return cqrs.Empty{}, s.workSessions.update(ctx, command.Session)
}

func (s *Store) deleteWorkSession(ctx context.Context, command commands.DeleteWorkSession) (cqrs.Empty, error) {
	return cqrs.Empty{}, s.workSessions.delete(ctx, command.ID)
}

func (s *Store) createScoreSnapshot(ctx context.Context, command commands.CreateScoreSnapshot) (commands.IDResult, error) {
	if err := s.scoreSnapshots.create(ctx, command.Snapshot); err != nil {
		return commands.IDResult{}, err
	}
	return commands.IDResult{ID: command.Snapshot.ID}, nil
}

func (s *Store) deleteScoreSnapshot(ctx context.Context, command commands.DeleteScoreSnapshot) (cqrs.Empty, error) {
	return cqrs.Empty{}, s.scoreSnapshots.delete(ctx, command.ID)
}

func (s *Store) createModelConfig(ctx context.Context, command commands.CreateModelConfig) (commands.IDResult, error) {
	if err := s.modelConfigs.create(ctx, command.Config); err != nil {
		return commands.IDResult{}, err
	}
	return commands.IDResult{ID: command.Config.ID}, nil
}

func (s *Store) setDefaultModelConfig(ctx context.Context, command commands.SetDefaultModelConfig) (cqrs.Empty, error) {
	return cqrs.Empty{}, s.modelConfigs.setDefault(ctx, command.ID)
}

func (s *Store) ensureDefaultModelConfig(ctx context.Context, command commands.EnsureDefaultModelConfig) (cqrs.Empty, error) {
	_, err := s.modelConfigs.getDefault(ctx, domain.DefaultModelVersion)
	if err == nil {
		return cqrs.Empty{}, nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return cqrs.Empty{}, err
	}
	cfg := domain.DefaultModelConfig()
	return cqrs.Empty{}, s.modelConfigs.create(ctx, &cfg)
}

func (s *Store) createExport(ctx context.Context, command commands.CreateExport) (commands.IDResult, error) {
	if err := s.exports.create(ctx, command.Export); err != nil {
		return commands.IDResult{}, err
	}
	return commands.IDResult{ID: command.Export.ID}, nil
}

func (s *Store) updateExportStatus(ctx context.Context, command commands.UpdateExportStatus) (cqrs.Empty, error) {
	return cqrs.Empty{}, s.exports.updateStatus(ctx, command.ID, command.Status, command.FilePath)
}

func (s *Store) deleteExport(ctx context.Context, command commands.DeleteExport) (cqrs.Empty, error) {
	return cqrs.Empty{}, s.exports.delete(ctx, command.ID)
}

func (s *Store) getFrictionEventByID(ctx context.Context, query queries.GetFrictionEventByID) (*domain.FrictionEvent, error) {
	return s.frictionEvents.getByID(ctx, query.ID)
}

func (s *Store) listFrictionEvents(ctx context.Context, query queries.ListFrictionEvents) ([]domain.FrictionEvent, error) {
	return s.frictionEvents.list(ctx, query.Filter)
}

func (s *Store) getLLMEnrichmentJobByID(ctx context.Context, query queries.GetLLMEnrichmentJobByID) (*domain.LLMEnrichmentJob, error) {
	return s.llmJobs.getByID(ctx, query.ID)
}

func (s *Store) getWorkGoalByID(ctx context.Context, query queries.GetWorkGoalByID) (*domain.WorkGoal, error) {
	return s.workGoals.getByID(ctx, query.ID)
}

func (s *Store) listWorkGoals(ctx context.Context, query queries.ListWorkGoals) ([]domain.WorkGoal, error) {
	return s.workGoals.list(ctx, query.Status, query.Limit)
}

func (s *Store) getWorkSessionByID(ctx context.Context, query queries.GetWorkSessionByID) (*domain.WorkSession, error) {
	return s.workSessions.getByID(ctx, query.ID)
}

func (s *Store) listWorkSessions(ctx context.Context, query queries.ListWorkSessions) ([]domain.WorkSession, error) {
	return s.workSessions.list(ctx, query.From, query.To, query.GoalID, query.Limit)
}

func (s *Store) getScoreSnapshotByID(ctx context.Context, query queries.GetScoreSnapshotByID) (*domain.ScoreSnapshot, error) {
	return s.scoreSnapshots.getByID(ctx, query.ID)
}

func (s *Store) listScoreSnapshots(ctx context.Context, query queries.ListScoreSnapshots) ([]domain.ScoreSnapshot, error) {
	return s.scoreSnapshots.list(ctx, query.From, query.To, query.ScoreType, query.Limit)
}

func (s *Store) getModelConfigByID(ctx context.Context, query queries.GetModelConfigByID) (*domain.ModelConfig, error) {
	return s.modelConfigs.getByID(ctx, query.ID)
}

func (s *Store) getDefaultModelConfig(ctx context.Context, query queries.GetDefaultModelConfig) (*domain.ModelConfig, error) {
	return s.modelConfigs.getDefault(ctx, query.ModelVersion)
}

func (s *Store) listModelConfigs(ctx context.Context, query queries.ListModelConfigs) ([]domain.ModelConfig, error) {
	return s.modelConfigs.list(ctx)
}

func (s *Store) getExportByID(ctx context.Context, query queries.GetExportByID) (*domain.ExportRecord, error) {
	return s.exports.getByID(ctx, query.ID)
}

func (s *Store) listExports(ctx context.Context, query queries.ListExports) ([]domain.ExportRecord, error) {
	return s.exports.list(ctx, query.Status, query.Limit)
}

func missingCQRSDependency(name string) error {
	return fmt.Errorf("%w: missing %s", store.ErrInvalidInput, name)
}
