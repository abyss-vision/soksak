package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	hclog "github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"

	"abyss-view/pkg/plugin"
)

// LoadedPlugin represents a running plugin process managed by go-plugin.
type LoadedPlugin struct {
	Name      string
	Path      string
	client    *goplugin.Client
	iface     plugin.PluginInterface
}

// Call dispatches a call to the underlying plugin interface.
func (lp *LoadedPlugin) Interface() plugin.PluginInterface {
	return lp.iface
}

// Kill terminates the plugin process immediately.
func (lp *LoadedPlugin) Kill() {
	lp.client.Kill()
}

// PluginLoader discovers, loads, and unloads go-plugin–based plugins.
type PluginLoader struct {
	mu      sync.Mutex
	plugins map[string]*LoadedPlugin
	logger  hclog.Logger
}

// NewPluginLoader creates a PluginLoader with a no-op logger. Pass a custom
// hclog.Logger via WithLogger if you want log output.
func NewPluginLoader() *PluginLoader {
	return &PluginLoader{
		plugins: make(map[string]*LoadedPlugin),
		logger:  hclog.NewNullLogger(),
	}
}

// WithLogger replaces the internal logger.
func (l *PluginLoader) WithLogger(logger hclog.Logger) *PluginLoader {
	l.logger = logger
	return l
}

// Discover scans dir for executable files and returns their paths.
// Sub-directories are not traversed.
func (l *PluginLoader) Discover(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("discover plugins in %s: %w", dir, err)
	}

	var paths []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		// Accept any executable file.
		if info.Mode()&0o111 != 0 {
			paths = append(paths, filepath.Join(dir, entry.Name()))
		}
	}
	return paths, nil
}

// Load starts a plugin process at path and returns the loaded plugin.
// The name must be unique; Load returns an error if a plugin with that name
// is already loaded.
func (l *PluginLoader) Load(name, path string) (*LoadedPlugin, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, exists := l.plugins[name]; exists {
		return nil, fmt.Errorf("plugin already loaded: %s", name)
	}

	client := goplugin.NewClient(&goplugin.ClientConfig{
		HandshakeConfig: plugin.Handshake,
		Plugins:         plugin.PluginMap,
		Cmd:             commandForPath(path),
		Logger:          l.logger.Named(name),
		AllowedProtocols: []goplugin.Protocol{
			goplugin.ProtocolNetRPC,
		},
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("connect to plugin %s: %w", name, err)
	}

	raw, err := rpcClient.Dispense("plugin")
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("dispense plugin %s: %w", name, err)
	}

	iface, ok := raw.(plugin.PluginInterface)
	if !ok {
		client.Kill()
		return nil, fmt.Errorf("plugin %s does not implement PluginInterface", name)
	}

	lp := &LoadedPlugin{
		Name:   name,
		Path:   path,
		client: client,
		iface:  iface,
	}
	l.plugins[name] = lp
	return lp, nil
}

// Unload calls OnUnload on the plugin and then kills the process.
// Returns an error if the plugin is not found.
func (l *PluginLoader) Unload(name string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	lp, exists := l.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not loaded: %s", name)
	}

	// Best-effort graceful shutdown.
	_ = lp.iface.OnUnload()
	lp.client.Kill()
	delete(l.plugins, name)
	return nil
}

// Get returns the loaded plugin by name, or nil if not found.
func (l *PluginLoader) Get(name string) *LoadedPlugin {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.plugins[name]
}
