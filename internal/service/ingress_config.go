package service

import (
	"encoding/json"
	"flag"

	"github.com/infinytum/injector"
	v1 "k8s.io/api/core/v1"
)

func init() {
	injector.Singleton(newIngressConfig)
}

var (
	className  = flag.String("class-name", "", "class name of the ingress controller")
	kubeConfig = flag.String("kube-config", "", "path to kube config file")
	namespace  = flag.String("namespace", v1.NamespaceAll, "namespace to watch for ingress resources")
	configMap  = flag.String("config-map", "infinytum-ingress-cfg", "name of the config map to use for configuration")
)

type IngressConfig struct {
	// Ingress ClassName to use
	ClassName string
	// Path to kube config file (useful for development outside of a kubernetes cluster)
	KubeConfig string
	// Namespace to watch for ingress resources
	Namespace string
	// ConfigMap to watch for configuration changes
	ConfigMap string
}

func (c IngressConfig) String() string {
	s, _ := json.MarshalIndent(c, "", "  ")
	return string(s)
}

func newIngressConfig() IngressConfig {
	flag.Parse()
	return IngressConfig{
		ClassName:  *className,
		KubeConfig: *kubeConfig,
		Namespace:  *namespace,
		ConfigMap:  *configMap,
	}
}
