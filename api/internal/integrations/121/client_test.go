package one21

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchBookBuilds121QueryAndDecodesBookInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("bookid"); got != "7673480334440139800" {
			t.Fatalf("bookid = %q", got)
		}
		if got := r.URL.Query().Get("platform"); got != "2" {
			t.Fatalf("platform = %q", got)
		}
		if got := r.URL.Query().Get("max_txt"); got != "2000" {
			t.Fatalf("max_txt = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"msg":"获取章节内容成功","data":"正文","bookinfo":{"book_id":"7673480334440139800","book_name":"港岛雨停，再无爱意","category":"男生生活","genre":8}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.FetchBook(context.Background(), "7673480334440139800", "2", 2000)
	if err != nil {
		t.Fatal(err)
	}
	if result.Data != "正文" || result.BookInfo.Category != "男生生活" || result.BookInfo.Genre != 8 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestFetchBookRejectsUpstreamErrorCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":500,"msg":"抓取失败","data":""}`))
	}))
	defer server.Close()

	_, err := NewClient(server.URL).FetchBook(context.Background(), "123", "2", 2000)
	if err == nil {
		t.Fatal("expected upstream error")
	}
}

func TestFetchBookRejectsEmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"msg":"ok","data":"","bookinfo":{"category":"男生生活"}}`))
	}))
	defer server.Close()

	_, err := NewClient(server.URL).FetchBook(context.Background(), "123", "2", 2000)
	if err == nil {
		t.Fatal("expected empty body error")
	}
}

func TestFetchBookRejectsInvalidArguments(t *testing.T) {
	client := NewClient("http://example.invalid")
	for _, tc := range []struct {
		bookID   string
		platform string
		maxTxt   int
	}{
		{"", "2", 2000},
		{"123", "", 2000},
		{"123", "2", 0},
	} {
		if _, err := client.FetchBook(context.Background(), tc.bookID, tc.platform, tc.maxTxt); err == nil {
			t.Fatalf("expected invalid argument error for %#v", tc)
		}
	}
}
