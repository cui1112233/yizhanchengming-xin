package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localexecutor"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/safehttp"
)

func TestSubjectNativePreflightRejectsLegacyBeforeSend(t *testing.T) {
	s := scriptVideoService{resolver: func(context.Context, string, string) (batchfactoryv11.VideoProviderConfig, error) {
		return batchfactoryv11.VideoProviderConfig{Provider: batchfactoryv11.VideoProviderYFAISeedance}, nil
	}}
	if err := s.ValidateInput(context.Background(), "alice", scriptVideoLocalInput{ModelKey: "seedance-2-0-official", Prompt: "@阿青", SubjectStrategy: "seedance_subject", SubjectReferenceURIs: []string{"asset://abc"}}); err == nil {
		t.Fatal("legacy protocol must reject subject references")
	}
}

func TestSubjectNativeHTTPBody(t *testing.T) {
	for _, provider := range []string{scriptVideoProviderArk, scriptVideoProviderSora} {
		t.Run(provider, func(t *testing.T) {
			var body map[string]any
			calls := 0
			s := scriptVideoService{client: &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				json.NewDecoder(r.Body).Decode(&body)
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"id":"fake-task"}`))}, nil
			})}, resolver: func(context.Context, string, string) (batchfactoryv11.VideoProviderConfig, error) {
				return batchfactoryv11.VideoProviderConfig{ConfigID: "own-config-42", Provider: provider, Model: "own-real-model", APIKey: "OWN_FAKE_KEY", CreateURL: "https://8.8.8.8/create", TasksURL: "https://8.8.8.8/tasks/{id}"}, nil
			}}
			input := scriptVideoLocalInput{ModelKey: "own-config-42", Prompt: "@Alice walks in @Park", Duration: 6, AspectRatio: "9:16", Resolution: "720p", GenerationMode: "text_to_video"}
			if provider == scriptVideoProviderArk {
				input.SubjectStrategy = "seedance_subject"
				input.SubjectReferenceURIs = []string{"asset://asset-own-123", "asset://asset-own-scene", "asset://asset-own-123"}
			} else {
				input.SubjectStrategy = "sora_text"
				input.Prompt = "OWN @asset_123 REPLACED"
			}
			if err := s.ValidateInput(context.Background(), "alice", input); err != nil {
				t.Fatal(err)
			}
			if calls != 0 {
				t.Fatal("preflight sent HTTP")
			}
			out, err := s.Submit(context.Background(), "alice", input)
			if err != nil {
				t.Fatal(err)
			}
			_, p, key := nativeScriptTaskParts(out.TaskID)
			if key != input.ModelKey || p != provider || body["model"] != "own-real-model" {
				t.Fatalf("wrong routing %s %#v", out.TaskID, body)
			}

			expected := `{"images":[],"model":"own-real-model","orientation":"portrait","prompt":"OWN @asset_123 REPLACED","size":"small","duration":6,"watermark":false}`
			if provider == scriptVideoProviderArk {
				expected = `{"model":"own-real-model","content":[{"type":"text","text":"@Alice walks in @Park"},{"type":"image_url","image_url":{"url":"asset://asset-own-123"},"role":"reference_image"},{"type":"image_url","image_url":{"url":"asset://asset-own-scene"},"role":"reference_image"}],"ratio":"9:16","duration":6,"resolution":"720p","watermark":false}`
			}
			var want map[string]any
			json.Unmarshal([]byte(expected), &want)
			recordSubjectNativeEvidence(t, provider+"-enabled.json", body)
			if !reflect.DeepEqual(body, want) {
				t.Fatalf("original fixed-input request mismatch got=%#v want=%#v", body, want)
			}
			if provider == scriptVideoProviderArk {
				content := body["content"].([]any)
				if len(content) != 3 || content[0].(map[string]any)["text"] != input.Prompt {
					t.Fatalf("wrong content %#v", content)
				}
				for i, want := range []string{"asset://asset-own-123", "asset://asset-own-scene"} {
					item := content[i+1].(map[string]any)
					if item["role"] != "reference_image" || item["image_url"].(map[string]any)["url"] != want {
						t.Fatalf("wrong ref %#v", item)
					}
				}
			} else if body["prompt"] != input.Prompt || len(body["images"].([]any)) != 0 || body["orientation"] != "portrait" {
				t.Fatalf("wrong sora %#v", body)
			}

			if provider == scriptVideoProviderArk {
				input.SubjectReferenceURIs = nil
				input.SubjectStrategy = ""
				if _, err := s.Submit(context.Background(), "alice", input); err != nil {
					t.Fatal(err)
				}
				recordSubjectNativeEvidence(t, provider+"-disabled.json", body)
				content := body["content"].([]any)
				if len(content) != 1 || content[0].(map[string]any)["text"] != input.Prompt {
					t.Fatalf("disabled subject mode changed original body %#v", body)
				}
			}
			if provider == scriptVideoProviderSora {
				input.SubjectStrategy = ""
				input.Prompt = "OWN @Alice RAW"
				if _, err := s.Submit(context.Background(), "alice", input); err != nil {
					t.Fatal(err)
				}
				want["prompt"] = input.Prompt
				if !reflect.DeepEqual(body, want) {
					t.Fatalf("Sora off mode changed more than prompt %#v", body)
				}
				recordSubjectNativeEvidence(t, provider+"-disabled.json", body)
			}

		})
	}
}
func TestSubjectNativeInvalidPreflight(t *testing.T) {
	s := scriptVideoService{resolver: func(context.Context, string, string) (batchfactoryv11.VideoProviderConfig, error) {
		return batchfactoryv11.VideoProviderConfig{ConfigID: "own", Provider: scriptVideoProviderArk, Model: "own", APIKey: "fake", CreateURL: "https://8.8.8.8/create", TasksURL: "https://8.8.8.8/tasks/{id}"}, nil
	}}
	for _, refs := range [][]string{{"asset://{}"}, {"asset://a/b"}, {"http://cdn.example/a.png"}, {"https://127.0.0.1/a"}, []string{"asset://asset-a", "asset://asset-b", "asset://asset-c", "asset://asset-d", "asset://asset-e", "asset://asset-f", "asset://asset-g", "asset://asset-h", "asset://asset-i", "asset://asset-j"}} {
		if err := s.ValidateInput(context.Background(), "alice", scriptVideoLocalInput{ModelKey: "own", Prompt: "raw", SubjectStrategy: "seedance_subject", SubjectReferenceURIs: refs}); err == nil {
			t.Fatalf("invalid refs accepted %#v", refs)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if s.ValidateInput(ctx, "alice", scriptVideoLocalInput{ModelKey: "own", Prompt: "raw"}) == nil {
		t.Fatal("canceled accepted")
	}
}
func TestSubjectNativeConfiguredCatalogID(t *testing.T) {
	config := map[string]any{"modelCatalog": []any{map[string]any{"id": "own-ark", "kind": "video", "provider": scriptVideoProviderArk, "modelId": "doubao-own-model", "endpoint": "https://8.8.8.8/tasks", "tasksEndpoint": "https://8.8.8.8/tasks/{id}", "credential": "fake"}}}
	cfg, err := scriptVideoModelConfig(context.Background(), nil, "alice", config, "own-ark")
	if err != nil || cfg.Model != "doubao-own-model" || cfg.Provider != scriptVideoProviderArk {
		t.Fatalf("cfg %#v err %v", cfg, err)
	}
}

func TestSubjectNativeRouterRegistrationSmoke(t *testing.T) {
	for _, withCatalog := range []bool{false, true} {
		options := RouterOptions{BridgeSecret: "secret", LocalExecutors: localexecutor.NewService(localexecutor.NewMemoryStore(), time.Now)}
		if withCatalog {
			options.NovelFetchStore = novelfetchworkshop.NewMemoryStore()
		}
		handler := NewRouter(options)
		req := httptest.NewRequest(http.MethodPost, "/api/script-video", strings.NewReader(`{}`))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("catalog=%v status=%d", withCatalog, w.Code)
		}
	}
}

func TestSubjectNativeHTTPFailureSafety(t *testing.T) {
	for _, kind := range []string{"redirect", "oversize", "error-body", "bad-json"} {
		t.Run(kind, func(t *testing.T) {
			calls := 0
			a := nativeScriptVideoAdapter{cfg: batchfactoryv11.VideoProviderConfig{Provider: scriptVideoProviderArk, Model: "own", APIKey: "PRIVATE_KEY", CreateURL: "https://8.8.8.8/create", TasksURL: "https://8.8.8.8/tasks/{id}"}, Client: &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				status := 200
				body := "PRIVATE_KEY"
				headers := make(http.Header)
				switch kind {
				case "redirect":
					status = 302
					headers.Set("Location", "https://1.1.1.1/stolen")
				case "oversize":
					body = strings.Repeat("x", (1<<20)+1)
				case "error-body":
					status = 401
				}
				return &http.Response{StatusCode: status, Header: headers, Body: io.NopCloser(strings.NewReader(body))}, nil
			})}}
			_, err := a.do(context.Background(), http.MethodPost, a.cfg.CreateURL, map[string]string{"prompt": "private"})
			if err == nil || strings.Contains(err.Error(), "PRIVATE_KEY") || calls != 1 {
				t.Fatalf("unsafe response kind=%s err=%v calls=%d", kind, err, calls)
			}
		})
	}
}

func TestSubjectNativeTaskKeepsActualCatalogIDFromModelAlias(t *testing.T) {
	config := map[string]any{"modelCatalog": []any{map[string]any{"id": "actual-catalog-entry", "kind": "video", "provider": scriptVideoProviderArk, "modelId": "own-deployment-id", "endpoint": "https://8.8.8.8/create", "tasksEndpoint": "https://8.8.8.8/tasks/{id}", "credential": "fake"}}}
	var resolved []string
	s := scriptVideoService{resolver: func(ctx context.Context, owner, key string) (batchfactoryv11.VideoProviderConfig, error) {
		resolved = append(resolved, key)
		return scriptVideoModelConfig(ctx, nil, owner, config, key)
	}, client: &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"id":"fake-task"}`))}, nil
	})}}
	out, err := s.Submit(context.Background(), "alice", scriptVideoLocalInput{ModelKey: "own-deployment-id", Prompt: "raw"})
	if err != nil {
		t.Fatal(err)
	}
	_, _, key := nativeScriptTaskParts(out.TaskID)
	if key != "actual-catalog-entry" {
		t.Fatalf("wrong task configuration %s", out.TaskID)
	}
	// Poll resolves the original catalog entry even if the model alias has changed.
	_, _ = s.Poll(context.Background(), "alice", out.TaskID)
	if resolved[len(resolved)-1] != "actual-catalog-entry" {
		t.Fatalf("wrong poll resolution %#v", resolved)
	}
	config["modelCatalog"] = []any{}
	if _, err = s.Poll(context.Background(), "alice", out.TaskID); err == nil {
		t.Fatal("deleted model silently polled another model")
	}
}

