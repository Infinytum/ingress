package ingress

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/reverseproxy"
	"github.com/infinytum/ingress/internal/annotations"
	"github.com/infinytum/injector"
	"github.com/infinytum/reactive"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
				MaxIdleConnsPerHost: annotations.GetAnnotationInt(ctx.Ingress.ObjectMeta, annotations.AnnotationKeepAlive, 1024),
			},
			Versions: annotations.GetAnnotationList(ctx.Ingress.ObjectMeta, annotations.AnnotationProxyHTTPVersion, []string{"1.1", "2"}),
		}

		// Configure backend protocol, if supported
		backend := annotations.GetAnnotationOrDefault(ctx.Ingress.ObjectMeta, annotations.AnnotationBackendProtocol, string(BackendHTTP))
		switch SupportedBackend(strings.ToLower(backend)) {
		case BackendHTTP: // Do nothing
			break
		case BackendHTTPS:
			transport.TLS = &reverseproxy.TLSConfig{
				InsecureSkipVerify: annotations.GetAnnotationBool(ctx.Ingress.ObjectMeta, annotations.AnnotationInsecureSkipVerify, true),
			}
		default:
			return append(errs, errors.New("Unsupported backend protocol: "+backend))
		}

		if ctx.Path.Backend.Service.Port.Number == 0 {
			injector.Call(func(client *kubernetes.Clientset) {
				srv, err := client.CoreV1().Services(ctx.Ingress.Namespace).Get(context.Background(), ctx.Path.Backend.Service.Name, metav1.GetOptions{})
				if err != nil {
					return
				}
				for _, port := range srv.Spec.Ports {
					if port.Name == ctx.Path.Backend.Service.Port.Name {
						ctx.Path.Backend.Service.Port.Number = port.Port
					}
				}
			})
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
					TryDuration: caddy.Duration(time.Second * time.Duration(annotations.GetAnnotationInt(ctx.Ingress.ObjectMeta, annotations.AnnotationProxyNextUpstreamTimeout, 5))),
				},
				TrustedProxies: annotations.GetAnnotationList(ctx.Ingress.ObjectMeta, annotations.AnnotationTrustedProxies, []string{}),
			},
			"handler",
			"reverse_proxy",
			nil,
		)
		ctx.Route.HandlersRaw = append(ctx.Route.HandlersRaw, handlerModule)
		return errs
	})
}
