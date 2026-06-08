package testsupport

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/MisterVVP/logarift/backend/internal/config"
	"github.com/MisterVVP/logarift/backend/internal/database"
)

const defaultMongoTestURI = "mongodb://127.0.0.1:27017/?directConnection=true"

// MongoURI returns the MongoDB URI used by tests that exercise the real Mongo
// store. The default uses IPv4 and directConnection to avoid CI localhost/IPv6
// resolution surprises with a single test container.
func MongoURI() string {
	if uri := strings.TrimSpace(os.Getenv("LOGARIFT_TEST_MONGODB_URI")); uri != "" {
		return uri
	}
	return defaultMongoTestURI
}

// ConnectMongo connects to MongoDB for tests. When MongoDB is not available,
// local runs skip these integration-like tests by default. CI can set
// LOGARIFT_REQUIRE_MONGODB_TESTS=true to fail fast if its MongoDB service is not
// reachable.
func ConnectMongo(t *testing.T, databaseName string) *database.Client {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)

	client, err := database.Connect(ctx, config.Config{MongoDBURI: MongoURI(), MongoDBDatabase: databaseName})
	if err != nil {
		if strings.EqualFold(os.Getenv("LOGARIFT_REQUIRE_MONGODB_TESTS"), "true") {
			t.Fatalf("database.Connect: %v", err)
		}
		t.Skipf("MongoDB unavailable at %s: %v", MongoURI(), err)
	}
	return client
}
