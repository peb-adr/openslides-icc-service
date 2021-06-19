package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	// notifyKey is the name of the notify stream name.
	notifyKey = "notify"

	// applauseKey is the name of the redis key for applause.
	applauseKey = "applause"
)

// Redis implements the icc backend by saving the data to redis.
//
// Has to be created with redis.New().
type Redis struct {
	pool         *redis.Pool
	lastNotifyID string
}

// New creates a new initializes redis instance.
func New(addr string) *Redis {
	pool := redis.Pool{
		MaxActive:   100,
		Wait:        true,
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}

	return &Redis{
		pool: &pool,
	}
}

// SendNotify saves a valid icc message.
func (r *Redis) SendNotify(message []byte) error {
	conn := r.pool.Get()
	defer conn.Close()

	_, err := conn.Do("XADD", notifyKey, "*", "content", message)
	if err != nil {
		return fmt.Errorf("xadd: %w", err)
	}
	return nil
}

// ReceiveNotify is a blocking function that receives the messages.
//
// The first call returnes the first icc message, the next call the second
// an so on. If there are no more messages to read, the function blocks
// until there is or the context ist canceled.
//
// It is expected, that only one goroutine is calling this function.
func (r *Redis) ReceiveNotify(ctx context.Context) (message []byte, err error) {
	conn := r.pool.Get()
	defer conn.Close()

	id := r.lastNotifyID
	if id == "" {
		id = "$"
	}

	var data []byte
	received := make(chan struct{})

	go func() {
		id, data, err = stream(conn.Do("XREAD", "COUNT", 1, "BLOCK", "0", "STREAMS", notifyKey, id))
		close(received)
	}()

	select {
	case <-received:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	if id != "" {
		r.lastNotifyID = id
	}

	if err != nil {
		return nil, fmt.Errorf("read notify from redis: %w", err)
	}

	return data, nil
}

// SendApplause saves an applause fro the user. A user can save as many
// applause as he wonts, but the backend only counds it once per user in a
// given time.
func (r *Redis) SendApplause(userID int) error {
	conn := r.pool.Get()
	defer conn.Close()

	if _, err := conn.Do("ZADD", applauseKey, time.Now().Unix(), userID); err != nil {
		return fmt.Errorf("adding applause in redis: %w", err)
	}

	return nil
}

// ReceiveApplause returned all applause since a given time in seconds. Each user is
// only called once.
func (r *Redis) ReceiveApplause(since int64) (int, error) {
	conn := r.pool.Get()
	defer conn.Close()

	// Delete old applause.
	before := time.Now().Add(-time.Minute)
	if _, err := conn.Do("ZREMRANGEBYSCORE", applauseKey, 0, before.Unix()); err != nil {
		return 0, fmt.Errorf("removing old applause from redis: %w", err)
	}

	n, err := redis.Int(conn.Do("ZCOUNT", applauseKey, since, "+inf"))
	if err != nil {
		return 0, fmt.Errorf("getting applause from redis: %w", err)
	}

	return n, nil
}
