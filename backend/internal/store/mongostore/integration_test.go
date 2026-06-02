//go:build integration

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

func TestRepositoriesIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cfg := config.Config{MongoDBURI: "mongodb://localhost:27017", MongoDBDatabase: "logarift_test"}
	db, err := database.Connect(ctx, cfg)
	if err != nil {
		t.Skipf("MongoDB unavailable: %v", err)
	}
	defer db.Database().Drop(ctx)
	if err := db.EnsureIndexes(ctx); err != nil {
		t.Fatalf("EnsureIndexes: %v", err)
	}
	s := New(db)
	if err := EnsureDefaultModelConfig(ctx, s.ModelConfigs); err != nil {
		t.Fatalf("EnsureDefaultModelConfig: %v", err)
	}
	now := time.Now().UTC()
	e := domain.FrictionEvent{TimestampStart: now, WorkflowStage: "build", FrictionLayer: "technical", FrictionType: "slow_feedback", SeveritySelf: 3, CognitiveLoadSelf: 3, EmotionValence: 0, Source: "manual"}
	if err := s.FrictionEvents.Create(ctx, &e); err != nil {
		t.Fatalf("create event: %v", err)
	}
	if _, err := s.FrictionEvents.GetByID(ctx, e.ID); err != nil {
		t.Fatalf("get event: %v", err)
	}
	if got, err := s.FrictionEvents.List(ctx, store.FrictionEventFilter{}); err != nil || len(got) == 0 {
		t.Fatalf("list event len=%d err=%v", len(got), err)
	}
	e.Notes = "updated"
	if err := s.FrictionEvents.Update(ctx, &e); err != nil {
		t.Fatalf("update event: %v", err)
	}
	if err := s.FrictionEvents.Delete(ctx, e.ID); err != nil {
		t.Fatalf("delete event: %v", err)
	}
	g := domain.WorkGoal{Title: "goal", Status: "active"}
	if err := s.WorkGoals.Create(ctx, &g); err != nil {
		t.Fatalf("create goal: %v", err)
	}
	if _, err := s.WorkGoals.List(ctx, "active", 10); err != nil {
		t.Fatalf("list goals: %v", err)
	}
	g.Status = "completed"
	if err := s.WorkGoals.Update(ctx, &g); err != nil {
		t.Fatalf("update goal: %v", err)
	}
	if err := s.WorkGoals.Delete(ctx, g.ID); err != nil {
		t.Fatalf("delete goal: %v", err)
	}
	ws := domain.WorkSession{Title: "session", StartedAt: now}
	if err := s.WorkSessions.Create(ctx, &ws); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.WorkSessions.List(ctx, nil, nil, 10); err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	ws.Notes = "updated"
	if err := s.WorkSessions.Update(ctx, &ws); err != nil {
		t.Fatalf("update session: %v", err)
	}
	if err := s.WorkSessions.Delete(ctx, ws.ID); err != nil {
		t.Fatalf("delete session: %v", err)
	}
	snap := domain.ScoreSnapshot{ModelVersion: "mvp-0.1", PeriodStart: now, PeriodEnd: now, ScoreType: "daily", Scores: map[string]float64{"fcs": 1}}
	if err := s.ScoreSnapshots.Create(ctx, &snap); err != nil {
		t.Fatalf("create snapshot: %v", err)
	}
	if _, err := s.ScoreSnapshots.List(ctx, now.Add(-time.Hour), now.Add(time.Hour), "daily", 10); err != nil {
		t.Fatalf("list snapshots: %v", err)
	}
	if err := s.ScoreSnapshots.Delete(ctx, snap.ID); err != nil {
		t.Fatalf("delete snapshot: %v", err)
	}
	if _, err := s.ModelConfigs.GetDefault(ctx, domain.DefaultModelVersion); err != nil {
		t.Fatalf("get default: %v", err)
	}
	cfg2 := domain.DefaultModelConfig()
	cfg2.Name = "Alternate"
	cfg2.IsDefault = false
	if err := s.ModelConfigs.Create(ctx, &cfg2); err != nil {
		t.Fatalf("create cfg: %v", err)
	}
	if err := s.ModelConfigs.SetDefault(ctx, cfg2.ID); err != nil {
		t.Fatalf("set default: %v", err)
	}
	if _, err := s.ModelConfigs.List(ctx); err != nil {
		t.Fatalf("list configs: %v", err)
	}
	exp := domain.ExportRecord{ExportType: "json", Status: "pending"}
	if err := s.Exports.Create(ctx, &exp); err != nil {
		t.Fatalf("create export: %v", err)
	}
	if _, err := s.Exports.List(ctx, "pending", 10); err != nil {
		t.Fatalf("list exports: %v", err)
	}
	if err := s.Exports.UpdateStatus(ctx, exp.ID, "completed", "/tmp/out.json"); err != nil {
		t.Fatalf("update export: %v", err)
	}
	if err := s.Exports.Delete(ctx, exp.ID); err != nil {
		t.Fatalf("delete export: %v", err)
	}
}
