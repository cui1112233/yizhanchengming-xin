package batchfactoryv11

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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
	// 关键守卫：分镜输入（原文 digest + 冻结配置 + 模式 + Hook）完全没变时，
	// 不能产生新修订，更不能把当前 active 视频槽位全部标成 orphaned。
	// 旧逻辑每次修订都无条件轮换槽位，自动化重试「先 director 再 video」会把
	// 已经生成成功、仍然可播的视频整批废掉。
	if reused, ok, err := s.reusableDirectorRevision(ctx, tx, owner, book, snapshot, digest, hookID, string(snapshotJSON)); err != nil {
		return DirectorRevision{}, err
	} else if ok {
		if err := tx.Commit(); err != nil {
			return DirectorRevision{}, err
		}
		return reused, nil
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
	// 逐槽位保留：上一修订的每个槽位都记录了自己的输入 digest，本次只有
	// digest 真正变化的槽位才轮换；未变化槽位连同已成功的生产任务一起保留，
	// 避免一次局部修订把整本书已生成好的视频全部废掉。
	previousDigests, err := loadPreviousSlotDigests(ctx, tx, owner, book, currentRevision)
	if err != nil {
		return DirectorRevision{}, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT v.id,v.ordinal,COALESCE(r.label,''),COALESCE(r.video_prompt,''),COALESCE(r.visual_prompt,''),COALESCE(r.duration_seconds,0),COALESCE(v.revision,1) FROM batch_factory_v12_videos v LEFT JOIN batch_factory_v12_video_records r ON r.video_id=v.id WHERE v.batch_id=? AND v.book_id=? AND v.owner_username=? AND v.compatibility_state='active' ORDER BY v.ordinal,v.id`, book.BatchID, book.ID, owner)
	if err != nil {
		return DirectorRevision{}, err
	}
	type activeSlot struct {
		video   Video
		ordinal int
	}
	oldSlots := []activeSlot{}
	oldByOrdinal := map[int]activeSlot{}
	for rows.Next() {
		var slot activeSlot
		if err := rows.Scan(&slot.video.ID, &slot.ordinal, &slot.video.Label, &slot.video.VideoPrompt, &slot.video.VisualPrompt, &slot.video.DurationSeconds, &slot.video.Revision); err != nil {
			rows.Close()
			return DirectorRevision{}, err
		}
		slot.video.BatchID, slot.video.BookID, slot.video.CompatibilityState = book.BatchID, book.ID, "active"
		slot.video.Ordinal = slot.ordinal
		oldSlots = append(oldSlots, slot)
		oldByOrdinal[slot.ordinal] = slot
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return DirectorRevision{}, err
	}
	rows.Close()
	newDigests := make([]string, len(output.Storyboard))
	kept := map[int]string{}
	for ordinal, draft := range output.Storyboard {
		newDigests[ordinal] = directorSlotDigest(snapshot.Mode, hookID, string(snapshotJSON), draft)
		previous, exists := oldByOrdinal[ordinal]
		if exists && newDigests[ordinal] != "" && previousDigests[previous.video.ID] == newDigests[ordinal] {
			kept[ordinal] = previous.video.ID
		}
	}
	previousSelections := map[int]VideoAssetSelection{}
	previousSelectionExists := map[int]bool{}
	for _, slot := range oldSlots {
		if _, isKept := kept[slot.ordinal]; isKept {
			// 保留槽位的设置补丁原样留在原 video 上，不做 orphan 审计，
			// 也不需要为素材 reconcile 读取它。
			continue
		}
		videoID := slot.video.ID
		patch, err := loadPatch(ctx, tx, owner, ScopeRef{Kind: ScopeVideo, BatchID: book.BatchID, BookID: book.ID, VideoID: videoID})
		if err != nil {
			return DirectorRevision{}, err
		}
		if selection, ok := decodeVideoAssetSelection(patch); ok {
			previousSelections[slot.ordinal], previousSelectionExists[slot.ordinal] = selection, true
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
	for _, slot := range oldSlots {
		if _, isKept := kept[slot.ordinal]; isKept {
			continue
		}
		if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_videos SET compatibility_state='orphaned',updated_at=? WHERE id=? AND owner_username=? AND compatibility_state='active'`, now, slot.video.ID, owner); err != nil {
			return DirectorRevision{}, err
		}
	}

	videos := make([]Video, 0, len(output.Storyboard))
	comicMode := normalizeProductionMode(snapshot.Effective) == "comic"
	for ordinal, draft := range output.Storyboard {
		if keptID, isKept := kept[ordinal]; isKept {
			if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_director_video_links(director_revision_id,video_id,ordinal,slot_digest) VALUES(?,?,?,?)`, revisionID, keptID, ordinal, newDigests[ordinal]); err != nil {
				return DirectorRevision{}, err
			}
			videos = append(videos, oldByOrdinal[ordinal].video)
			continue
		}
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
		if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_director_video_links(director_revision_id,video_id,ordinal,slot_digest) VALUES(?,?,?,?)`, revisionID, videoID, ordinal, newDigests[ordinal]); err != nil {
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
		if _, isKept := kept[ordinal]; isKept {
			// 保留槽位的 settings_patches 仍挂在原 video 上，重复插入会产生
			// 同 scope 的重复补丁行。
			continue
		}
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

// loadPreviousSlotDigests returns the per-slot input digests recorded on the
// links of the revision that preceded currentRevision. Rows written before the
// slot_digest column existed read back as "" and therefore never match, which
// keeps the legacy whole-book rotation as the safe fallback.
func loadPreviousSlotDigests(ctx context.Context, q batchQueryer, owner string, book Book, previousRevision int64) (map[string]string, error) {
	digests := map[string]string{}
	if previousRevision < 1 {
		return digests, nil
	}
	rows, err := q.QueryContext(ctx, `SELECT l.video_id,COALESCE(l.slot_digest,'') FROM batch_factory_v12_director_video_links l JOIN batch_factory_v12_director_revisions d ON d.id=l.director_revision_id WHERE d.batch_id=? AND d.book_id=? AND d.owner_username=? AND d.revision=?`, book.BatchID, book.ID, owner, previousRevision)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var videoID, digest string
		if err := rows.Scan(&videoID, &digest); err != nil {
			return nil, err
		}
		digests[videoID] = digest
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return digests, nil
}

// directorSlotDigest 哈希一个分镜槽位「会影响产出」的全部冻结输入：模式、
// Hook、书本级冻结配置以及该槽位的分镜内容。ID/SceneID 是模型每次生成时
// 随手分配的标识，同一输入重跑也会变，参与 digest 会让保留永远不命中。
func directorSlotDigest(mode, hookID, snapshotJSON string, draft DirectorVideo) string {
	type digestDraft struct {
		DurationSec  int            `json:"duration_sec"`
		Characters   []string       `json:"characters"`
		Props        []string       `json:"props"`
		Scene        string         `json:"scene"`
		SceneMemory  string         `json:"scene_memory,omitempty"`
		PrefixKey    string         `json:"prefix_key"`
		Shots        []DirectorShot `json:"shots"`
		VideoDesc    string         `json:"video_desc"`
		VisualPrompt string         `json:"visual_prompt,omitempty"`
		FinalPrompt  string         `json:"final_prompt,omitempty"`
	}
	payload, err := json.Marshal(struct {
		Mode      string          `json:"mode"`
		Hook      string          `json:"hook,omitempty"`
		Effective json.RawMessage `json:"effective"`
		Draft     digestDraft     `json:"draft"`
	}{
		Mode:      mode,
		Hook:      hookID,
		Effective: json.RawMessage(snapshotJSON),
		Draft: digestDraft{
			DurationSec:  draft.DurationSec,
			Characters:   draft.Characters,
			Props:        draft.Props,
			Scene:        draft.Scene,
			SceneMemory:  draft.SceneMemory,
			PrefixKey:    draft.PrefixKey,
			Shots:        draft.Shots,
			VideoDesc:    draft.VideoDesc,
			VisualPrompt: draft.VisualPrompt,
			FinalPrompt:  draft.FinalPrompt,
		},
	})
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
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

// reusableDirectorRevision returns the latest revision when every director
// input is unchanged. The check compares digest, mode, Hook and the frozen
// effective settings; digest alone only covers the source text, so a settings
// change without a text change must still create a new revision.
func (s *MySQLStore) reusableDirectorRevision(ctx context.Context, q batchQueryer, owner string, book Book, snapshot DirectorSnapshot, digest, hookID, snapshotJSON string) (DirectorRevision, bool, error) {
	latest, err := loadLatestDirectorRevision(ctx, q, owner, book.BatchID, book.ID)
	if errors.Is(err, ErrNotFound) {
		return DirectorRevision{}, false, nil
	}
	if err != nil {
		return DirectorRevision{}, false, err
	}
	if latest.SourceDigest != digest || latest.Mode != snapshot.Mode || latest.HookRevisionID != hookID {
		return DirectorRevision{}, false, nil
	}
	var storedEffective string
	err = q.QueryRowContext(ctx, `SELECT effective_json FROM batch_factory_v12_config_snapshots WHERE id=? AND owner_username=? AND batch_id=? AND book_id=?`, latest.SnapshotID, owner, book.BatchID, book.ID).Scan(&storedEffective)
	if errors.Is(err, sql.ErrNoRows) {
		return DirectorRevision{}, false, nil
	}
	if err != nil {
		return DirectorRevision{}, false, err
	}
	if !canonicalJSONEqual(storedEffective, snapshotJSON) {
		return DirectorRevision{}, false, nil
	}
	videos, err := loadActiveDirectorVideos(ctx, q, owner, book)
	if err != nil {
		return DirectorRevision{}, false, err
	}
	latest.Videos = videos
	latest.BatchID, latest.BookID = book.BatchID, book.ID
	return latest, true, nil
}

func loadActiveDirectorVideos(ctx context.Context, q batchQueryer, owner string, book Book) ([]Video, error) {
	rows, err := q.QueryContext(ctx, `SELECT v.id,COALESCE(r.label,''),COALESCE(r.video_prompt,''),COALESCE(r.visual_prompt,''),COALESCE(r.duration_seconds,0),v.ordinal,COALESCE(v.revision,1) FROM batch_factory_v12_videos v LEFT JOIN batch_factory_v12_video_records r ON r.video_id=v.id WHERE v.batch_id=? AND v.book_id=? AND v.owner_username=? AND v.compatibility_state='active' ORDER BY v.ordinal,v.id`, book.BatchID, book.ID, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	videos := []Video{}
	for rows.Next() {
		var video Video
		if err := rows.Scan(&video.ID, &video.Label, &video.VideoPrompt, &video.VisualPrompt, &video.DurationSeconds, &video.Ordinal, &video.Revision); err != nil {
			return nil, err
		}
		video.BatchID, video.BookID, video.CompatibilityState = book.BatchID, book.ID, "active"
		videos = append(videos, video)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return videos, nil
}

// canonicalJSONEqual compares two JSON documents after normalizing key order
// and number formatting, so a stored snapshot can be compared with a freshly
// marshalled one.
func canonicalJSONEqual(left, right string) bool {
	var leftValue, rightValue any
	if json.Unmarshal([]byte(left), &leftValue) != nil || json.Unmarshal([]byte(right), &rightValue) != nil {
		return false
	}
	leftBytes, err := json.Marshal(leftValue)
	if err != nil {
		return false
	}
	rightBytes, err := json.Marshal(rightValue)
	if err != nil {
		return false
	}
	return string(leftBytes) == string(rightBytes)
}
