package handlers

import (
	"github.com/go-mojito/mojito"
	"github.com/go-mojito/mojito/log"
	"github.com/infinytum/ingress/internal/service"
)

func Ask(ctx mojito.Context, state *service.State) error {
	domainToCheck := ctx.Request().GetRequest().FormValue("domain")
	line := log.Field("domain", domainToCheck)
	if domainToCheck == "" {
		ctx.Response().GetWriter().WriteHeader(403)
		line.Warn("[OnDemand TLS] No domain provided, rejecting certificate request by default.")
		return ctx.String("No domain provided")
	}

	if !state.IsHostConfigured(domainToCheck) {
		ctx.Response().GetWriter().WriteHeader(403)
		line.Debug("[OnDemand TLS] Domain is not configured on this caddy instance, rejecting certificate request.")
		return ctx.String("Domain is not configured on this caddy instance, rejecting certificate request.")
	}

	line.Info("[OnDemand TLS] Caddy knows this domain, issue a certificate for it.")
	return ctx.String("Caddy knows this domain, issue a certificate for it.")
}
