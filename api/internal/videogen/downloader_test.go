package videogen

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPDownloaderReturnsVideoStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Length", "5")
		_, _ = w.Write([]byte("video"))
	}))
	defer server.Close()

	d := HTTPDownloader{Client: server.Client(), ValidateURL: func(string) error { return nil }, MaxBytes: 1024}
	result, err := d.Download(context.Background(), server.URL+"/result.mp4")
	if err != nil { t.Fatal(err) }
	defer result.Body.Close()
	body, _ := io.ReadAll(result.Body)
	if string(body) != "video" || result.ContentType != "video/mp4" || result.ContentLength != 5 {
		t.Fatalf("result=%#v body=%q", result, body)
	}
}

func TestHTTPDownloaderRejectsNonVideoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("no"))
	}))
	defer server.Close()
	d := HTTPDownloader{Client: server.Client(), ValidateURL: func(string) error { return nil }}
	if _, err := d.Download(context.Background(), server.URL); err == nil { t.Fatal("expected content type validation") }
}

func TestHTTPDownloaderRejectsUnsafeURLBeforeRequest(t *testing.T) {
	d := HTTPDownloader{ValidateURL: func(raw string) error {
		if strings.Contains(raw, "localhost") { return io.ErrUnexpectedEOF }
		return nil
	}}
	if _, err := d.Download(context.Background(), "http://localhost/video.mp4"); err == nil { t.Fatal("expected URL validation") }
}
