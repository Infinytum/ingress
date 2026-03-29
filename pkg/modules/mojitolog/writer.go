package mojitolog

import (
	"encoding/json"
	"fmt"

	"log/slog"
)

type MojitoWriter struct {
}

func (m *MojitoWriter) Write(p []byte) (n int, err error) {
	rawLog := make(map[string]interface{})
	if err := json.Unmarshal(p, &rawLog); err != nil {
		fmt.Println(string(p), err)
		return 0, err
	}

	if filter(rawLog) {
		return len(p), nil
	}

	lvl, msg, rawFields := extractLevelAndMessage(rawLog)
	fields := make(map[string]interface{})
	flatten(rawFields, fields, "")

	// Convert map to slog-style key-value args
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}

	line := slog.With(args...)
	switch lvl {
	case "debug":
		line.Debug(msg)
	case "info":
		line.Info(msg)
	case "warn":
		line.Warn(msg)
	case "error", "panic":
		line.Error(msg)
	case "fatal":
		line.Error(msg)
	default:
		line.Info(msg)
	}

	return len(p), nil
}

func (m *MojitoWriter) Close() error {
	return nil
}
