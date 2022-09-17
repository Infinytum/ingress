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
	return false
}
