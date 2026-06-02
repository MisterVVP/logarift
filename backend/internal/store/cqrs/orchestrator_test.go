package cqrs

import (
	"context"
	"errors"
	"testing"
)

type testCommand struct{ Value int }
type testQuery struct{ Prefix string }

func TestOrchestratorDispatchesRegisteredCommand(t *testing.T) {
	bus := New()
	if err := RegisterCommand[testCommand, int](bus, func(ctx context.Context, command testCommand) (int, error) {
		return command.Value + 1, nil
	}); err != nil {
		t.Fatalf("RegisterCommand() error: %v", err)
	}

	got, err := SendCommand[testCommand, int](context.Background(), bus, testCommand{Value: 41})
	if err != nil {
		t.Fatalf("SendCommand() error: %v", err)
	}
	if got != 42 {
		t.Fatalf("expected command result 42, got %d", got)
	}
}

func TestOrchestratorDispatchesRegisteredQuery(t *testing.T) {
	bus := New()
	if err := RegisterQuery[testQuery, string](bus, func(ctx context.Context, query testQuery) (string, error) {
		return query.Prefix + "result", nil
	}); err != nil {
		t.Fatalf("RegisterQuery() error: %v", err)
	}

	got, err := SendQuery[testQuery, string](context.Background(), bus, testQuery{Prefix: "cqrs-"})
	if err != nil {
		t.Fatalf("SendQuery() error: %v", err)
	}
	if got != "cqrs-result" {
		t.Fatalf("expected query result cqrs-result, got %q", got)
	}
}

func TestOrchestratorKeepsCommandAndQueryHandlersSeparate(t *testing.T) {
	bus := New()
	if err := RegisterCommand[testCommand, int](bus, func(ctx context.Context, command testCommand) (int, error) {
		return command.Value, nil
	}); err != nil {
		t.Fatalf("RegisterCommand() error: %v", err)
	}

	_, err := SendQuery[testCommand, int](context.Background(), bus, testCommand{Value: 1})
	if !errors.Is(err, ErrHandlerNotFound) {
		t.Fatalf("expected ErrHandlerNotFound for query side, got %v", err)
	}
}

func TestOrchestratorRejectsDuplicateCommandRegistration(t *testing.T) {
	bus := New()
	handler := func(ctx context.Context, command testCommand) (int, error) { return command.Value, nil }
	if err := RegisterCommand[testCommand, int](bus, handler); err != nil {
		t.Fatalf("RegisterCommand() first error: %v", err)
	}
	if err := RegisterCommand[testCommand, int](bus, handler); !errors.Is(err, ErrHandlerAlreadyRegistered) {
		t.Fatalf("expected ErrHandlerAlreadyRegistered, got %v", err)
	}
}
