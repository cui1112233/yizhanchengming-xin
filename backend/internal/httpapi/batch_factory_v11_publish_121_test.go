package httpapi

import (
	"encoding/json"
	"testing"
)

func TestPublishOrganizationUsesNestedBatchPublishSettings(t *testing.T) {
	settings := map[string]json.RawMessage{
		"publishSettings": json.RawMessage(`{"organization":1}`),
	}
	if got := publishOrganization(map[string]any{}, settings); got != "1" {
		t.Fatalf("publishOrganization() = %q, want 1", got)
	}
}
