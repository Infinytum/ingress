package config

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/bep/debounce"
	"github.com/caddyserver/caddy/v2"
	"log/slog"
)

var rwlock sync.RWMutex = sync.RWMutex{}
var debounced = debounce.New(1 * time.Second)

// Edit locks the configuration while the editor function is being executed.
// the lock is automatically returned and the caddy config is updated.
func Edit(editor func(config *Config)) error {
	rwlock.Lock()
	defer rwlock.Unlock()

	var copy Config
	v, _ := json.Marshal(config)
	json.Unmarshal(v, &copy)
	editor(&copy)

	caddyCfg, err := AsCaddyConfig(copy)
	if err != nil {
		return err
	}
	if err := caddy.Validate(caddyCfg); err != nil {
		return err
	}

	config = copy
	debounced(Reload)
	return nil
}

func AsCaddyConfig(cfg Config) (*caddy.Config, error) {
	j, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	var caddyCfg caddy.Config
	err = json.Unmarshal(j, &caddyCfg)
	if err != nil {
		return nil, err
	}
	return &caddyCfg, nil
}

// Read returns a copy of the current caddy config.
// Any modifications will be lost and never applied.
func Read() Config {
	rwlock.RLock()
	defer rwlock.RUnlock()
	copy := config
	return copy
}

func Reload() {
	rwlock.RLock()
	j, err := json.Marshal(config)
	rwlock.RUnlock()
	if err != nil {
		slog.Error("Failed to marshal config", "error", err)
		return
	}
	err = caddy.Load(j, false)
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		return
	}
}
