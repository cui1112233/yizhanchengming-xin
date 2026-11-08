package productionmedia

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"qiantie/backend/internal/localartifact"
)

type Resolver struct {
	Repository Repository
	Local      *localartifact.Store
	Remote     RemoteStore
	Now        func() time.Time
}

func (r *Resolver) RestoreStorageRef(ctx context.Context, storageRef string) error {
	if r == nil || r.Repository == nil || r.Local == nil || r.Remote == nil {
		return os.ErrNotExist
	}
	item, found, err := r.Repository.FindVerifiedByStorageRef(ctx, storageRef)
	if err != nil {
		return err
	}
	if !found || (item.State != StateVerified && item.State != StateCleanupPending && item.State != StateCleaned) || item.ByteSize < 0 || item.SHA256 == "" {
		return os.ErrNotExist
	}
	body, err := r.Remote.Open(ctx, item.ObjectKey)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return os.ErrNotExist
		}
		return err
	}
	defer body.Close()
	if err := r.Local.Restore(storageRef, body, item.ByteSize, item.SHA256); err != nil {
		return fmt.Errorf("restore production media %s: %w", storageRef, err)
	}
	now := r.Now
	if now == nil {
		now = time.Now
	}
	return r.Repository.MarkRehydrated(ctx, item.ID, now().Add(time.Duration(NormalizeRetentionDays(item.RetentionDays))*24*time.Hour))
}
