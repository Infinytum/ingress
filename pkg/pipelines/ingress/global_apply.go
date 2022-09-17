package ingress

import (
	"strings"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/infinytum/ingress/internal/config"
	"github.com/infinytum/reactive"
)

// GlobalApply applies the routes to the caddy server
func GlobalApply() reactive.Pipe {
	return GlobalPipe(func(ctx *GlobalContext, errs []error) []error {
		config.Edit(func(c *config.Config) {
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
				for _, route := range ctx.Routes {
					app.Routes = append(app.Routes, *route)
				}
			}
		})
		return errs
	})
}
