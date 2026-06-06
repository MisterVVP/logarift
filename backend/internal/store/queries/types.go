package queries

import (
	"context"
	"time"

	"github.com/MisterVVP/logarift/backend/internal/store"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type GetFrictionEventByID struct {
	Context context.Context
	ID      bson.ObjectID
}
type ListFrictionEvents struct {
	Context context.Context
	Filter  store.FrictionEventFilter
}

type GetLLMEnrichmentJobByID struct {
	Context context.Context
	ID      bson.ObjectID
}

type GetWorkGoalByID struct {
	Context context.Context
	ID      bson.ObjectID
}
type ListWorkGoals struct {
	Context context.Context
	Status  string
	Limit   int64
}

type GetWorkSessionByID struct {
	Context context.Context
	ID      bson.ObjectID
}
type ListWorkSessions struct {
	Context context.Context
	From    *time.Time
	To      *time.Time
	GoalID  *bson.ObjectID
	Limit   int64
}

type GetScoreSnapshotByID struct {
	Context context.Context
	ID      bson.ObjectID
}
type ListScoreSnapshots struct {
	Context   context.Context
	From      time.Time
	To        time.Time
	ScoreType string
	Limit     int64
}

type GetModelConfigByID struct {
	Context context.Context
	ID      bson.ObjectID
}
type GetDefaultModelConfig struct {
	Context      context.Context
	ModelVersion string
}
type ListModelConfigs struct{ Context context.Context }

type GetExportByID struct {
	Context context.Context
	ID      bson.ObjectID
}
type ListExports struct {
	Context context.Context
	Status  string
	Limit   int64
}
