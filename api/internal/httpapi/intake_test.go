package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/batchfactory"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/novel"
)

type fakeIntakeCreator struct {
	input  batchfactory.CreateIntakeInput
	result batchfactory.CreateIntakeResult
	err    error
}
func (f *fakeIntakeCreator) Create(_ context.Context, input batchfactory.CreateIntakeInput) (batchfactory.CreateIntakeResult, error) {
	f.input = input
	return f.result, f.err
}

func TestIntakeHandlerCreatesGroupedBookstoreIntake(t *testing.T) {
	creator := &fakeIntakeCreator{result: batchfactory.CreateIntakeResult{IntakeID:"intake-1", GroupCount:2, BookCount:3}}
	handler := NewIntakeHandler(creator, func(*http.Request) (string, error) { return "user-1", nil })
	body := `{"title":"批量","groups":[{"platform_id":"zhihu","platform_name":"知乎","max_txt":2000,"books":[{"book_id":"z1"},{"book_id":"z2","manual_gender":"女频"}]},{"platform_id":"dianzhong","platform_name":"点众","books":[{"book_id":"d1","style":"现代通用"}]}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/batch-factory/intakes", strings.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated { t.Fatalf("status=%d body=%s", w.Code, w.Body.String()) }
	if creator.input.Owner != "user-1" || len(creator.input.Groups) != 2 || creator.input.Groups[0].Books[1].ManualGender != novel.GenderFemale {
		t.Fatalf("input=%+v", creator.input)
	}
	var got batchfactory.CreateIntakeResult
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil { t.Fatal(err) }
	if got.IntakeID != "intake-1" || got.BookCount != 3 { t.Fatalf("response=%+v", got) }
}

func TestIntakeHandlerDoesNotTrustOwnerFromJSON(t *testing.T) {
	creator := &fakeIntakeCreator{result: batchfactory.CreateIntakeResult{IntakeID:"intake-1", GroupCount:1, BookCount:1}}
	handler := NewIntakeHandler(creator, func(*http.Request) (string, error) { return "trusted-owner", nil })
	body := `{"owner":"spoofed","groups":[{"platform_id":"2","platform_name":"番茄","books":[{"book_id":"b1"}]}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/batch-factory/intakes", strings.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest { t.Fatalf("expected unknown owner field rejection, status=%d body=%s", w.Code, w.Body.String()) }
}

func TestIntakeHandlerRequiresResolvedOwner(t *testing.T) {
	creator := &fakeIntakeCreator{}
	handler := NewIntakeHandler(creator, func(*http.Request) (string, error) { return "", errors.New("unauthorized") })
	req := httptest.NewRequest(http.MethodPost, "/api/batch-factory/intakes", strings.NewReader(`{"groups":[]}`))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized { t.Fatalf("status=%d", w.Code) }
}
