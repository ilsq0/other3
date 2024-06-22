package mmap

import "sync"

func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{mp: make(map[K]V)}
}

type Map[K comparable, V any] struct {
	mp map[K]V
	mu sync.RWMutex
}

func (m *Map[K, V]) Get(k K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, b := m.mp[k]
	return v, b
}

func (m *Map[K, V]) Load(k K, f func() V) V {
	m.mu.RLock()
	v0, ok := m.mp[k]
	m.mu.RUnlock()
	if !ok {
		m.mu.Lock()
		defer m.mu.Unlock()
		if v0, ok = m.mp[k]; !ok {
			v0 = f()
			m.mp[k] = v0
		}
	}
	return v0
}

func (m *Map[K, V]) Put(k K, v V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mp[k] = v
}

func (m *Map[K, V]) Remove(k K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.mp, k)
}

func (m *Map[K, V]) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.mp)
}

func (m *Map[K, V]) Inner() map[K]V {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mp
}

func (m *Map[K, V]) Loop(f func(K, V)) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for k, v := range m.mp {
		f(k, v)
	}
}
