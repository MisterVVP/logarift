package database

import (
	"context"
	"testing"
	"time"

	"github.com/MisterVVP/logarift/backend/internal/config"
)

func TestConnectUsesConfiguredDatabaseName(t *testing.T) {
	cfg := config.Config{MongoDBURI: "mongodb://localhost:27017", MongoDBDatabase: "logarift_test"}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	client, err := Connect(ctx, cfg)
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	if client.DatabaseName() != "logarift_test" {
		t.Fatalf("expected database name logarift_test, got %q", client.DatabaseName())
	}
}
func TestEnsureIndexesDoesNotRequireCollectionsToExist(t *testing.T) {
	cfg := config.Config{MongoDBURI: "mongodb://localhost:27017", MongoDBDatabase: "logarift_test_indexes"}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	client, err := Connect(ctx, cfg)
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	if err := client.EnsureIndexes(ctx); err != nil {
		t.Fatalf("EnsureIndexes() error: %v", err)
	}
}
