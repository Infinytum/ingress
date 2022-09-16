package signals

import (
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

/** ConfigMap Service **/
type ConfigMapUpdated func(configMap *v1.ConfigMap) error

/** Ingress Service **/
type IngressUpdated func(ingress *networkingv1.Ingress)
type IngressDeleted func(ingress *networkingv1.Ingress)
