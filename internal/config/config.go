package config

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
	"github.com/infinytum/ingress/internal/service"
	"github.com/infinytum/injector"
)

var persist = false
var config Config = Config{
	Admin: caddy.AdminConfig{
		Config: &caddy.ConfigSettings{
			Persist: &persist,
		},
	},
	Storage: Storage{
		System: "kubestore",
	},
	Logging: caddy.Logging{
		Logs: map[string]*caddy.CustomLog{
			"default": {
				WriterRaw:  json.RawMessage(`{"output":"mojito_log"}`),
				EncoderRaw: json.RawMessage(`{"format":"json"}`),
				Level:      "INFO",
			},
		},
	},
	Apps: map[string]interface{}{
		"tls": &caddytls.TLS{CertificatesRaw: caddy.ModuleMap{
			"load_storage": json.RawMessage(`{"pairs": []}`),
		}},
		"http": &caddyhttp.App{
			Servers: map[string]*caddyhttp.Server{
				"https_server": {
					AutoHTTPS: &caddyhttp.AutoHTTPSConfig{
						DisableRedir: true,
					},
					Listen: []string{":8443"},
					TLSConnPolicies: caddytls.ConnectionPolicies{
						&caddytls.ConnectionPolicy{},
					},
				},
				"http_server": {
					AutoHTTPS: &caddyhttp.AutoHTTPSConfig{
						Disabled: true,
					},
					Listen: []string{":8080"},
					Routes: []caddyhttp.Route{
						{
							HandlersRaw: []json.RawMessage{
								json.RawMessage(`{"handler": "headers","response": {"deferred": true,"set": { "Server": ["Infinytum Gate"] }}}`),
								json.RawMessage(`{"handler": "static_response","headers": {"Location": ["https://{http.request.host}{http.request.uri}"]},"status_code": 302}`),
							},
						},
					},
				},
			},
		},
	},
}

func init() {
	injector.MustCall(func(ingressCfg service.IngressConfig) {
		config.GetHTTPApp().ExperimentalHTTP3 = ingressCfg.HTTP3
	})
}
