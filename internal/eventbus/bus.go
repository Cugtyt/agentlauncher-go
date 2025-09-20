package eventbus

import (
	"context"
	"log"
	"reflect"
	"runtime"
	"sync"
)

type VerboseLevel int

const (
	SILENT VerboseLevel = iota
	BASIC
	DETAILED
)

type CompiledHandler interface {
	Call(context.Context, Event)
}

type compiledHandler[T Event] struct {
	handler func(context.Context, T)
}

func (ch *compiledHandler[T]) Call(ctx context.Context, event Event) {
	if typedEvent, ok := event.(T); ok {
		ch.handler(ctx, typedEvent)
	}
}

type work struct {
	event   Event
	handler CompiledHandler
}

type EventBus struct {
	handlerMap map[reflect.Type][]CompiledHandler

	eventQueue chan Event
	workerPool chan work

	numWorkers int
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup

	verbose VerboseLevel
	logger  *log.Logger
}

func NewEventBus() *EventBus {
	bufferSize := 100
	numWorkers := runtime.NumCPU() * 2
	ctx, cancel := context.WithCancel(context.Background())

	eb := &EventBus{
		handlerMap: make(map[reflect.Type][]CompiledHandler),
		eventQueue: make(chan Event, bufferSize),
		workerPool: make(chan work, bufferSize*2),
		numWorkers: numWorkers,
		ctx:        ctx,
		cancel:     cancel,
		verbose:    SILENT,
		logger:     log.New(log.Writer(), "EventBus: ", log.LstdFlags),
	}

	eb.wg.Add(1)
	go eb.dispatcher()

	for i := 0; i < eb.numWorkers; i++ {
		eb.wg.Add(1)
		go eb.worker()
	}

	return eb
}

func (eb *EventBus) WithVerboseLevel(level VerboseLevel) {
	eb.verbose = level
}

func Subscribe[T Event](eb *EventBus, handler func(context.Context, T)) {
	var zero T
	eventType := reflect.TypeOf(zero)

	compiled := &compiledHandler[T]{handler: handler}
	eb.handlerMap[eventType] = append(eb.handlerMap[eventType], compiled)
}

func (eb *EventBus) Emit(event Event) {
	eb.eventQueue <- event
	eb.log(event)
}

func (eb *EventBus) dispatcher() {
	defer eb.wg.Done()

	for {
		select {
		case event := <-eb.eventQueue:
			eb.dispatchEvent(event)

		case <-eb.ctx.Done():
			eb.drainEvents()
			return
		}
	}
}

func (eb *EventBus) dispatchEvent(event Event) {
	eventType := reflect.TypeOf(event)

	handlers := eb.handlerMap[eventType]

	for _, handler := range handlers {
		eb.workerPool <- work{event: event, handler: handler}
	}
}

func (eb *EventBus) worker() {
	defer eb.wg.Done()

	for {
		select {
		case w := <-eb.workerPool:
			w.handler.Call(eb.ctx, w.event)

		case <-eb.ctx.Done():
			return
		}
	}
}

func (eb *EventBus) drainEvents() {
	for {
		select {
		case event := <-eb.eventQueue:
			eb.dispatchEvent(event)
		default:
			return
		}
	}
}

func (eb *EventBus) Shutdown(ctx context.Context) error {
	eb.cancel()

	done := make(chan struct{})
	go func() {
		eb.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (eb *EventBus) log(event Event) {
	if eb.verbose == SILENT {
		return
	}

	agentID := ""
	val := reflect.ValueOf(event)
	field := val.FieldByName("AgentID")
	if field.IsValid() && field.Kind() == reflect.String {
		agentID = field.String()
	}

	eventType := reflect.TypeOf(event).Name()

	if eb.verbose == BASIC {
		eb.logger.Printf("[%s] Event emitted: %s", agentID, eventType)
		return
	}

	eb.logger.Printf("----- Event emitted: %s -----", eventType)
	eb.logger.Printf("%+v\n", event)
	eb.logger.Println("-------------------------")
}
