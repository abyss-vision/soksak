package realtime

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	h := NewHub()
	if h == nil {
		t.Fatal("NewHub returned nil")
	}
	if h.clients == nil {
		t.Error("clients map is nil")
	}
	if h.rooms == nil {
		t.Error("rooms map is nil")
	}
}

func TestHub_RegisterUnregister(t *testing.T) {
	h := NewHub()
	go h.Run()

	client := &Client{
		companyID: "company-1",
		actorType: "agent",
		actorID:   "agent-1",
		subs:      make(map[string]bool),
		send:      make(chan []byte, 16),
	}

	h.Register(client)
	time.Sleep(20 * time.Millisecond)

	if _, ok := h.clients[client]; !ok {
		t.Error("client not registered in h.clients")
	}
	if _, ok := h.rooms["company-1"]; !ok {
		t.Error("room not created for company-1")
	}

	h.Unregister(client)
	time.Sleep(20 * time.Millisecond)

	if _, ok := h.clients[client]; ok {
		t.Error("client still in h.clients after unregister")
	}
	// Room should be cleaned up when empty.
	if _, ok := h.rooms["company-1"]; ok {
		t.Error("room should be deleted when empty")
	}
}

func TestHub_PublishToCompany(t *testing.T) {
	h := NewHub()
	go h.Run()

	client := &Client{
		companyID: "company-2",
		actorType: "user",
		actorID:   "user-1",
		subs:      make(map[string]bool),
		send:      make(chan []byte, 16),
	}

	h.Register(client)
	time.Sleep(20 * time.Millisecond)

	msg := WebSocketMessage{
		Type:      EventIssueUpdated,
		CompanyID: "company-2",
		Payload:   json.RawMessage(`{"id":"issue-1"}`),
	}
	h.PublishToCompany("company-2", msg)
	time.Sleep(20 * time.Millisecond)

	select {
	case data := <-client.send:
		var received WebSocketMessage
		if err := json.Unmarshal(data, &received); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if received.Type != EventIssueUpdated {
			t.Errorf("Type = %q, want %q", received.Type, EventIssueUpdated)
		}
	default:
		t.Error("no message received on client.send channel")
	}

	h.Unregister(client)
}

func TestHub_PublishToCompany_SetsTimestamp(t *testing.T) {
	h := NewHub()
	go h.Run()

	client := &Client{
		companyID: "company-3",
		send:      make(chan []byte, 4),
		subs:      make(map[string]bool),
	}
	h.Register(client)
	time.Sleep(20 * time.Millisecond)

	// Zero timestamp should be set automatically.
	msg := WebSocketMessage{Type: "test.event"}
	h.PublishToCompany("company-3", msg)
	time.Sleep(20 * time.Millisecond)

	select {
	case data := <-client.send:
		var received WebSocketMessage
		if err := json.Unmarshal(data, &received); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if received.Timestamp.IsZero() {
			t.Error("Timestamp should be set automatically when zero")
		}
	default:
		t.Error("no message received")
	}

	h.Unregister(client)
}

func TestHub_PublishToCompany_UnknownCompany(t *testing.T) {
	h := NewHub()
	go h.Run()

	// Publish to a company with no clients — should not block or panic.
	msg := WebSocketMessage{Type: "test.event", CompanyID: "nonexistent"}
	h.PublishToCompany("nonexistent", msg)
	time.Sleep(20 * time.Millisecond)
}

func TestHub_Register_Unregister_NonexistentClient(t *testing.T) {
	h := NewHub()
	go h.Run()

	// Unregistering a client that was never registered should be a no-op.
	client := &Client{
		companyID: "company-x",
		send:      make(chan []byte, 4),
		subs:      make(map[string]bool),
	}
	h.Unregister(client)
	time.Sleep(20 * time.Millisecond)
}
