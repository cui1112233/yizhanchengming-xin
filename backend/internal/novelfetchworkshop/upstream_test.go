package novelfetchworkshop

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestUpstreamFetcherFetchAndProcessPreservesRawAndProcessesText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("max_txt") {
			t.Fatalf("full-text fetch must not send processing limit: %s", r.URL.RawQuery)
		}
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

func TestUpstreamFetchRejectsHTMLErrorPageAsBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"msg":"ok","data":"<!DOCTYPE html><html><head><title>502 Bad Gateway</title></head><body><h1>502 Bad Gateway</h1><hr>nginx</body></html>"}`))
	}))
	defer server.Close()

	if _, err := (UpstreamFetcher{Client: server.Client(), Endpoint: server.URL}).Fetch(context.Background(), "book-123", "15", 4000); err == nil {
		t.Fatal("HTML error page must not be accepted as novel body")
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

func TestBookInfoUnmarshalAcceptsPersistedStringThemeAndLabels(t *testing.T) {
	var info BookInfo
	if err := json.Unmarshal([]byte(`{"category":"","genre":"","theme":"言情、现实情感","labels":"现代"}`), &info); err != nil {
		t.Fatal(err)
	}
	if info.Themes != "言情、现实情感" || info.Labels != "现代" {
		t.Fatalf("persisted book info = %#v", info)
	}
}

func TestBookInfoJSONRoundTripPreservesFetchedMetadata(t *testing.T) {
	want := BookInfo{
		Category:  "女频",
		Genre:     "现实情感",
		WorkTitle: "鼓励式老公，却把行动都给了闺蜜",
		Themes:    "言情、婚恋",
		Labels:    "现代",
	}
	raw, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got BookInfo
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("round trip = %#v, want %#v; json=%s", got, want, raw)
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

// 121 在输出 JSON 前会先 print PHP Warning/HTML。这里必须跳过前缀继续解析，
// 否则正常正文会被误判成 invalid JSON。
func TestUpstreamFetchSkipsHTMLWarningPrefixBeforeJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<br />
<b>Warning</b>:  Undefined array key "app_type_permission_status" in <b>/www/wwwroot/txt.121w.com/class/changduiaa.php</b> on line <b>107</b><br />
{"code":200,"msg":"获取章节内容成功","data":"正文内容","bookinfo":{"book_name":"书名"}}`))
	}))
	defer server.Close()

	result, err := (UpstreamFetcher{Client: server.Client(), Endpoint: server.URL}).Fetch(context.Background(), "book-123", "7", 4000)
	if err != nil {
		t.Fatalf("err = %v; want the JSON after the HTML warning to be parsed", err)
	}
	if result.OriginalRaw != "正文内容" || result.BookInfo.WorkTitle != "书名" {
		t.Fatalf("result = %#v; want parsed payload", result)
	}
}

// 业务失败时前缀同样存在，必须把真正的上游错误抛出来，而不是首个字符 '<'。
func TestUpstreamFetchReportsBusinessErrorBehindHTMLWarning(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<br />
<b>Warning</b>:  Undefined array key "app_type_permission_status" in <b>/www/wwwroot/txt.121w.com/class/changduiaa.php</b> on line <b>107</b><br />
书籍 7707248192760319001 数据解析失败
{"code":400,"msg":"获取书籍信息失败","data":null}`))
	}))
	defer server.Close()

	_, err := (UpstreamFetcher{Client: server.Client(), Endpoint: server.URL}).Fetch(context.Background(), "7707248192760319001", "7", 4000)
	if err == nil || !strings.Contains(err.Error(), "获取书籍信息失败") {
		t.Fatalf("err = %v; want the upstream business error behind the HTML warning", err)
	}
	if strings.Contains(err.Error(), "invalid character") {
		t.Fatalf("err = %v; must not surface the JSON decoder error", err)
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
	const secret = "example-query-secret"
	transport := roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("network failure")
	})
	client := &http.Client{Transport: transport}

	result, err := (UpstreamFetcher{Client: client, Endpoint: "http://upstream.invalid?token=" + secret}).Fetch(context.Background(), "book-123", "15", 4000)
	if err == nil || !strings.Contains(err.Error(), "request failed") || result.OriginalRaw != "" {
		t.Fatalf("result = %#v, err = %v; want request failed error without body", result, err)
	}
	if class, ok := ClassifyUpstreamFailure(err); !ok || class != UpstreamFailureTransient {
		t.Fatalf("got error %v, want a transient request failure", err)
	}
	for cause := err; cause != nil; cause = errors.Unwrap(cause) {
		if strings.Contains(cause.Error(), secret) {
			t.Fatalf("error chain exposed the endpoint token: %v", cause)
		}
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
	if class, ok := ClassifyUpstreamFailure(err); !ok || class != UpstreamFailureTransient {
		t.Fatalf("got error %v, want a transient timeout", err)
	}
}

// Reproduces the production "参数不完整" failure: missing bookid/platform must
// be rejected locally so we never call the upstream with incomplete params.
func TestUpstreamFetchRejectsMissingParametersWithoutRequest(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		_, _ = w.Write([]byte(`{"code":200,"data":"正文"}`))
	}))
	defer server.Close()

	for _, params := range [][2]string{{"", "15"}, {"book-123", ""}, {" ", "15"}} {
		result, err := (UpstreamFetcher{Client: server.Client(), Endpoint: server.URL}).Fetch(context.Background(), params[0], params[1], 4000)
		if err == nil || !strings.Contains(err.Error(), "参数不完整") || result.OriginalRaw != "" {
			t.Fatalf("params=%v result=%#v err=%v; want local incomplete-params error", params, result, err)
		}
		if class, ok := ClassifyUpstreamFailure(err); !ok || class != UpstreamFailurePermanent {
			t.Fatalf("got error %v, want a permanent input failure", err)
		}
	}
	if called {
		t.Fatal("upstream must not be called when bookid/platform is empty")
	}
}

func TestUpstreamFetchClassifiesInvalidEndpoint(t *testing.T) {
	const secret = "example-query-secret"
	fetcher := UpstreamFetcher{Endpoint: "https://upstream.invalid/%zz?token=" + secret}
	_, err := fetcher.Fetch(context.Background(), "book-123", "15", 0)
	if class, ok := ClassifyUpstreamFailure(err); !ok || class != UpstreamFailureConfiguration {
		t.Fatalf("got error %v, want a configuration failure", err)
	}
	for cause := err; cause != nil; cause = errors.Unwrap(cause) {
		if strings.Contains(cause.Error(), secret) {
			t.Fatalf("error chain exposed the endpoint token: %v", cause)
		}
	}
}

// The upstream business error must carry the real request parameters so a
// "参数不完整" response can be traced back to the book that caused it.
func TestUpstreamFetchBusinessErrorIncludesRequestParameters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":400,"msg":"参数不完整"}`))
	}))
	defer server.Close()

	_, err := (UpstreamFetcher{Client: server.Client(), Endpoint: server.URL}).Fetch(context.Background(), "book-123", "15", 0)
	if err == nil {
		t.Fatal("want business error")
	}
	for _, want := range []string{"400", "参数不完整", "bookid=book-123", "platform=15"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %q, want to contain %q", err.Error(), want)
		}
	}
}

