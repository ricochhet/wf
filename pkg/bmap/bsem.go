package bmap

import (
	"sync"
)

type Bsem struct {
	mutex sync.RWMutex
	sem   sync.WaitGroup
}

// Add adds a the specified delta to the wait group.
func (bsem *Bsem) Add(delta int) {
	bsem.mutex.Lock()
	bsem.sem.Add(delta)
	bsem.mutex.Unlock()
}

// Done calls done on the wait group.
func (bsem *Bsem) Done() {
	bsem.sem.Done()
}

// Wait calls wait on the wait group.
func (bsem *Bsem) Wait() {
	bsem.mutex.RLock()
	bsem.sem.Wait()
	bsem.mutex.RUnlock()
}
