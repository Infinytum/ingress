package handlers

import (
	"net"

	"github.com/go-mojito/mojito"
	"github.com/go-mojito/mojito/log"
)

func Ask(ctx mojito.Context, inftm *services.InfinytumService) error {
	domainToCheck := ctx.Request().GetRequest().FormValue("domain")
	if domainToCheck == "" {
		return ctx.String("No domain provided")
	}

	ips, err := net.LookupIP(domainToCheck)
	if err != nil {
		log.Errorf("Failed to resolve domain: %s", err)
	}

	for _, ip := range ips {
		if !inftm.IsInfinytum(ip.String()) {
			ctx.Response().WriteHeader(403)
			return ctx.String("Domain is not pointing to Infinytum, make sure you have set up your DNS records correctly.")
		}
	}
	return ctx.String("Domain is pointing to Infinytum, issue a certificate for it.")
}
