package ingress

import (
	"errors"
	"fmt"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/reverseproxy"
	"github.com/infinytum/ingress/internal/annotations"
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

		transport := &reverseproxy.HTTPTransport{
			KeepAlive: &reverseproxy.KeepAlive{
				MaxIdleConnsPerHost: 1024,
			},
		}

		// Configure backend protocol, if supported
		backend := annotations.GetAnnotationOrDefault(ctx.Ingress.ObjectMeta, annotations.AnnotationBackendProtocol, string(BackendHTTP))
		switch SupportedBackend(backend) {
		case BackendHTTP: // Do nothing
			break
		case BackendHTTPS:
			transport.TLS = &reverseproxy.TLSConfig{
				InsecureSkipVerify: annotations.GetAnnotationBool(ctx.Ingress.ObjectMeta, annotations.AnnotationInsecureSkipVerify, true),
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
				LoadBalancing: &reverseproxy.LoadBalancing{
					TryDuration: caddy.Duration(time.Second * 5),
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