func TestUpstreamFetchClassifiesResponseFailures(t *testing.T) {
	const secret = "example-secret-never-log"
	tests := []struct {
		name         string
		httpStatus   int
		body         string
		class        UpstreamFailureClass
		businessCode int
	}{
		{name: "HTTP 400", httpStatus: http.StatusBadRequest, body: secret, class: UpstreamFailurePermanent},
		{name: "HTTP 404", httpStatus: http.StatusNotFound, class: UpstreamFailurePermanent},
		{name: "HTTP 401", httpStatus: http.StatusUnauthorized, class: UpstreamFailureConfiguration},
		{name: "HTTP 403", httpStatus: http.StatusForbidden, class: UpstreamFailureConfiguration},
		{name: "HTTP 408", httpStatus: http.StatusRequestTimeout, class: UpstreamFailureTransient},
		{name: "HTTP 429", httpStatus: http.StatusTooManyRequests, class: UpstreamFailureTransient},
		{name: "HTTP 503", httpStatus: http.StatusServiceUnavailable, class: UpstreamFailureTransient},
		{name: "business 400", body: `{"code":400,"msg":"token=` + secret + `","data":"错误内容"}`, class: UpstreamFailurePermanent, businessCode: 400},
		{name: "invalid JSON", body: `{"token":"` + secret + `"`, class: UpstreamFailureTransient},
		{name: "HTML response", body: `<!DOCTYPE html><html>` + secret + `</html>`, class: UpstreamFailureTransient},
		{name: "HTML in data", body: `{"code":200,"data":"<html>` + secret + `</html>"}`, class: UpstreamFailureTransient},
		{name: "empty response", class: UpstreamFailureTransient},
		{name: "empty data", body: `{"code":200,"data":" "}`, class: UpstreamFailureTransient},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if tt.httpStatus != 0 {
					w.WriteHeader(tt.httpStatus)
				}
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			result, err := (UpstreamFetcher{Client: server.Client(), Endpoint: server.URL}).Fetch(context.Background(), "book-123", "15", 0)
			var upstreamErr *UpstreamError
			if !errors.As(err, &upstreamErr) {
				t.Fatalf("got result %+v and error %v, want *UpstreamError", result, err)
			}
			if upstreamErr.Class != tt.class || upstreamErr.HTTPStatus != tt.httpStatus || upstreamErr.BusinessCode != tt.businessCode {
				t.Fatalf("got %+v, want class %q, HTTP status %d, business code %d", upstreamErr, tt.class, tt.httpStatus, tt.businessCode)
			}
			if result.OriginalRaw != "" || !strings.Contains(err.Error(), "bookid=book-123") || !strings.Contains(err.Error(), "platform=15") {
				t.Fatalf("got result %+v and error %v, want an empty result and request IDs in the error", result, err)
			}
			if strings.Contains(err.Error(), secret) {
				t.Fatalf("error exposed the response secret: %v", err)
			}
		})
	}
}

func TestUpstreamFetchClassifiesReadFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "100")
		_, _ = w.Write([]byte(`{"code":200,"data":"正文"}`))
	}))
	defer server.Close()

	_, err := (UpstreamFetcher{Client: server.Client(), Endpoint: server.URL}).Fetch(context.Background(), "book-123", "15", 0)
	if class, ok := ClassifyUpstreamFailure(err); !ok || class != UpstreamFailureTransient {
		t.Fatalf("got error %v, want a transient read failure", err)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
