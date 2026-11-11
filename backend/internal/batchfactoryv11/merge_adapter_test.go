package batchfactoryv11

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestHTTPMergeAdapterSubmitsOrderedSourcesAndPolls(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			if r.Header.Get("Authorization") != "Bearer secret" {
				t.Fatalf("auth=%q", r.Header.Get("Authorization"))
			}
			if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
				t.Fatal(err)
			}
			_, _ = w.Write([]byte(`{"data":{"id":"merge-task-1"}}`))
			return
		}
		if !strings.HasSuffix(r.URL.Path, "/merge-task-1") {
			t.Fatalf("path=%s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"completed","outputUrl":"https://media.example/merged.mp4"}`))
	}))
	defer server.Close()
	adapter := &HTTPMergeAdapter{Endpoint: server.URL + "/merge", PollEndpoint: server.URL + "/merge/{id}", APIKey: "secret", Client: server.Client(), ValidateURL: func(raw string) (*url.URL, error) { return url.Parse(raw) }}
	sources := []MergeMedia{
		{ProductionJobID: "production-1", VideoID: "v1", MediaURL: "https://media.example/1.mp4", Order: 0},
		{ProductionJobID: "production-1", VideoID: "v2", MediaURL: "https://media.example/2.mp4", Order: 1},
	}
	job, err := adapter.Submit(context.Background(), "batch-1", sources, MergeOptions{TimingMode: "speed", Speed: 1.5})
	if err != nil {
		t.Fatal(err)
	}
	if job.Status != MergeQueued || job.ProviderTaskID != "merge-task-1" {
		t.Fatalf("job=%+v", job)
	}
	if received["batchId"] != "batch-1" || received["timingMode"] != "speed" {
		t.Fatalf("payload=%+v", received)
	}
	result, err := adapter.Poll(context.Background(), "batch-1", job)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != MergeSucceeded || result.OutputURL != "https://media.example/merged.mp4" {
		t.Fatalf("result=%+v", result)
	}
}

func TestHTTPMergeAdapterProviderContractIncludesSourceIdentityOrder(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("auth=%q", r.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"taskId":"merge-task-contract"}`))
	}))
	defer server.Close()

	adapter := &HTTPMergeAdapter{
		Endpoint: server.URL + "/v1/merge",
		APIKey:   "secret",
		Client:   server.Client(),
		ValidateURL: func(raw string) (*url.URL, error) {
			return url.Parse(raw)
		},
	}
	_, err := adapter.Submit(context.Background(), "batch-1", []MergeMedia{
		{ProductionJobID: "production-1", VideoID: "video-1", MediaURL: "https://media.example/1.mp4", Order: 0},
		{ProductionJobID: "production-1", VideoID: "video-2", MediaURL: "https://media.example/2.mp4", Order: 1},
	}, MergeOptions{TimingMode: "speed", Speed: 1})
	if err != nil {
		t.Fatal(err)
	}

	rawSources, ok := received["sources"].([]any)
	if !ok || len(rawSources) != 2 {
		t.Fatalf("sources=%#v", received["sources"])
	}
	for index, raw := range rawSources {
		source, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("source[%d]=%#v", index, raw)
		}
		if strings.TrimSpace(asString(source["productionJobId"])) == "" {
			t.Fatalf("source[%d] missing productionJobId: %#v", index, source)
		}
		if strings.TrimSpace(asString(source["videoId"])) == "" {
			t.Fatalf("source[%d] missing videoId: %#v", index, source)
		}
		if strings.TrimSpace(asString(source["mediaUrl"])) == "" {
			t.Fatalf("source[%d] missing mediaUrl: %#v", index, source)
		}
		if got, ok := source["order"].(float64); !ok || int(got) != index {
			t.Fatalf("source[%d] order=%#v", index, source["order"])
		}
	}
}

func asString(value any) string {
	text, _ := value.(string)
	return text
}

func TestHTTPMergeAdapterRejectsUnsafeSource(t *testing.T) {
	adapter := &HTTPMergeAdapter{Endpoint: "https://merge.example/submit", APIKey: "secret"}
	if _, err := adapter.Submit(context.Background(), "batch-1", []MergeMedia{{ProductionJobID: "production-1", VideoID: "v1", MediaURL: "http://media.example/1.mp4", Order: 0}}, MergeOptions{}); err == nil {
		t.Fatal("expected unsafe source rejection")
	}
}
