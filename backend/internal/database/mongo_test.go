package database

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/MisterVVP/logarift/backend/internal/config"
)

const defaultMongoTestURI = "mongodb://127.0.0.1:27017/?directConnection=true"

func mongoTestURI() string {
	if uri := strings.TrimSpace(os.Getenv("LOGARIFT_TEST_MONGODB_URI")); uri != "" {
		return uri
	}
	return defaultMongoTestURI
}

func connectTestMongo(t *testing.T, databaseName string) *Client {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	client, err := Connect(ctx, config.Config{MongoDBURI: mongoTestURI(), MongoDBDatabase: databaseName})
	if err != nil {
		if strings.EqualFold(os.Getenv("LOGARIFT_REQUIRE_MONGODB_TESTS"), "true") {
			t.Fatalf("Connect() error: %v", err)
		}
		t.Skipf("MongoDB unavailable at %s: %v", mongoTestURI(), err)
	}
	return client
}

func TestConnectUsesConfiguredDatabaseName(t *testing.T) {
	client := connectTestMongo(t, "logarift_test")
	if client.DatabaseName() != "logarift_test" {
		t.Fatalf("expected database name logarift_test, got %q", client.DatabaseName())
	}
}

func TestEnsureIndexesDoesNotRequireCollectionsToExist(t *testing.T) {
	client := connectTestMongo(t, "logarift_test_indexes")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.EnsureIndexes(ctx); err != nil {
		t.Fatalf("EnsureIndexes() error: %v", err)
	}
}
