package commands

import (
	"context"

	"github.com/MisterVVP/logarift/backend/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type IDResult struct {
	ID bson.ObjectID
}

type CreateFrictionEvent struct {
	Context context.Context
	Event   *domain.FrictionEvent
}
type UpdateFrictionEvent struct {
	Context context.Context
	Event   *domain.FrictionEvent
}
type DeleteFrictionEvent struct {
	Context context.Context
	ID      bson.ObjectID
}

type CreateLLMEnrichmentJob struct {
	Context context.Context
	Job     *domain.LLMEnrichmentJob
}
type UpdateLLMEnrichmentJob struct {
	Context context.Context
	Job     *domain.LLMEnrichmentJob
}

type CreateWorkGoal struct {
	Context context.Context
	Goal    *domain.WorkGoal
}
type UpdateWorkGoal struct {
	Context context.Context
	Goal    *domain.WorkGoal
}
type DeleteWorkGoal struct {
	Context context.Context
	ID      bson.ObjectID
}

type CreateWorkSession struct {
	Context context.Context
	Session *domain.WorkSession
}
type UpdateWorkSession struct {
	Context context.Context
	Session *domain.WorkSession
}
type DeleteWorkSession struct {
	Context context.Context
	ID      bson.ObjectID
}

type CreateScoreSnapshot struct {
	Context  context.Context
	Snapshot *domain.ScoreSnapshot
}
type DeleteScoreSnapshot struct {
	Context context.Context
	ID      bson.ObjectID
}

type CreateModelConfig struct {
	Context context.Context
	Config  *domain.ModelConfig
}
type SetDefaultModelConfig struct {
	Context context.Context
	ID      bson.ObjectID
}
type EnsureDefaultModelConfig struct{ Context context.Context }

type CreateExport struct {
	Context context.Context
	Export  *domain.ExportRecord
}
type UpdateExportStatus struct {
	Context  context.Context
	ID       bson.ObjectID
	Status   string
	FilePath string
}
type DeleteExport struct {
	Context context.Context
	ID      bson.ObjectID
}
