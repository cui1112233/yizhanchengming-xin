package batchfactoryv11

import (
	"context"
	"database/sql"
	"errors"
)

// ReadbackMySQLStore keeps MySQLStore as the single write authority while
// hydrating the sparse settings state required by the V11 browser contract.
// It deliberately does not copy or reinterpret patches: loadPatch returns the
// exact persisted JSON values, including false, "", and 0.
type ReadbackMySQLStore struct {
	*MySQLStore
}

var _ Store = (*ReadbackMySQLStore)(nil)

func NewReadbackMySQLStore(db *sql.DB) *ReadbackMySQLStore {
	return &ReadbackMySQLStore{MySQLStore: NewMySQLStore(db)}
}

func (s *ReadbackMySQLStore) hydrateSettingsState(ctx context.Context, owner string, batch Batch) (Batch, error) {
	patch, err := loadPatch(ctx, s.db, owner, ScopeRef{Kind: ScopeBatch, BatchID: batch.ID})
	if err != nil {
		return Batch{}, err
	}
	batch.SettingsState = SettingsState{Patch: clonePatch(patch), Revision: batch.Revision}

	for bookIndex := range batch.Books {
		book := &batch.Books[bookIndex]
		patch, err := loadPatch(ctx, s.db, owner, ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID})
		if err != nil {
			return Batch{}, err
		}
		book.SettingsState = SettingsState{Patch: clonePatch(patch), Revision: book.Revision}
		effective := ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
		book.Mode = rawString(effective, "productionMode", rawString(effective, "mode", "original"))
		if book.Mode == "original_direct" {
			book.Mode = "original"
		}
		if book.Mode == "viral_hook" {
			book.Mode = "viral"
		}
		if hook, err := s.LatestHookRevision(ctx, owner, batch.ID, book.ID); err == nil {
			book.Hook = &hook
		} else if !errors.Is(err, ErrNotFound) {
			return Batch{}, err
		}
		if revision, err := loadLatestDirectorRevision(ctx, s.db, owner, batch.ID, book.ID); err == nil {
			revision.Videos = append([]Video(nil), book.Videos...)
			book.DirectorRevision = &revision
			book.Assets = DirectorAssets{Characters: revision.Output.Characters, Scenes: revision.Output.Scenes, Props: revision.Output.Props}
		} else if !errors.Is(err, ErrNotFound) {
			return Batch{}, err
		}
		assets, err := s.ListBookAssets(ctx, owner, batch.ID, book.ID)
		if err != nil {
			return Batch{}, err
		}
		book.AssetRecords = assets
		if len(assets) > 0 {
			book.Assets = directorAssetsFromRecords(assets)
		}

		for videoIndex := range book.Videos {
			video := &book.Videos[videoIndex]
			patch, err := loadPatch(ctx, s.db, owner, ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID})
			if err != nil {
				return Batch{}, err
			}
			video.SettingsState = SettingsState{Patch: clonePatch(patch), Revision: video.Revision}
			applyVideoPromptOverrides(video)
			if _, overridden := optionalPromptValue(video.SettingsState.Patch, "videoPrompt"); !overridden && book.DirectorRevision != nil && videoIndex < len(book.DirectorRevision.Output.Storyboard) {
				video.VideoPrompt = storyboardVideoPrompt(book.DirectorRevision.Output.Storyboard[videoIndex])
			}
		}
		// 非 active（orphaned）的历史视频记录也要下发：自动化据此判断“这个分镜
		// 已有成功成片、不必再次生成”，媒体库也据此展示所有历史版本。
		history, err := s.loadVideoHistory(ctx, owner, batch.ID, book.ID)
		if err != nil {
			return Batch{}, err
		}
		book.VideoHistory = history
		segmentation, err := s.GetComicSourceSegmentation(ctx, owner, batch.ID, book.ID)
		if err == nil {
			segmentsByVideo := make(map[string]ComicSourceSegment, len(segmentation.Segments))
			for _, segment := range segmentation.Segments {
				segmentsByVideo[segment.VideoID] = segment
			}
			for videoIndex := range book.Videos {
				if segment, ok := segmentsByVideo[book.Videos[videoIndex].ID]; ok {
					book.Videos[videoIndex].ComicSourceSegment = &segment
				}
			}
		} else if !errors.Is(err, ErrNotFound) {
			return Batch{}, err
		}
	}
	return batch, nil
}

func (s *ReadbackMySQLStore) GetBatch(ctx context.Context, owner, id string) (Batch, error) {
	batch, err := s.MySQLStore.GetBatch(ctx, owner, id)
	if err != nil {
		return Batch{}, err
	}
	return s.hydrateSettingsState(ctx, owner, batch)
}

// loadVideoHistory 读取本书所有非 active（被轮换成 orphaned）的视频槽位。
// 只取轻量字段：前端要把历史成功成片挂回同一个分镜，自动化要用它判断
// 该分镜是否已经有成功成片、从而跳过重复生成。
func (s *ReadbackMySQLStore) loadVideoHistory(ctx context.Context, owner, batchID, bookID string) ([]Video, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,ordinal,compatibility_state FROM batch_factory_v12_videos WHERE batch_id=? AND book_id=? AND owner_username=? AND compatibility_state<>'active' ORDER BY ordinal,updated_at DESC,id`, batchID, bookID, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var videos []Video
	for rows.Next() {
		var video Video
		if err := rows.Scan(&video.ID, &video.Ordinal, &video.CompatibilityState); err != nil {
			return nil, err
		}
		videos = append(videos, video)
	}
	return videos, rows.Err()
}

func (s *ReadbackMySQLStore) ListBatches(ctx context.Context, owner string) ([]Batch, error) {
	// The library route only needs identity and book-count data. Hydrating
	// settings, director documents, assets and videos here makes one damaged
	// historical record hide every otherwise-readable batch from the library.
	return s.MySQLStore.ListBatches(ctx, owner)
}
