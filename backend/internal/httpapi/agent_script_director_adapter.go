package httpapi

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/batchfactoryv11"
)

// adaptDirectorResultToScriptShots keeps one Batch Factory VIDEO as one Agent
// Canvas generation unit. The full director timeline remains in VideoPrompt,
// so image, video, and TTS nodes continue to consume the same ScriptShot data.
func adaptDirectorResultToScriptShots(scriptNodeID string, result batchfactoryv11.DirectorResult) ([]agentworkspace.ScriptShot, error) {
	scriptNodeID = strings.TrimSpace(scriptNodeID)
	if scriptNodeID == "" || len(result.Storyboard) == 0 {
		return nil, agentworkspace.ErrScriptShotInvalid
	}
	shots := make([]agentworkspace.ScriptShot, 0, len(result.Storyboard))
	for index, video := range result.Storyboard {
		videoPrompt := renderAgentDirectorVideoPrompt(video)
		if videoPrompt == "" {
			return nil, fmt.Errorf("%w: director video %d has no prompt", agentworkspace.ErrScriptShotInvalid, index+1)
		}
		imagePrompt := strings.TrimSpace(video.VisualPrompt)
		if imagePrompt == "" {
			imagePrompt = strings.TrimSpace(strings.Join(nonEmptyDirectorText(video.Scene, video.VideoDesc), "，"))
		}
		shots = append(shots, agentworkspace.ScriptShot{
			ID:          fmt.Sprintf("%s-shot-%d", scriptNodeID, index+1),
			Sequence:    index + 1,
			Scene:       strings.TrimSpace(video.Scene),
			Characters:  strings.Join(nonEmptyDirectorText(video.Characters...), "、"),
			Narration:   renderAgentDirectorNarration(video),
			ImagePrompt: imagePrompt,
			VideoPrompt: videoPrompt,
			TTSText:     renderAgentDirectorAudio(video),
		})
	}
	return shots, nil
}

func renderAgentDirectorVideoPrompt(video batchfactoryv11.DirectorVideo) string {
	if prompt := strings.TrimSpace(video.FinalPrompt); prompt != "" {
		return prompt
	}
	if len(video.Shots) == 0 {
		return strings.TrimSpace(video.VideoDesc)
	}
	lines := []string{"镜头画面："}
	for _, shot := range video.Shots {
		lines = append(lines, fmt.Sprintf("%02d:%02d-%02d:%02d | %s｜%s | %s", shot.StartSec/60, shot.StartSec%60, shot.EndSec/60, shot.EndSec%60, strings.TrimSpace(shot.ShotType), strings.TrimSpace(shot.Camera), strings.TrimSpace(shot.Description)))
	}
	return strings.Join(lines, "\n")
}

func renderAgentDirectorNarration(video batchfactoryv11.DirectorVideo) string {
	parts := make([]string, 0, len(video.Shots)+1)
	if summary := strings.TrimSpace(video.VideoDesc); summary != "" {
		parts = append(parts, summary)
	}
	for _, shot := range video.Shots {
		if description := strings.TrimSpace(shot.Description); description != "" {
			parts = append(parts, description)
		}
	}
	return strings.Join(nonEmptyDirectorText(parts...), "；")
}

func renderAgentDirectorAudio(video batchfactoryv11.DirectorVideo) string {
	audio := make([]string, 0, len(video.Shots))
	for _, shot := range video.Shots {
		value := strings.TrimSpace(shot.Audio)
		if value != "" && value != "无" {
			audio = append(audio, value)
		}
	}
	return strings.Join(nonEmptyDirectorText(audio...), "\n")
}

func nonEmptyDirectorText(values ...string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func buildAgentDirectorContract(scriptNodeID, title, sourceText string) (batchfactoryv11.PromptContract, error) {
	book := batchfactoryv11.Book{
		ID:         strings.TrimSpace(scriptNodeID),
		Title:      strings.TrimSpace(title),
		SourceText: strings.TrimSpace(sourceText),
		Mode:       "original",
	}
	snapshot := batchfactoryv11.DirectorSnapshot{
		Effective:        batchfactoryv11.SettingsPatch{},
		Mode:             "original",
		MaxVideoDuration: 15,
		AspectRatio:      "9:16",
	}
	contract, err := batchfactoryv11.BuildDirectorContract(book, batchfactoryv11.HookRevision{}, snapshot)
	if err != nil {
		return batchfactoryv11.PromptContract{}, err
	}
	// The canvas creates image nodes directly from each persisted ScriptShot,
	// so it needs the visual prompt together with the director timeline.
	contract.Normalization.RequireVisualPrompt = true
	contract.SystemPrompt += "\n每个 storyboard 必须额外输出非空 visual_prompt，作为同一 VIDEO 的图片生成提示词；它要和角色、场景、道具及镜头动作保持一致。"
	return contract, nil
}

func completeAgentDirectorContract(complete AgentTextCompletionExecutor, ctx context.Context, owner, modelID string, contract batchfactoryv11.PromptContract) (batchfactoryv11.DirectorResult, error) {
	if complete == nil {
		return batchfactoryv11.DirectorResult{}, errors.New("文字模型执行器未配置")
	}
	request := batchfactoryv11.TextCompletionRequest{SystemPrompt: contract.SystemPrompt, UserPrompt: contract.UserPrompt, Temperature: contract.Temperature, MaxTokens: contract.MaxTokens, DisableJSONResponse: contract.SDTextProtocol}
	completion, err := complete(ctx, owner, modelID, request)
	if err != nil {
		return batchfactoryv11.DirectorResult{}, err
	}
	result, parseErr := parseAgentDirectorCompletion(completion, contract)
	if parseErr == nil {
		return result, nil
	}
	request.SystemPrompt += "\n上一轮输出无法解析。现在只输出完整、合法的结果，不要解释，不要 Markdown。"
	request.Temperature = 0.2
	request.DisableJSONResponse = true
	retryCompletion, retryErr := complete(ctx, owner, modelID, request)
	if retryErr != nil {
		return batchfactoryv11.DirectorResult{}, parseErr
	}
	return parseAgentDirectorCompletion(retryCompletion, contract)
}

func parseAgentDirectorCompletion(completion string, contract batchfactoryv11.PromptContract) (batchfactoryv11.DirectorResult, error) {
	if contract.SDTextProtocol {
		return batchfactoryv11.ParseSDDirectorText(completion, contract.Normalization.MaxVideoDuration)
	}
	raw, err := batchfactoryv11.ParseDirectorJSON(completion)
	if err != nil {
		return batchfactoryv11.DirectorResult{}, err
	}
	return batchfactoryv11.NormalizeDirectorOutput(raw, contract.Normalization)
}
