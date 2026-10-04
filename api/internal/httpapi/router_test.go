package httpapi

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/cui1112233/yizhanchengming-xin/api/internal/agent"
    "github.com/cui1112233/yizhanchengming-xin/api/internal/batchfactory"
    "github.com/cui1112233/yizhanchengming-xin/api/internal/media"
)

func TestRouterExposesHealthAndBatchFactory(t *testing.T) {
    starter := &fakeStarter{}
    handler := NewRouter(starter)

    health := httptest.NewRecorder()
    handler.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
    if health.Code != http.StatusOK { t.Fatalf("health status=%d", health.Code) }

    batch := httptest.NewRecorder()
    handler.ServeHTTP(batch, httptest.NewRequest(http.MethodGet, "/api/batch-factory/jobs", nil))
    if batch.Code != http.StatusMethodNotAllowed { t.Fatalf("batch route status=%d", batch.Code) }
}

func TestRouterWithIntakesExposesGroupedIntakeAndRunRoutes(t *testing.T) {
    starter := &fakeStarter{}
    creator := &fakeIntakeCreator{result:batchfactory.CreateIntakeResult{IntakeID:"intake-1", GroupCount:1, BookCount:1}}
    handler := NewRouterWithIntakes(starter, creator, func(*http.Request) (string,error) { return "user-1", nil })

    intake := httptest.NewRecorder()
    body := `{"groups":[{"platform_id":"2","platform_name":"番茄","books":[{"book_id":"b1"}]}]}`
    handler.ServeHTTP(intake, httptest.NewRequest(http.MethodPost, "/api/batch-factory/intakes", strings.NewReader(body)))
    if intake.Code != http.StatusCreated { t.Fatalf("intake route status=%d body=%s", intake.Code, intake.Body.String()) }

    run := httptest.NewRecorder()
    handler.ServeHTTP(run, httptest.NewRequest(http.MethodPost, "/api/batch-factory/runs", strings.NewReader(body)))
    if run.Code != http.StatusCreated { t.Fatalf("run route status=%d body=%s", run.Code, run.Body.String()) }
    if starter.input.IntakeID != "intake-1" { t.Fatalf("run did not start created intake: %+v", starter.input) }
}

func TestRouterWithAgentExposesAgentThreadsWithoutBreakingBatchRoutes(t *testing.T) {
    starter := &fakeStarter{}
    agentAPI := &fakeAgentAPI{threads: []agent.Thread{{ID: "thread-1", Title: "测试", Status: "active"}}}
    owner := func(*http.Request) (string,error) { return "user-1", nil }
    handler := NewRouterWithAgent(starter, nil, owner, agentAPI)

    threads := httptest.NewRecorder()
    handler.ServeHTTP(threads, httptest.NewRequest(http.MethodGet, "/api/agent/threads", nil))
    if threads.Code != http.StatusOK { t.Fatalf("agent route status=%d body=%s", threads.Code, threads.Body.String()) }

    batch := httptest.NewRecorder()
    handler.ServeHTTP(batch, httptest.NewRequest(http.MethodGet, "/api/batch-factory/jobs", nil))
    if batch.Code != http.StatusMethodNotAllowed { t.Fatalf("batch route status=%d", batch.Code) }
}

func TestUnifiedRouterExposesMediaPreview(t *testing.T) {
    starter := &fakeStarter{}
    owner := func(*http.Request) (string,error) { return "user-1", nil }
    resolver := &mediaResolverFake{result: media.ResolvedAsset{Asset: media.Reference{ID: "asset_1", MediaType: media.TypeImage}, PreviewURL: "https://signed.example/a.png"}}
    handler := NewRouterWithServices(starter, nil, owner, nil, resolver)

    preview := httptest.NewRecorder()
    handler.ServeHTTP(preview, httptest.NewRequest(http.MethodGet, "/api/media/assets/asset_1", nil))
    if preview.Code != http.StatusOK { t.Fatalf("media route status=%d body=%s", preview.Code, preview.Body.String()) }
}
