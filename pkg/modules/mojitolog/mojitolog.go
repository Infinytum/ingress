package mojitolog

import (
	"io"

	"github.com/caddyserver/caddy/v2"
)

type MojitoLog struct {
}

func (MojitoLog) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "caddy.logging.writers.mojito_log",
		New: func() caddy.Module { return new(MojitoLog) },
	}
}

func (l *MojitoLog) String() string {
	return "mojito_log"
}

func (l *MojitoLog) WriterKey() string {
	return "mojito_log"
}

func (l *MojitoLog) OpenWriter() (io.WriteCloser, error) {
	return &MojitoWriter{}, nil
}
