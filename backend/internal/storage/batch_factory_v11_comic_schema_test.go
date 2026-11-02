package storage

import (
	"strings"
	"testing"
)

func TestV11ComicPageMigrationIsAppendOnlyAfterBookAssetOverrides(t *testing.T) {
	var historical, comic, snapshot Migration
	for _, candidate := range V11Migrations() {
		if candidate.Version == 1100028 {
			historical = candidate
		}
		if candidate.Version == 1100029 {
			comic = candidate
		}
		if candidate.Version == 1100030 {
			snapshot = candidate
		}
	}
	ddl := strings.ToLower(strings.Join(comic.SQL, "\n"))
	if historical.Version != 1100028 || !checksumMatches(historical, "402a7b76a317be5a5a753a64e179f6bee5d426b2382ba6b1fd8ad3f605e904c2", ChecksumFor(historical)) {
		t.Fatalf("historical comic migration does not accept the verified local ledger: %+v", historical)
	}
	if comic.Version != 1100029 || comic.CallbackChecksum != "batch-factory-v11-comic-pages-repair-v1" {
		t.Fatalf("comic migration=%+v", comic)
	}
	for _, table := range []string{"batch_factory_v11_comic_pages", "batch_factory_v11_comic_page_images", "batch_factory_v11_comic_collections", "batch_factory_v11_comic_collection_items"} {
		if !strings.Contains(ddl, table) {
			t.Fatalf("migration missing %s", table)
		}
	}
	if snapshot.Version != 1100030 || !strings.Contains(strings.ToLower(strings.Join(snapshot.SQL, "\n")), "snapshot_hash") {
		t.Fatalf("comic snapshot migration=%+v", snapshot)
	}
}

func TestV11ComicSourceSegmentMigrationIsAppendOnly(t *testing.T) {
	var migration Migration
	for _, candidate := range V11Migrations() {
		if candidate.Version == 1100032 {
			migration = candidate
			break
		}
	}
	if migration.Version != 1100032 || migration.CallbackChecksum != "batch-factory-v11-comic-source-segments-v1" {
		t.Fatalf("comic source migration=%+v", migration)
	}
	ddl := strings.ToLower(strings.Join(migration.SQL, "\n"))
	for _, fragment := range []string{
		"batch_factory_v11_comic_source_segments", "video_id", "start_offset", "end_offset",
		"source_text", "character_count", "source_digest", "prompt_stale", "media_stale",
		"revision", "foreign key (video_id)", "idx_bfv11_comic_source_book",
	} {
		if !strings.Contains(ddl, fragment) {
			t.Fatalf("comic source migration missing %q", fragment)
		}
	}
}
