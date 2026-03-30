package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

const (
	// InternalHTTPSAddr is the internal address the HTTPS server listens on
	// when layer4 is active and handling the external port.
	InternalHTTPSAddr = "127.0.0.1:18443"

	// ExternalHTTPSAddr is the normal external HTTPS listen address.
	ExternalHTTPSAddr = ":8443"
)

// Layer4Config represents the caddy-l4 layer4 app configuration.
type Layer4Config struct {
	Servers map[string]*Layer4Server `json:"servers"`
}

// Layer4Server represents a layer4 server.
type Layer4Server struct {
	Listen []string      `json:"listen"`
	Routes []Layer4Route `json:"routes,omitempty"`
}

// Layer4Route represents a layer4 route with matchers and handlers.
type Layer4Route struct {
	Match  []json.RawMessage `json:"match,omitempty"`
	Handle []json.RawMessage `json:"handle,omitempty"`
}

// PassthroughEntry represents a single SSL passthrough host and its backend.
type PassthroughEntry struct {
	Host string
	Dial string // e.g. "backend.namespace.svc.cluster.local:443"
}

// BuildPassthroughRoute builds a layer4 route that matches on TLS SNI
// and proxies the raw TCP connection to the backend.
func BuildPassthroughRoute(host, dial string) Layer4Route {
	return Layer4Route{
		Match: []json.RawMessage{
			json.RawMessage(fmt.Sprintf(`{"tls":{"sni":[%q]}}`, host)),
		},
		Handle: []json.RawMessage{
			json.RawMessage(fmt.Sprintf(`{"handler":"proxy","upstreams":[{"dial":[%q]}]}`, dial)),
		},
	}
}

// BuildCatchAllRoute builds the catch-all layer4 route that forwards the raw
// TLS connection to the internal HTTP server, which handles TLS termination itself.
func BuildCatchAllRoute() Layer4Route {
	return Layer4Route{
		Handle: []json.RawMessage{
			json.RawMessage(fmt.Sprintf(`{"handler":"proxy","upstreams":[{"dial":[%q]}]}`, InternalHTTPSAddr)),
		},
	}
}

// GetLayer4Config returns the current layer4 config, or nil if not set.
func (c Config) GetLayer4Config() *Layer4Config {
	if c.Apps.Layer4 == nil {
		return nil
	}
	var l4 Layer4Config
	if err := json.Unmarshal(c.Apps.Layer4, &l4); err != nil {
		slog.Error("Failed to unmarshal layer4 config", "error", err)
		return nil
	}
	return &l4
}

// SetLayer4Config sets the layer4 config, or removes it if nil.
func (c *Config) SetLayer4Config(l4 *Layer4Config) {
	if l4 == nil {
		c.Apps.Layer4 = nil
		return
	}
	data, err := json.Marshal(l4)
	if err != nil {
		slog.Error("Failed to marshal layer4 config", "error", err)
		return
	}
	c.Apps.Layer4 = data
}
