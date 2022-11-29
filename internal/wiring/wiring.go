package wiring

import (
	"github.com/infinytum/ingress/internal/signals"
	"github.com/infinytum/ingress/pkg/pipelines"
	"github.com/infinytum/injector"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

func init() {
	/** ConfigMap Service **/
	injector.DeferredSingleton(func(pipeline *pipelines.Configmap) signals.ConfigMapUpdated {
		return func(configMap *v1.ConfigMap) error {
			return pipeline.Configure(configMap)
		}
	})

	/** Ingress Service **/
	injector.DeferredSingleton(func(pipeline *pipelines.Ingress) signals.IngressUpdated {
		return func(ingress *networkingv1.Ingress) {
			pipeline.Configure(ingress)
		}
	})
	injector.DeferredSingleton(func(pipeline *pipelines.Ingress) signals.IngressDeleted {
		return func(ingress *networkingv1.Ingress) {
			pipeline.Delete(ingress)
		}
	})
}
