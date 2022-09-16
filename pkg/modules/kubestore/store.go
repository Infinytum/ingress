package kubestore

import (
	"context"

	"github.com/infinytum/injector"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (k *KubeStore) Store(ctx context.Context, key string, value []byte) error {
	client, err := injector.Inject[*kubernetes.Clientset]()
	if err != nil {
		return err
	}

	name := generateSecretName(key)
	secret, err := client.CoreV1().Secrets(k.Namespace()).Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		secret, err = client.CoreV1().Secrets(k.Namespace()).Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   name,
				Labels: labels,
			},
			Data: map[string][]byte{},
		}, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}

	secret.Data[dataKey] = value
	secret.Data[nameKey] = []byte(key)
	_, err = client.CoreV1().Secrets(k.Namespace()).Update(ctx, secret, metav1.UpdateOptions{})
	return err
}
