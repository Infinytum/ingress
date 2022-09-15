package kubestore

import (
	"context"

	"github.com/go-mojito/mojito/log"
	"github.com/infinytum/injector"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (k *KubeStore) Delete(ctx context.Context, key string) error {
	client, err := injector.Inject[*kubernetes.Clientset]()
	if err != nil {
		return err
	}

	name := cleanKey(key)
	_, err = client.CoreV1().Secrets(k.Namespace()).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.Field("secret", name).Warn("Secret deletion requested, but secret does not exist. Continuing anyway.")
			return nil
		}
		return err
	}

	return client.CoreV1().Secrets(k.Namespace()).Delete(ctx, name, metav1.DeleteOptions{})
}
