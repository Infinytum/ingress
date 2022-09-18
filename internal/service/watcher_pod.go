package service

import (
	"context"
	"os"

	"github.com/go-mojito/mojito/log"
	"github.com/infinytum/ingress/internal/signals"
	"github.com/infinytum/ingress/pkg/utils"
	"github.com/infinytum/injector"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func init() {
	injector.Singleton(newPodWatcher)
}

type PodWatcher struct {
	namespace  string
	pod        *apiv1.Pod
	nodeIp     string
	serviceIps []string
}

func (p *PodWatcher) Namespace() string {
	return p.namespace
}

func (p *PodWatcher) IPs() []string {
	if p.nodeIp == "" {
		return p.serviceIps
	}
	return append(p.serviceIps, p.nodeIp)
}

func (p *PodWatcher) onAdd(obj interface{}) {
	p.onUpdate(nil, obj)
}

func (p *PodWatcher) onUpdate(_, obj interface{}) {
	pod := obj.(*apiv1.Pod)
	if string(pod.UID) != string(p.pod.UID) {
		return
	}
	p.pod = pod
	if err := injector.Call(p.refreshIps); err != nil {
		log.Field("name", pod.Name).Field("namespace", pod.Namespace).Errorf("Error refreshing pod IPs: %s", err)
	}
}

func (p *PodWatcher) refreshIps(clientset *kubernetes.Clientset) {
	// Get services that may select this pod
	svcs, err := clientset.CoreV1().Services(p.pod.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Field("name", p.pod.Name).Field("namespace", p.pod.Namespace).Errorf("Error getting services: %s", err)
		return
	}

	// Get IPs of services that select this pod
	serviceIps := make([]string, 0)
	for _, svc := range svcs.Items {
		if svc.Spec.Selector == nil {
			continue
		}
		if utils.IsSubset(p.pod.Labels, svc.Spec.Selector) {
			if ips := utils.GetAddressesFromService(&svc); len(ips) > 0 {
				serviceIps = append(serviceIps, ips...)
			}
		}
	}
	p.serviceIps = serviceIps

	// Compatibility with hostPort deployments
	if len(p.serviceIps) == 0 {
		p.nodeIp = p.pod.Status.HostIP
	} else {
		p.nodeIp = ""
	}
}

func newPodWatcher() *PodWatcher {
	podNamespace := os.Getenv("POD_NAMESPACE")
	podName := os.Getenv("POD_NAME")
	if podNamespace == "" || podName == "" {
		log.Fatal("POD_NAMESPACE and POD_NAME must be set")
	}

	k := &PodWatcher{
		namespace:  podNamespace,
		serviceIps: make([]string, 0),
	}
	injector.MustCall(func(clientset *kubernetes.Clientset) {
		pod, err := clientset.CoreV1().Pods(podNamespace).Get(context.TODO(), podName, metav1.GetOptions{})
		if errors.IsNotFound(err) || pod == nil {
			log.Error("Could not find pod in kubernetes, make sure POD_NAME and POD_NAMESPACE are set")
			return
		}
		k.pod = pod
		k.refreshIps(clientset)

		informer := informers.NewSharedInformerFactoryWithOptions(
			clientset,
			resourcesSyncInterval,
			informers.WithNamespace(podNamespace),
		).Core().V1().Pods().Informer()

		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    k.onAdd,
			UpdateFunc: k.onUpdate,
			DeleteFunc: nil,
		})

		go informer.Run(injector.MustInject[signals.Signal](signals.STOP))
	})
	return k
}
