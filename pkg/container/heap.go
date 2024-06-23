package container

import "container/heap"

type MinHeap[T any] struct {
	capacity, size int
	items          []T
	less           func(T, T) bool
}

func NewMinHeap[T any](capacity int, less func(T, T) bool) *MinHeap[T] {
	h := &MinHeap[T]{
		capacity: capacity,
		items:    make([]T, 0, capacity),
		less:     less,
	}
	heap.Init(h)
	return h
}

func (h MinHeap[T]) Len() int           { return h.size }
func (h MinHeap[T]) Less(i, j int) bool { return h.less(h.items[i], h.items[j]) }
func (h MinHeap[T]) Swap(i, j int)      { h.items[i], h.items[j] = h.items[j], h.items[i] }

func (h *MinHeap[T]) Push(x any) {
	h.items = append(h.items, x.(T))
	h.size++
}

func (h *MinHeap[T]) Pop() any {
	item := h.items[h.size-1]
	h.items = h.items[:h.size-1]
	h.size--
	return item
}

func (h *MinHeap[T]) TryPush(x T) {
	if h.size < h.capacity {
		heap.Push(h, x)
	} else {
		if h.less(h.items[0], x) {
			h.items[0] = x
			heap.Fix(h, 0)
		}
	}
}