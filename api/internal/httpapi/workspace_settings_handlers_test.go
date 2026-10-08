package httpapi

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

func TestWorkspaceSettingsPersistsOnlySafeUserPreferences(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	auth := &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{"batch.read"}}}
	h := NewHandler(Dependencies{Auth: auth, Database: db})
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO user_workspace_preferences(user_id,theme,notifications_enabled,storage_preference,pet_id,sound_volume,pet_visible,companion_active) VALUES(?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE theme=VALUES(theme),notifications_enabled=VALUES(notifications_enabled),storage_preference=VALUES(storage_preference),pet_id=VALUES(pet_id),sound_volume=VALUES(sound_volume),pet_visible=VALUES(pet_visible),companion_active=VALUES(companion_active)")).WithArgs(driver.Value(int64(7)), "dark", true, "tos", "fox", 35, false, true).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT theme,notifications_enabled,storage_preference,pet_id,sound_volume,pet_visible,companion_active FROM user_workspace_preferences WHERE user_id=?")).WithArgs(driver.Value(int64(7))).
		WillReturnRows(sqlmock.NewRows([]string{"theme", "notifications_enabled", "storage_preference", "pet_id", "sound_volume", "pet_visible", "companion_active"}).AddRow("dark", true, "tos", "fox", 35, false, true))

	req := httptest.NewRequest(http.MethodPut, "/api/v1/workspace/settings", bytes.NewBufferString(`{"theme":"dark","notificationsEnabled":true,"storagePreference":"tos","petId":"fox","soundVolume":35,"petVisible":false,"companionActive":true,"serverPath":"/root"}`))
	sameOrigin(req)
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	for _, forbidden := range []string{"serverPath", "token", "secret", "cookie"} {
		if strings.Contains(strings.ToLower(rec.Body.String()), strings.ToLower(forbidden)) {
			t.Fatalf("response leaked %q: %s", forbidden, rec.Body.String())
		}
	}
	var body struct {
		Settings map[string]any `json:"settings"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Settings["petId"] != "fox" || body.Settings["soundVolume"] != float64(35) {
		t.Fatalf("settings=%#v", body.Settings)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWorkspaceSettingsRejectsCrossOriginBeforeWriting(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	auth := &fakeAuthService{user: authn.User{ID: 7}}
	h := NewHandler(Dependencies{Auth: auth, Database: db})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/workspace/settings", bytes.NewBufferString(`{"theme":"dark","storagePreference":"tos"}`))
	req.Host = "app.example"
	req.Header.Set("Origin", "https://evil.example")
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
