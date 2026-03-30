package main

import (
	_ "github.com/caddyserver/caddy/v2/modules/caddyhttp/caddyauth"
	_ "github.com/caddyserver/caddy/v2/modules/logging"
	_ "github.com/caddyserver/caddy/v2/modules/metrics"
	_ "github.com/infinytum/ingress/internal/wiring"
	_ "github.com/mholt/caddy-l4/modules/l4proxy"
	_ "github.com/mholt/caddy-l4/modules/l4tls"

	"k8s.io/client-go/kubernetes"

	"flag"
	"log/slog"
	"os"

	"github.com/go-mojito/mojito"
	"github.com/infinytum/ingress/internal/signals"
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
	flag.Parse()
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

	// Start HTTP server in background
	go func() {
		if err := mojito.ListenAndServe(":8123"); err != nil {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	// Wait for interrupt signal
	interrupt := injector.MustInject[signals.Signal](signals.INTERRUPT)
	<-interrupt
	slog.Info("Shutting down gracefully...")

	// Signal informers to stop
	stop := injector.MustInject[signals.Signal](signals.STOP)
	stop <- struct{}{}

	// Shutdown Mojito HTTP server
	if err := mojito.Shutdown(); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
	}

	// Shutdown Caddy
	if err := caddyv2.Stop(); err != nil {
		slog.Error("Caddy shutdown error", "error", err)
	}

	slog.Info("Shutdown complete")
}
