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

func TestAutoDLH3VideoAdapterUsesConfiguredWorkflowWhenNoReferenceImages(t *testing.T) {
	const workflowID = "saved_workflow_from_catalog"
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"id":"task-1"}`))
	}))
	defer server.Close()
	adapter := &AutoDLH3VideoAdapter{CreateURL: server.URL + "/{workflow}", TasksURL: server.URL + "/result/{id}", APIKey: "fake-key", Model: AutoDLH3Model, WorkflowID: workflowID, ValidateURL: func(raw string) (*url.URL, error) { return url.Parse(raw) }}
	_, err := adapter.Submit(context.Background(), FrozenVideoModel{ID: AutoDLH3Model, MaxDuration: 15}, FinalPrompt{CompiledPrompt: "prompt"})
	if err != nil { t.Fatal(err) }
	if !strings.Contains(gotPath, workflowID) { t.Fatalf("workflow path=%q", gotPath) }
}

func TestAutoDLH3VideoAdapterAlwaysIncludesFallbackReferenceImage(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/comfyui/comfyui_workflow/minimax_h3_lightx2v_v5_15s" {
			t.Fatalf("request=%s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "autodl-secret" {
			t.Fatalf("authorization=%q", r.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"code":"Success","data":{"task_id":"h3-task-1","status":"QUEUED"}}`))
	}))
	defer server.Close()

	adapter := &AutoDLH3VideoAdapter{
		CreateURL: server.URL + "/api/v1/comfyui/comfyui_workflow/{workflow}",
		TasksURL:  server.URL + "/api/v1/comfyui/comfyui_workflow/result/{id}",
		APIKey:    "autodl-secret",
		Model:     AutoDLH3Model,
		Client:    server.Client(),
		ValidateURL: func(raw string) (*url.URL, error) {
			return url.Parse(raw)
		},
	}
	prompt := FinalPrompt{
		CompiledPrompt:  "夜晚的街道，电影感运镜",
		DurationSeconds: 15,
		EffectiveSettings: EffectiveSettings{Values: SettingsPatch{
			"imageAspectRatio": json.RawMessage(`"1:1"`),
			"videoResolution":  json.RawMessage(`"480p"`),
		}},
	}

	ref, err := adapter.Submit(context.Background(), FrozenVideoModel{ID: AutoDLH3Model, MaxDuration: 15}, prompt)
	if err != nil {
		t.Fatal(err)
	}
	if ref.ProviderTaskID != "h3-task-1" || ref.State != ProductionQueued {
		t.Fatalf("ref=%+v", ref)
	}
	if received["prompt"] != prompt.CompiledPrompt || received["duration"] != float64(15) || received["resolution"] != "480p竖" {
		t.Fatalf("payload=%+v", received)
	}
	// Acceptance: even a book with zero reference images must submit a usable
	// ref_image_0, otherwise the provider answers 缺少必填参数：ref_image_0.
	if received["ref_image_0"] != "https://yzcm.site/assets/empty-1x1.png" {
		t.Fatalf("ref_image_0=%v payload=%+v", received["ref_image_0"], received)
	}
}

func TestAutoDLH3VideoAdapterUsesConfiguredWorkflowID(t *testing.T) {
	const workflowID = "minimax_h3_image_audio_to_video_v1"
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantPath := "/api/v1/comfyui/comfyui_workflow/" + workflowID
		if r.Method != http.MethodPost || r.URL.Path != wantPath {
			t.Fatalf("request=%s %s, want POST %s", r.Method, r.URL.Path, wantPath)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"code":"Success","data":{"task_id":"h3-configured-workflow","status":"QUEUED"}}`))
	}))
	defer server.Close()

	adapter := &AutoDLH3VideoAdapter{
		CreateURL:   server.URL + "/api/v1/comfyui/comfyui_workflow/{workflow}",
		TasksURL:    server.URL + "/api/v1/comfyui/comfyui_workflow/result/{id}",
		APIKey:      "autodl-secret",
		Model:       AutoDLH3Model,
		WorkflowID:  workflowID,
		Client:      server.Client(),
		ValidateURL: func(raw string) (*url.URL, error) { return url.Parse(raw) },
	}

	ref, err := adapter.Submit(context.Background(), FrozenVideoModel{ID: AutoDLH3Model, MaxDuration: 15}, FinalPrompt{
		CompiledPrompt: "镜头推进",
		EffectiveSettings: EffectiveSettings{Values: SettingsPatch{
			"videoResolution": json.RawMessage(`"720p"`),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if ref.ProviderTaskID != "h3-configured-workflow" {
		t.Fatalf("ref=%+v", ref)
	}
	if received["resolution"] != "768p竖" {
		t.Fatalf("resolution=%q, want 768p竖", received["resolution"])
	}
}

func TestAutoDLH3VideoAdapterReportsSafeProviderRejection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":"InvalidWorkflow","msg":"workflow is not available"}`))
	}))
	defer server.Close()

	adapter := &AutoDLH3VideoAdapter{
		CreateURL:   server.URL + "/api/v1/comfyui/comfyui_workflow/{workflow}",
		TasksURL:    server.URL + "/api/v1/comfyui/comfyui_workflow/result/{id}",
		APIKey:      "autodl-secret",
		Model:       AutoDLH3Model,
		WorkflowID:  "minimax_h3_image_audio_to_video_v2_15s",
		Client:      server.Client(),
		ValidateURL: func(raw string) (*url.URL, error) { return url.Parse(raw) },
	}

	_, err := adapter.Submit(context.Background(), FrozenVideoModel{ID: AutoDLH3Model, MaxDuration: 15}, FinalPrompt{CompiledPrompt: "镜头推进"})
	if err == nil {
		t.Fatal("Submit() error = nil")
	}
	if got := err.Error(); !strings.Contains(got, "code=InvalidWorkflow") || !strings.Contains(got, "workflow is not available") || strings.Contains(got, "autodl-secret") {
		t.Fatalf("unexpected error: %q", got)
	}
}

