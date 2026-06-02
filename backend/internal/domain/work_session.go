package domain

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type WorkSession struct {
	ID            bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	SchemaVersion int             `bson:"schema_version" json:"schema_version"`
	Title         string          `bson:"title" json:"title"`
	StartedAt     time.Time       `bson:"started_at" json:"started_at"`
	EndedAt       *time.Time      `bson:"ended_at,omitempty" json:"ended_at,omitempty"`
	GoalIDs       []bson.ObjectID `bson:"goal_ids,omitempty" json:"goal_ids,omitempty"`
	Notes         string          `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt     time.Time       `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time       `bson:"updated_at" json:"updated_at"`
}
