package util

import (
	"runtime"
	"sync"
)

func WaitGroup(i int) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(i)
	return wg
}

// Work runs fn() with default parallelism
func Work(fn func()) *sync.WaitGroup {
	return NWork(fn, runtime.GOMAXPROCS(0))
}

// NWork runs fn() with parallelism p
func NWork(fn func(), p int) *sync.WaitGroup {
	wg := WaitGroup(p)
	for i := 0; i < p; i++ {
		go func() {
			fn()
			wg.Done()
		}()
	}
	return wg
}
