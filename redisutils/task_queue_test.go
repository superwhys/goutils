package redisutils

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/superwhys/goutils/dialer"
)

var (
	queue *TaskQueue
)

type TestObj struct {
	Age int
}

func TestMain(m *testing.M) {
	queue = NewTaskQueue(
		dialer.DialRedisPool("localhost:6379", 10, 100),
		"testQueue",
		&TestObj{},
		WithBucket(6),
	)
	m.Run()
}

func TestRedisTaskQueue(t *testing.T) {
	go func() {
		for i := 0; i < 20; i++ {
			key := fmt.Sprintf("taskkey-%v", i)
			task := &TestObj{Age: i}
			bucket := rand.Intn(6)
			if err := queue.PushToBucket(key, task, bucket, true); err != nil {
				t.Errorf("push to bucket err: %v", err)
				return
			}
		}

		time.Sleep(time.Second * 3)
		queue.Close()
	}()

	for task := range queue.IterTask() {
		fmt.Printf("receive task: %#v\n", task.Payload.(*TestObj))
		time.Sleep(time.Second)
	}
}

func TestRedisTaskQueueDup(t *testing.T) {
	task := &TestObj{Age: 1}

	if err := queue.PushToBucket("dupKey", task, 1, true); err != nil {
		t.Error("push to bucket err")
		return
	}
	if err := queue.PushToBucket("dupKey", task, 1, true); err != nil {
		if !errors.Is(err, ErrDuplicated) {
			t.Error("err is not dup")
			return
		}
	} else {
		t.Error("except a dup err")
	}
}
