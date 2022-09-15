package main

import (
	zerolog "github.com/go-mojito/logger-zerolog"
	"github.com/go-mojito/mojito/log"
	"github.com/infinytum/injector"

	caddyv2 "github.com/caddyserver/caddy/v2"
	"github.com/infinytum/ingress/internal/service"
	"github.com/infinytum/ingress/pkg/caddy"
	"github.com/infinytum/ingress/pkg/modules/kubestore"

	zlog "github.com/rs/zerolog"
)

func init() {
	zerolog.Pretty()
	zerolog.AsDefault()
	zlog.SetGlobalLevel(zlog.InfoLevel)
}

func main() {
	injector.MustCall(func(cfg service.IngressConfig) {
		log.Infof("Loaded config: %+v", cfg)
	})

	_ = injector.MustInject[*service.K8sIngress]()

	caddyv2.RegisterModule(kubestore.KubeStore{})
	caddy.Reload()
	select {}
}
