package conc

import "sync"

type KeyLock struct {
	giantLock sync.RWMutex
	locks     map[string]*sync.Mutex
}

func NewKeyLock() *KeyLock {
	return &KeyLock{
		giantLock: sync.RWMutex{},
		locks:     map[string]*sync.Mutex{},
	}
}

func (l *KeyLock) getLock(key string) *sync.Mutex {
	l.giantLock.RLock()
	if lock, ok := l.locks[key]; ok {
		l.giantLock.RUnlock()
		return lock
	}

	l.giantLock.RUnlock()
	l.giantLock.Lock()

	if lock, ok := l.locks[key]; ok {
		l.giantLock.Unlock()
		return lock
	}

	lock := &sync.Mutex{}
	l.locks[key] = lock
	l.giantLock.Unlock()
	return lock
}

func (l *KeyLock) Lock(key string) {
	l.getLock(key).Lock()
}

func (l *KeyLock) Unlock(key string) {
	l.getLock(key).Unlock()
}

func (l *KeyLock) KeyLocker(key string) sync.Locker {
	return l.getLock(key)
}

type KeyRWLock struct {
	giantLock sync.RWMutex
	locks     map[string]*sync.RWMutex
}

func NewKeyRWLock() *KeyRWLock {
	return &KeyRWLock{
		giantLock: sync.RWMutex{},
		locks:     map[string]*sync.RWMutex{},
	}
}

func (l *KeyRWLock) getLock(key string) *sync.RWMutex {
	l.giantLock.RLock()
	if lock, ok := l.locks[key]; ok {
		l.giantLock.RUnlock()
		return lock
	}

	l.giantLock.RUnlock()
	l.giantLock.Lock()

	if lock, ok := l.locks[key]; ok {
		l.giantLock.Unlock()
		return lock
	}

	lock := &sync.RWMutex{}
	l.locks[key] = lock
	l.giantLock.Unlock()
	return lock
}

func (l *KeyRWLock) Lock(key string) {
	l.getLock(key).Lock()
}

func (l *KeyRWLock) Unlock(key string) {
	l.getLock(key).Unlock()
}

func (l *KeyRWLock) RLock(key string) {
	l.getLock(key).RLock()
}

func (l *KeyRWLock) RUnlock(key string) {
	l.getLock(key).RUnlock()
}

func (l *KeyRWLock) KeyLocker(key string) sync.Locker {
	return l.getLock(key)
}

func (l *KeyRWLock) KeyRLocker(key string) sync.Locker {
	return l.getLock(key).RLocker()
}
