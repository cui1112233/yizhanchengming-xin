package videomanagementconnector

import (
	"context"
	"fmt"
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
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true,"id":"submission-1"}`))
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
}

func TestTargetHTTPServiceRejectsInvalidTargetURL(t *testing.T) {
	service := &TargetHTTPService{BaseURL: "://invalid"}
	_, err := service.Login(context.Background(), LoginInput{Username: "u", Password: "p"})
	if err == nil || !strings.Contains(fmt.Sprint(err), "target site") {
		t.Fatalf("err=%v", err)
	}
}
