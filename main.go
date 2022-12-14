package main

import (
	_ "github.com/caddyserver/caddy/v2/modules/caddyhttp/caddyauth"
	_ "github.com/caddyserver/caddy/v2/modules/logging"
	_ "github.com/caddyserver/caddy/v2/modules/metrics"
	_ "github.com/infinytum/ingress/internal/wiring"

	"k8s.io/client-go/kubernetes"

	zerolog "github.com/go-mojito/logger-zerolog"
	"github.com/go-mojito/mojito"
	"github.com/go-mojito/mojito/log"
	"github.com/infinytum/injector"

	caddyv2 "github.com/caddyserver/caddy/v2"
	"github.com/infinytum/ingress/internal/service"
	"github.com/infinytum/ingress/pkg/handlers"
	"github.com/infinytum/ingress/pkg/modules/kubestore"
	"github.com/infinytum/ingress/pkg/modules/mojitolog"
	"github.com/infinytum/ingress/pkg/modules/proxy"

	zlog "github.com/rs/zerolog"
)

func init() {
	zerolog.Pretty()
	zerolog.AsDefault()
	zlog.SetGlobalLevel(zlog.InfoLevel)
}

func main() {
	log.Info("Discovering Kubernetes API server...")
	_ = injector.MustInject[*kubernetes.Clientset]()
	injector.MustCall(func(cfg service.IngressConfig) {
		log.Infof("Loaded config: %+v", cfg)
	})

	log.Info("Creating Pod Watcher")
	_ = injector.MustInject[*service.PodWatcher]()
	log.Info("Creating ConfigMap Watcher")
	_ = injector.MustInject[*service.ConfigMapWatcher]()
	log.Info("Creating Ingress Watcher")
	_ = injector.MustInject[*service.IngressWatcher]()

	log.Info("Registering Kubernetes TLS Storage module")
	caddyv2.RegisterModule(mojitolog.MojitoLog{})
	caddyv2.RegisterModule(kubestore.KubeStore{})
	caddyv2.RegisterModule(proxy.Proxy{})

	mojito.GET("/ask", handlers.Ask)
	mojito.ListenAndServe(":8123")
}
