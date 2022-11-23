package mojitolog

func filter(msg map[string]interface{}) bool {
	// Only does HTTP -> HTTPS redirs, clutters log with useless warnings
	if val, ok := msg["server_name"]; ok && val == "http_server" {
		return true
	}

	// Hide misleading caddy warning, this is intentional and not a bug
	if val, ok := msg["server_name"]; ok && val == "https_server" {
		if val, ok := msg["msg"]; ok && val == "automatic HTTP->HTTPS redirects are disabled" {
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
	return false
}
