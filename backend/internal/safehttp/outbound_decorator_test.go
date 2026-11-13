package safehttp

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type decoratorTestTransport func(*http.Request) (*http.Response, error)

func (f decoratorTestTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestOutboundDecoratorCoversAllClientConstructors(t *testing.T) {
	calls := 0
	restore := SetOutboundDecorator(func(next http.RoundTripper) http.RoundTripper {
		return decoratorTestTransport(func(r *http.Request) (*http.Response, error) {
			calls++
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("ok")), Request: r}, nil
		})
	})
	defer restore()
	clients := []*http.Client{NewClient(Options{}), NewClientWithTimeout(time.Second), NewClientWithTransport(time.Second, http.DefaultTransport.(*http.Transport)), NewClientWithTransportWithoutTimeout(http.DefaultTransport.(*http.Transport))}
	for _, client := range clients {
		req, _ := http.NewRequest("POST", "https://capacity.test/v1/images/generations", nil)
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
	}
	if calls != 4 {
		t.Fatalf("decorated calls=%d want4", calls)
	}
}

func TestManagedDecoratorOwnsProviderTimeoutAfterAdmission(t *testing.T) {
	var duration time.Duration
	restore := SetManagedOutboundDecorator(func(base http.RoundTripper, timeout time.Duration) http.RoundTripper { duration = timeout; return base })
	defer restore()
	client := NewClientWithTimeout(73 * time.Millisecond)
	if client.Timeout != 0 || duration != 73*time.Millisecond {
		t.Fatalf("outer timeout=%s managed=%s", client.Timeout, duration)
	}
}
