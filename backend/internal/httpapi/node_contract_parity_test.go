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

// The inventory is the cutover guardrail: every historical Express mount is
// now fully owned by Go. The file remains as an audit record, not as a runtime
// routing table, and a shadow/blocked entry is never acceptable.
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
	for _, route := range inventory.Routes {
		if route.Path == "" || route.Method == "" || route.Owner == "" || route.Status == "" {
			t.Fatalf("incomplete route record: %+v", route)
		}
		switch route.Status {
		case "migrated":
			if route.Owner != "go" {
				t.Fatalf("migrated route is not Go-owned: %+v", route)
			}
		case "shadow", "blocked":
			t.Fatalf("unmigrated public route remains: %+v", route)
		default:
			t.Fatalf("unknown route status %q for %s", route.Status, route.Path)
		}
	}
}
