package httpapi

import (
	"context"
	"errors"
	"sync"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

type workshopQueueAction int

const (
	ackWorkshopJob workshopQueueAction = iota
	failWorkshopJob
	retryWorkshopJob
)

// Stop joins the renewer before the worker transfers or releases ownership.
func startWorkshopProcessLease(parent context.Context, queue novelfetchworkshop.ProcessLeaseQueue, job novelfetchworkshop.ProcessJob, interval time.Duration) (context.Context, func() error) {
	if queue == nil {
		return parent, func() error { return context.Cause(parent) }
	}
	if interval <= 0 {
		interval = novelfetchworkshop.DefaultProcessLeaseHeartbeat
	}
	work, cancelWork := context.WithCancelCause(parent)
	heartbeat, stopHeartbeat := context.WithCancel(parent)
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-heartbeat.Done():
				return
			case <-ticker.C:
				err := queue.RenewLease(heartbeat, job, novelfetchworkshop.DefaultProcessLeaseTTL)
				if err != nil {
					if heartbeat.Err() == nil || !errors.Is(err, heartbeat.Err()) {
						cancelWork(err)
					}
					return
				}
			}
		}
	}()
	stop := sync.OnceValue(func() error {
		stopHeartbeat()
		<-done
		err := context.Cause(work)
		cancelWork(nil)
		return err
	})
	return work, stop
}

// Some stores also save inline text in the document. Check cancellation at
// both write entry points, even when the provider ignored its cancelled context.
type workshopProcessLeaseStore struct {
	novelfetchworkshop.Store
	checkOwnership func() error
}

func (s workshopProcessLeaseStore) PutBody(ctx context.Context, owner string, body novelfetchworkshop.BodyRecord) (novelfetchworkshop.BodyRef, error) {
	if err := ctx.Err(); err != nil {
		return novelfetchworkshop.BodyRef{}, err
	}
	if err := s.checkOwnership(); err != nil {
		return novelfetchworkshop.BodyRef{}, err
	}
	return s.Store.PutBody(ctx, owner, body)
}

func (s workshopProcessLeaseStore) PutDocument(ctx context.Context, owner string, doc novelfetchworkshop.Document) error {
	if err := context.Cause(ctx); err != nil {
		return err
	}
	return s.Store.PutDocument(ctx, owner, doc)
}
