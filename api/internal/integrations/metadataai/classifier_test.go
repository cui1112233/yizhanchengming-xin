package metadataai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/novel"
	workerpkg "github.com/cui1112233/yizhanchengming-xin/api/internal/worker"
)

func TestClassifierUsesLegacyV88OpenAICompatibleProtocol(t *testing.T) {
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

	classifier, err := New(Config{BaseURL: server.URL, APIKey: "secret", Model: "legacy-text-model"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	got, err := classifier.Classify(context.Background(), workerpkg.IntakeBook{
		BookID: "7673480334440139800", Category: "都市生活", Genre: 8,
		SourceText: "这是一段用于判断小说频道与风格的正文。",
	})
	if err != nil {
		t.Fatalf("Classify: %v", err)
	}
	if got.Gender != novel.GenderMale || got.Style != "男频都市" {
		t.Fatalf("unexpected classification: %#v", got)
	}
	if gotPath != "/v1/chat/completions" {
		t.Fatalf("path=%q", gotPath)
	}
	if gotAuth != "Bearer secret" {
		t.Fatalf("Authorization=%q", gotAuth)
	}
	if gotBody["model"] != "legacy-text-model" {
		t.Fatalf("model=%#v", gotBody["model"])
	}
	messages, ok := gotBody["messages"].([]any)
	if !ok || len(messages) != 2 {
		t.Fatalf("messages=%#v", gotBody["messages"])
	}
	system, _ := messages[0].(map[string]any)["content"].(string)
	if !strings.Contains(system, `{"gender":"男或女","style":"一个适合的风格"}`) {
		t.Fatalf("legacy JSON contract missing: %q", system)
	}
	if !strings.Contains(system, "古风虐文") || !strings.Contains(system, "职场打脸") {
		t.Fatalf("legacy style enum missing: %q", system)
	}
}

func TestClassifierNormalizesV1BaseURLAndAcceptsBareJSON(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"gender\":\"女\",\"style\":\"现代女主\"}"}}]}`))
	}))
	defer server.Close()

	classifier, err := New(Config{BaseURL: server.URL + "/v1", APIKey: "k", Model: "m"})
	if err != nil {
		t.Fatal(err)
	}
	got, err := classifier.Classify(context.Background(), workerpkg.IntakeBook{SourceText: "正文"})
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/v1/chat/completions" {
		t.Fatalf("path=%q", gotPath)
	}
	if got.Gender != novel.GenderFemale || got.Style != "现代女主" {
		t.Fatalf("unexpected classification: %#v", got)
	}
}

func TestClassifierRejectsInvalidLegacyValues(t *testing.T) {
	for _, content := range []string{
		`{"gender":"未知","style":"现代通用"}`,
		`{"gender":"男","style":"不存在的风格"}`,
	} {
		t.Run(content, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]any{"content": content}}}})
			}))
			defer server.Close()
			classifier, err := New(Config{BaseURL: server.URL, APIKey: "k", Model: "m"})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := classifier.Classify(context.Background(), workerpkg.IntakeBook{SourceText: "正文"}); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestNewRequiresCompleteLegacyTextConfig(t *testing.T) {
	for _, cfg := range []Config{
		{},
		{BaseURL: "https://example.com"},
		{BaseURL: "https://example.com", Model: "m"},
		{BaseURL: "https://example.com", APIKey: "k"},
	} {
		if _, err := New(cfg); err == nil {
			t.Fatalf("expected config error for %+v", cfg)
		}
	}
}
