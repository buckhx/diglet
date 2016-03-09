package util

import (
	"sync"
)

func WaitGroup(i int) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(i)
	return wg
}

func Work(fn func(), workers int) *sync.WaitGroup {
	wg := WaitGroup(workers)
	for i := 0; i < workers; i++ {
		go func() {
			fn()
			wg.Done()
		}()
	}
	return wg
}
