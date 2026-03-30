package service

import (
	"sync"

	"github.com/infinytum/injector"
	"github.com/thoas/go-funk"
)

func init() {
	injector.Singleton(func() *State {
		return &State{
			configuredHosts: make(map[string][]string),
		}
	})
}

type State struct {
	mu              sync.RWMutex
	configuredHosts map[string][]string
}

func (s *State) IsHostConfigured(host string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, hosts := range s.configuredHosts {
		if funk.Contains(hosts, host) {
			return true
		}
	}
	return false
}

func (s *State) SetHosts(uid string, hosts []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.configuredHosts, uid)
	s.configuredHosts[uid] = hosts
}
