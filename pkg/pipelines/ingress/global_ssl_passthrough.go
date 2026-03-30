package ingress

import (
	"fmt"
	"log/slog"

	"github.com/infinytum/ingress/internal/annotations"
	"github.com/infinytum/ingress/internal/config"
	"github.com/infinytum/reactive"
	"github.com/infinytum/structures"
)

// GlobalSSLPassthrough manages layer4 SSL passthrough routes for ingresses
// annotated with ssl-passthrough: "true". It builds the caddy-l4 config
// that proxies raw TLS connections to backends without terminating TLS.
func GlobalSSLPassthrough() reactive.Pipe {
	// Track passthrough entries per ingress UID for cleanup on update/delete
	passthroughMap := structures.NewMap[string, []config.PassthroughEntry]()

	return GlobalPipe(func(ctx *GlobalContext, errs []error) []error {
		isPassthrough := annotations.GetAnnotationBool(
			ctx.Ingress.ObjectMeta, annotations.AnnotationSSLPassthrough, false,
		)

		uid := string(ctx.Ingress.UID)
		hadPassthrough := passthroughMap.Contains(uid)

		// Collect passthrough entries from the ingress spec
		var entries []config.PassthroughEntry
		if isPassthrough && ctx.Mode == ContextModeConfigure {
			for _, rule := range ctx.Ingress.Spec.Rules {
				if rule.Host == "" || rule.HTTP == nil {
					continue
				}
				// Use the first path's backend as the passthrough target
				for _, path := range rule.HTTP.Paths {
					port := path.Backend.Service.Port.Number
					if port == 0 {
						port = 443 // default for passthrough
					}
					entries = append(entries, config.PassthroughEntry{
						Host: rule.Host,
						Dial: fmt.Sprintf("%s.%s.svc.cluster.local:%d",
							path.Backend.Service.Name,
							ctx.Ingress.Namespace,
							port,
						),
					})
					break // one backend per host for passthrough
				}
			}
		}

		// Update or remove this ingress's passthrough entries
		if len(entries) > 0 {
			passthroughMap.Set(uid, entries)
		} else {
			passthroughMap.Delete(uid)
		}

		// Rebuild layer4 config from all tracked passthrough entries
		var allEntries []config.PassthroughEntry
		for _, e := range passthroughMap.ToMap() {
			allEntries = append(allEntries, e...)
		}

		err := config.Edit(func(c *config.Config) {
			httpApp := c.GetHTTPApp()

			if len(allEntries) == 0 {
				// No passthrough hosts — remove layer4, restore direct listen
				c.SetLayer4Config(nil)
				if hadPassthrough {
					httpApp.Listen = []string{config.ExternalHTTPSAddr}
					slog.Info("SSL passthrough disabled, restored direct HTTPS listener")
				}
				return
			}

			// Build layer4 routes: one per passthrough host + catch-all
			routes := make([]config.Layer4Route, 0, len(allEntries)+1)
			for _, entry := range allEntries {
				routes = append(routes, config.BuildPassthroughRoute(entry.Host, entry.Dial))
				slog.With("host", entry.Host, "backend", entry.Dial).Debug("SSL passthrough route")
			}
			routes = append(routes, config.BuildCatchAllRoute())

			l4 := &config.Layer4Config{
				Servers: map[string]*config.Layer4Server{
					"ssl_mux": {
						Listen:           []string{config.ExternalHTTPSAddr},
						Routes:           routes,
						MaxMatchingBytes: 4096,
					},
				},
			}
			c.SetLayer4Config(l4)

			// Move HTTPS server to internal address so layer4 can front it
			httpApp.Listen = []string{config.InternalHTTPSAddr}
		})

		if err != nil {
			slog.Error("Error configuring SSL passthrough", "error", err)
			errs = append(errs, err)
		}

		return errs
	})
}
