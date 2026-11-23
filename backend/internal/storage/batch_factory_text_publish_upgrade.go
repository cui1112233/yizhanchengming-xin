package storage

import (
	"cmp"
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

type textReceiptBook struct{ owner, batch, book string }

// Both data migrations read the same private, owner-scoped receipt formats.
func readTextReceiptDrafts(ctx context.Context, tx *sql.Tx, visit func(textReceiptBook, string, cp.Record)) error {
	const currentKind = "n8-text-publish"
	rows, err := tx.QueryContext(ctx, `SELECT owner_username,scope,kind,draft_key,content FROM batch_factory_v12_drafts WHERE kind IN ('n8-text-publish','n8-text-publish-history') AND content IS NOT NULL`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var owner, batch, kind, rowKey string
		var content []byte
		if err := rows.Scan(&owner, &batch, &kind, &rowKey, &content); err != nil {
			return err
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
		archiveKey := ""
		if kind != currentKind {
			archiveKey = rowKey
		}
		visit(textReceiptBook{owner, batch, rec.BookID}, archiveKey, rec)
	}
	return rows.Err()
}

// Older manual uploads wrote receipts before all entry points used the global
// confirmation registry. Reconstruct their scopes without changing any receipt.
func adoptTextPublishReceipts(ctx context.Context, tx *sql.Tx) (bool, error) {
	type success struct {
		id, archive string
		at          time.Time
	}
	last := make(map[textReceiptBook]success)
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
	err := readTextReceiptDrafts(ctx, tx, func(book textReceiptBook, archiveKey string, rec cp.Record) {
		req := cp.PublishRequest{Owner: book.owner, BatchID: book.batch, BookID: rec.BookID, CollectionID: rec.CollectionID}
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
			return
		}
		old := last[book]
		// For equal times, choose deterministically. Prefer the current draft
		// when it contains the same physical success as an archived receipt.
		tie := h.CollectionID > old.id || h.CollectionID == old.id && archiveKey < old.archive
		if h.PublishedAt.After(old.at) || h.PublishedAt.Equal(old.at) && tie {
			last[book] = success{id: h.CollectionID, archive: archiveKey, at: h.PublishedAt}
		}
	})
	if err != nil {
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

// 1200114 could copy an older collection's last success while recording its
// next pending attempt. Compare physical publication times across the book.
func correctTextPublishHistory(ctx context.Context, tx *sql.Tx) (bool, error) {
	type receiptRef struct {
		id, archive string
		at          time.Time
	}
	// Receipt writes lock the same book row. Take those locks before reading
	// the receipts so a concurrent publication cannot be replaced by an old snapshot.
	rows, err := tx.QueryContext(ctx, `SELECT b.owner_username,b.batch_id,b.id,COALESCE(JSON_UNQUOTE(JSON_EXTRACT(r.source_metadata_json,'$.textLastPublishedCollectionId')),''),COALESCE(JSON_UNQUOTE(JSON_EXTRACT(r.source_metadata_json,'$.textLastPublishedArchiveKey')),'') FROM batch_factory_v12_books b JOIN batch_factory_v12_book_records r ON r.book_id=b.id WHERE COALESCE(JSON_UNQUOTE(JSON_EXTRACT(r.source_metadata_json,'$.textCollectionId')),'')<>'' ORDER BY b.id FOR UPDATE`)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	existingPointers := map[textReceiptBook]receiptRef{}
	for rows.Next() {
		var book textReceiptBook
		var ref receiptRef
		if err := rows.Scan(&book.owner, &book.batch, &book.book, &ref.id, &ref.archive); err != nil {
			return false, err
		}
		existingPointers[book] = ref
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	if err := rows.Close(); err != nil {
		return false, err
	}
	if len(existingPointers) == 0 {
		return true, nil
	}

	candidates := map[textReceiptBook][]receiptRef{}
	err = readTextReceiptDrafts(ctx, tx, func(book textReceiptBook, archive string, rec cp.Record) {
		h := rec.HistoricalPublished()
		if h != nil && h.Kind == cp.CollectionKindText && h.BookID != "" && h.BookID == book.book && h.CollectionID != "" && h.CollectionID == rec.CollectionID && !h.PublishedAt.IsZero() {
			existing := candidates[book]
			ref := receiptRef{h.CollectionID, archive, h.PublishedAt}
			if len(existing) == 0 || ref.at.After(existing[0].at) {
				candidates[book] = []receiptRef{ref}
			} else if ref.at.Equal(existing[0].at) {
				candidates[book] = append(existing, ref)
			}
		}
	})
	if err != nil {
		return false, err
	}
	for book, newest := range candidates {
		prior, ok := existingPointers[book]
		if !ok {
			continue
		}
		id, archive := prior.id, prior.archive
		// When different publications share a timestamp, keep an already valid
		// pointer. Otherwise prefer its collection, then a deterministic copy.
		slices.SortFunc(newest, func(a, b receiptRef) int {
			if v := cmp.Compare(textHistoryPreference(a.id == id, a.archive == archive), textHistoryPreference(b.id == id, b.archive == archive)); v != 0 {
				return v
			}
			if v := cmp.Compare(a.id, b.id); v != 0 {
				return v
			}
			return cmp.Compare(a.archive, b.archive)
		})
		best := newest[0]
		if best.id == id && best.archive == archive {
			continue
		}
		if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_book_records SET source_metadata_json=JSON_SET(source_metadata_json,'$.textLastPublishedCollectionId',?,'$.textLastPublishedArchiveKey',?) WHERE book_id=?`, best.id, best.archive, book.book); err != nil {
			return false, err
		}
	}
	return true, nil
}

func textHistoryPreference(sameCollection, sameArchive bool) int {
	if sameCollection && sameArchive {
		return 0
	}
	if sameCollection {
		return 1
	}
	return 2
}
