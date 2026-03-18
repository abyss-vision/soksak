package adapter

import "fmt"

// Registry holds all registered ServerAdapter implementations and provides
// thread-safe lookup by adapter name.
//
// Registrations happen at startup (single-threaded), so the map itself does not
// need a mutex — reads during request handling are safe once all registrations
// are complete.
type Registry struct {
	adapters map[string]ServerAdapter
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		adapters: make(map[string]ServerAdapter),
	}
}

// Register adds an adapter to the registry. It panics if an adapter with the
// same name has already been registered, preventing silent misconfiguration.
func (r *Registry) Register(a ServerAdapter) {
	name := a.Name()
	if _, exists := r.adapters[name]; exists {
		panic(fmt.Sprintf("adapter: duplicate registration for %q", name))
	}
	r.adapters[name] = a
}

// Get looks up an adapter by name. The second return value reports whether the
// adapter was found.
func (r *Registry) Get(name string) (ServerAdapter, bool) {
	a, ok := r.adapters[name]
	return a, ok
}

// List returns the names of all registered adapters in an unspecified order.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.adapters))
	for name := range r.adapters {
		names = append(names, name)
	}
	return names
}
