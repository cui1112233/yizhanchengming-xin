package media

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

type execFake struct {
	query string
	args  []any
	rows  int64
}

func (f *execFake) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	f.query = query
	f.args = append([]any(nil), args...)
	rows := f.rows
	if rows == 0 { rows = 1 }
	return mediaResult(rows), nil
}

type mediaResult int64
func (r mediaResult) LastInsertId() (int64, error) { return int64(r), nil }
func (r mediaResult) RowsAffected() (int64, error) { return int64(r), nil }

func TestValidateCreateInputRequiresCanonicalTOSObject(t *testing.T) {
	valid := CreateInput{Owner: "owner", MediaType: TypeImage, TOSBucket: "bucket", TOSKey: "images/a.png"}
	if err := validateCreateInput(valid); err != nil { t.Fatal(err) }
	for _, tc := range []CreateInput{
		{MediaType: TypeImage, TOSBucket: "bucket", TOSKey: "images/a.png"},
		{Owner: "owner", MediaType: "audio", TOSBucket: "bucket", TOSKey: "a.mp3"},
		{Owner: "owner", MediaType: TypeImage, TOSKey: "images/a.png"},
		{Owner: "owner", MediaType: TypeVideo, TOSBucket: "bucket"},
	} {
		if err := validateCreateInput(tc); err == nil { t.Fatalf("expected invalid input: %#v", tc) }
	}
}

func TestSQLStoreCreatePersistsOnlyTOSLocatorAndMetadata(t *testing.T) {
	exec := &execFake{}
	store := newSQLStoreWithExecutor(exec)
	asset, err := store.Create(context.Background(), CreateInput{
		Owner: "owner-1",
		MediaType: TypeVideo,
		TOSBucket: "prod-media",
		TOSKey: "video/2026/10/result.mp4",
		MimeType: "video/mp4",
		SizeBytes: 1024,
		WidthPx: 1080,
		HeightPx: 1920,
		DurationMs: 15000,
		SourceTaskID: "task-1",
	})
	if err != nil { t.Fatal(err) }
	if asset.ID == "" || asset.TOSKey != "video/2026/10/result.mp4" { t.Fatalf("asset=%#v", asset) }
	if !strings.Contains(exec.query, "INSERT INTO media_assets") { t.Fatalf("query=%s", exec.query) }
	if strings.Contains(exec.query, "local_path") || strings.Contains(exec.query, "file_path") { t.Fatalf("query contains local path: %s", exec.query) }
	if got := exec.args[4]; got != "video/mp4" { t.Fatalf("mime arg=%v", got) }
}

func TestAssetReferenceNeverExposesCredentials(t *testing.T) {
	asset := Asset{ID: "asset_1", Owner: "owner", MediaType: TypeImage, TOSBucket: "bucket", TOSKey: "images/a.png"}
	ref := asset.Reference()
	if ref.ID != "asset_1" || ref.MediaType != TypeImage { t.Fatalf("ref=%#v", ref) }
	if strings.Contains(ref.StorageKey, "AK") || strings.Contains(ref.StorageKey, "secret") { t.Fatalf("unsafe ref=%#v", ref) }
}
