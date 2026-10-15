package videomanagementconnector

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
			_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
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
	if err != nil || readback.ID != "book-1" || readback.Status != "readback" {
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
