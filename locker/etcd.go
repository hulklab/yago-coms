package locker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/hulklab/yago/coms/etcd"
	"go.etcd.io/etcd/clientv3/concurrency"
)

type etcdLock struct {
	eIns  *etcd.Etcd
	retry int
	key   string
	ctx   context.Context
	mu    sync.Mutex
	mutex *concurrency.Mutex
}

func (e *etcdLock) Lock(key string, timeout int64) error {
	e.ctx = context.Background()
	var err error

	for i := 0; i < e.retry; i++ {
		err = e.lock(key)
		if err == nil {
			break
		}

		log.Printf("etcd lock err:%s,retry:%d", err.Error(), i)
	}

	if err != nil {
		return err
	}

	go func() {
		<-time.After(time.Duration(timeout) * time.Second)
		e.Unlock()
	}()
	return nil
}

func (e *etcdLock) lock(key string) error {
	session, err := concurrency.NewSession(e.eIns.Client)
	if err != nil {
		return err
	}

	mutex := concurrency.NewMutex(session, "/lock_"+key)
	err = mutex.Lock(e.ctx)
	if err != nil {
		return err
	}

	e.mutex = mutex
	return nil
}

func (e *etcdLock) Unlock() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.mutex != nil {
		_ = e.mutex.Unlock(e.ctx)
		e.mutex = nil
	}
}
