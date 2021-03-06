package etcd

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/hulklab/yago"
	"github.com/hulklab/yago-coms/locker/lock"

	"github.com/hulklab/yago/coms/etcd"
	"go.etcd.io/etcd/clientv3/concurrency"
)

func init() {
	lock.RegisterLocker("etcd", func(name string) lock.ILocker {
		//driver := yago.Config.GetString(name + ".driver")

		driverInsId := yago.Config.GetString(name + ".driver_instance_id")
		retry := yago.Config.GetInt(name + ".retry")
		if retry == 0 {
			retry = 3
		}
		eIns := etcd.Ins(driverInsId)
		val := &etcdLock{
			eIns:  eIns,
			retry: retry,
		}

		return val
	})

}

type etcdLock struct {
	eIns  *etcd.Etcd
	retry int
	key   string
	ctx   context.Context
	mu    sync.Mutex
	mutex *concurrency.Mutex
}

func (e *etcdLock) Lock(key string, opts ...lock.SessionOption) error {
	var ctx context.Context
	ctx = context.Background()
	ops := &lock.SessionOptions{TTL: lock.DefaultSessionTTL}
	for _, opt := range opts {
		opt(ops)
	}

	if ops.WaitTime > 0 {
		var cancelFunc context.CancelFunc

		ctx, cancelFunc = context.WithTimeout(context.Background(), ops.WaitTime)
		defer cancelFunc()
	}

	e.ctx = ctx

	var err error

	for i := 0; i < e.retry; i++ {

		err = e.lock(key, ops.TTL)
		if err == nil {
			break
		}

		if errors.Is(err, context.DeadlineExceeded) {
			break
		}

		log.Printf("etcd lock err:%s,retry:%d", err.Error(), i)
	}

	if err != nil {
		return err
	}

	if ops.DisableKeepAlive {
		go func() {
			<-time.After(time.Duration(ops.TTL) * time.Second)
			e.Unlock()
		}()
	}
	return nil
}

func (e *etcdLock) lock(key string, ttl int64) error {
	response, err := e.eIns.Client.Grant(e.ctx, ttl)
	if err != nil {
		return err
	}

	session, err := concurrency.NewSession(e.eIns.Client, concurrency.WithLease(response.ID))
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
