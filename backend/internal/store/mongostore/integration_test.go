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
	if err := ensureDefaultModelConfig(ctx, s.modelConfigs); err != nil {
		t.Fatalf("ensureDefaultModelConfig: %v", err)
	}
	now := time.Now().UTC()
	e := domain.FrictionEvent{TimestampStart: now, WorkflowStage: "build", FrictionLayer: "technical", FrictionType: "slow_feedback", SeveritySelf: 3, CognitiveLoadSelf: 3, EmotionValence: 0, Source: "manual"}
	if err := s.frictionEvents.create(ctx, &e); err != nil {
		t.Fatalf("create event: %v", err)
	}
	if _, err := s.frictionEvents.getByID(ctx, e.ID); err != nil {
		t.Fatalf("get event: %v", err)
	}
	if got, err := s.frictionEvents.list(ctx, store.FrictionEventFilter{}); err != nil || len(got) == 0 {
		t.Fatalf("list event len=%d err=%v", len(got), err)
	}
	e.Notes = "updated"
	if err := s.frictionEvents.update(ctx, &e); err != nil {
		t.Fatalf("update event: %v", err)
	}
	if err := s.frictionEvents.delete(ctx, e.ID); err != nil {
		t.Fatalf("delete event: %v", err)
	}
	g := domain.WorkGoal{Title: "goal", Status: "active"}
	if err := s.workGoals.create(ctx, &g); err != nil {
		t.Fatalf("create goal: %v", err)
	}
	if _, err := s.workGoals.list(ctx, "active", 10); err != nil {
		t.Fatalf("list goals: %v", err)
	}
	g.Status = "completed"
	if err := s.workGoals.update(ctx, &g); err != nil {
		t.Fatalf("update goal: %v", err)
	}
	if err := s.workGoals.delete(ctx, g.ID); err != nil {
		t.Fatalf("delete goal: %v", err)
	}
	ws := domain.WorkSession{Title: "session", StartedAt: now}
	if err := s.workSessions.create(ctx, &ws); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.workSessions.list(ctx, nil, nil, nil, 10); err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	ws.Notes = "updated"
	if err := s.workSessions.update(ctx, &ws); err != nil {
		t.Fatalf("update session: %v", err)
	}
	if err := s.workSessions.delete(ctx, ws.ID); err != nil {
		t.Fatalf("delete session: %v", err)
	}
	snap := domain.ScoreSnapshot{ModelVersion: "mvp-0.1", PeriodStart: now, PeriodEnd: now, ScoreType: "daily", Scores: map[string]float64{"fcs": 1}}
	if err := s.scoreSnapshots.create(ctx, &snap); err != nil {
		t.Fatalf("create snapshot: %v", err)
	}
	if _, err := s.scoreSnapshots.list(ctx, now.Add(-time.Hour), now.Add(time.Hour), "daily", 10); err != nil {
		t.Fatalf("list snapshots: %v", err)
	}
	if err := s.scoreSnapshots.delete(ctx, snap.ID); err != nil {
		t.Fatalf("delete snapshot: %v", err)
	}
	if _, err := s.modelConfigs.getDefault(ctx, domain.DefaultModelVersion); err != nil {
		t.Fatalf("get default: %v", err)
	}
	cfg2 := domain.DefaultModelConfig()
	cfg2.Name = "Alternate"
	cfg2.IsDefault = false
	if err := s.modelConfigs.create(ctx, &cfg2); err != nil {
		t.Fatalf("create cfg: %v", err)
	}
	if err := s.modelConfigs.setDefault(ctx, cfg2.ID); err != nil {
		t.Fatalf("set default: %v", err)
	}
	if _, err := s.modelConfigs.list(ctx); err != nil {
		t.Fatalf("list configs: %v", err)
	}
	exp := domain.ExportRecord{ExportType: "json", Status: "pending"}
	if err := s.exports.create(ctx, &exp); err != nil {
		t.Fatalf("create export: %v", err)
	}
	if _, err := s.exports.list(ctx, "pending", 10); err != nil {
		t.Fatalf("list exports: %v", err)
	}
	if err := s.exports.updateStatus(ctx, exp.ID, "completed", "/tmp/out.json"); err != nil {
		t.Fatalf("update export: %v", err)
	}
	if err := s.exports.delete(ctx, exp.ID); err != nil {
		t.Fatalf("delete export: %v", err)
	}
}
