package service

import (
	"context"
	"os"
	"sync"
	"time"

	"log/slog"

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
	injector.DeferredSingleton(newPodWatcher)
}

type PodWatcher struct {
	mu         sync.RWMutex
	namespace  string
	pod        *apiv1.Pod
	nodeIp     string
	serviceIps []string
}

func (p *PodWatcher) Namespace() string {
	return p.namespace
}

func (p *PodWatcher) IPs() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.nodeIp == "" {
		ips := make([]string, len(p.serviceIps))
		copy(ips, p.serviceIps)
		return ips
	}
	ips := make([]string, len(p.serviceIps)+1)
	copy(ips, p.serviceIps)
	ips[len(p.serviceIps)] = p.nodeIp
	return ips
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
		slog.With("name", pod.Name, "namespace", pod.Namespace).Error("Error refreshing pod IPs", "error", err)
	}
}

func (p *PodWatcher) refreshIps(clientset *kubernetes.Clientset) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get services that may select this pod
	svcs, err := clientset.CoreV1().Services(p.pod.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		slog.With("name", p.pod.Name, "namespace", p.pod.Namespace).Error("Error getting services", "error", err)
		return
	}

	// Get IPs of services that select this pod
	serviceIps := make([]string, 0)
	for _, svc := range svcs.Items {
		if svc.Spec.Selector == nil {
			continue
		}
		if utils.IsSubset(svc.Spec.Selector, p.pod.Labels) {
			if ips := utils.GetAddressesFromService(&svc); len(ips) > 0 {
				serviceIps = append(serviceIps, ips...)
			}
		}
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.serviceIps = serviceIps

	// Compatibility with hostPort deployments
	if len(p.serviceIps) == 0 {
		p.nodeIp = p.pod.Status.HostIP
	} else {
		p.nodeIp = ""
	}
}

func newPodWatcher(clientset *kubernetes.Clientset) *PodWatcher {
	podNamespace := os.Getenv("POD_NAMESPACE")
	podName := os.Getenv("POD_NAME")
	if podNamespace == "" || podName == "" {
		slog.Error("POD_NAMESPACE and POD_NAME must be set")
		os.Exit(1)
	}

	k := &PodWatcher{
		namespace:  podNamespace,
		serviceIps: make([]string, 0),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pod, err := clientset.CoreV1().Pods(podNamespace).Get(ctx, podName, metav1.GetOptions{})
	if errors.IsNotFound(err) || pod == nil {
		slog.Error("Could not find pod in kubernetes, make sure POD_NAME and POD_NAMESPACE are set")
		return nil
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
	return k
}
