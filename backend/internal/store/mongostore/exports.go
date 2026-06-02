package mongostore

import (
	"context"
	"fmt"
	"github.com/MisterVVP/logarift/backend/internal/domain"
	"github.com/MisterVVP/logarift/backend/internal/store"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type exportRepo struct{ c *mongo.Collection }

func (r *exportRepo) Create(ctx context.Context, e *domain.ExportRecord) error {
	if e == nil {
		return fmt.Errorf("%w: nil export", store.ErrInvalidInput)
	}
	prepareCreate(&e.ID, &e.SchemaVersion, &e.CreatedAt, &e.UpdatedAt)
	if err := e.Validate(); err != nil {
		return fmt.Errorf("%w: %v", store.ErrInvalidInput, err)
	}
	_, err := r.c.InsertOne(ctx, e)
	return err
}
func (r *exportRepo) GetByID(ctx context.Context, id bson.ObjectID) (*domain.ExportRecord, error) {
	return one[domain.ExportRecord](ctx, r.c, id)
}
func (r *exportRepo) List(ctx context.Context, status string, limit int64) ([]domain.ExportRecord, error) {
	q := bson.M{}
	if status != "" {
		q["status"] = status
	}
	return findAll[domain.ExportRecord](ctx, r.c, q, bson.D{{Key: "created_at", Value: -1}}, limit)
}
func (r *exportRepo) UpdateStatus(ctx context.Context, id bson.ObjectID, status string, filePath string) error {
	rec, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	rec.Status = status
	rec.FilePath = filePath
	touch(&rec.UpdatedAt)
	if err := rec.Validate(); err != nil {
		return fmt.Errorf("%w: %v", store.ErrInvalidInput, err)
	}
	return replaceOne(ctx, r.c, rec.ID, rec)
}
func (r *exportRepo) Delete(ctx context.Context, id bson.ObjectID) error {
	return deleteOne(ctx, r.c, id)
}
