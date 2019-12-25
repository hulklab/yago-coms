package locker

import (
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/hulklab/yago/coms/rds"
)

type redisLock struct {
	rIns    *rds.Rds
	retry   int
	key     string
	expired int64
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
		lock, err := redis.Int(r.rIns.SetNx(r.key, t))
		if err != nil {
			return err
		}
		if lock == 1 {
			break
		}

		// key 已经超时，并且 getset 获取任务超时
		val, err := redis.Int64(r.rIns.Get(r.key))
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
	val, _ := redis.Int64(r.rIns.Get(r.key))
	if val > 0 && val == r.expired {
		_, _ = r.rIns.Del(r.key)
	}
}
