package config

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
)

var config Config

func init() {

	config = Config{
		Admin: caddy.AdminConfig{},
		Storage: Storage{
			System: "kubestore",
		},
		Logging: caddy.Logging{
			Logs: map[string]*caddy.CustomLog{
				"default": {
					Level: "WARN",
				},
			},
		},
		Apps: map[string]interface{}{
			"tls": &caddytls.TLS{CertificatesRaw: caddy.ModuleMap{}},
			"http": &caddyhttp.App{
				Servers: map[string]*caddyhttp.Server{
					"http_server": {
						AutoHTTPS: &caddyhttp.AutoHTTPSConfig{},
						Listen:    []string{":80", ":443"},
						TLSConnPolicies: caddytls.ConnectionPolicies{
							&caddytls.ConnectionPolicy{},
						},
					},
				},
			},
		},
	}
}

type Config struct {
	Admin   caddy.AdminConfig      `json:"admin,omitempty"`
	Storage Storage                `json:"storage"`
	Apps    map[string]interface{} `json:"apps"`
	Logging caddy.Logging          `json:"logging"`
}

func (c Config) GetHTTPApp() *caddyhttp.Server {
	return c.Apps["http"].(*caddyhttp.App).Servers["http_server"]
}

func (c Config) GetTLSApp() *caddytls.TLS {
	return c.Apps["tls"].(*caddytls.TLS)
}

// Storage represents the certmagic storage configuration.
type Storage struct {
	System string `json:"module"`
}
