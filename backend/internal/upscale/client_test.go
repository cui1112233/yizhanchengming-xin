package upscale

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSubmitEnhanceBuildsOfficialRequest(t *testing.T) {
	var gotPath, gotAuth, gotContentType string
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"task_id":"amk-tool-enhance-video-123","request_id":"req-1"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "ak-test")
	taskID, err := client.SubmitEnhance(context.Background(), "mediakit://file-1", "2k")
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if taskID != "amk-tool-enhance-video-123" {
		t.Fatalf("taskID=%q", taskID)
	}
	if gotPath != "/api/v1/tools/enhance-video-generative" {
		t.Fatalf("path=%q", gotPath)
	}
	if gotAuth != "Bearer ak-test" {
		t.Fatalf("authorization=%q", gotAuth)
	}
	if gotContentType != "application/json" {
		t.Fatalf("content-type=%q", gotContentType)
	}
	if body["video_url"] != "mediakit://file-1" || body["resolution"] != "2k" {
		t.Fatalf("body=%v", body)
	}
}

func TestSubmitEnhanceRequiresConfiguration(t *testing.T) {
	client := NewClient("", "")
	if client.Configured() {
		t.Fatal("client without api key must not be configured")
	}
	if _, err := client.SubmitEnhance(context.Background(), "mediakit://file-1", "1080p"); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("err=%v", err)
	}
}

func TestSubmitEnhanceMapsUpstreamHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"success":false,"error":{"code":"AuthenticationFailed","type":"Unauthorized","message":"The API Key provided is invalid."}}`))
	}))
	defer server.Close()

	_, err := NewClient(server.URL, "bad-key").SubmitEnhance(context.Background(), "mediakit://file-1", "1080p")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err=%v", err)
	}
	if apiErr.Status != http.StatusUnauthorized || apiErr.Code != "AuthenticationFailed" {
		t.Fatalf("apiErr=%+v", apiErr)
	}
	if strings.Contains(err.Error(), "bad-key") {
		t.Fatal("error message must not leak the api key")
	}
}

func TestSubmitEnhanceMapsBusinessFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":false,"error":{"code":"InvalidParameter","message":"resolution is not supported"}}`))
	}))
	defer server.Close()

	_, err := NewClient(server.URL, "ak").SubmitEnhance(context.Background(), "mediakit://file-1", "4k")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "InvalidParameter" {
		t.Fatalf("err=%v", err)
	}
}

func TestQueryTaskMapsStates(t *testing.T) {
	cases := []struct {
		name     string
		payload  string
		status   string
		videoURL string
		errCode  string
	}{
		{
			name:    "running",
			payload: `{"success":true,"task_id":"t1","status":"running"}`,
			status:  "running",
		},
		{
			name:     "completed",
			payload:  `{"success":true,"task_id":"t1","status":"completed","result":{"video_url":"https://cdn.example.com/out.mp4"}}`,
			status:   "completed",
			videoURL: "https://cdn.example.com/out.mp4",
		},
		{
			name:    "failed",
			payload: `{"success":true,"task_id":"t1","status":"failed","error":{"code":"DownloadFailed","message":"cannot download source"}}`,
			status:  "failed",
			errCode: "DownloadFailed",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var gotPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				_, _ = w.Write([]byte(testCase.payload))
			}))
			defer server.Close()

			task, err := NewClient(server.URL, "ak").QueryTask(context.Background(), "t1")
			if err != nil {
				t.Fatalf("query: %v", err)
			}
			if gotPath != "/api/v1/tasks/t1" {
				t.Fatalf("path=%q", gotPath)
			}
			if task.Status != testCase.status || task.VideoURL != testCase.videoURL || task.ErrorCode != testCase.errCode {
				t.Fatalf("task=%+v", task)
			}
		})
	}
}

