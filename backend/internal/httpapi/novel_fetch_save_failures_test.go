package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"qiantie/backend/internal/novelfetchworkshop"
)

type workshopSaveFailureStore struct {
	novelfetchworkshop.Store
	failConfig bool
	failRead   bool
}

func (s workshopSaveFailureStore) GetConfig(ctx context.Context, owner string) (map[string]any, error) {
	if s.failRead {
		return nil, errors.New("database unavailable")
	}
	return s.Store.GetConfig(ctx, owner)
}
func (s workshopSaveFailureStore) PutConfig(context.Context, string, map[string]any) error {
	return errors.New("database unavailable")
}
func (s workshopSaveFailureStore) PutDocument(ctx context.Context, owner string, d novelfetchworkshop.Document) error {
	if d.BookID == "bad" {
		return errors.New("database unavailable")
	}
	return s.Store.PutDocument(ctx, owner, d)
}

type workshopOptimizeTestProvider struct{}

func (workshopOptimizeTestProvider) Complete(context.Context, novelfetchworkshop.RewriteRequest) (string, error) {
	return `{"item":{"text":"优化结果"}}`, nil
}

func TestWorkshopLegacyOptimizeReportsStorageFailure(t *testing.T) {
	for _, failRead := range []bool{false, true} {
		t.Run(map[bool]string{false: "write", true: "read"}[failRead], func(t *testing.T) {
			base := novelfetchworkshop.NewMemoryStore()
			cfg := map[string]any{}
			novelfetchworkshop.SaveKnowledge(cfg, "opening_phrases", map[string]any{"id": "one", "text": "原文"})
			base.PutConfig(context.Background(), "alice", cfg)
			mux := http.NewServeMux()
			registerNovelFetchLegacyWorkflowRoutes(mux, workshopSaveFailureStore{Store: base, failRead: failRead}, novelfetchworkshop.NewMemoryQueue(), "", func(*http.Request) (novelfetchworkshop.RewriteProvider, error) {
				return workshopOptimizeTestProvider{}, nil
			})
			req := httptest.NewRequest(http.MethodPost, "/api/novel-fetch-workshop/knowledge/optimize", strings.NewReader(`{"kind":"opening_phrases","id":"one"}`))
			req = req.WithContext(context.WithValue(req.Context(), bridgeIdentityKey{}, BridgeIdentity{Username: "alice"}))
			res := httptest.NewRecorder()
			mux.ServeHTTP(res, req)
			if res.Code != http.StatusInternalServerError || strings.Contains(res.Body.String(), `"ok":true`) {
				t.Fatalf("false success: %d %s", res.Code, res.Body.String())
			}
		})
	}
}

func TestWorkshopTaskMetaCountsOnlySavedDocuments(t *testing.T) {
	base := novelfetchworkshop.NewMemoryStore()
	for _, id := range []string{"good", "bad"} {
		base.PutDocument(context.Background(), "alice", novelfetchworkshop.Document{BookID: id})
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"ids":["good","bad"],"count":3}`))
	req = req.WithContext(context.WithValue(req.Context(), bridgeIdentityKey{}, BridgeIdentity{Username: "alice"}))
	res := httptest.NewRecorder()
	updateTaskMeta(res, req, workshopSaveFailureStore{Store: base}, "aiCount")
	var out struct {
		Updated int      `json:"updated"`
		Failed  int      `json:"failed"`
		IDs     []string `json:"task_ids"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Updated != 1 || out.Failed != 1 || len(out.IDs) != 1 || out.IDs[0] != "good" {
		t.Fatalf("incorrect save results: %s", res.Body.String())
	}
}
