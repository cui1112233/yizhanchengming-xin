package batchfactoryv11

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Comic pages are written after a relatively long provider request. Several
// pages can finish at nearly the same time, so the short default retry window
// used to turn a transient InnoDB deadlock into a lost generated image.
const comicPageWriteRetryLimit = 8

func retryableComicPageWriteError(err error) bool {
	if errors.Is(err, driver.ErrBadConn) {
		return true
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1213 || mysqlErr.Number == 1205
	}
	return false
}

func waitComicPageWriteRetry(ctx context.Context, attempt int) error {
	delay := time.Duration(attempt*250) * time.Millisecond
	if delay > 2*time.Second {
		delay = 2 * time.Second
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func ensureComicPageOwnership(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, owner, batchID, bookID, pageID string) error {
	var one int
	err := q.QueryRowContext(ctx, `SELECT 1 FROM batch_factory_v12_comic_pages WHERE id=? AND owner_username=? AND batch_id=? AND book_id=?`, pageID, owner, batchID, bookID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func ensureComicVideoOwnership(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, owner, batchID, bookID, videoID string) error {
	var one int
	err := q.QueryRowContext(ctx, `SELECT 1 FROM batch_factory_v12_videos WHERE id=? AND owner_username=? AND batch_id=? AND book_id=? AND compatibility_state='active'`, videoID, owner, batchID, bookID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (s *MySQLStore) UpsertComicPage(ctx context.Context, owner, batchID, bookID, videoID string, input UpsertComicPageInput) (ComicPage, error) {
	var lastErr error
	for attempt := 1; attempt <= comicPageWriteRetryLimit; attempt++ {
		page, err := s.upsertComicPageOnce(ctx, owner, batchID, bookID, videoID, input)
		if err == nil || !retryableComicPageWriteError(err) || attempt == comicPageWriteRetryLimit {
			return page, err
		}
		lastErr = err
		if waitErr := waitComicPageWriteRetry(ctx, attempt); waitErr != nil {
			return ComicPage{}, waitErr
		}
	}
	return ComicPage{}, lastErr
}

func (s *MySQLStore) upsertComicPageOnce(ctx context.Context, owner, batchID, bookID, videoID string, input UpsertComicPageInput) (ComicPage, error) {
	input, err := normalizeComicPageInput(input)
	if err != nil {
		return ComicPage{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ComicPage{}, err
	}
	defer tx.Rollback()
	if err := ensureBookOwnership(ctx, tx, owner, batchID, bookID); err != nil {
		return ComicPage{}, err
	}
	if err := ensureComicVideoOwnership(ctx, tx, owner, batchID, bookID, videoID); err != nil {
		return ComicPage{}, err
	}
	now := time.Now().UTC()
	var page ComicPage
	err = tx.QueryRowContext(ctx, `SELECT id,revision,created_at FROM batch_factory_v12_comic_pages WHERE owner_username=? AND batch_id=? AND book_id=? AND video_id=? FOR UPDATE`, owner, batchID, bookID, videoID).Scan(&page.ID, &page.Revision, &page.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		if input.ExpectedRevision > 0 {
			// 调用方认为自己持有某一版页面，但这一行已经不存在，说明页面被
			// 删除或重建过；按版本冲突处理，避免把旧结果写到新页面上。
			return ComicPage{}, fmt.Errorf("%w: 漫画页版本已变化", ErrConflict)
		}
		page.ID, err = newID("comic-page")
		if err != nil {
			return ComicPage{}, err
		}
		page.Revision, page.CreatedAt = 1, now
		_, err = tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_comic_pages(id,owner_username,batch_id,book_id,video_id,comic_prompt,image_prompt,model_id,aspect_ratio,status,snapshot_hash,reference_image_urls,revision,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,1,?,?)`, page.ID, owner, batchID, bookID, videoID, input.ComicPrompt, input.ImagePrompt, nullableString(input.ModelID), input.AspectRatio, input.Status, nullableString(input.SnapshotHash), comicReferenceURLsJSON(input.ReferenceImageURLs), now, now)
	} else if err == nil {
		if input.ExpectedRevision > 0 && page.Revision != input.ExpectedRevision {
			return ComicPage{}, fmt.Errorf("%w: 漫画页版本已变化", ErrConflict)
		}
		page.Revision++
		_, err = tx.ExecContext(ctx, `UPDATE batch_factory_v12_comic_pages SET comic_prompt=?,image_prompt=?,model_id=?,aspect_ratio=?,status=?,snapshot_hash=?,reference_image_urls=?,revision=?,updated_at=? WHERE id=? AND owner_username=?`, input.ComicPrompt, input.ImagePrompt, nullableString(input.ModelID), input.AspectRatio, input.Status, nullableString(input.SnapshotHash), comicReferenceURLsJSON(input.ReferenceImageURLs), page.Revision, now, page.ID, owner)
	}
	if err != nil {
		return ComicPage{}, err
	}
	if err := tx.Commit(); err != nil {
		return ComicPage{}, err
	}
	page.BatchID, page.BookID, page.VideoID = batchID, bookID, videoID
	page.ComicPrompt, page.ImagePrompt, page.ModelID = input.ComicPrompt, input.ImagePrompt, input.ModelID
	page.AspectRatio, page.Status, page.SnapshotHash, page.ReferenceImageURLs, page.UpdatedAt = input.AspectRatio, input.Status, input.SnapshotHash, input.ReferenceImageURLs, now
	return page, nil
}

func (s *MySQLStore) PatchComicPage(ctx context.Context, owner, batchID, bookID, videoID string, patch PatchComicPageInput) (ComicPage, error) {
	var lastErr error
	for attempt := 1; attempt <= comicPageWriteRetryLimit; attempt++ {
		page, err := s.patchComicPageOnce(ctx, owner, batchID, bookID, videoID, patch)
		if err == nil || !retryableComicPageWriteError(err) || attempt == comicPageWriteRetryLimit {
			return page, err
		}
		lastErr = err
		if waitErr := waitComicPageWriteRetry(ctx, attempt); waitErr != nil {
			return ComicPage{}, waitErr
		}
	}
	return ComicPage{}, lastErr
}

func (s *MySQLStore) patchComicPageOnce(ctx context.Context, owner, batchID, bookID, videoID string, patch PatchComicPageInput) (ComicPage, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ComicPage{}, err
	}
	defer tx.Rollback()
	if err := ensureBookOwnership(ctx, tx, owner, batchID, bookID); err != nil {
		return ComicPage{}, err
	}
	if err := ensureComicVideoOwnership(ctx, tx, owner, batchID, bookID, videoID); err != nil {
		return ComicPage{}, err
	}
	now := time.Now().UTC()
	var (
		pageID    string
		revision  int64
		createdAt time.Time
		existing  UpsertComicPageInput
		refsJSON  []byte
	)
	err = tx.QueryRowContext(ctx, `SELECT id,revision,created_at,comic_prompt,image_prompt,COALESCE(model_id,''),aspect_ratio,status,COALESCE(snapshot_hash,''),COALESCE(reference_image_urls,JSON_ARRAY()) FROM batch_factory_v12_comic_pages WHERE owner_username=? AND batch_id=? AND book_id=? AND video_id=? FOR UPDATE`, owner, batchID, bookID, videoID).Scan(&pageID, &revision, &createdAt, &existing.ComicPrompt, &existing.ImagePrompt, &existing.ModelID, &existing.AspectRatio, &existing.Status, &existing.SnapshotHash, &refsJSON)
	if errors.Is(err, sql.ErrNoRows) {
		merged, mergeErr := applyComicPagePatch(UpsertComicPageInput{}, patch)
		if mergeErr != nil {
			return ComicPage{}, mergeErr
		}
		pageID, err = newID("comic-page")
		if err != nil {
			return ComicPage{}, err
		}
		revision, createdAt = 1, now
		if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_comic_pages(id,owner_username,batch_id,book_id,video_id,comic_prompt,image_prompt,model_id,aspect_ratio,status,snapshot_hash,reference_image_urls,revision,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,1,?,?)`, pageID, owner, batchID, bookID, videoID, merged.ComicPrompt, merged.ImagePrompt, nullableString(merged.ModelID), merged.AspectRatio, merged.Status, nullableString(merged.SnapshotHash), comicReferenceURLsJSON(merged.ReferenceImageURLs), now, now); err != nil {
			return ComicPage{}, err
		}
		if err := tx.Commit(); err != nil {
			return ComicPage{}, err
		}
		return comicPageFromUpsertInput(pageID, revision, createdAt, now, batchID, bookID, videoID, merged), nil
	}
	if err != nil {
		return ComicPage{}, err
	}
	_ = json.Unmarshal(refsJSON, &existing.ReferenceImageURLs)
	merged, err := applyComicPagePatch(existing, patch)
	if err != nil {
		return ComicPage{}, err
	}
	revision++
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_comic_pages SET comic_prompt=?,image_prompt=?,model_id=?,aspect_ratio=?,status=?,snapshot_hash=?,reference_image_urls=?,revision=?,updated_at=? WHERE id=? AND owner_username=?`, merged.ComicPrompt, merged.ImagePrompt, nullableString(merged.ModelID), merged.AspectRatio, merged.Status, nullableString(merged.SnapshotHash), comicReferenceURLsJSON(merged.ReferenceImageURLs), revision, now, pageID, owner); err != nil {
		return ComicPage{}, err
	}
	if err := tx.Commit(); err != nil {
		return ComicPage{}, err
	}
	return comicPageFromUpsertInput(pageID, revision, createdAt, now, batchID, bookID, videoID, merged), nil
}

func comicPageFromUpsertInput(pageID string, revision int64, createdAt, updatedAt time.Time, batchID, bookID, videoID string, input UpsertComicPageInput) ComicPage {
	return ComicPage{
		ID: pageID, BatchID: batchID, BookID: bookID, VideoID: videoID,
		ComicPrompt: input.ComicPrompt, ImagePrompt: input.ImagePrompt, ModelID: input.ModelID,
		AspectRatio: input.AspectRatio, Status: input.Status, SnapshotHash: input.SnapshotHash,
		ReferenceImageURLs: input.ReferenceImageURLs, Revision: revision, CreatedAt: createdAt, UpdatedAt: updatedAt,
	}
}

func (s *MySQLStore) ListComicPages(ctx context.Context, owner, batchID, bookID string) ([]ComicPage, error) {
	if err := ensureBookOwnership(ctx, s.db, owner, batchID, bookID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT p.id,p.video_id,p.comic_prompt,p.image_prompt,COALESCE(p.model_id,''),p.aspect_ratio,p.status,COALESCE(p.snapshot_hash,''),COALESCE(p.reference_image_urls,JSON_ARRAY()),p.revision,p.created_at,p.updated_at FROM batch_factory_v12_comic_pages p JOIN batch_factory_v12_videos v ON v.id=p.video_id WHERE p.owner_username=? AND p.batch_id=? AND p.book_id=? AND v.compatibility_state='active' ORDER BY v.ordinal,v.id`, owner, batchID, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	pages := []ComicPage{}
	for rows.Next() {
		var page ComicPage
		page.BatchID, page.BookID = batchID, bookID
		var referencesJSON []byte
		if err := rows.Scan(&page.ID, &page.VideoID, &page.ComicPrompt, &page.ImagePrompt, &page.ModelID, &page.AspectRatio, &page.Status, &page.SnapshotHash, &referencesJSON, &page.Revision, &page.CreatedAt, &page.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(referencesJSON, &page.ReferenceImageURLs)
		pages = append(pages, page)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(pages) == 0 {
		return pages, nil
	}
	pageIndexes := make(map[string]int, len(pages))
	for index := range pages {
		pageIndexes[pages[index].ID] = index
	}
	imageRows, err := s.db.QueryContext(ctx, `SELECT i.id,COALESCE(i.url,''),COALESCE(i.storage_ref,''),i.media_type,i.source,i.is_primary,i.revision,i.created_at,p.id FROM batch_factory_v12_comic_page_images i JOIN batch_factory_v12_comic_pages p ON p.id=i.comic_page_id AND p.owner_username=i.owner_username JOIN batch_factory_v12_videos v ON v.id=p.video_id WHERE i.owner_username=? AND p.batch_id=? AND p.book_id=? AND v.compatibility_state='active' ORDER BY v.ordinal,v.id,i.created_at,i.id`, owner, batchID, bookID)
	if err != nil {
		return nil, err
	}
	defer imageRows.Close()
	for imageRows.Next() {
		var image ComicPageImage
		if err := imageRows.Scan(&image.ID, &image.URL, &image.StorageRef, &image.MediaType, &image.Source, &image.IsPrimary, &image.Revision, &image.CreatedAt, &image.ComicPageID); err != nil {
			return nil, err
		}
		if index, ok := pageIndexes[image.ComicPageID]; ok {
			pages[index].Images = append(pages[index].Images, image)
		}
	}
	if err := imageRows.Err(); err != nil {
		return nil, err
	}
	return pages, nil
}

func comicReferenceURLsJSON(urls []string) string {
	payload, _ := json.Marshal(normalizeComicReferenceURLs(urls))
	return string(payload)
}

// GetBookComicSourceIndex reads only the book source text and the per-video
// segmentation digests. The comic-pages read path used to call GetBatch here,
// which hydrated every book, video and draft in the batch just to compare one
// book's digests; with many books that dominated the "enter batch" latency.
func (s *MySQLStore) GetBookComicSourceIndex(ctx context.Context, owner, batchID, bookID string) (ComicSourceIndex, bool, error) {
	var sourceText string
	err := s.db.QueryRowContext(ctx, `SELECT COALESCE(r.source_text,'') FROM batch_factory_v12_books b LEFT JOIN batch_factory_v12_book_records r ON r.book_id=b.id WHERE b.id=? AND b.batch_id=? AND b.owner_username=?`, bookID, batchID, owner).Scan(&sourceText)
	if errors.Is(err, sql.ErrNoRows) {
		return ComicSourceIndex{}, false, nil
	}
	if err != nil {
		return ComicSourceIndex{}, false, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT s.video_id,s.source_digest FROM batch_factory_v12_comic_source_segments s JOIN batch_factory_v12_videos v ON v.id=s.video_id AND v.owner_username=s.owner_username WHERE s.owner_username=? AND s.batch_id=? AND s.book_id=? AND v.compatibility_state='active'`, owner, batchID, bookID)
	if err != nil {
		return ComicSourceIndex{}, false, err
	}
	defer rows.Close()
	digests := map[string]string{}
	for rows.Next() {
		var videoID, digest string
		if err := rows.Scan(&videoID, &digest); err != nil {
			return ComicSourceIndex{}, false, err
		}
		digests[videoID] = digest
	}
	if err := rows.Err(); err != nil {
		return ComicSourceIndex{}, false, err
	}
	return ComicSourceIndex{SourceText: sourceText, VideoDigests: digests}, true, nil
}

func (s *MySQLStore) SetComicPageReferenceImages(ctx context.Context, owner, batchID, bookID, videoID string, urls []string) (ComicPage, error) {
	if err := ensureBookOwnership(ctx, s.db, owner, batchID, bookID); err != nil {
		return ComicPage{}, err
	}
	urls = normalizeComicReferenceURLs(urls)
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx, `UPDATE batch_factory_v12_comic_pages SET reference_image_urls=?,revision=revision+1,updated_at=? WHERE owner_username=? AND batch_id=? AND book_id=? AND video_id=?`, comicReferenceURLsJSON(urls), now, owner, batchID, bookID, videoID)
	if err != nil {
		return ComicPage{}, err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return ComicPage{}, err
	}
	if changed == 0 {
		return ComicPage{}, ErrNotFound
	}
	pages, err := s.ListComicPages(ctx, owner, batchID, bookID)
	if err != nil {
		return ComicPage{}, err
	}
	for _, page := range pages {
		if page.VideoID == videoID {
			return page, nil
		}
	}
	return ComicPage{}, ErrNotFound
}

func (s *MySQLStore) ApplyComicPageReferenceImages(ctx context.Context, owner, batchID, bookID, sourceVideoID string) (int, error) {
	pages, err := s.ListComicPages(ctx, owner, batchID, bookID)
	if err != nil {
		return 0, err
	}
	var references []string
	found := false
	for _, page := range pages {
		if page.VideoID == sourceVideoID {
			references, found = page.ReferenceImageURLs, true
			break
		}
	}
	if !found {
		return 0, ErrNotFound
	}
	result, err := s.db.ExecContext(ctx, `UPDATE batch_factory_v12_comic_pages SET reference_image_urls=?,revision=revision+1,updated_at=? WHERE owner_username=? AND batch_id=? AND book_id=?`, comicReferenceURLsJSON(references), time.Now().UTC(), owner, batchID, bookID)
	if err != nil {
		return 0, err
	}
	count, err := result.RowsAffected()
	return int(count), err
}

func (s *MySQLStore) CreateComicPageImage(ctx context.Context, owner, batchID, bookID, pageID string, input CreateComicPageImageInput) (ComicPageImage, error) {
	var lastErr error
	for attempt := 1; attempt <= comicPageWriteRetryLimit; attempt++ {
		image, err := s.createComicPageImageOnce(ctx, owner, batchID, bookID, pageID, input)
		if err == nil || !retryableComicPageWriteError(err) || attempt == comicPageWriteRetryLimit {
			return image, err
		}
		lastErr = err
		if waitErr := waitComicPageWriteRetry(ctx, attempt); waitErr != nil {
			return ComicPageImage{}, waitErr
		}
	}
	return ComicPageImage{}, lastErr
}

func (s *MySQLStore) createComicPageImageOnce(ctx context.Context, owner, batchID, bookID, pageID string, input CreateComicPageImageInput) (ComicPageImage, error) {
	input, err := normalizeComicPageImageInput(input)
	if err != nil {
		return ComicPageImage{}, err
	}
	// A commit can succeed while the client sees a deadlock/connection error.
	// Reusing the same provider artifact must be idempotent, otherwise the
	// automatic retry would create duplicate candidates for one generated image.
	if strings.TrimSpace(input.StorageRef) != "" {
		var image ComicPageImage
		err := s.db.QueryRowContext(ctx, `SELECT id,COALESCE(url,''),COALESCE(storage_ref,''),media_type,source,is_primary,revision,created_at FROM batch_factory_v12_comic_page_images WHERE owner_username=? AND comic_page_id=? AND storage_ref=? LIMIT 1`, owner, pageID, input.StorageRef).Scan(&image.ID, &image.URL, &image.StorageRef, &image.MediaType, &image.Source, &image.IsPrimary, &image.Revision, &image.CreatedAt)
		if err == nil {
			image.ComicPageID = pageID
			return image, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return ComicPageImage{}, err
		}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ComicPageImage{}, err
	}
	defer tx.Rollback()
	if err := ensureComicPageOwnership(ctx, tx, owner, batchID, bookID, pageID); err != nil {
		return ComicPageImage{}, err
	}
	var primaryCount int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM batch_factory_v12_comic_page_images WHERE owner_username=? AND comic_page_id=? AND is_primary=TRUE FOR UPDATE`, owner, pageID).Scan(&primaryCount); err != nil {
		return ComicPageImage{}, err
	}
	id, err := newID("comic-image")
	if err != nil {
		return ComicPageImage{}, err
	}
	now := time.Now().UTC()
	image := ComicPageImage{ID: id, ComicPageID: pageID, URL: input.URL, StorageRef: input.StorageRef, MediaType: input.MediaType, Source: input.Source, IsPrimary: primaryCount == 0, Revision: 1, CreatedAt: now}
	if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_comic_page_images(id,owner_username,comic_page_id,url,storage_ref,media_type,source,is_primary,revision,created_at) VALUES(?,?,?,?,?,?,?,?,1,?)`, image.ID, owner, pageID, nullableString(image.URL), nullableString(image.StorageRef), image.MediaType, image.Source, image.IsPrimary, now); err != nil {
		return ComicPageImage{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_comic_pages SET status='image_ready',revision=revision+1,updated_at=? WHERE id=? AND owner_username=?`, now, pageID, owner); err != nil {
		return ComicPageImage{}, err
	}
	if err := tx.Commit(); err != nil {
		return ComicPageImage{}, err
	}
	return image, nil
}

func (s *MySQLStore) SetPrimaryComicPageImage(ctx context.Context, owner, batchID, bookID, pageID, imageID string) (ComicPageImage, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ComicPageImage{}, err
	}
	defer tx.Rollback()
	if err := ensureComicPageOwnership(ctx, tx, owner, batchID, bookID, pageID); err != nil {
		return ComicPageImage{}, err
	}
	var image ComicPageImage
	image.ID, image.ComicPageID = imageID, pageID
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(url,''),COALESCE(storage_ref,''),media_type,source,is_primary,revision,created_at FROM batch_factory_v12_comic_page_images WHERE id=? AND owner_username=? AND comic_page_id=? FOR UPDATE`, imageID, owner, pageID).Scan(&image.URL, &image.StorageRef, &image.MediaType, &image.Source, &image.IsPrimary, &image.Revision, &image.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ComicPageImage{}, ErrNotFound
	}
	if err != nil {
		return ComicPageImage{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_comic_page_images SET is_primary=FALSE,revision=revision+1 WHERE owner_username=? AND comic_page_id=? AND is_primary=TRUE AND id<>?`, owner, pageID, imageID); err != nil {
		return ComicPageImage{}, err
	}
	if !image.IsPrimary {
		if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_comic_page_images SET is_primary=TRUE,revision=revision+1 WHERE id=? AND owner_username=? AND comic_page_id=?`, imageID, owner, pageID); err != nil {
			return ComicPageImage{}, err
		}
		image.IsPrimary, image.Revision = true, image.Revision+1
	}
	if err := tx.Commit(); err != nil {
		return ComicPageImage{}, err
	}
	return image, nil
}

func (s *MySQLStore) CreateComicCollection(ctx context.Context, owner, batchID, bookID string, requested ...[]string) (ComicCollection, []string, error) {
	var lastErr error
	for attempt := 1; attempt <= comicPageWriteRetryLimit; attempt++ {
		collection, missing, err := s.createComicCollectionOnce(ctx, owner, batchID, bookID, requested...)
		if err == nil || !retryableComicPageWriteError(err) || attempt == comicPageWriteRetryLimit {
			return collection, missing, err
		}
		lastErr = err
		if waitErr := waitComicPageWriteRetry(ctx, attempt); waitErr != nil {
			return ComicCollection{}, nil, waitErr
		}
	}
	return ComicCollection{}, nil, lastErr
}

func (s *MySQLStore) createComicCollectionOnce(ctx context.Context, owner, batchID, bookID string, requested ...[]string) (ComicCollection, []string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ComicCollection{}, nil, err
	}
	defer tx.Rollback()
	if err := ensureBookOwnership(ctx, tx, owner, batchID, bookID); err != nil {
		return ComicCollection{}, nil, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id FROM batch_factory_v12_videos WHERE owner_username=? AND batch_id=? AND book_id=? AND compatibility_state='active' ORDER BY ordinal,id`, owner, batchID, bookID)
	if err != nil {
		return ComicCollection{}, nil, err
	}
	videoIDs := []string{}
	for rows.Next() {
		var videoID string
		if err := rows.Scan(&videoID); err != nil {
			rows.Close()
			return ComicCollection{}, nil, err
		}
		videoIDs = append(videoIDs, videoID)
	}
	rows.Close()
	var requestedIDs []string
	if len(requested) > 0 {
		requestedIDs = requested[0]
	}
	selectedIDs, unknown := SelectComicCollectionVideoIDsFromIDs(videoIDs, requestedIDs)
	if len(unknown) > 0 || len(selectedIDs) == 0 {
		return ComicCollection{}, unknown, ErrConflict
	}
	selectedSet := make(map[string]struct{}, len(selectedIDs))
	for _, id := range selectedIDs {
		selectedSet[id] = struct{}{}
	}
	items, missing := []ComicCollectionItem{}, []string{}
	for _, videoID := range videoIDs {
		if _, ok := selectedSet[videoID]; !ok {
			continue
		}
		var item ComicCollectionItem
		item.Ordinal, item.VideoID = len(items)+1, videoID
		err := tx.QueryRowContext(ctx, `SELECT p.id,i.id,COALESCE(i.storage_ref,''),i.media_type FROM batch_factory_v12_comic_pages p JOIN batch_factory_v12_comic_page_images i ON i.comic_page_id=p.id AND i.owner_username=p.owner_username AND i.is_primary=TRUE WHERE p.owner_username=? AND p.batch_id=? AND p.book_id=? AND p.video_id=?`, owner, batchID, bookID, videoID).Scan(&item.ComicPageID, &item.ImageID, &item.StorageRef, &item.MediaType)
		if errors.Is(err, sql.ErrNoRows) {
			missing = append(missing, videoID)
			continue
		}
		if err != nil {
			return ComicCollection{}, nil, err
		}
		items = append(items, item)
	}
	if len(missing) > 0 {
		return ComicCollection{}, missing, ErrConflict
	}
	existingRows, err := tx.QueryContext(ctx, `SELECT id,status,created_at FROM batch_factory_v12_comic_collections WHERE owner_username=? AND batch_id=? AND book_id=? ORDER BY created_at,id FOR UPDATE`, owner, batchID, bookID)
	if err != nil {
		return ComicCollection{}, nil, err
	}
	// The MySQL driver streams one result set per connection at a time. Loading
	// collection items while the FOR UPDATE rows are still open issues a nested
	// query on the same connection and fails with "driver: bad connection".
	// Drain and close the outer rows first, then load items per collection.
	type existingCollectionMeta struct {
		id        string
		status    string
		createdAt time.Time
	}
	var existingMetas []existingCollectionMeta
	for existingRows.Next() {
		var meta existingCollectionMeta
		if err := existingRows.Scan(&meta.id, &meta.status, &meta.createdAt); err != nil {
			existingRows.Close()
			return ComicCollection{}, nil, err
		}
		existingMetas = append(existingMetas, meta)
	}
	if err := existingRows.Close(); err != nil {
		return ComicCollection{}, nil, err
	}
	for _, meta := range existingMetas {
		var existing ComicCollection
		existing.ID, existing.Status, existing.CreatedAt = meta.id, meta.status, meta.createdAt
		existing.BatchID, existing.BookID = batchID, bookID
		itemRows, itemErr := tx.QueryContext(ctx, `SELECT ordinal_no,video_id,comic_page_id,image_id,COALESCE(storage_ref,''),media_type FROM batch_factory_v12_comic_collection_items WHERE collection_id=? ORDER BY ordinal_no`, existing.ID)
		if itemErr != nil {
			return ComicCollection{}, nil, itemErr
		}
		for itemRows.Next() {
			var item ComicCollectionItem
			if itemErr := itemRows.Scan(&item.Ordinal, &item.VideoID, &item.ComicPageID, &item.ImageID, &item.StorageRef, &item.MediaType); itemErr != nil {
				itemRows.Close()
				return ComicCollection{}, nil, itemErr
			}
			existing.Items = append(existing.Items, item)
		}
		if itemErr := itemRows.Close(); itemErr != nil {
			return ComicCollection{}, nil, itemErr
		}
		if sameComicCollectionItems(existing.Items, items) {
			return existing, nil, nil
		}
	}
	id, err := newID("comic-collection")
	if err != nil {
		return ComicCollection{}, nil, err
	}
	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_comic_collections(id,owner_username,batch_id,book_id,status,created_at) VALUES(?,?,?,?,?,?)`, id, owner, batchID, bookID, "ready", now); err != nil {
		return ComicCollection{}, nil, err
	}
	for _, item := range items {
		if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_comic_collection_items(collection_id,ordinal_no,video_id,comic_page_id,image_id,storage_ref,media_type) VALUES(?,?,?,?,?,?,?)`, id, item.Ordinal, item.VideoID, item.ComicPageID, item.ImageID, nullableString(item.StorageRef), item.MediaType); err != nil {
			return ComicCollection{}, nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return ComicCollection{}, nil, err
	}
	return ComicCollection{ID: id, BatchID: batchID, BookID: bookID, Status: "ready", Items: items, CreatedAt: now}, nil, nil
}

func (s *MySQLStore) ListComicCollections(ctx context.Context, owner, batchID, bookID string) ([]ComicCollection, error) {
	if err := ensureBookOwnership(ctx, s.db, owner, batchID, bookID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,status,created_at FROM batch_factory_v12_comic_collections WHERE owner_username=? AND batch_id=? AND book_id=? ORDER BY created_at,id`, owner, batchID, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	collections := []ComicCollection{}
	for rows.Next() {
		var collection ComicCollection
		collection.BatchID, collection.BookID = batchID, bookID
		if err := rows.Scan(&collection.ID, &collection.Status, &collection.CreatedAt); err != nil {
			return nil, err
		}
		collections = append(collections, collection)
	}
	for index := range collections {
		itemRows, err := s.db.QueryContext(ctx, `SELECT ordinal_no,video_id,comic_page_id,image_id,COALESCE(storage_ref,''),media_type FROM batch_factory_v12_comic_collection_items WHERE collection_id=? ORDER BY ordinal_no`, collections[index].ID)
		if err != nil {
			return nil, err
		}
		for itemRows.Next() {
			var item ComicCollectionItem
			if err := itemRows.Scan(&item.Ordinal, &item.VideoID, &item.ComicPageID, &item.ImageID, &item.StorageRef, &item.MediaType); err != nil {
				itemRows.Close()
				return nil, err
			}
			collections[index].Items = append(collections[index].Items, item)
		}
		itemRows.Close()
	}
	return latestDistinctComicCollections(collections), nil
}

var _ ComicPageStore = (*MySQLStore)(nil)
