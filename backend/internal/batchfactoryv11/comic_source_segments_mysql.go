package batchfactoryv11

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (s *MySQLStore) ReplaceComicSourceSegments(ctx context.Context, owner, batchID, bookID, digest string, drafts []ComicSourceSegmentDraft) (ComicSourceSegmentation, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ComicSourceSegmentation{}, err
	}
	defer tx.Rollback()
	var source string
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(NULLIF(d.content,''),r.source_text,'') FROM batch_factory_v12_books b JOIN batch_factory_v12_book_records r ON r.book_id=b.id LEFT JOIN batch_factory_v12_drafts d ON d.owner_username=b.owner_username AND d.draft_key=CONCAT('working-front:',b.id) AND d.kind='working-front-content' AND d.scope=b.batch_id WHERE b.id=? AND b.batch_id=? AND b.owner_username=? FOR UPDATE`, bookID, batchID, owner).Scan(&source); errors.Is(err, sql.ErrNoRows) {
		return ComicSourceSegmentation{}, ErrNotFound
	} else if err != nil {
		return ComicSourceSegmentation{}, err
	}
	if digest != sourceDigest(source) {
		return ComicSourceSegmentation{}, fmt.Errorf("%w: 漫画正文已变化", ErrConflict)
	}
	if err := ValidateComicSourceSegments(source, drafts); err != nil {
		return ComicSourceSegmentation{}, err
	}
	var version int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(segmentation_version),0)+1 FROM batch_factory_v12_comic_source_segments WHERE owner_username=? AND batch_id=? AND book_id=?`, owner, batchID, bookID).Scan(&version); err != nil {
		return ComicSourceSegmentation{}, err
	}
	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_videos SET compatibility_state='orphaned',updated_at=? WHERE owner_username=? AND batch_id=? AND book_id=? AND compatibility_state='active'`, now, owner, batchID, bookID); err != nil {
		return ComicSourceSegmentation{}, err
	}
	for ordinal, draft := range drafts {
		videoID, err := newID("video")
		if err != nil {
			return ComicSourceSegmentation{}, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_videos(id,batch_id,book_id,owner_username,ordinal,compatibility_state,revision,created_at,updated_at) VALUES(?,?,?,?,?,'active',1,?,?)`, videoID, batchID, bookID, owner, ordinal, now, now); err != nil {
			return ComicSourceSegmentation{}, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_video_records(video_id,label,video_prompt,visual_prompt,duration_seconds) VALUES(?,?,NULL,NULL,NULL)`, videoID, fmt.Sprintf("漫画页 %02d", ordinal+1)); err != nil {
			return ComicSourceSegmentation{}, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_comic_source_segments(video_id,owner_username,batch_id,book_id,start_offset,end_offset,source_text,character_count,boundary_type,reason,source_digest,segmentation_version,prompt_stale,media_stale,revision,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,FALSE,FALSE,1,?,?)`, videoID, owner, batchID, bookID, draft.StartOffset, draft.EndOffset, draft.SourceText, draft.CharacterCount, draft.BoundaryType, draft.Reason, digest, version, now, now); err != nil {
			return ComicSourceSegmentation{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_books SET revision=revision+1,updated_at=? WHERE id=? AND batch_id=? AND owner_username=?`, now, bookID, batchID, owner); err != nil {
		return ComicSourceSegmentation{}, err
	}
	if err := tx.Commit(); err != nil {
		return ComicSourceSegmentation{}, err
	}
	return s.GetComicSourceSegmentation(ctx, owner, batchID, bookID)
}

func (s *MySQLStore) BindComicSourceSegments(ctx context.Context, owner, batchID, bookID, digest string, videoIDs []string, drafts []ComicSourceSegmentDraft) (ComicSourceSegmentation, error) {
	if len(videoIDs) != len(drafts) || len(videoIDs) == 0 {
		return ComicSourceSegmentation{}, fmt.Errorf("%w: 漫画页与原文分段数量不一致", ErrInvalid)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ComicSourceSegmentation{}, err
	}
	defer tx.Rollback()
	var source string
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(NULLIF(d.content,''),r.source_text,'') FROM batch_factory_v12_books b JOIN batch_factory_v12_book_records r ON r.book_id=b.id LEFT JOIN batch_factory_v12_drafts d ON d.owner_username=b.owner_username AND d.draft_key=CONCAT('working-front:',b.id) AND d.kind='working-front-content' AND d.scope=b.batch_id WHERE b.id=? AND b.batch_id=? AND b.owner_username=? FOR UPDATE`, bookID, batchID, owner).Scan(&source); errors.Is(err, sql.ErrNoRows) {
		return ComicSourceSegmentation{}, ErrNotFound
	} else if err != nil {
		return ComicSourceSegmentation{}, err
	}
	if digest != sourceDigest(source) {
		return ComicSourceSegmentation{}, fmt.Errorf("%w: 漫画正文已变化", ErrConflict)
	}
	if err := ValidateComicSourceSegments(source, drafts); err != nil {
		return ComicSourceSegmentation{}, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id FROM batch_factory_v12_videos WHERE owner_username=? AND batch_id=? AND book_id=? AND compatibility_state='active' ORDER BY ordinal,id FOR UPDATE`, owner, batchID, bookID)
	if err != nil {
		return ComicSourceSegmentation{}, err
	}
	activeIDs := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return ComicSourceSegmentation{}, err
		}
		activeIDs = append(activeIDs, id)
	}
	if err := rows.Close(); err != nil {
		return ComicSourceSegmentation{}, err
	}
	if len(activeIDs) != len(videoIDs) {
		return ComicSourceSegmentation{}, fmt.Errorf("%w: 漫画页已变化", ErrConflict)
	}
	for index := range activeIDs {
		if activeIDs[index] != videoIDs[index] {
			return ComicSourceSegmentation{}, fmt.Errorf("%w: 漫画页顺序已变化", ErrConflict)
		}
	}
	var version int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(segmentation_version),0)+1 FROM batch_factory_v12_comic_source_segments WHERE owner_username=? AND batch_id=? AND book_id=?`, owner, batchID, bookID).Scan(&version); err != nil {
		return ComicSourceSegmentation{}, err
	}
	now := time.Now().UTC()
	for ordinal, draft := range drafts {
		if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_comic_source_segments(video_id,owner_username,batch_id,book_id,start_offset,end_offset,source_text,character_count,boundary_type,reason,source_digest,segmentation_version,prompt_stale,media_stale,revision,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,FALSE,FALSE,1,?,?) ON DUPLICATE KEY UPDATE start_offset=VALUES(start_offset),end_offset=VALUES(end_offset),source_text=VALUES(source_text),character_count=VALUES(character_count),boundary_type=VALUES(boundary_type),reason=VALUES(reason),source_digest=VALUES(source_digest),segmentation_version=VALUES(segmentation_version),prompt_stale=FALSE,media_stale=FALSE,revision=revision+1,updated_at=VALUES(updated_at)`, videoIDs[ordinal], owner, batchID, bookID, draft.StartOffset, draft.EndOffset, draft.SourceText, draft.CharacterCount, draft.BoundaryType, draft.Reason, digest, version, now, now); err != nil {
			return ComicSourceSegmentation{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_books SET revision=revision+1,updated_at=? WHERE id=? AND batch_id=? AND owner_username=?`, now, bookID, batchID, owner); err != nil {
		return ComicSourceSegmentation{}, err
	}
	if err := tx.Commit(); err != nil {
		return ComicSourceSegmentation{}, err
	}
	return s.GetComicSourceSegmentation(ctx, owner, batchID, bookID)
}

