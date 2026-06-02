package domain

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

const DefaultModelVersion = "mvp-0.1"

type ModelConfig struct {
	ID            bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	SchemaVersion int             `bson:"schema_version" json:"schema_version"`
	ModelVersion  string          `bson:"model_version" json:"model_version"`
	Name          string          `bson:"name" json:"name"`
	Parameters    ModelParameters `bson:"parameters" json:"parameters"`
	IsDefault     bool            `bson:"is_default" json:"is_default"`
	CreatedAt     time.Time       `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time       `bson:"updated_at" json:"updated_at"`
}
type ModelParameters struct {
	CLADecay                float64 `bson:"cla_decay" json:"cla_decay"`
	SeverityMultiplier      float64 `bson:"severity_multiplier" json:"severity_multiplier"`
	CognitiveLoadMultiplier float64 `bson:"cognitive_load_multiplier" json:"cognitive_load_multiplier"`
	InterruptionMultiplier  float64 `bson:"interruption_multiplier" json:"interruption_multiplier"`
	RecoveryMultiplier      float64 `bson:"recovery_multiplier" json:"recovery_multiplier"`
	FCIHalfLifeMinutes      float64 `bson:"fci_half_life_minutes" json:"fci_half_life_minutes"`
}

func DefaultModelConfig() ModelConfig {
	return ModelConfig{SchemaVersion: CurrentSchemaVersion, ModelVersion: DefaultModelVersion, Name: "Default MVP model", Parameters: ModelParameters{CLADecay: 0.85, SeverityMultiplier: 1.2, CognitiveLoadMultiplier: 1.5, InterruptionMultiplier: 2.0, RecoveryMultiplier: 0.3, FCIHalfLifeMinutes: 90}, IsDefault: true}
}
