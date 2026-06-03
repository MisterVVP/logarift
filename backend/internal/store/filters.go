package store

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type FrictionEventFilter struct {
	From          *time.Time
	To            *time.Time
	WorkflowStage string
	FrictionLayer string
	FrictionType  string
	GoalID        *bson.ObjectID
	SessionID     *bson.ObjectID
	Source        string
	Limit         int64
}
