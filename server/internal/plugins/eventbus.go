package plugins

import (
	"sync"
	"time"
)

// EventHandler is the function signature for event subscribers.
type EventHandler func(event BusEvent) error

// BusEvent is an event delivered through the EventBus.
type BusEvent struct {
	// Type is the event type string (e.g. "issue.created").
	Type string
	// Payload is arbitrary data attached to the event.
	Payload map[string]interface{}
	// CompanyUUID is the company this event belongs to.
	CompanyUUID string
	// Timestamp is when the event occurred.
	Timestamp time.Time
}

type subscription struct {
	id      uint64
	handler EventHandler
}

// EventBus is an async, in-process event bus with a bounded worker pool.
// Events are dispatched to all matching subscribers concurrently.
type EventBus struct {
	mu   sync.RWMutex
	subs map[string][]subscription // key: eventType

	nextID uint64

	// Worker pool
	workerCount int
	queue       chan busJob
	wg          sync.WaitGroup
	stopCh      chan struct{}
}

type busJob struct {
	handlers []EventHandler
	event    BusEvent
}

// NewEventBus creates an EventBus with the given worker pool size and buffer.
// Call Start() before publishing events and Stop() to drain and shut down.
func NewEventBus(workers, queueBuffer int) *EventBus {
	if workers <= 0 {
		workers = 4
	}
	if queueBuffer <= 0 {
		queueBuffer = 256
	}
	return &EventBus{
		subs:        make(map[string][]subscription),
		workerCount: workers,
		queue:       make(chan busJob, queueBuffer),
		stopCh:      make(chan struct{}),
	}
}

// Start launches the background worker goroutines.
func (b *EventBus) Start() {
	for i := 0; i < b.workerCount; i++ {
		b.wg.Add(1)
		go b.worker()
	}
}

// Stop drains in-flight events and shuts down the workers.
func (b *EventBus) Stop() {
	close(b.stopCh)
	close(b.queue)
	b.wg.Wait()
}

func (b *EventBus) worker() {
	defer b.wg.Done()
	for job := range b.queue {
		for _, h := range job.handlers {
			_ = h(job.event) // errors are swallowed per-handler to isolate failures
		}
	}
}

// Subscribe registers handler for the given eventType. Returns a subscription ID
// that can be passed to Unsubscribe.
func (b *EventBus) Subscribe(eventType string, handler EventHandler) uint64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.nextID++
	id := b.nextID
	b.subs[eventType] = append(b.subs[eventType], subscription{id: id, handler: handler})
	return id
}

// Unsubscribe removes the subscription with the given ID from eventType.
func (b *EventBus) Unsubscribe(eventType string, id uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subs := b.subs[eventType]
	filtered := subs[:0]
	for _, s := range subs {
		if s.id != id {
			filtered = append(filtered, s)
		}
	}
	b.subs[eventType] = filtered
}

// Publish enqueues event for async delivery to all subscribers of event.Type.
// If the internal queue is full the event is dropped (non-blocking).
func (b *EventBus) Publish(event BusEvent) {
	b.mu.RLock()
	subs := b.subs[event.Type]
	if len(subs) == 0 {
		b.mu.RUnlock()
		return
	}
	// Copy handlers under the read lock so Unsubscribe can run concurrently.
	handlers := make([]EventHandler, len(subs))
	for i, s := range subs {
		handlers[i] = s.handler
	}
	b.mu.RUnlock()

	select {
	case b.queue <- busJob{handlers: handlers, event: event}:
	default:
		// Queue full: drop event to avoid blocking callers.
	}
}
