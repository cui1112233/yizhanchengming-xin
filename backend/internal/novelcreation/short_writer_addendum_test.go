package novelcreation

import (
	"context"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/prompt/novelcreation"
	"strings"
	"testing"
)

// Native evidence wraps writer clients only. These requests must not spread
// writer-only constraints into analysis, planning or editorial review prompts.
func TestShortImitationWriterAddendumBoundaries(t *testing.T) {
	defs := []struct {
		id     string
		writer bool
	}{
		{novelcreation.ShortBodyPrompt.ID, true},
		{novelcreation.ShortBodyRevision.ID, true},
		{novelcreation.ShortImitationDeconstruction.ID, false},
		{novelcreation.ShortImitationBlueprint.ID, false},
		{novelcreation.ShortReviewPrompt.ID, false},
		{novelcreation.ShortImitationEditorial.ID, false},
	}
	provider := &queuedPipelineProvider{responses: []string{`{"content":"ok"}`, `{"content":"ok"}`, `{"content":"ok"}`, `{"content":"ok"}`, `{"content":"ok"}`, `{"content":"ok"}`}}
	p, _, in, _ := controlledShortFixture(t, nil, NormalizeAgentSettings(AgentSettings{}, 2000))
	p.resolveProvider = func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }
	in.ShortImitationVersion = 2
	in.SystemPrompts = &SystemPromptSnapshot{Version: 1, Entries: map[string]FrozenSystemPrompt{}}
	for _, d := range novelcreation.AdminPrompts() {
		in.SystemPrompts.Entries[d.ID] = FrozenSystemPrompt{ID: d.ID, Version: 1, Revision: 1, Body: d.Definition.SystemPrompt}
	}
	const marker = "writer-only-frozen-proof"
	in.SystemPrompts.Entries[novelcreation.ShortImitationWritingID] = FrozenSystemPrompt{ID: novelcreation.ShortImitationWritingID, Version: 1, Revision: 1, Body: marker}
	for i, spec := range defs {
		var target shortText
		var found bool
		for _, d := range novelcreation.AdminPrompts() {
			if d.ID != spec.id {
				continue
			}
			found = true
			if _, err := p.callJSON(context.Background(), in, &recordingEventSink{}, nil, modelCallSpec{definition: d.Definition, key: ":boundary", step: RunWriting, userPayload: map[string]string{"test": "only"}, maxTokens: 1000}, &target); err != nil {
				t.Fatal(err)
			}
			break
		}
		if !found {
			t.Fatal("test definition missing", spec.id)
		}
		if strings.Contains(provider.requests[i].SystemPrompt, marker) != spec.writer {
			t.Fatal("wrong writer boundary", spec.id)
		}
	}
}
