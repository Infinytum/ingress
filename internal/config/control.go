package config

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/bep/debounce"
	"github.com/caddyserver/caddy/v2"
	"github.com/go-mojito/mojito/log"
)

var rwlock sync.RWMutex = sync.RWMutex{}
var debounced = debounce.New(1 * time.Second)

// Edit locks the configuration while the editor function is being executed.
// the lock is automatically returned and the caddy config is updated.
func Edit(editor func(config *Config)) {
	rwlock.Lock()
	defer rwlock.Unlock()
	copy := config
	editor(&copy)
	config = copy
	debounced(Reload)
}

// Read returns a copy of the current caddy config.
// Any modifications will be lost and never applied.
func Read() Config {
	copy := config
	return copy
}

func Reload() {
	j, err := json.Marshal(config)
	if err != nil {
		log.Errorf("Failed to marshal config: %v", err)
		return
	}
	err = caddy.Load(j, false)
	if err != nil {
		log.Errorf("Failed to load config: %v", err)
		return
	}
}
