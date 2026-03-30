package utils

import (
	"testing"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func TestGetAddressesFromService_ClusterIP(t *testing.T) {
	svc := &apiv1.Service{
		Spec: apiv1.ServiceSpec{
			Type:      apiv1.ServiceTypeClusterIP,
			ClusterIP: "10.0.0.1",
		},
	}
	addrs := GetAddressesFromService(svc)
	if len(addrs) != 1 || addrs[0] != "10.0.0.1" {
		t.Errorf("expected [10.0.0.1], got %v", addrs)
	}
}

func TestGetAddressesFromService_ExternalName(t *testing.T) {
	svc := &apiv1.Service{
		Spec: apiv1.ServiceSpec{
			Type:         apiv1.ServiceTypeExternalName,
			ExternalName: "example.com",
		},
	}
	addrs := GetAddressesFromService(svc)
	if len(addrs) != 1 || addrs[0] != "example.com" {
		t.Errorf("expected [example.com], got %v", addrs)
	}
}

func TestGetAddressesFromService_NodePort(t *testing.T) {
	svc := &apiv1.Service{
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeNodePort,
		},
	}
	addrs := GetAddressesFromService(svc)
	if addrs != nil {
		t.Errorf("expected nil, got %v", addrs)
	}
}

func TestGetAddressesFromService_LoadBalancerIP(t *testing.T) {
	svc := &apiv1.Service{
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeLoadBalancer,
		},
		Status: apiv1.ServiceStatus{
			LoadBalancer: apiv1.LoadBalancerStatus{
				Ingress: []apiv1.LoadBalancerIngress{
					{IP: "203.0.113.1"},
				},
			},
		},
	}
	addrs := GetAddressesFromService(svc)
	if len(addrs) != 1 || addrs[0] != "203.0.113.1" {
		t.Errorf("expected [203.0.113.1], got %v", addrs)
	}
}

func TestGetAddressesFromService_LoadBalancerHostname(t *testing.T) {
	svc := &apiv1.Service{
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeLoadBalancer,
		},
		Status: apiv1.ServiceStatus{
			LoadBalancer: apiv1.LoadBalancerStatus{
				Ingress: []apiv1.LoadBalancerIngress{
					{Hostname: "lb.example.com"},
				},
			},
		},
	}
	addrs := GetAddressesFromService(svc)
	if len(addrs) != 1 || addrs[0] != "lb.example.com" {
		t.Errorf("expected [lb.example.com], got %v", addrs)
	}
}

func TestGetAddressesFromService_LoadBalancerBothIPAndHostname(t *testing.T) {
	svc := &apiv1.Service{
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeLoadBalancer,
		},
		Status: apiv1.ServiceStatus{
			LoadBalancer: apiv1.LoadBalancerStatus{
				Ingress: []apiv1.LoadBalancerIngress{
					{IP: "203.0.113.1", Hostname: "lb.example.com"},
				},
			},
		},
	}
	addrs := GetAddressesFromService(svc)
	// Hostname takes precedence over IP in the code
	if len(addrs) != 1 || addrs[0] != "lb.example.com" {
		t.Errorf("expected [lb.example.com], got %v", addrs)
	}
}

func TestGetAddressesFromService_LoadBalancerMultiple(t *testing.T) {
	svc := &apiv1.Service{
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeLoadBalancer,
		},
		Status: apiv1.ServiceStatus{
			LoadBalancer: apiv1.LoadBalancerStatus{
				Ingress: []apiv1.LoadBalancerIngress{
					{IP: "203.0.113.1"},
					{IP: "203.0.113.2"},
				},
			},
		},
	}
	addrs := GetAddressesFromService(svc)
	if len(addrs) != 2 {
		t.Errorf("expected 2 addresses, got %v", addrs)
	}
}

func TestGetAddressesFromService_LoadBalancerEmpty(t *testing.T) {
	svc := &apiv1.Service{
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeLoadBalancer,
		},
	}
	addrs := GetAddressesFromService(svc)
	if len(addrs) != 0 {
		t.Errorf("expected empty, got %v", addrs)
	}
}

func TestIsSubset_BothEmpty(t *testing.T) {
	if !IsSubset(labels.Set{}, labels.Set{}) {
		t.Error("empty subset of empty superset should be true")
	}
}

func TestIsSubset_EmptySuperSet(t *testing.T) {
	// Note: IsSubset returns true when superSet is empty (matches k8s behavior)
	if !IsSubset(labels.Set{"a": "1"}, labels.Set{}) {
		t.Error("any set is a subset of empty superset per implementation")
	}
}

func TestIsSubset_EmptySubSet(t *testing.T) {
	if !IsSubset(labels.Set{}, labels.Set{"a": "1"}) {
		t.Error("empty subset should always be true")
	}
}

func TestIsSubset_ProperSubset(t *testing.T) {
	sub := labels.Set{"a": "1"}
	super := labels.Set{"a": "1", "b": "2"}
	if !IsSubset(sub, super) {
		t.Error("proper subset should return true")
	}
}

func TestIsSubset_ExactMatch(t *testing.T) {
	s := labels.Set{"a": "1", "b": "2"}
	if !IsSubset(s, s) {
		t.Error("equal sets should return true")
	}
}

func TestIsSubset_MissingKey(t *testing.T) {
	sub := labels.Set{"a": "1", "c": "3"}
	super := labels.Set{"a": "1", "b": "2"}
	if IsSubset(sub, super) {
		t.Error("subset with key not in superset should return false")
	}
}

func TestIsSubset_DifferentValues(t *testing.T) {
	sub := labels.Set{"a": "1"}
	super := labels.Set{"a": "2"}
	if IsSubset(sub, super) {
		t.Error("same key different value should return false")
	}
}

func TestGetAddressesFromService_LoadBalancerNoIPNoHostname(t *testing.T) {
	svc := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeLoadBalancer,
		},
		Status: apiv1.ServiceStatus{
			LoadBalancer: apiv1.LoadBalancerStatus{
				Ingress: []apiv1.LoadBalancerIngress{
					{}, // neither IP nor hostname
				},
			},
		},
	}
	addrs := GetAddressesFromService(svc)
	if len(addrs) != 0 {
		t.Errorf("expected empty for ingress with no IP or hostname, got %v", addrs)
	}
}
