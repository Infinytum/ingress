package service

import (
	"github.com/go-mojito/mojito/log"
	"github.com/infinytum/ingress/internal/signals"
	"github.com/infinytum/ingress/pkg/pipelines"
	"github.com/infinytum/injector"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func init() {
	injector.Singleton(newK8sConfigMap)
}

type K8sConfigMap struct {
	config   IngressConfig
	pipeline *pipelines.Configmap
}

func (k *K8sConfigMap) IsManagedByController(cm v1.ConfigMap) bool {
	return cm.Name == k.config.ConfigMap
}

func (k *K8sConfigMap) onAdd(obj interface{}) {
	configMap, ok := obj.(*v1.ConfigMap)
	if ok && k.IsManagedByController(*configMap) {
		log.Field("name", configMap.Name).Field("namespace", configMap.Namespace).Info("Ingress configuration loaded")
		k.pipeline.Configure(configMap)
	}
}

func (k *K8sConfigMap) onUpdate(oldObj, newObj interface{}) {
	configMap, ok := newObj.(*v1.ConfigMap)
	if ok && k.IsManagedByController(*configMap) {
		log.Field("name", configMap.Name).Field("namespace", configMap.Namespace).Info("Ingress configuration reloaded")
		k.pipeline.Configure(configMap)
	}
}

func newK8sConfigMap(clientset *kubernetes.Clientset, config IngressConfig, pipeline *pipelines.Configmap, podinfo PodInfo) *K8sConfigMap {
	k := &K8sConfigMap{
		config:   config,
		pipeline: pipeline,
	}

	informer := informers.NewSharedInformerFactoryWithOptions(
		clientset,
		resourcesSyncInterval,
		informers.WithNamespace(podinfo.Namespace),
	).Core().V1().ConfigMaps().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    k.onAdd,
		UpdateFunc: k.onUpdate,
		DeleteFunc: nil,
	})

	go informer.Run(injector.MustInject[signals.Signal](signals.STOP))
	return k
}
