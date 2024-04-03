package redisutils

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/superwhys/goutils/dialer"
)

type TestObj struct {
	Age int
}

func TestRedisTaskQueue(t *testing.T) {
	queue := NewTaskQueue(
		dialer.DialRedisPool("localhost:6379", 10, 100),
		"testQueue",
		&TestObj{},
		WithBucket(6),
	)

	go func() {
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("taskkey-%v", i)
			task := &TestObj{Age: i}
			bucket := rand.Intn(6)
			queue.PushToBucket(key, task, bucket, true)
		}

		time.Sleep(time.Second * 3)
		queue.Close()
	}()

	for task := range queue.IterTask() {
		fmt.Printf("receive task: %#v\n", task.Payload.(*TestObj))
	}
}
