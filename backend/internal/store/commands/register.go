package commands

import "github.com/MisterVVP/logarift/backend/internal/store/cqrs"

func Register(dispatcher *cqrs.Dispatcher, handlers Handlers) error {
	registrations := []func() error{
		func() error {
			return cqrs.RegisterCommand(dispatcher, handlers.createFrictionEvent)
		},
		func() error {
			return cqrs.RegisterCommand(dispatcher, handlers.updateFrictionEvent)
		},
		func() error {
			return cqrs.RegisterCommand(dispatcher, handlers.deleteFrictionEvent)
		},
		func() error {
			return cqrs.RegisterCommand(dispatcher, handlers.createWorkGoal)
		},
		func() error {
			return cqrs.RegisterCommand(dispatcher, handlers.updateWorkGoal)
		},
		func() error {
			return cqrs.RegisterCommand(dispatcher, handlers.deleteWorkGoal)
		},
		func() error {
			return cqrs.RegisterCommand(dispatcher, handlers.createWorkSession)
		},
		func() error {
			return cqrs.RegisterCommand(dispatcher, handlers.updateWorkSession)
		},
		func() error {
			return cqrs.RegisterCommand(dispatcher, handlers.deleteWorkSession)
		},
		func() error {
			return cqrs.RegisterCommand(dispatcher, handlers.createScoreSnapshot)
		},
		func() error {
			return cqrs.RegisterCommand(dispatcher, handlers.deleteScoreSnapshot)
		},
		func() error {
			return cqrs.RegisterCommand(dispatcher, handlers.createModelConfig)
		},
		func() error {
			return cqrs.RegisterCommand(dispatcher, handlers.setDefaultModelConfig)
		},
		func() error {
			return cqrs.RegisterCommand(dispatcher, handlers.ensureDefaultModelConfig)
		},
		func() error {
			return cqrs.RegisterCommand(dispatcher, handlers.createExport)
		},
		func() error {
			return cqrs.RegisterCommand(dispatcher, handlers.updateExportStatus)
		},
		func() error {
			return cqrs.RegisterCommand(dispatcher, handlers.deleteExport)
		},
	}
	for _, register := range registrations {
		if err := register(); err != nil {
			return err
		}
	}
	return nil
}
