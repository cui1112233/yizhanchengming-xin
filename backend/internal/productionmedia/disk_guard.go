package productionmedia

import (
	"context"
	"fmt"
	"log"
	"time"

	"qiantie/backend/internal/localartifact"
)

const (
	PressureCleanupStartPercent = 75.0
	PressureCleanupStopPercent  = 65.0
	WriteStopPercent            = 85.0
)

type UsageReader func() (float64, error)

func NewDiskWriteGuard(read UsageReader) func() error {
	return func() error {
		if read == nil {
			return nil
		}
		used, err := read()
		if err != nil {
			return fmt.Errorf("read disk usage before media write: %w", err)
		}
		if used >= WriteStopPercent {
			return fmt.Errorf("%w: %.1f%% used", localartifact.ErrDiskPressure, used)
		}
		return nil
	}
}

type CleanupSupervisor struct {
	Purger Purger
	Usage  UsageReader
}

func (s CleanupSupervisor) RunOnce(ctx context.Context) (PurgeSummary, error) {
	total, err := s.Purger.RunOnce(ctx, false)
	if err != nil {
		return total, err
	}
	if s.Usage == nil {
		return total, nil
	}
	used, err := s.Usage()
	if err != nil || used < PressureCleanupStartPercent {
		return total, err
	}
	for used > PressureCleanupStopPercent {
		batch, err := s.Purger.RunOnce(ctx, true)
		if err != nil {
			return total, err
		}
		total.Deleted += batch.Deleted
		total.Retried += batch.Retried
		if batch.Deleted == 0 {
			break
		}
		used, err = s.Usage()
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

func (s CleanupSupervisor) Start(interval time.Duration) {
	if interval <= 0 {
		interval = time.Hour
	}
	go func() {
		run := func() {
			if _, err := s.RunOnce(context.Background()); err != nil {
				log.Printf("production media cleanup failed: %v", err)
			}
		}
		run()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			run()
		}
	}()
}
