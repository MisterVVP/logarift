package queries

import "github.com/MisterVVP/logarift/backend/internal/store/cqrs"

func Register(dispatcher *cqrs.Dispatcher, handlers Handlers) error {
	registrations := []func() error{
		func() error {
			return cqrs.RegisterQuery(dispatcher, handlers.getFrictionEventByID)
		},
		func() error {
			return cqrs.RegisterQuery(dispatcher, handlers.listFrictionEvents)
		},
		func() error {
			return cqrs.RegisterQuery(dispatcher, handlers.getWorkGoalByID)
		},
		func() error {
			return cqrs.RegisterQuery(dispatcher, handlers.listWorkGoals)
		},
		func() error {
			return cqrs.RegisterQuery(dispatcher, handlers.getWorkSessionByID)
		},
		func() error {
			return cqrs.RegisterQuery(dispatcher, handlers.listWorkSessions)
		},
		func() error {
			return cqrs.RegisterQuery(dispatcher, handlers.getScoreSnapshotByID)
		},
		func() error {
			return cqrs.RegisterQuery(dispatcher, handlers.listScoreSnapshots)
		},
		func() error {
			return cqrs.RegisterQuery(dispatcher, handlers.getModelConfigByID)
		},
		func() error {
			return cqrs.RegisterQuery(dispatcher, handlers.getDefaultModelConfig)
		},
		func() error {
			return cqrs.RegisterQuery(dispatcher, handlers.listModelConfigs)
		},
		func() error {
			return cqrs.RegisterQuery(dispatcher, handlers.getExportByID)
		},
		func() error {
			return cqrs.RegisterQuery(dispatcher, handlers.listExports)
		},
	}
	for _, register := range registrations {
		if err := register(); err != nil {
			return err
		}
	}
	return nil
}
