package container

import (
	"container/list"
	"log"
	"sync"
	"time"
)

type Int64 interface {
	Int64() int64
}

type warp[T Int64] struct {
	v T
	t time.Time
}

type Lru[T Int64] struct {
	cap, size int
	exp       time.Duration
	list      *list.List
	items     map[int64]*list.Element
	mu        sync.RWMutex
}

func NewLru[T Int64](cap int, exp time.Duration) *Lru[T] {
	return &Lru[T]{
		cap:   cap,
		exp:   exp,
		list:  list.New(),
		items: make(map[int64]*list.Element),
	}
}

// 后面的新
func (l *Lru[T]) Put(v T) {
	id := v.Int64()
	l.mu.Lock()
	defer l.mu.Unlock()
	var wa *warp[T]
	ele, ok := l.items[id]
	if !ok || ele == nil {
		wa = &warp[T]{v, time.Now()}
		if l.size < l.cap {
			l.size++
		} else {
			// 满了，移除第一个
			first := l.list.Front()
			l.list.Remove(first)
			delete(l.items, first.Value.(*warp[T]).v.Int64())
		}
		//放入lru
		ele = l.list.PushBack(wa)
		l.items[id] = ele
	} else {
		// 更新element的value，并移到后面去
		wa = ele.Value.(*warp[T])
		wa.v = v
		wa.t = time.Now()
		l.list.MoveToBack(ele)
	}
}

func (l *Lru[T]) Get(id int64) T {
	var v T
	l.mu.Lock()
	defer l.mu.Unlock()
	if ele, ok := l.items[id]; ok && ele != nil {
		// touch了则移到后面（后面的新
		l.list.MoveToBack(ele)
		wa := ele.Value.(*warp[T])
		wa.t = time.Now()
		v = wa.v
	}
	return v
}

func (l *Lru[T]) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.exp == 0 {
		return
	}
	for id, ele := range l.items {
		if time.Since(ele.Value.(*warp[T]).t) > l.exp {
			l.list.Remove(ele)
			delete(l.items, id)
			l.size--
		}
	}
}

func (l *Lru[T]) Print(s string) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for e := l.list.Front(); e != nil; e = e.Next() {
		log.Println(s, e.Value.(*warp[T]).v.Int64())
	}
}
