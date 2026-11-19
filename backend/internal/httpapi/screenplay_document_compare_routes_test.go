package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/screenplay"
)

func TestScreenplayRoutesDocumentVersionCompare(t *testing.T) {
	now := time.Date(2026, 11, 19, 1, 30, 0, 0, time.UTC)
	store := screenplay.NewMemoryStore()
	ctx := context.Background()
	if _, err := store.CreateProject(ctx, screenplay.Project{ID: "compare-project", OwnerUsername: "alice", Title: "版本择优"}); err != nil {
		t.Fatal(err)
	}
	if _, err := screenplay.SaveManualDocument(ctx, store, "alice", "compare-project", "# 第1集\n\n第一场\n林川走进巷口。"); err != nil {
		t.Fatal(err)
	}
	historical, err := screenplay.SaveManualDocument(ctx, store, "alice", "compare-project", "# 第1集\n\n第一场\n林川走进巷口，雨下起来了。")
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterScreenplayRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, nil, func() time.Time { return now })
	base := "/api/comic-screenplays/projects/compare-project"

	// GET 单个版本返回完整正文（列表接口不返回正文）。
	fetched := signedJSONRequest(t, mux, now, "alice", http.MethodGet, base+"/documents/"+historical.ID, nil)
	if fetched.Code != http.StatusOK {
		t.Fatalf("get document=%d %s", fetched.Code, fetched.Body.String())
	}
	var payload struct {
		Document screenplay.Document `json:"document"`
	}
	if err := json.Unmarshal(fetched.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Document.Content == "" || !strings.Contains(payload.Document.Content, "雨下起来了") {
		t.Fatalf("document content missing: %q", payload.Document.Content)
	}

	// 版本状态梳理：后保存的 v2 是当前活动版本，v1 已被取代为历史版本。
	documents, err := store.ListDocuments(ctx, "alice", "compare-project")
	if err != nil {
		t.Fatal(err)
	}
	var activeID, supersededID string
	for _, doc := range documents {
		if doc.Status == "active" {
			activeID = doc.ID
		} else {
			supersededID = doc.ID
		}
	}
	if activeID == "" || supersededID == "" {
		t.Fatalf("expected one active and one superseded document, got %+v", documents)
	}

	// DELETE 当前活动版本被拒绝。
	rejected := signedJSONRequest(t, mux, now, "alice", http.MethodDelete, base+"/documents/"+activeID, nil)
	if rejected.Code == http.StatusOK || !containsText(rejected.Body.String(), "当前版本不可删除") {
		t.Fatalf("delete active=%d %s", rejected.Code, rejected.Body.String())
	}

	// DELETE 历史版本成功，列表只剩当前版本。
	removed := signedJSONRequest(t, mux, now, "alice", http.MethodDelete, base+"/documents/"+supersededID, nil)
	if removed.Code != http.StatusNoContent {
		t.Fatalf("delete historical=%d %s", removed.Code, removed.Body.String())
	}
	documents, err = store.ListDocuments(ctx, "alice", "compare-project")
	if err != nil {
		t.Fatal(err)
	}
	if len(documents) != 1 {
		t.Fatalf("expected 1 document left, got %d", len(documents))
	}
	if _, err := store.GetDocument(ctx, "alice", "compare-project", supersededID); err == nil {
		t.Fatalf("deleted document still fetchable")
	}

	// 越权访问返回 404。
	foreign := signedJSONRequest(t, mux, now, "bob", http.MethodGet, base+"/documents/"+activeID, nil)
	if foreign.Code != http.StatusNotFound {
		t.Fatalf("foreign get=%d %s", foreign.Code, foreign.Body.String())
	}
}
