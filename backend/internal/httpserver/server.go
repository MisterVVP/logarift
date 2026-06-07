package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/MisterVVP/logarift/backend/internal/config"
	"github.com/MisterVVP/logarift/backend/internal/friction"
	"github.com/MisterVVP/logarift/backend/internal/goals"
	"github.com/MisterVVP/logarift/backend/internal/llmadapter"
	"github.com/MisterVVP/logarift/backend/internal/llmqueue"
	"github.com/MisterVVP/logarift/backend/internal/scoring"
	"github.com/MisterVVP/logarift/backend/internal/sessions"
	"github.com/MisterVVP/logarift/backend/internal/store/cqrs"
	"github.com/MisterVVP/logarift/backend/internal/version"
)

// HealthChecker is satisfied by the MongoDB client wrapper and by tests.
type HealthChecker interface {
	Ping(ctx context.Context) error
	DatabaseName() string
}

type Server struct {
	cfg     config.Config
	build   version.BuildInfo
	checker HealthChecker
	router  *http.ServeMux
	now     func() time.Time
	api     apiServices
}

func New(cfg config.Config, checker HealthChecker, build version.BuildInfo) *Server {
	return newServer(cfg, checker, build, apiServices{})
}

func NewWithDispatcher(cfg config.Config, checker HealthChecker, build version.BuildInfo, dispatcher *cqrs.Dispatcher) *Server {
	frictionService := friction.NewService(dispatcher, nil)
	if cfg.LLMAdapterEnabled {
		frictionService = friction.NewServiceWithLLM(dispatcher, nil, llmadapter.NewClient(cfg.LLMAdapterURL, cfg.LLMAdapterTimeout), cfg.LLMAdapterMinConfidence, cfg.LLMAdapterPromptPrivacyMode == "markdown")
		frictionService.SetLLMAdapterURL(cfg.LLMAdapterURL)
		if cfg.ValkeyEnabled {
			queue, err := llmqueue.NewValkeyStream(llmqueue.Options{URL: cfg.ValkeyURL, Stream: cfg.ValkeyStream, Group: cfg.ValkeyGroup, Consumer: cfg.ValkeyConsumer})
			if err == nil {
				frictionService.SetLLMJobQueue(queue)
			}
		}
	}
	return newServer(cfg, checker, build, apiServices{friction: frictionService, goals: goals.NewService(dispatcher, nil), sessions: sessions.NewService(dispatcher, nil), scoring: scoring.NewService(dispatcher, cfg.MathEngineURL, nil)})
}

func newServer(cfg config.Config, checker HealthChecker, build version.BuildInfo, api apiServices) *Server {
	s := &Server{
		cfg:     cfg,
		build:   build,
		checker: checker,
		router:  http.NewServeMux(),
		now:     time.Now,
		api:     api,
	}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	return corsMiddleware(requestIDMiddleware(loggingMiddleware(s.router)))
}

func (s *Server) StartBackground(ctx context.Context) {
	if s.api.friction != nil {
		s.api.friction.StartLLMJobConsumer(ctx)
	}
}

func (s *Server) routes() {
	s.router.HandleFunc("GET /", s.handleIndex)
	s.router.HandleFunc("GET /health/live", s.handleLiveness)
	s.router.HandleFunc("GET /health/ready", s.handleReadiness)
	s.router.HandleFunc("GET /api/v1/status", s.handleStatus)
	s.router.HandleFunc("POST /api/v1/uploads", s.uploadAttachment)
	s.router.HandleFunc("GET /uploads/{filename}", s.serveUploadedAttachment)
	s.registerAPIRoutes()
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service": s.build.Service,
		"version": s.build.Version,
		"links": map[string]string{
			"liveness":  "/health/live",
			"readiness": "/health/ready",
			"status":    "/api/v1/status",
		},
	})
}

func (s *Server) handleLiveness(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status:    "ok",
		Service:   s.build.Service,
		Version:   s.build.Version,
		Timestamp: s.now().UTC(),
		Checks: map[string]string{
			"process": "ok",
		},
	})
}

func (s *Server) handleReadiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.ReadinessTimeout)
	defer cancel()

	status := http.StatusOK
	checks := map[string]string{"mongodb": "ok"}
	overall := "ok"
	if s.checker == nil {
		status = http.StatusServiceUnavailable
		checks["mongodb"] = "not_configured"
		overall = "unavailable"
	} else if err := s.checker.Ping(ctx); err != nil {
		status = http.StatusServiceUnavailable
		checks["mongodb"] = "unavailable"
		overall = "unavailable"
	}

	writeJSON(w, status, healthResponse{
		Status:    overall,
		Service:   s.build.Service,
		Version:   s.build.Version,
		Timestamp: s.now().UTC(),
		Checks:    checks,
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	databaseName := s.cfg.MongoDBDatabase
	ready := false
	if s.checker != nil {
		if name := s.checker.DatabaseName(); name != "" {
			databaseName = name
		}
		ctx, cancel := context.WithTimeout(r.Context(), s.cfg.ReadinessTimeout)
		defer cancel()
		ready = s.checker.Ping(ctx) == nil
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"service": s.build.Service,
		"version": s.build.Version,
		"commit":  s.build.Commit,
		"runtime": map[string]any{
			"config": s.cfg.PublicStatus(),
		},
		"database": map[string]any{
			"kind":          "mongodb",
			"database_name": databaseName,
			"driver":        "go.mongodb.org/mongo-driver/v2",
			"ready":         ready,
		},
		"capabilities": map[string]bool{
			"local_first":          true,
			"single_user":          true,
			"authentication":       false,
			"cloud_sync":           false,
			"hidden_telemetry":     false,
			"event_crud":           s.api.friction != nil && s.api.goals != nil && s.api.sessions != nil,
			"quick_logging":        s.api.friction != nil,
			"local_uploads":        true,
			"rich_notes":           true,
			"deterministic_rules":  s.api.friction != nil,
			"scoring":              s.api.scoring != nil,
			"async_llm_enrichment": s.api.friction != nil && s.cfg.LLMAdapterEnabled,
			"valkey_streams":       s.cfg.ValkeyEnabled,
		},
	})
}

type healthResponse struct {
	Status    string            `json:"status"`
	Service   string            `json:"service"`
	Version   string            `json:"version"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(true)
	_ = encoder.Encode(payload)
}
