package httpapi

import (
	"encoding/json"
	"net/http"
	"qiantie/backend/internal/screenplay"
	"testing"
	"time"
)

func TestScreenplayPromptCatalogExposesProtocolAndDisabledReasons(t *testing.T) {
	mux := http.NewServeMux()
	now := time.Now()
	RegisterScreenplayCreativeRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, screenplay.NewMemoryStore(), nil)
	response := signedJSONRequest(t, mux, now, "alice", "GET", "/api/screenplays/creative-library", nil)
	var body struct {
		Prompts []struct {
			SourceID  int    `json:"source_id"`
			Supported bool   `json:"supported"`
			Reason    string `json:"unsupported_reason"`
			Protocol  string `json:"output_protocol"`
			Parser    string `json:"parser_version"`
			Body      string `json:"body"`
		}
	}
	if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &body) != nil {
		t.Fatal(response.Code, response.Body.String())
	}
	found := map[int]bool{}
	for _, p := range body.Prompts {
		if p.Body != "" {
			t.Fatal("public metadata leaked system body")
		}
		if p.SourceID == 138 || p.SourceID == 293 {
			if p.Supported || p.Reason == "" {
				t.Fatal("unsupported choice became enabled", p)
			}
			found[p.SourceID] = true
		}
		if p.SourceID == 139 {
			if !p.Supported || p.Reason != "" || p.Parser != "cm_storyboard_139_optional_image_v2" {
				t.Fatal("reviewed optional-image adapter missing", p)
			}
			found[139] = true
		}
		if p.SourceID == 137 {
			if !p.Supported || p.Protocol == "" || p.Parser == "" {
				t.Fatal("contract lost at HTTP boundary", p)
			}
			found[137] = true
		}
	}
	if len(found) != 4 {
		t.Fatal("catalog incomplete", found)
	}
}
