package productionmedia

import (
	"context"
	"fmt"

	"qiantie/backend/internal/localartifact"
)

type PurgeSummary struct {
	Deleted int
	Retried int
}

type Purger struct {
	Repository Repository
	Local      *localartifact.Store
	MaxPerRun  int
}

func (p Purger) RunOnce(ctx context.Context, pressure bool) (PurgeSummary, error) {
	summary := PurgeSummary{}
	if p.Repository == nil || p.Local == nil {
		return summary, fmt.Errorf("production media purger is not configured")
	}
	limit := p.MaxPerRun
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	for count := 0; count < limit; count++ {
		item, found, err := p.Repository.ClaimDuePurge(ctx, pressure)
		if err != nil {
			return summary, err
		}
		if !found {
			return summary, nil
		}
		if item.State != StateVerified && item.State != StateCleanupPending {
			err = fmt.Errorf("unverified production media cannot be purged")
		} else {
			err = p.Local.Remove(item.LocalStorageRef)
		}
		if err == nil {
			if err := p.Repository.MarkLocalDeleted(ctx, item.ID); err != nil {
				return summary, err
			}
			summary.Deleted++
			continue
		}
		if err := p.Repository.MarkPurgeRetry(ctx, item.ID, err); err != nil {
			return summary, err
		}
		summary.Retried++
	}
	return summary, nil
}