func TestSubjectNativeServicePreservesFankePollConfiguration(t *testing.T) {
	selected := ""
	s := scriptVideoService{resolver: func(_ context.Context, _ string, key string) (batchfactoryv11.VideoProviderConfig, error) {
		selected = key
		return batchfactoryv11.VideoProviderConfig{Provider: batchfactoryv11.VideoProviderYFAISeedance, Model: "fanke-model", CreateURL: "https://yf.token6688.com", APIKey: "fake"}, nil
	}, client: &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/tasks/fanke-task" {
			t.Fatalf("wrong task path %s", r.URL.Path)
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"status":"completed","output_url":"https://cdn.example/fanke.mp4"}`))}, nil
	})}}
	result, err := s.Poll(context.Background(), "alice", "yfai:fanke-video-api:fanke-task")
	if err != nil || selected != "fanke-video-api" || result.Status != "succeeded" {
		t.Fatalf("wrong fanke routing selected=%s status=%#v err=%v", selected, result, err)
	}
}

func TestSubjectNativeRouteValidatesAndForwardsSpecializedReferences(t *testing.T) {
	now := time.Unix(100, 0)
	calls := 0
	var body map[string]any
	resolver := func(context.Context, string, string) (batchfactoryv11.VideoProviderConfig, error) {
		return batchfactoryv11.VideoProviderConfig{ConfigID: "own-config", Provider: scriptVideoProviderArk, Model: "own-model", APIKey: "fake", CreateURL: "https://8.8.8.8/create", TasksURL: "https://8.8.8.8/tasks/{id}"}, nil
	}
	client := &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		json.NewDecoder(r.Body).Decode(&body)
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"id":"fake-task"}`))}, nil
	})}
	mux := http.NewServeMux()
	RegisterScriptVideoRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, client, resolver)
	for _, payload := range []string{`{"modelKey":"own-config","prompt":"@Alice","subjectStrategy":"seedance_subject","subjectReferenceUris":["asset://asset-own-123"],"generationMode":"image_to_video"}`, `{"modelKey":"own-config","prompt":"@Alice","subjectReferenceUris":["asset://asset-own-123"]}`} {
		req := httptest.NewRequest(http.MethodPost, "/api/script-video", strings.NewReader(payload))
		SignBridgeRequest(req, "alice", false, now, "secret")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if strings.Contains(payload, `"subjectStrategy"`) {
			if w.Code != http.StatusAccepted || len(body["content"].([]any)) != 2 {
				t.Fatalf("route lost subject contract status=%d body=%s outgoing=%#v", w.Code, w.Body.String(), body)
			}
		} else if w.Code != http.StatusBadRequest || calls != 1 {
			t.Fatalf("route silently accepted unmodeled refs status=%d calls=%d", w.Code, calls)
		}
	}
}

