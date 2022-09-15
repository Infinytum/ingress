package ingress

import (
	"github.com/infinytum/reactive"
)

// SpecificMarker marks the route generated with the ingress uid and generation
func SpecificMarker() reactive.Pipe {
	return SpecificPipe(func(ctx *SpecificContext, errs []error) []error {
		// If there are already errors or the ingress is being deleted, skip this pipe
		if len(errs) > 0 || ctx.Mode == ContextModeDelete {
			return errs
		}

		// By adding this marker, we can later match routes back to the ingress they came from.
		ctx.Route.Group = ctx.RouteIdentifier()
		return errs
	})
}
