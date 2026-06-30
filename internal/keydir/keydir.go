package keydir

import "sync"

type KeyDir struct {
	mu   sync.RWMutex
	data map[string]Entry
}

func New() *KeyDir {
	return &KeyDir{
		data: make(map[string]Entry),
	}
}

func (k *KeyDir) Put(key string, entry Entry) {
	k.mu.Lock()
	defer k.mu.Unlock()

	k.data[key] = entry
}

func (k *KeyDir) Get(key string) (Entry, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	e, ok := k.data[key]
	return e, ok
}
