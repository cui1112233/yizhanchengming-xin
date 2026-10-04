package media

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

type uploaderFake struct {
	bucket string
	key string
	contentType string
	contentLength int64
	body string
	err error
}

func (f *uploaderFake) Put(_ context.Context, key, contentType string, contentLength int64, body io.Reader) (string, string, error) {
	f.key, f.contentType, f.contentLength = key, contentType, contentLength
	if body != nil { data, _ := io.ReadAll(body); f.body = string(data) }
	if f.err != nil { return "", "", f.err }
	return f.bucket, key, nil
}

type assetWriterFake struct {
	input CreateInput
	asset Asset
	err error
	calls int
}

func (f *assetWriterFake) Create(_ context.Context, input CreateInput) (Asset, error) {
	f.calls++
	f.input = input
	if f.err != nil { return Asset{}, f.err }
	if f.asset.ID == "" { f.asset = Asset{ID: "asset_1", Owner: input.Owner, MediaType: input.MediaType, TOSBucket: input.TOSBucket, TOSKey: input.TOSKey, MimeType: input.MimeType} }
	return f.asset, nil
}

func TestIngestUploadsToTOSBeforeCreatingAsset(t *testing.T) {
	uploader := &uploaderFake{bucket: "prod-media"}
	writer := &assetWriterFake{}
	service := IngestService{Uploader: uploader, Assets: writer}

	asset, err := service.Ingest(context.Background(), IngestInput{
		Owner: "owner-1",
		MediaType: TypeImage,
		ContentType: "image/png",
		ContentLength: 3,
		Body: strings.NewReader("png"),
		SourceTaskID: "task_123",
		ObjectKey: "images/2026/10/a.png",
	})
	if err != nil { t.Fatal(err) }
	if uploader.key != "images/2026/10/a.png" || uploader.body != "png" { t.Fatalf("upload=%#v", uploader) }
	if writer.calls != 1 { t.Fatalf("asset creates=%d", writer.calls) }
	if writer.input.TOSBucket != "prod-media" || writer.input.TOSKey != uploader.key || writer.input.SourceTaskID != "task_123" { t.Fatalf("asset input=%#v", writer.input) }
	if asset.ID != "asset_1" { t.Fatalf("asset=%#v", asset) }
}

func TestIngestNeverCreatesDatabaseAssetWhenTOSUploadFails(t *testing.T) {
	uploader := &uploaderFake{bucket: "prod-media", err: errors.New("tos unavailable")}
	writer := &assetWriterFake{}
	service := IngestService{Uploader: uploader, Assets: writer}
	_, err := service.Ingest(context.Background(), IngestInput{Owner: "owner", MediaType: TypeVideo, ContentType: "video/mp4", Body: strings.NewReader("video")})
	if err == nil { t.Fatal("expected upload error") }
	if writer.calls != 0 { t.Fatalf("database asset must not exist before TOS upload; calls=%d", writer.calls) }
}

func TestIngestGeneratesNamespacedObjectKeyWhenMissing(t *testing.T) {
	uploader := &uploaderFake{bucket: "prod-media"}
	writer := &assetWriterFake{}
	service := IngestService{Uploader: uploader, Assets: writer}
	_, err := service.Ingest(context.Background(), IngestInput{Owner: "user@example.com", MediaType: TypeVideo, ContentType: "video/mp4", Body: strings.NewReader("video")})
	if err != nil { t.Fatal(err) }
	if !strings.HasPrefix(uploader.key, "video/") || !strings.HasSuffix(uploader.key, ".mp4") { t.Fatalf("generated key=%q", uploader.key) }
	if strings.Contains(uploader.key, "@") { t.Fatalf("generated key must not leak owner identifier: %q", uploader.key) }
}

func TestIngestRejectsUnsupportedMimeForMediaType(t *testing.T) {
	service := IngestService{Uploader: &uploaderFake{}, Assets: &assetWriterFake{}}
	if _, err := service.Ingest(context.Background(), IngestInput{Owner: "owner", MediaType: TypeImage, ContentType: "video/mp4", Body: strings.NewReader("x")}); err == nil {
		t.Fatal("expected media type/mime mismatch")
	}
}
