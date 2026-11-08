package productionmedia

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"

	"qiantie/backend/internal/localartifact"
)

type MigrationSummary struct {
	Verified int
	Failed   int
}

type Migrator struct {
	Repository Repository
	Local      *localartifact.Store
	Remote     RemoteStore
	Now        func() time.Time
	MaxPerRun  int
}

func (m Migrator) RunOnce(ctx context.Context) (MigrationSummary, error) {
	summary := MigrationSummary{}
	if m.Repository == nil || m.Local == nil || m.Remote == nil {
		return summary, fmt.Errorf("production media migrator is not configured")
	}
	limit := m.MaxPerRun
	if limit <= 0 || limit > 50 {
		limit = 50
	}
	now := m.Now
	if now == nil {
		now = time.Now
	}
	for count := 0; count < limit; count++ {
		item, found, err := m.Repository.ClaimUpload(ctx)
		if err != nil {
			return summary, err
		}
		if !found {
			return summary, nil
		}
		if err := m.uploadOne(ctx, &item, now()); err != nil {
			if markErr := m.Repository.MarkUploadFailed(ctx, item.ID, err); markErr != nil {
				return summary, markErr
			}
			summary.Failed++
			continue
		}
		if err := m.Repository.MarkVerified(ctx, item); err != nil {
			return summary, err
		}
		summary.Verified++
	}
	return summary, nil
}

func (m Migrator) uploadOne(ctx context.Context, item *Object, verifiedAt time.Time) error {
	path, err := m.Local.LocalPath(item.LocalStorageRef)
	if err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	digest := hex.EncodeToString(hash.Sum(nil))
	if err := m.Remote.PutFile(ctx, item.ObjectKey, path, item.MediaType, digest, size); err != nil {
		return err
	}
	remote, err := m.Remote.Head(ctx, item.ObjectKey)
	if err != nil {
		return err
	}
	if remote.ByteSize != size || remote.SHA256 != digest {
		return fmt.Errorf("remote verification mismatch: size=%d/%d sha256=%s/%s", remote.ByteSize, size, remote.SHA256, digest)
	}
	item.ByteSize = size
	item.SHA256 = digest
	item.State = StateVerified
	item.RetentionDays = NormalizeRetentionDays(item.RetentionDays)
	item.RemoteVerifiedAt = verifiedAt
	item.PurgeAfter = verifiedAt.Add(time.Duration(item.RetentionDays) * 24 * time.Hour)
	return nil
}

func (m Migrator) Start(interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}
	go func() {
		_, _ = m.RunOnce(context.Background())
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			_, _ = m.RunOnce(context.Background())
		}
	}()
}
