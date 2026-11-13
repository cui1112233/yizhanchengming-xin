package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"qiantie/backend/internal/novelfetchworkshop"
)

type countRouteRewriteProvider struct {
	calls   int
	prompts []string
}

func (p *countRouteRewriteProvider) Complete(_ context.Context, req novelfetchworkshop.RewriteRequest) (string, error) {
	p.calls++
	p.prompts = append(p.prompts, req.UserPrompt)
	return "改后首行", nil
}

func TestWorkshopGenerateRouteUsesSavedCountAndPreservesOutputs(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	store.PutConfig(ctx, "alice", map[string]any{"appConfig": map[string]any{"rewrite": map[string]any{"default_template_id": "plain", "process_line_count": 1}}, "knowledgeLibraries": map[string]any{"rewrite_templates": map[string]any{"profiles": []any{map[string]any{"id": "plain", "prompt": "保持原文", "response_mode": "plain_text"}}}}})
	store.PutDocument(ctx, "alice", novelfetchworkshop.Document{BookID: "123", Original: "首行原文", Meta: map[string]any{"aiCount": 2}, Versions: map[string]any{"ai1": "旧版本"}})
	provider := &countRouteRewriteProvider{}
	mux := http.NewServeMux()
	registerNovelFetchWorkshopRoutes(mux, store, nil, nil, nil, nil, provider, nil, "", nil, nil, nil)
	call := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/novel-fetch-workshop/tasks/123/generate-ai", strings.NewReader(`{}`))
		req = req.WithContext(context.WithValue(ctx, bridgeIdentityKey{}, BridgeIdentity{Username: "alice"}))
		res := httptest.NewRecorder()
		mux.ServeHTTP(res, req)
		return res
	}
	res := call()
	if res.Code != http.StatusOK {
		t.Fatalf("generation: %d %s", res.Code, res.Body.String())
	}
	doc, _ := store.GetDocument(ctx, "alice", "123")
	if doc.Versions["ai1"] != "旧版本" || doc.Versions["ai2"] != "改后首行" || provider.calls != 1 || !strings.Contains(provider.prompts[0], "第 2 个") {
		t.Fatalf("wrong target: %#v calls=%d", doc, provider.calls)
	}
	res = call()
	if res.Code != http.StatusOK || provider.calls != 1 {
		t.Fatalf("repeat generated output: %d %s calls=%d", res.Code, res.Body.String(), provider.calls)
	}
}

func TestWorkshopSavedAICountTargetsOnlyMissingVersions(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	store.PutDocument(ctx, "alice", novelfetchworkshop.Document{BookID: "one", Meta: map[string]any{"status": "running"}, Versions: map[string]any{"ai1": "preserved body", "ai5": "extra body"}})
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"ids":["one"],"count":3}`))
	req = req.WithContext(context.WithValue(ctx, bridgeIdentityKey{}, BridgeIdentity{Username: "alice"}))
	res := httptest.NewRecorder()
	updateTaskMeta(res, req, store, "aiCount")
	if res.Code != http.StatusOK {
		t.Fatal(res.Body.String())
	}
	doc, _ := store.GetDocument(ctx, "alice", "one")
	if doc.Versions["ai1"] != "preserved body" || doc.Versions["ai5"] != "extra body" || doc.Meta["status"] != "running" {
		t.Fatalf("changed existing work: %#v", doc)
	}
	count := workshopAIRequestedCount(doc, map[string]any{}, 0)
	if count != 3 || !reflect.DeepEqual(workshopMissingAIVersions(doc, count), []string{"ai2", "ai3"}) {
		t.Fatalf("saved count not honored: count=%d missing=%v", count, workshopMissingAIVersions(doc, count))
	}
	// Reducing the next target never deletes already generated extra versions.
	if missing := workshopMissingAIVersions(doc, 1); len(missing) != 0 {
		t.Fatalf("regenerates existing ai1: %v", missing)
	}
}

func TestWorkshopAICountLimitAndDurableReadyVersions(t *testing.T) {
	cfg := map[string]any{"appConfig": map[string]any{"rewrite": map[string]any{"max_ai_count": 3, "default_ai_count": 2}}}
	doc := novelfetchworkshop.Document{BodyRefs: map[string]novelfetchworkshop.BodyRef{"ai1": {State: "ready"}}, Versions: map[string]any{"ai2": map[string]any{"state": "ready"}}}
	if got := workshopAIRequestedCount(doc, cfg, 20); got != 3 {
		t.Fatalf("max limit ignored: %d", got)
	}
	if got := workshopAIRequestedCount(doc, cfg, 0); got != 2 {
		t.Fatalf("default ignored: %d", got)
	}
	if got := workshopMissingAIVersions(doc, 3); !reflect.DeepEqual(got, []string{"ai3"}) {
		t.Fatalf("durable output ignored: %v", got)
	}
}

func TestWorkshopAICountRejectsInvalidValues(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	for _, body := range []string{`{"count":-1}`, `{"count":21}`} {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		res := httptest.NewRecorder()
		updateTaskMeta(res, req, store, "aiCount")
		if res.Code != http.StatusBadRequest {
			t.Fatalf("invalid value accepted: %s", body)
		}
	}
}
