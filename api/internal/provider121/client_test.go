package provider121

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestClientFetchSuccessPreservesBookInfo(t *testing.T) {
	var gotQuery url.Values
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"msg":"ok","data":"第一章\n正文","bookinfo":{"book_name":"港岛雨停，再无爱意","category":"男生生活","genre":8}}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{Endpoint: server.URL + "/api.php", HTTPClient: server.Client(), Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	result, err := client.Fetch(context.Background(), Request{BookID: "7673480334440139800", PlatformID: "2", MaxText: 4000})
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if result.Text != "第一章\n正文" {
		t.Fatalf("Text = %q", result.Text)
	}
	if result.BookInfo.BookName != "港岛雨停，再无爱意" || result.BookInfo.Category != "男生生活" {
		t.Fatalf("BookInfo = %+v", result.BookInfo)
	}
	if result.BookInfo.Genre != float64(8) {
		t.Fatalf("Genre = %#v, want 8", result.BookInfo.Genre)
	}
	if gotQuery.Get("bookid") != "7673480334440139800" || gotQuery.Get("platform") != "2" || gotQuery.Get("max_txt") != "4000" {
		t.Fatalf("query = %v", gotQuery)
	}
}

func TestClientFetchSupportsWorkTitle(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"data":"正文","bookinfo":{"work_title":"测试作品","category":"女生言情"}}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{Endpoint: server.URL + "/api.php", HTTPClient: server.Client(), Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Fetch(context.Background(), Request{BookID: "1", PlatformID: "15", MaxText: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if result.BookInfo.BookName != "测试作品" {
		t.Fatalf("BookName = %q", result.BookInfo.BookName)
	}
}

func TestNewClientRejectsNonHTTPS(t *testing.T) {
	_, err := NewClient(Config{Endpoint: "http://txt.121w.com/api.php"})
	if err == nil || !strings.Contains(err.Error(), "https") {
		t.Fatalf("err = %v, want HTTPS validation error", err)
	}
}

func TestClientFetchRejectsNon2xx(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	defer server.Close()
	client, _ := NewClient(Config{Endpoint: server.URL, HTTPClient: server.Client(), Timeout: time.Second})

	_, err := client.Fetch(context.Background(), Request{BookID: "1", PlatformID: "2", MaxText: 1000})
	var remote *RemoteError
	if !errors.As(err, &remote) || remote.HTTPStatus != http.StatusBadGateway {
		t.Fatalf("err = %#v, want RemoteError 502", err)
	}
}

func TestClientFetchRejectsInvalidJSON(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not-json"))
	}))
	defer server.Close()
	client, _ := NewClient(Config{Endpoint: server.URL, HTTPClient: server.Client(), Timeout: time.Second})

	_, err := client.Fetch(context.Background(), Request{BookID: "1", PlatformID: "2", MaxText: 1000})
	if err == nil || !strings.Contains(err.Error(), "JSON") {
		t.Fatalf("err = %v, want invalid JSON error", err)
	}
}

func TestClientFetchReturnsUpstreamApplicationError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":500,"msg":"获取失败","data":null}`))
	}))
	defer server.Close()
	client, _ := NewClient(Config{Endpoint: server.URL, HTTPClient: server.Client(), Timeout: time.Second})

	_, err := client.Fetch(context.Background(), Request{BookID: "1", PlatformID: "2", MaxText: 1000})
	var upstream *UpstreamError
	if !errors.As(err, &upstream) || upstream.Code != 500 || upstream.Message != "获取失败" {
		t.Fatalf("err = %#v", err)
	}
}

func TestClientFetchRejectsMissingText(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"msg":"ok","data":"","bookinfo":{"category":"男生生活"}}`))
	}))
	defer server.Close()
	client, _ := NewClient(Config{Endpoint: server.URL, HTTPClient: server.Client(), Timeout: time.Second})

	_, err := client.Fetch(context.Background(), Request{BookID: "1", PlatformID: "2", MaxText: 1000})
	if err == nil || !strings.Contains(err.Error(), "正文") {
		t.Fatalf("err = %v, want missing text error", err)
	}
}

func TestClientFetchHonorsTimeout(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(80 * time.Millisecond)
		_, _ = w.Write([]byte(`{"code":200,"data":"正文"}`))
	}))
	defer server.Close()
	client, _ := NewClient(Config{Endpoint: server.URL, HTTPClient: server.Client(), Timeout: 20 * time.Millisecond})

	_, err := client.Fetch(context.Background(), Request{BookID: "1", PlatformID: "2", MaxText: 1000})
	if err == nil {
		t.Fatal("Fetch error = nil, want timeout")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(strings.ToLower(err.Error()), "deadline") {
		t.Fatalf("err = %v, want deadline exceeded", err)
	}
}

func TestClientFetchValidatesRequestBeforeNetwork(t *testing.T) {
	client, err := NewClient(Config{Endpoint: "https://txt.121w.com/api.php", Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	cases := []Request{
		{BookID: "", PlatformID: "2", MaxText: 1000},
		{BookID: "1", PlatformID: "", MaxText: 1000},
		{BookID: "1", PlatformID: "2", MaxText: 99},
		{BookID: "1", PlatformID: "2", MaxText: 100001},
	}
	for _, tc := range cases {
		if _, err := client.Fetch(context.Background(), tc); err == nil {
			t.Fatalf("Fetch(%+v) error = nil", tc)
		}
	}
}
