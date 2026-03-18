package plugins

import (
	"testing"
)

func TestPluginRegistry_RegisterAndGet(t *testing.T) {
	r := NewPluginRegistry()

	err := r.Register("my-plugin", "1.0.0", "A test plugin", []string{"issue.created"}, nil)
	if err != nil {
		t.Fatalf("Register: unexpected error: %v", err)
	}

	entry, ok := r.Get("my-plugin")
	if !ok {
		t.Fatal("Get: expected ok=true, got false")
	}
	if entry.Name != "my-plugin" {
		t.Errorf("Name = %q, want %q", entry.Name, "my-plugin")
	}
	if entry.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", entry.Version, "1.0.0")
	}
	if entry.Description != "A test plugin" {
		t.Errorf("Description = %q, want %q", entry.Description, "A test plugin")
	}
	if len(entry.EventSubscriptions) != 1 || entry.EventSubscriptions[0] != "issue.created" {
		t.Errorf("EventSubscriptions = %v, want [issue.created]", entry.EventSubscriptions)
	}
}

func TestPluginRegistry_Register_Duplicate(t *testing.T) {
	r := NewPluginRegistry()
	_ = r.Register("plugin-a", "1.0", "desc", nil, nil)
	err := r.Register("plugin-a", "2.0", "desc2", nil, nil)
	if err == nil {
		t.Fatal("Register duplicate: expected error, got nil")
	}
}

func TestPluginRegistry_Get_Missing(t *testing.T) {
	r := NewPluginRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Fatal("Get missing: expected ok=false, got true")
	}
}

func TestPluginRegistry_Unregister(t *testing.T) {
	r := NewPluginRegistry()
	_ = r.Register("plugin-b", "1.0", "desc", nil, nil)

	err := r.Unregister("plugin-b")
	if err != nil {
		t.Fatalf("Unregister: unexpected error: %v", err)
	}

	_, ok := r.Get("plugin-b")
	if ok {
		t.Fatal("Get after Unregister: expected ok=false, got true")
	}
}

func TestPluginRegistry_Unregister_Missing(t *testing.T) {
	r := NewPluginRegistry()
	err := r.Unregister("ghost-plugin")
	if err == nil {
		t.Fatal("Unregister missing: expected error, got nil")
	}
}

func TestPluginRegistry_List(t *testing.T) {
	r := NewPluginRegistry()

	entries := r.List()
	if len(entries) != 0 {
		t.Errorf("List on empty registry: expected 0 entries, got %d", len(entries))
	}

	_ = r.Register("p1", "1.0", "d1", []string{"ev.a"}, nil)
	_ = r.Register("p2", "2.0", "d2", []string{"ev.b"}, nil)

	entries = r.List()
	if len(entries) != 2 {
		t.Errorf("List: expected 2 entries, got %d", len(entries))
	}
}

func TestPluginRegistry_GetByEvent(t *testing.T) {
	r := NewPluginRegistry()
	_ = r.Register("p1", "1.0", "d1", []string{"issue.created", "issue.updated"}, nil)
	_ = r.Register("p2", "1.0", "d2", []string{"issue.deleted"}, nil)

	results := r.GetByEvent("issue.created")
	if len(results) != 1 {
		t.Fatalf("GetByEvent(issue.created): expected 1 result, got %d", len(results))
	}
	if results[0].Name != "p1" {
		t.Errorf("GetByEvent result name = %q, want %q", results[0].Name, "p1")
	}

	results = r.GetByEvent("issue.deleted")
	if len(results) != 1 {
		t.Fatalf("GetByEvent(issue.deleted): expected 1 result, got %d", len(results))
	}

	results = r.GetByEvent("issue.nonexistent")
	if len(results) != 0 {
		t.Errorf("GetByEvent(nonexistent): expected 0 results, got %d", len(results))
	}
}
