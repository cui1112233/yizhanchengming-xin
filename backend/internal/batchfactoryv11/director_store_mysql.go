package batchfactoryv11

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

func (s *MySQLStore) CreateHookRevision(ctx context.Context, owner, batchID, bookID, text, digest string) (HookRevision, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return HookRevision{}, err
	}
	defer tx.Rollback()
	var one int
	if err := tx.QueryRowContext(ctx, `SELECT 1 FROM batch_factory_v12_books WHERE id=? AND batch_id=? AND owner_username=? FOR UPDATE`, bookID, batchID, owner).Scan(&one); errors.Is(err, sql.ErrNoRows) {
		return HookRevision{}, ErrNotFound
	} else if err != nil {
		return HookRevision{}, err
	}
	var current int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(revision),0) FROM batch_factory_v12_hook_revisions WHERE batch_id=? AND book_id=? AND owner_username=?`, batchID, bookID, owner).Scan(&current); err != nil {
		return HookRevision{}, err
	}
	id, err := newID("hook")
	if err != nil {
		return HookRevision{}, err
	}
	now := time.Now().UTC()
	value := HookRevision{ID: id, BatchID: batchID, BookID: bookID, Revision: current + 1, Status: "draft", Text: text, SourceDigest: digest, CreatedAt: now}
	if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_hook_revisions(id,batch_id,book_id,owner_username,revision,status,hook_text,source_digest,created_at) VALUES(?,?,?,?,?,'draft',?,?,?)`, id, batchID, bookID, owner, value.Revision, text, digest, now); err != nil {
		return HookRevision{}, err
	}
	if err := tx.Commit(); err != nil {
		return HookRevision{}, err
	}
	return value, nil
}

