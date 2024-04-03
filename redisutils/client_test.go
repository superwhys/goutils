package redisutils

import (
	"fmt"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/superwhys/goutils/dialer"
)

func TestClientCommandDo(t *testing.T) {
	client := NewRedisClient(dialer.DialRedisPool("localhost:6379", 12, 100))

	res, err := redis.String(client.Do("set", "test-key", 10))
	if err != nil {
		t.Errorf("command do err: %v", err)
		return
	}
	fmt.Println("command resp: ", res)

	res, err = redis.String(client.Do("get", "test-key"))
	if err != nil {
		t.Errorf("command do err: %v", err)
		return
	}
	fmt.Println("command resp: ", res)
	if res != "10" {
		t.Error("resp no equal")
		return
	}
}

func TestClientLock(t *testing.T) {
	client := NewRedisClient(dialer.DialRedisPool("localhost:6379", 12, 100))

	go func() {
		if err := client.Lock("testLock"); err != nil {
			t.Errorf("lock error: %v", err)
			return
		}

		fmt.Println("after lockA")
		time.Sleep(time.Second * 3)
		client.UnLock("testLock")
	}()

	time.Sleep(time.Second)

	if err := client.LockWithBlock("testLock", 10); err != nil {
		t.Errorf("lock error: %v", err)
		return
	}

	fmt.Println("after lockB")
}
