package batchfactoryexecution

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type OutboxRepository interface {
	Claim(context.Context, int, time.Time) ([]OutboxEvent, error)
	MarkDelivered(context.Context, string, time.Time) error
	MarkFailed(context.Context, string, string, time.Time) error
}

type Dispatcher struct {
	Repository OutboxRepository
	Handler    func(context.Context, OutboxEvent) error
	Now        func() time.Time
	RetryDelay time.Duration
}

func (d *Dispatcher) DispatchOnce(ctx context.Context, limit int) error {
	if d == nil || d.Repository == nil || d.Handler == nil {
		return fmt.Errorf("execution outbox dispatcher is not configured")
	}
	if limit <= 0 {
		limit = 20
	}
	now := time.Now().UTC()
	if d.Now != nil {
		now = d.Now().UTC()
	}
	delay := d.RetryDelay
	if delay <= 0 {
		delay = time.Minute
	}
	events, err := d.Repository.Claim(ctx, limit, now)
	if err != nil {
		return err
	}
	var failures []error
	for _, event := range events {
		if handleErr := d.Handler(ctx, event); handleErr != nil {
			if markErr := d.Repository.MarkFailed(ctx, event.ID, handleErr.Error(), now.Add(delay)); markErr != nil {
				failures = append(failures, fmt.Errorf("event %s failed: %v; persist retry: %w", event.ID, handleErr, markErr))
			} else {
				failures = append(failures, fmt.Errorf("event %s: %w", event.ID, handleErr))
			}
			continue
		}
		if markErr := d.Repository.MarkDelivered(ctx, event.ID, now); markErr != nil {
			failures = append(failures, fmt.Errorf("mark event %s delivered: %w", event.ID, markErr))
		}
	}
	return errors.Join(failures...)
}
