package httpapi

import "testing"

func TestScriptVideoModelSpecMapsPublicModelKeys(t *testing.T) {
	cases := []struct {
		modelKey       string
		wantProvider   string
		acceptsImages  bool
		wantTaskPrefix string
	}{
		{"yd2-mini-video", "personal_api", true, ""},
		{"minimax-h3-video", "autodl_comfyui", true, "h3:"},
		{"seedance-2-0-official", "yfai_seedance", true, "yfai:"},
		{"local-doubao-executor-video", "doubao_local_executor", false, "go-local:"},
	}
	for _, tc := range cases {
		spec, ok := scriptVideoModelSpec(tc.modelKey)
		if !ok {
			t.Fatalf("model %q was not recognized", tc.modelKey)
		}
		if spec.Provider != tc.wantProvider || spec.AcceptsReferenceImages != tc.acceptsImages || spec.PublicTaskPrefix != tc.wantTaskPrefix {
			t.Fatalf("spec for %q = %#v", tc.modelKey, spec)
		}
	}
}

func TestScriptVideoModelSpecRejectsUnknownModel(t *testing.T) {
	if _, ok := scriptVideoModelSpec("unknown-video-model"); ok {
		t.Fatal("unknown model must not silently use a provider")
	}
}

func TestLocalScriptVideoInputRejectsNonLocalModel(t *testing.T) {
	if err := validateLocalScriptVideoInput(scriptVideoLocalInput{ModelKey: "minimax-h3-video", Prompt: "镜头"}); err == nil {
		t.Fatal("H3 must not be submitted to the local executor handler")
	}
	if err := validateLocalScriptVideoInput(scriptVideoLocalInput{ModelKey: "local-doubao-executor-video", Prompt: "镜头"}); err != nil {
		t.Fatalf("local model rejected: %v", err)
	}
}