func TestAutoDLH3VideoAdapterUsesReferenceWorkflowAndPollsResult(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodPost {
			var received map[string]any
			if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
				t.Fatal(err)
			}
			if received["ref_image_0"] != "https://assets.example/ref.png" {
				t.Fatalf("payload=%+v", received)
			}
			_, _ = w.Write([]byte(`{"code":"Success","data":{"task_id":"h3-task-2","status":"QUEUED"}}`))
			return
		}
		if r.URL.Path != "/api/v1/comfyui/comfyui_workflow/result/h3-task-2" {
			t.Fatalf("poll path=%s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"code":"Success","data":{"task_id":"h3-task-2","status":"SUCCESS","results":["https://media.example/h3.mp4"]}}`))
	}))
	defer server.Close()

	adapter := &AutoDLH3VideoAdapter{
		CreateURL: server.URL + "/api/v1/comfyui/comfyui_workflow/{workflow}",
		TasksURL:  server.URL + "/api/v1/comfyui/comfyui_workflow/result/{id}",
		APIKey:    "autodl-secret",
		Model:     AutoDLH3Model,
		Client:    server.Client(),
		ValidateURL: func(raw string) (*url.URL, error) {
			return url.Parse(raw)
		},
	}
	prompt := FinalPrompt{
		CompiledPrompt:  "人物转身，镜头推进",
		DurationSeconds: 10,
		EffectiveSettings: EffectiveSettings{Values: SettingsPatch{
			"imageUrls":  json.RawMessage(`["https://assets.example/ref.png"]`),
			"resolution": json.RawMessage(`"768p横"`),
		}},
	}

	ref, err := adapter.Submit(context.Background(), FrozenVideoModel{ID: AutoDLH3Model, MaxDuration: 15}, prompt)
	if err != nil {
		t.Fatal(err)
	}
	result, err := adapter.Poll(context.Background(), FrozenVideoModel{ID: AutoDLH3Model, MaxDuration: 15}, ref)
	if err != nil {
		t.Fatal(err)
	}
	if result.State != ProductionSucceeded || result.MediaURL != "https://media.example/h3.mp4" {
		t.Fatalf("result=%+v", result)
	}
	if len(requests) != 2 || !strings.Contains(requests[0], "minimax_h3_lightx2v_v5_15s") {
		t.Fatalf("requests=%v", requests)
	}
}

func TestAutoDLH3VideoProviderNormalizesAndKeepsKeyOutOfView(t *testing.T) {
	registry := NewMemoryVideoProviderRegistry()
	if err := registry.Put(context.Background(), "alice", VideoProviderConfig{
		Provider: VideoProviderAutoDLH3,
		APIKey:   "autodl-secret",
	}); err != nil {
		t.Fatal(err)
	}
	view, err := registry.View(context.Background(), "alice", "h3")
	if err != nil {
		t.Fatal(err)
	}
	if view.Provider != VideoProviderAutoDLH3 || view.Model != AutoDLH3Model || !view.Configured {
		t.Fatalf("view=%+v", view)
	}
}

