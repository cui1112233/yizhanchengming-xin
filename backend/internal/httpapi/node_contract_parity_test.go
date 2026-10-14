package httpapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type nodeRouteInventory struct {
	Routes []struct {
		Method string `json:"method"`
		Path   string `json:"path"`
		Owner  string `json:"owner"`
		Status string `json:"status"`
	} `json:"routes"`
}

// The inventory is the cutover guardrail: every remaining Express mount must
// either have a Go shadow implementation or be explicitly named as a blocker.
// This keeps cleanup honest and prevents a newly added Node mount from being
// silently treated as part of the Go-owned API.
func TestNodeRouteInventoryIsExplicit(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "migration", "node-api-contract-inventory.json"))
	if err != nil {
		t.Fatal(err)
	}
	var inventory nodeRouteInventory
	if err := json.Unmarshal(raw, &inventory); err != nil {
		t.Fatal(err)
	}
	if len(inventory.Routes) == 0 {
		t.Fatal("node route inventory is empty")
	}
	blocked := 0
	for _, route := range inventory.Routes {
		if route.Path == "" || route.Method == "" || route.Owner == "" || route.Status == "" {
			t.Fatalf("incomplete route record: %+v", route)
		}
		switch route.Status {
		case "shadow":
			if route.Owner != "go" {
				t.Fatalf("shadow route is not Go-owned: %+v", route)
			}
		case "blocked":
			blocked++
		default:
			t.Fatalf("unknown route status %q for %s", route.Status, route.Path)
		}
	}
	if blocked != 0 {
		t.Fatalf("expected no blocked legacy contracts, got %d", blocked)
	}
}
