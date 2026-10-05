package video

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

type recordingObjectUploader struct {
	bucket string
	key    string
	body   string
}

func (u *recordingObjectUploader) PutObjectFromFile(_ context.Context, bucket, key, filename string) error {
	data, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		return err
	}
	u.bucket, u.key, u.body = bucket, key, string(data)
	return nil
}

func TestTOSArtifactStoreDownloadsThenUploadsProviderArtifact(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		_, _ = io.WriteString(w, "fake-mp4-bytes")
	}))
	defer server.Close()

	uploader := &recordingObjectUploader{}
	store, err := NewArtifactStore(ArtifactStoreConfig{
		Bucket: "video-bucket", PublicBaseURL: "https://cdn.example/video-bucket/",
		HTTPClient: server.Client(), Uploader: uploader, AllowInsecureLoopback: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := store.Persist(context.Background(), server.URL+"/provider.mp4", "task14/51/31/out.mp4")
	if err != nil {
		t.Fatal(err)
	}
	if uploader.bucket != "video-bucket" || uploader.key != "task14/51/31/out.mp4" || uploader.body != "fake-mp4-bytes" {
		t.Fatalf("upload = bucket=%q key=%q body=%q", uploader.bucket, uploader.key, uploader.body)
	}
	if artifact.Bucket != "video-bucket" || artifact.ObjectKey != "task14/51/31/out.mp4" || artifact.URL != "https://cdn.example/video-bucket/task14/51/31/out.mp4" {
		t.Fatalf("artifact = %+v", artifact)
	}
}

func TestTOSArtifactStoreRejectsUnsafeSourceURL(t *testing.T) {
	store, err := NewArtifactStore(ArtifactStoreConfig{Bucket: "video-bucket", PublicBaseURL: "https://cdn.example/", Uploader: &recordingObjectUploader{}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Persist(context.Background(), "http://provider.example/out.mp4", "out.mp4"); err == nil {
		t.Fatal("expected non-https provider artifact URL rejection")
	}
}
