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

type workSessionRepo struct{ c *mongo.Collection }

func (r *workSessionRepo) Create(ctx context.Context, s *domain.WorkSession) error {
	if s == nil {
		return fmt.Errorf("%w: nil work session", store.ErrInvalidInput)
	}
	prepareCreate(&s.ID, &s.SchemaVersion, &s.CreatedAt, &s.UpdatedAt)
	if err := s.Validate(); err != nil {
		return fmt.Errorf("%w: %v", store.ErrInvalidInput, err)
	}
	_, err := r.c.InsertOne(ctx, s)
	return err
}
func (r *workSessionRepo) GetByID(ctx context.Context, id bson.ObjectID) (*domain.WorkSession, error) {
	return one[domain.WorkSession](ctx, r.c, id)
}
func (r *workSessionRepo) List(ctx context.Context, from *time.Time, to *time.Time, limit int64) ([]domain.WorkSession, error) {
	q := bson.M{}
	tm := bson.M{}
	if from != nil {
		tm["$gte"] = *from
	}
	if to != nil {
		tm["$lte"] = *to
	}
	if len(tm) > 0 {
		q["started_at"] = tm
	}
	return findAll[domain.WorkSession](ctx, r.c, q, bson.D{{Key: "started_at", Value: -1}}, limit)
}
func (r *workSessionRepo) Update(ctx context.Context, s *domain.WorkSession) error {
	if s == nil {
		return fmt.Errorf("%w: nil work session", store.ErrInvalidInput)
	}
	if s.ID.IsZero() {
		return fmt.Errorf("%w: missing id", store.ErrInvalidInput)
	}
	if s.SchemaVersion == 0 {
		s.SchemaVersion = domain.CurrentSchemaVersion
	}
	touch(&s.UpdatedAt)
	if err := s.Validate(); err != nil {
		return fmt.Errorf("%w: %v", store.ErrInvalidInput, err)
	}
	return replaceOne(ctx, r.c, s.ID, s)
}
func (r *workSessionRepo) Delete(ctx context.Context, id bson.ObjectID) error {
	return deleteOne(ctx, r.c, id)
}
