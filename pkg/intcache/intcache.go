package intcache

import (
	"sync"
	"time"
)

type item struct {
	mu       sync.RWMutex
	expireAt time.Time
	count    int
}

func (it *item) expired(t time.Time) bool {
	it.mu.RLock()
	defer it.mu.RUnlock()
	return t.After(it.expireAt)
}

type Cache struct {
	onEvicted func(any, int)

	cleanupInterval time.Duration
	stop            chan bool

	items sync.Map
}

func (c *Cache) Init(cleanupInterval time.Duration, onEvicted func(any, int)) {
	c.cleanupInterval = cleanupInterval
	c.stop = make(chan bool)
	c.onEvicted = onEvicted
	go c.loopClean()
}

func (c *Cache) Stop() {
	c.stop <- true
}

func (c *Cache) Incr(key string, dur time.Duration) int {
	if val, ok := c.items.Load(key); ok {
		if it, ok := val.(*item); ok && it != nil {
			it.mu.Lock()
			defer it.mu.Unlock()
			it.count++
			return it.count
		}
	}
	c.items.Store(key, &item{expireAt: time.Now().Add(dur), count: 1})
	return 1
}

func (c *Cache) loopClean() {
	ticker := time.NewTicker(c.cleanupInterval)
	for {
		select {
		case <-ticker.C:
			c.deleteExpired()
		case <-c.stop:
			ticker.Stop()
			return
		}
	}
}

func (c *Cache) deleteExpired() {
	now := time.Now()
	if c.onEvicted != nil {
		c.items.Range(func(key, value any) bool {
			if it, ok := value.(*item); ok && it != nil {
				if it.expired(now) {
					if _, loaded := c.items.LoadAndDelete(key); loaded {
						c.onEvicted(key, it.count)
					}
				}
			} else {
				c.items.Delete(key)
			}
			return true
		})
	} else {
		c.items.Range(func(key, value any) bool {
			if it, ok := value.(*item); ok && it != nil {
				if it.expired(now) {
					c.items.Delete(key)
				}
			} else {
				c.items.Delete(key)
			}
			return true
		})
	}
}
