package ingress

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/infinytum/ingress/pkg/utils"
	"github.com/infinytum/reactive"
	v1 "k8s.io/api/networking/v1"
)

// SpecificMatcher configures the route's matcher to only match the given ingress route and path
func SpecificMatcher() reactive.Pipe {
	return SpecificPipe(func(ctx *SpecificContext, errs []error) []error {
		// If there are already errors or the ingress is being deleted, skip this pipe
		if len(errs) > 0 || ctx.Mode == ContextModeDelete {
			return errs
		}

		modMap := caddy.ModuleMap{}

		// Configure HTTPS-only listener if SSL redirect is enabled
		if utils.GetAnnotationBool(ctx.Ingress.ObjectMeta, utils.AnnotationSSLRedirect, true) {
			modMap["protocol"] = caddyconfig.JSON(caddyhttp.MatchProtocol("https"), nil)
		}

		if ctx.Rule.Host != "" {
			modMap["host"] = caddyconfig.JSON(caddyhttp.MatchHost{ctx.Rule.Host}, nil)
		}

		// Configure the path matcher for the route
		// By default, the path matcher is a prefix matcher
		switch *ctx.Path.PathType {
		case v1.PathTypeImplementationSpecific, v1.PathTypePrefix:
			modMap["path"] = caddyconfig.JSON(caddyhttp.MatchPath{ctx.Path.Path + "*"}, nil)
		case v1.PathTypeExact:
			modMap["path"] = caddyconfig.JSON(caddyhttp.MatchPath{ctx.Path.Path}, nil)
		}

		ctx.Route.MatcherSetsRaw = append(ctx.Route.MatcherSetsRaw, modMap)
		return errs
	})
}
