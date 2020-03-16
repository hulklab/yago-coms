package locker

import (
	"fmt"
	"testing"
	"time"

	"github.com/hulklab/yago/example/app/g"

	"github.com/hulklab/yago"
	// un-comment before test
	_ "github.com/hulklab/yago-coms/locker/etcd"
)

// go test -v . -args "-c=${APP_PATH}/app.toml"

func TestRedis(t *testing.T) {
	yago.Config.Set("locker", g.Hash{
		"driver":             "redis",
		"driver_instance_id": "redis",
	})
	yago.Config.Set("redis", g.Hash{
		"addr": "127.0.0.1:6379",
	})

	doTest()
}

func doTest() {
	key := "lock_test"

	go func() {
		r1 := New()
		err := r1.Lock(key, 5)
		if err != nil {
			fmt.Println("get lock in fun1 err", err.Error())
			return
		}
		defer r1.Unlock()
		fmt.Println("get lock in fun1")
		for i := 0; i < 10; i++ {
			fmt.Println("fun1:", i)
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		r2 := New()
		err := r2.Lock(key, 5)
		if err != nil {
			fmt.Println("get lock in fun2 err", err.Error())
			return
		}
		defer r2.Unlock()
		fmt.Println("get lock in fun2")
		for i := 0; i < 10; i++ {
			fmt.Println("fun2:", i)
			time.Sleep(1 * time.Second)
		}
	}()

	time.Sleep(12 * time.Second)

	r3 := New()
	err := r3.Lock(key, 10)
	if err != nil {
		fmt.Println("get lock in fun3 err:", err.Error())
		return
	}

	fmt.Println("get lock in fun3")
	r3.Unlock()

}

func TestEtcd(t *testing.T) {
	yago.Config.Set("locker", g.Hash{
		"driver":             "etcd",
		"driver_instance_id": "etcd",
	})
	yago.Config.Set("etcd", g.Hash{
		"endpoints": []string{"127.0.0.1:2379"},
	})

	doTest()
}
