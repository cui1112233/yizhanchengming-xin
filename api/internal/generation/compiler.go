package generation

import "strings"

type FinalPromptCompiler struct{}

func (FinalPromptCompiler) Compile(in FinalPromptInput) string {
	parts := []struct{ label, value string }{
		{"SYSTEM PRESET", in.SystemPreset},
		{"SCRIPT", in.Script},
		{"HOOK", in.Hook},
		{"DIRECTOR", in.Director},
		{"PROCESSING RULES", in.ProcessingRules},
		{"KNOWLEDGE BASE", in.KnowledgeBase},
		{"PROJECT CONFIG", in.ProjectConfig},
		{"USER CONFIG", in.UserConfig},
		{"MODEL CONFIG", in.ModelConfig},
	}
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part.value)
		if value == "" { continue }
		out = append(out, part.label+":\n"+value)
	}
	return strings.Join(out, "\n\n")
}
