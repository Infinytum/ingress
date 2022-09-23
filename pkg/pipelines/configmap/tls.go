package configmap

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
	"github.com/infinytum/ingress/internal/config"
	"github.com/infinytum/reactive"
	"github.com/mholt/acmez/acme"
)

func TLS() reactive.Pipe {
	return Pipe(func(ctx *Context, errs []error) []error {
		config.Edit(func(config *config.Config) {
			tlsApp := config.GetTLSApp()
			acmeIssuer := caddytls.ACMEIssuer{
				Challenges: &caddytls.ChallengesConfig{
					HTTP: &caddytls.HTTPChallengeConfig{
						AlternatePort: 8080,
					},
					TLSALPN: &caddytls.TLSALPNChallengeConfig{
						AlternatePort: 8443,
					},
				},
			}

			if ctx.AcmeCA != "" {
				acmeIssuer.CA = ctx.AcmeCA
			}

			if ctx.AcmeEmail != "" {
				acmeIssuer.Email = ctx.AcmeEmail
			}

			if ctx.AcmeEABKeyId != "" && ctx.AcmeEABMacKey != "" {
				acmeIssuer.ExternalAccount = &acme.EAB{
					KeyID:  ctx.AcmeEABKeyId,
					MACKey: ctx.AcmeEABMacKey,
				}
			}

			var onDemandConfig *caddytls.OnDemandConfig
			if ctx.OnDemandTLS {
				ask := ctx.OnDemandAsk
				if ctx.OnDemandInternalAsk {
					ask = "http://localhost:8123/ask"
				}
				onDemandConfig = &caddytls.OnDemandConfig{
					RateLimit: &caddytls.RateLimit{
						Interval: ctx.OnDemandRateLimitInterval,
						Burst:    ctx.OnDemandRateLimitBurst,
					},
					Ask: ask,
				}
			}

			tlsApp.Automation = &caddytls.AutomationConfig{
				OnDemand:          onDemandConfig,
				OCSPCheckInterval: ctx.OCSPCheckInterval,
				Policies: []*caddytls.AutomationPolicy{
					{
						IssuersRaw: []json.RawMessage{
							caddyconfig.JSONModuleObject(acmeIssuer, "module", "acme", nil),
						},
						OnDemand: ctx.OnDemandTLS,
					},
				},
			}
		})

		return errs
	})
}
