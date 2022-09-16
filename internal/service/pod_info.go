package service

import (
	"os"

	"github.com/infinytum/injector"
)

func init() {
	injector.Singleton(newPodInfo)
}

type PodInfo struct {
	// Name of the pod
	Name string `json:"name"`
	// Namespace of the pod
	Namespace string `json:"namespace"`
}

func newPodInfo() PodInfo {
	return PodInfo{
		Name:      os.Getenv("POD_NAME"),
		Namespace: os.Getenv("POD_NAMESPACE"),
	}
}
