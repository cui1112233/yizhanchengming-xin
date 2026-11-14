package httpapi

import (
	"context"
	"errors"
	"net/http"
	"qiantie/backend/internal/screenplay"
	"testing"
	"time"
)

type screenplayAssetReadFailure struct{ screenplay.Store }

func (s screenplayAssetReadFailure) ListAssets(context.Context, string, string, screenplay.AssetType) ([]screenplay.Asset, error) {
	return nil, errors.New("asset store unavailable")
}
func TestScreenplayWorkspaceDoesNotDisguiseAssetFailureAsEmptyLibrary(t *testing.T) {
	now := time.Now()
	memory := screenplay.NewMemoryStore()
	_, err := memory.CreateProject(context.Background(), screenplay.Project{ID: "read-failure", OwnerUsername: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterScreenplayRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, screenplayAssetReadFailure{memory}, nil, nil)
	r := signedJSONRequest(t, mux, now, "alice", "GET", "/api/comic-screenplays/projects/read-failure/workspace", nil)
	if r.Code != http.StatusInternalServerError || containsText(r.Body.String(), `"assets":[]`) {
		t.Fatalf("read failure reported as a successful empty workspace: %d %s", r.Code, r.Body.String())
	}
}
