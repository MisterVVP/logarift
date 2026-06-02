package domain

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type FrictionEvent struct {
	ID                bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	SchemaVersion     int            `bson:"schema_version" json:"schema_version"`
	TimestampStart    time.Time      `bson:"timestamp_start" json:"timestamp_start"`
	TimestampEnd      *time.Time     `bson:"timestamp_end,omitempty" json:"timestamp_end,omitempty"`
	WorkflowStage     string         `bson:"workflow_stage" json:"workflow_stage"`
	FrictionLayer     string         `bson:"friction_layer" json:"friction_layer"`
	FrictionType      string         `bson:"friction_type" json:"friction_type"`
	SeveritySelf      int            `bson:"severity_self" json:"severity_self"`
	CognitiveLoadSelf int            `bson:"cognitive_load_self" json:"cognitive_load_self"`
	EmotionValence    int            `bson:"emotion_valence" json:"emotion_valence"`
	TimeLostMinutes   int            `bson:"time_lost_minutes" json:"time_lost_minutes"`
	ResumeTimeMinutes int            `bson:"resume_time_minutes" json:"resume_time_minutes"`
	RecoveryMinutes   int            `bson:"recovery_minutes" json:"recovery_minutes"`
	InterruptionCount int            `bson:"interruption_count" json:"interruption_count"`
	GoalID            *bson.ObjectID `bson:"goal_id,omitempty" json:"goal_id,omitempty"`
	SessionID         *bson.ObjectID `bson:"session_id,omitempty" json:"session_id,omitempty"`
	Tags              []string       `bson:"tags,omitempty" json:"tags,omitempty"`
	Notes             string         `bson:"notes,omitempty" json:"notes,omitempty"`
	Source            string         `bson:"source" json:"source"`
	CreatedAt         time.Time      `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time      `bson:"updated_at" json:"updated_at"`
}
