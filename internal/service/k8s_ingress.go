package service

import (
	"time"

	"github.com/go-mojito/mojito/log"
	"github.com/infinytum/ingress/internal/signals"
	"github.com/infinytum/ingress/pkg/pipelines"
	"github.com/infinytum/injector"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	resourcesSyncInterval = time.Hour
)

func init() {
	injector.Singleton(newK8sIngress)
}

type K8sIngress struct {
	config   IngressConfig
	pipeline *pipelines.Ingress
}

// IsManagedByController returns true if the ingress resource has the correct ingress class
func (k *K8sIngress) IsManagedByController(ingress networkingv1.Ingress) bool {
	ingressClass := ingress.Annotations["kubernetes.io/ingress.class"]
	if ingressClass == "" && ingress.Spec.IngressClassName != nil {
		ingressClass = *ingress.Spec.IngressClassName
	}
	return ingressClass == k.config.ClassName
}

func (k *K8sIngress) onAdd(obj interface{}) {
	ingress, ok := obj.(*networkingv1.Ingress)
	if ok && k.IsManagedByController(*ingress) {
		log.Field("name", ingress.Name).Field("namespace", ingress.Namespace).Info("Discovered ingress")
		k.pipeline.Configure(ingress)
	}
}

func (k *K8sIngress) onUpdate(oldObj, newObj interface{}) {
	ingress, ok := newObj.(*networkingv1.Ingress)
	if ok && k.IsManagedByController(*ingress) {
		log.Field("name", ingress.Name).Field("namespace", ingress.Namespace).Info("Updated ingress")
		k.pipeline.Configure(ingress)
	}
}

func (k *K8sIngress) onDelete(obj interface{}) {
	ingress, ok := obj.(*networkingv1.Ingress)
	if ok && k.IsManagedByController(*ingress) {
		log.Field("name", ingress.Name).Field("namespace", ingress.Namespace).Info("Deleted ingress")
		k.pipeline.Delete(ingress)
	}
}

func newK8sIngress(clientset *kubernetes.Clientset, config IngressConfig, ingPipeline *pipelines.Ingress) *K8sIngress {
	k := &K8sIngress{
		config:   config,
		pipeline: ingPipeline,
	}

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
	return k
}
