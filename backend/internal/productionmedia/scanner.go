package productionmedia

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"qiantie/backend/internal/localartifact"
)

type Enqueuer interface {
	Enqueue(context.Context, Object) error
}

type ScanSummary struct {
	Enqueued int
	Skipped  int
}

type Scanner struct {
	DB            *sql.DB
	Repository    Enqueuer
	Local         *localartifact.Store
	Prefix        string
	RetentionDays func(string) int
}

func (s Scanner) RunOnce(ctx context.Context) (ScanSummary, error) {
	summary := ScanSummary{}
	if s.DB == nil || s.Repository == nil || s.Local == nil {
		return summary, fmt.Errorf("production media scanner is not configured")
	}
	queries := []struct {
		query string
		kind  string
	}{
		{`SELECT i.owner_username,COALESCE(i.storage_ref,''),COALESCE(i.media_type,'application/octet-stream') FROM batch_factory_v12_comic_page_images i WHERE COALESCE(i.storage_ref,'')<>''`, KindComicImage},
		{`SELECT i.owner_username,COALESCE(i.storage_ref,''),COALESCE(i.media_type,'application/octet-stream') FROM batch_factory_v12_book_asset_images i WHERE COALESCE(i.storage_ref,'')<>''`, KindAssetImage},
		{`SELECT owner_username,COALESCE(storage_ref,''),COALESCE(media_type,'application/octet-stream') FROM local_executor_artifacts WHERE COALESCE(storage_ref,'')<>''`, "local_executor"},
	}
	for _, source := range queries {
		rows, err := s.DB.QueryContext(ctx, source.query)
		if err != nil {
			return summary, err
		}
		for rows.Next() {
			var owner, storageRef, mediaType string
			if err := rows.Scan(&owner, &storageRef, &mediaType); err != nil {
				rows.Close()
				return summary, err
			}
			kind := source.kind
			if kind == "local_executor" {
				switch {
				case strings.HasPrefix(strings.ToLower(mediaType), "audio/"):
					kind = KindAudio
				case strings.HasPrefix(strings.ToLower(mediaType), "image/"):
					kind = KindComicImage
				default:
					kind = KindVideo
				}
			}
			enqueued, err := s.enqueueIfLocal(ctx, owner, kind, mediaType, storageRef)
			if err != nil {
				rows.Close()
				return summary, err
			}
			if enqueued {
				summary.Enqueued++
			} else {
				summary.Skipped++
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return summary, err
		}
		rows.Close()
	}
	rows, err := s.DB.QueryContext(ctx, `SELECT owner_username,COALESCE(output_url,'') FROM batch_factory_v12_merge_jobs WHERE status='succeeded' AND COALESCE(output_url,'')<>''`)
	if err != nil {
		return summary, err
	}
	defer rows.Close()
	for rows.Next() {
		var owner, outputURL string
		if err := rows.Scan(&owner, &outputURL); err != nil {
			return summary, err
		}
		storageRef, ok := localMergeStorageRef(outputURL)
		if !ok {
			summary.Skipped++
			continue
		}
		enqueued, err := s.enqueueIfLocal(ctx, owner, KindMergeVideo, "video/mp4", storageRef)
		if err != nil {
			return summary, err
		}
		if enqueued {
			summary.Enqueued++
		} else {
			summary.Skipped++
		}
	}
	return summary, rows.Err()
}

func (s Scanner) enqueueIfLocal(ctx context.Context, owner, kind, mediaType, storageRef string) (bool, error) {
	if _, err := s.Local.LocalPath(storageRef); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	key, err := ObjectKey(s.Prefix, kind, storageRef)
	if err != nil {
		return false, err
	}
	days := 7
	if s.RetentionDays != nil {
		days = s.RetentionDays(owner)
	}
	if err := s.Repository.Enqueue(ctx, Object{Owner: owner, MediaKind: kind, MediaType: mediaType, LocalStorageRef: storageRef, ObjectKey: key, RetentionDays: days}); err != nil {
		return false, err
	}
	return true, nil
}

func localMergeStorageRef(raw string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", false
	}
	value := path.Base(parsed.Path)
	if value == "." || value == "/" || !strings.Contains(parsed.Path, "/merge-media/") || !strings.HasPrefix(value, "merge_") {
		return "", false
	}
	if !strings.HasSuffix(value, ".mp4") {
		value += ".mp4"
	}
	return value, true
}

func (s Scanner) Start(interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}
	go func() {
		run := func() {
			if _, err := s.RunOnce(context.Background()); err != nil {
				log.Printf("production media scanner failed: %v", err)
			}
		}
		run()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			run()
		}
	}()
}
