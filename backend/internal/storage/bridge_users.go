package storage

import (
	"context"
	"database/sql"
	"sync"
	"time"
)

type BridgeUsers struct {
	DB *sql.DB

	// cache tracks recently resolved usernames to avoid writing to MySQL on
	// every single API request. The write is idempotent (ON DUPLICATE KEY
	// UPDATE), so skipping it for a short window is safe.
	cache    sync.Map
	cacheTTL time.Duration
}

func (s BridgeUsers) ResolveBridgeUser(ctx context.Context, username string, isOwner bool) error {
	ttl := s.cacheTTL
	if ttl == 0 {
		ttl = 5 * time.Minute
	}
	if cached, ok := s.cache.Load(username); ok {
		if entry, valid := cached.(bridgeUserCacheEntry); valid && time.Since(entry.at) < ttl {
			return nil
		}
	}
	_, err := s.DB.ExecContext(ctx, `INSERT INTO batch_factory_v12_bridge_users(username,is_owner,created_at,updated_at) VALUES(?,?,CURRENT_TIMESTAMP(6),CURRENT_TIMESTAMP(6)) ON DUPLICATE KEY UPDATE is_owner=VALUES(is_owner), updated_at=CURRENT_TIMESTAMP(6)`, username, isOwner)
	if err == nil {
		s.cache.Store(username, bridgeUserCacheEntry{at: time.Now()})
	}
	return err
}

type bridgeUserCacheEntry struct{ at time.Time }
