package prompt_test

import (
	"testing"

	promptcatalog "qiantie/backend/internal/prompt"
	promptagent "qiantie/backend/internal/prompt/agent"
	promptasset "qiantie/backend/internal/prompt/asset"
	promptdirector "qiantie/backend/internal/prompt/director"
	promptstoryboard "qiantie/backend/internal/prompt/storyboard"
)

func TestFirstAgentPromptDefinitionsHaveStableContracts(t *testing.T) {
	definitions := []promptcatalog.Definition{
		promptagent.WorkspaceChat,
		promptagent.DirectorPlan,
		promptagent.TextNode,
		promptasset.ContextReference,
		promptasset.VisualReference,
		promptstoryboard.ScriptSplit,
		promptstoryboard.ShotMerge,
		promptdirector.SceneSync,
	}
	seen := map[string]struct{}{}
	for _, definition := range definitions {
		if definition.ID == "" || definition.Version < 1 || definition.SystemPrompt == "" || len(definition.RequiredInputs) == 0 || definition.OutputContract == "" || len(definition.Errors) == 0 {
			t.Fatalf("incomplete definition: %+v", definition)
		}
		if _, exists := seen[definition.ID]; exists {
			t.Fatalf("duplicate prompt id %q", definition.ID)
		}
		seen[definition.ID] = struct{}{}
	}
}

func TestPromptDefinitionsKeepExpectedIDsAndVersions(t *testing.T) {
	expected := map[string]int{
		"agent.workspace-chat":    1,
		"agent.director-plan":     1,
		"agent.text-node":         1,
		"asset.context-reference": 1,
		"asset.visual-reference":  1,
		"storyboard.script-split": 1,
		"storyboard.shot-merge":   1,
		"director.scene-sync":     1,
	}
	for _, definition := range []promptcatalog.Definition{
		promptagent.WorkspaceChat, promptagent.DirectorPlan, promptagent.TextNode,
		promptasset.ContextReference, promptasset.VisualReference,
		promptstoryboard.ScriptSplit, promptstoryboard.ShotMerge, promptdirector.SceneSync,
	} {
		id, version := definition.ID, definition.Version
		if expected[id] != version {
			t.Fatalf("identity=%s@%d expected=%d", id, version, expected[id])
		}
		delete(expected, id)
	}
	if len(expected) != 0 {
		t.Fatalf("missing definitions: %v", expected)
	}
}
