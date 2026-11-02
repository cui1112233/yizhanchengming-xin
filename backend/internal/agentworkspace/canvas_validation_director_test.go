package agentworkspace

import "testing"

func TestValidateCanvasForStorageAllowsDirectorNode(t *testing.T) {
	canvas := Canvas{Nodes: []Node{{
		ID:     "director-1",
		Type:   "director",
		Title:  "导演台预演",
		Status: "draft",
		Config: map[string]any{
			"directorDesk": map[string]any{
				"scenePrompt": "雨夜城市天桥",
				"camera":      map[string]any{"zoom": 1.0},
			},
		},
	}}}

	if err := ValidateCanvasForStorage(canvas); err != nil {
		t.Fatalf("director node must be valid for canvas storage: %v", err)
	}
}
