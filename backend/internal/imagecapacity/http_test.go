package imagecapacity

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestHTTPImageCountValidationBeforeAdmission(t *testing.T) {
	for _, tc := range []struct {
		name, count            string
		multipart, missingBody bool
		want                   int
	}{
		{name: "omitted", want: 1}, {name: "two", count: "2", want: 2},
		{name: "zero", count: "0"}, {name: "negative", count: "-1"}, {name: "too many", count: "21"},
		{name: "string", count: `"2"`}, {name: "fraction", count: "1.5"}, {name: "null", count: "null"},
		{name: "multipart omitted", multipart: true, want: 1}, {name: "multipart two", multipart: true, count: "2", want: 2},
		{name: "multipart zero", multipart: true, count: "0"}, {name: "multipart fraction", multipart: true, count: "1.5"},
		{name: "multipart too many", multipart: true, count: "21"}, {name: "missing replay", missingBody: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			payload := `{"model":"x"}`
			contentType := "application/json"
			if tc.count != "" {
				payload = `{"model":"x","n":` + tc.count + `}`
			}
			if tc.multipart {
				var data bytes.Buffer
				writer := multipart.NewWriter(&data)
				writer.WriteField("model", "x")
				writer.WriteField("prompt", strings.Repeat("long prompt ", 256))
				if tc.count != "" {
					writer.WriteField("n", tc.count)
				}
				writer.Close()
				payload, contentType = data.String(), writer.FormDataContentType()
			}
			admitted, sent := 0, 0
			transport := WrapHTTP(httpTestTransport(func(*http.Request) (*http.Response, error) {
				sent++
				return &http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader("rejected"))}, nil
			}), func(_ context.Context, r Request) (HTTPPermit, error) {
				admitted++
				if r.Count != tc.want {
					t.Errorf("count=%d want=%d", r.Count, tc.want)
				}
				return &httpTestPermit{}, nil
			})
			req, _ := http.NewRequest("POST", "https://provider.test/v1/images/edits", strings.NewReader(payload))
			req.Header.Set("Content-Type", contentType)
			if tc.missingBody {
				req.GetBody = nil
			}
			res, err := transport.RoundTrip(req)
			if tc.want == 0 {
				if err == nil || admitted != 0 || sent != 0 {
					t.Fatalf("invalid count reached admission/upstream: err=%v admitted=%d sent=%d", err, admitted, sent)
				}
				return
			}
			if err != nil || admitted != 1 || sent != 1 {
				t.Fatalf("valid request failed: %v admitted=%d sent=%d", err, admitted, sent)
			}
			io.ReadAll(res.Body)
			res.Body.Close()
		})
	}
}

func TestHTTPMalformedImagePayloadNeverReachesAdmission(t *testing.T) {
	for _, tc := range []struct{ payload, contentType string }{
		{`{"model":"x","n":`, "application/json"},
		{"--broken\r\n", "multipart/form-data; boundary=broken"},
		{"--b\r\nContent-Disposition: form-data; name=\"model\"\r\n\r\nx\r\n--b\r\nContent-Disposition: form-data; name=\"n\"\r\n\r\n20\r\n--b\r\nContent-Disposition: form-data; name=\"n\"\r\n\r\n1\r\n--b--\r\n", "multipart/form-data; boundary=b"},
		{`{"model":"x"}`, "multipart/form-data; boundary="},
	} {
		transport := WrapHTTP(httpTestTransport(func(*http.Request) (*http.Response, error) {
			t.Fatal("malformed request reached upstream")
			return nil, nil
		}), func(context.Context, Request) (HTTPPermit, error) {
			t.Fatal("malformed request reached admission")
			return nil, nil
		})
		req, _ := http.NewRequest("POST", "https://provider.test/custom", strings.NewReader(tc.payload))
		ctx, _ := BeginOperation(req.Context())
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", tc.contentType)
		if _, err := transport.RoundTrip(req); err == nil {
			t.Fatal("malformed image request accepted")
		}
	}
}

type httpTestPermit struct{ outcomes []Outcome }

func (p *httpTestPermit) Context() context.Context { return context.Background() }
func (p *httpTestPermit) Finish(o Outcome)         { p.outcomes = append(p.outcomes, o) }

type httpTestTransport func(*http.Request) (*http.Response, error)

