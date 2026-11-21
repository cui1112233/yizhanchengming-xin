package httpapi

import "testing"

func TestTargetProfileOrganizationAcceptsNumericID(t *testing.T) {
	data := map[string]any{"organization": float64(1)}
	if got := targetProfileOrganization(data, nil); got != "1" {
		t.Fatalf("numeric organization id should be preserved, got %q", got)
	}
}

func TestTargetProfileOrganizationKeepsStringID(t *testing.T) {
	data := map[string]any{"organization_id": "9"}
	if got := targetProfileOrganization(data, nil); got != "9" {
		t.Fatalf("string organization id should be preserved, got %q", got)
	}
}
