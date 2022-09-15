package ingress

import (
	"errors"
	"fmt"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/reverseproxy"
	"github.com/infinytum/ingress/pkg/utils"
	"github.com/infinytum/reactive"
)

type SupportedBackend string

const (
	BackendHTTP  SupportedBackend = "http"
	BackendHTTPS SupportedBackend = "https"
)

// SpecificReverseProxy configures the route's backend to be a reverse proxy to the configured ingress service
func SpecificReverseProxy() reactive.Pipe {
	return SpecificPipe(func(ctx *SpecificContext, errs []error) []error {
		// If there are already errors or the ingress is being deleted, skip this pipe
		if len(errs) > 0 || ctx.Mode == ContextModeDelete {
			return errs
		}

		transport := &reverseproxy.HTTPTransport{}

		// Configure backend protocol, if supported
		backend := utils.GetAnnotationOrDefault(ctx.Ingress.ObjectMeta, utils.AnnotationBackendProtocol, string(BackendHTTP))
		switch SupportedBackend(backend) {
		case BackendHTTP: // Do nothing
			break
		case BackendHTTPS:
			transport.TLS = &reverseproxy.TLSConfig{
				InsecureSkipVerify: utils.GetAnnotationBool(ctx.Ingress.ObjectMeta, utils.AnnotationInsecureSkipVerify, true),
			}
		default:
			return append(errs, errors.New("Unsupported backend protocol: "+backend))
		}

		// Create the reverse proxy handler
		// Uses the cluster DNS to resolve backend pods
		handlerModule := caddyconfig.JSONModuleObject(
			reverseproxy.Handler{
				TransportRaw: caddyconfig.JSONModuleObject(transport, "protocol", "http", nil),
				Upstreams: reverseproxy.UpstreamPool{
					{Dial: fmt.Sprintf("%v.%v.svc.cluster.local:%d", ctx.Path.Backend.Service.Name, ctx.Ingress.Namespace, ctx.Path.Backend.Service.Port.Number)},
				},
			},
			"handler",
			"reverse_proxy",
			nil,
		)
		ctx.Route.HandlersRaw = append(ctx.Route.HandlersRaw, handlerModule)
		return errs
	})
}
