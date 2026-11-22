package novelfetchworkshop

import (
	"context"
	"errors"
	"time"
)

const (
	DefaultProcessLeaseTTL       = 10 * time.Minute
	DefaultProcessLeaseHeartbeat = 30 * time.Second
)

var ErrProcessLeaseLost = errors.New("novel fetch process lease lost")

type ProcessLeaseQueue interface {
	RenewLease(context.Context, ProcessJob, time.Duration) error
	OwnsLease(context.Context, ProcessJob) (bool, error)
}
