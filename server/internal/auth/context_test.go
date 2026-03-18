package auth

import (
	"context"
	"testing"
)

func TestWithActorAndActorFromContext(t *testing.T) {
	actor := Actor{
		Type:    ActorTypeAgent,
		ID:      "agent-1",
		AgentID: "agent-1",
	}

	ctx := WithActor(context.Background(), actor)
	got, ok := ActorFromContext(ctx)
	if !ok {
		t.Fatal("ActorFromContext: expected ok=true, got false")
	}
	if got.Type != ActorTypeAgent {
		t.Errorf("Type = %v, want %v", got.Type, ActorTypeAgent)
	}
	if got.ID != "agent-1" {
		t.Errorf("ID = %q, want %q", got.ID, "agent-1")
	}
}

func TestActorFromContext_Missing(t *testing.T) {
	got, ok := ActorFromContext(context.Background())
	if ok {
		t.Fatal("ActorFromContext on empty context: expected ok=false, got true")
	}
	if got.Type != ActorTypeNone {
		t.Errorf("Type = %v, want %v", got.Type, ActorTypeNone)
	}
}

func TestWithActor_Overwrite(t *testing.T) {
	first := Actor{Type: ActorTypeUser, ID: "user-1"}
	second := Actor{Type: ActorTypeAgent, ID: "agent-2"}

	ctx := WithActor(context.Background(), first)
	ctx = WithActor(ctx, second)

	got, ok := ActorFromContext(ctx)
	if !ok {
		t.Fatal("expected ok=true after overwrite")
	}
	if got.ID != "agent-2" {
		t.Errorf("ID = %q, want %q", got.ID, "agent-2")
	}
}
