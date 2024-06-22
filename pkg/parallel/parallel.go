package parallel

import (
	"runtime"
	"sync"
)

func Process(step int, f1 func() (int, int), do func(int, int)) {
	min, max := f1()
	if min == max && min == 0 {
		return
	}
	num := runtime.NumCPU()
	batch := (max - min) / num
	var wg sync.WaitGroup
	wg.Add(num)
	for i := 0; i < num; i++ {
		start := min + batch*i + i
		if start > max {
			wg.Done()
			continue
		}
		end := start + batch
		if end > max {
			end = max
		}
		go func() {
			defer wg.Done()
			batchDeal(start, end, step, do)
		}()
	}
	wg.Wait()
}

func batchDeal(start, end, step int, do func(int, int)) {
	batchEnd := start + step
	if batchEnd >= end {
		batchEnd = end
	}
	do(start, batchEnd)
	if batchEnd < end {
		batchDeal(batchEnd+1, end, step, do)
	}
}
