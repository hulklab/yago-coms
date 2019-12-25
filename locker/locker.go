package locker

import (
	"log"

	"github.com/hulklab/yago/coms/etcd"

	"github.com/hulklab/yago"
	"github.com/hulklab/yago/coms/rds"
)

type ILocker interface {
	Lock(key string, timeout int64) error
	Unlock()
}

func New(id ...string) ILocker {
	var name string

	if len(id) == 0 {
		name = "locker"
	} else if len(id) > 0 {
		name = id[0]
	}

	driver := yago.Config.GetString(name + ".driver")
	driverInsId := yago.Config.GetString(name + ".driver_instance_id")
	retry := yago.Config.GetInt(name + ".retry")
	if retry == 0 {
		retry = 3
	}

	var val ILocker
	if driver == "redis" {
		if len(driverInsId) == 0 {
			log.Fatalln("driver_id is required in locker config")
		}

		rIns := rds.Ins(driverInsId)
		val = &redisLock{
			rIns:  rIns,
			retry: retry,
		}
	} else if driver == "etcd" {
		if len(driverInsId) == 0 {
			log.Fatalln("driver_id is required in locker config")
		}

		eIns := etcd.Ins(driverInsId)
		val = &etcdLock{
			eIns:  eIns,
			retry: retry,
		}
	} else {
		log.Fatalf("unsupport driver %s", driver)
	}

	return val
}