func TestSubjectNativeSoraPollUsesObservedQueryProtocol(t *testing.T) {
	calls := 0
	a := nativeScriptVideoAdapter{cfg: batchfactoryv11.VideoProviderConfig{Provider: scriptVideoProviderSora, Model: "own", APIKey: "fake", CreateURL: "https://8.8.8.8/v1/video/create"}, Client: &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method != http.MethodGet || r.URL.Path != "/v1/video/query" || r.URL.Query().Get("id") != "OWN_TASK" || r.Body != nil {
			t.Fatalf("wrong original Sora query request %s %s", r.Method, r.URL)
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"status":"processing"}`))}, nil
	})}}
	_, _ = a.Poll(context.Background(), batchfactoryv11.FrozenVideoModel{ID: "own"}, batchfactoryv11.ProviderTaskRef{ProviderTaskID: "OWN_TASK"})
	if calls != 1 {
		t.Fatalf("observed query never sent: %d", calls)
	}
}

func TestSubjectNativeOriginalPollGoldenFixtures(t *testing.T) {
	bytes, err := os.ReadFile("testdata/script_video_native_poll.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Rows []struct {
			Provider string         `json:"provider"`
			Input    map[string]any `json:"input"`
			Status   string         `json:"status"`
			VideoURL string         `json:"video_url"`
			Error    string         `json:"error"`
		}
	}
	if err = json.Unmarshal(bytes, &fixture); err != nil {
		t.Fatal(err)
	}
	for i, row := range fixture.Rows {
		status, video, message := nativeScriptResponseFields(row.Provider, row.Input)
		if status != row.Status || video != row.VideoURL || message != row.Error {
			t.Fatalf("original poll case %d: got=(%s,%s,%s) want=(%s,%s,%s)", i, status, video, message, row.Status, row.VideoURL, row.Error)
		}
	}
}

func TestSubjectNativePollHTTPAndTerminalStates(t *testing.T) {
	cases := []struct {
		provider, reply, wantState, wantURL string
		wantError                           bool
	}{
		{scriptVideoProviderArk, `{"status":"succeeded","content":{"video_url":"https://own.invalid/video.mp4"}}`, string(batchfactoryv11.ProductionSucceeded), "https://own.invalid/video.mp4", false},
		{scriptVideoProviderSora, `{"status":"completed","video_url":"https://own.invalid/video.mp4"}`, string(batchfactoryv11.ProductionSucceeded), "https://own.invalid/video.mp4", false},
		{scriptVideoProviderArk, `{"status":"failed","error":{"message":"OWN_FAKE_KEY failure"}}`, string(batchfactoryv11.ProductionFailed), "", false},
		{scriptVideoProviderSora, `{"status":"cancelled"}`, string(batchfactoryv11.ProductionFailed), "", false},
		{scriptVideoProviderSora, `{"status":"completed","video_url":"https://127.0.0.1/private"}`, "", "", true},
		{scriptVideoProviderArk, `{"status":"succeeded","content":{"video_url":"asset://not-video"}}`, "", "", true},
		{scriptVideoProviderSora, `{"data":{"status":"completed","video_url":"https://own.invalid/video.mp4"}}`, string(batchfactoryv11.ProductionRunning), "", false},
		{scriptVideoProviderArk, `{"status":"succeeded"}`, string(batchfactoryv11.ProductionRunning), "", false},
	}
	for i, c := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			create := "https://8.8.8.8/v1/video/create"
			if c.provider == scriptVideoProviderArk {
				create = "https://8.8.8.8/api/v3/contents/generations/tasks"
			}
			calls := 0
			a := nativeScriptVideoAdapter{cfg: batchfactoryv11.VideoProviderConfig{Provider: c.provider, APIKey: "OWN_FAKE_KEY", Model: "own", CreateURL: create}, Client: &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Method != http.MethodGet {
					t.Fatal("poll did not GET")
				}
				if c.provider == scriptVideoProviderArk {
					if r.URL.Path != "/api/v3/contents/generations/tasks/OWN_TASK" || r.Header.Get("Content-Type") != "" {
						t.Fatalf("wrong Ark request %s headers=%v", r.URL, r.Header)
					}
				} else if r.URL.Path != "/v1/video/query" || r.URL.Query().Get("id") != "OWN_TASK" {
					t.Fatalf("wrong Sora request %s", r.URL)
				}
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(c.reply))}, nil
			})}}
			out, err := a.Poll(context.Background(), batchfactoryv11.FrozenVideoModel{ID: "own"}, batchfactoryv11.ProviderTaskRef{ProviderTaskID: "OWN_TASK"})
			if (err != nil) != c.wantError {
				t.Fatalf("error=%v wantError=%v", err, c.wantError)
			}
			if err == nil && (string(out.State) != c.wantState || out.MediaURL != c.wantURL || strings.Contains(out.ErrorMessage, "OWN_FAKE_KEY")) {
				t.Fatalf("wrong poll normalization %#v", out)
			}
			if calls != 1 {
				t.Fatalf("calls %d", calls)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if _, err = a.Poll(ctx, batchfactoryv11.FrozenVideoModel{}, batchfactoryv11.ProviderTaskRef{ProviderTaskID: "OWN_TASK"}); err == nil || calls != 1 {
				t.Fatal("canceled poll sent HTTP")
			}
		})
	}
}

func TestSubjectNativeSoraCreateAcceptsOnlyOriginalTopLevelID(t *testing.T) {
	for _, reply := range []string{`{"id":"OWN_TASK"}`, `{"task_id":"OWN_TASK"}`, `{"data":{"id":"OWN_TASK"}}`} {
		a := nativeScriptVideoAdapter{cfg: batchfactoryv11.VideoProviderConfig{Provider: scriptVideoProviderSora, Model: "own", APIKey: "fake", CreateURL: "https://8.8.8.8/v1/video/create"}, Client: &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(reply))}, nil
		})}}
		out, err := a.Submit(context.Background(), batchfactoryv11.FrozenVideoModel{ID: "own"}, batchfactoryv11.FinalPrompt{CompiledPrompt: "raw", DurationSeconds: 6})
		wantOK := strings.HasPrefix(reply, `{"id"`)
		if (err == nil) != wantOK || (wantOK && out.ProviderTaskID != "OWN_TASK") {
			t.Fatalf("original create shape mismatch reply=%s out=%#v err=%v", reply, out, err)
		}
	}
}
func TestSubjectNativeTasksEndpointOverride(t *testing.T) {
	for _, provider := range []string{scriptVideoProviderArk, scriptVideoProviderSora} {
		a := nativeScriptVideoAdapter{cfg: batchfactoryv11.VideoProviderConfig{Provider: provider, Model: "own", APIKey: "fake", CreateURL: "https://8.8.8.8/custom-create", TasksURL: "https://8.8.8.8/own-poll"}, Client: &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			if provider == scriptVideoProviderArk && r.URL.Path != "/own-poll/OWN_TASK" {
				t.Fatalf("override ignored %s", r.URL)
			}
			if provider == scriptVideoProviderSora && (r.URL.Path != "/own-poll" || r.URL.Query().Get("id") != "OWN_TASK") {
				t.Fatalf("override ignored %s", r.URL)
			}
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"status":"pending"}`))}, nil
		})}}
		if _, err := a.Poll(context.Background(), batchfactoryv11.FrozenVideoModel{}, batchfactoryv11.ProviderTaskRef{ProviderTaskID: "OWN_TASK"}); err != nil {
			t.Fatal(err)
		}
	}
}

