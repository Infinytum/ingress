package kubestore

import (
	"context"

	"github.com/go-mojito/mojito/log"
	"github.com/infinytum/injector"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (k *KubeStore) Exists(ctx context.Context, key string) bool {
	client, err := injector.Inject[*kubernetes.Clientset]()
	if err != nil {
		log.Errorf("Failed to inject kubernetes clientset during kubestore.Exists: %s", err)
		return false
	}

	name := cleanKey(key)
	_, err = client.CoreV1().Secrets(k.Namespace()).Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		if !errors.IsNotFound(err) {
			log.Errorf("Could not check for secret during kubestore.Exists: %s", err)
		}
		return false
	}

	return true
}
