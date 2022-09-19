package ingress

import (
	"net/http"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/headers"
	"github.com/infinytum/ingress/internal/service"
	"github.com/infinytum/injector"
	"github.com/infinytum/reactive"
)

// SpecificHeaders applies custom headers to the response.
// For example caching directives, CORS, etc.
func SpecificHeaders() reactive.Pipe {
	ingressConfig := injector.MustInject[service.IngressConfig]()
	return SpecificPipe(func(ctx *SpecificContext, errs []error) []error {
		handler := headers.Handler{
			Response: &headers.RespHeaderOps{
				HeaderOps: &headers.HeaderOps{
					Set: http.Header{
						"Server": []string{"Infinytum Gate"},
					},
				},
				Deferred: true,
			},
		}

		if ingressConfig.EnableHSTS {
			handler.Response.HeaderOps.Set.Set("Strict-Transport-Security", "max-age=31536000")
		}

		ctx.Route.HandlersRaw = append(ctx.Route.HandlersRaw, caddyconfig.JSONModuleObject(
			handler,
			"handler",
			"headers",
			nil,
		))
		return errs
	})
}
