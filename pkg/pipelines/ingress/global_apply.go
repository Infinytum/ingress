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
			existingRoutes := extractRoutes(string(ctx.Ingress.UID), app)
			if len(existingRoutes) > 0 {
				// Protect against race conditions
				if !ctx.IsNewer(existingRoutes[0].Group) {
					ctx.Routes = existingRoutes // Use existing routes since they are newer
				}
			}

			// Only configure new routes if the ingress is being configured
			if ctx.Mode == ContextModeConfigure {
				for _, route := range ctx.Routes {
					app.Routes = append(app.Routes, *route)
				}
			}
		})
		return errs
	})
}

// extractRoutes extracts all routes that belong to the given ingress uid
func extractRoutes(uid string, c *caddyhttp.Server) []*caddyhttp.Route {
	routes := make([]*caddyhttp.Route, 0)
	newRoutes := make([]caddyhttp.Route, 0)
	for _, route := range c.Routes {
		if strings.HasPrefix(route.Group, uid) {
			routes = append(routes, &route)
		} else {
			newRoutes = append(newRoutes, route)
		}
	}
	c.Routes = newRoutes
	return routes
}
