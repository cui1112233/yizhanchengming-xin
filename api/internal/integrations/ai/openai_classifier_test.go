package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/novel"
)

func TestOpenAICompatibleClassifierUsesLegacyProtocol(t *testing.T) {
	var gotAuth string
	var gotPath string
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"### 一、分析结果\n{\"gender\":\"男\",\"style\":\"男频都市\"}"}}]}`))
	}))
	defer server.Close()

	classifier, err := NewOpenAICompatibleClassifier(Config{
		BaseURL: server.URL,
		Model:   "legacy-model",
		APIKey:  "secret-key",
	})
	if err != nil {
		t.Fatalf("new classifier: %v", err)
	}

	result, err := classifier.Classify(context.Background(), MetadataInput{
		BookID:     "7673480334440139800",
		Category:   "都市生活",
		Genre:      8,
		SourceText: "这是一段小说正文。",
	})
	if err != nil {
		t.Fatalf("classify: %v", err)
	}
	if result.Gender != novel.GenderMale || result.Style != "男频都市" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if gotPath != "/v1/chat/completions" {
		t.Fatalf("expected legacy OpenAI-compatible path, got %q", gotPath)
	}
	if gotAuth != "Bearer secret-key" {
		t.Fatalf("unexpected Authorization: %q", gotAuth)
	}
	if gotBody["model"] != "legacy-model" {
		t.Fatalf("unexpected model: %#v", gotBody["model"])
	}
	messages, ok := gotBody["messages"].([]any)
	if !ok || len(messages) != 2 {
		t.Fatalf("unexpected messages: %#v", gotBody["messages"])
	}
	system := messages[0].(map[string]any)["content"].(string)
	if !strings.Contains(system, `{"gender":"男或女","style":"一个适合的风格"}`) {
		t.Fatalf("system prompt lost legacy output contract: %q", system)
	}
	if !strings.Contains(system, "男频都市") || !strings.Contains(system, "职场打脸") {
		t.Fatalf("system prompt lost legacy style enum: %q", system)
	}
}

func TestOpenAICompatibleClassifierAcceptsBareJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"gender\":\"女\",\"style\":\"现代女主\"}"}}]}`))
	}))
	defer server.Close()

	classifier, err := NewOpenAICompatibleClassifier(Config{BaseURL: server.URL + "/v1", Model: "m", APIKey: "k"})
	if err != nil {
		t.Fatal(err)
	}
	result, err := classifier.Classify(context.Background(), MetadataInput{SourceText: "正文"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Gender != novel.GenderFemale || result.Style != "现代女主" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestOpenAICompatibleClassifierRejectsInvalidLegacyValues(t *testing.T) {
	tests := []string{
		`{"gender":"未知","style":"现代通用"}`,
		`{"gender":"男","style":"不存在的风格"}`,
	}
	for _, content := range tests {
		t.Run(content, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]any{"content": content}}}})
			}))
			defer server.Close()

			classifier, err := NewOpenAICompatibleClassifier(Config{BaseURL: server.URL, Model: "m", APIKey: "k"})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := classifier.Classify(context.Background(), MetadataInput{SourceText: "正文"}); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestNewOpenAICompatibleClassifierRequiresCompleteConfig(t *testing.T) {
	for _, cfg := range []Config{
		{},
		{BaseURL: "https://example.com"},
		{BaseURL: "https://example.com", Model: "m"},
		{BaseURL: "https://example.com", APIKey: "k"},
	} {
		if _, err := NewOpenAICompatibleClassifier(cfg); err == nil {
			t.Fatalf("expected config error for %+v", cfg)
		}
	}
}
