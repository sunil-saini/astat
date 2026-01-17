package cache

import (
	"sync"
	"time"
)

var metaMu sync.Mutex

type ServiceMeta struct {
	LastUpdated time.Time `json:"last_updated"`
	Refreshing  bool      `json:"refreshing"`
	BusyPID     int       `json:"busy_pid"`
}

type Meta struct {
	LastUpdated time.Time              `json:"last_updated"`
	Services    map[string]ServiceMeta `json:"services"`
}

func LockMeta() {
	metaMu.Lock()
}

func UnlockMeta() {
	metaMu.Unlock()
}
