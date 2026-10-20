package httpapi

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPublicManagedModelCapabilitiesExcludeCredentialsAndValidateImageSettings(t *testing.T) {
	model := map[string]any{
		"id": "image-primary", "kind": "image", "displayName": "主图片模型",
		"provider": "personal_api", "credential": "secret-key", "enabled": true,
	}
	public := publicManagedModel(model)
	encoded, err := json.Marshal(public)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "secret-key") || strings.Contains(string(encoded), "credential") {
		t.Fatalf("public model leaks credential: %s", encoded)
	}
	capabilities, ok := public["capabilities"].(PublicModelCapabilities)
	if !ok {
		t.Fatalf("capabilities type = %T", public["capabilities"])
	}
	if strings.Join(capabilities.Resolutions, ",") != "1K,2K,4K" || capabilities.MaxOutputCount != 4 {
		t.Fatalf("capabilities = %+v", capabilities)
	}
	if err := ValidateModelSettings(public, "image", map[string]any{"ratio": "16:9", "resolution": "8K", "count": 1}); err == nil || !strings.Contains(err.Error(), "resolution") {
		t.Fatalf("invalid resolution error = %v", err)
	}
}

func TestPublicManagedModelCapabilitiesAreKindSpecific(t *testing.T) {
	video := publicManagedModel(map[string]any{"id": "video-a", "kind": "video", "enabled": true})
	capabilities := video["capabilities"].(PublicModelCapabilities)
	if strings.Join(capabilities.ReferenceKinds, ",") != "image" || len(capabilities.QualityLevels) != 0 {
		t.Fatalf("video capabilities = %+v", capabilities)
	}
	if err := ValidateModelSettings(video, "image", map[string]any{}); err == nil || !strings.Contains(err.Error(), "kind") {
		t.Fatalf("kind mismatch error = %v", err)
	}
}
