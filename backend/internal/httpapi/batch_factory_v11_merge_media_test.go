package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
)

func TestSliceFiveMergeMediaSupportsRangeRequests(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "b", Books: []batchfactoryv11.CreateBookInput{{Title: "k", SourceText: "正文"}}})
	if err != nil {
		t.Fatal(err)
	}
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	artifactID := "merge_range_media_1"
	payload := "0000ftypisom-local-merged-video" + strings.Repeat("0123456789", 100)
	if _, err := files.SaveMP4(artifactID, strings.NewReader(payload)); err != nil {
		t.Fatal(err)
	}
	output := "/api/batch-factory/v12/batches/" + batch.ID + "/merge-media/" + artifactID
	if _, err := store.CreateMergeJob(context.Background(), batchfactoryv11.MergeJob{Owner: "alice", BatchID: batch.ID, RequestID: "range-media", Status: batchfactoryv11.MergeSucceeded, OutputURL: output, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 5, Store: store, Merge: &batchfactoryv11.MergeService{Store: store, Enabled: true}, LocalArtifacts: files})

	// 普通 GET 必须带 Content-Length 与 Accept-Ranges，否则浏览器会把直出当成不可拖动的流。
	full := signedJSONRequest(t, api, now, "alice", http.MethodGet, output, nil)
	if full.Code != http.StatusOK || full.Body.Len() != len(payload) {
		t.Fatalf("full status=%d bytes=%d", full.Code, full.Body.Len())
	}
	if got := full.Header().Get("Accept-Ranges"); got != "bytes" {
		t.Fatalf("Accept-Ranges=%q", got)
	}
	if got := full.Header().Get("Content-Length"); got != strconv.Itoa(len(payload)) {
		t.Fatalf("Content-Length=%q want %d", got, len(payload))
	}

	// Range 请求必须返回 206 + Content-Range，这是播放器起播的关键。
	req := httptest.NewRequest(http.MethodGet, output, nil)
	req.Header.Set("Range", "bytes=0-99")
	SignBridgeRequest(req, "alice", false, now, "secret")
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusPartialContent {
		t.Fatalf("range status=%d body=%s", rec.Code, rec.Body.String())
	}
	if got, want := rec.Header().Get("Content-Range"), "bytes 0-99/"+strconv.Itoa(len(payload)); got != want {
		t.Fatalf("Content-Range=%q want %q", got, want)
	}
	if rec.Body.Len() != 100 || rec.Body.String() != payload[:100] {
		t.Fatalf("range bytes=%d body=%q", rec.Body.Len(), rec.Body.String())
	}

	// 越界范围必须明确返回 416，并给出总长度，不能让播放器拿到整包。
	beyond := httptest.NewRequest(http.MethodGet, output, nil)
	beyond.Header.Set("Range", "bytes=2000-3000")
	SignBridgeRequest(beyond, "alice", false, now, "secret")
	over := httptest.NewRecorder()
	api.ServeHTTP(over, beyond)
	if over.Code != http.StatusRequestedRangeNotSatisfiable {
		t.Fatalf("out-of-range status=%d", over.Code)
	}
	if got, want := over.Header().Get("Content-Range"), "bytes */"+strconv.Itoa(len(payload)); got != want {
		t.Fatalf("out-of-range Content-Range=%q want %q", got, want)
	}

	// HEAD 只返回头部，长度与普通 GET 一致。
	head := signedJSONRequest(t, api, now, "alice", http.MethodHead, output, nil)
	if head.Code != http.StatusOK || head.Body.Len() != 0 {
		t.Fatalf("head status=%d bytes=%d", head.Code, head.Body.Len())
	}
	if got := head.Header().Get("Content-Length"); got != strconv.Itoa(len(payload)) {
		t.Fatalf("head Content-Length=%q want %d", got, len(payload))
	}

	missing := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/batches/"+batch.ID+"/merge-media/merge_missing_1", nil)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing status=%d", missing.Code)
	}
}
