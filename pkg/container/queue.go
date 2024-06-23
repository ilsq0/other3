package container

type FixedSizeQueue[T any] struct {
	queue    []T
	capacity int //容量
	size     int //当前数量
	head     int //第一个的index
}

func NewFixedSizeQueue[T any](capacity int) *FixedSizeQueue[T] {
	return &FixedSizeQueue[T]{
		queue:    make([]T, capacity),
		capacity: capacity,
	}
}

func (fq *FixedSizeQueue[T]) Push(value T) {
	if fq.size < fq.capacity {
		// 队列未满，直接添加到队尾
		fq.queue[(fq.head+fq.size)%fq.capacity] = value
		fq.size++
	} else {
		// 队列已满，替换队首元素并调整队首索引
		fq.queue[fq.head] = value
		fq.head = (fq.head + 1) % fq.capacity
	}
}

func (fq *FixedSizeQueue[T]) GetAll() []T {
	result := make([]T, fq.size)
	index := fq.head
	for i := 0; i < fq.size; i++ {
		result[i] = fq.queue[index]
		index = (index + 1) % fq.capacity
	}
	return result
}
