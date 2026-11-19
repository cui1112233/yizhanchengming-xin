package screenplay

import (
	"context"
	"fmt"
	"qiantie/backend/internal/prompt"
	"strings"
)

func generateStoryboardPlanContinuation(ctx context.Context, generator TextGenerator, model, system, user, density, seed string, observers ...func(PlanningProgress)) (storyboardPlanRun, error) {
	out := storyboardPlanRun{Text: strings.TrimSpace(seed)}
	messages := []TextMessage{{Role: "system", Content: system}, {Role: "user", Content: user}}
	if out.Text != "" {
		messages = append(messages, TextMessage{Role: "assistant", Content: out.Text}, TextMessage{Role: "user", Content: prompt.ScreenplayNativePlanContinue(out.Text)})
	}
	if len(messages) == 4 {
		r := []rune(messages[2].Content)
		if len(r) > 32000 {
			messages[2].Content = string(r[len(r)-32000:])
		}
	}
	for _, observe := range observers {
		if observe != nil {
			observe(planningProgressFrame(1, out.Text, density, "requesting"))
		}
	}
	if err := ctx.Err(); err != nil {
		return out, err
	}
	response, err := generator.Generate(ctx, TextRequest{Model: model, SystemPrompt: system, UserPrompt: messages[len(messages)-1].Content, Messages: messages, MaxOutputTokens: 16000, Temperature: .4})
	out.Rounds = 1
	if ctx.Err() != nil {
		return out, ctx.Err()
	}
	if err != nil {
		return out, err
	}
	out.Text = appendStoryboardPlanChunk(out.Text, response.Content)
	out.Appended = strings.TrimSpace(response.Content) != ""
	if out.Text == "" {
		return out, fmt.Errorf("%w: 规划生成为空", ErrArtifactInvalid)
	}
	for _, observe := range observers {
		if observe != nil {
			observe(planningProgressFrame(1, out.Text, density, "received"))
		}
	}
	return out, nil
}
