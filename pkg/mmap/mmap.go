package mmap

import "sync"

func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{mp: make(map[K]V)}
}

type Map[K comparable, V any] struct {
	mp map[K]V
	mu sync.RWMutex
}

func (m *Map[K, V]) Lock() {
	m.mu.Lock()
}
func (m *Map[K, V]) Unlock() {
	m.mu.Unlock()
}

func (m *Map[K, V]) RLock() {
	m.mu.RLock()
}
func (m *Map[K, V]) RUnlock() {
	m.mu.RUnlock()
}

func (m *Map[K, V]) Get(k K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Get0(k)
}
func (m *Map[K, V]) Get0(k K) (V, bool) {
	v, b := m.mp[k]
	return v, b
}

func (m *Map[K, V]) Contain(k K) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, b := m.mp[k]
	return b
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

func (m *Map[K, V]) Load0(k K, f func() V) V {
	v0, ok := m.mp[k]
	if !ok {
		v0 = f()
		m.mp[k] = v0
	}
	return v0
}

func (m *Map[K, V]) Put(k K, v V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Put0(k, v)
}
func (m *Map[K, V]) Put0(k K, v V) {
	m.mp[k] = v
}

func (m *Map[K, V]) Remove(k K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Remove0(k)
}
func (m *Map[K, V]) Remove0(k K) {
	delete(m.mp, k)
}

func (m *Map[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Len0()
}
func (m *Map[K, V]) Len0() int {
	return len(m.mp)
}

func (m *Map[K, V]) Mp() map[K]V {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Mp0()
}
func (m *Map[K, V]) Mp0() map[K]V {
	return m.mp
}

func (m *Map[K, V]) Loop(f func(K, V)) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for k, v := range m.mp {
		f(k, v)
	}
}
func (m *Map[K, V]) Loop0(f func(K, V)) {
	for k, v := range m.mp {
		f(k, v)
	}
}

func (m *Map[K, V]) SetMp(mp map[K]V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mp = mp
}
