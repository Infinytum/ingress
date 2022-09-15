package kubestore

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/certmagic"
	"github.com/infinytum/ingress/internal/service"
	"github.com/infinytum/injector"
)

var labels = map[string]string{
	"infinytum.co/managed": "true",
}

const dataKey = "data"

type KubeStore struct {
	LeaseId string
}

func (KubeStore) Namespace() string {
	podInfo := injector.MustInject[service.PodInfo]()
	return podInfo.Namespace
}

func (KubeStore) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "caddy.storage.kubestore",
		New: func() caddy.Module { return new(KubeStore) },
	}
}

func (k *KubeStore) CertMagicStorage() (certmagic.Storage, error) {
	return k, nil
}
