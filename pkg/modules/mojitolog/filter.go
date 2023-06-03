package mojitolog

import (
	"fmt"
	"strings"
)

func filter(msg map[string]interface{}) bool {
	// Only does HTTP -> HTTPS redirs, clutters log with useless warnings
	if val, ok := msg["server_name"]; ok && val == "http_server" {
		return true
	}

	if val, ok := msg["server_name"]; ok && val == "https_server" {
		// Hide misleading caddy warning, this is intentional and not a bug
		if val, ok := msg["msg"]; ok && val == "automatic HTTP->HTTPS redirects are disabled" {
			return true
		}

		// Hide spammy message about skipping automatic cert management
		// E.g: skipping automatic certificate management because one or more matching certificates are already loaded
		if val, ok := msg["msg"]; ok && strings.Contains(fmt.Sprint(val), "skipping automatic certificate management") {
			return true
		}
	}

	if val, ok := msg["server_name"]; ok && val == "metrics" {
		if val, ok := msg["msg"]; ok && val == "automatic HTTPS is completely disabled for server" {
			return true
		}
	}

	// Disable useless tls start stop logs
	if val, ok := msg["logger"]; ok && val == "tls.cache.maintenance" {
		return true
	}

	if val, ok := msg["logger"]; ok && val == "tls" {
		if val, ok := msg["error"]; ok && strings.Contains(fmt.Sprint(val), "cloudflare origin certificate") {
			return true
		}
	}

	// Disable "enable automatic TLS certificate management" since domain list can get long
	if val, ok := msg["logger"]; ok && val == "http" {
		if val, ok := msg["msg"]; ok && strings.Contains(fmt.Sprint(val), "enabling automatic TLS certificate management") {
			return true
		}
	}
	return false
}
