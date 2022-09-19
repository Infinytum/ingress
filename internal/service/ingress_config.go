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
	className        = flag.String("class-name", "", "class name of the ingress controller")
	configMap        = flag.String("config-map", "infinytum-ingress-cfg", "name of the config map to use for configuration")
	enableHSTS       = flag.Bool("enable-hsts", true, "enable hsts header for all ingress routes")
	http3            = flag.Bool("http3", true, "enable experimental http3 support")
	kubeConfig       = flag.String("kube-config", "", "path to kube config file")
	namespace        = flag.String("namespace", v1.NamespaceAll, "namespace to watch for ingress resources")
	nginxAnnotations = flag.Bool("nginx-annotations", false, "enables the ingress to use some nginx-specific annotations")
)

type IngressConfig struct {
	// Ingress ClassName to use
	ClassName string
	// ConfigMap to watch for configuration changes
	ConfigMap string
	// Enable HSTS header for all ingress routes
	EnableHSTS bool
	// Whether to enable experimental HTTP3 support
	HTTP3 bool
	// Path to kube config file (useful for development outside of a kubernetes cluster)
	KubeConfig string
	// Namespace to watch for ingress resources
	Namespace string
	// Whether to watch for nginx-specific annotations
	NginxAnnotations bool
}

func (c IngressConfig) String() string {
	s, _ := json.MarshalIndent(c, "", "  ")
	return string(s)
}

func newIngressConfig() IngressConfig {
	flag.Parse()
	return IngressConfig{
		ClassName:        *className,
		ConfigMap:        *configMap,
		EnableHSTS:       *enableHSTS,
		HTTP3:            *http3,
		KubeConfig:       *kubeConfig,
		Namespace:        *namespace,
		NginxAnnotations: *nginxAnnotations,
	}
}
