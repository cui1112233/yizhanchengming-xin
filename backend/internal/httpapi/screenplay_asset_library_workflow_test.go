package httpapi

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/screenplay"
)

func TestScreenplayAssetLibraryWorkflowRegistersPersonalLibrary(t *testing.T) {
	store := screenplay.NewMemoryStore()
	_, err := store.CreateProject(context.Background(), screenplay.Project{ID: "library-workflow", OwnerUsername: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1700000000, 0)
	mux := http.NewServeMux()
	registerScreenplayWorkflow(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, RouterOptions{Screenplay: store}, nil, nil)
	response := signedJSONRequest(t, mux, now, "alice", "GET", "/api/comic-screenplays/asset-library?type=character", nil)
	if response.Code != 200 || !strings.Contains(response.Body.String(), `"entries":[]`) {
		t.Fatal(response.Code, response.Body.String())
	}
}
