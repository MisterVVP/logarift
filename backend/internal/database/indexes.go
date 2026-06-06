package database

import (
	"context"
	"fmt"
	"github.com/MisterVVP/logarift/backend/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (c *Client) EnsureIndexes(ctx context.Context) error {
	specs := map[string][]mongo.IndexModel{
		domain.CollectionFrictionEvents:    {idx("timestamp_start", 1), idx("workflow_stage", 1), idx("friction_layer", 1), idx("friction_type", 1), idx("goal_id", 1), idx("session_id", 1), idx("source", 1), idx("created_at", 1), idxD(bson.D{{Key: "timestamp_start", Value: -1}, {Key: "workflow_stage", Value: 1}}), idxD(bson.D{{Key: "timestamp_start", Value: -1}, {Key: "friction_layer", Value: 1}}), idxD(bson.D{{Key: "timestamp_start", Value: -1}, {Key: "friction_type", Value: 1}})},
		domain.CollectionWorkSessions:      {idx("started_at", 1), idx("ended_at", 1), idx("goal_ids", 1), idx("created_at", 1)},
		domain.CollectionWorkGoals:         {idx("status", 1), idx("created_at", 1), idx("updated_at", 1)},
		domain.CollectionScoreSnapshots:    {idx("period_start", 1), idx("period_end", 1), idx("score_type", 1), idx("model_version", 1), idx("created_at", 1), idxD(bson.D{{Key: "period_start", Value: 1}, {Key: "period_end", Value: 1}, {Key: "score_type", Value: 1}})},
		domain.CollectionModelConfigs:      {idx("model_version", 1), idx("is_default", 1), idx("created_at", 1), {Keys: bson.D{{Key: "model_version", Value: 1}, {Key: "is_default", Value: 1}}, Options: options.Index().SetName("model_version_default")}},
		domain.CollectionExports:           {idx("created_at", 1), idx("export_type", 1), idx("status", 1)},
		domain.CollectionLLMEnrichmentJobs: {idx("event_id", 1), idx("request_id", 1), idx("trace_id", 1), idx("status", 1), idx("created_at", 1), idx("updated_at", 1)},
	}
	for name, models := range specs {
		if _, err := c.database.Collection(name).Indexes().CreateMany(ctx, models); err != nil {
			return fmt.Errorf("create indexes for %s: %w", name, err)
		}
	}
	return nil
}
func idx(field string, order int) mongo.IndexModel {
	return mongo.IndexModel{Keys: bson.D{{Key: field, Value: order}}}
}
func idxD(keys bson.D) mongo.IndexModel { return mongo.IndexModel{Keys: keys} }
