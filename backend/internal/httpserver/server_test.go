package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MisterVVP/logarift/backend/internal/config"
	"github.com/MisterVVP/logarift/backend/internal/version"
)

type fakeChecker struct {
	err error
}

func (f fakeChecker) Ping(ctx context.Context) error {
	return f.err
}
func (f fakeChecker) DatabaseName() string { return "logarift" }

func testConfig() config.Config {
	return config.Config{
		APIHost:          "127.0.0.1",
		APIPort:          "8080",
		MongoDBURI:       "mongodb://localhost:27017",
		MongoDBDatabase:  "logarift",
		MathEnginePath:   "./bin/friction-math",
		ExportDir:        "./exports",
		ReadinessTimeout: 100 * time.Millisecond,
		ShutdownTimeout:  100 * time.Millisecond,
	}
}

func TestLivenessReturnsOK(t *testing.T) {
	server := New(testConfig(), fakeChecker{}, version.BuildInfo{Service: "logarift-api", Version: "test"})
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if payload["status"] != "ok" {
		t.Fatalf("expected status ok, got %#v", payload["status"])
	}
}

func TestReadinessReturnsOKWhenMongoPingSucceeds(t *testing.T) {
	server := New(testConfig(), fakeChecker{}, version.BuildInfo{Service: "logarift-api", Version: "test"})
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}

func TestReadinessReturnsUnavailableWhenMongoPingFails(t *testing.T) {
	server := New(testConfig(), fakeChecker{err: errors.New("down")}, version.BuildInfo{Service: "logarift-api", Version: "test"})
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", rr.Code)
	}
	var payload healthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if payload.Checks["mongodb"] != "unavailable" {
		t.Fatalf("expected mongodb unavailable check, got %#v", payload.Checks)
	}
}

func TestStatusExposesMVP2DatabaseAndCapabilities(t *testing.T) {
	server := New(testConfig(), fakeChecker{}, version.BuildInfo{Service: "logarift-api", Version: "test", Commit: "abc"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	capabilities, ok := payload["capabilities"].(map[string]any)
	if !ok {
		t.Fatalf("missing capabilities map: %#v", payload)
	}
	if capabilities["local_first"] != true {
		t.Fatalf("expected local_first=true, got %#v", capabilities["local_first"])
	}
	if capabilities["authentication"] != false {
		t.Fatalf("expected authentication=false, got %#v", capabilities["authentication"])
	}
	if capabilities["event_crud"] != false {
		t.Fatalf("MVP-2 should not expose event CRUD yet")
	}
	database, ok := payload["database"].(map[string]any)
	if !ok {
		t.Fatalf("missing database map: %#v", payload)
	}
	if database["driver"] != "go.mongodb.org/mongo-driver/v2" || database["database_name"] != "logarift" || database["ready"] != true {
		t.Fatalf("unexpected database status: %#v", database)
	}
}
