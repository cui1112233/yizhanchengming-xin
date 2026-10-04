package videogen

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPersonalProviderSubmitsV88Payload(t *testing.T) {
	var auth string
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/create" { http.NotFound(w, r); return }
		auth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil { t.Fatal(err) }
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"task_id":"task-1","status":"queued"}`))
	}))
	defer server.Close()

	provider := PersonalProvider{
		Config: Config{Provider: ProviderPersonalAPI, APIKey: "secret", Model: "yd2.0-mini", CreateURL: server.URL + "/create", TasksURL: server.URL + "/tasks", ResultURL: server.URL + "/tasks/{id}/result"},
		Client: server.Client(),
		ValidateURL: allowTestServer(server.URL),
	}
	task, err := provider.Submit(context.Background(), Request{Model: "yd2-mini-video", Prompt: "镜头推进", Duration: 10, AspectRatio: "9:16", Resolution: "720p", ReferenceURLs: []string{"https://example.com/ref.png"}})
	if err != nil { t.Fatal(err) }
	if task.ID != "task-1" || task.State != StateQueued { t.Fatalf("task=%#v", task) }
	if auth != "Bearer secret" { t.Fatalf("auth=%q", auth) }
	if body["model"] != "yd2.0-mini" || body["prompt"] != "镜头推进" || body["duration"] != "10" || body["aspect_ratio"] != "9:16" || body["resolution"] != "720p" {
		t.Fatalf("body=%#v", body)
	}
	images, ok := body["image_urls"].([]any)
	if !ok || len(images) != 2 { t.Fatalf("image_urls=%#v", body["image_urls"]) }
	if images[0] != DefaultPersonalFirstFrameURL || images[1] != "https://example.com/ref.png" { t.Fatalf("image_urls=%#v", images) }
}

func TestPersonalProviderRejectsMoreThanThreeReferences(t *testing.T) {
	provider := PersonalProvider{Config: Config{Provider: ProviderPersonalAPI, APIKey: "k", Model: "yd2.0-mini", CreateURL: "https://ydapi.yadiai.cn/create", TasksURL: "https://ydapi.yadiai.cn/tasks", ResultURL: "https://ydapi.yadiai.cn/tasks/{id}/result"}}
	_, err := provider.Submit(context.Background(), Request{Model: "yd2.0-mini", Prompt: "x", Duration: 5, ReferenceURLs: []string{"https://a.example/1", "https://a.example/2", "https://a.example/3", "https://a.example/4"}})
	if err == nil || !strings.Contains(err.Error(), "at most 3") { t.Fatalf("err=%v", err) }
}

func TestPersonalProviderPollsCompletedTaskAndResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/tasks/task-1":
			_, _ = w.Write([]byte(`{"status":"SUCCESS"}`))
		case "/tasks/task-1/result":
			_, _ = w.Write([]byte(`{"data":{"urls":["https://cdn.example.com/result.mp4"]},"duration_seconds":12.5}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	provider := PersonalProvider{
		Config: Config{Provider: ProviderPersonalAPI, APIKey: "secret", Model: "yd2.0-mini", CreateURL: server.URL + "/create", TasksURL: server.URL + "/tasks", ResultURL: server.URL + "/tasks/{id}/result"},
		Client: server.Client(), ValidateURL: allowTestServer(server.URL),
	}
	task, err := provider.Poll(context.Background(), Task{ID: "task-1", State: StateQueued})
	if err != nil { t.Fatal(err) }
	if task.State != StateSucceeded || task.MediaURL != "https://cdn.example.com/result.mp4" || task.DurationMs != 12500 { t.Fatalf("task=%#v", task) }
}

func TestPersonalProviderSurfacesHTTP200ProviderRejection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"quota exceeded"}`))
	}))
	defer server.Close()
	provider := PersonalProvider{Config: Config{Provider: ProviderPersonalAPI, APIKey: "secret", Model: "yd2.0-mini", CreateURL: server.URL, TasksURL: server.URL, ResultURL: server.URL + "/{id}"}, Client: server.Client(), ValidateURL: allowTestServer(server.URL)}
	_, err := provider.Submit(context.Background(), Request{Model: "yd2.0-mini", Prompt: "x", Duration: 5})
	if err == nil || !strings.Contains(err.Error(), "quota exceeded") { t.Fatalf("err=%v", err) }
}

func allowTestServer(base string) URLValidator {
	return func(raw string) error {
		if strings.HasPrefix(raw, base) || strings.HasPrefix(raw, "https://example.com/") || strings.HasPrefix(raw, "https://cdn.example.com/") { return nil }
		return nil
	}
}
