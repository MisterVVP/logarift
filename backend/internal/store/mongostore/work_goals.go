package mongostore

import (
	"context"
	"fmt"
	"github.com/MisterVVP/logarift/backend/internal/domain"
	"github.com/MisterVVP/logarift/backend/internal/store"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type workGoalRepo struct{ c *mongo.Collection }

func (r *workGoalRepo) Create(ctx context.Context, g *domain.WorkGoal) error {
	if g == nil {
		return fmt.Errorf("%w: nil work goal", store.ErrInvalidInput)
	}
	prepareCreate(&g.ID, &g.SchemaVersion, &g.CreatedAt, &g.UpdatedAt)
	if err := g.Validate(); err != nil {
		return fmt.Errorf("%w: %v", store.ErrInvalidInput, err)
	}
	_, err := r.c.InsertOne(ctx, g)
	return err
}
func (r *workGoalRepo) GetByID(ctx context.Context, id bson.ObjectID) (*domain.WorkGoal, error) {
	return one[domain.WorkGoal](ctx, r.c, id)
}
func (r *workGoalRepo) List(ctx context.Context, status string, limit int64) ([]domain.WorkGoal, error) {
	q := bson.M{}
	if status != "" {
		q["status"] = status
	}
	return findAll[domain.WorkGoal](ctx, r.c, q, bson.D{{Key: "updated_at", Value: -1}}, limit)
}
func (r *workGoalRepo) Update(ctx context.Context, g *domain.WorkGoal) error {
	if g == nil {
		return fmt.Errorf("%w: nil work goal", store.ErrInvalidInput)
	}
	if g.ID.IsZero() {
		return fmt.Errorf("%w: missing id", store.ErrInvalidInput)
	}
	if g.SchemaVersion == 0 {
		g.SchemaVersion = domain.CurrentSchemaVersion
	}
	touch(&g.UpdatedAt)
	if err := g.Validate(); err != nil {
		return fmt.Errorf("%w: %v", store.ErrInvalidInput, err)
	}
	return replaceOne(ctx, r.c, g.ID, g)
}
func (r *workGoalRepo) Delete(ctx context.Context, id bson.ObjectID) error {
	return deleteOne(ctx, r.c, id)
}
