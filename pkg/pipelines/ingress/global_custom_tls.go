package ingress

import (
	"github.com/infinytum/ingress/internal/config"
	"github.com/infinytum/reactive"
	"github.com/infinytum/structures"
	"github.com/thoas/go-funk"
)

func GlobalCustomTLS() reactive.Pipe {
	hostMap := structures.NewMap[string, []string]()
	secretMap := structures.NewMap[string, []string]()
	return GlobalPipe(func(ctx *GlobalContext, errs []error) []error {
		config.Edit(func(c *config.Config) {
			httpApp := c.GetHTTPApp()
			knownHosts := hostMap.GetOrDefault(string(ctx.Ingress.UID), []string{})

			// Create a list of all TLS hosts and the assigned secrets
			// which is stored in the hostMap and secretMap for later comparison and removal
			hostsToSkip := make([]string, 0)
			secrets := make([]string, 0)
			for _, tls := range ctx.Ingress.Spec.TLS {
				hostsToSkip = append(hostsToSkip, tls.Hosts...)
				secrets = append(secrets, ctx.Ingress.Namespace+"/"+tls.SecretName)
			}
			hostMap.Set(string(ctx.Ingress.UID), hostsToSkip)

			// If there are any existing hosts, we clear them out to prevent lingering hosts that have been removed
			if len(knownHosts) > 0 {
				httpApp.AutoHTTPS.SkipCerts = funk.Subtract(httpApp.AutoHTTPS.SkipCerts, knownHosts).([]string)
			}

			if len(hostsToSkip) > 0 && ctx.Mode == ContextModeConfigure {
				httpApp.AutoHTTPS.SkipCerts = append(httpApp.AutoHTTPS.SkipCerts, hostsToSkip...)
			}

			// Configure the TLS app to load the certificates for the hosts from the assigned secret
			// This is done by prefixing all files with "secret/" which is identified by kubestore.
			// Kubestore will then lookup the certificate in its appropriate namespace and load it.
			certs := c.GetTLSCertificates()
			knownSecrets := secretMap.GetOrDefault(string(ctx.Ingress.UID), []string{})
			secretMap.Set(string(ctx.Ingress.UID), secrets)
			if len(knownSecrets) > 0 {
				certs.Pairs = funk.Filter(certs.Pairs, func(pair config.TLSCertificatePair) bool {
					return string(ctx.Ingress.UID) != pair.Tags[0]
				}).([]config.TLSCertificatePair)
			}
			if len(secrets) > 0 && ctx.Mode == ContextModeConfigure {
				for _, secret := range secrets {
					// By tagging all pairs with the ingress UID, we know which pairs to remove if the ingress
					// changes or is removed.
					certs.Pairs = append(certs.Pairs, config.TLSCertificatePair{
						Certificate: "secret/" + secret + "/tls.crt",
						Key:         "secret/" + secret + "/tls.key",
						Format:      "pem",
						Tags:        []string{string(ctx.Ingress.UID)},
					})
				}
			}
			c.SetTLSCertificates(certs)
		})
		return errs
	})
}
