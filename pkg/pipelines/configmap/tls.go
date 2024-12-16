package configmap

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
	"github.com/go-mojito/mojito/log"
	"github.com/infinytum/ingress/internal/config"
	"github.com/infinytum/reactive"
	"github.com/mholt/acmez/v2/acme"
)

func TLS() reactive.Pipe {
	return Pipe(func(ctx *Context, errs []error) []error {
		err := config.Edit(func(config *config.Config) {
			tlsApp := config.GetTLSApp()

			var onDemandConfig *caddytls.OnDemandConfig
			if ctx.OnDemandTLS {
				ask := ctx.OnDemandAsk
				if ctx.OnDemandInternalAsk {
					ask = "http://localhost:8123/ask"
				}
				onDemandConfig = &caddytls.OnDemandConfig{
					Ask: ask,
				}
			}

			issuersRaw := make([]json.RawMessage, 0)
			if issuer := generateBuyPassIssuer(ctx); issuer != nil {
				issuersRaw = append(issuersRaw, caddyconfig.JSONModuleObject(*issuer, "module", "acme", nil))
			}
			if issuer := generateZeroSSLIssuer(ctx); issuer != nil {
				issuersRaw = append(issuersRaw, caddyconfig.JSONModuleObject(*issuer, "module", "acme", nil))
			}
			if issuer := generateLEIssuer(ctx); issuer != nil {
				issuersRaw = append(issuersRaw, caddyconfig.JSONModuleObject(*issuer, "module", "acme", nil))
			}

			tlsApp.Automation = &caddytls.AutomationConfig{
				OnDemand:          onDemandConfig,
				OCSPCheckInterval: ctx.OCSPCheckInterval,
				Policies: []*caddytls.AutomationPolicy{
					{
						IssuersRaw: issuersRaw,
						OnDemand:   ctx.OnDemandTLS,
					},
				},
			}
		})

		if err != nil {
			log.Errorf("Error while configuring TLS: %v", err)
			errs = append(errs, err)
		}

		return errs
	})
}

func generateBuyPassIssuer(ctx *Context) *caddytls.ACMEIssuer {

	if ctx.AcmeEmail == "" {
		return nil
	}

	return &caddytls.ACMEIssuer{
		Challenges: &caddytls.ChallengesConfig{
			HTTP: &caddytls.HTTPChallengeConfig{
				AlternatePort: 8080,
			},
			TLSALPN: &caddytls.TLSALPNChallengeConfig{
				AlternatePort: 8443,
			},
		},
		CA:    "https://api.buypass.com/acme/directory",
		Email: ctx.AcmeEmail,
	}
}

func generateLEIssuer(ctx *Context) *caddytls.ACMEIssuer {

	if ctx.AcmeEmail == "" {
		return nil
	}

	return &caddytls.ACMEIssuer{
		Challenges: &caddytls.ChallengesConfig{
			HTTP: &caddytls.HTTPChallengeConfig{
				AlternatePort: 8080,
			},
			TLSALPN: &caddytls.TLSALPNChallengeConfig{
				AlternatePort: 8443,
			},
		},
		CA:    "https://acme-v02.api.letsencrypt.org/directory",
		Email: ctx.AcmeEmail,
	}
}

func generateZeroSSLIssuer(ctx *Context) *caddytls.ACMEIssuer {

	if ctx.AcmeEmail == "" {
		return nil
	}

	if ctx.AcmeEABKeyId == "" || ctx.AcmeEABMacKey == "" {
		return nil
	}

	return &caddytls.ACMEIssuer{
		Challenges: &caddytls.ChallengesConfig{
			HTTP: &caddytls.HTTPChallengeConfig{
				AlternatePort: 8080,
			},
			TLSALPN: &caddytls.TLSALPNChallengeConfig{
				AlternatePort: 8443,
			},
		},
		CA:    "https://acme.zerossl.com/v2/DV90",
		Email: ctx.AcmeEmail,
		ExternalAccount: &acme.EAB{
			KeyID:  ctx.AcmeEABKeyId,
			MACKey: ctx.AcmeEABMacKey,
		},
	}
}
