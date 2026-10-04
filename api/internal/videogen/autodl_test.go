package videogen

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAutoDLProviderSelectsReferenceWorkflowAndRawAuthorization(t *testing.T) {
	var path, auth string
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		auth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil { t.Fatal(err) }
		_, _ = w.Write([]byte(`{"task_id":"h3-1","status":"queued"}`))
	}))
	defer server.Close()
	provider := AutoDLProvider{Config:Config{Provider:ProviderAutoDLH3, APIKey:"raw-token", Model:"minimax-h3-video", CreateURL:server.URL+"/{workflow}", TasksURL:server.URL+"/result/{id}"}, Client:server.Client(), ValidateURL:allowTestServer(server.URL)}
	task, err := provider.Submit(context.Background(), Request{Model:"minimax-h3-video", Prompt:"镜头环绕", Duration:15, Resolution:"480p竖", ReferenceURLs:[]string{"https://example.com/a.png","https://example.com/b.png"}})
	if err != nil { t.Fatal(err) }
	if task.ID != "h3-1" { t.Fatalf("task=%#v", task) }
	if path != "/"+AutoDLReferenceWorkflow { t.Fatalf("path=%q", path) }
	if auth != "raw-token" { t.Fatalf("auth=%q", auth) }
	if body["ref_image_0"] != "https://example.com/a.png" || body["ref_image_1"] != "https://example.com/b.png" { t.Fatalf("body=%#v", body) }
}

func TestAutoDLProviderUsesNoImageWorkflow(t *testing.T) {
	var path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { path=r.URL.Path; _,_=w.Write([]byte(`{"task_id":"h3-1"}`)) }))
	defer server.Close()
	provider := AutoDLProvider{Config:Config{Provider:ProviderAutoDLH3, APIKey:"token", Model:"minimax-h3-video", CreateURL:server.URL+"/{workflow}", TasksURL:server.URL+"/result/{id}"}, Client:server.Client(), ValidateURL:allowTestServer(server.URL)}
	if _, err := provider.Submit(context.Background(), Request{Model:"minimax-h3-video", Prompt:"x", Duration:5}); err != nil { t.Fatal(err) }
	if path != "/"+AutoDLNoImageWorkflow { t.Fatalf("path=%q", path) }
}

func TestAutoDLProviderRejectsMoreThanNineReferences(t *testing.T) {
	refs := make([]string, 10)
	for i := range refs { refs[i] = "https://example.com/a.png" }
	provider := AutoDLProvider{Config:Config{Provider:ProviderAutoDLH3, APIKey:"k", Model:"minimax-h3-video", CreateURL:"https://autodl.art/{workflow}", TasksURL:"https://autodl.art/result/{id}"}}
	if _, err := provider.Submit(context.Background(), Request{Model:"minimax-h3-video", Prompt:"x", Duration:5, ReferenceURLs:refs}); err == nil { t.Fatal("expected reference limit") }
}

func TestAutoDLProviderPollsResultURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _,_=w.Write([]byte(`{"status":"SUCCESS","results":[{"url":"https://cdn.example.com/h3.mp4"}],"duration_seconds":15}`)) }))
	defer server.Close()
	provider := AutoDLProvider{Config:Config{Provider:ProviderAutoDLH3, APIKey:"token", Model:"minimax-h3-video", CreateURL:server.URL+"/{workflow}", TasksURL:server.URL+"/result/{id}"}, Client:server.Client(), ValidateURL:allowTestServer(server.URL)}
	task, err := provider.Poll(context.Background(), Task{ID:"h3-1"})
	if err != nil { t.Fatal(err) }
	if task.State != StateSucceeded || task.MediaURL != "https://cdn.example.com/h3.mp4" || task.DurationMs != 15000 { t.Fatalf("task=%#v", task) }
}
