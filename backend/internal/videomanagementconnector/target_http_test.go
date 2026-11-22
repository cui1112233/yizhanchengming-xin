package videomanagementconnector

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestTextSubmissionHTTPAuthFailureIsExplicitSessionExpiry(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(status) }))
			defer server.Close()
			service := &TargetHTTPService{BaseURL: server.URL, Client: server.Client()}
			_, err := service.Submit(context.Background(), Session{Cookie: "sid=fixture"}, SubmissionInput{BookID: "123", Content: []byte("测试正文")})
			if !errors.Is(err, ErrSessionExpired) {
				t.Fatalf("auth failure classified as unknown: %v", err)
			}
		})
	}
}

func TestTargetHTTPServiceUsesConfiguredProxy(t *testing.T) {
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "http://two.121w.com/tttadmin/login.php" {
			t.Fatalf("proxy endpoint = %s", r.URL.String())
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer proxy.Close()
	previous := os.Getenv("QIANTIE_VIDEO_CONNECTOR_PROXY")
	t.Setenv("QIANTIE_VIDEO_CONNECTOR_PROXY", proxy.URL)
	defer os.Setenv("QIANTIE_VIDEO_CONNECTOR_PROXY", previous)
	response, err := (&TargetHTTPService{BaseURL: "http://two.121w.com"}).client().Get("http://two.121w.com/tttadmin/login.php")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
}

func TestTargetHTTPServiceLoginAndSubmitUseOneCookieContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/tttadmin/login.php":
			w.Header().Add("Set-Cookie", "sid=first; Path=/")
			_, _ = w.Write([]byte("{}"))
		case "/tttadmin/api/login.php":
			if !strings.Contains(r.Header.Get("Cookie"), "sid=first") {
				t.Errorf("missing landing cookie")
			}
			w.Header().Add("Set-Cookie", "sid=second; Path=/")
			_, _ = w.Write([]byte(`{"success":true}`))
		case "/tttadmin/api/zbooklist_upload.php":
			if !strings.Contains(r.Header.Get("Cookie"), "sid=first") || !strings.Contains(r.Header.Get("Cookie"), "sid=second") {
				t.Errorf("missing login cookie: %q", r.Header.Get("Cookie"))
			}
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				t.Fatal(err)
			}
			if r.FormValue("bookid") != "book-1" {
				t.Errorf("bookid=%q", r.FormValue("bookid"))
			}
			if _, _, err := r.FormFile("files[]"); err != nil {
				t.Errorf("files[] upload missing: %v", err)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true,"id":"submission-1"}`))
		case "/tttadmin/api/zbooklist_get.php":
			if (r.URL.Query().Get("bookid") != "book-1" && r.URL.Query().Get("bookid") != "1") || r.URL.Query().Get("page") != "1" || r.URL.Query().Get("pageSize") != "10" {
				t.Errorf("unexpected readback query: %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"success":true,"data":[{"bookid":"book-1","id":"remote-record-1"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	service := &TargetHTTPService{BaseURL: server.URL, Client: server.Client()}
	session, err := service.Login(context.Background(), LoginInput{Username: "u", Password: "p"})
	if err != nil {
		t.Fatal(err)
	}
	if session.Cookie != "sid=first; sid=second" {
		t.Fatalf("cookie=%q", session.Cookie)
	}
	result, err := service.Submit(context.Background(), session, SubmissionInput{BookID: "book-1", Filename: "book-1.txt", Content: []byte("正文"), FormFields: map[string]string{"bookid": "book-1"}})
	if err != nil || result.ID != "submission-1" || result.Status != "submitted" {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	readback, err := service.GetSubmission(context.Background(), session, "book-1")
	if err != nil || readback.ID != "remote-record-1" || readback.Status != "readback" {
		t.Fatalf("readback=%#v err=%v", readback, err)
	}
	book, err := service.GetBook(context.Background(), session, "1", 1, 10)
	if err != nil {
		t.Fatalf("book readback: %v", err)
	}
	if value, ok := book.(map[string]any); !ok || value["success"] != true {
		t.Fatalf("book=%#v", book)
	}
}

func TestTargetHTTPBookReadbackPreservesLongIdentifiers(t *testing.T) {
	const bookID, receiptID = "2100242549395006289", "9100242549395006293"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Query().Get("bookid") != bookID {
			t.Errorf("unexpected book readback: %s %s", r.Method, r.URL)
		}
		_, _ = fmt.Fprintf(w, `{"success":true,"total":1,"data":[{"bookid":%s,"id":%s}]}`, bookID, receiptID)
	}))
	defer server.Close()
	ctx := context.Background()
	service := &TargetHTTPService{BaseURL: server.URL, Client: server.Client()}
	session := Session{Cookie: "sid=fixture"}
	raw, err := service.GetBook(ctx, session, bookID, 1, 100)
	value, ok := raw.(map[string]any)
	if err != nil || !ok {
		t.Fatalf("book listing = %#v, err = %v", raw, err)
	}
	rows := arrayOfMaps(value["data"])
	if len(rows) != 1 || fmt.Sprint(rows[0]["bookid"]) != bookID || fmt.Sprint(rows[0]["id"]) != receiptID {
		t.Errorf("listing changed book/receipt IDs: %+v", rows)
	}
	receipt, err := service.GetSubmission(ctx, session, bookID)
	if err != nil || receipt.ID != receiptID {
		t.Errorf("readback changed receipt ID: %+v %v", receipt, err)
	}
}

func TestTargetHTTPServiceCloneSubmissionUsesRemoteCloneContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tttadmin/api/zbooklist_clone.php" {
			http.NotFound(w, r)
			return
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		if got := r.FormValue("id"); got != "296749" {
			t.Errorf("source id = %q, want 296749", got)
		}
		if got := r.FormValue("action"); got != "clone" {
			t.Errorf("action = %q, want clone", got)
		}
		if got := r.FormValue("text_content"); got != "再次上传正文" {
			t.Errorf("text_content = %q", got)
		}
		config := r.FormValue("config")
		for _, want := range []string{`"jieya_num":8`, `"gunping_num":0`, `"ai_head":"3"`, `"organization":"1"`} {
			if !strings.Contains(config, want) {
				t.Errorf("config missing %s: %s", want, config)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"id":9100242549395006293}`))
	}))
	defer server.Close()

	service := &TargetHTTPService{BaseURL: server.URL, Client: server.Client()}
	result, err := service.CloneSubmission(context.Background(), Session{Cookie: "sid=ok"}, "296749", SubmissionInput{
		BookID:  "738579",
		Content: []byte("再次上传正文"),
		FormFields: map[string]string{
			"organization":  "1",
			"category":      "NEW_BOOK",
			"gender":        "2",
			"style":         "304",
			"jieya_num":     "8",
			"jieya_ai_head": "3",
			"gunping_num":   "0",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != "9100242549395006293" || result.Status != "submitted" {
		t.Fatalf("result = %#v", result)
	}
}

// 目标站会话失效时 zidingyi.php 会 302 到 login.php，客户端跟随跳转后最终
// 仍是 HTTP 200。CheckSession 必须识别正文里的登录页，否则会把过期会话
// 判成“已登录并验证”，与同步风格目录的结果自相矛盾。
func TestTargetHTTPServiceCheckSessionDetectsLoginPageAfterRedirect(t *testing.T) {
	loginPage := `<!DOCTYPE html><html><head><title>管理员登录</title></head><body>
<form id="loginForm"><input type="text" name="username"><input type="password" name="password"></form></body></html>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/tttadmin/zidingyi.php":
			http.Redirect(w, r, "/tttadmin/login.php", http.StatusFound)
		case "/tttadmin/login.php":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(loginPage))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	service := &TargetHTTPService{BaseURL: server.URL, Client: server.Client()}
	err := service.CheckSession(context.Background(), Session{Cookie: "PHPSESSID=stale"})
	if !errors.Is(err, ErrSessionExpired) {
		t.Fatalf("expired session err=%v, want ErrSessionExpired", err)
	}
	if err := service.CheckSession(context.Background(), Session{}); !errors.Is(err, ErrSessionMissing) {
		t.Fatalf("missing cookie err=%v, want ErrSessionMissing", err)
	}
}

func TestTargetHTTPServiceCheckSessionAcceptsRealStylePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tttadmin/zidingyi.php" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`<select id="style"><option value="1">古风虐文</option></select>`))
	}))
	defer server.Close()
	service := &TargetHTTPService{BaseURL: server.URL, Client: server.Client()}
	if err := service.CheckSession(context.Background(), Session{Cookie: "PHPSESSID=ok"}); err != nil {
		t.Fatalf("valid session err=%v", err)
	}
}

func TestTargetHTTPServiceUploadVideoClassifiesExpiredSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tttadmin/api/music_put_url.php" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":false,"message":"未登录"}`))
	}))
	defer server.Close()

	service := &TargetHTTPService{BaseURL: server.URL, Client: server.Client()}
	_, err := service.UploadVideo(context.Background(), Session{Cookie: "PHPSESSID=stale"}, "head.mp4", []byte("video"))
	if !errors.Is(err, ErrSessionExpired) {
		t.Fatalf("UploadVideo() error = %v, want ErrSessionExpired", err)
	}
	if !strings.Contains(err.Error(), "视频后台登录已失效") {
		t.Fatalf("UploadVideo() error = %q, want a clear Chinese login-expired message", err)
	}
}

