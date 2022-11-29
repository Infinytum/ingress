package service

import (
	"time"

	"github.com/go-mojito/mojito/log"
	"github.com/infinytum/injector"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// high enough QPS to fit all expected use cases.
	defaultQPS = 1e6

	// high enough Burst to fit all expected use cases.
	defaultBurst = 1e6
)

func init() {
	injector.DeferredSingleton(newClientSet)
}

func createApiserverClient(config IngressConfig) (*kubernetes.Clientset, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", config.KubeConfig)
	if err != nil {
		return nil, err
	}

	log.Infof("Creating API client for %s", cfg.Host)

	cfg.QPS = defaultQPS
	cfg.Burst = defaultBurst
	cfg.ContentType = "application/vnd.kubernetes.protobuf"

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	// The client may fail to connect to the API server on the first request
	defaultRetry := wait.Backoff{
		Steps:    10,
		Duration: 1 * time.Second,
		Factor:   1.5,
		Jitter:   0.1,
	}

	var retries int
	var lastErr error

	err = wait.ExponentialBackoff(defaultRetry, func() (bool, error) {
		v, err := client.Discovery().ServerVersion()
		if err == nil {
			log.Infof("Discovered kubernetes API server version %s", v.String())
			return true, nil
		}

		lastErr = err
		log.Infof("Unexpected error discovering Kubernetes version (attempt %v): %v", retries, err)
		retries++
		return false, nil
	})

	// err is returned in case of timeout in the exponential backoff (ErrWaitTimeout)
	if err != nil {
		return nil, lastErr
	}

	if retries > 0 {
		log.Warnf("Initial connection to the Kubernetes API server was retried %d times.", retries)
	}

	return client, nil
}

func newClientSet(config IngressConfig) *kubernetes.Clientset {
	c, err := createApiserverClient(config)
	if err != nil {
		log.Fatal(err)
	}
	if c == nil {
		log.Fatal("Kubernetes Clientset was nil")
	}
	return c
}
