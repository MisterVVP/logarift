package cqrs

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

var (
	ErrHandlerNotFound          = errors.New("handler not found")
	ErrHandlerAlreadyRegistered = errors.New("handler already registered")
)

type Empty struct{}

type CommandHandler[C any, R any] func(context.Context, C) (R, error)
type QueryHandler[Q any, R any] func(context.Context, Q) (R, error)

type Orchestrator struct {
	mu              sync.RWMutex
	commandHandlers map[reflect.Type]any
	queryHandlers   map[reflect.Type]any
}

func New() *Orchestrator {
	return &Orchestrator{
		commandHandlers: make(map[reflect.Type]any),
		queryHandlers:   make(map[reflect.Type]any),
	}
}

func RegisterCommand[C any, R any](o *Orchestrator, handler CommandHandler[C, R]) error {
	if o == nil {
		return errors.New("cqrs orchestrator is nil")
	}
	if handler == nil {
		return errors.New("command handler is nil")
	}

	key := typeKey[C]()
	o.mu.Lock()
	defer o.mu.Unlock()
	if _, exists := o.commandHandlers[key]; exists {
		return fmt.Errorf("%w: command %s", ErrHandlerAlreadyRegistered, key)
	}
	o.commandHandlers[key] = handler
	return nil
}

func RegisterQuery[Q any, R any](o *Orchestrator, handler QueryHandler[Q, R]) error {
	if o == nil {
		return errors.New("cqrs orchestrator is nil")
	}
	if handler == nil {
		return errors.New("query handler is nil")
	}

	key := typeKey[Q]()
	o.mu.Lock()
	defer o.mu.Unlock()
	if _, exists := o.queryHandlers[key]; exists {
		return fmt.Errorf("%w: query %s", ErrHandlerAlreadyRegistered, key)
	}
	o.queryHandlers[key] = handler
	return nil
}

func SendCommand[C any, R any](ctx context.Context, o *Orchestrator, command C) (R, error) {
	var zero R
	if o == nil {
		return zero, errors.New("cqrs orchestrator is nil")
	}

	key := typeKey[C]()
	o.mu.RLock()
	handler, exists := o.commandHandlers[key]
	o.mu.RUnlock()
	if !exists {
		return zero, fmt.Errorf("%w: command %s", ErrHandlerNotFound, key)
	}

	typed, ok := handler.(CommandHandler[C, R])
	if !ok {
		return zero, fmt.Errorf("handler result type mismatch for command %s", key)
	}
	return typed(ctx, command)
}

func SendQuery[Q any, R any](ctx context.Context, o *Orchestrator, query Q) (R, error) {
	var zero R
	if o == nil {
		return zero, errors.New("cqrs orchestrator is nil")
	}

	key := typeKey[Q]()
	o.mu.RLock()
	handler, exists := o.queryHandlers[key]
	o.mu.RUnlock()
	if !exists {
		return zero, fmt.Errorf("%w: query %s", ErrHandlerNotFound, key)
	}

	typed, ok := handler.(QueryHandler[Q, R])
	if !ok {
		return zero, fmt.Errorf("handler result type mismatch for query %s", key)
	}
	return typed(ctx, query)
}

func typeKey[T any]() reflect.Type {
	var value T
	if t := reflect.TypeOf(value); t != nil {
		return t
	}
	return reflect.TypeOf((*T)(nil)).Elem()
}
