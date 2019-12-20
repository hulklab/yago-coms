package locker

import (
	"fmt"
	"testing"
	"time"

	"github.com/hulklab/yago/example/app/g"

	"github.com/hulklab/yago"
)

// go test -v . -test.run TestRedis

func TestRedis(t *testing.T) {
	yago.Config.Set("locker", g.Hash{
		"driver":             "redis",
		"driver_instance_id": "redis",
	})
	yago.Config.Set("redis", g.Hash{
		"addr": "127.0.0.1:6379",
	})

	key := "lock_test"

	r1 := Ins("locker")
	go func() {
		r1.Lock(key, 10)
		defer r1.Unlock(key)
		fmt.Println("get lock in fun1")
	}()

	go func() {
		r1.Lock(key, 10)
		defer r1.Unlock(key)
		fmt.Println("get lock in fun2")
	}()

	for i := 0; i < 13; i++ {
		fmt.Println(i)
		time.Sleep(1 * time.Second)
	}
	r1.Unlock(key)
	r1.Lock(key, 10)
	fmt.Println("get lock in fun3")
	r1.Unlock(key)
}
