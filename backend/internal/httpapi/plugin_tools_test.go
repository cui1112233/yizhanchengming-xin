package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

type recordingPluginToolProvider struct {
	request batchfactoryv11.TextCompletionRequest
}

func (p *recordingPluginToolProvider) Complete(_ context.Context, request batchfactoryv11.TextCompletionRequest) (string, error) {
	p.request = request
	return "处理结果", nil
}

func TestPluginToolRouteUsesBackendOwnedPrompt(t *testing.T) {
	for _, toolID := range []string{
		"novel-to-drama", "ai-rewrite", "storyboard-master", "viral-opening", "character-edit",
	} {
		t.Run(toolID, func(t *testing.T) {
			provider := &recordingPluginToolProvider{}
			mux := http.NewServeMux()
			registerPluginToolRoutes(mux, BridgeAuth{Secret: "secret"}, provider)
			req := httptest.NewRequest(http.MethodPost, "/api/plugins/"+toolID+"/run", strings.NewReader(`{"input":"原始文本","promptId":"default"}`))
			SignBridgeRequest(req, "alice", false, time.Now(), "secret")
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "处理结果") {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
			if provider.request.SystemPrompt == "" || strings.Contains(provider.request.UserPrompt, "你是") || !strings.Contains(provider.request.UserPrompt, "原始文本") {
				t.Fatalf("request=%+v", provider.request)
			}
		})
	}
}

func TestPluginToolConfigOnlyExposesPromptNames(t *testing.T) {
	mux := http.NewServeMux()
	registerPluginToolRoutes(mux, BridgeAuth{Secret: "secret"}, &recordingPluginToolProvider{})
	req := httptest.NewRequest(http.MethodGet, "/api/plugins/viral-opening/config", nil)
	SignBridgeRequest(req, "alice", false, time.Now(), "secret")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "爆款开头") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "SystemPrompt") || strings.Contains(rec.Body.String(), "你是中文") {
		t.Fatalf("prompt body leaked: %s", rec.Body.String())
	}
}

func TestPluginCoverPromptUsesLuoshuiCorpus(t *testing.T) {
	mux := http.NewServeMux()
	registerPluginToolRoutes(mux, BridgeAuth{Secret: "secret"}, &recordingPluginToolProvider{})
	req := httptest.NewRequest(http.MethodPost, "/api/plugins/cover-generation/prompt", strings.NewReader(`{"content":"女主复仇","dramaName":"复仇千金","aspectRatio":"4:3","fontStyle":"黑体"}`))
	SignBridgeRequest(req, "alice", false, time.Now(), "secret")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, "番茄小说、七猫免费小说平台顶级封面设计专家") || !strings.Contains(body, "短剧名：复仇千金") || !strings.Contains(body, "--ar 4:3") {
		t.Fatalf("status=%d body=%s", rec.Code, body)
	}
}

func TestCoverDesignUsesTextProviderAndAppendsRatio(t *testing.T) {
	provider := &recordingPluginToolProvider{}
	mux := http.NewServeMux()
	registerPluginToolRoutes(mux, BridgeAuth{Secret: "secret"}, provider)
	req := httptest.NewRequest(http.MethodPost, "/api/plugins/cover-generation/design", strings.NewReader(`{"content":"女主复仇","dramaName":"复仇千金","aspectRatio":"3:4"}`))
	SignBridgeRequest(req, "alice", false, time.Now(), "secret")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, "处理结果") || !strings.Contains(body, "--ar 3:4") {
		t.Fatalf("status=%d body=%s", rec.Code, body)
	}
	if !strings.Contains(provider.request.SystemPrompt, "番茄小说、七猫免费小说平台顶级封面设计专家") {
		t.Fatalf("system=%s", provider.request.SystemPrompt)
	}
}

func TestPluginToolRouteRejectsUnknownToolAndEmptyInput(t *testing.T) {
	provider := &recordingPluginToolProvider{}
	mux := http.NewServeMux()
	registerPluginToolRoutes(mux, BridgeAuth{Secret: "secret"}, provider)
	for _, tc := range []struct {
		path, body string
		status     int
	}{
		{"/api/plugins/missing/run", `{"input":"文本"}`, http.StatusNotFound},
		{"/api/plugins/ai-rewrite/run", `{"input":""}`, http.StatusBadRequest},
	} {
		req := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(tc.body))
		SignBridgeRequest(req, "alice", false, time.Now(), "secret")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != tc.status {
			t.Fatalf("path=%s status=%d body=%s", tc.path, rec.Code, rec.Body.String())
		}
	}
}

func TestDocxPlainTextExtractsParagraphs(t *testing.T) {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	document, err := writer.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	document.Write([]byte(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>第一段</w:t></w:r></w:p><w:p><w:r><w:t>第二段</w:t></w:r></w:p></w:body></w:document>`))
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	text, err := docxPlainText(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if text != "第一段\n第二段" {
		t.Fatalf("text=%q", text)
	}
}
