package sort

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
