package service

import (
	"os"

	"github.com/infinytum/injector"
)

var podInfo PodInfo = PodInfo{}

func init() {
	podInfo.Name = os.Getenv("POD_NAME")
	podInfo.Namespace = os.Getenv("POD_NAMESPACE")
	injector.Singleton(newPodInfo)
}

type PodInfo struct {
	// Name of the pod
	Name string `json:"name"`
	// Namespace of the pod
	Namespace string `json:"namespace"`
}

func newPodInfo() PodInfo {
	return podInfo
}
