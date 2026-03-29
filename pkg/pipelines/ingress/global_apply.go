package ingress

import (
	"strings"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"log/slog"
	"github.com/infinytum/ingress/internal/config"
	"github.com/infinytum/ingress/internal/service"
	"github.com/infinytum/injector"
	"github.com/infinytum/reactive"
)

// GlobalApply applies the routes to the caddy server
func GlobalApply() reactive.Pipe {
	return GlobalPipe(func(ctx *GlobalContext, errs []error) []error {

		if len(errs) > 0 {
			for _, err := range errs {
				slog.Error("Error while configuring global ingress", "error", err)
			}
			return errs
		}

		err := config.Edit(func(c *config.Config) {
			app := c.GetHTTPApp()

			routes := make([]caddyhttp.Route, 0)
			routesExist := false
			for _, route := range app.Routes {
				if !routesExist && strings.HasPrefix(route.Group, string(ctx.Ingress.UID)) {
					if !ctx.IsNewer(route.Group) {
						routesExist = true
						routes = append(routes, route)
					}
				} else {
					routes = append(routes, route)
				}
			}

			app.Routes = routes
			if !routesExist && ctx.Mode == ContextModeConfigure {
				slog.With("name", ctx.Ingress.Name, "namespace", ctx.Ingress.Namespace).Info("Reconfigured routes")
				for _, route := range ctx.Routes {
					app.Routes = append(app.Routes, *route)
				}
			}
		})

		if err != nil {
			slog.Error("Error while applying ingress", "name", ctx.Ingress.Name, "namespace", ctx.Ingress.Namespace, "error", err)
			errs = append(errs, err)
			return errs
		}

		injector.Call(func(state *service.State) {
			delete(state.ConfiguredHosts, string(ctx.Ingress.UID))
			state.ConfiguredHosts[string(ctx.Ingress.UID)] = ctx.Hosts
		})

		return errs
	})
}
