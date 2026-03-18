// Package plugin defines the SDK interfaces and types for the abyss-view plugin system.
// Plugins communicate with the host via the hashicorp/go-plugin gRPC transport.
package plugin

import (
	"net/rpc"
	"time"

	"github.com/hashicorp/go-plugin"
)

// Handshake is the shared configuration used to verify plugin/host compatibility.
// Both host and plugin must use the same magic cookie.
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "ABYSS_PLUGIN",
	MagicCookieValue: "abyss-plugin-v1",
}

// PluginInfo describes a plugin's identity and capabilities.
type PluginInfo struct {
	// Name is the unique plugin key (e.g. "acme.linear").
	Name string
	// Version is the semver version string (e.g. "1.0.0").
	Version string
	// Description is a short human-readable description of the plugin.
	Description string
	// EventSubscriptions lists the event types this plugin wants to receive.
	EventSubscriptions []string
}

// Event is a domain event delivered to a plugin.
type Event struct {
	// Type is the event type (e.g. "issue.created", "plugin.acme.linear.sync-done").
	Type string
	// Payload is arbitrary JSON-serializable data attached to the event.
	Payload map[string]interface{}
	// CompanyUUID is the UUID of the company this event belongs to.
	CompanyUUID string
	// Timestamp is when the event occurred.
	Timestamp time.Time
}

// PluginInterface is the interface that all abyss-view plugins must implement.
// It is used both as the Go interface and as the hashicorp/go-plugin RPC contract.
type PluginInterface interface {
	// GetInfo returns metadata describing the plugin.
	GetInfo() (PluginInfo, error)
	// OnLoad is called once after the plugin process starts and the host has
	// verified compatibility. config is the operator-supplied JSON config blob.
	OnLoad(config map[string]interface{}) error
	// OnUnload is called before the plugin process is terminated. Plugins should
	// flush any in-flight work and release external resources.
	OnUnload() error
	// HandleEvent is called by the host when a subscribed event fires.
	// The plugin should process the event synchronously and return any error.
	HandleEvent(event Event) error
}

// --- hashicorp/go-plugin net/rpc glue ---

// PluginRPC is the client-side RPC proxy that calls into the plugin process.
type PluginRPC struct{ client *rpc.Client }

func (p *PluginRPC) GetInfo() (PluginInfo, error) {
	var resp PluginInfo
	err := p.client.Call("Plugin.GetInfo", new(interface{}), &resp)
	return resp, err
}

func (p *PluginRPC) OnLoad(config map[string]interface{}) error {
	return p.client.Call("Plugin.OnLoad", config, new(interface{}))
}

func (p *PluginRPC) OnUnload() error {
	return p.client.Call("Plugin.OnUnload", new(interface{}), new(interface{}))
}

func (p *PluginRPC) HandleEvent(event Event) error {
	return p.client.Call("Plugin.HandleEvent", event, new(interface{}))
}

// PluginRPCServer is the server-side RPC wrapper that delegates to the real implementation.
type PluginRPCServer struct{ Impl PluginInterface }

func (s *PluginRPCServer) GetInfo(_ interface{}, resp *PluginInfo) error {
	info, err := s.Impl.GetInfo()
	*resp = info
	return err
}

func (s *PluginRPCServer) OnLoad(config map[string]interface{}, _ *interface{}) error {
	return s.Impl.OnLoad(config)
}

func (s *PluginRPCServer) OnUnload(_ interface{}, _ *interface{}) error {
	return s.Impl.OnUnload()
}

func (s *PluginRPCServer) HandleEvent(event Event, _ *interface{}) error {
	return s.Impl.HandleEvent(event)
}

// GRPCPlugin implements hashicorp/go-plugin's Plugin interface using net/rpc.
// Each plugin binary registers this in its main() via plugin.Serve.
type GRPCPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	// Impl is set by the plugin binary; unused on the host side.
	Impl PluginInterface
}

func (p *GRPCPlugin) Server(_ *plugin.MuxBroker) (interface{}, error) {
	return &PluginRPCServer{Impl: p.Impl}, nil
}

func (p *GRPCPlugin) Client(_ *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &PluginRPC{client: c}, nil
}

// PluginMap is the plugin map passed to plugin.NewClient / plugin.Serve.
var PluginMap = map[string]plugin.Plugin{
	"plugin": &GRPCPlugin{},
}
