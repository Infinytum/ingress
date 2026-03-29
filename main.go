package main

import (
	_ "github.com/caddyserver/caddy/v2/modules/caddyhttp/caddyauth"
	_ "github.com/caddyserver/caddy/v2/modules/logging"
	_ "github.com/caddyserver/caddy/v2/modules/metrics"
	_ "github.com/infinytum/ingress/internal/wiring"

	"k8s.io/client-go/kubernetes"

	"log/slog"
	"os"

	"github.com/go-mojito/mojito"
	"github.com/infinytum/injector"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	slogzerolog "github.com/samber/slog-zerolog/v2"

	caddyv2 "github.com/caddyserver/caddy/v2"
	"github.com/infinytum/ingress/internal/service"
	"github.com/infinytum/ingress/pkg/handlers"
	"github.com/infinytum/ingress/pkg/modules/kubestore"
	"github.com/infinytum/ingress/pkg/modules/mojitolog"
	"github.com/infinytum/ingress/pkg/modules/proxy"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	slog.SetDefault(slog.New(slogzerolog.Option{Level: slog.LevelInfo, Logger: &log.Logger}.NewZerologHandler()))
}

func main() {
	slog.Info("Discovering Kubernetes API server...")
	_ = injector.MustInject[*kubernetes.Clientset]()
	injector.MustCall(func(cfg service.IngressConfig) {
		slog.Info("Loaded config", "config", cfg)
	})

	slog.Info("Creating Pod Watcher")
	_ = injector.MustInject[*service.PodWatcher]()
	slog.Info("Creating ConfigMap Watcher")
	_ = injector.MustInject[*service.ConfigMapWatcher]()
	slog.Info("Creating Ingress Watcher")
	_ = injector.MustInject[*service.IngressWatcher]()

	slog.Info("Registering Kubernetes TLS Storage module")
	caddyv2.RegisterModule(mojitolog.MojitoLog{})
	caddyv2.RegisterModule(kubestore.KubeStore{})
	caddyv2.RegisterModule(proxy.Proxy{})

	mojito.GET("/ask", handlers.Ask)
	mojito.ListenAndServe(":8123")
}