func recordSubjectNativeEvidence(t *testing.T, name string, body map[string]any) {
	t.Helper()
	dir := os.Getenv("SUBJECT_VIDEO_PROTOCOL_EVIDENCE_DIR")
	if dir == "" {
		return
	}
	data, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestSubjectNativePollRejectsDeletedIDAliasCollision(t *testing.T) {
	config := map[string]any{"modelCatalog": []any{map[string]any{"id": "config-B", "kind": "video", "provider": scriptVideoProviderArk, "modelId": "config-A", "apiKey": "FAKE_B", "endpoint": "https://8.8.8.8/create", "tasksEndpoint": "https://8.8.8.8/tasks/{id}"}}}
	calls := 0
	s := scriptVideoService{resolver: func(ctx context.Context, owner, key string) (batchfactoryv11.VideoProviderConfig, error) {
		return scriptVideoModelConfig(ctx, nil, owner, config, key)
	}, client: &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("must not poll different configuration")
	})}}
	if _, err := s.Poll(context.Background(), "alice", nativeScriptTaskPrefix(scriptVideoProviderArk, "config-A")+"old-task"); err == nil || calls != 0 {
		t.Fatalf("deleted configuration used alias error=%v calls=%d", err, calls)
	}
}
func TestSubjectNativeStaticEndpointPreflightAlwaysRejectsPrivate(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "[::1]", "192.168.1.5", "100.100.100.200", "localhost", "LOCALHOST.", "host.localhost"} {
		cfg := batchfactoryv11.VideoProviderConfig{ConfigID: "own", Provider: scriptVideoProviderArk, Model: "own", APIKey: "fake", CreateURL: "https://" + host + "/create", TasksURL: "https://" + host + "/tasks/{id}"}
		s := scriptVideoService{resolver: func(context.Context, string, string) (batchfactoryv11.VideoProviderConfig, error) { return cfg, nil }}
		if s.ValidateInput(context.Background(), "alice", scriptVideoLocalInput{ModelKey: "own", Prompt: "raw"}) == nil {
			t.Fatalf("private endpoint %s accepted in preflight", host)
		}
	}
}

