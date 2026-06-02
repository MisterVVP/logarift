package mongostore

import (
	"context"
	"testing"
	"time"

	"github.com/MisterVVP/logarift/backend/internal/config"
	"github.com/MisterVVP/logarift/backend/internal/database"
	"github.com/MisterVVP/logarift/backend/internal/domain"
	"github.com/MisterVVP/logarift/backend/internal/store"
)

func TestFrictionEventListAppliesFiltersAndLimit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client, err := database.Connect(ctx, config.Config{MongoDBURI: "mongodb://localhost:27017", MongoDBDatabase: "unit_filters"})
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer client.Database().Drop(ctx)
	stores := New(client)

	now := time.Now().UTC()
	matching := domain.FrictionEvent{TimestampStart: now, WorkflowStage: "build", FrictionLayer: "technical", FrictionType: "slow_feedback", SeveritySelf: 3, CognitiveLoadSelf: 3, EmotionValence: 0, Source: "manual"}
	other := domain.FrictionEvent{TimestampStart: now.Add(-time.Hour), WorkflowStage: "test", FrictionLayer: "cognitive", FrictionType: "unclear_error", SeveritySelf: 3, CognitiveLoadSelf: 3, EmotionValence: 0, Source: "manual"}
	if err := stores.FrictionEvents.Create(ctx, &matching); err != nil {
		t.Fatalf("Create(matching) error: %v", err)
	}
	if err := stores.FrictionEvents.Create(ctx, &other); err != nil {
		t.Fatalf("Create(other) error: %v", err)
	}

	from := now.Add(-time.Minute)
	got, err := stores.FrictionEvents.List(ctx, store.FrictionEventFilter{From: &from, WorkflowStage: "build", FrictionLayer: "technical", FrictionType: "slow_feedback", Limit: 1})
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(got) != 1 || got[0].ID != matching.ID {
		t.Fatalf("expected only matching event, got %#v", got)
	}
}
