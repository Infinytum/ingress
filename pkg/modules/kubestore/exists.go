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
		log.Error("Failed to inject kubernetes clientset during kubestore.Exists", "error", err)
		return false
	}

	name := generateSecretName(key)
	_, err = client.CoreV1().Secrets(k.Namespace()).Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		if !errors.IsNotFound(err) {
			log.Error("Could not check for secret during kubestore.Exists", "error", err)
		}
		return false
	}

	return true
}
