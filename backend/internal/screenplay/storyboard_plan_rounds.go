package screenplay

import (
	"context"
	"fmt"
	"qiantie/backend/internal/prompt"
	"strings"
)

func storyboardPlanningRoundLimit(in StoryboardPlanRequest) int {
	n := 8
	if in.MaxRounds != nil && *in.MaxRounds != 0 {
		n = *in.MaxRounds
	}
	if n < 1 {
		n = 1
	}
	if n > 8 {
		n = 8
	}
	if in.AutoContinue != nil && !*in.AutoContinue {
		n = 1
	}
	return n
}

type storyboardPlanRun struct {
	Text              string
	Rounds            int
	ContinuationError string
	Appended          bool
}

func generateStoryboardPlanRounds(ctx context.Context, generator TextGenerator, model, system, user, density string, limit int, observers ...func(PlanningProgress)) (storyboardPlanRun, error) {
	var out storyboardPlanRun
	base := []TextMessage{{Role: "system", Content: system}, {Role: "user", Content: user}}
	messages := append([]TextMessage(nil), base...)
	empty := 0
	for round := 1; round <= limit; round++ {
		if ctx.Err() != nil {
			return out, ctx.Err()
		}
		for _, observe := range observers {
			if observe != nil {
				observe(planningProgressFrame(round, out.Text, density, "requesting"))
			}
		}
		if ctx.Err() != nil {
			return out, ctx.Err()
		}
		request := TextRequest{Model: model, SystemPrompt: system, UserPrompt: messages[len(messages)-1].Content, Messages: append([]TextMessage(nil), messages...), MaxOutputTokens: 16000, Temperature: .4}
		response, err := generator.Generate(ctx, request)
		out.Rounds = round
		if ctx.Err() != nil {
			return out, ctx.Err()
		}
		if err != nil {
			if out.Text == "" {
				return out, err
			}
			out.ContinuationError = err.Error()
			break
		}
		chunk := strings.TrimSpace(response.Content)
		if chunk == "" {
			if out.Text == "" {
				return out, fmt.Errorf("%w: 规划生成为空", ErrArtifactInvalid)
			}
			empty++
			if empty >= 2 {
				out.ContinuationError = "续写连续两次未返回正文，已停止续写"
				break
			}
			continue
		}
		empty = 0
		next := appendStoryboardPlanChunk(out.Text, chunk)
		if len([]rune(next)) > 100000 {
			err = fmt.Errorf("%w: 规划正文过长", ErrArtifactInvalid)
			if out.Text == "" {
				return out, err
			}
			out.ContinuationError = err.Error()
			break
		}
		out.Text = next
		for _, observe := range observers {
			if observe != nil {
				observe(planningProgressFrame(round, out.Text, density, "received"))
			}
		}
		if EvaluateStoryboardPlan(out.Text, density).OK {
			break
		}
		messages = append(append([]TextMessage(nil), base...), TextMessage{Role: "assistant", Content: out.Text}, TextMessage{Role: "user", Content: prompt.ScreenplayNativePlanContinue(out.Text)})
	}
	return out, nil
}
