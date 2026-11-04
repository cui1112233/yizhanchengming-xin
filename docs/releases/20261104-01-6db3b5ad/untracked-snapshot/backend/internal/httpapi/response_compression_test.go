package httpapi

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func runCompressedRequest(t *testing.T, handler http.HandlerFunc, acceptEncoding string, extraHeaders map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	if acceptEncoding != "" {
		request.Header.Set("Accept-Encoding", acceptEncoding)
	}
	for key, value := range extraHeaders {
		request.Header.Set(key, value)
	}
	recorder := httptest.NewRecorder()
	withJSONCompression(handler).ServeHTTP(recorder, request)
	return recorder
}

func TestJSONCompressionCompressesLargeJSON(t *testing.T) {
	payload := strings.Repeat(`{"key":"value"},`, 200)
	recorder := runCompressedRequest(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = io.WriteString(w, payload)
	}, "gzip, deflate", nil)
	if recorder.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("Content-Encoding=%q", recorder.Header().Get("Content-Encoding"))
	}
	if recorder.Header().Get("Vary") != "Accept-Encoding" {
		t.Fatalf("Vary=%q", recorder.Header().Get("Vary"))
	}
	reader, err := gzip.NewReader(recorder.Body)
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	decoded, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if string(decoded) != payload {
		t.Fatalf("decoded body mismatch: %d bytes", len(decoded))
	}
	if recorder.Body.Len() >= len(payload) {
		t.Fatalf("compressed body=%d, plain=%d", recorder.Body.Len(), len(payload))
	}
}

func TestJSONCompressionKeepsSmallJSONPlain(t *testing.T) {
	recorder := runCompressedRequest(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true}`)
	}, "gzip", nil)
	if recorder.Header().Get("Content-Encoding") != "" {
		t.Fatalf("small response was compressed: %q", recorder.Header().Get("Content-Encoding"))
	}
	if recorder.Body.String() != `{"ok":true}` {
		t.Fatalf("body=%q", recorder.Body.String())
	}
}

func TestJSONCompressionSkipsRangeAndNonJSON(t *testing.T) {
	recorder := runCompressedRequest(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = io.WriteString(w, strings.Repeat("png", 4096))
	}, "gzip", map[string]string{"Range": "bytes=0-10"})
	if recorder.Code != http.StatusPartialContent {
		t.Fatalf("status=%d", recorder.Code)
	}
	if recorder.Header().Get("Content-Encoding") != "" {
		t.Fatalf("range response was compressed")
	}
	if recorder.Body.Len() != 4096*3 {
		t.Fatalf("range body=%d", recorder.Body.Len())
	}

	recorder = runCompressedRequest(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = io.WriteString(w, strings.Repeat("png", 4096))
	}, "gzip", nil)
	if recorder.Header().Get("Content-Encoding") != "" {
		t.Fatalf("image response was compressed")
	}
	if recorder.Body.Len() != 4096*3 {
		t.Fatalf("image body=%d", recorder.Body.Len())
	}
}

func TestJSONCompressionKeepsNotModifiedPlain(t *testing.T) {
	recorder := runCompressedRequest(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotModified)
	}, "gzip", nil)
	if recorder.Code != http.StatusNotModified || recorder.Header().Get("Content-Encoding") != "" {
		t.Fatalf("status=%d encoding=%q", recorder.Code, recorder.Header().Get("Content-Encoding"))
	}
}

func TestJSONCompressionIgnoresMissingAcceptEncoding(t *testing.T) {
	payload := strings.Repeat(`{"key":"value"},`, 200)
	recorder := runCompressedRequest(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, payload)
	}, "", nil)
	if recorder.Header().Get("Content-Encoding") != "" || recorder.Body.Len() != len(payload) {
		t.Fatalf("encoding=%q body=%d", recorder.Header().Get("Content-Encoding"), recorder.Body.Len())
	}
}
