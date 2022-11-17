package configmap

import (
	"encoding/json"

	"github.com/infinytum/ingress/internal/config"
	"github.com/infinytum/reactive"
)

func Proxy() reactive.Pipe {
	return Pipe(func(ctx *Context, errs []error) []error {
		config.Edit(func(config *config.Config) {
			if ctx.ProxyProtocol {
				allowedIps, err := json.Marshal(ctx.ProxyProtocolAllowedIPs)
				if err != nil {
					errs = append(errs, err)
					return
				}

				config.GetHTTPApp().ListenerWrappersRaw = []json.RawMessage{
					json.RawMessage(`{"wrapper": "proxy_protocol", "timeout": "5s", "allow": ` + string(allowedIps) + `}`),
					json.RawMessage(`{"wrapper": "tls"}`),
				}

				config.GetRedirApp().ListenerWrappersRaw = []json.RawMessage{
					json.RawMessage(`{"wrapper": "proxy_protocol", "timeout": "5s", "allow": ` + string(allowedIps) + `}`),
					json.RawMessage(`{"wrapper": "tls"}`),
				}
			}
		})

		return errs
	})
}
