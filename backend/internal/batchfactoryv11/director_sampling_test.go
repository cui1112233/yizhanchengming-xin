package batchfactoryv11

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNovelSamplingOmissionPreservesOtherDefaults(t *testing.T) {
	for _, omit := range []bool{false, true} {
		var captured map[string]any
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&captured)
			_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{}"}}]}`))
		}))
		p := &OpenAICompatibleProvider{Endpoint: server.URL, APIKey: "test", Model: "actual", Client: server.Client()}
		request := TextCompletionRequest{Temperature: .65, MaxTokens: 2000}
		raw, _ := json.Marshal(map[string]any{"OmitTemperature": omit, "OmitMaxTokens": omit})
		_ = json.Unmarshal(raw, &request)
		_, err := p.Complete(WithSingleAttempt(context.Background()), request)
		server.Close()
		if err != nil {
			t.Fatal(err)
		}
		_, temp := captured["temperature"]
		_, tokens := captured["max_tokens"]
		if temp == omit || tokens == omit {
			t.Fatalf("omit=%v body=%v", omit, captured)
		}
	}
}

func TestNovelSafeMetadataDoesNotFingerprintCredentials(t *testing.T) {
	first := NewTextModelMetadata("actual", "catalog-id", "https://user:password@example.test/v1/private-token?api_key=secret", map[string]any{"credential": "first-secret", "capabilities": map[string]any{"supportsTemperature": false, "supportsMaxTokens": true}})
	second := NewTextModelMetadata("actual", "catalog-id", "https://changed:other@example.test/changed-token?api_key=other", map[string]any{"credential": "rotated-secret", "capabilities": map[string]any{"supportsTemperature": false, "supportsMaxTokens": true}})
	if first.SafeConfigDigest != second.SafeConfigDigest {
		t.Fatal("safe configuration digest includes private credentials or URL components")
	}
	raw, _ := json.Marshal(first)
	for _, secret := range []string{"password", "private-token", "api_key", "first-secret", "example.test"} {
		if strings.Contains(string(raw), secret) {
			t.Fatalf("private configuration leaked: %s", raw)
		}
	}
	if first.SupportsTemperature == nil || *first.SupportsTemperature || first.SupportsMaxTokens == nil || !*first.SupportsMaxTokens {
		t.Fatal("advertised explicit capability lost")
	}
}
