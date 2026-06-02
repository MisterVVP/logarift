package mongostore

import (
	"context"
	"fmt"
	"github.com/MisterVVP/logarift/backend/internal/domain"
	"github.com/MisterVVP/logarift/backend/internal/store"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type frictionEventRepo struct{ c *mongo.Collection }

func (r *frictionEventRepo) Create(ctx context.Context, e *domain.FrictionEvent) error {
	if e == nil {
		return fmt.Errorf("%w: nil friction event", store.ErrInvalidInput)
	}
	prepareCreate(&e.ID, &e.SchemaVersion, &e.CreatedAt, &e.UpdatedAt)
	if err := e.Validate(); err != nil {
		return fmt.Errorf("%w: %v", store.ErrInvalidInput, err)
	}
	_, err := r.c.InsertOne(ctx, e)
	return err
}
func (r *frictionEventRepo) GetByID(ctx context.Context, id bson.ObjectID) (*domain.FrictionEvent, error) {
	return one[domain.FrictionEvent](ctx, r.c, id)
}
func (r *frictionEventRepo) List(ctx context.Context, f store.FrictionEventFilter) ([]domain.FrictionEvent, error) {
	q := bson.M{}
	tm := bson.M{}
	if f.From != nil {
		tm["$gte"] = *f.From
	}
	if f.To != nil {
		tm["$lte"] = *f.To
	}
	if len(tm) > 0 {
		q["timestamp_start"] = tm
	}
	if f.WorkflowStage != "" {
		q["workflow_stage"] = f.WorkflowStage
	}
	if f.FrictionLayer != "" {
		q["friction_layer"] = f.FrictionLayer
	}
	if f.FrictionType != "" {
		q["friction_type"] = f.FrictionType
	}
	if f.GoalID != nil {
		q["goal_id"] = *f.GoalID
	}
	if f.SessionID != nil {
		q["session_id"] = *f.SessionID
	}
	return findAll[domain.FrictionEvent](ctx, r.c, q, bson.D{{Key: "timestamp_start", Value: -1}}, f.Limit)
}
func (r *frictionEventRepo) Update(ctx context.Context, e *domain.FrictionEvent) error {
	if e == nil {
		return fmt.Errorf("%w: nil friction event", store.ErrInvalidInput)
	}
	if e.ID.IsZero() {
		return fmt.Errorf("%w: missing id", store.ErrInvalidInput)
	}
	if e.SchemaVersion == 0 {
		e.SchemaVersion = domain.CurrentSchemaVersion
	}
	touch(&e.UpdatedAt)
	if err := e.Validate(); err != nil {
		return fmt.Errorf("%w: %v", store.ErrInvalidInput, err)
	}
	return replaceOne(ctx, r.c, e.ID, e)
}
func (r *frictionEventRepo) Delete(ctx context.Context, id bson.ObjectID) error {
	return deleteOne(ctx, r.c, id)
}
