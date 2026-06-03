package mongostore

import (
	"context"
	"fmt"
	"github.com/MisterVVP/logarift/backend/internal/domain"
	"github.com/MisterVVP/logarift/backend/internal/store"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type modelConfigRepo struct{ c *mongo.Collection }

func (r *modelConfigRepo) create(ctx context.Context, m *domain.ModelConfig) error {
	if m == nil {
		return fmt.Errorf("%w: nil model config", store.ErrInvalidInput)
	}
	prepareCreate(&m.ID, &m.SchemaVersion, &m.CreatedAt, &m.UpdatedAt)
	if err := m.Validate(); err != nil {
		return fmt.Errorf("%w: %v", store.ErrInvalidInput, err)
	}
	if m.IsDefault {
		if err := r.unsetDefaults(ctx, m.ModelVersion); err != nil {
			return err
		}
	}
	_, err := r.c.InsertOne(ctx, m)
	return err
}
func (r *modelConfigRepo) getByID(ctx context.Context, id bson.ObjectID) (*domain.ModelConfig, error) {
	return one[domain.ModelConfig](ctx, r.c, id)
}
func (r *modelConfigRepo) getDefault(ctx context.Context, mv string) (*domain.ModelConfig, error) {
	var out domain.ModelConfig
	err := r.c.FindOne(ctx, bson.M{"model_version": mv, "is_default": true}).Decode(&out)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("%w: default model config %s", store.ErrNotFound, mv)
		}
		return nil, err
	}
	return &out, nil
}
func (r *modelConfigRepo) list(ctx context.Context) ([]domain.ModelConfig, error) {
	return findAll[domain.ModelConfig](ctx, r.c, bson.M{}, bson.D{{Key: "created_at", Value: -1}}, maxLimit)
}
func (r *modelConfigRepo) setDefault(ctx context.Context, id bson.ObjectID) error {
	cfg, err := r.getByID(ctx, id)
	if err != nil {
		return err
	}
	if err := r.unsetDefaults(ctx, cfg.ModelVersion); err != nil {
		return err
	}
	cfg.IsDefault = true
	touch(&cfg.UpdatedAt)
	return replaceOne(ctx, r.c, cfg.ID, cfg)
}
func (r *modelConfigRepo) unsetDefaults(ctx context.Context, mv string) error {
	configs, err := findAll[domain.ModelConfig](ctx, r.c, bson.M{"model_version": mv, "is_default": true}, bson.D{{Key: "created_at", Value: -1}}, maxLimit)
	if err != nil {
		return err
	}
	for i := range configs {
		configs[i].IsDefault = false
		touch(&configs[i].UpdatedAt)
		if err := replaceOne(ctx, r.c, configs[i].ID, &configs[i]); err != nil {
			return err
		}
	}
	return nil
}