func TestTargetHTTPServiceRejectsInvalidTargetURL(t *testing.T) {
	service := &TargetHTTPService{BaseURL: "://invalid"}
	_, err := service.Login(context.Background(), LoginInput{Username: "u", Password: "p"})
	if err == nil || !strings.Contains(fmt.Sprint(err), "target site") {
		t.Fatalf("err=%v", err)
	}
}

func TestTargetHTTPServiceOwnsMediaAndConfigurationReads(t *testing.T) {
	var uploaded []byte
	serverURLPlaceholder := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/tttadmin/api/music_put_url.php":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(fmt.Sprintf(`{"success":true,"data":{"files":[{"upload_url":%q,"object_key":"obj-1","bucket_url":"https://bucket.example/obj-1","headers":{"X-Test":"ok"}}]}}`, serverURLPlaceholder)))
		case "/object":
			if r.Method != http.MethodPut || r.Header.Get("X-Test") != "ok" {
				t.Errorf("unexpected object upload: %s %q", r.Method, r.Header.Get("X-Test"))
			}
			var err error
			uploaded, err = io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
		case "/tttadmin/api/zdy_config.php":
			_, _ = w.Write([]byte(`{"data":[{"id":"1","config_name":"测试档"}]}`))
		case "/tttadmin/zidingyi.php":
			_, _ = w.Write([]byte(`<option value="101">古风虐文</option>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	// The handler needs the server URL before serving the presign response.
	serverURLPlaceholder = server.URL + "/object"
	service := &TargetHTTPService{BaseURL: server.URL, Client: server.Client()}
	session := Session{Cookie: "sid=ok"}
	if _, err := service.UploadVideo(context.Background(), session, "head.mp4", []byte("video")); err != nil {
		t.Fatal(err)
	}
	if string(uploaded) != "video" {
		t.Fatalf("uploaded=%q", uploaded)
	}
	profiles, err := service.RefreshProfiles(context.Background(), session)
	if err != nil || len(arrayOfMaps(mapOrEmpty(profiles)["data"])) != 1 {
		t.Fatalf("profiles=%#v err=%v", profiles, err)
	}
	styles, err := service.RefreshStyles(context.Background(), session)
	if err != nil || !strings.Contains(styles, "古风虐文") {
		t.Fatalf("styles=%q err=%v", styles, err)
	}
}
