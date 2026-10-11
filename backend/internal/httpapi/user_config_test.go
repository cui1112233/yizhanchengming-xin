package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
)

func TestUserConfigReportsModelManagementForOwner(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	store := &authTestStore{account: auth.Account{Username: "tester", Active: true, IsOwner: true}}
	mux := http.NewServeMux()
	registerUserConfigRoutes(mux, BridgeAuth{Secret: "test-secret", Now: func() time.Time { return now }, AuthStore: store}, NewUserConfigStore(t.TempDir()), store)
	handler := mux
	request := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	SignBridgeRequest(request, "tester", true, now, "test-secret")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("config status=%d body=%s", response.Code, response.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["canManageApi"] != true {
		t.Fatalf("owner config must expose model management: %#v", payload)
	}
	if _, ok := payload["managedBy"]; !ok {
		t.Fatalf("owner config must expose managedBy for frontend parity: %#v", payload)
	}
}
