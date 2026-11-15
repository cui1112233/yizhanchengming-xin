package screenplay

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

func (s *MySQLStore) ReserveAssetLibraryBulk(ctx context.Context, owner, key, digest string) error {
	if owner == "" || key == "" || len(key) > 191 || digest == "" {
		return ErrInvalidArgument
	}
	// Persist the intent before writing any external object. A failed upload or
	// rolled-back metadata transaction still binds this key to the same payload.
	_, err := s.db.ExecContext(ctx, `INSERT INTO screenplay_asset_library_bulk_operations(owner_username,request_key,digest,result_json) VALUES(?,?,?,NULL) ON DUPLICATE KEY UPDATE request_key=request_key`, owner, key, digest)
	if err != nil {
		return err
	}
	var stored string
	if err = s.db.QueryRowContext(ctx, `SELECT digest FROM screenplay_asset_library_bulk_operations WHERE owner_username=? AND request_key=?`, owner, key).Scan(&stored); err != nil {
		return err
	}
	if stored != digest {
		return ErrRevisionConflict
	}
	return nil
}

func (s *MySQLStore) LookupAssetLibraryBulk(ctx context.Context, owner, key, digest string) (AssetLibraryBulkResult, error) {
	var oldDigest string
	var raw []byte
	var result AssetLibraryBulkResult
	err := s.db.QueryRowContext(ctx, `SELECT digest,result_json FROM screenplay_asset_library_bulk_operations WHERE owner_username=? AND request_key=?`, owner, key).Scan(&oldDigest, &raw)
	if errors.Is(err, sql.ErrNoRows) {
		return result, ErrNotFound
	}
	if err != nil {
		return result, err
	}
	if oldDigest != digest {
		return result, ErrRevisionConflict
	}
	if len(raw) == 0 {
		return result, ErrNotFound
	}
	err = json.Unmarshal(raw, &result)
	for i := range result.Entries {
		result.Entries[i].Owner = owner
	}
	return result, err
}

func (s *MySQLStore) ApplyAssetLibraryBulk(ctx context.Context, owner, key, digest, action string, entries []AssetLibraryEntry, expected []AssetLibraryExpected) (AssetLibraryBulkResult, error) {
	empty := AssetLibraryBulkResult{}
	if owner == "" || key == "" || len(key) > 191 || digest == "" {
		return empty, ErrInvalidArgument
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return empty, err
	}
	defer tx.Rollback()
	// Serialize bulk requests for an owner before acquiring range locks. Two
	// simultaneous imports into an empty range must not hold compatible gap locks
	// and then deadlock while inserting each other's first character.
	_, err = tx.ExecContext(ctx, `INSERT INTO screenplay_asset_library_bulk_owners(owner_username) VALUES(?) ON DUPLICATE KEY UPDATE owner_username=owner_username`, owner)
	if err != nil {
		return empty, err
	}
	// Reserve the idempotency key first. Same-key concurrent requests serialize.
	_, err = tx.ExecContext(ctx, `INSERT INTO screenplay_asset_library_bulk_operations(owner_username,request_key,digest,result_json) VALUES(?,?,?,NULL) ON DUPLICATE KEY UPDATE request_key=request_key`, owner, key, digest)
	if err != nil {
		return empty, err
	}
	var oldDigest string
	var resultJSON []byte
	if err = tx.QueryRowContext(ctx, `SELECT digest,result_json FROM screenplay_asset_library_bulk_operations WHERE owner_username=? AND request_key=? FOR UPDATE`, owner, key).Scan(&oldDigest, &resultJSON); err != nil {
		return empty, err
	}
	if oldDigest != digest {
		return empty, ErrRevisionConflict
	}
	if len(resultJSON) > 0 {
		var result AssetLibraryBulkResult
		if err = json.Unmarshal(resultJSON, &result); err != nil {
			return empty, err
		}
		for i := range result.Entries {
			result.Entries[i].Owner = owner
		}
		if err = tx.Commit(); err != nil {
			return empty, err
		}
		return result, nil
	}
	// Use the owner/type index and next-key locks, including the empty range.
	// Ordinary inserts into this account's character range must wait as well.
	rows, err := tx.QueryContext(ctx, `SELECT `+assetLibraryColumns+` FROM screenplay_asset_library FORCE INDEX (idx_screenplay_library_owner_type) WHERE owner_username=? AND asset_type=? FOR UPDATE`, owner, string(AssetCharacter))
	if err != nil {
		return empty, err
	}
	current := []AssetLibraryEntry{}
	for rows.Next() {
		e, readErr := scanAssetLibrary(rows)
		if readErr != nil {
			rows.Close()
			return empty, readErr
		}
		if e.DeletedAt == nil {
			current = append(current, e)
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return empty, err
	}
	prepared, err := prepareLibraryBulk(owner, key, digest, action, entries, current, expected)
	if err != nil {
		return empty, err
	}
	now := time.Now().UTC()
	result := AssetLibraryBulkResult{Entries: []AssetLibraryEntry{}}
	if action == "clear" {
		_, err = tx.ExecContext(ctx, `UPDATE screenplay_asset_library SET deleted_at=?,updated_at=?,revision=revision+1 WHERE owner_username=? AND asset_type=? AND deleted_at IS NULL`, now, now, owner, string(AssetCharacter))
		if err != nil {
			return empty, err
		}
		result.ClearedCount = len(current)
	} else {
		for _, e := range prepared {
			e.CreatedAt = now
			e.UpdatedAt = now
			e.Revision = 1
			_, err = tx.ExecContext(ctx, `INSERT INTO screenplay_asset_library (`+assetLibraryColumns+`) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,1,?,?,NULL)`, e.ID, e.Owner, string(e.Type), e.Name, e.Description, e.Category, bulkJSON(e.Image), bulkJSON(e.Snapshot), e.SourceProjectID, e.SourceAssetID, e.SourceMediaID, bulkJSON(e.ImageObject), now, now)
			if err != nil {
				return empty, err
			}
			result.Entries = append(result.Entries, e)
		}
	}
	_, err = tx.ExecContext(ctx, `UPDATE screenplay_asset_library_bulk_operations SET result_json=? WHERE owner_username=? AND request_key=?`, bulkJSON(result), owner, key)
	if err != nil {
		return empty, err
	}
	if err = tx.Commit(); err != nil {
		return empty, err
	}
	return result, nil
}
