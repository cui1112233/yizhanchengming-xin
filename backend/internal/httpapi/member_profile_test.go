package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
)

func TestMemberProfileIncludesDetailsAndOwnerTeamShape(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "member-profile-details.json"), []byte(`[{"username":"alice","bio":"个人简介","email":"alice@example.com","teamTitle":"负责人"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	store := NewMemberProfileStore(dir)
	accounts := &authTestStore{account: auth.Account{Username: "alice", IsOwner: true, Active: true, CreatedAt: time.Unix(100, 0)}}
	member := store.member("alice", accounts)
	if member["bio"] != "个人简介" || member["email"] != "alice@example.com" || member["teamTitle"] != "负责人" {
		t.Fatalf("member details = %#v", member)
	}
	if member["apiEnabled"] != false {
		t.Fatalf("owner apiEnabled should remain account preference: %#v", member["apiEnabled"])
	}
	team := goTeam(member)
	if team["managerUsername"] != "alice" || team["name"] != "alice的团队" {
		t.Fatalf("team = %#v", team)
	}
	usage := emptyMemberUsage()
	if _, ok := usage["month"].(map[string]any); !ok {
		t.Fatalf("usage = %#v", usage)
	}
	if _, err := json.Marshal(map[string]any{"member": member, "team": team, "usage": usage}); err != nil {
		t.Fatal(err)
	}
}

func TestMemberProfileThemePreferenceDefaultsPersistsAndValidates(t *testing.T) {
	dir := t.TempDir()
	store := NewMemberProfileStore(dir)
	accounts := &authTestStore{account: auth.Account{Username: "alice", IsOwner: true, Active: true, CreatedAt: time.Unix(100, 0)}}

	if got := store.member("alice", accounts)["themeId"]; got != "system" {
		t.Fatalf("default themeId=%v, want system", got)
	}
	updated, err := store.update("alice", map[string]any{"themeId": "dark"}, accounts)
	if err != nil {
		t.Fatalf("update theme: %v", err)
	}
	if got := updated["themeId"]; got != "dark" {
		t.Fatalf("updated themeId=%v, want dark", got)
	}
	if got := store.member("alice", accounts)["themeId"]; got != "dark" {
		t.Fatalf("persisted themeId=%v, want dark", got)
	}
	if _, err := store.update("alice", map[string]any{"themeId": "ocean-dark"}, accounts); err != nil {
		t.Fatalf("scheme themeId ocean-dark should be accepted: %v", err)
	}
	if got := store.member("alice", accounts)["themeId"]; got != "ocean-dark" {
		t.Fatalf("scheme themeId=%v, want ocean-dark", got)
	}
	if _, err := store.update("alice", map[string]any{"themeId": "coral-dark"}, accounts); err != nil {
		t.Fatalf("scheme themeId coral-dark should be accepted: %v", err)
	}
	if _, err := store.update("alice", map[string]any{"themeId": "unknown-theme"}, accounts); err == nil {
		t.Fatal("unsupported themeId should be rejected")
	}
	if _, err := store.update("alice", map[string]any{"themeId": "mint"}, accounts); err == nil {
		t.Fatal("retired themeId mint should be rejected")
	}
	if got := store.member("alice", accounts)["themeId"]; got != "coral-dark" {
		t.Fatalf("rejected update changed themeId=%v, want coral-dark", got)
	}
}

func TestMemberProfileWindowBackgroundPreference(t *testing.T) {
	store := NewMemberProfileStore(t.TempDir())
	accounts := &authTestStore{account: auth.Account{Username: "alice", Active: true}}
	if got := store.member("alice", accounts)["windowBackgroundId"]; got != "none" {
		t.Fatalf("default windowBackgroundId=%v, want none", got)
	}
	updated, err := store.update("alice", map[string]any{"windowBackgroundId": "preset3"}, accounts)
	if err != nil {
		t.Fatal(err)
	}
	if got := updated["windowBackgroundId"]; got != "preset3" {
		t.Fatalf("windowBackgroundId=%v, want preset3", got)
	}
	if _, err := store.update("alice", map[string]any{"windowBackgroundId": "preset9"}, accounts); err != nil {
		t.Fatalf("preset9 should be accepted: %v", err)
	}
	if _, err := store.update("alice", map[string]any{"windowBackgroundId": "unknown"}, accounts); err == nil {
		t.Fatal("unsupported windowBackgroundId should be rejected")
	}
}

func TestCinematicGlassThemeAndVideoWallpaperPersistIndependently(t *testing.T) {
	store := NewMemberProfileStore(t.TempDir())
	accounts := &authTestStore{account: auth.Account{Username: "alice", Active: true}}
	for _, background := range []string{"cinema-auto", "cinema-0", "cinema-1", "cinema-2", "cinema-3"} {
		if _, err := store.update("alice", map[string]any{"themeId": "glass-dark", "windowBackgroundId": background}, accounts); err != nil {
			t.Fatal(err)
		}
		updated, err := store.update("alice", map[string]any{"themeId": "glass-light"}, accounts)
		if err != nil {
			t.Fatal(err)
		}
		if updated["windowBackgroundId"] != background {
			t.Fatal("changing theme overwrote wallpaper")
		}
	}
}

func TestMemberWindowBackgroundUploadAndServe(t *testing.T) {
	now := time.Date(2026, 11, 9, 15, 0, 0, 0, time.UTC)
	accounts := &authTestStore{account: auth.Account{Username: "alice", Active: true}}
	avatarDir := t.TempDir()
	mux := http.NewServeMux()
	RegisterMemberProfileRoutes(mux, BridgeAuth{Secret: "test-secret", Now: func() time.Time { return now }, AuthStore: accounts}, NewMemberProfileStore(t.TempDir()), accounts, avatarDir)
	imageBytes := []byte("\x89PNG\r\n\x1a\nbackground-bytes")
	body := `{"backgroundDataUrl":"data:image/png;base64,` + base64.StdEncoding.EncodeToString(imageBytes) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/member/window-background", strings.NewReader(body))
	SignBridgeRequest(req, "alice", true, now, "test-secret")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("upload status=%d body=%s", res.Code, res.Body.String())
	}
	var payload struct {
		Member map[string]any `json:"member"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	backgroundURL, _ := payload.Member["windowBackgroundUrl"].(string)
	if !strings.HasPrefix(backgroundURL, "/user-content/backgrounds/alice.png?v=") {
		t.Fatalf("windowBackgroundUrl=%q", backgroundURL)
	}
	readRes := httptest.NewRecorder()
	mux.ServeHTTP(readRes, httptest.NewRequest(http.MethodGet, backgroundURL, nil))
	if readRes.Code != http.StatusOK || readRes.Body.String() != string(imageBytes) {
		t.Fatalf("background fetch status=%d body=%q", readRes.Code, readRes.Body.String())
	}
}

func TestMemberSecuritySeparatesRememberedSessions(t *testing.T) {
	now := time.Date(2026, 10, 11, 3, 0, 0, 0, time.UTC)
	accounts := &authTestStore{
		account: auth.Account{Username: "alice", IsOwner: true, Active: true, CreatedAt: now},
		sessions: []auth.Session{
			{Token: "runtime-token-hash", Username: "alice", IssuedAt: now, ExpiresAt: now.Add(24 * time.Hour), IPHint: "127.0.0.1"},
			{Token: "remembered-token-hash", Username: "alice", IssuedAt: now, ExpiresAt: now.Add(30 * 24 * time.Hour), IPHint: "127.0.0.1"},
		},
	}
	mux := http.NewServeMux()
	RegisterMemberProfileRoutes(mux, BridgeAuth{Secret: "test-secret", Now: func() time.Time { return now }, AuthStore: accounts}, NewMemberProfileStore(t.TempDir()), accounts, t.TempDir())
	req := httptest.NewRequest(http.MethodGet, "/api/member/security", nil)
	SignBridgeRequest(req, "alice", true, now, "test-secret")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("security status=%d body=%s", res.Code, res.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(strings.NewReader(res.Body.String())).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if got := len(payload["sessions"].([]any)); got != 1 {
		t.Fatalf("runtime sessions=%d, want 1: %#v", got, payload["sessions"])
	}
	persistent, ok := payload["persistentSessions"].([]any)
	if !ok || len(persistent) != 1 {
		t.Fatalf("persistent sessions=%#v, want one remembered session", payload["persistentSessions"])
	}
	if _, ok := persistent[0].(map[string]any)["expiresAt"]; !ok {
		t.Fatalf("remembered session must expose expiry: %#v", persistent[0])
	}
}

func TestMemberAvatarURLVersionsLegacyRows(t *testing.T) {
	legacy := map[string]any{"avatarUrl": "/user-content/avatars/owner.png", "updatedAt": "2026-11-04T11:43:33Z"}
	got, _ := memberAvatarURL(legacy).(string)
	if !strings.HasPrefix(got, "/user-content/avatars/owner.png?v=") {
		t.Fatalf("legacy avatarUrl=%q, want a versioned URL", got)
	}
	versioned := map[string]any{"avatarUrl": "/user-content/avatars/owner.png?v=abc123", "updatedAt": "2026-11-04T11:43:33Z"}
	if memberAvatarURL(versioned) != "/user-content/avatars/owner.png?v=abc123" {
		t.Fatalf("versioned avatarUrl was rewritten: %v", memberAvatarURL(versioned))
	}
}

func TestAuthAccountPayloadIncludesOwnerAvatar(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "member-profiles.json"), []byte(`[{"username":"owner","avatarUrl":"/user-content/avatars/owner.png?v=abc123","displayName":"站长","themeId":"mint"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	payload := authAccountPayload(auth.Account{Username: "owner", IsOwner: true, Active: true}, NewMemberProfileStore(dir))
	if payload["role"] != "dev" {
		t.Fatalf("owner role=%v, want dev", payload["role"])
	}
	if payload["avatarUrl"] != "/user-content/avatars/owner.png?v=abc123" {
		t.Fatalf("owner avatarUrl=%v, want the profile avatar", payload["avatarUrl"])
	}
	if payload["displayName"] != "站长" {
		t.Fatalf("owner displayName=%v, want profile displayName", payload["displayName"])
	}
	if payload["themeId"] != "system" {
		t.Fatalf("retired stored themeId=%v, want graceful fallback to system", payload["themeId"])
	}
}

// TestMemberAvatarUploadIsServableWithVersionedURL covers the full upload ->
// store -> read-back loop. Upload used to report success while the recorded
// /user-content/avatars/… URL had no serving route, so the browser got 404 and
// Antd Avatar silently kept showing the initial-letter placeholder.
func TestMemberAvatarUploadIsServableWithVersionedURL(t *testing.T) {
	now := time.Date(2026, 11, 4, 3, 0, 0, 0, time.UTC)
	systemDir := t.TempDir()
	avatarDir := t.TempDir()
	accounts := &authTestStore{account: auth.Account{Username: "alice", IsOwner: true, Active: true, CreatedAt: now}}
	mux := http.NewServeMux()
	RegisterMemberProfileRoutes(mux, BridgeAuth{Secret: "test-secret", Now: func() time.Time { return now }, AuthStore: accounts}, NewMemberProfileStore(systemDir), accounts, avatarDir)

	imageBytes := []byte("\x89PNG\r\n\x1a\navatar-bytes")
	body := `{"avatarDataUrl":"data:image/png;base64,` + base64.StdEncoding.EncodeToString(imageBytes) + `"}`
	uploadReq := httptest.NewRequest(http.MethodPost, "/api/member/avatar", strings.NewReader(body))
	SignBridgeRequest(uploadReq, "alice", true, now, "test-secret")
	uploadRes := httptest.NewRecorder()
	mux.ServeHTTP(uploadRes, uploadReq)
	if uploadRes.Code != http.StatusOK {
		t.Fatalf("upload status=%d body=%s", uploadRes.Code, uploadRes.Body.String())
	}
	var payload struct {
		Member map[string]any `json:"member"`
	}
	if err := json.NewDecoder(strings.NewReader(uploadRes.Body.String())).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	avatarURL, _ := payload.Member["avatarUrl"].(string)
	if !strings.HasPrefix(avatarURL, "/user-content/avatars/alice.png?v=") {
		t.Fatalf("avatarUrl=%q, want a versioned /user-content/avatars URL", avatarURL)
	}
	if _, err := os.Stat(filepath.Join(avatarDir, "alice.png")); err != nil {
		t.Fatalf("avatar file was not written: %v", err)
	}

	readReq := httptest.NewRequest(http.MethodGet, avatarURL, nil)
	readRes := httptest.NewRecorder()
	mux.ServeHTTP(readRes, readReq)
	if readRes.Code != http.StatusOK {
		t.Fatalf("avatar fetch status=%d, want 200", readRes.Code)
	}
	if contentType := readRes.Header().Get("Content-Type"); contentType != "image/png" {
		t.Fatalf("avatar Content-Type=%q, want image/png", contentType)
	}
	if readRes.Body.String() != string(imageBytes) {
		t.Fatalf("avatar body=%q, want the uploaded bytes", readRes.Body.String())
	}

	for _, path := range []string{"/user-content/avatars/alice.svg", "/user-content/avatars/missing.png"} {
		missing := httptest.NewRecorder()
		mux.ServeHTTP(missing, httptest.NewRequest(http.MethodGet, path, nil))
		if missing.Code != http.StatusNotFound {
			t.Fatalf("%s status=%d, want 404", path, missing.Code)
		}
	}
}

var _ interface {
	ListSessionsForUser(context.Context, string) ([]auth.Session, error)
} = (*authTestStore)(nil)
