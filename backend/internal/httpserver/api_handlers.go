package httpserver

import (
	"net/http"
	"time"

	"github.com/MisterVVP/logarift/backend/internal/friction"
	"github.com/MisterVVP/logarift/backend/internal/goals"
	"github.com/MisterVVP/logarift/backend/internal/scoring"
	"github.com/MisterVVP/logarift/backend/internal/serviceerror"
	"github.com/MisterVVP/logarift/backend/internal/sessions"
)

type apiServices struct {
	friction *friction.Service
	goals    *goals.Service
	sessions *sessions.Service
	scoring  *scoring.Service
}

func (s *Server) registerAPIRoutes() {
	if s.api.friction != nil {
		s.router.HandleFunc("POST /api/v1/friction-events/quick", s.createQuickFrictionEvent)
		s.router.HandleFunc("POST /api/v1/friction-events", s.createFrictionEvent)
		s.router.HandleFunc("GET /api/v1/friction-events", s.listFrictionEvents)
		s.router.HandleFunc("GET /api/v1/friction-events/{id}", s.getFrictionEvent)
		s.router.HandleFunc("GET /api/v1/enrichment-jobs/{id}", s.getLLMEnrichmentJob)
		s.router.HandleFunc("PUT /api/v1/friction-events/{id}", s.updateFrictionEvent)
		s.router.HandleFunc("DELETE /api/v1/friction-events/{id}", s.deleteFrictionEvent)
	}
	if s.api.goals != nil {
		s.router.HandleFunc("POST /api/v1/work-goals", s.createWorkGoal)
		s.router.HandleFunc("GET /api/v1/work-goals", s.listWorkGoals)
		s.router.HandleFunc("GET /api/v1/work-goals/{id}", s.getWorkGoal)
		s.router.HandleFunc("PUT /api/v1/work-goals/{id}", s.updateWorkGoal)
		s.router.HandleFunc("DELETE /api/v1/work-goals/{id}", s.deleteWorkGoal)
	}
	if s.api.sessions != nil {
		s.router.HandleFunc("POST /api/v1/work-sessions", s.createWorkSession)
		s.router.HandleFunc("GET /api/v1/work-sessions", s.listWorkSessions)
		s.router.HandleFunc("GET /api/v1/work-sessions/{id}", s.getWorkSession)
		s.router.HandleFunc("PUT /api/v1/work-sessions/{id}", s.updateWorkSession)
		s.router.HandleFunc("DELETE /api/v1/work-sessions/{id}", s.deleteWorkSession)
	}

	if s.api.scoring != nil {
		s.router.HandleFunc("POST /api/v1/scores/calculate", s.calculateScores)
		s.router.HandleFunc("GET /api/v1/score-snapshots", s.listScoreSnapshots)
		s.router.HandleFunc("GET /api/v1/score-snapshots/{id}", s.getScoreSnapshot)
	}
}

func (s *Server) createQuickFrictionEvent(w http.ResponseWriter, r *http.Request) {
	var req friction.QuickRequest
	if err := decodeJSON(r, &req); err != nil {
		writeInvalidJSON(w)
		return
	}
	result, err := s.api.friction.CreateQuick(r.Context(), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"event": result.Event, "enrichment": result.Enrichment})
}

func (s *Server) createFrictionEvent(w http.ResponseWriter, r *http.Request) {
	var req friction.Request
	if err := decodeJSON(r, &req); err != nil {
		writeInvalidJSON(w)
		return
	}
	event, err := s.api.friction.Create(r.Context(), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"event": event})
}
func (s *Server) listFrictionEvents(w http.ResponseWriter, r *http.Request) {
	limit, err := parseLimit(r.URL.Query().Get("limit"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	from, err := queryTime(r, "from")
	if err != nil {
		writeServiceError(w, err)
		return
	}
	to, err := queryTime(r, "to")
	if err != nil {
		writeServiceError(w, err)
		return
	}
	q := r.URL.Query()
	page, err := s.api.friction.List(r.Context(), friction.Filter{From: from, To: to, WorkflowStage: q.Get("workflow_stage"), FrictionLayer: q.Get("friction_layer"), FrictionType: q.Get("friction_type"), GoalID: q.Get("goal_id"), SessionID: q.Get("session_id"), Source: q.Get("source"), Limit: limit, Cursor: q.Get("cursor")})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": page.Events, "next_cursor": page.NextCursor})
}
func (s *Server) getFrictionEvent(w http.ResponseWriter, r *http.Request) {
	event, err := s.api.friction.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"event": event})
}

func (s *Server) getLLMEnrichmentJob(w http.ResponseWriter, r *http.Request) {
	job, err := s.api.friction.GetLLMEnrichmentJob(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"job": job})
}

