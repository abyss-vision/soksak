// Package plugins implements the abyss-view plugin subsystem: loading,
// registry, event delivery, state persistence, job scheduling, and Node.js bridging.
package plugins

import (
	"fmt"
	"sync"
)

// pluginEntry holds a loaded plugin alongside its metadata.
type pluginEntry struct {
	name   string
	loader *PluginLoader
	info   pluginInfo
}

type pluginInfo struct {
	version            string
	description        string
	eventSubscriptions []string
}

// PluginRegistry keeps track of all loaded plugins and allows lookup by name
// or by subscribed event type.
type PluginRegistry struct {
	mu      sync.RWMutex
	plugins map[string]*pluginEntry // key: plugin name
}

// NewPluginRegistry creates an empty PluginRegistry.
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins: make(map[string]*pluginEntry),
	}
}

// Register adds a loaded plugin to the registry.
// Returns an error if a plugin with the same name is already registered.
func (r *PluginRegistry) Register(name, version, description string, eventSubscriptions []string, loader *PluginLoader) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin already registered: %s", name)
	}

	r.plugins[name] = &pluginEntry{
		name:   name,
		loader: loader,
		info: pluginInfo{
			version:            version,
			description:        description,
			eventSubscriptions: eventSubscriptions,
		},
	}
	return nil
}

// Unregister removes a plugin from the registry.
// Returns an error if the plugin is not found.
func (r *PluginRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}
	delete(r.plugins, name)
	return nil
}

// RegistryEntry is the public view of a registered plugin.
type RegistryEntry struct {
	Name               string
	Version            string
	Description        string
	EventSubscriptions []string
}

// Get returns the entry for a plugin by name.
func (r *PluginRegistry) Get(name string) (RegistryEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	e, ok := r.plugins[name]
	if !ok {
		return RegistryEntry{}, false
	}
	return RegistryEntry{
		Name:               e.name,
		Version:            e.info.version,
		Description:        e.info.description,
		EventSubscriptions: e.info.eventSubscriptions,
	}, true
}

// List returns all registered plugin entries.
func (r *PluginRegistry) List() []RegistryEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := make([]RegistryEntry, 0, len(r.plugins))
	for _, e := range r.plugins {
		entries = append(entries, RegistryEntry{
			Name:               e.name,
			Version:            e.info.version,
			Description:        e.info.description,
			EventSubscriptions: e.info.eventSubscriptions,
		})
	}
	return entries
}

// GetByEvent returns all plugin entries that have subscribed to the given event type.
func (r *PluginRegistry) GetByEvent(eventType string) []RegistryEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []RegistryEntry
	for _, e := range r.plugins {
		for _, sub := range e.info.eventSubscriptions {
			if sub == eventType {
				results = append(results, RegistryEntry{
					Name:               e.name,
					Version:            e.info.version,
					Description:        e.info.description,
					EventSubscriptions: e.info.eventSubscriptions,
				})
				break
			}
		}
	}
	return results
}
