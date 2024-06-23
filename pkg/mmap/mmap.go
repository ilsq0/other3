package mmap

import "sync"

func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{mp: make(map[K]V)}
}

type Map[K comparable, V any] struct {
	mp map[K]V
	sync.RWMutex
}

func (m *Map[K, V]) Get(k K) (V, bool) {
	m.RLock()
	defer m.RUnlock()
	v, b := m.mp[k]
	return v, b
}

func (m *Map[K, V]) Load(k K, f func() V) V {
	m.RLock()
	v0, ok := m.mp[k]
	m.RUnlock()
	if !ok {
		m.Lock()
		defer m.Unlock()
		if v0, ok = m.mp[k]; !ok {
			v0 = f()
			m.mp[k] = v0
		}
	}
	return v0
}

func (m *Map[K, V]) Put(k K, v V) {
	m.Lock()
	defer m.Unlock()
	m.mp[k] = v
}

func (m *Map[K, V]) Remove(k K) {
	m.Lock()
	defer m.Unlock()
	delete(m.mp, k)
}

func (m *Map[K, V]) Count() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.mp)
}

func (m *Map[K, V]) Inner() map[K]V {
	m.RLock()
	defer m.RUnlock()
	return m.mp
}

func (m *Map[K, V]) Loop(f func(K, V)) {
	m.RLock()
	defer m.RUnlock()
	for k, v := range m.mp {
		f(k, v)
	}
}
