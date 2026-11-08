package imagegenerationcache

import (
	"context"
	"time"
)

const (
	StateGenerating = "generating"
	StateReady      = "ready"
	StateFailed     = "failed"

	ClaimAcquired = "acquired"
	ClaimReady    = "ready"
	ClaimBusy     = "busy"
)

type Timing struct {
	Queue    time.Duration
	Header   time.Duration
	Body     time.Duration
	Download time.Duration
	Save     time.Duration
	Total    time.Duration
}

type Entry struct {
	ID         int64
	State      string
	StorageRef string
	MediaType  string
	LeaseUntil time.Time
	Timing     Timing
}

type Repository interface {
	Claim(context.Context, string, string, string, time.Duration) (Entry, string, error)
	MarkReady(context.Context, string, string, string, string, Timing) error
	MarkFailed(context.Context, string, string, error) error
}
