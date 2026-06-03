package cqrs

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

var (
	ErrHandlerNotFound          = errors.New("handler not found")
	ErrHandlerAlreadyRegistered = errors.New("handler already registered")
)

const (
	commandPackageSuffix = "/store/commands"
	queryPackageSuffix   = "/store/queries"
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
	return d.registerCommand(key, func(ctx context.Context, message any) (any, error) {
		command, ok := message.(C)
		if !ok {
			return nil, fmt.Errorf("handler command type mismatch for %s", key)
		}
		return handler(ctx, command)
	})
}

func RegisterQuery[Q Query, R any](d *Dispatcher, handler QueryHandler[Q, R]) error {
	if d == nil {
		return errors.New("cqrs dispatcher is nil")
	}
	if handler == nil {
		return errors.New("query handler is nil")
	}

	key := typeKey[Q]()
	return d.registerQuery(key, func(ctx context.Context, message any) (any, error) {
		query, ok := message.(Q)
		if !ok {
			return nil, fmt.Errorf("handler query type mismatch for %s", key)
		}
		return handler(ctx, query)
	})
}

// RegisterHandlers reflects over typed handler functions and registers each as a
// command or query handler based on the package of its message argument. Handler
// functions must have this shape:
//
//	func(context.Context, commands.SomeCommand) (SomeResult, error)
//	func(context.Context, queries.SomeQuery) (SomeResult, error)
//
// Message types in backend/internal/store/commands are registered on the command
// side. Message types in backend/internal/store/queries are registered on the
// query side.
func RegisterHandlers(d *Dispatcher, handlers ...any) error {
	for _, handler := range handlers {
		if err := RegisterHandler(d, handler); err != nil {
			return err
		}
	}
	return nil
}

// RegisterHandler registers one reflected command/query handler. Prefer
// RegisterHandlers when registering a module's full handler set.
func RegisterHandler(d *Dispatcher, handler any) error {
	if d == nil {
		return errors.New("cqrs dispatcher is nil")
	}
	value := reflect.ValueOf(handler)
	if !value.IsValid() || value.Kind() != reflect.Func || value.IsNil() {
		return errors.New("cqrs handler must be a non-nil function")
	}

	messageType, wrapper, err := reflectedHandler(value)
	if err != nil {
		return err
	}

	messagePackage := messageType.PkgPath()
	switch {
	case strings.HasSuffix(messagePackage, commandPackageSuffix):
		return d.registerCommand(messageType, wrapper)
	case strings.HasSuffix(messagePackage, queryPackageSuffix):
		return d.registerQuery(messageType, wrapper)
	default:
		return fmt.Errorf("cqrs handler message %s is not from %s or %s", messageType, commandPackageSuffix, queryPackageSuffix)
	}
}

func reflectedHandler(value reflect.Value) (reflect.Type, registeredHandler, error) {
	typ := value.Type()
	if typ.NumIn() != 2 || typ.NumOut() != 2 {
		return nil, nil, fmt.Errorf("cqrs handler %s must have signature func(context.Context, Message) (Result, error)", typ)
	}
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	if !typ.In(0).Implements(contextType) {
		return nil, nil, fmt.Errorf("cqrs handler %s first argument must implement context.Context", typ)
	}
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if !typ.Out(1).Implements(errorType) {
		return nil, nil, fmt.Errorf("cqrs handler %s second return value must implement error", typ)
	}

	messageType := typ.In(1)
	if messageType.Kind() == reflect.Pointer {
		return nil, nil, fmt.Errorf("cqrs handler %s message argument must be a value type", typ)
	}
	wrapper := func(ctx context.Context, message any) (any, error) {
		messageValue := reflect.ValueOf(message)
		if !messageValue.IsValid() || messageValue.Type() != messageType {
			return nil, fmt.Errorf("handler message type mismatch for %s", messageType)
		}
		out := value.Call([]reflect.Value{reflect.ValueOf(ctx), messageValue})
		if !out[1].IsNil() {
			return out[0].Interface(), out[1].Interface().(error)
		}
		return out[0].Interface(), nil
	}
	return messageType, wrapper, nil
}

func (d *Dispatcher) registerCommand(key reflect.Type, handler registeredHandler) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, exists := d.commandHandlers[key]; exists {
		return fmt.Errorf("%w: command %s", ErrHandlerAlreadyRegistered, key)
	}
	d.commandHandlers[key] = handler
	return nil
}

func (d *Dispatcher) registerQuery(key reflect.Type, handler registeredHandler) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, exists := d.queryHandlers[key]; exists {
		return fmt.Errorf("%w: query %s", ErrHandlerAlreadyRegistered, key)
	}
	d.queryHandlers[key] = handler
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
