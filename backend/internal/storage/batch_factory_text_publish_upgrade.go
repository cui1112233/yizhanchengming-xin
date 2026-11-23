package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"maps"
	"slices"
	"time"

	cp "qiantie/backend/internal/collectionpublish"
)

// Older manual uploads wrote receipts before all entry points used the global
// confirmation registry. Reconstruct their scopes without changing any receipt.
func adoptTextPublishReceipts(ctx context.Context, tx *sql.Tx) (bool, error) {
	const currentKind = "n8-text-publish"
	rows, err := tx.QueryContext(ctx, `SELECT owner_username,scope,kind,draft_key,content FROM batch_factory_v12_drafts WHERE kind IN ('n8-text-publish','n8-text-publish-history') AND content IS NOT NULL`)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	type bookScope struct{ owner, batch, book string }
	type success struct {
		id, archive string
		at          time.Time
	}
	last := make(map[bookScope]success)
	claims := make(map[string][32]byte)
	var blocked [32]byte
	remember := func(key string, digest [32]byte) {
		if key == "" || len(key) > 128 {
			return
		}
		if prior, ok := claims[key]; ok && prior != digest {
			digest = blocked
		}
		claims[key] = digest
	}
	for rows.Next() {
		var owner, batch, kind, rowKey string
		var content []byte
		if err := rows.Scan(&owner, &batch, &kind, &rowKey, &content); err != nil {
			return false, err
		}
		var rec cp.Record
		if json.Unmarshal(content, &rec) != nil || rec.Kind != cp.CollectionKindText {
			continue
		}
		if kind == currentKind && rowKey != rec.CollectionID {
			continue
		}
		if kind != currentKind {
			sum := sha256.Sum256([]byte(rec.CollectionID + "\x00" + rec.RequestKey))
			if rowKey != hex.EncodeToString(sum[:]) {
				continue
			}
		}
		req := cp.PublishRequest{Owner: owner, BatchID: batch, BookID: rec.BookID, CollectionID: rec.CollectionID}
		digest := sha256.Sum256([]byte(req.TextKeyScope()))
		for r := &rec; r != nil; r = r.LastPublished {
			claim := digest
			if rec.BookID == "" || rec.CollectionID == "" || r.BookID != rec.BookID || r.CollectionID != rec.CollectionID || r.Kind != cp.CollectionKindText {
				claim = blocked
			}
			remember(r.RequestKey, claim)
		}
		h := rec.HistoricalPublished()
		if h == nil || h.BookID == "" || h.BookID != rec.BookID || h.CollectionID != rec.CollectionID || h.Kind != cp.CollectionKindText || h.PublishedAt.IsZero() {
			continue
		}
		book := bookScope{owner, batch, rec.BookID}
		old := last[book]
		archiveKey := ""
		if kind != currentKind {
			archiveKey = rowKey
		}
		// For equal times, choose deterministically. Prefer the current draft
		// when it contains the same physical success as an archived receipt.
		tie := h.CollectionID > old.id || h.CollectionID == old.id && archiveKey < old.archive
		if h.PublishedAt.After(old.at) || h.PublishedAt.Equal(old.at) && tie {
			last[book] = success{id: h.CollectionID, archive: archiveKey, at: h.PublishedAt}
		}
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	if err := rows.Close(); err != nil {
		return false, err
	}
	// A worker may have persisted its request before writing its first receipt.
	// The stage fingerprint occupies 64 bytes, followed by the collection ID.
	stages, err := tx.QueryContext(ctx, `SELECT owner_username,batch_id,book_id,COALESCE(SUBSTRING(input_revision,65),''),request_id FROM batch_factory_v12_book_stage_runs WHERE stage='text_publish' AND request_id IS NOT NULL AND request_id<>''`)
	if err != nil {
		return false, err
	}
	defer stages.Close()
	for stages.Next() {
		var req cp.PublishRequest
		if err := stages.Scan(&req.Owner, &req.BatchID, &req.BookID, &req.CollectionID, &req.RequestKey); err != nil {
			return false, err
		}
		digest := blocked
		if req.CollectionID != "" {
			digest = sha256.Sum256([]byte(req.TextKeyScope()))
		}
		remember(req.RequestKey, digest)
	}
	if err := stages.Err(); err != nil {
		return false, err
	}
	if err := stages.Close(); err != nil {
		return false, err
	}
	for _, key := range slices.Sorted(maps.Keys(claims)) {
		digest := claims[key]
		// A key found in multiple scopes, or in conflict with an existing
		// reservation, becomes unusable in every scope. A fresh key still works.
		if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_text_publish_keys(request_key,scope_digest) VALUES(?,?) ON DUPLICATE KEY UPDATE scope_digest=IF(scope_digest=?,scope_digest,?)`, key, digest[:], digest[:], blocked[:]); err != nil {
			return false, err
		}
	}
	for book, h := range last {
		_, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_book_records r JOIN batch_factory_v12_books b ON b.id=r.book_id SET r.source_metadata_json=JSON_SET(r.source_metadata_json,'$.textLastPublishedCollectionId',?,'$.textLastPublishedArchiveKey',?) WHERE b.id=? AND b.owner_username=? AND b.batch_id=? AND COALESCE(JSON_UNQUOTE(JSON_EXTRACT(r.source_metadata_json,'$.textCollectionId')),'')<>'' AND COALESCE(JSON_UNQUOTE(JSON_EXTRACT(r.source_metadata_json,'$.textLastPublishedCollectionId')),'')=''`, h.id, h.archive, book.book, book.owner, book.batch)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}
