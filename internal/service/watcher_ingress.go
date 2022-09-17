package service

import (
	"time"

	"github.com/go-mojito/mojito/log"
	"github.com/infinytum/ingress/internal/signals"
	"github.com/infinytum/injector"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	resourcesSyncInterval = time.Minute * 30
)

func init() {
	injector.Singleton(newIngressWatcher)
}

type IngressWatcher struct {
	config          IngressConfig          `injector:"type"`
	onIngressDelete signals.IngressDeleted `injector:"type"`
	onIngressUpdate signals.IngressUpdated `injector:"type"`
}

// IsManagedByController returns true if the ingress resource has the correct ingress class
func (k *IngressWatcher) IsManagedByController(ingress networkingv1.Ingress) bool {
	ingressClass := ingress.Annotations["kubernetes.io/ingress.class"]
	if ingressClass == "" && ingress.Spec.IngressClassName != nil {
		ingressClass = *ingress.Spec.IngressClassName
	}
	return ingressClass == k.config.ClassName
}

func (k *IngressWatcher) onAdd(obj interface{}) {
	ingress, ok := obj.(*networkingv1.Ingress)
	if ok && k.IsManagedByController(*ingress) {
		log.Field("name", ingress.Name).Field("namespace", ingress.Namespace).Info("Discovered ingress")
		if k.onIngressUpdate != nil {
			k.onIngressUpdate(ingress)
		}
	}
}

func (k *IngressWatcher) onUpdate(oldObj, newObj interface{}) {
	ingress, ok := newObj.(*networkingv1.Ingress)
	if ok && k.IsManagedByController(*ingress) {
		log.Field("name", ingress.Name).Field("namespace", ingress.Namespace).Info("Updated ingress")
		if k.onIngressUpdate != nil {
			k.onIngressUpdate(ingress)
		}
	}
}

func (k *IngressWatcher) onDelete(obj interface{}) {
	ingress, ok := obj.(*networkingv1.Ingress)
	if ok && k.IsManagedByController(*ingress) {
		log.Field("name", ingress.Name).Field("namespace", ingress.Namespace).Info("Deleted ingress")
		if k.onIngressDelete != nil {
			k.onIngressDelete(ingress)
		}
	}
}

func newIngressWatcher() *IngressWatcher {
	k := &IngressWatcher{}
	injector.MustFill(k)
	injector.MustCall(func(clientset *kubernetes.Clientset, config IngressConfig) {
		informer := informers.NewSharedInformerFactoryWithOptions(
			clientset,
			resourcesSyncInterval,
			informers.WithNamespace(config.Namespace),
		).Networking().V1().Ingresses().Informer()

		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    k.onAdd,
			UpdateFunc: k.onUpdate,
			DeleteFunc: k.onDelete,
		})

		go informer.Run(injector.MustInject[signals.Signal](signals.STOP))
	})
	return k
}
