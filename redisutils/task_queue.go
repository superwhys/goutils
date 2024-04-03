package redisutils

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/superwhys/goutils/lg"
	"github.com/superwhys/goutils/slices"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type Task struct {
	Key     string
	Payload any
}

type TaskQueue struct {
	ctx      context.Context
	cancel   func()
	rc       *RedisClient
	name     string
	taskTmpl reflect.Type
	buckets  []int
}

type QueueOption func(*TaskQueue)

func WithBucket(size int) QueueOption {
	return func(tq *TaskQueue) {
		if size <= 0 {
			lg.Fatal("Bucket size is invalid", size)
		}

		for i := 0; i < size; i++ {
			tq.buckets = append(tq.buckets, i)
		}
	}
}

func NewTaskQueue(pool *redis.Pool, queueName string, taskObj any, opts ...QueueOption) *TaskQueue {
	t := reflect.TypeOf(taskObj)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		lg.Fatal("NewTaskQueue: typeObj should be ptr to struct")
	}

	ctx, cancel := context.WithCancel(context.TODO())

	q := &TaskQueue{
		ctx:      ctx,
		cancel:   cancel,
		rc:       NewRedisClient(pool),
		name:     queueName,
		taskTmpl: t.Elem(),
	}

	for _, opt := range opts {
		opt(q)
	}

	// if no buckets provide, use `0` as default
	if len(q.buckets) == 0 {
		q.buckets = append(q.buckets, 0)
	}

	q.buckets = slices.DupInt(q.buckets)
	return q
}

func (q *TaskQueue) genWipKey() string {
	return fmt.Sprintf("wip:%v", q.name)
}

func (q *TaskQueue) genQueueKey(bucket int) string {
	return fmt.Sprintf("queue:%v:%v", q.name, bucket)
}

func (q *TaskQueue) checkWorkInProcess(conn redis.Conn, key string) error {
	exists, err := redis.Bool(conn.Do("SISMEMBER", q.genWipKey(), key))
	if err != nil {
		return errors.Wrap(err, "SISMEMBER")
	}

	if exists {
		return ErrDuplicated
	}
	return nil
}

func (q *TaskQueue) pushToBucket(conn redis.Conn, bucket int, key string, val []byte) error {
	// open a transaction
	if err := conn.Send("MULTI"); err != nil {
		return errors.Wrap(err, "multi")
	}

	if err := conn.Send("RPUSH", q.genQueueKey(bucket), val); err != nil {
		return errors.Wrap(err, "push list")
	}

	if err := conn.Send("SADD", q.genWipKey(), key); err != nil {
		return errors.Wrap(err, "add set")
	}

	if err := conn.Send("EXEC"); err != nil {
		return errors.Wrap(err, "exec")
	}

	if err := conn.Flush(); err != nil {
		return errors.Wrap(err, "flush")
	}

	return nil
}

func (q *TaskQueue) PushToBucket(key string, obj any, bucket int, noDup bool) error {
	if bucket < 0 {
		return errors.Errorf("Bucket: %v incorrect", bucket)
	}
	conn, err := q.rc.GetConnWithContext(q.ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	// if noDup. it need to check whether this key is in process
	if noDup {
		err = q.checkWorkInProcess(conn, key)
		if err != nil {
			return err
		}
	}

	if ot := reflect.TypeOf(obj); ot != reflect.PtrTo(q.taskTmpl) {
		return errors.Errorf("object tpye not correct: %v", ot)
	}

	task := &Task{
		Key:     key,
		Payload: obj,
	}

	b, err := msgpack.Marshal(task)
	if err != nil {
		return errors.Wrap(err, "encode")
	}

	if err := q.pushToBucket(conn, bucket, key, b); err != nil {
		return errors.Wrap(err, "pushToBucket")
	}

	return nil
}

func (q *TaskQueue) Push(key string, obj any) error {
	return q.PushToBucket(key, obj, 0, false)
}

func (q *TaskQueue) IterTask() chan *Task {
	c := make(chan *Task)
	go q.iter(c)
	return c
}

func (q *TaskQueue) checkCancel() bool {
	select {
	case <-q.ctx.Done():
		return true
	default:
		return false
	}
}

func (q *TaskQueue) iter(c chan *Task) {
	conn, err := q.rc.GetConnWithContext(q.ctx)
	if err != nil {
		lg.Errorf("get conn err: %v", err)
		return
	}
	defer conn.Close()

	for {
		if q.checkCancel() {
			close(c)
			return
		}

		bulk, err := redis.Values(q.blpop(conn))
		if err != nil || len(bulk) != 2 {
			lg.Errorf("queue %v blpop err(maybe timeout): %v", q.name, err)
			time.Sleep(time.Second * 2)
			conn.Close()
			conn, err = q.rc.GetConnWithContext(q.ctx)
			if err != nil {
				lg.Errorf("get conn err: %v", err)
				return
			}
			continue
		}

		if q.checkCancel() {
			close(c)
			return
		}

		task, err := q.parseQueueData(bulk)
		if err != nil {
			lg.Errorf("parse bulk data error: %v", err)
			continue
		}

		if err := q.removeWip(conn, task.Key); err != nil {
			lg.Errorf("remove task process status error: %v", err)
			continue
		}
		c <- task
	}
}

func (q *TaskQueue) removeWip(conn redis.Conn, key string) error {
	if _, err := redis.Bool(conn.Do("SREM", q.genWipKey(), key)); err != nil {
		return errors.Wrap(err, "redis srem")
	}
	return nil
}

func (q *TaskQueue) parseQueueData(bulk []any) (*Task, error) {
	task := &Task{
		Payload: reflect.New(q.taskTmpl).Interface(),
	}

	if err := msgpack.Unmarshal(bulk[1].([]byte), task); err != nil {
		return nil, errors.Wrap(err, "decode")
	}

	return task, nil
}

func (q *TaskQueue) blpop(conn redis.Conn) (any, error) {
	keys := make([]any, len(q.buckets)+1)
	for i, b := range rand.Perm(len(q.buckets)) {
		keys[i] = q.genQueueKey(q.buckets[b])
	}

	// timeout
	keys[len(q.buckets)] = 5

	return conn.Do("BLPOP", keys...)
}

func (q *TaskQueue) Close() error {
	q.cancel()

	conn := q.rc.GetConn()
	defer conn.Close()

	_, err := redis.Int(conn.Do("DEL", "wip:"+q.name, "queue:"+q.name))
	return err
}
