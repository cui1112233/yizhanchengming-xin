package batchfactoryv11

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"maps"
	"slices"

	cp "qiantie/backend/internal/collectionpublish"
)

func (s *MemoryStore) BindTextPublishKeys(_ context.Context, keys map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, scope := range keys {
		if previous, ok := s.textPublishKeys[key]; ok && previous != scope {
			return ErrConflict
		}
	}
	if s.textPublishKeys == nil {
		s.textPublishKeys = map[string]string{}
	}
	maps.Copy(s.textPublishKeys, keys)
	return nil
}

func (s *MySQLStore) BindTextPublishKeys(ctx context.Context, keys map[string]string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, key := range slices.Sorted(maps.Keys(keys)) {
		digest := sha256.Sum256([]byte(keys[key]))
		if _, err = tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_text_publish_keys(request_key,scope_digest) VALUES(?,?) ON DUPLICATE KEY UPDATE request_key=request_key`, key, digest[:]); err != nil {
			return err
		}
		var actual []byte
		if err = tx.QueryRowContext(ctx, `SELECT scope_digest FROM batch_factory_v12_text_publish_keys WHERE request_key=? FOR UPDATE`, key).Scan(&actual); err != nil {
			return err
		}
		if !bytes.Equal(actual, digest[:]) {
			return ErrConflict
		}
	}
	return tx.Commit()
}

// Use the physical receipt, never an older success rebound to a later run,
// to advance the history pointer. Old books can have only the current pointer.
func textReceiptPointers(meta map[string]any, rec cp.Record, load func(string) (cp.Record, error)) (map[string]any, bool, error) {
	current, _ := meta["textCollectionId"].(string)
	last, _ := meta["textLastPublishedCollectionId"].(string)
	old, err := load(rec.CollectionID)
	if err != nil {
		return nil, false, err
	}
	nextCurrent, nextLast := current, last
	if last == "" && current != "" {
		previous := old
		if current != rec.CollectionID {
			previous, err = load(current)
			if err != nil {
				return nil, false, err
			}
		}
		if previous.HistoricalPublished() != nil {
			nextLast = current
		}
	}
	if current == "" || current == rec.CollectionID || old.CollectionID == "" || old.IntentID != rec.IntentID || old.RunID != rec.RunID || old.RequestKey != rec.RequestKey {
		nextCurrent = rec.CollectionID
	}
	if rec.Confirmed() && (!rec.Reused && (!old.Confirmed() || old.IntentID != rec.IntentID) || nextLast == "") {
		nextLast = rec.CollectionID
	}
	if current == nextCurrent && last == nextLast {
		return meta, false, nil
	}
	out := maps.Clone(meta)
	if out == nil {
		out = map[string]any{}
	}
	out["textCollectionId"] = nextCurrent
	if nextLast != "" {
		out["textLastPublishedCollectionId"] = nextLast
	}
	return out, true, nil
}

func parseTextReceipt(content, bookID, id string) (cp.Record, error) {
	if content == "" {
		return cp.Record{}, nil
	}
	var rec cp.Record
	if err := json.Unmarshal([]byte(content), &rec); err != nil {
		return cp.Record{}, err
	}
	if rec.Kind != cp.CollectionKindText || rec.BookID != bookID || rec.CollectionID != id {
		return cp.Record{}, ErrInvalid
	}
	return rec, nil
}

func (s *MemoryStore) SaveTextPublishReceipt(_ context.Context, owner string, rec cp.Record, d Draft, archive *Draft) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.batches[d.Scope]
	if !ok || b.Owner != owner {
		return ErrNotFound
	}
	for i := range b.Value.Books {
		book := &b.Value.Books[i]
		if book.ID != rec.BookID {
			continue
		}
		meta, changed, err := textReceiptPointers(book.SourceMetadata, rec, func(id string) (cp.Record, error) {
			return parseTextReceipt(s.drafts[draftKey(owner, id, d.Kind, d.Scope)].Content, book.ID, id)
		})
		if err != nil {
			return err
		}
		if archive != nil {
			s.saveDraftLocked(owner, *archive)
		}
		s.saveDraftLocked(owner, d)
		if changed {
			book.SourceMetadata = meta
			book.Revision++
			s.batches[d.Scope] = b
		}
		return nil
	}
	return ErrNotFound
}

func (s *MySQLStore) SaveTextPublishReceipt(ctx context.Context, owner string, rec cp.Record, d Draft, archive *Draft) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var raw []byte
	err = tx.QueryRowContext(ctx, `SELECT r.source_metadata_json FROM batch_factory_v12_books b JOIN batch_factory_v12_book_records r ON r.book_id=b.id WHERE b.id=? AND b.batch_id=? AND b.owner_username=? FOR UPDATE`, rec.BookID, d.Scope, owner).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	meta, changed, err := textReceiptPointers(decodeSourceMetadata(raw), rec, func(id string) (cp.Record, error) {
		var content string
		err := tx.QueryRowContext(ctx, `SELECT content FROM batch_factory_v12_drafts WHERE owner_username=? AND draft_key=? AND kind=? AND scope=? FOR UPDATE`, owner, id, d.Kind, d.Scope).Scan(&content)
		if errors.Is(err, sql.ErrNoRows) {
			return cp.Record{}, nil
		}
		if err != nil {
			return cp.Record{}, err
		}
		return parseTextReceipt(content, rec.BookID, id)
	})
	if err != nil {
		return err
	}
	if archive != nil {
		if _, err = s.SaveDraftTx(ctx, tx, owner, *archive); err != nil {
			return err
		}
	}
	if _, err = s.SaveDraftTx(ctx, tx, owner, d); err != nil {
		return err
	}
	if changed {
		data, err := json.Marshal(meta)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE batch_factory_v12_book_records SET source_metadata_json=? WHERE book_id=?`, string(data), rec.BookID); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE batch_factory_v12_books SET revision=revision+1,updated_at=NOW(6) WHERE id=?`, rec.BookID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
