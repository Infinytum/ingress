package ingress

import (
	"net/http"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/headers"
	"github.com/infinytum/reactive"
)

// SpecificHeaders applies custom headers to the response.
// For example caching directives, CORS, etc.
func SpecificHeaders() reactive.Pipe {
	return SpecificPipe(func(ctx *SpecificContext, errs []error) []error {
		// Create the reverse proxy handler
		// Uses the cluster DNS to resolve backend pods
		handlerModule := caddyconfig.JSONModuleObject(
			headers.Handler{
				Response: &headers.RespHeaderOps{
					HeaderOps: &headers.HeaderOps{
						Set: http.Header{
							"Server": []string{"Infinytum Gate"},
						},
					},
					Deferred: true,
				},
			},
			"handler",
			"headers",
			nil,
		)
		ctx.Route.HandlersRaw = append(ctx.Route.HandlersRaw, handlerModule)
		return errs
	})
}
