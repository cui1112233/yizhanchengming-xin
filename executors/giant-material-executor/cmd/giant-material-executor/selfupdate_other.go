//go:build !windows && !darwin

package main

import (
	"context"
	"sync"

	"qiantie/giant-material-executor/internal/agent"
	"qiantie/giant-material-executor/internal/update"
)

func startSelfUpdater(_ context.Context, _ context.CancelFunc, _ func() agent.Snapshot, _ *sync.RWMutex, _ *string) *update.SelfUpdater {
	return nil
}
