package kubestore

import (
	"context"
	"strings"

	"github.com/infinytum/injector"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubelabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func (k *KubeStore) List(ctx context.Context, prefix string, recursive bool) ([]string, error) {
	client, err := injector.Inject[*kubernetes.Clientset]()
	if err != nil {
		return nil, err
	}

	secrets, err := client.CoreV1().Secrets(k.Namespace()).List(ctx, metav1.ListOptions{
		LabelSelector: kubelabels.SelectorFromSet(labels).String(),
	})

	if err != nil {
		return nil, err
	}

	var keys []string
	for _, secret := range secrets.Items {
		if strings.HasPrefix(secret.Name, prefix) {
			keys = append(keys, secret.Name)
		}
	}

	return keys, nil
}