func TestSubjectNativeDefinitePreSendErrorsRemainDistinctFromUnknown(t *testing.T) {
	cases := []struct {
		name      string
		cause     error
		preflight bool
	}{{"blocked-address", safehttp.ErrBlockedAddress, true}, {"blocked-scheme", safehttp.ErrBlockedScheme, true}, {"resolve-failed", safehttp.ErrResolveFailed, true}, {"network-unknown", errors.New("connection reset after write"), false}, {"redirect-after-send", safehttp.ErrTooManyRedirects, false}}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a := nativeScriptVideoAdapter{cfg: batchfactoryv11.VideoProviderConfig{Provider: scriptVideoProviderArk, APIKey: "fake"}, Client: &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) { return nil, c.cause })}}
			_, err := a.do(context.Background(), http.MethodPost, "https://8.8.8.8/create", map[string]any{"prompt": "raw"})
			var preflight *scriptVideoPreflightError
			if err == nil || errors.As(err, &preflight) != c.preflight {
				t.Fatalf("wrong certainty error=%v preflight=%v", err, c.preflight)
			}
		})
	}
}

func TestSubjectNativeMissingConfigIDNeverCreatesUnpollableTask(t *testing.T) {
	calls := 0
	s := scriptVideoService{resolver: func(context.Context, string, string) (batchfactoryv11.VideoProviderConfig, error) {
		return batchfactoryv11.VideoProviderConfig{Provider: scriptVideoProviderArk, Model: "own", APIKey: "fake", CreateURL: "https://8.8.8.8/create", TasksURL: "https://8.8.8.8/tasks/{id}"}, nil
	}, client: &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) { calls++; return nil, errors.New("must not send") })}}
	_, err := s.Submit(context.Background(), "alice", scriptVideoLocalInput{ModelKey: "own", Prompt: "raw"})
	var preflight *scriptVideoPreflightError
	if !errors.As(err, &preflight) || calls != 0 {
		t.Fatalf("missing ID created untrackable task err=%v calls=%d", err, calls)
	}
}

