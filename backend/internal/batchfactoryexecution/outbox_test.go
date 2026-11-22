package batchfactoryexecution

import (
	"context"
	"errors"
	"testing"
	"time"
)

type memoryOutboxRepository struct {
	events []OutboxEvent
	failed map[string]string
}

func (r *memoryOutboxRepository) Claim(_ context.Context, limit int, now time.Time) ([]OutboxEvent, error) {
	var claimed []OutboxEvent
	for i := range r.events {
		e := &r.events[i]
		if e.DeliveredAt != nil || e.AvailableAt.After(now) {
			continue
		}
		if e.ClaimedAt != nil && e.ClaimedAt.After(now.Add(-time.Minute)) {
			continue
		}
		claimedAt := now
		e.ClaimedAt = &claimedAt
		e.Attempts++
		claimed = append(claimed, *e)
		if len(claimed) >= limit {
			break
		}
	}
	return claimed, nil
}
func (r *memoryOutboxRepository) MarkDelivered(_ context.Context, id string, at time.Time) error {
	for i := range r.events {
		if r.events[i].ID == id {
			r.events[i].DeliveredAt = &at
			return nil
		}
	}
	return ErrNotFound
}
func (r *memoryOutboxRepository) MarkFailed(_ context.Context, id, message string, next time.Time) error {
	if r.failed == nil {
		r.failed = map[string]string{}
	}
	r.failed[id] = message
	for i := range r.events {
		if r.events[i].ID == id {
			r.events[i].ClaimedAt = nil
			r.events[i].AvailableAt = next
			r.events[i].LastError = message
			return nil
		}
	}
	return ErrNotFound
}

func TestDispatcherDeliversEachEventOnce(t *testing.T) {
	now := time.Date(2026, 11, 22, 9, 0, 0, 0, time.UTC)
	repo := &memoryOutboxRepository{events: []OutboxEvent{{ID: "event-1", AvailableAt: now}}}
	calls := 0
	d := Dispatcher{Repository: repo, Handler: func(context.Context, OutboxEvent) error { calls++; return nil }, Now: func() time.Time { return now }}
	if err := d.DispatchOnce(context.Background(), 10); err != nil {
		t.Fatal(err)
	}
	if err := d.DispatchOnce(context.Background(), 10); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || repo.events[0].DeliveredAt == nil {
		t.Fatalf("calls=%d event=%#v", calls, repo.events[0])
	}
}

func TestDispatcherRetriesFailedAndStaleClaims(t *testing.T) {
	now := time.Date(2026, 11, 22, 9, 0, 0, 0, time.UTC)
	stale := now.Add(-2 * time.Minute)
	repo := &memoryOutboxRepository{events: []OutboxEvent{{ID: "event-1", AvailableAt: now.Add(-time.Hour), ClaimedAt: &stale}}}
	calls := 0
	d := Dispatcher{Repository: repo, Handler: func(context.Context, OutboxEvent) error { calls++; return errors.New("temporary") }, Now: func() time.Time { return now }, RetryDelay: time.Minute}
	if err := d.DispatchOnce(context.Background(), 10); err == nil {
		t.Fatal("expected dispatch error")
	}
	if calls != 1 || repo.events[0].ClaimedAt != nil || !repo.events[0].AvailableAt.Equal(now.Add(time.Minute)) {
		t.Fatalf("calls=%d event=%#v", calls, repo.events[0])
	}
}

func TestDispatcherRejectsMissingDependencies(t *testing.T) {
	if err := (&Dispatcher{}).DispatchOnce(context.Background(), 1); err == nil {
		t.Fatal("expected configuration error")
	}
}
