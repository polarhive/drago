package workflow

import (
	"sync"
)

type KV struct {
	store map[string]any
	mu    sync.RWMutex
}

func NewKV() *KV {
	return &KV{
		store: make(map[string]any),
	}
}

// Get retrieves the value associated with the given key.
func (k *KV) Get(key string) (any, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	val, exists := k.store[key]
	return val, exists
}

// Set associates the given value with the given key.
func (k *KV) Set(key string, value any) {
	k.mu.Lock()
	defer k.mu.Unlock()

	k.store[key] = value
}

// Delete removes the value associated with the given key.
func (k *KV) Delete(key string) {
	k.mu.Lock()
	defer k.mu.Unlock()

	delete(k.store, key)
}
