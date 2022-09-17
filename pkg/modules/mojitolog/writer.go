package mojitolog

import (
	"encoding/json"
	"fmt"

	"github.com/go-mojito/mojito/log"
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

	line := log.Fields(fields)
	switch lvl {
	case "debug":
		line.Debug(msg)
	case "info":
		line.Info(msg)
	case "warn":
		line.Warn(msg)
	case "error":
	case "panic":
		line.Error(msg)
	case "fatal":
		line.Fatal(msg)
	default:
		line.Info(msg)
	}

	return len(p), nil
}

func (m *MojitoWriter) Close() error {
	return nil
}
