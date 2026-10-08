package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/shuihuo"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type shuihuoHTTPFake struct {
	got         shuihuo.CreateMediaTaskInput
	upload      shuihuo.UploadAssetInput
	listedMedia []shuihuo.MediaTask
}

func (*shuihuoHTTPFake) CreateSegment(context.Context, shuihuo.CreateSegmentInput) (shuihuo.Segment, error) {
	return shuihuo.Segment{}, nil
}
func (*shuihuoHTTPFake) UpdateSegment(context.Context, int64, int64, int64, shuihuo.UpdateSegmentInput) (shuihuo.Segment, error) {
	return shuihuo.Segment{}, nil
}
func (*shuihuoHTTPFake) ListSegments(context.Context, int64, int64) ([]shuihuo.Segment, error) {
	return nil, nil
}
func (*shuihuoHTTPFake) CreateAsset(context.Context, shuihuo.CreateAssetInput) (shuihuo.Asset, error) {
	return shuihuo.Asset{}, nil
}
func (*shuihuoHTTPFake) ListAssets(context.Context, int64, int64, int64) ([]shuihuo.Asset, error) {
	return nil, nil
}
func (f *shuihuoHTTPFake) CreateMediaTask(_ context.Context, i shuihuo.CreateMediaTaskInput) (shuihuo.MediaTask, error) {
	f.got = i
	return shuihuo.MediaTask{ID: 1, Status: shuihuo.MediaPendingExecutor}, nil
}
func (f *shuihuoHTTPFake) ListMediaTasks(context.Context, int64, int64) ([]shuihuo.MediaTask, error) {
	return f.listedMedia, nil
}
func (*shuihuoHTTPFake) ReorderSegments(context.Context, int64, int64, []int64) ([]shuihuo.Segment, error) {
	return nil, nil
}
func (*shuihuoHTTPFake) ListCandidates(context.Context, int64, int64, int64) ([]shuihuo.Candidate, error) {
	return nil, nil
}
func (*shuihuoHTTPFake) SelectCandidate(context.Context, int64, int64, int64, int64) (shuihuo.Candidate, error) {
	return shuihuo.Candidate{}, nil
}
func (*shuihuoHTTPFake) RetryMediaTask(context.Context, int64, int64, int64) (shuihuo.MediaTask, error) {
	return shuihuo.MediaTask{}, nil
}
func (f *shuihuoHTTPFake) UploadAsset(_ context.Context, in shuihuo.UploadAssetInput) (shuihuo.Asset, error) {
	f.upload = in
	return shuihuo.Asset{ID: 9, ObjectKey: "server-generated"}, nil
}

func TestShuihuoAssetUploadUsesScopedMultipartInput(t *testing.T) {
	f := &shuihuoHTTPFake{}
	h := NewHandler(Dependencies{ShuihuoMedia: f})
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("type", "image")
	_ = writer.WriteField("segmentId", "12")
	part, err := writer.CreateFormFile("file", "still.png")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("image data"))
	_ = writer.Close()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/batch-projects/7/books/8/shuihuo/assets/upload", body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	h.ServeHTTP(w, r)
	if w.Code != http.StatusCreated || f.upload.BatchProjectID != 7 || f.upload.BookID != 8 || f.upload.SegmentID != 12 || f.upload.Type != shuihuo.AssetImage || f.upload.Filename != "still.png" {
		t.Fatalf("code=%d upload=%+v", w.Code, f.upload)
	}
}
func (*shuihuoHTTPFake) OpenAsset(context.Context, int64, int64, int64) (shuihuo.Asset, io.ReadCloser, error) {
	return shuihuo.Asset{}, nil, shuihuo.ErrNotFound
}
func TestShuihuoMediaTaskHTTPIsProjectScopedAndPending(t *testing.T) {
	f := &shuihuoHTTPFake{}
	h := NewHandler(Dependencies{ShuihuoMedia: f})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/batch-projects/7/books/8/shuihuo/media-tasks", strings.NewReader(`{"kind":"image","requestId":"request-1"}`))
	h.ServeHTTP(w, r)
	if w.Code != http.StatusAccepted || f.got.BatchProjectID != 7 || f.got.BookID != 8 || f.got.Kind != shuihuo.MediaImage {
		t.Fatalf("code=%d input=%+v", w.Code, f.got)
	}
}

func TestShuihuoMediaTaskHTTPUsesFrontendJSONContract(t *testing.T) {
	f := &shuihuoHTTPFake{listedMedia: []shuihuo.MediaTask{{
		ID: 18, BatchProjectID: 7, BookID: 8, Kind: shuihuo.MediaAudio,
		Status: shuihuo.MediaFailed, ErrorCode: "executor_unavailable",
	}}}
	h := NewHandler(Dependencies{ShuihuoMedia: f})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/batch-projects/7/books/8/shuihuo/media-tasks", nil)
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
	var body []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body) != 1 || body[0]["id"] != float64(18) || body[0]["status"] != "failed" || body[0]["errorCode"] != "executor_unavailable" {
		t.Fatalf("frontend media contract=%v", body)
	}
}
