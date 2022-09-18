package config

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
	"github.com/go-mojito/mojito/log"
)

type Config struct {
	Admin   caddy.AdminConfig      `json:"admin,omitempty"`
	Storage Storage                `json:"storage"`
	Apps    map[string]interface{} `json:"apps"`
	Logging caddy.Logging          `json:"logging"`
}

func (c Config) GetHTTPApp() *caddyhttp.Server {
	return c.Apps["http"].(*caddyhttp.App).Servers["https_server"]
}

func (c Config) GetTLSApp() *caddytls.TLS {
	return c.Apps["tls"].(*caddytls.TLS)
}

func (c Config) GetTLSCertificates() (t TLSCertificates) {
	tlsApp := c.GetTLSApp()
	if err := json.Unmarshal(tlsApp.CertificatesRaw["load_storage"], &t); err != nil {
		log.Errorf("Failed to unmarshal TLS Certificates configuration: %v", err)
	}
	return
}

func (c Config) SetTLSCertificates(t TLSCertificates) {
	tlsApp := c.GetTLSApp()
	data, err := json.Marshal(t)
	if err != nil {
		log.Errorf("Failed to marshal TLS Certificates configuration: %v", err)
		return
	}
	tlsApp.CertificatesRaw["load_storage"] = data
}

// Storage represents the certmagic storage configuration.
type Storage struct {
	System string `json:"module"`
}

// Covers apps.tls.certificates.load_storage
type TLSCertificates struct {
	Pairs []TLSCertificatePair `json:"pairs"`
}

// Covers apps.tls.certificates.load_storage.pairs
type TLSCertificatePair struct {
	Certificate string   `json:"certificate"`
	Key         string   `json:"key"`
	Format      string   `json:"format"`
	Tags        []string `json:"tags"`
}
