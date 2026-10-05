package generation

import (
	"strings"
	"testing"
)

func TestFinalPromptCompilerUsesStableBackendOrder(t *testing.T) {
	compiler := FinalPromptCompiler{}
	out := compiler.Compile(FinalPromptInput{
		SystemPreset: "SYS",
		Script: "SCRIPT",
		Hook: "HOOK",
		Director: "DIRECTOR",
		ProcessingRules: "RULES",
		KnowledgeBase: "KB",
		ProjectConfig: "PROJECT",
		UserConfig: "USER",
		ModelConfig: "MODEL",
	})

	want := []string{"SYS", "SCRIPT", "HOOK", "DIRECTOR", "RULES", "KB", "PROJECT", "USER", "MODEL"}
	last := -1
	for _, part := range want {
		idx := strings.Index(out, part)
		if idx < 0 { t.Fatalf("compiled prompt missing %q: %s", part, out) }
		if idx <= last { t.Fatalf("compiled prompt order unstable at %q: %s", part, out) }
		last = idx
	}
}

func TestFinalPromptCompilerIsDeterministic(t *testing.T) {
	in := FinalPromptInput{SystemPreset: "a", Script: "b", Director: "c"}
	compiler := FinalPromptCompiler{}
	if compiler.Compile(in) != compiler.Compile(in) { t.Fatal("same input must compile identically") }
}
