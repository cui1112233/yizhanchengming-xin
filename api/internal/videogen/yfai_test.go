package videogen

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestYFAIProviderUsesReferenceModeAndV88Params(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/media/generate" { http.NotFound(w,r); return }
		if r.Header.Get("Authorization") != "Bearer secret" { t.Fatalf("auth=%q", r.Header.Get("Authorization")) }
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil { t.Fatal(err) }
		_, _ = w.Write([]byte(`{"task_id":"task-y"}`))
	}))
	defer server.Close()
	provider := YFAIProvider{Config: Config{Provider:ProviderYFAISeedance, APIKey:"secret", Model:"seedance-2-0-official", CreateURL:server.URL}, Client:server.Client(), ValidateURL:allowTestServer(server.URL)}
	task, err := provider.Submit(context.Background(), Request{Model:"seedance-2-0-official", Prompt:"人物回头", Duration:10, Resolution:"720p", AspectRatio:"9:16", ReferenceURLs:[]string{"https://example.com/a.png"}})
	if err != nil { t.Fatal(err) }
	if task.ID != "task-y" || task.State != StateQueued { t.Fatalf("task=%#v", task) }
	params := body["params"].(map[string]any)
	if params["mode"] != "reference" || params["duration"] != "10" || params["resolution"] != "720p" || params["aspect_ratio"] != "9:16" { t.Fatalf("params=%#v", params) }
}

func TestYFAIProviderRejectsDurationOver15Seconds(t *testing.T) {
	provider := YFAIProvider{Config:Config{Provider:ProviderYFAISeedance, APIKey:"k", Model:"seedance-2-0-official", CreateURL:"https://yf.token6688.com"}}
	if _, err := provider.Submit(context.Background(), Request{Model:"seedance-2-0-official", Prompt:"x", Duration:16}); err == nil { t.Fatal("expected duration error") }
}

func TestYFAIProviderPollsCompletedOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"status":"completed","output_url":"https://cdn.example.com/y.mp4"}`)) }))
	defer server.Close()
	provider := YFAIProvider{Config:Config{Provider:ProviderYFAISeedance, APIKey:"k", Model:"seedance-2-0-official", CreateURL:server.URL}, Client:server.Client(), ValidateURL:allowTestServer(server.URL)}
	task, err := provider.Poll(context.Background(), Task{ID:"task-y"})
	if err != nil { t.Fatal(err) }
	if task.State != StateSucceeded || task.MediaURL != "https://cdn.example.com/y.mp4" { t.Fatalf("task=%#v", task) }
}
