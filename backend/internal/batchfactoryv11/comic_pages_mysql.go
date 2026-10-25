package batchfactoryv11

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

func ensureComicPageOwnership(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, owner, batchID, bookID, pageID string) error {
	var one int
	err := q.QueryRowContext(ctx, `SELECT 1 FROM batch_factory_v11_comic_pages WHERE id=? AND owner_username=? AND batch_id=? AND book_id=?`, pageID, owner, batchID, bookID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func ensureComicVideoOwnership(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, owner, batchID, bookID, videoID string) error {
	var one int
	err := q.QueryRowContext(ctx, `SELECT 1 FROM batch_factory_v11_videos WHERE id=? AND owner_username=? AND batch_id=? AND book_id=? AND compatibility_state='active'`, videoID, owner, batchID, bookID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (s *MySQLStore) UpsertComicPage(ctx context.Context, owner, batchID, bookID, videoID string, input UpsertComicPageInput) (ComicPage, error) {
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
	err = tx.QueryRowContext(ctx, `SELECT id,revision,created_at FROM batch_factory_v11_comic_pages WHERE owner_username=? AND batch_id=? AND book_id=? AND video_id=? FOR UPDATE`, owner, batchID, bookID, videoID).Scan(&page.ID, &page.Revision, &page.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		page.ID, err = newID("comic-page")
		if err != nil {
			return ComicPage{}, err
		}
		page.Revision, page.CreatedAt = 1, now
		_, err = tx.ExecContext(ctx, `INSERT INTO batch_factory_v11_comic_pages(id,owner_username,batch_id,book_id,video_id,comic_prompt,image_prompt,model_id,aspect_ratio,status,snapshot_hash,revision,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,1,?,?)`, page.ID, owner, batchID, bookID, videoID, input.ComicPrompt, input.ImagePrompt, nullableString(input.ModelID), input.AspectRatio, input.Status, nullableString(input.SnapshotHash), now, now)
	} else if err == nil {
		page.Revision++
		_, err = tx.ExecContext(ctx, `UPDATE batch_factory_v11_comic_pages SET comic_prompt=?,image_prompt=?,model_id=?,aspect_ratio=?,status=?,snapshot_hash=?,revision=?,updated_at=? WHERE id=? AND owner_username=?`, input.ComicPrompt, input.ImagePrompt, nullableString(input.ModelID), input.AspectRatio, input.Status, nullableString(input.SnapshotHash), page.Revision, now, page.ID, owner)
	}
	if err != nil {
		return ComicPage{}, err
	}
	if err := tx.Commit(); err != nil {
		return ComicPage{}, err
	}
	page.BatchID, page.BookID, page.VideoID = batchID, bookID, videoID
	page.ComicPrompt, page.ImagePrompt, page.ModelID = input.ComicPrompt, input.ImagePrompt, input.ModelID
	page.AspectRatio, page.Status, page.SnapshotHash, page.UpdatedAt = input.AspectRatio, input.Status, input.SnapshotHash, now
	return page, nil
}

func (s *MySQLStore) ListComicPages(ctx context.Context, owner, batchID, bookID string) ([]ComicPage, error) {
	if err := ensureBookOwnership(ctx, s.db, owner, batchID, bookID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT p.id,p.video_id,p.comic_prompt,p.image_prompt,COALESCE(p.model_id,''),p.aspect_ratio,p.status,COALESCE(p.snapshot_hash,''),p.revision,p.created_at,p.updated_at FROM batch_factory_v11_comic_pages p JOIN batch_factory_v11_videos v ON v.id=p.video_id WHERE p.owner_username=? AND p.batch_id=? AND p.book_id=? AND v.compatibility_state='active' ORDER BY v.ordinal,v.id`, owner, batchID, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	pages := []ComicPage{}
	for rows.Next() {
		var page ComicPage
		page.BatchID, page.BookID = batchID, bookID
		if err := rows.Scan(&page.ID, &page.VideoID, &page.ComicPrompt, &page.ImagePrompt, &page.ModelID, &page.AspectRatio, &page.Status, &page.SnapshotHash, &page.Revision, &page.CreatedAt, &page.UpdatedAt); err != nil {
			return nil, err
		}
		pages = append(pages, page)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for index := range pages {
		imageRows, err := s.db.QueryContext(ctx, `SELECT id,COALESCE(url,''),COALESCE(storage_ref,''),media_type,source,is_primary,revision,created_at FROM batch_factory_v11_comic_page_images WHERE owner_username=? AND comic_page_id=? ORDER BY created_at,id`, owner, pages[index].ID)
		if err != nil {
			return nil, err
		}
		for imageRows.Next() {
			var image ComicPageImage
			image.ComicPageID = pages[index].ID
			if err := imageRows.Scan(&image.ID, &image.URL, &image.StorageRef, &image.MediaType, &image.Source, &image.IsPrimary, &image.Revision, &image.CreatedAt); err != nil {
				imageRows.Close()
				return nil, err
			}
			pages[index].Images = append(pages[index].Images, image)
		}
		if err := imageRows.Close(); err != nil {
			return nil, err
		}
	}
	return pages, nil
}

func (s *MySQLStore) CreateComicPageImage(ctx context.Context, owner, batchID, bookID, pageID string, input CreateComicPageImageInput) (ComicPageImage, error) {
	input, err := normalizeComicPageImageInput(input)
	if err != nil {
		return ComicPageImage{}, err
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
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM batch_factory_v11_comic_page_images WHERE owner_username=? AND comic_page_id=? AND is_primary=TRUE FOR UPDATE`, owner, pageID).Scan(&primaryCount); err != nil {
		return ComicPageImage{}, err
	}
	id, err := newID("comic-image")
	if err != nil {
		return ComicPageImage{}, err
	}
	now := time.Now().UTC()
	image := ComicPageImage{ID: id, ComicPageID: pageID, URL: input.URL, StorageRef: input.StorageRef, MediaType: input.MediaType, Source: input.Source, IsPrimary: primaryCount == 0, Revision: 1, CreatedAt: now}
	if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v11_comic_page_images(id,owner_username,comic_page_id,url,storage_ref,media_type,source,is_primary,revision,created_at) VALUES(?,?,?,?,?,?,?,?,1,?)`, image.ID, owner, pageID, nullableString(image.URL), nullableString(image.StorageRef), image.MediaType, image.Source, image.IsPrimary, now); err != nil {
		return ComicPageImage{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v11_comic_pages SET status='image_ready',revision=revision+1,updated_at=? WHERE id=? AND owner_username=?`, now, pageID, owner); err != nil {
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
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(url,''),COALESCE(storage_ref,''),media_type,source,is_primary,revision,created_at FROM batch_factory_v11_comic_page_images WHERE id=? AND owner_username=? AND comic_page_id=? FOR UPDATE`, imageID, owner, pageID).Scan(&image.URL, &image.StorageRef, &image.MediaType, &image.Source, &image.IsPrimary, &image.Revision, &image.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ComicPageImage{}, ErrNotFound
	}
	if err != nil {
		return ComicPageImage{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v11_comic_page_images SET is_primary=FALSE,revision=revision+1 WHERE owner_username=? AND comic_page_id=? AND is_primary=TRUE AND id<>?`, owner, pageID, imageID); err != nil {
		return ComicPageImage{}, err
	}
	if !image.IsPrimary {
		if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v11_comic_page_images SET is_primary=TRUE,revision=revision+1 WHERE id=? AND owner_username=? AND comic_page_id=?`, imageID, owner, pageID); err != nil {
			return ComicPageImage{}, err
		}
		image.IsPrimary, image.Revision = true, image.Revision+1
	}
	if err := tx.Commit(); err != nil {
		return ComicPageImage{}, err
	}
	return image, nil
}

func (s *MySQLStore) CreateComicCollection(ctx context.Context, owner, batchID, bookID string) (ComicCollection, []string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ComicCollection{}, nil, err
	}
	defer tx.Rollback()
	if err := ensureBookOwnership(ctx, tx, owner, batchID, bookID); err != nil {
		return ComicCollection{}, nil, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id FROM batch_factory_v11_videos WHERE owner_username=? AND batch_id=? AND book_id=? AND compatibility_state='active' ORDER BY ordinal,id`, owner, batchID, bookID)
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
	items, missing := []ComicCollectionItem{}, []string{}
	for ordinal, videoID := range videoIDs {
		var item ComicCollectionItem
		item.Ordinal, item.VideoID = ordinal+1, videoID
		err := tx.QueryRowContext(ctx, `SELECT p.id,i.id,COALESCE(i.storage_ref,''),i.media_type FROM batch_factory_v11_comic_pages p JOIN batch_factory_v11_comic_page_images i ON i.comic_page_id=p.id AND i.owner_username=p.owner_username AND i.is_primary=TRUE WHERE p.owner_username=? AND p.batch_id=? AND p.book_id=? AND p.video_id=?`, owner, batchID, bookID, videoID).Scan(&item.ComicPageID, &item.ImageID, &item.StorageRef, &item.MediaType)
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
	id, err := newID("comic-collection")
	if err != nil {
		return ComicCollection{}, nil, err
	}
	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v11_comic_collections(id,owner_username,batch_id,book_id,status,created_at) VALUES(?,?,?,?,?,?)`, id, owner, batchID, bookID, "ready", now); err != nil {
		return ComicCollection{}, nil, err
	}
	for _, item := range items {
		if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v11_comic_collection_items(collection_id,ordinal_no,video_id,comic_page_id,image_id,storage_ref,media_type) VALUES(?,?,?,?,?,?,?)`, id, item.Ordinal, item.VideoID, item.ComicPageID, item.ImageID, nullableString(item.StorageRef), item.MediaType); err != nil {
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
	rows, err := s.db.QueryContext(ctx, `SELECT id,status,created_at FROM batch_factory_v11_comic_collections WHERE owner_username=? AND batch_id=? AND book_id=? ORDER BY created_at,id`, owner, batchID, bookID)
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
		itemRows, err := s.db.QueryContext(ctx, `SELECT ordinal_no,video_id,comic_page_id,image_id,COALESCE(storage_ref,''),media_type FROM batch_factory_v11_comic_collection_items WHERE collection_id=? ORDER BY ordinal_no`, collections[index].ID)
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
	return collections, nil
}

var _ ComicPageStore = (*MySQLStore)(nil)
