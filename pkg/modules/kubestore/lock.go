package kubestore

import (
	"context"
	"fmt"
	"time"

	"github.com/infinytum/injector"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

const (
	leaseDuration      = 5 * time.Second
	leaseRenewInterval = 2 * time.Second
	leasePollInterval  = 5 * time.Second
)

func (k *KubeStore) Lock(ctx context.Context, key string) error {
	key = cleanKey(key)
	for {
		_, err := k.tryAcquireOrRenew(ctx, key, false)
		if err == nil {
			go k.keepLockUpdated(ctx, key)
			return nil
		}

		select {
		case <-time.After(leasePollInterval):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (k *KubeStore) keepLockUpdated(ctx context.Context, key string) {
	for {
		time.Sleep(leaseRenewInterval)
		done, err := k.tryAcquireOrRenew(ctx, key, true)
		if err != nil {
			return
		}
		if done {
			return
		}
	}
}

func (k *KubeStore) tryAcquireOrRenew(ctx context.Context, key string, shouldExist bool) (bool, error) {
	client, err := injector.Inject[*kubernetes.Clientset]()
	if err != nil {
		return false, err
	}

	now := metav1.Now()
	lock := resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      key,
			Namespace: k.Namespace(),
		},
		Client: client.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: k.LeaseId,
		},
	}

	ler := resourcelock.LeaderElectionRecord{
		HolderIdentity:       lock.Identity(),
		LeaseDurationSeconds: 5,
		AcquireTime:          now,
		RenewTime:            now,
	}

	currLer, _, err := lock.Get(ctx)

	// 1. obtain or create the ElectionRecord
	if err != nil {
		if !errors.IsNotFound(err) {
			return true, err
		}
		if shouldExist {
			return true, nil // Lock has been released
		}
		if err = lock.Create(ctx, ler); err != nil {
			return true, err
		}
		return false, nil
	}

	// 2. Record obtained, check the Identity & Time
	if currLer.HolderIdentity != "" &&
		currLer.RenewTime.Add(leaseDuration).After(now.Time) &&
		currLer.HolderIdentity != lock.Identity() {
		return true, fmt.Errorf("lock is held by %v and has not yet expired", currLer.HolderIdentity)
	}

	// 3. We're going to try to update the existing one
	if currLer.HolderIdentity == lock.Identity() {
		ler.AcquireTime = currLer.AcquireTime
		ler.LeaderTransitions = currLer.LeaderTransitions
	} else {
		ler.LeaderTransitions = currLer.LeaderTransitions + 1
	}

	if err = lock.Update(ctx, ler); err != nil {
		return true, fmt.Errorf("failed to update lock: %v", err)
	}
	return false, nil
}
