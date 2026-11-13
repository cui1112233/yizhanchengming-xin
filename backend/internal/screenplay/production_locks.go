package screenplay

import (
	"context"
	"fmt"
	"time"
)

type productionLockRenewer interface {
	RenewProductionLock(context.Context, string, string, time.Duration) error
}

func (s *MySQLStore) RenewProductionLock(ctx context.Context, name, token string, ttl time.Duration) error {
	now := time.Now().UTC()
	res, e := s.db.ExecContext(ctx, `UPDATE screenplay_locks SET expires_at=? WHERE lock_name=? AND owner_token=? AND expires_at>?`, now.Add(ttl), name, token, now)
	if e != nil {
		return e
	}
	n, e := res.RowsAffected()
	if e == nil && n == 0 {
		return fmt.Errorf("%w: 执行锁已失效", ErrLockBusy)
	}
	return e
}
func (m *MemoryStore) RenewProductionLock(_ context.Context, name, token string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	lock, ok := m.locks[name]
	now := time.Now().UTC()
	if !ok || lock.token != token || !lock.expires.After(now) {
		return ErrLockBusy
	}
	lock.expires = now.Add(ttl)
	m.locks[name] = lock
	return nil
}
