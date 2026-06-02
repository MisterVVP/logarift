package cqrs

import (
	"context"
	"errors"
	"testing"
)

type testCommand struct {
	Context context.Context
	Value   int
}
type testQuery struct {
	Context context.Context
	Prefix  string
}

func TestDispatcherDispatchesRegisteredCommand(t *testing.T) {
	dispatcher := NewDispatcher()
	if err := RegisterCommand[testCommand, int](dispatcher, func(ctx context.Context, command testCommand) (int, error) {
		return command.Value + 1, nil
	}); err != nil {
		t.Fatalf("RegisterCommand() error: %v", err)
	}

	got, err := dispatcher.SendCommand(testCommand{Value: 41})
	if err != nil {
		t.Fatalf("SendCommand() error: %v", err)
	}
	if got != 42 {
		t.Fatalf("expected command result 42, got %#v", got)
	}
}

func TestDispatcherDispatchesRegisteredQuery(t *testing.T) {
	dispatcher := NewDispatcher()
	if err := RegisterQuery[testQuery, string](dispatcher, func(ctx context.Context, query testQuery) (string, error) {
		return query.Prefix + "result", nil
	}); err != nil {
		t.Fatalf("RegisterQuery() error: %v", err)
	}

	got, err := dispatcher.SendQuery(testQuery{Prefix: "cqrs-"})
	if err != nil {
		t.Fatalf("SendQuery() error: %v", err)
	}
	if got != "cqrs-result" {
		t.Fatalf("expected query result cqrs-result, got %#v", got)
	}
}

func TestDispatcherUsesOptionalMessageContext(t *testing.T) {
	type contextKey string
	const key contextKey = "value"
	ctx := context.WithValue(context.Background(), key, "from-context")
	dispatcher := NewDispatcher()
	if err := RegisterQuery[testQuery, string](dispatcher, func(ctx context.Context, query testQuery) (string, error) {
		value, _ := ctx.Value(key).(string)
		return value, nil
	}); err != nil {
		t.Fatalf("RegisterQuery() error: %v", err)
	}

	got, err := dispatcher.SendQuery(testQuery{Context: ctx})
	if err != nil {
		t.Fatalf("SendQuery() error: %v", err)
	}
	if got != "from-context" {
		t.Fatalf("expected context value, got %#v", got)
	}
}

func TestDispatcherKeepsCommandAndQueryHandlersSeparate(t *testing.T) {
	dispatcher := NewDispatcher()
	if err := RegisterCommand[testCommand, int](dispatcher, func(ctx context.Context, command testCommand) (int, error) {
		return command.Value, nil
	}); err != nil {
		t.Fatalf("RegisterCommand() error: %v", err)
	}

	_, err := dispatcher.SendQuery(testCommand{Value: 1})
	if !errors.Is(err, ErrHandlerNotFound) {
		t.Fatalf("expected ErrHandlerNotFound for query side, got %v", err)
	}
}

func TestDispatcherRejectsDuplicateCommandRegistration(t *testing.T) {
	dispatcher := NewDispatcher()
	handler := func(ctx context.Context, command testCommand) (int, error) { return command.Value, nil }
	if err := RegisterCommand[testCommand, int](dispatcher, handler); err != nil {
		t.Fatalf("RegisterCommand() first error: %v", err)
	}
	if err := RegisterCommand[testCommand, int](dispatcher, handler); !errors.Is(err, ErrHandlerAlreadyRegistered) {
		t.Fatalf("expected ErrHandlerAlreadyRegistered, got %v", err)
	}
}
