package queue

import (
	"context"
	"errors"
	"testing"
)

type pingFunc func(context.Context) error

func (f pingFunc) Ping(ctx context.Context) error { return f(ctx) }

func TestRuntimeQueueUnconfiguredIsUnavailable(t *testing.T) {
	status := RuntimeQueueStatusFor("", nil)
	if status.Configured || status.Reachable {
		t.Fatalf("status = %#v, want unconfigured and unreachable", status)
	}
	if status.Reason == "" {
		t.Fatal("expected a safe reason")
	}
}

func TestRuntimeQueueReportsInvalidAddress(t *testing.T) {
	status := RuntimeQueueStatusFor("://bad", nil)
	if !status.Configured || status.Reachable {
		t.Fatalf("status = %#v, want configured and unreachable", status)
	}
	if status.Reason == "" {
		t.Fatal("expected a safe reason")
	}
}

func TestRuntimeQueueUsesInjectedPing(t *testing.T) {
	want := errors.New("redis offline")
	status := RuntimeQueueStatusFor("redis://localhost:6379", pingFunc(func(context.Context) error { return want }))
	if !status.Configured || status.Reachable {
		t.Fatalf("status = %#v, want configured and unreachable", status)
	}
	if status.Reason != "redis unavailable" {
		t.Fatalf("reason = %q, want safe reason", status.Reason)
	}

	status = RuntimeQueueStatusFor("redis://localhost:6379", pingFunc(func(context.Context) error { return nil }))
	if !status.Configured || !status.Reachable || status.Reason != "" {
		t.Fatalf("status = %#v, want reachable", status)
	}
}
