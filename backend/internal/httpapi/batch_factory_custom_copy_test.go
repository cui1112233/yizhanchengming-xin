package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/videomanagementconnector"
	"strings"
	"testing"
	"time"
)

func TestCustomCopyAudioUsesAllLines(t *testing.T) {
	book := batchfactoryv11.Book{SourceText: "一\n二\n三\n四\n五\n六\n结尾", SourceMetadata: map[string]any{"sourceMode": "custom_copy", "contentRangeLines": 5}}
	if got := productionLines(book); len(got) != 7 || got[6] != "结尾" {
		t.Fatalf("audio lost lines: %v", got)
	}
	book.SettingsState.Patch = batchfactoryv11.SettingsPatch{"audioPlanningEnabled": json.RawMessage(`true`), "audioLineDurations": json.RawMessage(`[1,1,1,1,1]`), "audioDurationSeconds": json.RawMessage(`5`)}
	if hasReusableAudioPlan(book) {
		t.Fatal("truncated audio plan reused")
	}
	if got := string(publish121Content(book, "新一\n新二\n新三\n新四\n新五\n新六\n新结尾")); !strings.HasSuffix(got, "新结尾") || strings.Contains(got, "\n六\n") {
		t.Fatalf("publish mixed versions: %q", got)
	}
}

type copyReaderFixture struct {
	failingVideoManagementLogin
	cookie string
}

func (f *copyReaderFixture) CustomCopies(_ context.Context, s videomanagementconnector.Session, id string) ([]videomanagementconnector.CustomCopy, error) {
	f.cookie = s.Cookie
	if s.Cookie == "" {
		return nil, videomanagementconnector.ErrSessionMissing
	}
	return []videomanagementconnector.CustomCopy{{ID: "280894", BookID: id, Platform: "15", Content: "第一份"}, {ID: "280893", BookID: id, Platform: "15", Content: "第二份"}}, nil
}
func TestCustomCopyRouteUsesOnlyCurrentUserSession(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "alice"), 0700)
	raw, _ := json.Marshal(map[string]string{"cookie": "alice-secret"})
	os.WriteFile(filepath.Join(dir, "alice", "upload-target.json"), raw, 0600)
	target := &copyReaderFixture{}
	store := &authTestStore{account: auth.Account{Username: "alice", Active: true}, session: auth.Session{Token: "fixture", Username: "alice", ExpiresAt: time.Now().Add(time.Hour)}}
	mux := http.NewServeMux()
	RegisterNovelFetchUploadSessionCompat(mux, BridgeAuth{AuthStore: store}, dir, target)
	for _, token := range []string{"", "fixture"} {
		req := httptest.NewRequest("GET", "/api/batch-factory/v12/custom-copies/2100242549395006289?owner=bob", nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		res := httptest.NewRecorder()
		mux.ServeHTTP(res, req)
		if token == "" {
			if res.Code != 401 {
				t.Fatalf("anonymous %d", res.Code)
			}
			continue
		}
		if res.Code != 200 || !strings.Contains(res.Body.String(), "280893") || strings.Contains(res.Body.String(), "alice-secret") || target.cookie != "alice-secret" {
			t.Fatalf("%d %s", res.Code, res.Body.String())
		}
	}
}
