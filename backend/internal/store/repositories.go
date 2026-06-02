package store

import (
	"context"
	"github.com/MisterVVP/logarift/backend/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type FrictionEventRepository interface {
	Create(context.Context, *domain.FrictionEvent) error
	GetByID(context.Context, bson.ObjectID) (*domain.FrictionEvent, error)
	List(context.Context, FrictionEventFilter) ([]domain.FrictionEvent, error)
	Update(context.Context, *domain.FrictionEvent) error
	Delete(context.Context, bson.ObjectID) error
}
type WorkGoalRepository interface {
	Create(context.Context, *domain.WorkGoal) error
	GetByID(context.Context, bson.ObjectID) (*domain.WorkGoal, error)
	List(context.Context, string, int64) ([]domain.WorkGoal, error)
	Update(context.Context, *domain.WorkGoal) error
	Delete(context.Context, bson.ObjectID) error
}
type WorkSessionRepository interface {
	Create(context.Context, *domain.WorkSession) error
	GetByID(context.Context, bson.ObjectID) (*domain.WorkSession, error)
	List(context.Context, *time.Time, *time.Time, int64) ([]domain.WorkSession, error)
	Update(context.Context, *domain.WorkSession) error
	Delete(context.Context, bson.ObjectID) error
}
type ScoreSnapshotRepository interface {
	Create(context.Context, *domain.ScoreSnapshot) error
	GetByID(context.Context, bson.ObjectID) (*domain.ScoreSnapshot, error)
	List(context.Context, time.Time, time.Time, string, int64) ([]domain.ScoreSnapshot, error)
	Delete(context.Context, bson.ObjectID) error
}
type ModelConfigRepository interface {
	Create(context.Context, *domain.ModelConfig) error
	GetByID(context.Context, bson.ObjectID) (*domain.ModelConfig, error)
	GetDefault(context.Context, string) (*domain.ModelConfig, error)
	List(context.Context) ([]domain.ModelConfig, error)
	SetDefault(context.Context, bson.ObjectID) error
}
type ExportRepository interface {
	Create(context.Context, *domain.ExportRecord) error
	GetByID(context.Context, bson.ObjectID) (*domain.ExportRecord, error)
	List(context.Context, string, int64) ([]domain.ExportRecord, error)
	UpdateStatus(context.Context, bson.ObjectID, string, string) error
	Delete(context.Context, bson.ObjectID) error
}
