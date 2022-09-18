package kubestore

import (
	"context"
	"io/fs"

	"github.com/infinytum/injector"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (k *KubeStore) Load(ctx context.Context, key string) ([]byte, error) {
	client, err := injector.Inject[*kubernetes.Clientset]()
	if err != nil {
		return nil, err
	}

	name := generateSecretName(key)
	namespace := getNamespace(key, k.Namespace())
	dataKey := getDataKey(key)
	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fs.ErrNotExist
		}
		return nil, err
	}

	if _, ok := secret.Data[dataKey]; !ok {
		return nil, fs.ErrNotExist
	}

	return secret.Data[dataKey], nil
}
