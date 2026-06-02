package domain

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type ScoreSnapshot struct {
	ID              bson.ObjectID      `bson:"_id,omitempty" json:"id"`
	SchemaVersion   int                `bson:"schema_version" json:"schema_version"`
	ModelVersion    string             `bson:"model_version" json:"model_version"`
	ModelConfigID   *bson.ObjectID     `bson:"model_config_id,omitempty" json:"model_config_id,omitempty"`
	PeriodStart     time.Time          `bson:"period_start" json:"period_start"`
	PeriodEnd       time.Time          `bson:"period_end" json:"period_end"`
	ScoreType       string             `bson:"score_type" json:"score_type"`
	Scores          map[string]float64 `bson:"scores" json:"scores"`
	EventScores     []EventScore       `bson:"event_scores,omitempty" json:"event_scores,omitempty"`
	TopContributors []TopContributor   `bson:"top_contributors,omitempty" json:"top_contributors,omitempty"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}
type EventScore struct {
	EventID bson.ObjectID `bson:"event_id" json:"event_id"`
	FCS     float64       `bson:"fcs" json:"fcs"`
}
type TopContributor struct {
	EventID bson.ObjectID `bson:"event_id" json:"event_id"`
	Reason  string        `bson:"reason" json:"reason"`
}
