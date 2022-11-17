package ingress

import (
	"context"
	"net"
	"strings"

	"github.com/infinytum/ingress/internal/service"
	"github.com/infinytum/injector"
	"github.com/infinytum/reactive"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GlobalStatus adds the LB status to the ingress objects
func GlobalStatus() reactive.Pipe {
	return GlobalPipe(func(ctx *GlobalContext, errs []error) []error {
		err := injector.Call(func(clientset *kubernetes.Clientset, podWatcher *service.PodWatcher) {
			if len(podWatcher.IPs()) == 0 {
				return
			}

			modified := false
			for _, ip := range podWatcher.IPs() {
				for _, lb := range ctx.Ingress.Status.LoadBalancer.Ingress {
					if lb.IP == ip || strings.EqualFold(lb.Hostname, ip) {
						goto skip
					}
				}

				if net.ParseIP(ip) != nil {
					ctx.Ingress.Status.LoadBalancer.Ingress = append(ctx.Ingress.Status.LoadBalancer.Ingress, v1.LoadBalancerIngress{IP: ip})
				} else {
					ctx.Ingress.Status.LoadBalancer.Ingress = append(ctx.Ingress.Status.LoadBalancer.Ingress, v1.LoadBalancerIngress{Hostname: ip})
				}
				modified = true

			skip:
			}

			if modified {
				ing, err := clientset.NetworkingV1().Ingresses(ctx.Ingress.Namespace).UpdateStatus(context.TODO(), ctx.Ingress, metav1.UpdateOptions{})
				if err != nil {
					errs = append(errs, err)
					return
				}
				ctx.Ingress = ing
			}
		})
		if err != nil {
			errs = append(errs, err)
		}
		return errs
	})
}
