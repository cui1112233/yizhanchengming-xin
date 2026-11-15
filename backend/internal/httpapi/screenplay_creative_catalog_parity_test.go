package httpapi

import (
	"encoding/json"
	"net/http"
	"qiantie/backend/internal/screenplay"
	"testing"
	"time"
)

func TestScreenplayCreativeHTTPOriginalStoryboardCatalog(t *testing.T) {
	now := time.Date(2026, 11, 15, 4, 0, 0, 0, time.UTC)
	mux := http.NewServeMux()
	RegisterScreenplayCreativeRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, screenplay.NewMemoryStore(), nil)
	response := signedJSONRequest(t, mux, now, "alice", "GET", "/api/screenplays/creative-library", nil)
	if response.Code != 200 {
		t.Fatal(response.Code, response.Body.String())
	}
	var result struct {
		Prompts []struct {
			ID        string `json:"id"`
			Purpose   string `json:"purpose"`
			Group     string `json:"catalog_group"`
			Created   string `json:"created_at"`
			Updated   string `json:"updated_at"`
			Supported bool   `json:"supported"`
			Digest    string `json:"sha256"`
		} `json:"prompts"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	original, compat, unavailable := 0, 0, 0
	for _, row := range result.Prompts {
		if row.Purpose != "storyboard" {
			continue
		}
		switch row.Group {
		case "original":
			original++
			if row.Created == "" || row.Updated == "" || len(row.Digest) != 64 {
				t.Fatal("missing real metadata", row.ID)
			}
			if !row.Supported {
				unavailable++
			}
		case "compatibility":
			compat++
		default:
			t.Fatal("unclassified prompt", row.ID)
		}
	}
	if original != 41 || compat != 2 || unavailable != 2 {
		t.Fatalf("original=%d compatibility=%d unavailable=%d", original, compat, unavailable)
	}
}
