package lock

import "sync"

var locks sync.Map

type ILocker interface {
	Lock(key string, timeout int64) error
	Unlock()
}

type NewFunc func(configId string) ILocker

func RegisterLocker(name string, f NewFunc) {
	locks.Store(name, f)
}

func LoadLocker(name string) (NewFunc, bool) {
	val, b := locks.Load(name)
	if !b {
		return nil, false
	}

	newFunc, _ := val.(NewFunc)
	return newFunc, true

}
