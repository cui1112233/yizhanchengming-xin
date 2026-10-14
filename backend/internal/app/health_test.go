package app

import (
	"context"
	"errors"
	"testing"
)

type healthProbeFunc func(context.Context) error

func (f healthProbeFunc) PingContext(ctx context.Context) error {
	return f(ctx)
}

func TestDatabaseHealthReportsMySQLPingFailure(t *testing.T) {
	want := errors.New("mysql is unavailable")
	health := databaseHealth(healthProbeFunc(func(context.Context) error { return want }))

	if err := health(context.Background()); !errors.Is(err, want) {
		t.Fatalf("health error = %v, want %v", err, want)
	}
}

func TestDatabaseHealthSucceedsWhenMySQLResponds(t *testing.T) {
	health := databaseHealth(healthProbeFunc(func(context.Context) error { return nil }))

	if err := health(context.Background()); err != nil {
		t.Fatalf("health error = %v, want nil", err)
	}
}
