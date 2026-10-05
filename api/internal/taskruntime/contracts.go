package taskruntime

import (
	"context"
	"errors"
	"time"
)

var (
	ErrQueueUnavailable = errors.New("runtime queue unavailable")
	ErrLeaseUnavailable = errors.New("runtime lease store unavailable")
	ErrLeaseNotOwner = errors.New("runtime lease owner/token mismatch")
)

type Message struct {
	TaskKey string `json:"task_key"`
	Attempt int    `json:"attempt"`
}

type Delivery struct {
	Message Message
	Receipt string
}

type Queue interface {
	Enqueue(context.Context, Message) error
	Claim(context.Context, string, time.Duration) (Delivery, error)
	Ack(context.Context, Delivery) error
	Nack(context.Context, Delivery, time.Duration) error
}

type Lease struct {
	TaskKey      string
	Owner        string
	FencingToken uint64
	ExpiresAt    time.Time
}

type ExpiredLease struct {
	TaskKey      string
	FencingToken uint64
}

type LeaseStore interface {
	Claim(context.Context, string, string, uint64, time.Duration) (Lease, bool, error)
	Renew(context.Context, Lease, time.Duration) (bool, error)
	Release(context.Context, Lease) (bool, error)
	RequeueExpired(context.Context, time.Time, int) ([]ExpiredLease, error)
}

// RedisQueue and RedisLeaseStore deliberately expose the stable runtime
// contracts before the concrete Redis transport is wired in the next GREEN
// slice. MySQL remains the execution authority; these types never grant it.
type RedisQueue struct{}
func NewRedisQueue(string, string) (*RedisQueue, error) { return &RedisQueue{}, nil }
func (*RedisQueue) Close() error { return nil }
func (*RedisQueue) Enqueue(context.Context, Message) error { return ErrQueueUnavailable }
func (*RedisQueue) Claim(context.Context, string, time.Duration) (Delivery, error) { return Delivery{}, ErrQueueUnavailable }
func (*RedisQueue) Ack(context.Context, Delivery) error { return ErrQueueUnavailable }
func (*RedisQueue) Nack(context.Context, Delivery, time.Duration) error { return ErrQueueUnavailable }

type RedisLeaseStore struct{}
func NewRedisLeaseStore(string, string) (*RedisLeaseStore, error) { return &RedisLeaseStore{}, nil }
func (*RedisLeaseStore) Close() error { return nil }
func (*RedisLeaseStore) Claim(context.Context, string, string, uint64, time.Duration) (Lease, bool, error) { return Lease{}, false, ErrLeaseUnavailable }
func (*RedisLeaseStore) Renew(context.Context, Lease, time.Duration) (bool, error) { return false, ErrLeaseUnavailable }
func (*RedisLeaseStore) Release(context.Context, Lease) (bool, error) { return false, ErrLeaseUnavailable }
func (*RedisLeaseStore) RequeueExpired(context.Context, time.Time, int) ([]ExpiredLease, error) { return nil, ErrLeaseUnavailable }
