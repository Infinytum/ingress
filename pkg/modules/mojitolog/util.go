package mojitolog

func extractLevelAndMessage(log map[string]interface{}) (string, string, map[string]interface{}) {
	level := log["level"]
	message := log["msg"]
	delete(log, "level")
	delete(log, "msg")
	delete(log, "ts") // Mojito provides its own timestamp
	return level.(string), "[Caddy] " + message.(string), log
}

func flatten(m map[string]interface{}, fields map[string]interface{}, prefix string) map[string]interface{} {
	for k, v := range m {
		key := prefix + k

		if v2, ok := v.(map[string]interface{}); ok {
			flatten(v2, fields, key+"_")
		} else {
			fields[key] = v
		}
	}
	return m
}
