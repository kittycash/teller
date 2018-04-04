package util

import "sync"

type SafeInt struct {
	m sync.RWMutex
	v int
}

func (i *SafeInt) Set(v int) {
	i.m.Lock()
	defer i.m.Unlock()
	i.v = v
}

func (i *SafeInt) Val() int {
	i.m.RLock()
	defer i.m.RUnlock()
	return i.v
}

func (i *SafeInt) Inc() int {
	i.m.Lock()
	defer i.m.Unlock()
	i.v += 1
	return i.v
}