func (s *MySQLStore) ApproveHookRevision(ctx context.Context, owner, batchID, bookID, hookID string) (HookRevision, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return HookRevision{}, err
	}
	defer tx.Rollback()
	value, err := loadHookRevision(ctx, tx, owner, batchID, bookID, hookID, true)
	if err != nil {
		return HookRevision{}, err
	}
	if value.Status == "approved" {
		return value, nil
	}
	now := time.Now().UTC()
	res, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_hook_revisions SET status='approved',approved_at=? WHERE id=? AND batch_id=? AND book_id=? AND owner_username=?`, now, hookID, batchID, bookID, owner)
	if err != nil {
		return HookRevision{}, err
	}
	if count, _ := res.RowsAffected(); count != 1 {
		return HookRevision{}, ErrNotFound
	}
	if err := tx.Commit(); err != nil {
		return HookRevision{}, err
	}
	value.Status, value.ApprovedAt = "approved", &now
	return value, nil
}

func (s *MySQLStore) LatestHookRevision(ctx context.Context, owner, batchID, bookID string) (HookRevision, error) {
	return loadHookRevision(ctx, s.db, owner, batchID, bookID, "", false)
}

type hookQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func loadHookRevision(ctx context.Context, q hookQueryer, owner, batchID, bookID, hookID string, forUpdate bool) (HookRevision, error) {
	query := `SELECT id,revision,status,hook_text,source_digest,created_at,approved_at FROM batch_factory_v12_hook_revisions WHERE batch_id=? AND book_id=? AND owner_username=?`
	args := []any{batchID, bookID, owner}
	if hookID != "" {
		query += ` AND id=?`
		args = append(args, hookID)
	} else {
		query += ` ORDER BY revision DESC LIMIT 1`
	}
	if forUpdate {
		query += ` FOR UPDATE`
	}
	var value HookRevision
	var approved sql.NullTime
	err := q.QueryRowContext(ctx, query, args...).Scan(&value.ID, &value.Revision, &value.Status, &value.Text, &value.SourceDigest, &value.CreatedAt, &approved)
	if errors.Is(err, sql.ErrNoRows) {
		return HookRevision{}, ErrNotFound
	}
	if err != nil {
		return HookRevision{}, err
	}
	value.BatchID, value.BookID = batchID, bookID
	if approved.Valid {
		value.ApprovedAt = &approved.Time
	}
	return value, nil
}

func (s *MySQLStore) PersistDirectorRevision(ctx context.Context, owner string, book Book, snapshot DirectorSnapshot, digest, hookID string, output DirectorResult) (DirectorRevision, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return DirectorRevision{}, err
	}
	defer tx.Rollback()
	var currentBookRevision int64
	if err := tx.QueryRowContext(ctx, `SELECT revision FROM batch_factory_v12_books WHERE id=? AND batch_id=? AND owner_username=? FOR UPDATE`, book.ID, book.BatchID, owner).Scan(&currentBookRevision); errors.Is(err, sql.ErrNoRows) {
		return DirectorRevision{}, ErrNotFound
	} else if err != nil {
		return DirectorRevision{}, err
	}
	if snapshot.Mode == "viral" {
		hook, err := loadHookRevision(ctx, tx, owner, book.BatchID, book.ID, hookID, true)
		if err != nil || hook.Status != "approved" {
			return DirectorRevision{}, fmt.Errorf("%w: approved Hook required", ErrConflict)
		}
	}
	snapshotID, err := newID("snapshot")
	if err != nil {
		return DirectorRevision{}, err
	}
	snapshotJSON, err := json.Marshal(snapshot.Effective)
	if err != nil {
		return DirectorRevision{}, err
	}
	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_config_snapshots(id,owner_username,batch_id,book_id,video_id,effective_json,created_at) VALUES(?,?,?,?,NULL,?,?)`, snapshotID, owner, book.BatchID, book.ID, string(snapshotJSON), now); err != nil {
		return DirectorRevision{}, err
	}
	var currentRevision int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(revision),0) FROM batch_factory_v12_director_revisions WHERE batch_id=? AND book_id=? AND owner_username=?`, book.BatchID, book.ID, owner).Scan(&currentRevision); err != nil {
		return DirectorRevision{}, err
	}
	revisionID, err := newID("director")
	if err != nil {
		return DirectorRevision{}, err
	}
	outputJSON, err := json.Marshal(output)
	if err != nil {
		return DirectorRevision{}, err
	}
	var hookValue any
	if hookID != "" {
		hookValue = hookID
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_director_revisions(id,batch_id,book_id,owner_username,revision,snapshot_id,mode,source_digest,hook_revision_id,output_json,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, revisionID, book.BatchID, book.ID, owner, currentRevision+1, snapshotID, snapshot.Mode, digest, hookValue, string(outputJSON), now); err != nil {
		return DirectorRevision{}, err
	}

	orphaned := []OrphanedOverride{}
	rows, err := tx.QueryContext(ctx, `SELECT id,ordinal FROM batch_factory_v12_videos WHERE batch_id=? AND book_id=? AND owner_username=? AND compatibility_state='active' ORDER BY ordinal,id`, book.BatchID, book.ID, owner)
	if err != nil {
		return DirectorRevision{}, err
	}
	oldIDs := []string{}
	oldOrdinals := map[string]int{}
	for rows.Next() {
		var id string
		var ordinal int
		if err := rows.Scan(&id, &ordinal); err != nil {
			rows.Close()
			return DirectorRevision{}, err
		}
		oldIDs = append(oldIDs, id)
		oldOrdinals[id] = ordinal
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return DirectorRevision{}, err
	}
	rows.Close()
	previousSelections := map[int]VideoAssetSelection{}
	previousSelectionExists := map[int]bool{}
	for _, videoID := range oldIDs {
		patch, err := loadPatch(ctx, tx, owner, ScopeRef{Kind: ScopeVideo, BatchID: book.BatchID, BookID: book.ID, VideoID: videoID})
		if err != nil {
			return DirectorRevision{}, err
		}
		if selection, ok := decodeVideoAssetSelection(patch); ok {
			previousSelections[oldOrdinals[videoID]], previousSelectionExists[oldOrdinals[videoID]] = selection, true
		}
		if len(patch) > 0 {
			encoded, err := json.Marshal(patch)
			if err != nil {
				return DirectorRevision{}, err
			}
			auditID, err := newID("orphan")
			if err != nil {
				return DirectorRevision{}, err
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_orphaned_overrides(id,director_revision_id,video_id,patch_json,state,created_at) VALUES(?,?,?,?, 'orphaned',?)`, auditID, revisionID, videoID, string(encoded), now); err != nil {
				return DirectorRevision{}, err
			}
			orphaned = append(orphaned, OrphanedOverride{VideoID: videoID, Patch: patch, State: "orphaned"})
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_videos SET compatibility_state='orphaned',updated_at=? WHERE batch_id=? AND book_id=? AND owner_username=? AND compatibility_state='active'`, now, book.BatchID, book.ID, owner); err != nil {
		return DirectorRevision{}, err
	}

	videos := make([]Video, 0, len(output.Storyboard))
	comicMode := normalizeProductionMode(snapshot.Effective) == "comic"
	for ordinal, draft := range output.Storyboard {
		videoID, err := newID("video")
		if err != nil {
			return DirectorRevision{}, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_videos(id,batch_id,book_id,owner_username,ordinal,compatibility_state,revision,created_at,updated_at) VALUES(?,?,?,?,?,'active',1,?,?)`, videoID, book.BatchID, book.ID, owner, ordinal, now, now); err != nil {
			return DirectorRevision{}, err
		}
		label := fmt.Sprintf("VIDEO %02d", ordinal+1)
		if comicMode {
			label = fmt.Sprintf("漫画页 %02d", ordinal+1)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_video_records(video_id,label,video_prompt,visual_prompt,duration_seconds) VALUES(?,?,?,?,?)`, videoID, label, storyboardVideoPrompt(draft), nullableString(draft.VisualPrompt), draft.DurationSec); err != nil {
			return DirectorRevision{}, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_director_video_links(director_revision_id,video_id,ordinal) VALUES(?,?,?)`, revisionID, videoID, ordinal); err != nil {
			return DirectorRevision{}, err
		}
		videos = append(videos, Video{ID: videoID, BatchID: book.BatchID, BookID: book.ID, Label: label, VideoPrompt: storyboardVideoPrompt(draft), VisualPrompt: draft.VisualPrompt, DurationSeconds: float64(draft.DurationSec), CompatibilityState: "active", Revision: 1})
	}
	if err := persistDirectorBookAssets(ctx, tx, owner, book, snapshot, output, now); err != nil {
		return DirectorRevision{}, err
	}
	assetRows, err := tx.QueryContext(ctx, `SELECT id,kind,name FROM batch_factory_v12_book_assets WHERE batch_id=? AND book_id=? AND owner_username=?`, book.BatchID, book.ID, owner)
	if err != nil {
		return DirectorRevision{}, err
	}
	assetIDs := map[string]string{}
	nameRowsByKind := map[string][][2]string{}
	for assetRows.Next() {
		var id, kind, name string
		if err := assetRows.Scan(&id, &kind, &name); err != nil {
			assetRows.Close()
			return DirectorRevision{}, err
		}
		assetIDs[kind+"\x00"+name] = id
		nameRowsByKind[kind] = append(nameRowsByKind[kind], [2]string{name, id})
	}
	if err := assetRows.Err(); err != nil {
		assetRows.Close()
		return DirectorRevision{}, err
	}
	assetRows.Close()
	// Rename-tolerant binding: shot text may carry a parenthesised variant.
	for kind, rows := range nameRowsByKind {
		addCoreNameLookups(assetIDs, kind, rows)
	}
	for ordinal, video := range videos {
		selection := reconcileVideoAssetSelection(previousSelections[ordinal], previousSelectionExists[ordinal], automaticAssetIDsForDirectorVideo(output.Storyboard[ordinal], assetIDs))
		if selection.effectivelyEmpty() {
			// 空选择不落库：否则会被当成“用户一个都没选”，把默认资产全过滤掉。
			continue
		}
		encoded, err := json.Marshal(selection)
		if err != nil {
			return DirectorRevision{}, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_settings_patches(scope_type,scope_id,owner_username,batch_id,book_id,video_id,patch_json,revision,updated_at) VALUES('video',?,?,?,?,?,?,1,?)`, video.ID, owner, book.BatchID, book.ID, video.ID, string(encoded), now); err != nil {
			return DirectorRevision{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_books SET revision=?,updated_at=? WHERE id=? AND batch_id=? AND owner_username=?`, currentBookRevision+1, now, book.ID, book.BatchID, owner); err != nil {
		return DirectorRevision{}, err
	}
	if err := tx.Commit(); err != nil {
		return DirectorRevision{}, err
	}
	return DirectorRevision{ID: revisionID, BatchID: book.BatchID, BookID: book.ID, Revision: currentRevision + 1, Mode: snapshot.Mode, SnapshotID: snapshotID, SourceDigest: digest, HookRevisionID: hookID, Output: output, Videos: videos, OrphanedOverrides: orphaned, CreatedAt: now}, nil
}

func loadLatestDirectorRevision(ctx context.Context, q batchQueryer, owner, batchID, bookID string) (DirectorRevision, error) {
	var value DirectorRevision
	var hook sql.NullString
	var raw []byte
	err := q.QueryRowContext(ctx, `SELECT id,revision,snapshot_id,mode,source_digest,hook_revision_id,output_json,created_at FROM batch_factory_v12_director_revisions WHERE batch_id=? AND book_id=? AND owner_username=? ORDER BY revision DESC LIMIT 1`, batchID, bookID, owner).Scan(&value.ID, &value.Revision, &value.SnapshotID, &value.Mode, &value.SourceDigest, &hook, &raw, &value.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return DirectorRevision{}, ErrNotFound
	}
	if err != nil {
		return DirectorRevision{}, err
	}
	value.BatchID, value.BookID = batchID, bookID
	if hook.Valid {
		value.HookRevisionID = hook.String
	}
	if err := json.Unmarshal(raw, &value.Output); err != nil {
		return DirectorRevision{}, err
	}
	return value, nil
}