func TestSubjectNativeIdentifierBoundsBeforeSendAndRoundTrip(t *testing.T) {
	calls := 0
	configID := strings.Repeat("c", nativeScriptConfigIDMaxBytes)
	providerID := strings.Repeat("t", nativeScriptTaskIDMaxBytes)
	s := scriptVideoService{resolver: func(context.Context, string, string) (batchfactoryv11.VideoProviderConfig, error) {
		return batchfactoryv11.VideoProviderConfig{ConfigID: configID, Provider: scriptVideoProviderArk, Model: "own", APIKey: "fake", CreateURL: "https://8.8.8.8/create", TasksURL: "https://8.8.8.8/tasks/{id}"}, nil
	}, client: &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		body := `{"status":"pending"}`
		if r.Method == http.MethodPost {
			b, _ := json.Marshal(map[string]string{"id": providerID})
			body = string(b)
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})}}
	out, err := s.Submit(context.Background(), "alice", scriptVideoLocalInput{ModelKey: configID, Prompt: "raw"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Poll(context.Background(), "alice", out.TaskID); err != nil {
		t.Fatalf("maximum accepted identifiers cannot poll: %v", err)
	}
	if calls != 2 {
		t.Fatalf("calls=%d", calls)
	}
	configID += "c"
	_, err = s.Submit(context.Background(), "alice", scriptVideoLocalInput{ModelKey: configID, Prompt: "raw"})
	var preflight *scriptVideoPreflightError
	if !errors.As(err, &preflight) || calls != 2 {
		t.Fatalf("oversized config ID sent request err=%v calls=%d", err, calls)
	}
}
