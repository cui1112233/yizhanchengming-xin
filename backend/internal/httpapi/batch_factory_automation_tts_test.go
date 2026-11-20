package httpapi

import "testing"

func TestAutomationTTSSemanticHashIsDeterministic(t *testing.T) {
	lines := []string{"第一句台词", "第二句台词"}
	base := automationTTSSemanticHash(lines, nil, false)
	if base != automationTTSSemanticHash([]string{"第一句台词", "第二句台词"}, nil, false) {
		t.Fatal("the same TTS inputs must produce the same semantic hash")
	}
	if base == automationTTSSemanticHash([]string{"第一句台词", "改过的第二句"}, nil, false) {
		t.Fatal("a changed line must produce a different semantic hash")
	}
	if base == automationTTSSemanticHash([]string{"第二句台词", "第一句台词"}, nil, false) {
		t.Fatal("line order must participate in the semantic hash")
	}
}

func TestAutomationTTSSemanticHashTracksModelSelection(t *testing.T) {
	lines := []string{"同一句台词"}
	model := map[string]any{"id": "tts-model-a", "endpoint": "https://tts.example/a"}
	withModel := automationTTSSemanticHash(lines, model, true)
	if withModel == automationTTSSemanticHash(lines, nil, false) {
		t.Fatal("switching the audio model must change the semantic hash")
	}
	if withModel != automationTTSSemanticHash(lines, map[string]any{"id": "tts-model-a", "endpoint": "https://tts.example/a"}, true) {
		t.Fatal("the same audio model must stay stable")
	}
	if withModel == automationTTSSemanticHash(lines, map[string]any{"id": "tts-model-b", "endpoint": "https://tts.example/b"}, true) {
		t.Fatal("a different audio model must change the semantic hash")
	}
}
