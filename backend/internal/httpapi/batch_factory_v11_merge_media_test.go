package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
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

func TestMergeMediaCacheValidatorsAndSignedPlayback(t *testing.T) {
	now := time.Unix(1700000100, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "b", Books: []batchfactoryv11.CreateBookInput{{Title: "k", SourceText: "正文"}}})
	if err != nil {
		t.Fatal(err)
	}
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	artifactID := "merge_cache_media_1"
	payload := "0000ftypisom-local-merged-video" + strings.Repeat("0123456789", 100)
	if _, err := files.SaveMP4(artifactID, strings.NewReader(payload)); err != nil {
		t.Fatal(err)
	}
	output := "/api/batch-factory/v12/batches/" + batch.ID + "/merge-media/" + artifactID
	if _, err := store.CreateMergeJob(context.Background(), batchfactoryv11.MergeJob{Owner: "alice", BatchID: batch.ID, RequestID: "cache-media", Status: batchfactoryv11.MergeSucceeded, OutputURL: output, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 5, Store: store, Merge: &batchfactoryv11.MergeService{Store: store, Enabled: true}, LocalArtifacts: files})

	full := signedJSONRequest(t, api, now, "alice", http.MethodGet, output, nil)
	if full.Code != http.StatusOK {
		t.Fatalf("full status=%d body=%s", full.Code, full.Body.String())
	}
	etag := strings.TrimSpace(full.Header().Get("ETag"))
	if etag == "" {
		t.Fatal("merge media response is missing ETag")
	}
	if cacheControl := full.Header().Get("Cache-Control"); !strings.Contains(cacheControl, "private") || !strings.Contains(cacheControl, "max-age=3600") {
		t.Fatalf("Cache-Control=%q", cacheControl)
	}

	// 同一个 URL 携带 If-None-Match 二次请求必须命中 304，避免播放器重复整包回源。
	revalidate := httptest.NewRequest(http.MethodGet, output, nil)
	revalidate.Header.Set("If-None-Match", etag)
	SignBridgeRequest(revalidate, "alice", false, now, "secret")
	revalidated := httptest.NewRecorder()
	api.ServeHTTP(revalidated, revalidate)
	if revalidated.Code != http.StatusNotModified || revalidated.Body.Len() != 0 {
		t.Fatalf("revalidate status=%d bytes=%d", revalidated.Code, revalidated.Body.Len())
	}

	// Range 仍然返回 206，且 206 也带 ETag。
	ranged := httptest.NewRequest(http.MethodGet, output, nil)
	ranged.Header.Set("Range", "bytes=0-99")
	SignBridgeRequest(ranged, "alice", false, now, "secret")
	partial := httptest.NewRecorder()
	api.ServeHTTP(partial, ranged)
	if partial.Code != http.StatusPartialContent {
		t.Fatalf("range status=%d", partial.Code)
	}
	if got := partial.Header().Get("ETag"); got != etag {
		t.Fatalf("range ETag=%q want %q", got, etag)
	}

	tokenResponse := signedJSONRequest(t, api, now, "alice", http.MethodPost, output+"/token", nil)
	if tokenResponse.Code != http.StatusOK {
		t.Fatalf("token status=%d body=%s", tokenResponse.Code, tokenResponse.Body.String())
	}
	var signed struct {
		URL       string `json:"url"`
		ExpiresAt string `json:"expiresAt"`
	}
	if err := json.Unmarshal(tokenResponse.Body.Bytes(), &signed); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(signed.URL, "/merge-media/"+artifactID+"?token=") || signed.ExpiresAt == "" {
		t.Fatalf("signed playback url=%q expiresAt=%q", signed.URL, signed.ExpiresAt)
	}

	// 原生 <video> 无法带 Authorization：签名地址必须能在没有任何 bridge 头的情况下播放。
	playback := httptest.NewRequest(http.MethodGet, signed.URL, nil)
	played := httptest.NewRecorder()
	api.ServeHTTP(played, playback)
	if played.Code != http.StatusOK || played.Body.Len() != len(payload) {
		t.Fatalf("token playback status=%d bytes=%d", played.Code, played.Body.Len())
	}
	if got := played.Header().Get("Content-Type"); got != "video/mp4" {
		t.Fatalf("token playback Content-Type=%q", got)
	}

	// 首帧只需要首个 Range 分片，也必须能在无鉴权头情况下拿到 206。
	firstFrame := httptest.NewRequest(http.MethodGet, signed.URL, nil)
	firstFrame.Header.Set("Range", "bytes=0-99")
	firstChunk := httptest.NewRecorder()
	api.ServeHTTP(firstChunk, firstFrame)
	if firstChunk.Code != http.StatusPartialContent {
		t.Fatalf("token range status=%d", firstChunk.Code)
	}

	expiredToken, tokenErr := localexecutor.ArtifactPublicToken("secret", "alice", artifactID, now.Add(-time.Minute))
	if tokenErr != nil {
		t.Fatal(tokenErr)
	}
	expired := httptest.NewRequest(http.MethodGet, "/api/batch-factory/v12/batches/"+batch.ID+"/merge-media/"+artifactID+"?token="+expiredToken, nil)
	expiredRecorder := httptest.NewRecorder()
	api.ServeHTTP(expiredRecorder, expired)
	if expiredRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("expired token status=%d", expiredRecorder.Code)
	}
}
