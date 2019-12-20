package locker

import (
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/hulklab/yago/coms/rds"
)

type redisLock struct {
	rIns *rds.Rds
}

func (r *redisLock) Lock(key string, timeout int64) {

	for {

		t := time.Now().Unix() + timeout + 1

		// key 不存在
		lock, err := redis.Int(r.rIns.SetNx(key, t))
		if err != nil {
			//fmt.Println("set nx err:", err)
			continue

		}
		if lock == 1 {
			break
		}

		// key 已经超时，并且 getset 获取任务超时
		val, err := redis.Int64(r.rIns.Get(key))
		if err != nil {
			//fmt.Println("get err:", err)
			continue
		}
		if time.Now().Unix() > val {
			old, err := redis.Int64(r.rIns.GetSet(key, t))
			if err != nil {
				continue
				//fmt.Println("get set err:", err)
			}
			// 超时
			if time.Now().Unix() > old {
				break
			}
		}

		time.Sleep(10 * time.Microsecond)
	}
}

func (r *redisLock) Unlock(key string) {
	val, _ := redis.Int64(r.rIns.Get(key))
	if val > 0 && val > time.Now().Unix() {
		_, _ = r.rIns.Del(key)
	}
}
