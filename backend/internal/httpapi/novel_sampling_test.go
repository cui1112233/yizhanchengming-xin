package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelcreation"
	"testing"
)

func TestNovelIntegrationSamplingOmissionPreservesOtherDefaults(t *testing.T) {
	for _, omit := range []bool{false, true} {
		var captured map[string]any
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&captured)
			_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{}"}}]}`))
		}))
		p := &integrationTextProvider{Endpoint: server.URL, APIKey: "test", Model: "actual", Client: server.Client()}
		request := batchfactoryv11.TextCompletionRequest{Temperature: .65, MaxTokens: 2000}
		raw, _ := json.Marshal(map[string]any{"OmitTemperature": omit, "OmitMaxTokens": omit})
		_ = json.Unmarshal(raw, &request)
		_, err := p.Complete(batchfactoryv11.WithSingleAttempt(context.Background()), request)
		server.Close()
		if err != nil {
			t.Fatal(err)
		}
		_, temp := captured["temperature"]
		_, tokens := captured["max_tokens"]
		if temp == omit || tokens == omit {
			t.Fatalf("omit=%v body=%v", omit, captured)
		}
	}
}

func TestNovelSavedChannelCapabilitiesReachActualPayload(t *testing.T) {
	var captured map[string]any
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_ = json.NewDecoder(r.Body).Decode(&captured)
		response := `{"title":"故事","genre":"悬疑","world":"旧城","characters":[{"name":"林","role":"主角","goal":"追查","conflict":"失踪"}],"coreConflict":"追查真相","volumes":[{"number":1,"title":"第一卷","goal":"追查","chapterRange":[1,10]}]}`
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]any{"content": response}}}})
	}))
	defer server.Close()
	saved, ok := sanitizeIntegrationChannel(map[string]any{"id": "task5", "name": "不支持采样参数", "group": "text", "baseUrl": server.URL, "model": "real-id", "apiKey": "fake-test-key", "capabilities": map[string]any{"supportsTemperature": false, "supportsMaxTokens": false, "secret": "discard-me"}}, nil)
	if !ok {
		t.Fatal("channel save rejected")
	}
	config := newIntegrationTestUserConfig(t, map[string]any{"customChannels": []any{saved}})
	service := novelcreation.NewService(novelcreation.NewMemoryStore(), novelcreation.NewMemoryRuntime(), novelcreation.NewPipeline(func(ctx context.Context, owner, id string) (batchfactoryv11.DirectorProvider, error) {
		provider, _, err := ResolveIntegrationTextProvider(ctx, config, owner, id)
		return provider, err
	}))
	ctx := context.Background()
	project, err := service.CreateProject(ctx, "tester", novelcreation.CreateProjectInput{Title: "故事", Genre: "悬疑", ModelID: "custom:task5"})
	if err != nil {
		t.Fatal(err)
	}
	session, err := service.CreateSession(ctx, "tester", novelcreation.CreateSessionInput{ProjectID: project.ID, Mode: "long_original"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := service.SendMessage(ctx, "tester", novelcreation.SendMessageInput{ProjectID: project.ID, SessionID: session.ID, Content: "生成蓝图", ModelID: "custom:task5", IdempotencyKey: "sampling-save"})
	if err != nil {
		t.Fatal(err)
	}
	done, err := service.Run(ctx, "tester", run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if done.Status != novelcreation.RunAwaitingConfirmation || calls != 1 {
		t.Fatalf("real chain status=%s calls=%d", done.Status, calls)
	}
	if _, ok = captured["temperature"]; ok {
		t.Fatalf("saved advertised capability was dropped: %v", captured)
	}
	if _, ok = captured["max_tokens"]; ok {
		t.Fatalf("saved token capability was dropped: %v", captured)
	}
	var runtime struct {
		Calls map[string]struct {
			Chosen   string                            `json:"chosenModelId"`
			Provider batchfactoryv11.TextModelMetadata `json:"provider"`
		} `json:"modelCalls"`
	}
	_ = json.Unmarshal(done.RuntimeSnapshot, &runtime)
	meta := runtime.Calls["novel-creation.foundation-plan"]
	if meta.Chosen != "custom:task5" || meta.Provider.ActualModel != "real-id" {
		t.Fatalf("actual selected/real model identity missing: %s", done.RuntimeSnapshot)
	}
	if caps := saved["capabilities"].(map[string]any); caps["secret"] != nil {
		t.Fatal("unrelated capability metadata persisted")
	}
}