func (s *Server) updateFrictionEvent(w http.ResponseWriter, r *http.Request) {
	var req friction.Request
	if err := decodeJSON(r, &req); err != nil {
		writeInvalidJSON(w)
		return
	}
	event, err := s.api.friction.Update(r.Context(), r.PathValue("id"), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"event": event})
}
func (s *Server) deleteFrictionEvent(w http.ResponseWriter, r *http.Request) {
	if err := s.api.friction.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) createWorkGoal(w http.ResponseWriter, r *http.Request) {
	var req goals.Request
	if err := decodeJSON(r, &req); err != nil {
		writeInvalidJSON(w)
		return
	}
	goal, err := s.api.goals.Create(r.Context(), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"goal": goal})
}
func (s *Server) listWorkGoals(w http.ResponseWriter, r *http.Request) {
	limit, err := parseLimit(r.URL.Query().Get("limit"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	page, err := s.api.goals.List(r.Context(), goals.Filter{Status: r.URL.Query().Get("status"), Limit: limit, Cursor: r.URL.Query().Get("cursor")})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"goals": page.Goals, "next_cursor": page.NextCursor})
}
func (s *Server) getWorkGoal(w http.ResponseWriter, r *http.Request) {
	goal, err := s.api.goals.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"goal": goal})
}
func (s *Server) updateWorkGoal(w http.ResponseWriter, r *http.Request) {
	var req goals.Request
	if err := decodeJSON(r, &req); err != nil {
		writeInvalidJSON(w)
		return
	}
	goal, err := s.api.goals.Update(r.Context(), r.PathValue("id"), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"goal": goal})
}
func (s *Server) deleteWorkGoal(w http.ResponseWriter, r *http.Request) {
	if err := s.api.goals.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) createWorkSession(w http.ResponseWriter, r *http.Request) {
	var req sessions.Request
	if err := decodeJSON(r, &req); err != nil {
		writeInvalidJSON(w)
		return
	}
	session, err := s.api.sessions.Create(r.Context(), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"session": session})
}
func (s *Server) listWorkSessions(w http.ResponseWriter, r *http.Request) {
	limit, err := parseLimit(r.URL.Query().Get("limit"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	from, err := queryTime(r, "from")
	if err != nil {
		writeServiceError(w, err)
		return
	}
	to, err := queryTime(r, "to")
	if err != nil {
		writeServiceError(w, err)
		return
	}
	page, err := s.api.sessions.List(r.Context(), sessions.Filter{From: from, To: to, GoalID: r.URL.Query().Get("goal_id"), Limit: limit, Cursor: r.URL.Query().Get("cursor")})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": page.Sessions, "next_cursor": page.NextCursor})
}
func (s *Server) getWorkSession(w http.ResponseWriter, r *http.Request) {
	session, err := s.api.sessions.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"session": session})
}
func (s *Server) updateWorkSession(w http.ResponseWriter, r *http.Request) {
	var req sessions.Request
	if err := decodeJSON(r, &req); err != nil {
		writeInvalidJSON(w)
		return
	}
	session, err := s.api.sessions.Update(r.Context(), r.PathValue("id"), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"session": session})
}
func (s *Server) deleteWorkSession(w http.ResponseWriter, r *http.Request) {
	if err := s.api.sessions.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func queryTime(r *http.Request, key string) (*time.Time, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return nil, serviceerror.ValidationError{Fields: []serviceerror.FieldError{{Field: key, Message: "must be an RFC3339 timestamp"}}}
	}
	t = t.UTC()
	return &t, nil
}

func (s *Server) calculateScores(w http.ResponseWriter, r *http.Request) {
	var req scoring.CalculateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeInvalidJSON(w)
		return
	}
	snapshot, err := s.api.scoring.Calculate(r.Context(), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"snapshot": snapshot})
}

func (s *Server) listScoreSnapshots(w http.ResponseWriter, r *http.Request) {
	limit, err := parseLimit(r.URL.Query().Get("limit"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	from, err := queryTime(r, "from")
	if err != nil {
		writeServiceError(w, err)
		return
	}
	to, err := queryTime(r, "to")
	if err != nil {
		writeServiceError(w, err)
		return
	}
	page, err := s.api.scoring.List(r.Context(), scoring.Filter{From: from, To: to, ScoreType: r.URL.Query().Get("score_type"), Limit: limit})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"snapshots": page.Snapshots, "next_cursor": page.NextCursor})
}

func (s *Server) getScoreSnapshot(w http.ResponseWriter, r *http.Request) {
	snapshot, err := s.api.scoring.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"snapshot": snapshot})
}
