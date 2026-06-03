package mongostore

import (
	"context"
	"fmt"
	"github.com/MisterVVP/logarift/backend/internal/domain"
	"github.com/MisterVVP/logarift/backend/internal/store"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"time"
)

type scoreSnapshotRepo struct{ c *mongo.Collection }

func (r *scoreSnapshotRepo) create(ctx context.Context, s *domain.ScoreSnapshot) error {
	if s == nil {
		return fmt.Errorf("%w: nil score snapshot", store.ErrInvalidInput)
	}
	prepareCreate(&s.ID, &s.SchemaVersion, &s.CreatedAt, &s.UpdatedAt)
	if err := s.Validate(); err != nil {
		return fmt.Errorf("%w: %v", store.ErrInvalidInput, err)
	}
	_, err := r.c.InsertOne(ctx, s)
	return err
}
func (r *scoreSnapshotRepo) getByID(ctx context.Context, id bson.ObjectID) (*domain.ScoreSnapshot, error) {
	return one[domain.ScoreSnapshot](ctx, r.c, id)
}
func (r *scoreSnapshotRepo) list(ctx context.Context, from time.Time, to time.Time, scoreType string, limit int64) ([]domain.ScoreSnapshot, error) {
	q := bson.M{"period_start": bson.M{"$gte": from}, "period_end": bson.M{"$lte": to}}
	if scoreType != "" {
		q["score_type"] = scoreType
	}
	return findAll[domain.ScoreSnapshot](ctx, r.c, q, bson.D{{Key: "period_start", Value: -1}}, limit)
}
func (r *scoreSnapshotRepo) delete(ctx context.Context, id bson.ObjectID) error {
	return deleteOne(ctx, r.c, id)
}
