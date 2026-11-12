package aicut

import (
	"encoding/json"
	"testing"
)

func TestClipEditorRepairPreservesCatalogParametersAndZeroOpacity(t *testing.T) {
	original := Clip{CatalogResourceID: "filter-bright", CatalogParameters: map[string]float64{"intensity": 0}}
	data, err := encodeClipEditorJSON(original)
	if err != nil {
		t.Fatal(err)
	}
	var restored Clip
	if err := decodeClipEditorJSON(data, &restored); err != nil {
		t.Fatal(err)
	}
	if value, ok := restored.CatalogParameters["intensity"]; !ok || value != 0 {
		t.Fatalf("intensity lost: %s", data)
	}
	zero := 0.0
	settings := TrackSettings{TextOpacity: &zero}
	data, err = json.Marshal(settings)
	if err != nil {
		t.Fatal(err)
	}
	var decoded TrackSettings
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.TextOpacity == nil || *decoded.TextOpacity != 0 {
		t.Fatalf("opacity lost: %s", data)
	}
}
