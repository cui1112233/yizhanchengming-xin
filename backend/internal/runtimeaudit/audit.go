package runtimeaudit

import "sync"

type Snapshot struct {
	BridgeHits int            `json:"bridgeHits"`
	Paths      map[string]int `json:"paths"`
}

type Registry struct {
	mu    sync.RWMutex
	paths map[string]int
}

var defaultRegistry = NewRegistry()

func Default() *Registry { return defaultRegistry }

func NewRegistry() *Registry {
	return &Registry{paths: make(map[string]int)}
}

func (r *Registry) RecordBridge(path string) {
	if r == nil || path == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.paths == nil {
		r.paths = make(map[string]int)
	}
	r.paths[path]++
}

func (r *Registry) Snapshot() Snapshot {
	if r == nil {
		return Snapshot{Paths: map[string]int{}}
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	paths := make(map[string]int, len(r.paths))
	total := 0
	for path, count := range r.paths {
		paths[path] = count
		total += count
	}
	return Snapshot{BridgeHits: total, Paths: paths}
}
