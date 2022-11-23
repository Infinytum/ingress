package config

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
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
	Apps: ConfigApps{
		TLS: &caddytls.TLS{CertificatesRaw: caddy.ModuleMap{
			"load_storage": json.RawMessage(`{"pairs": []}`),
		}},
		HTTP: &caddyhttp.App{
			Servers: map[string]*caddyhttp.Server{
				"https_server": {
					AutoHTTPS: &caddyhttp.AutoHTTPSConfig{
						DisableRedir: true,
					},
					Listen:  []string{":8443"},
					Metrics: &caddyhttp.Metrics{},
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
				"metrics": {
					AutoHTTPS: &caddyhttp.AutoHTTPSConfig{
						Disabled: true,
					},
					Listen: []string{":9765"},
					Routes: []caddyhttp.Route{
						{
							HandlersRaw: []json.RawMessage{json.RawMessage(`{ "handler": "metrics" }`)},
							MatcherSetsRaw: []caddy.ModuleMap{{
								"path": caddyconfig.JSON(caddyhttp.MatchPath{"/metrics"}, nil),
							}},
						},
					},
				},
			},
		},
	},
}
