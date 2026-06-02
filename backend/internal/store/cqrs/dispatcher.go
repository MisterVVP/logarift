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

type Command interface{}
type Query interface{}

type ContextProvider interface {
	DispatchContext() context.Context
}

type CommandHandler[C Command, R any] func(context.Context, C) (R, error)
type QueryHandler[Q Query, R any] func(context.Context, Q) (R, error)

type Dispatcher struct {
	mu              sync.RWMutex
	commandHandlers map[reflect.Type]registeredHandler
	queryHandlers   map[reflect.Type]registeredHandler
}

type registeredHandler func(context.Context, any) (any, error)

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		commandHandlers: make(map[reflect.Type]registeredHandler),
		queryHandlers:   make(map[reflect.Type]registeredHandler),
	}
}

func New() *Dispatcher { return NewDispatcher() }

func RegisterCommand[C Command, R any](d *Dispatcher, handler CommandHandler[C, R]) error {
	if d == nil {
		return errors.New("cqrs dispatcher is nil")
	}
	if handler == nil {
		return errors.New("command handler is nil")
	}

	key := typeKey[C]()
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, exists := d.commandHandlers[key]; exists {
		return fmt.Errorf("%w: command %s", ErrHandlerAlreadyRegistered, key)
	}
	d.commandHandlers[key] = func(ctx context.Context, message any) (any, error) {
		command, ok := message.(C)
		if !ok {
			return nil, fmt.Errorf("handler command type mismatch for %s", key)
		}
		return handler(ctx, command)
	}
	return nil
}

func RegisterQuery[Q Query, R any](d *Dispatcher, handler QueryHandler[Q, R]) error {
	if d == nil {
		return errors.New("cqrs dispatcher is nil")
	}
	if handler == nil {
		return errors.New("query handler is nil")
	}

	key := typeKey[Q]()
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, exists := d.queryHandlers[key]; exists {
		return fmt.Errorf("%w: query %s", ErrHandlerAlreadyRegistered, key)
	}
	d.queryHandlers[key] = func(ctx context.Context, message any) (any, error) {
		query, ok := message.(Q)
		if !ok {
			return nil, fmt.Errorf("handler query type mismatch for %s", key)
		}
		return handler(ctx, query)
	}
	return nil
}

func (d *Dispatcher) SendCommand(command Command) (any, error) {
	if d == nil {
		return nil, errors.New("cqrs dispatcher is nil")
	}

	key, err := messageType(command)
	if err != nil {
		return nil, err
	}
	d.mu.RLock()
	handler, exists := d.commandHandlers[key]
	d.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("%w: command %s", ErrHandlerNotFound, key)
	}
	return handler(messageContext(command), command)
}

func (d *Dispatcher) SendQuery(query Query) (any, error) {
	if d == nil {
		return nil, errors.New("cqrs dispatcher is nil")
	}

	key, err := messageType(query)
	if err != nil {
		return nil, err
	}
	d.mu.RLock()
	handler, exists := d.queryHandlers[key]
	d.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("%w: query %s", ErrHandlerNotFound, key)
	}
	return handler(messageContext(query), query)
}

func typeKey[T any]() reflect.Type {
	var value T
	if t := reflect.TypeOf(value); t != nil {
		return t
	}
	return reflect.TypeOf((*T)(nil)).Elem()
}

func messageType(message any) (reflect.Type, error) {
	if message == nil {
		return nil, errors.New("cqrs message is nil")
	}
	return reflect.TypeOf(message), nil
}

func messageContext(message any) context.Context {
	if provider, ok := message.(ContextProvider); ok {
		if ctx := provider.DispatchContext(); ctx != nil {
			return ctx
		}
	}

	value := reflect.ValueOf(message)
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return context.Background()
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return context.Background()
	}

	field := value.FieldByName("Context")
	if !field.IsValid() || !field.CanInterface() {
		return context.Background()
	}
	if isNilReflectValue(field) {
		return context.Background()
	}
	if ctx, ok := field.Interface().(context.Context); ok && ctx != nil {
		return ctx
	}
	return context.Background()
}

func isNilReflectValue(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