func (f httpTestTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestHTTPAdmissionHoldsPermitThroughOperationAndDownload(t *testing.T) {
	permit := &httpTestPermit{}
	transport := WrapHTTP(httpTestTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"data":[{"url":"https://image.test/a.jpg"}]}`))}, nil
	}), func(ctx context.Context, r Request) (HTTPPermit, error) {
		if r.Model != "actual-model" || r.Size != "1024x1536" || r.Owner != "alice" {
			t.Fatalf("request=%+v", r)
		}
		return permit, nil
	})
	ctx, finish := BeginOperation(WithIdentity(context.Background(), "alice", "book-a"))
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://provider.test/v1/images/generations", strings.NewReader(`{"model":"actual-model","size":"1024x1536"}`))
	req.Header.Set("Authorization", "Bearer private-key")
	response, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(permit.outcomes) != 0 {
		t.Fatal("headers released image slot")
	}
	io.ReadAll(response.Body)
	response.Body.Close()
	if len(permit.outcomes) != 0 {
		t.Fatal("JSON read released slot before image download")
	}
	ConfirmResult(ctx)
	finish(nil)
	finish(nil)
	if len(permit.outcomes) != 1 || !permit.outcomes[0].Success {
		t.Fatalf("outcomes=%+v", permit.outcomes)
	}
}

func TestProviderTimeoutStartsAfterWaitingForCapacity(t *testing.T) {
	permit := &httpTestPermit{}
	transport := WrapHTTPWithTimeout(httpTestTransport(func(r *http.Request) (*http.Response, error) {
		time.Sleep(15 * time.Millisecond)
		if err := r.Context().Err(); err != nil {
			return nil, err
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"data":[]}`))}, nil
	}), func(context.Context, Request) (HTTPPermit, error) {
		time.Sleep(60 * time.Millisecond)
		return permit, nil
	}, 40*time.Millisecond)
	ctx, finish := BeginOperation(context.Background())
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://provider.test/v1/images/generations", strings.NewReader(`{"model":"x"}`))
	res, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	io.ReadAll(res.Body)
	res.Body.Close()
	ConfirmResult(ctx)
	finish(nil)
	if !permit.outcomes[0].Success {
		t.Fatalf("outcome=%+v", permit.outcomes)
	}
}

func TestNonImageTimeoutBudgetSpansRedirects(t *testing.T) {
	transport := WrapHTTPWithTimeout(httpTestTransport(func(r *http.Request) (*http.Response, error) {
		select {
		case <-time.After(35 * time.Millisecond):
		case <-r.Context().Done():
			return nil, r.Context().Err()
		}
		return &http.Response{StatusCode: 302, Header: http.Header{"Location": []string{"/next"}}, Body: io.NopCloser(strings.NewReader(""))}, nil
	}), func(context.Context, Request) (HTTPPermit, error) {
		t.Fatal("text entered image limiter")
		return nil, nil
	}, 55*time.Millisecond)
	first, _ := http.NewRequest("GET", "https://provider.test/text", nil)
	response, err := transport.RoundTrip(first)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	next, _ := http.NewRequest("GET", "https://provider.test/next", nil)
	next.Response = response
	if _, err := transport.RoundTrip(next); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("redirect reset total timeout: %v", err)
	}
}

func TestHTTPAdmissionRetainsUnconfirmed200Task(t *testing.T) {
	permit := &httpTestPermit{}
	transport := WrapHTTP(httpTestTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"task_id":"still-running"}`))}, nil
	}), func(context.Context, Request) (HTTPPermit, error) { return permit, nil })
	ctx, finish := BeginOperation(context.Background())
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://provider.test/v1/images/generations", strings.NewReader(`{"model":"x"}`))
	res, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	io.ReadAll(res.Body)
	res.Body.Close()
	finish(errors.New("no terminal image"))
	if len(permit.outcomes) != 1 || !permit.outcomes[0].Unknown {
		t.Fatal("unconfirmed 200 released slot")
	}
}

func TestHTTPAdmissionFailClosedAndLeavesTextUntouched(t *testing.T) {
	calls := 0
	transport := WrapHTTP(httpTestTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("ok"))}, nil
	}), func(context.Context, Request) (HTTPPermit, error) { return nil, errors.New("redis unavailable") })
	req, _ := http.NewRequest("POST", "https://provider.test/v1/images/generations", strings.NewReader(`{"model":"x"}`))
	if _, err := transport.RoundTrip(req); err == nil || calls != 0 {
		t.Fatal("Redis failure reached upstream")
	}
	text, _ := http.NewRequest("POST", "https://provider.test/v1/chat/completions", strings.NewReader(`{"model":"x"}`))
	if _, err := transport.RoundTrip(text); err != nil || calls != 1 {
		t.Fatal("text request affected")
	}
}

func TestHTTPAdmissionSeparatesRateLimitAndUncertainNetwork(t *testing.T) {
	permit := &httpTestPermit{}
	transport := WrapHTTP(httpTestTransport(func(*http.Request) (*http.Response, error) { return nil, io.ErrUnexpectedEOF }), func(context.Context, Request) (HTTPPermit, error) { return permit, nil })
	req, _ := http.NewRequest("POST", "https://provider.test/v1/images/generations", strings.NewReader(`{"model":"x"}`))
	if _, err := transport.RoundTrip(req); !errors.Is(err, ErrOutcomeUnknown) {
		t.Fatalf("error=%v", err)
	}
	if len(permit.outcomes) != 1 || !permit.outcomes[0].Unknown {
		t.Fatal("uncertain request freed occupancy")
	}
	permit = &httpTestPermit{}
	transport = WrapHTTP(httpTestTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 429, Header: http.Header{"Retry-After": []string{"9"}}, Body: io.NopCloser(strings.NewReader("limited"))}, nil
	}), func(context.Context, Request) (HTTPPermit, error) { return permit, nil })
	res, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	io.ReadAll(res.Body)
	res.Body.Close()
	if len(permit.outcomes) != 1 || !permit.outcomes[0].Limited || permit.outcomes[0].RetryAfter.Seconds() != 9 {
		t.Fatalf("outcomes=%+v", permit.outcomes)
	}
}
