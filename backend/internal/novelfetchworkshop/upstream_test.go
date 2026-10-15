package novelfetchworkshop

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestUpstreamFetcherFetchAndProcessPreservesRawAndProcessesText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"msg":"成功","data":"春风又绿江南岸","bookinfo":{"category":"男频"}}`))
	}))
	defer server.Close()

	result, err := (UpstreamFetcher{Client: server.Client(), Endpoint: server.URL}).FetchAndProcess(context.Background(), "book-123", "15", 4)
	if err != nil {
		t.Fatal(err)
	}
	if result.OriginalRaw != "春风又绿江南岸" || result.ProcessedText != "春风又绿" || result.BookInfo.Category != "男频" {
		t.Fatalf("result = %+v", result)
	}
	if result.OriginalRawChars != 7 || result.ProcessedChars != 4 || result.MaxTxt != 4 {
		t.Fatalf("counts = %+v", result)
	}
}

func TestUpstreamFetcherFetchAndProcessZeroAndNegativeMaxTxt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"data":"正文"}`))
	}))
	defer server.Close()

	for _, maxTxt := range []int{0, -1} {
		result, err := (UpstreamFetcher{Client: server.Client(), Endpoint: server.URL}).FetchAndProcess(context.Background(), "book-123", "15", maxTxt)
		if err != nil {
			t.Fatal(err)
		}
		if result.ProcessedText != "" || result.ProcessedChars != 0 || result.MaxTxt != 0 {
			t.Fatalf("maxTxt=%d result = %+v", maxTxt, result)
		}
	}
}

func TestUpstreamFetchSendsRequiredQueryParameters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if got := query.Get("bookid"); got != "book-123" {
			t.Fatalf("bookid = %q, want book-123", got)
		}
		if got := query.Get("platform"); got != "15" {
			t.Fatalf("platform = %q, want 15", got)
		}
		if got := query.Get("max_txt"); got != "4000" {
			t.Fatalf("max_txt = %q, want 4000", got)
		}
		_, _ = w.Write([]byte(`{"code":200,"data":"正文"}`))
	}))
	defer server.Close()

	result, err := (UpstreamFetcher{Client: server.Client(), Endpoint: server.URL}).Fetch(context.Background(), "book-123", "15", 4000)
	if err != nil {
		t.Fatal(err)
	}
	if result.OriginalRaw != "正文" {
		t.Fatalf("OriginalRaw = %q, want 正文", result.OriginalRaw)
	}
}

func TestUpstreamFetchParsesBodyAndBookInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"msg":"成功","data":"完整正文","bookinfo":{"category":"男频","genre":"都市"}}`))
	}))
	defer server.Close()

	result, err := (UpstreamFetcher{Client: server.Client(), Endpoint: server.URL}).Fetch(context.Background(), "book-123", "15", 4000)
	if err != nil {
		t.Fatal(err)
	}
	if result.Code != 200 || result.Message != "成功" || result.OriginalRaw != "完整正文" || result.BookInfo != (BookInfo{Category: "男频", Genre: "都市"}) {
		t.Fatalf("result = %#v", result)
	}
}

func TestUpstreamFetchParsesNestedThemeAndLabels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"msg":"成功","data":"完整正文","bookinfo":{"work_title":"测试书","theme":[{"id":1,"name":"现实情感"},{"id":2,"name":"婚恋"}],"labels":[{"id":75,"name":"现代"}]}}`))
	}))
	defer server.Close()

	result, err := (UpstreamFetcher{Client: server.Client(), Endpoint: server.URL}).Fetch(context.Background(), "book-123", "15", 4000)
	if err != nil {
		t.Fatal(err)
	}
	if result.BookInfo.WorkTitle != "测试书" || result.BookInfo.Themes != "现实情感、婚恋" || result.BookInfo.Labels != "现代" {
		t.Fatalf("nested bookinfo = %#v", result.BookInfo)
	}
}

func TestUpstreamFetchRejectsBusinessErrorWithoutBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":403,"msg":"无权访问","data":"错误页面"}`))
	}))
	defer server.Close()

	result, err := (UpstreamFetcher{Client: server.Client(), Endpoint: server.URL}).Fetch(context.Background(), "book-123", "15", 4000)
	if err == nil || result.OriginalRaw != "" {
		t.Fatalf("result = %#v, err = %v; want business error without body", result, err)
	}
}

func TestUpstreamFetchRejectsNonJSONWithoutBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("upstream unavailable"))
	}))
	defer server.Close()

	result, err := (UpstreamFetcher{Client: server.Client(), Endpoint: server.URL}).Fetch(context.Background(), "book-123", "15", 4000)
	if err == nil || result.OriginalRaw != "" {
		t.Fatalf("result = %#v, err = %v; want invalid JSON error without body", result, err)
	}
}

func TestUpstreamFetchRejectsHTTP500WithoutBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"code":200,"data":"不应保存"}`))
	}))
	defer server.Close()

	result, err := (UpstreamFetcher{Client: server.Client(), Endpoint: server.URL}).Fetch(context.Background(), "book-123", "15", 4000)
	if err == nil || !strings.Contains(err.Error(), "HTTP status 500") || result.OriginalRaw != "" {
		t.Fatalf("result = %#v, err = %v; want HTTP 500 error without body", result, err)
	}
}

func TestUpstreamFetchRejectsNetworkFailureWithoutBody(t *testing.T) {
	transport := roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("network failure")
	})
	client := &http.Client{Transport: transport}

	result, err := (UpstreamFetcher{Client: client, Endpoint: "http://upstream.invalid"}).Fetch(context.Background(), "book-123", "15", 4000)
	if err == nil || !strings.Contains(err.Error(), "request failed") || result.OriginalRaw != "" {
		t.Fatalf("result = %#v, err = %v; want request failed error without body", result, err)
	}
}

func TestUpstreamFetchReturnsTimeoutErrorWithoutBody(t *testing.T) {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		<-req.Context().Done()
		return nil, req.Context().Err()
	})
	client := &http.Client{Transport: transport, Timeout: 10 * time.Millisecond}

	result, err := (UpstreamFetcher{Client: client, Endpoint: "http://upstream.invalid"}).Fetch(context.Background(), "book-123", "15", 4000)
	if !errors.Is(err, context.DeadlineExceeded) || result.OriginalRaw != "" {
		t.Fatalf("result = %#v, err = %v; want timeout without body", result, err)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
