package taskruntime

import (
	"context"
	"errors"
	"time"
)

var (
	ErrQueueUnavailable = errors.New("runtime queue unavailable")
	ErrLeaseUnavailable = errors.New("runtime lease store unavailable")
	ErrLeaseNotOwner    = errors.New("runtime lease owner/token mismatch")
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

// ProcessingReclaimer is an optional delivery recovery capability. Business
// execution ownership remains in the durable store and its fencing checks.
type ProcessingReclaimer interface {
	ReclaimExpired(context.Context, time.Time, int) (int, error)
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
