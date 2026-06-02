package domain

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type WorkGoal struct {
	ID            bson.ObjectID `bson:"_id,omitempty" json:"id"`
	SchemaVersion int           `bson:"schema_version" json:"schema_version"`
	Title         string        `bson:"title" json:"title"`
	Description   string        `bson:"description,omitempty" json:"description,omitempty"`
	Status        string        `bson:"status" json:"status"`
	CreatedAt     time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time     `bson:"updated_at" json:"updated_at"`
}
