package videomanagementconnector

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContractDetectionSeparatesSchemaAuthAndTemporaryFailure(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{"html", 200, "<html>new application</html>", ErrSiteContractChanged},
		{"missing-endpoint", 404, "missing", ErrSiteContractChanged},
		{"missing-success", 200, `{"ok":true}`, ErrSiteContractChanged},
		{"login", 200, `<input name="username"><input name="password">`, ErrSessionExpired},
		{"temporary", 503, "unavailable", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()
			service := &TargetHTTPService{BaseURL: srv.URL, Client: srv.Client()}
			_, err := service.Submit(context.Background(), Session{Cookie: "fixture"}, SubmissionInput{BookID: "123"})
			if err == nil || (tc.want != nil && !errors.Is(err, tc.want)) || (tc.want == nil && errors.Is(err, ErrSiteContractChanged)) {
				t.Fatalf("misclassified: %v", err)
			}
		})
	}
}

func TestContractProbeNeverPostsAndRejectsMalformedReadback(t *testing.T) {
	valid := false
	writes := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writes++
			t.Error("probe wrote to target")
		}
		if r.URL.Path == "/tttadmin/zidingyi.php" {
			_, _ = w.Write([]byte("signed in"))
			return
		}
		if valid {
			_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
		} else {
			_, _ = w.Write([]byte(`{"changed":true}`))
		}
	}))
	defer srv.Close()
	service := &TargetHTTPService{BaseURL: srv.URL, Client: srv.Client()}
	if err := service.ProbeSubmissionContract(context.Background(), Session{Cookie: "fixture"}); !errors.Is(err, ErrSiteContractChanged) {
		t.Fatal(err)
	}
	valid = true
	if err := service.ProbeSubmissionContract(context.Background(), Session{Cookie: "fixture"}); err != nil {
		t.Fatal(err)
	}
	if writes != 0 {
		t.Fatal("probe submitted content")
	}
}
