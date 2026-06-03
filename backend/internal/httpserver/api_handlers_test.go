package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MisterVVP/logarift/backend/internal/config"
	"github.com/MisterVVP/logarift/backend/internal/database"
	"github.com/MisterVVP/logarift/backend/internal/store/cqrs"
	"github.com/MisterVVP/logarift/backend/internal/store/mongostore"
	"github.com/MisterVVP/logarift/backend/internal/version"
)

func newAPITestServer(t *testing.T) *Server {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)
	client, err := database.Connect(ctx, config.Config{MongoDBURI: "mongodb://localhost:27017", MongoDBDatabase: "api_handlers_test" + time.Now().Format("150405.000000000")})
	if err != nil {
		t.Fatalf("database.Connect: %v", err)
	}
	t.Cleanup(func() { _ = client.Database().Drop(context.Background()) })
	stores := mongostore.New(client)
	dispatcher := cqrs.NewDispatcher()
	if err := stores.RegisterCQRS(dispatcher); err != nil {
		t.Fatalf("RegisterCQRS: %v", err)
	}
	return NewWithDispatcher(testConfig(), fakeChecker{}, version.BuildInfo{Service: "logarift-api", Version: "test"}, dispatcher)
}

func doJSON(t *testing.T, server *Server, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	return rr
}
func stringField(t *testing.T, body []byte, keys ...string) string {
	t.Helper()
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatalf("invalid JSON: %v body=%s", err, string(body))
	}
	cur := v
	for _, key := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			t.Fatalf("missing map for %s in %#v", key, cur)
		}
		cur = m[key]
	}
	s, ok := cur.(string)
	if !ok {
		t.Fatalf("expected string at %v, got %#v", keys, cur)
	}
	return s
}

func TestWorkGoalHandlersCRUDAndErrors(t *testing.T) {
	server := newAPITestServer(t)
	create := doJSON(t, server, http.MethodPost, "/api/v1/work-goals", `{"title":" Implement event API ","status":"active"}`)
	if create.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", create.Code, create.Body.String())
	}
	id := stringField(t, create.Body.Bytes(), "goal", "id")
	list := doJSON(t, server, http.MethodGet, "/api/v1/work-goals?status=active&limit=20", "")
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), "Implement event API") {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	get := doJSON(t, server, http.MethodGet, "/api/v1/work-goals/"+id, "")
	if get.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", get.Code, get.Body.String())
	}
	update := doJSON(t, server, http.MethodPut, "/api/v1/work-goals/"+id, `{"title":"Done","status":"completed"}`)
	if update.Code != http.StatusOK || !strings.Contains(update.Body.String(), "completed") {
		t.Fatalf("update status=%d body=%s", update.Code, update.Body.String())
	}
	bad := doJSON(t, server, http.MethodPost, "/api/v1/work-goals", `{"title":"x","unknown":true}`)
	if bad.Code != http.StatusBadRequest || !strings.Contains(bad.Body.String(), "invalid_json") {
		t.Fatalf("bad status=%d body=%s", bad.Code, bad.Body.String())
	}
	del := doJSON(t, server, http.MethodDelete, "/api/v1/work-goals/"+id, "")
	if del.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d body=%s", del.Code, del.Body.String())
	}
	missing := doJSON(t, server, http.MethodGet, "/api/v1/work-goals/"+id, "")
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing status=%d body=%s", missing.Code, missing.Body.String())
	}
}

func TestSessionAndFrictionEventHandlers(t *testing.T) {
	server := newAPITestServer(t)
	sessionCreate := doJSON(t, server, http.MethodPost, "/api/v1/work-sessions", `{"title":"Morning work","started_at":"2026-06-01T08:30:00Z"}`)
	if sessionCreate.Code != http.StatusCreated {
		t.Fatalf("session create status=%d body=%s", sessionCreate.Code, sessionCreate.Body.String())
	}
	sessionID := stringField(t, sessionCreate.Body.Bytes(), "session", "id")
	eventBody := `{"timestamp_start":"2026-06-01T09:15:00Z","workflow_stage":"test","friction_layer":"technical","friction_type":"failed_feedback","severity_self":4,"cognitive_load_self":3,"emotion_valence":-1,"time_lost_minutes":20,"resume_time_minutes":8,"interruption_count":1,"session_id":"` + sessionID + `","tags":["ci","ci"],"notes":"CI failed locally."}`
	eventCreate := doJSON(t, server, http.MethodPost, "/api/v1/friction-events", eventBody)
	if eventCreate.Code != http.StatusCreated {
		t.Fatalf("event create status=%d body=%s", eventCreate.Code, eventCreate.Body.String())
	}
	eventID := stringField(t, eventCreate.Body.Bytes(), "event", "id")
	list := doJSON(t, server, http.MethodGet, "/api/v1/friction-events?workflow_stage=test&session_id="+sessionID+"&limit=20", "")
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), "failed_feedback") {
		t.Fatalf("event list status=%d body=%s", list.Code, list.Body.String())
	}
	badValidation := doJSON(t, server, http.MethodPost, "/api/v1/friction-events", `{"timestamp_start":"2026-06-01T09:15:00Z","workflow_stage":"bad","friction_layer":"technical","friction_type":"failed_feedback","severity_self":9,"cognitive_load_self":3,"emotion_valence":0,"time_lost_minutes":0,"resume_time_minutes":0,"interruption_count":0}`)
	if badValidation.Code != http.StatusBadRequest || !strings.Contains(badValidation.Body.String(), "validation_failed") {
		t.Fatalf("validation status=%d body=%s", badValidation.Code, badValidation.Body.String())
	}
	update := doJSON(t, server, http.MethodPut, "/api/v1/friction-events/"+eventID, strings.Replace(eventBody, "CI failed locally.", "Updated notes.", 1))
	if update.Code != http.StatusOK || !strings.Contains(update.Body.String(), "Updated notes") {
		t.Fatalf("event update status=%d body=%s", update.Code, update.Body.String())
	}
	del := doJSON(t, server, http.MethodDelete, "/api/v1/friction-events/"+eventID, "")
	if del.Code != http.StatusNoContent {
		t.Fatalf("event delete status=%d body=%s", del.Code, del.Body.String())
	}
	sessionDel := doJSON(t, server, http.MethodDelete, "/api/v1/work-sessions/"+sessionID, "")
	if sessionDel.Code != http.StatusNoContent {
		t.Fatalf("session delete status=%d body=%s", sessionDel.Code, sessionDel.Body.String())
	}
}
