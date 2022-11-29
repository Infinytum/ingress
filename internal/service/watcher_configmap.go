package service

import (
	"github.com/go-mojito/mojito/log"
	"github.com/infinytum/ingress/internal/signals"
	"github.com/infinytum/injector"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func init() {
	injector.DeferredSingleton(newConfigMapWatcher)
}

type ConfigMapWatcher struct {
	config            IngressConfig            `injector:"type"`
	onConfigMapUpdate signals.ConfigMapUpdated `injector:"type"`
}

func (k *ConfigMapWatcher) IsManagedByController(cm v1.ConfigMap) bool {
	return cm.Name == k.config.ConfigMap
}

func (k *ConfigMapWatcher) onAdd(obj interface{}) {
	configMap, ok := obj.(*v1.ConfigMap)
	if ok && k.IsManagedByController(*configMap) {
		log.Field("name", configMap.Name).Field("namespace", configMap.Namespace).Info("Ingress configuration loaded")
		if k.onConfigMapUpdate != nil {
			k.onConfigMapUpdate(configMap)
		}
	}
}

func (k *ConfigMapWatcher) onUpdate(oldObj, newObj interface{}) {
	configMap, ok := newObj.(*v1.ConfigMap)
	if ok && k.IsManagedByController(*configMap) {
		log.Field("name", configMap.Name).Field("namespace", configMap.Namespace).Info("Ingress configuration reloaded")
		if k.onConfigMapUpdate != nil {
			k.onConfigMapUpdate(configMap)
		}
	}
}

func newConfigMapWatcher(clientset *kubernetes.Clientset, podWatcher *PodWatcher) *ConfigMapWatcher {
	k := &ConfigMapWatcher{}
	injector.MustFill(k)
	informer := informers.NewSharedInformerFactoryWithOptions(
		clientset,
		resourcesSyncInterval,
		informers.WithNamespace(podWatcher.Namespace()),
	).Core().V1().ConfigMaps().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    k.onAdd,
		UpdateFunc: k.onUpdate,
		DeleteFunc: nil,
	})

	go informer.Run(injector.MustInject[signals.Signal](signals.STOP))
	return k
}
