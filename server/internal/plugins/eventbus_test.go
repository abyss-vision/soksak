package plugins

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestEventBus_SubscribeAndPublish(t *testing.T) {
	bus := NewEventBus(2, 16)
	bus.Start()
	defer bus.Stop()

	var received atomic.Int32
	bus.Subscribe("issue.created", func(e BusEvent) error {
		received.Add(1)
		return nil
	})

	bus.Publish(BusEvent{Type: "issue.created", Payload: map[string]interface{}{"id": "1"}})

	// Give workers time to process.
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if received.Load() == 1 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	if received.Load() != 1 {
		t.Errorf("expected 1 delivery, got %d", received.Load())
	}
}

func TestEventBus_NoSubscribers_EventDropped(t *testing.T) {
	bus := NewEventBus(1, 4)
	bus.Start()
	defer bus.Stop()

	// Publishing with no subscribers should not block or panic.
	bus.Publish(BusEvent{Type: "unknown.event"})
}

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := NewEventBus(2, 16)
	bus.Start()
	defer bus.Stop()

	var count atomic.Int32
	id := bus.Subscribe("order.placed", func(e BusEvent) error {
		count.Add(1)
		return nil
	})

	bus.Publish(BusEvent{Type: "order.placed"})
	time.Sleep(50 * time.Millisecond)

	bus.Unsubscribe("order.placed", id)
	bus.Publish(BusEvent{Type: "order.placed"})
	time.Sleep(50 * time.Millisecond)

	if count.Load() != 1 {
		t.Errorf("expected exactly 1 delivery (before unsubscribe), got %d", count.Load())
	}
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	bus := NewEventBus(2, 32)
	bus.Start()
	defer bus.Stop()

	var c1, c2 atomic.Int32
	bus.Subscribe("ping", func(e BusEvent) error { c1.Add(1); return nil })
	bus.Subscribe("ping", func(e BusEvent) error { c2.Add(1); return nil })

	bus.Publish(BusEvent{Type: "ping"})

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if c1.Load() == 1 && c2.Load() == 1 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	if c1.Load() != 1 || c2.Load() != 1 {
		t.Errorf("expected both subscribers to receive event: c1=%d c2=%d", c1.Load(), c2.Load())
	}
}

func TestNewEventBus_DefaultsWhenZero(t *testing.T) {
	// workers <= 0 and buffer <= 0 should use defaults without panic.
	bus := NewEventBus(0, 0)
	bus.Start()
	defer bus.Stop()

	if bus.workerCount != 4 {
		t.Errorf("workerCount = %d, want 4", bus.workerCount)
	}
}

func TestEventBus_QueueFull_Drop(t *testing.T) {
	// Buffer of 1 so second publish may be dropped (non-blocking).
	bus := NewEventBus(0, 1)
	// Do NOT start workers — queue fills immediately.

	var called atomic.Int32
	bus.Subscribe("flood", func(e BusEvent) error { called.Add(1); return nil })

	// Fill the queue.
	bus.Publish(BusEvent{Type: "flood"})
	// Second publish should drop without blocking.
	bus.Publish(BusEvent{Type: "flood"})

	// Start workers to drain.
	bus.Start()
	time.Sleep(50 * time.Millisecond)
	bus.Stop()

	// At most 1 event was processed (queue size 1).
	if called.Load() > 1 {
		t.Errorf("expected at most 1 call, got %d", called.Load())
	}
}