func TestAutoDLH3VideoAdapterUpgradesConfiguredTextWorkflowForFallbackImage(t *testing.T) {
	var received map[string]any
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"id":"task-1"}`))
	}))
	defer server.Close()

	adapter := &AutoDLH3VideoAdapter{
		CreateURL: server.URL + "/{workflow}", TasksURL: server.URL + "/result/{id}",
		APIKey: "fake-key", Model: AutoDLH3Model,
		WorkflowID:                AutoDLH3NoImageWorkflow,
		FallbackReferenceImageURL: "https://yzcm.site/assets/empty-1x1.png",
		ValidateURL:               func(raw string) (*url.URL, error) { return url.Parse(raw) },
	}
	if _, err := adapter.Submit(context.Background(), FrozenVideoModel{ID: AutoDLH3Model, MaxDuration: 15}, FinalPrompt{CompiledPrompt: "prompt"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotPath, AutoDLH3ReferenceWorkflow) {
		t.Fatalf("workflow path=%q", gotPath)
	}
	if received["ref_image_0"] != "https://yzcm.site/assets/empty-1x1.png" {
		t.Fatalf("text workflow must still submit the fallback image: %+v", received)
	}
}

func TestAutoDLH3VideoAdapterInjectsFallbackImageForReferenceWorkflow(t *testing.T) {
	const fallback = "https://yzcm.site/assets/empty-1x1.png"
	var received map[string]any
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"id":"task-1"}`))
	}))
	defer server.Close()

	adapter := &AutoDLH3VideoAdapter{
		CreateURL: server.URL + "/{workflow}", TasksURL: server.URL + "/result/{id}",
		APIKey: "fake-key", Model: AutoDLH3Model,
		WorkflowID:                AutoDLH3ReferenceWorkflow,
		FallbackReferenceImageURL: fallback,
		ValidateURL:               func(raw string) (*url.URL, error) { return url.Parse(raw) },
	}
	if _, err := adapter.Submit(context.Background(), FrozenVideoModel{ID: AutoDLH3Model, MaxDuration: 15}, FinalPrompt{CompiledPrompt: "prompt"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotPath, AutoDLH3ReferenceWorkflow) {
		t.Fatalf("workflow path=%q", gotPath)
	}
	if received["ref_image_0"] != fallback {
		t.Fatalf("reference workflow must receive the fallback image, payload=%+v", received)
	}
}

func TestAutoDLH3VideoAdapterKeepsProvidedReferenceImages(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"id":"task-1"}`))
	}))
	defer server.Close()

	adapter := &AutoDLH3VideoAdapter{
		CreateURL: server.URL + "/{workflow}", TasksURL: server.URL + "/result/{id}",
		APIKey: "fake-key", Model: AutoDLH3Model,
		WorkflowID:                AutoDLH3ReferenceWorkflow,
		FallbackReferenceImageURL: "https://yzcm.site/assets/empty-1x1.png",
		ValidateURL:               func(raw string) (*url.URL, error) { return url.Parse(raw) },
	}
	prompt := FinalPrompt{CompiledPrompt: "prompt", ReferenceImageURLs: []string{"https://cdn.example.com/hero.png"}}
	if _, err := adapter.Submit(context.Background(), FrozenVideoModel{ID: AutoDLH3Model, MaxDuration: 15}, prompt); err != nil {
		t.Fatal(err)
	}
	if received["ref_image_0"] != "https://cdn.example.com/hero.png" {
		t.Fatalf("provided reference image must win over the fallback, payload=%+v", received)
	}
}

// Reference images that arrive through settings (not ReferenceImageURLs) must
// still be mapped to ref_image_0; otherwise the provider rejects the payload
// with "缺少必填参数: ref_image_0".
func TestAutoDLH3VideoAdapterMapsSettingsImagesToRefImage0(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"id":"task-1"}`))
	}))
	defer server.Close()

	adapter := &AutoDLH3VideoAdapter{
		CreateURL: server.URL + "/{workflow}", TasksURL: server.URL + "/result/{id}",
		APIKey: "fake-key", Model: AutoDLH3Model,
		WorkflowID:                AutoDLH3ReferenceWorkflow,
		FallbackReferenceImageURL: "https://yzcm.site/assets/empty-1x1.png",
		ValidateURL:               func(raw string) (*url.URL, error) { return url.Parse(raw) },
	}
	prompt := FinalPrompt{
		CompiledPrompt: "prompt",
		EffectiveSettings: EffectiveSettings{Values: SettingsPatch{
			"imageUrls": json.RawMessage(`["https://cdn.example.com/settings-hero.png"]`),
		}},
	}
	if _, err := adapter.Submit(context.Background(), FrozenVideoModel{ID: AutoDLH3Model, MaxDuration: 15}, prompt); err != nil {
		t.Fatal(err)
	}
	if received["ref_image_0"] != "https://cdn.example.com/settings-hero.png" {
		t.Fatalf("settings imageUrls must map to ref_image_0, payload=%+v", received)
	}
}