func (s *MySQLStore) GetComicSourceSegmentation(ctx context.Context, owner, batchID, bookID string) (ComicSourceSegmentation, error) {
	var source string
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(NULLIF(d.content,''),r.source_text,'') FROM batch_factory_v12_books b JOIN batch_factory_v12_book_records r ON r.book_id=b.id LEFT JOIN batch_factory_v12_drafts d ON d.owner_username=b.owner_username AND d.draft_key=CONCAT('working-front:',b.id) AND d.kind='working-front-content' AND d.scope=b.batch_id WHERE b.id=? AND b.batch_id=? AND b.owner_username=?`, bookID, batchID, owner).Scan(&source); errors.Is(err, sql.ErrNoRows) {
		return ComicSourceSegmentation{}, ErrNotFound
	} else if err != nil {
		return ComicSourceSegmentation{}, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT s.video_id,v.ordinal,s.start_offset,s.end_offset,s.source_text,s.character_count,s.boundary_type,s.reason,s.source_digest,s.segmentation_version,s.prompt_stale,s.media_stale,s.revision FROM batch_factory_v12_comic_source_segments s JOIN batch_factory_v12_videos v ON v.id=s.video_id WHERE s.owner_username=? AND s.batch_id=? AND s.book_id=? AND v.compatibility_state='active' ORDER BY v.ordinal,v.id`, owner, batchID, bookID)
	if err != nil {
		return ComicSourceSegmentation{}, err
	}
	defer rows.Close()
	result := ComicSourceSegmentation{BatchID: batchID, BookID: bookID, CurrentDigest: sourceDigest(source)}
	for rows.Next() {
		var segment ComicSourceSegment
		segment.BatchID, segment.BookID = batchID, bookID
		if err := rows.Scan(&segment.VideoID, &segment.Ordinal, &segment.StartOffset, &segment.EndOffset, &segment.SourceText, &segment.CharacterCount, &segment.BoundaryType, &segment.Reason, &segment.SourceDigest, &segment.SegmentationVersion, &segment.PromptStale, &segment.MediaStale, &segment.Revision); err != nil {
			return ComicSourceSegmentation{}, err
		}
		segment = comicSourceSegmentWithBoundaries(segment)
		result.Segments = append(result.Segments, segment)
		result.SourceDigest = segment.SourceDigest
		if segment.SegmentationVersion > result.Version {
			result.Version = segment.SegmentationVersion
		}
	}
	if err := rows.Err(); err != nil {
		return ComicSourceSegmentation{}, err
	}
	if len(result.Segments) == 0 {
		return ComicSourceSegmentation{}, ErrNotFound
	}
	result.Stale = result.SourceDigest != result.CurrentDigest
	return result, nil
}

