package kubestore

import (
	"context"

	"github.com/infinytum/injector"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (k *KubeStore) Unlock(ctx context.Context, key string) error {
	client, err := injector.Inject[*kubernetes.Clientset]()
	if err != nil {
		return err
	}

	key = cleanKey(key)
	return client.CoordinationV1().Leases(k.Namespace()).Delete(ctx, key, metav1.DeleteOptions{})
}
