package mongostore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MisterVVP/logarift/backend/internal/database"
	"github.com/MisterVVP/logarift/backend/internal/domain"
	"github.com/MisterVVP/logarift/backend/internal/store"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const defaultLimit int64 = 100
const maxLimit int64 = 500

type Store struct {
	FrictionEvents store.FrictionEventRepository
	WorkGoals      store.WorkGoalRepository
	WorkSessions   store.WorkSessionRepository
	ScoreSnapshots store.ScoreSnapshotRepository
	ModelConfigs   store.ModelConfigRepository
	Exports        store.ExportRepository
	db             *mongo.Database
}

func New(client *database.Client) *Store {
	db := client.Database()
	return &Store{db: db, FrictionEvents: &frictionEventRepo{db.Collection(domain.CollectionFrictionEvents)}, WorkGoals: &workGoalRepo{db.Collection(domain.CollectionWorkGoals)}, WorkSessions: &workSessionRepo{db.Collection(domain.CollectionWorkSessions)}, ScoreSnapshots: &scoreSnapshotRepo{db.Collection(domain.CollectionScoreSnapshots)}, ModelConfigs: &modelConfigRepo{db.Collection(domain.CollectionModelConfigs)}, Exports: &exportRepo{db.Collection(domain.CollectionExports)}}
}

func (s *Store) Repositories() (store.FrictionEventRepository, store.WorkGoalRepository, store.WorkSessionRepository, store.ScoreSnapshotRepository, store.ModelConfigRepository, store.ExportRepository) {
	return s.FrictionEvents, s.WorkGoals, s.WorkSessions, s.ScoreSnapshots, s.ModelConfigs, s.Exports
}
func EnsureDefaultModelConfig(ctx context.Context, repo store.ModelConfigRepository) error {
	_, err := repo.GetDefault(ctx, domain.DefaultModelVersion)
	if err == nil {
		return nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return err
	}
	cfg := domain.DefaultModelConfig()
	return repo.Create(ctx, &cfg)
}
func prepareCreate(id *bson.ObjectID, schema *int, created, updated *time.Time) {
	now := time.Now().UTC()
	if id.IsZero() {
		*id = bson.NewObjectID()
	}
	if *schema == 0 {
		*schema = domain.CurrentSchemaVersion
	}
	if created.IsZero() {
		*created = now
	}
	if updated.IsZero() {
		*updated = now
	}
}
func touch(updated *time.Time) { *updated = time.Now().UTC() }
func normalizeLimit(limit int64) int64 {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}
func one[T any](ctx context.Context, c *mongo.Collection, id bson.ObjectID) (*T, error) {
	var out T
	if err := c.FindOne(ctx, bson.M{"_id": id}).Decode(&out); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("%w: document %s", store.ErrNotFound, id.Hex())
		}
		return nil, err
	}
	return &out, nil
}
func deleteOne(ctx context.Context, c *mongo.Collection, id bson.ObjectID) error {
	res, err := c.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("%w: document %s", store.ErrNotFound, id.Hex())
	}
	return nil
}
func replaceOne(ctx context.Context, c *mongo.Collection, id bson.ObjectID, doc any) error {
	res, err := c.ReplaceOne(ctx, bson.M{"_id": id}, doc)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("%w: document %s", store.ErrNotFound, id.Hex())
	}
	return nil
}
func findAll[T any](ctx context.Context, c *mongo.Collection, filter bson.M, sort bson.D, limit int64) ([]T, error) {
	cur, err := c.Find(ctx, filter, options.Find().SetSort(sort).SetLimit(normalizeLimit(limit)))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []T
	return out, cur.All(ctx, &out)
}
