package kubestore

import (
	"context"
	"io/fs"

	"github.com/caddyserver/certmagic"
	"github.com/infinytum/injector"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (k *KubeStore) Stat(ctx context.Context, key string) (certmagic.KeyInfo, error) {
	client, err := injector.Inject[*kubernetes.Clientset]()
	if err != nil {
		return certmagic.KeyInfo{}, err
	}

	name := cleanKey(key)
	secret, err := client.CoreV1().Secrets(k.Namespace()).Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			return certmagic.KeyInfo{}, fs.ErrNotExist
		}
		return certmagic.KeyInfo{}, err
	}

	if _, ok := secret.Data[dataKey]; !ok {
		return certmagic.KeyInfo{}, fs.ErrNotExist
	}

	return certmagic.KeyInfo{
		Key:        key,
		Modified:   secret.GetCreationTimestamp().UTC(),
		Size:       int64(len(secret.Data[dataKey])),
		IsTerminal: false,
	}, nil
}
