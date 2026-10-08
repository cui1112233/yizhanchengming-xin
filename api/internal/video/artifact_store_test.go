package video

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/objectkey"
)

type recordingObjectUploader struct {
	bucket string
	key    string
	body   string
	calls  int
}

func (u *recordingObjectUploader) PutObjectFromFile(_ context.Context, bucket, key, filename string) error {
	u.calls++
	data, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		return err
	}
	u.bucket, u.key, u.body = bucket, key, string(data)
	return nil
}

type artifactRoundTripper func(*http.Request) (*http.Response, error)

func (f artifactRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestTOSArtifactStoreKeyPrefix(t *testing.T) {
	for _, raw := range []string{"", "staging/", "staging..local/"} {
		for _, method := range []string{"download", "file"} {
			t.Run(raw+method, func(t *testing.T) {
				prefix, err := objectkey.ParsePrefix(raw)
				if err != nil {
					t.Fatal(err)
				}
				uploader, requests := &recordingObjectUploader{}, 0
				client := &http.Client{Transport: artifactRoundTripper(func(*http.Request) (*http.Response, error) {
					requests++
					return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("fake-mp4"))}, nil
				})}
				store, err := NewArtifactStore(ArtifactStoreConfig{Bucket: "private", PublicBaseURL: "https://cdn.example/", Uploader: uploader, HTTPClient: client, KeyPrefix: prefix})
				if err != nil {
					t.Fatal(err)
				}
				var artifact Artifact
				want := raw + "video/7/9.mp4"
				if method == "download" {
					artifact, err = store.Persist(context.Background(), "https://provider.example/output", "video/7/9.mp4")
					if requests != 1 {
						t.Fatalf("downloads=%d", requests)
					}
				} else {
					f, e := os.CreateTemp(t.TempDir(), "merge-*.mp4")
					if e != nil {
						t.Fatal(e)
					}
					if _, e = io.WriteString(f, "fake-mp4"); e != nil {
						t.Fatal(e)
					}
					if e = f.Close(); e != nil {
						t.Fatal(e)
					}
					artifact, err = store.PersistFile(context.Background(), f.Name(), "merge/7/2.mp4")
					want = raw + "merge/7/2.mp4"
					if requests != 0 {
						t.Fatalf("downloads=%d", requests)
					}
				}
				if err != nil {
					t.Fatal(err)
				}
				if uploader.calls != 1 || uploader.key != want || artifact.ObjectKey != want || artifact.URL != "https://cdn.example/"+want || uploader.body != "fake-mp4" {
					t.Fatalf("upload=%+v artifact=%+v", uploader, artifact)
				}
			})
		}
	}
}

func TestTOSArtifactStoreKeyPrefixRejectsBeforeIO(t *testing.T) {
	prefix, err := objectkey.ParsePrefix("staging/")
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"/video/x", "video//x", "video/../x", `video\x`, "video/%2f/x", "staging/video/x", "staging/staging/video/x"} {
		t.Run(key, func(t *testing.T) {
			uploader, requests := &recordingObjectUploader{}, 0
			client := &http.Client{Transport: artifactRoundTripper(func(*http.Request) (*http.Response, error) { requests++; return nil, nil })}
			store, err := NewArtifactStore(ArtifactStoreConfig{Bucket: "private", PublicBaseURL: "https://cdn.example", Uploader: uploader, HTTPClient: client, KeyPrefix: prefix})
			if err != nil {
				t.Fatal(err)
			}
			if _, err = store.Persist(context.Background(), "https://provider.example/output", key); err == nil {
				t.Fatal("invalid download key accepted")
			}
			if _, err = store.PersistFile(context.Background(), "does-not-exist", key); err == nil || strings.Contains(err.Error(), "stat") {
				t.Fatalf("file key not rejected before stat: %v", err)
			}
			if requests != 0 || uploader.calls != 0 {
				t.Fatalf("downloads=%d uploads=%d", requests, uploader.calls)
			}
		})
	}
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
