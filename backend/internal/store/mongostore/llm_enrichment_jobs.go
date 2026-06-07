package mongostore

import (
	"context"
	"fmt"

	"github.com/MisterVVP/logarift/backend/internal/domain"
	"github.com/MisterVVP/logarift/backend/internal/store"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type llmEnrichmentJobRepo struct{ c *mongo.Collection }

func (r *llmEnrichmentJobRepo) create(ctx context.Context, j *domain.LLMEnrichmentJob) error {
	if j == nil {
		return fmt.Errorf("%w: nil llm enrichment job", store.ErrInvalidInput)
	}
	prepareCreate(&j.ID, &j.SchemaVersion, &j.CreatedAt, &j.UpdatedAt)
	if err := j.Validate(); err != nil {
		return fmt.Errorf("%w: %v", store.ErrInvalidInput, err)
	}
	_, err := r.c.InsertOne(ctx, j)
	return err
}

func (r *llmEnrichmentJobRepo) getByID(ctx context.Context, id bson.ObjectID) (*domain.LLMEnrichmentJob, error) {
	return one[domain.LLMEnrichmentJob](ctx, r.c, id)
}

func (r *llmEnrichmentJobRepo) update(ctx context.Context, j *domain.LLMEnrichmentJob) error {
	if j == nil {
		return fmt.Errorf("%w: nil llm enrichment job", store.ErrInvalidInput)
	}
	if j.ID.IsZero() {
		return fmt.Errorf("%w: missing id", store.ErrInvalidInput)
	}
	if j.SchemaVersion == 0 {
		j.SchemaVersion = domain.CurrentSchemaVersion
	}
	touch(&j.UpdatedAt)
	if err := j.Validate(); err != nil {
		return fmt.Errorf("%w: %v", store.ErrInvalidInput, err)
	}
	return replaceOne(ctx, r.c, j.ID, j)
}
