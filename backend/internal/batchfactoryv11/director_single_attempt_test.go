package batchfactoryv11

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestNovelSingleAttemptHTTPFailures(t *testing.T) {
	for _, status := range []int{400, 429, 500, 307} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			var count atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count.Add(1)
				w.Header().Set("Location", "/redirect")
				w.WriteHeader(status)
				fmt.Fprint(w, `{"error":"unsupported response_format"}`)
			}))
			defer server.Close()
			p := &OpenAICompatibleProvider{Endpoint: server.URL, APIKey: "local-test", Model: "test", Client: server.Client()}
			_, err := p.Complete(WithSingleAttempt(context.Background()), TextCompletionRequest{MaxTokens: 20000})
			if err == nil {
				t.Fatal("expected failure")
			}
			if got := count.Load(); got != 1 {
				t.Fatalf("novel sent %d HTTP requests", got)
			}
		})
	}
}

type novelAttemptTransport struct {
	calls    int
	response *http.Response
	err      error
}

func (t *novelAttemptTransport) RoundTrip(*http.Request) (*http.Response, error) {
	t.calls++
	return t.response, t.err
}

type novelBrokenBody struct{}

func (novelBrokenBody) Read([]byte) (int, error) {
	return 0, errors.New("connection reset during response")
}
func (novelBrokenBody) Close() error { return nil }
func TestNovelSingleAttemptTransportReadAndProtocolFailures(t *testing.T) {
	for _, kind := range []string{"transport", "read", "http2"} {
		t.Run(kind, func(t *testing.T) {
			transport := &novelAttemptTransport{err: errors.New("connection reset")}
			if kind == "read" {
				transport.err = nil
				transport.response = &http.Response{StatusCode: 200, Body: novelBrokenBody{}}
			}
			if kind == "http2" {
				transport.err = errors.New(`net/http: HTTP/1.x transport connection broken: malformed HTTP response "\\x00\\x00"`)
			}
			fallback := &novelAttemptTransport{err: errors.New("should not retry")}
			p := &OpenAICompatibleProvider{Endpoint: "http://127.0.0.1/unused", APIKey: "test", Model: "test", Client: &http.Client{Transport: transport}, retryClient: &http.Client{Transport: fallback}}
			if _, err := p.Complete(WithSingleAttempt(context.Background()), TextCompletionRequest{}); err == nil {
				t.Fatal("expected transport failure")
			}
			if transport.calls != 1 || fallback.calls != 0 {
				t.Fatalf("resend: transport=%d fallback=%d", transport.calls, fallback.calls)
			}
		})
	}
}

func TestNovelPolicyDoesNotChangeOtherModulesRetryOrRedirect(t *testing.T) {
	var count atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := count.Add(1)
		if n == 1 {
			w.WriteHeader(400)
			fmt.Fprint(w, `{"error":"unsupported response_format"}`)
			return
		}
		fmt.Fprint(w, `{"choices":[{"message":{"content":"{}"}}]}`)
	}))
	defer server.Close()
	client := server.Client()
	p := &OpenAICompatibleProvider{Endpoint: server.URL, APIKey: "test", Model: "test", Client: client}
	if _, err := p.Complete(context.Background(), TextCompletionRequest{}); err != nil {
		t.Fatal(err)
	}
	if count.Load() != 2 {
		t.Fatalf("existing retry changed: %d", count.Load())
	}
	if client.CheckRedirect != nil {
		t.Fatal("shared client was mutated")
	}
}
