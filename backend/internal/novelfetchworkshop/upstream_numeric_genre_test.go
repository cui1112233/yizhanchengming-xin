package novelfetchworkshop

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpstreamFetchAcceptsNumericBookInfoGenre(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"msg":"成功","data":"完整正文","bookinfo":{"category":"女频","genre":12}}`))
	}))
	defer server.Close()

	result, err := (UpstreamFetcher{Client: server.Client(), Endpoint: server.URL}).Fetch(context.Background(), "book-123", "15", 4000)
	if err != nil {
		t.Fatal(err)
	}
	if result.BookInfo != (BookInfo{Category: "女频", Genre: "12"}) {
		t.Fatalf("BookInfo = %#v", result.BookInfo)
	}
}
