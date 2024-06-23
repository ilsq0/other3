package sort

import (
	"container/heap"
)

func merge[T any](a, b []T, less func(T, T) bool) []T {
	final := []T{}
	i := 0
	j := 0
	for i < len(a) && j < len(b) {
		if less(a[i], b[j]) {
			final = append(final, a[i])
			i++
		} else {
			final = append(final, b[j])
			j++
		}
	}
	for i < len(a) {
		final = append(final, a[i])
		i++
	}
	for j < len(b) {
		final = append(final, b[j])
		j++
	}
	return final
}

func MergeSort[T any](items []T, less func(T, T) bool) []T {
	if len(items) < 2 {
		return items
	}
	first := MergeSort(items[:len(items)/2], less)
	second := MergeSort(items[len(items)/2:], less)
	return merge(first, second, less)
}

func InsertionSort[T any](arr []T, less func(T, T) bool) {
	l := len(arr)
	//从第二个元素开始
	for i := 1; i < l; i++ {
		j := i
		for j > 0 && less(arr[j], arr[j-1]) {
			arr[j-1], arr[j] = arr[j], arr[j-1]
			j--
		}
	}
}

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
