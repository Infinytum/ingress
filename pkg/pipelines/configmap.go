package pipelines

import (
	"github.com/infinytum/ingress/pkg/pipelines/configmap"
	"github.com/infinytum/injector"
	"github.com/infinytum/reactive"
	v1 "k8s.io/api/core/v1"
)

func init() {
	injector.Singleton(configmapFactory)
}

type Configmap struct {
	pipeline reactive.Subjectable
}

func (i *Configmap) Configure(c *v1.ConfigMap) error {
	ctx, err := configmap.ParseConfigMap(c)
	if err != nil {
		return err
	}
	i.Emit(ctx)
	return nil
}

func (i *Configmap) Emit(ctx *configmap.Context) {
	i.pipeline.Next(ctx, make([]error, 0))
}

func configmapFactory() *Configmap {
	i := &Configmap{
		pipeline: reactive.NewSubject(),
	}

	i.pipeline.Pipe(
		configmap.TLS(),
	)

	return i
}
