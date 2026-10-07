package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/workshop"
)

type fakeWorkshopService struct{ saved json.RawMessage }

func (s *fakeWorkshopService) Snapshot(context.Context, int64) (workshop.Snapshot, error) {
	return workshop.Snapshot{Intake: intake.Intake{ID: 7, Name: "任务", Status: intake.StatusCompleted}, Books: []intake.Book{{ID: 11, IntakeID: 7, Title: "书", OriginalText: "原文", Status: intake.BookStatusFetched}}, Settings: json.RawMessage(`{"processingRules":"r"}`), Prompts: []generation.Prompt{{Key: generation.PromptScript, Version: 1, Enabled: true}}}, nil
}
func (s *fakeWorkshopService) Save(_ context.Context, _ int64, value json.RawMessage) (json.RawMessage, error) {
	s.saved = value
	return value, nil
}

func TestWorkshopHandlerReturnsServerFactsAndPersistsSettings(t *testing.T) {
	service := &fakeWorkshopService{}
	handler := NewHandler(Dependencies{Workshop: service})
	get := httptest.NewRecorder()
	handler.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/api/v1/intakes/7/workshop", nil))
	if get.Code != http.StatusOK || !bytes.Contains(get.Body.Bytes(), []byte(`"originalText":"原文"`)) || !bytes.Contains(get.Body.Bytes(), []byte(`"prompts"`)) {
		t.Fatalf("GET %d %s", get.Code, get.Body.String())
	}
	put := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/api/v1/intakes/7/workshop", bytes.NewBufferString(`{"settings":{"knowledgeBase":"kb"}}`))
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(put, request)
	if put.Code != http.StatusOK || string(service.saved) != `{"knowledgeBase":"kb"}` {
		t.Fatalf("PUT %d saved=%s", put.Code, service.saved)
	}
}
