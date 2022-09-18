package utils

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func GetAddressesFromService(service *apiv1.Service) []string {
	switch service.Spec.Type {
	case apiv1.ServiceTypeNodePort:
		return nil
	case apiv1.ServiceTypeClusterIP:
		return []string{service.Spec.ClusterIP}
	case apiv1.ServiceTypeExternalName:
		return []string{service.Spec.ExternalName}
	case apiv1.ServiceTypeLoadBalancer:
		{
			var addrs []string
			for _, ingress := range service.Status.LoadBalancer.Ingress {
				if ingress.Hostname != "" {
					addrs = append(addrs, ingress.Hostname)
				} else if ingress.IP != "" {
					addrs = append(addrs, ingress.IP)
				}
			}
			return addrs
		}
	}
	return nil
}

// Copied from https://github.com/kubernetes/kubernetes/pull/95179
func IsSubset(subSet, superSet labels.Set) bool {
	if len(superSet) == 0 {
		return true
	}

	for k, v := range subSet {
		value, ok := superSet[k]
		if !ok {
			return false
		}
		if value != v {
			return false
		}
	}
	return true
}
