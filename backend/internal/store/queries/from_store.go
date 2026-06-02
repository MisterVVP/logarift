package queries

import "github.com/MisterVVP/logarift/backend/internal/store"

func FromStore(s interface {
	Repositories() (store.FrictionEventRepository, store.WorkGoalRepository, store.WorkSessionRepository, store.ScoreSnapshotRepository, store.ModelConfigRepository, store.ExportRepository)
}) Handlers {
	frictionEvents, workGoals, workSessions, scoreSnapshots, modelConfigs, exports := s.Repositories()
	return Handlers{FrictionEvents: frictionEvents, WorkGoals: workGoals, WorkSessions: workSessions, ScoreSnapshots: scoreSnapshots, ModelConfigs: modelConfigs, Exports: exports}
}
