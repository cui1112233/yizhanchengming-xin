package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLimitRequestBodyContentLengthTooLarge(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/test", strings.NewReader("hello"))
	request.ContentLength = 100
	if limitRequestBody(recorder, request, 10) {
		t.Fatal("expected request to be rejected")
	}
	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", recorder.Code)
	}
	var payload map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !strings.Contains(payload["error"], "请求体超过上限") {
		t.Fatalf("unexpected message: %q", payload["error"])
	}
}

func TestLimitRequestBodyAllowsWithinLimit(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/test", strings.NewReader("hello"))
	if !limitRequestBody(recorder, request, 10) {
		t.Fatal("expected request to be allowed")
	}
}

func TestMaxBytesReaderErrorDetection(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/test", strings.NewReader("0123456789"))
	request.Body = http.MaxBytesReader(recorder, request.Body, 4)
	buffer := make([]byte, 16)
	_, err := request.Body.Read(buffer)
	if !isBodyTooLarge(err) {
		t.Fatalf("expected MaxBytesError, got %v", err)
	}
}

func TestFormatByteLimit(t *testing.T) {
	cases := map[int64]string{
		16 << 20: "16MB",
		25 << 20: "25MB",
		2 << 30:  "2GB",
		5 << 30:  "5GB",
		2048:     "2KB",
	}
	for limit, want := range cases {
		if got := formatByteLimit(limit); got != want {
			t.Fatalf("formatByteLimit(%d) = %s, want %s", limit, got, want)
		}
	}
}
