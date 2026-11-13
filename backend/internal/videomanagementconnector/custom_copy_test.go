package videomanagementconnector

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCustomCopiesKeepsVersionsAndLargeBookID(t *testing.T) {
	const book = "2100242549395006289"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Cookie") != "session=test" {
			t.Error("missing session")
		}
		if r.URL.Path == "/tttadmin/api/zbooklist_get.php" {
			if r.URL.Query().Get("bookid") != book {
				t.Error("book id changed")
			}
			if r.URL.Query().Get("page") == "1" {
				fmt.Fprint(w, `{"success":true,"total":2,"data":[{"id":280894,"bookid":2100242549395006289,"bookname":"青霄赴","book_platform":15}]}`)
			} else {
				fmt.Fprint(w, `{"success":true,"total":2,"data":[{"id":280893,"bookid":"2100242549395006289","bookname":"青霄赴","book_platform":"15"}]}`)
			}
			return
		}
		fmt.Fprintf(w, `{"success":true,"content":"版本%s\n第二行"}`, r.URL.Query().Get("id"))
	}))
	defer server.Close()
	got, err := (&TargetHTTPService{BaseURL: server.URL, Client: server.Client()}).CustomCopies(context.Background(), Session{Cookie: "session=test"}, book)
	if err != nil || len(got) != 2 {
		t.Fatalf("%+v %v", got, err)
	}
	if got[0].BookID != book || got[0].Content == got[1].Content || got[1].Content != "版本280893\n第二行" {
		t.Fatalf("lost version: %+v", got)
	}
}

func TestCustomCopiesRejectsExpiredAndWrongBook(t *testing.T) {
	for _, body := range []string{`<html><form action="login.php">管理员登录</form></html>`, `{"success":true,"total":1,"data":[{"id":1,"bookid":"999999"}]}`} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, body) }))
		got, err := (&TargetHTTPService{BaseURL: server.URL, Client: server.Client()}).CustomCopies(context.Background(), Session{Cookie: "a=b"}, "2100242549395006289")
		server.Close()
		if body[0] == '<' {
			if !errors.Is(err, ErrSessionExpired) {
				t.Fatalf("%v", err)
			}
		} else if err != nil || len(got) != 0 {
			t.Fatalf("wrong book returned %+v %v", got, err)
		}
	}
}
