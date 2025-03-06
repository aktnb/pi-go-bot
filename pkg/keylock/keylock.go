package keylock

import "sync"

type KeyLock struct {
	lockMap    map[string]*lock
	globalLock sync.Mutex
}

type lock struct {
	m        *sync.Mutex
	refCount int
}

func New() *KeyLock {
	return &KeyLock{lockMap: make(map[string]*lock)}
}

func (kl *KeyLock) Lock(k string) {
	kl.globalLock.Lock()

	l, ok := kl.lockMap[k]
	if !ok {
		l = &lock{
			m: new(sync.Mutex),
		}
	}

	l.refCount++

	kl.globalLock.Unlock()

	l.m.Lock()
}

func (kl *KeyLock) Unlock(k string) {
	kl.globalLock.Lock()

	l, ok := kl.lockMap[k]

	if !ok {
		kl.globalLock.Unlock()
		return
	}

	l.refCount--

	if l.refCount <= 0 {
		delete(kl.lockMap, k)
	}

	kl.globalLock.Unlock()

	l.m.Unlock()
}
