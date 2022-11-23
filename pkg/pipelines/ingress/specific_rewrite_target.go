package ingress

import (
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/infinytum/ingress/internal/annotations"
	"github.com/infinytum/reactive"
)

type Uri struct {
	Regexp []UriRegexp `json:"path_regexp"`
}

type UriRegexp struct {
	Find    string `json:"find"`
	Replace string `json:"replace"`
}

// SpecificRewriteTarget configures the route to rewrite the backend's path
func SpecificRewriteTarget() reactive.Pipe {
	return SpecificPipe(func(ctx *SpecificContext, errs []error) []error {
		// If there are already errors or the ingress is being deleted, skip this pipe
		if len(errs) > 0 || ctx.Mode == ContextModeDelete {
			return errs
		}

		// Configure rewrite target
		rewriteTarget := annotations.GetAnnotationOrDefault(ctx.Ingress.ObjectMeta, annotations.AnnotationRewriteTarget, "")
		if rewriteTarget == "" {
			return errs
		}

		handlerModule := caddyconfig.JSONModuleObject(
			Uri{
				Regexp: []UriRegexp{
					{
						Find:    ctx.Path.Path,
						Replace: rewriteTarget,
					},
				},
			},
			"handler",
			"rewrite",
			nil,
		)
		ctx.Route.HandlersRaw = append(ctx.Route.HandlersRaw, handlerModule)
		return errs
	})
}
