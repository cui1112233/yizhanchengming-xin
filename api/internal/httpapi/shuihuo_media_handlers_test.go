package httpapi

import (
	"context"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/shuihuo"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type shuihuoHTTPFake struct{ got shuihuo.CreateMediaTaskInput }

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
func (*shuihuoHTTPFake) ListMediaTasks(context.Context, int64, int64) ([]shuihuo.MediaTask, error) {
	return nil, nil
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
