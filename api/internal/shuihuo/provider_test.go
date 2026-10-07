package shuihuo

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestHTTPProviderUsesServerSideCredentialAndNormalizesMediaResponse(t *testing.T) {
	roundTrip := roundTripper(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/generate" || r.Header.Get("Authorization") != "Bearer server-only-key" {
			t.Fatalf("path=%s authorization=%q", r.URL.Path, r.Header.Get("Authorization"))
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["kind"] != "image" || body["model"] != "image-v1" || body["prompt"] != "a quiet alley" {
			t.Fatalf("body=%v", body)
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"contentUrl":"https://provider.example/output.png","contentType":"image/png; charset=binary"}`))}, nil
	})
	provider := NewHTTPProvider(HTTPProviderConfig{BaseURL: "https://provider.test", APIKey: "server-only-key", ImageModel: "image-v1", Client: &http.Client{Transport: roundTrip}})
	if !provider.Available(MediaImage) || provider.Available(MediaAudio) {
		t.Fatal("unexpected provider availability")
	}
	output, err := provider.Generate(context.Background(), MediaTask{Kind: MediaImage}, "a quiet alley")
	if err != nil || output.URL != "https://provider.example/output.png" || output.ContentType != "image/png" {
		t.Fatalf("output=%+v err=%v", output, err)
	}
}

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestHTTPProviderDoesNotPretendUnavailableKindsCanRun(t *testing.T) {
	provider := NewHTTPProvider(HTTPProviderConfig{BaseURL: "https://provider.example", APIKey: "key", ImageModel: "image-v1"})
	if _, err := provider.Generate(context.Background(), MediaTask{Kind: MediaAudio}, "speak this"); err != ErrProviderUnavailable {
		t.Fatalf("err=%v", err)
	}
}
