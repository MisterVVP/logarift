package domain

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type ExportRecord struct {
	ID            bson.ObjectID `bson:"_id,omitempty" json:"id"`
	SchemaVersion int           `bson:"schema_version" json:"schema_version"`
	ExportType    string        `bson:"export_type" json:"export_type"`
	Status        string        `bson:"status" json:"status"`
	PeriodStart   *time.Time    `bson:"period_start,omitempty" json:"period_start,omitempty"`
	PeriodEnd     *time.Time    `bson:"period_end,omitempty" json:"period_end,omitempty"`
	FilePath      string        `bson:"file_path,omitempty" json:"file_path,omitempty"`
	CreatedAt     time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time     `bson:"updated_at" json:"updated_at"`
}
