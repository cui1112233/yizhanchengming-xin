package doubaoexecutor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestDownloadReferenceSupportsDataAndHTTPImages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("http-image"))
	}))
	defer server.Close()

	for index, value := range []string{"data:image/png;base64,ZGF0YS1pbWFnZQ==", server.URL + "/image.png"} {
		path, err := downloadReference(context.Background(), value, index)
		if err != nil {
			t.Fatalf("downloadReference(%q): %v", value, err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read downloaded image: %v", err)
		}
		want := []byte("data-image")
		if index == 1 {
			want = []byte("http-image")
		}
		if string(data) != string(want) {
			t.Fatalf("image data = %q, want %q", data, want)
		}
		if err := os.Remove(path); err != nil {
			t.Fatalf("remove temporary image: %v", err)
		}
	}
}

func TestAsStringsNormalizesPayloadShapes(t *testing.T) {
	if got := asStrings([]any{" one ", "", "two"}); len(got) != 2 || got[0] != "one" || got[1] != "two" {
		t.Fatalf("asStrings([]any) = %#v", got)
	}
	if got := asStrings([]string{" one ", ""}); len(got) != 1 || got[0] != "one" {
		t.Fatalf("asStrings([]string) = %#v", got)
	}
}
