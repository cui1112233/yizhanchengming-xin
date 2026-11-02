package batchfactoryv11

import (
	"context"
	"database/sql"
	"errors"
)

type batchOwnershipChecker interface {
	EnsureBatchOwnership(context.Context, string, string) error
}

func ensureBatchOwnership(ctx context.Context, store Store, owner, batchID string) error {
	if checker, ok := store.(batchOwnershipChecker); ok {
		return checker.EnsureBatchOwnership(ctx, owner, batchID)
	}
	_, err := store.GetBatch(ctx, owner, batchID)
	return err
}

func (s *MySQLStore) EnsureBatchOwnership(ctx context.Context, owner, batchID string) error {
	var one int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM batch_factory_v12_batches WHERE id=? AND owner_username=?`, batchID, owner).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