func (s *MySQLStore) SplitComicSourceSegment(ctx context.Context, owner, batchID, bookID, videoID string, input ComicSourceSplitInput) (ComicSourceSegmentation, error) {
	if input.Direction != "left" && input.Direction != "right" {
		return ComicSourceSegmentation{}, fmt.Errorf("%w: 拆分方向无效", ErrInvalid)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ComicSourceSegmentation{}, err
	}
	defer tx.Rollback()
	var source string
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(NULLIF(d.content,''),r.source_text,'') FROM batch_factory_v12_books b JOIN batch_factory_v12_book_records r ON r.book_id=b.id LEFT JOIN batch_factory_v12_drafts d ON d.owner_username=b.owner_username AND d.draft_key=CONCAT('working-front:',b.id) AND d.kind='working-front-content' AND d.scope=b.batch_id WHERE b.id=? AND b.batch_id=? AND b.owner_username=? FOR UPDATE`, bookID, batchID, owner).Scan(&source); errors.Is(err, sql.ErrNoRows) {
		return ComicSourceSegmentation{}, ErrNotFound
	} else if err != nil {
		return ComicSourceSegmentation{}, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT s.video_id,v.ordinal,s.start_offset,s.end_offset,s.source_text,s.character_count,s.boundary_type,s.reason,s.source_digest,s.segmentation_version,s.prompt_stale,s.media_stale,s.revision FROM batch_factory_v12_comic_source_segments s JOIN batch_factory_v12_videos v ON v.id=s.video_id WHERE s.owner_username=? AND s.batch_id=? AND s.book_id=? AND v.compatibility_state='active' ORDER BY v.ordinal,v.id FOR UPDATE`, owner, batchID, bookID)
	if err != nil {
		return ComicSourceSegmentation{}, err
	}
	segments := []ComicSourceSegment{}
	currentIndex := -1
	for rows.Next() {
		var segment ComicSourceSegment
		segment.BatchID, segment.BookID = batchID, bookID
		if err := rows.Scan(&segment.VideoID, &segment.Ordinal, &segment.StartOffset, &segment.EndOffset, &segment.SourceText, &segment.CharacterCount, &segment.BoundaryType, &segment.Reason, &segment.SourceDigest, &segment.SegmentationVersion, &segment.PromptStale, &segment.MediaStale, &segment.Revision); err != nil {
			rows.Close()
			return ComicSourceSegmentation{}, err
		}
		if segment.VideoID == videoID {
			currentIndex = len(segments)
		}
		segments = append(segments, segment)
	}
	rows.Close()
	if currentIndex < 0 {
		return ComicSourceSegmentation{}, ErrNotFound
	}
	current := segments[currentIndex]
	if current.Revision != input.ExpectedRevision {
		return ComicSourceSegmentation{}, ErrConflict
	}
	valid, boundaryType := false, "sentence"
	for _, boundary := range comicSourceSegmentWithBoundaries(current).SafeBoundaries {
		if boundary.Offset == input.BoundaryOffset {
			valid, boundaryType = true, boundary.Type
			break
		}
	}
	if !valid {
		return ComicSourceSegmentation{}, fmt.Errorf("%w: 只能在完整句子或连续对白结束处拆分", ErrInvalid)
	}
	cut := input.BoundaryOffset - current.StartOffset
	prefix, suffix := current.SourceText[:cut], current.SourceText[cut:]
	newID, err := newID("video")
	if err != nil {
		return ComicSourceSegmentation{}, err
	}
	version := current.SegmentationVersion + 1
	newSegment := ComicSourceSegment{VideoID: newID, BatchID: batchID, BookID: bookID, SourceDigest: current.SourceDigest, SegmentationVersion: version, PromptStale: true, MediaStale: true, Revision: 1, Reason: "manual_split"}
	current.SegmentationVersion, current.PromptStale, current.MediaStale, current.Revision, current.Reason = version, true, true, current.Revision+1, "manual_split"
	if input.Direction == "left" {
		newSegment.StartOffset, newSegment.EndOffset, newSegment.SourceText, newSegment.BoundaryType = current.StartOffset, input.BoundaryOffset, prefix, boundaryType
		current.StartOffset, current.SourceText = input.BoundaryOffset, suffix
		segments = append(segments[:currentIndex], append([]ComicSourceSegment{newSegment, current}, segments[currentIndex+1:]...)...)
	} else {
		originalEnd, originalBoundary := current.EndOffset, current.BoundaryType
		current.EndOffset, current.SourceText, current.BoundaryType = input.BoundaryOffset, prefix, boundaryType
		newSegment.StartOffset, newSegment.EndOffset, newSegment.SourceText, newSegment.BoundaryType = input.BoundaryOffset, originalEnd, suffix, originalBoundary
		segments = append(segments[:currentIndex], append([]ComicSourceSegment{current, newSegment}, segments[currentIndex+1:]...)...)
	}
	current.CharacterCount, newSegment.CharacterCount = comicSourceCharacterCount(current.SourceText), comicSourceCharacterCount(newSegment.SourceText)
	drafts := make([]ComicSourceSegmentDraft, len(segments))
	for index := range segments {
		segments[index].Ordinal, segments[index].SegmentationVersion = index, version
		drafts[index] = ComicSourceSegmentDraft{StartOffset: segments[index].StartOffset, EndOffset: segments[index].EndOffset, SourceText: segments[index].SourceText}
	}
	if err := ValidateComicSourceSegments(source, drafts); err != nil {
		return ComicSourceSegmentation{}, err
	}
	now := time.Now().UTC()
	if input.Direction == "left" {
		_, err = tx.ExecContext(ctx, `UPDATE batch_factory_v12_videos SET ordinal=ordinal+1,updated_at=? WHERE owner_username=? AND batch_id=? AND book_id=? AND compatibility_state='active' AND ordinal>=?`, now, owner, batchID, bookID, current.Ordinal)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE batch_factory_v12_videos SET ordinal=ordinal+1,updated_at=? WHERE owner_username=? AND batch_id=? AND book_id=? AND compatibility_state='active' AND ordinal>?`, now, owner, batchID, bookID, current.Ordinal)
	}
	if err != nil {
		return ComicSourceSegmentation{}, err
	}
	newOrdinal := current.Ordinal
	if input.Direction == "right" {
		newOrdinal++
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_videos(id,batch_id,book_id,owner_username,ordinal,compatibility_state,revision,created_at,updated_at) VALUES(?,?,?,?,?,'active',1,?,?)`, newID, batchID, bookID, owner, newOrdinal, now, now); err != nil {
		return ComicSourceSegmentation{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_video_records(video_id,label,video_prompt,visual_prompt,duration_seconds) VALUES(?,?,NULL,NULL,NULL)`, newID, fmt.Sprintf("漫画页 %02d", newOrdinal+1)); err != nil {
		return ComicSourceSegmentation{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_video_records r JOIN batch_factory_v12_videos v ON v.id=r.video_id SET r.label=CONCAT('漫画页 ',LPAD(v.ordinal+1,2,'0')) WHERE v.owner_username=? AND v.batch_id=? AND v.book_id=? AND v.compatibility_state='active'`, owner, batchID, bookID); err != nil {
		return ComicSourceSegmentation{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_video_records SET video_prompt=NULL,visual_prompt=NULL WHERE video_id=?`, videoID); err != nil {
		return ComicSourceSegmentation{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_comic_source_segments SET start_offset=?,end_offset=?,source_text=?,character_count=?,boundary_type=?,reason='manual_split',segmentation_version=?,prompt_stale=TRUE,media_stale=TRUE,revision=revision+1,updated_at=? WHERE video_id=? AND owner_username=?`, current.StartOffset, current.EndOffset, current.SourceText, current.CharacterCount, current.BoundaryType, version, now, videoID, owner); err != nil {
		return ComicSourceSegmentation{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_comic_source_segments(video_id,owner_username,batch_id,book_id,start_offset,end_offset,source_text,character_count,boundary_type,reason,source_digest,segmentation_version,prompt_stale,media_stale,revision,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,'manual_split',?,?,TRUE,TRUE,1,?,?)`, newID, owner, batchID, bookID, newSegment.StartOffset, newSegment.EndOffset, newSegment.SourceText, newSegment.CharacterCount, newSegment.BoundaryType, newSegment.SourceDigest, version, now, now); err != nil {
		return ComicSourceSegmentation{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_comic_source_segments SET segmentation_version=? WHERE owner_username=? AND batch_id=? AND book_id=?`, version, owner, batchID, bookID); err != nil {
		return ComicSourceSegmentation{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_comic_pages SET status='prompt_stale',revision=revision+1,updated_at=? WHERE owner_username=? AND batch_id=? AND book_id=? AND video_id=?`, now, owner, batchID, bookID, videoID); err != nil {
		return ComicSourceSegmentation{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_books SET revision=revision+1,updated_at=? WHERE id=? AND batch_id=? AND owner_username=?`, now, bookID, batchID, owner); err != nil {
		return ComicSourceSegmentation{}, err
	}
	if err := tx.Commit(); err != nil {
		return ComicSourceSegmentation{}, err
	}
	return s.GetComicSourceSegmentation(ctx, owner, batchID, bookID)
}

var _ ComicSourceSegmentStore = (*MySQLStore)(nil)
