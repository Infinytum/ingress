package pipelines

import (
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/infinytum/ingress/pkg/pipelines/ingress"
	"github.com/infinytum/injector"
	"github.com/infinytum/reactive"
	networkingv1 "k8s.io/api/networking/v1"
)

func init() {
	injector.Singleton(ingressFactory)
}

type Ingress struct {
	pipeline reactive.Subjectable
}

func (i *Ingress) Configure(ing *networkingv1.Ingress) {
	i.Emit(&ingress.GlobalContext{
		Context: ingress.Context{
			Ingress: ing,
			Mode:    ingress.ContextModeConfigure,
		},
		Hosts:  make([]string, 0),
		Routes: make([]*caddyhttp.Route, 0),
	})
}

func (i *Ingress) Delete(ing *networkingv1.Ingress) {
	i.Emit(&ingress.GlobalContext{
		Context: ingress.Context{
			Ingress: ing,
			Mode:    ingress.ContextModeDelete,
		},
		Hosts:  make([]string, 0),
		Routes: make([]*caddyhttp.Route, 0),
	})
}

func (i *Ingress) Emit(ctx *ingress.GlobalContext) {
	i.pipeline.Next(ctx, make([]error, 0))
}

func ingressFactory() *Ingress {
	i := &Ingress{
		pipeline: reactive.NewSubject(),
	}

	// Global Pipeline (Processes whole ingress objects)
	i.pipeline.Pipe(
		ingress.GlobalDiffuser(i.pipeline), // Calls the specific pipeline
		ingress.GlobalApply(),
		ingress.GlobalCustomTLS(),
		ingress.GlobalStatus(),
	)

	// Specific Pipeline (Processes individual paths inside ingress objects)
	i.pipeline.Pipe(
		ingress.SpecificMarker(),
		ingress.SpecificMatcher(),
		ingress.SpecificHeaders(),
		ingress.SpecificBasicAuth(),
		ingress.SpecificRewriteTarget(),
		ingress.SpecificReverseProxy(), // Must be last, because it creates the final route
	)

	return i
}
