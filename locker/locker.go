package locker

import (
	"log"

	"github.com/hulklab/yago"
	"github.com/hulklab/yago/coms/rds"
)

type ILocker interface {
	Lock(key string, timeout int64)
	Unlock(key string)
}

func Ins(id ...string) ILocker {
	var name string

	if len(id) == 0 {
		name = "locker"
	} else if len(id) > 0 {
		name = id[0]
	}

	v := yago.Component.Ins(name, func() interface{} {
		driver := yago.Config.GetString(name + ".driver")
		var val ILocker
		if driver == "redis" {
			redisId := yago.Config.GetString(name + ".driver_instance_id")
			if len(redisId) == 0 {
				log.Fatalln("driver_id is required in locker config")
			}

			rIns := rds.Ins(redisId)
			val = &redisLock{
				rIns: rIns,
			}
		} else {
			log.Fatalf("unsupport driver %s", driver)
		}
		return val
	})

	return v.(ILocker)
}
