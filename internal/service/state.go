package service

import (
	"github.com/infinytum/injector"
	"github.com/thoas/go-funk"
)

func init() {
	injector.Singleton(func() *State {
		return &State{
			ConfiguredHosts: make(map[string][]string),
		}
	})
}

type State struct {
	ConfiguredHosts map[string][]string
}

func (s State) IsHostConfigured(host string) bool {
	for _, hosts := range s.ConfiguredHosts {
		if funk.Contains(hosts, host) {
			return true
		}
	}
	return false
}
