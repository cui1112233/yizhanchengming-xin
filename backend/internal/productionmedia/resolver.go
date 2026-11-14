package productionmedia

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	// 本地已有一份尺寸一致的文件时只重试仓库标记，不再重复下载。上次恢复
	// 成功但 MarkRehydrated 失败时会走这里，避免每次访问都重新拉一遍 TOS。
	if r.localCopyReady(storageRef, item.ByteSize) {
		if err := r.markRehydrated(ctx, item); err != nil {
			// 文件已经可用，读取不应因为仓库标记失败而报错；记录日志，下一次
			// 访问会继续重试标记直到数据库恢复。
			log.Printf("production media rehydrate mark failed ref=%s id=%d err=%v", storageRef, item.ID, err)
		}
		return nil
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
	if err := r.markRehydrated(ctx, item); err != nil {
		log.Printf("production media rehydrate mark failed ref=%s id=%d err=%v", storageRef, item.ID, err)
	}
	return nil
}

func (r *Resolver) localCopyReady(storageRef string, expectedSize int64) bool {
	path, err := r.Local.LocalPath(storageRef)
	if err != nil {
		return false
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	return info.Size() == expectedSize
}

func (r *Resolver) markRehydrated(ctx context.Context, item Object) error {
	now := r.Now
	if now == nil {
		now = time.Now
	}
	return r.Repository.MarkRehydrated(ctx, item.ID, now().Add(time.Duration(NormalizeRetentionDays(item.RetentionDays))*24*time.Hour))
}
