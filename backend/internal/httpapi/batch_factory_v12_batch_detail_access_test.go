package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

// signedOwnerRequest 以系统 owner 身份签名，用来验证归属账号名只对该角色可见。
func signedOwnerRequest(t *testing.T, api http.Handler, now time.Time, username, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	SignBridgeRequest(req, username, true, now, "secret")
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)
	return rec
}

func createBatchForOwner(t *testing.T, api http.Handler, now time.Time, owner string) string {
	t.Helper()
	rec := signedJSONRequest(t, api, now, owner, http.MethodPost, "/api/batch-factory/v11/batches", map[string]any{"title": "11.5"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create batch status = %d body=%s", rec.Code, rec.Body.String())
	}
	created := decodeBody[struct {
		Batch struct {
			ID string `json:"id"`
		} `json:"batch"`
	}](t, rec)
	if created.Batch.ID == "" {
		t.Fatalf("create batch returned empty id: %s", rec.Body.String())
	}
	return created.Batch.ID
}

func TestBatchDetailCrossAccountReturnsForbiddenWithoutOwnerName(t *testing.T) {
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 3, Store: batchfactoryv11.NewMemoryStore()})
	batchID := createBatchForOwner(t, api, now, "weimeixia")

	for _, path := range []string{
		"/api/batch-factory/v12/batches/" + batchID,
		"/api/batch-factory/v11/batches/" + batchID,
	} {
		rec := signedJSONRequest(t, api, now, "choushiyiguai", http.MethodGet, path, nil)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("%s status = %d body=%s, want 403", path, rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), `"code":"batch_forbidden"`) {
			t.Fatalf("%s must carry batch_forbidden: %s", path, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "该批次属于其他账号") {
			t.Fatalf("%s must explain the ownership mismatch: %s", path, rec.Body.String())
		}
		// 普通账号不能拿到归属账号名，否则 403 会变成账号枚举入口。
		if strings.Contains(rec.Body.String(), "weimeixia") {
			t.Fatalf("%s leaked owner account to a non-owner: %s", path, rec.Body.String())
		}
	}
}

func TestBatchDetailOwnerIdentitySeesOwningAccount(t *testing.T) {
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 3, Store: batchfactoryv11.NewMemoryStore()})
	batchID := createBatchForOwner(t, api, now, "weimeixia")

	rec := signedOwnerRequest(t, api, now, "admin", http.MethodGet, "/api/batch-factory/v12/batches/"+batchID)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s, want 403", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"batch_forbidden"`) {
		t.Fatalf("must carry batch_forbidden: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "weimeixia") || !strings.Contains(rec.Body.String(), `"owner":"weimeixia"`) {
		t.Fatalf("system owner should see the owning account: %s", rec.Body.String())
	}
}

func TestBatchDetailMissingBatchReturnsStructuredNotFound(t *testing.T) {
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 3, Store: batchfactoryv11.NewMemoryStore()})

	rec := signedJSONRequest(t, api, now, "weimeixia", http.MethodGet, "/api/batch-factory/v12/batches/batch_does_not_exist", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s, want 404", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"batch_not_found"`) {
		t.Fatalf("missing batch must carry batch_not_found: %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"error":"not found"`) {
		t.Fatalf("missing batch must not return the bare not found body: %s", rec.Body.String())
	}
}

func TestBatchDetailOwnBatchStillReadable(t *testing.T) {
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 3, Store: batchfactoryv11.NewMemoryStore()})
	batchID := createBatchForOwner(t, api, now, "weimeixia")

	rec := signedJSONRequest(t, api, now, "weimeixia", http.MethodGet, "/api/batch-factory/v12/batches/"+batchID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s, want 200", rec.Code, rec.Body.String())
	}
}
