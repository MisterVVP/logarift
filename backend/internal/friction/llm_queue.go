package friction

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type LLMJobQueue interface {
	Enqueue(ctx context.Context, jobID bson.ObjectID) error
	Start(ctx context.Context, handler func(context.Context, bson.ObjectID))
}
