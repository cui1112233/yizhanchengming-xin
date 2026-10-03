package metadataai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/novel"
	workerpkg "github.com/cui1112233/yizhanchengming-xin/api/internal/worker"
)

func TestClassifierUsesOpenAICompatibleProtocol(t *testing.T) {
	var gotAuth string
	var gotModel string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		var body struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		gotModel = body.Model
		if len(body.Messages) != 2 || body.Messages[0].Role != "system" || body.Messages[1].Role != "user" {
			t.Fatalf("unexpected messages: %#v", body.Messages)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"gender\":\"男\",\"style\":\"男频都市\"}"}}]}`))
	}))
	defer server.Close()

	classifier, err := New(Config{
		Endpoint: server.URL,
		APIKey:   "secret",
		Model:    "legacy-text-model",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	got, err := classifier.Classify(context.Background(), workerpkg.IntakeBook{
		BookID: "7673480334440139800",
		SourceText: "这是一段用于判断小说频道与风格的正文。",
	})
	if err != nil {
		t.Fatalf("Classify: %v", err)
	}
	if got.Gender != novel.GenderMale || got.Style != "男频都市" {
		t.Fatalf("unexpected classification: %#v", got)
	}
	if gotAuth != "Bearer secret" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	if gotModel != "legacy-text-model" {
		t.Fatalf("model = %q", gotModel)
	}
}

func TestClassifierRejectsUnknownStyle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"gender\":\"女\",\"style\":\"不存在的风格\"}"}}]}`))
	}))
	defer server.Close()

	classifier, err := New(Config{Endpoint: server.URL, Model: "m"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := classifier.Classify(context.Background(), workerpkg.IntakeBook{SourceText: "正文"}); err == nil {
		t.Fatal("expected invalid style error")
	}
}

func TestClassifierAcceptsJSONCodeFence(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"choices\":[{\"message\":{\"content\":\"```json\\n{\\\"gender\\\":\\\"女\\\",\\\"style\\\":\\\"现代女主\\\"}\\n```\"}}]}"))
	}))
	defer server.Close()

	classifier, err := New(Config{Endpoint: server.URL, Model: "m"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	got, err := classifier.Classify(context.Background(), workerpkg.IntakeBook{SourceText: "正文"})
	if err != nil {
		t.Fatalf("Classify: %v", err)
	}
	if got.Gender != novel.GenderFemale || got.Style != "现代女主" {
		t.Fatalf("unexpected classification: %#v", got)
	}
}