func TestTerminateBuildsTaskIDsAndReadsFailures(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/tasks/terminate" {
			t.Fatalf("path=%q", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"success":true,"result":{"terminated_tasks":[{"task_id":"t1"}],"terminate_failed_tasks":[{"task_id":"t2","error":{"code":"TaskOperationConflict","message":"task termination conflicts with its current state"}}]}}`))
	}))
	defer server.Close()

	outcome, err := NewClient(server.URL, "ak").Terminate(context.Background(), []string{"t1", "t2"})
	if err != nil {
		t.Fatalf("terminate: %v", err)
	}
	ids, _ := body["task_ids"].([]any)
	if len(ids) != 2 || ids[0] != "t1" || ids[1] != "t2" {
		t.Fatalf("body=%v", body)
	}
	if len(outcome.Terminated) != 1 || outcome.Terminated[0] != "t1" {
		t.Fatalf("terminated=%v", outcome.Terminated)
	}
	if !strings.Contains(outcome.Failed["t2"], "conflicts") {
		t.Fatalf("failed=%v", outcome.Failed)
	}
}

func TestUploadMediaUsesPresignedContract(t *testing.T) {
	var uploadHeaders http.Header
	var uploaded []byte
	var requestBody map[string]any
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/tools-sync/request-media-upload-url", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer ak" {
			t.Fatalf("unexpected request method=%s auth=%q", r.Method, r.Header.Get("Authorization"))
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &requestBody)
		_, _ = w.Write([]byte(`{"success":true,"result":{"file_id":"FILE123","method":"PUT","upload_url":"` + "http://" + r.Host + `/upload-target","upload_headers":[{"key":"Content-Type","value":"video/mp4"},{"key":"X-Tos-Token","value":"token-1"}]}}`))
	})
	mux.HandleFunc("/upload-target", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("upload method=%s", r.Method)
		}
		uploadHeaders = r.Header.Clone()
		uploaded, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(server.URL, "ak")
	target, err := client.RequestUploadURL(context.Background())
	if err != nil {
		t.Fatalf("request upload url: %v", err)
	}
	if target.FileID != "FILE123" || target.Method != http.MethodPut {
		t.Fatalf("target=%+v", target)
	}
	if len(requestBody) != 0 {
		t.Fatalf("request body must be empty object, got %v", requestBody)
	}
	if err := client.UploadMedia(context.Background(), target, strings.NewReader("video-bytes"), int64(len("video-bytes")), "video/mp4"); err != nil {
		t.Fatalf("upload: %v", err)
	}
	if uploadHeaders.Get("X-Tos-Token") != "token-1" || uploadHeaders.Get("Content-Type") != "video/mp4" {
		t.Fatalf("headers=%v", uploadHeaders)
	}
	if string(uploaded) != "video-bytes" {
		t.Fatalf("uploaded=%q", string(uploaded))
	}
	if NormalizeFileID(target.FileID) != "mediakit://FILE123" {
		t.Fatalf("normalize=%q", NormalizeFileID(target.FileID))
	}
}

func TestUploadMediaMapsHTTPFailure(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/tools-sync/request-media-upload-url", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"result":{"file_id":"mediakit://abc","method":"PUT","upload_url":"` + "http://" + r.Host + `/upload-target","upload_headers":[]}}`))
	})
	mux.HandleFunc("/upload-target", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("signature expired"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(server.URL, "ak")
	target, err := client.RequestUploadURL(context.Background())
	if err != nil {
		t.Fatalf("request upload url: %v", err)
	}
	err = client.UploadMedia(context.Background(), target, strings.NewReader("x"), 1, "video/mp4")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("err=%v", err)
	}
	if NormalizeFileID(target.FileID) != "mediakit://abc" {
		t.Fatalf("normalize must keep prefix: %q", NormalizeFileID(target.FileID))
	}
}

func TestRequestUploadURLRejectsIncompleteResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"result":{"file_id":"only-id"}}`))
	}))
	defer server.Close()

	_, err := NewClient(server.URL, "ak").RequestUploadURL(context.Background())
	var apiErr *APIError
	if !errors.As(err, &apiErr) || !strings.Contains(apiErr.Message, "上传地址") {
		t.Fatalf("err=%v", err)
	}
}
