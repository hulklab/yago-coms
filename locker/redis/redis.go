package redis

import (
	"log"
	"time"

	"github.com/hulklab/yago-coms/locker/lock"

	"github.com/garyburd/redigo/redis"
	"github.com/hulklab/yago"
	"github.com/hulklab/yago/coms/rds"
)

type redisLock struct {
	rIns    *rds.Rds
	retry   int
	key     string
	expired int64
	done    chan struct{}
}

func init() {
	lock.RegisterLocker("redis", func(name string) lock.ILocker {
		driverInsId := yago.Config.GetString(name + ".driver_instance_id")
		retry := yago.Config.GetInt(name + ".retry")
		if retry == 0 {
			retry = 3
		}
		rIns := rds.Ins(driverInsId)
		val := &redisLock{
			rIns:  rIns,
			retry: retry,
		}
		return val
	})
}

func (r *redisLock) autoRenewal(ttl int64) {
	r.done = make(chan struct{})

	go func() {
		ticker := time.NewTicker(time.Duration(ttl) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.done:
				return
			case <-ticker.C:
				expired := r.expired + ttl
				ok, err := redis.String(r.rIns.Set(r.key, expired, "XX"))

				if err != nil {
					log.Printf("lock renewal err: %s\n", err.Error())
					break
				} else if len(ok) == 0 {
					//  续约失败
					log.Printf("lock renewal fail: key %s is not exists", r.key)
					break
				}

				r.expired = expired
			}
		}
	}()

}

func (r *redisLock) LockForever(key string, opts ...lock.SessionOption) error {
	ops := &lock.SessionOptions{TTL: lock.DefaultSessionTTL}
	for _, opt := range opts {
		opt(ops)
	}
	err := r.Lock(key, ops.TTL)
	if err != nil {
		return err
	}

	r.autoRenewal(ops.TTL)

	return nil

}

func (r *redisLock) Lock(key string, timeout int64) error {
	r.key = key
	var err error

	for i := 0; i < r.retry; i++ {
		err = r.lock(timeout)
		if err == nil {
			return nil
		}
		log.Printf("redis lock err:%s,retry:%d", err.Error(), i)
	}

	return err
}

func (r *redisLock) lock(timeout int64) error {
	i := 1
	for {
		t := time.Now().Unix() + timeout
		r.expired = t

		// key 不存在
		lo, err := redis.Int(r.rIns.SetNx(r.key, t))
		if err != nil {
			return err
		}
		if lo == 1 {
			break
		}

		// key 已经超时，并且 getset 获取任务超时
		reply, err := r.rIns.Get(r.key)
		if reply == nil && err == nil {
			continue
		}

		val, err := redis.Int64(reply, err)
		if err != nil {
			return err
		}

		if time.Now().Unix() > val {
			old, err := redis.Int64(r.rIns.GetSet(r.key, t))
			if err != nil {
				return err
			}

			// 超时
			if time.Now().Unix() > old {
				break
			}
		}

		time.Sleep(time.Duration(2*i) * time.Microsecond)
		i++
		// 超过 1 分钟归零
		if i >= 60*1000*1000 {
			i = 1
		}
	}
	return nil

}

func (r *redisLock) Unlock() {
	if r.done != nil {
		r.done <- struct{}{}
	}
	val, _ := redis.Int64(r.rIns.Get(r.key))
	if val > 0 && val == r.expired {
		_, _ = r.rIns.Del(r.key)
	}
}
