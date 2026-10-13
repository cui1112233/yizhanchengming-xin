package batchfactoryv11

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const sdOriginalStoryVersion = "sd.original-story.v1"

// buildSDStoryScriptContract is the first text-model stage of the SD flow.
// It converts the complete source into one row per source line; it does not
// decide duration, shot count, or supplier prompt text.
func buildSDStoryScriptContract(book Book, snapshot DirectorSnapshot) PromptContract {
	lines := make([]map[string]any, 0)
	for index, line := range strings.Split(strings.ReplaceAll(book.SourceText, "\r\n", "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines = append(lines, map[string]any{
			"source_index":     index + 1,
			"source_text":      line,
			"source_span_keys": []string{fmt.Sprintf("L%04d-P01", index+1)},
		})
	}
	contract := `你是 SD_STORY_SCRIPT 阶段。完整阅读原文、统一风格和冻结资产，只把每一行原文转换成一行可视化故事剧本。
原文是事实和顺序的最高权威；不得新增关键人物、行为、因果、地点、数字或结局，不得把后文事件提前。
source_text 必须逐字原样返回；每个 source_index 只能对应一行；行数、顺序和 source_span_keys 不得改变。
故事剧本只负责画面转换，不负责时长、镜头数量、供应商 Prompt 或配音。
只返回合法 JSON：{"rows":[{"source_index":1,"source_text":"原文","story_script_line":"非语言可见画面","characters":[],"objects":[],"continuity_in":{"summary":""},"continuity_out":{"summary":""},"covered_source_span_keys":["L0001-P01"]}]}。
不要 Markdown、解释或自检。`
	payload := map[string]any{
		"story_timeline_version": sdOriginalStoryVersion,
		"line_contract":          lines,
		"source_text":            book.SourceText,
		"assets":                 payloadAssets(book),
		"style_system_analysis":  rawString(snapshot.Effective, "h3StyleAnalysis", ""),
	}
	encoded, _ := json.MarshalIndent(payload, "", "  ")
	return PromptContract{SystemPrompt: contract, UserPrompt: string(encoded), Temperature: 0.2, MaxTokens: 6000}
}

func validateSDStoryScriptRows(completion string, book Book) (json.RawMessage, error) {
	root, err := ParseDirectorJSON(completion)
	if err != nil {
		return nil, err
	}
	var envelope struct {
		Rows []struct {
			SourceIndex           int      `json:"source_index"`
			SourceText            string   `json:"source_text"`
			StoryScriptLine       string   `json:"story_script_line"`
			CoveredSourceSpanKeys []string `json:"covered_source_span_keys"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(root, &envelope); err != nil {
		return nil, err
	}
	expected := make([]string, 0)
	for index, line := range strings.Split(strings.ReplaceAll(book.SourceText, "\r\n", "\n"), "\n") {
		if strings.TrimSpace(line) != "" {
			expected = append(expected, line)
			_ = index
		}
	}
	if len(envelope.Rows) != len(expected) {
		return nil, fmt.Errorf("SD 故事剧本 rows 数量为 %d，期望 %d", len(envelope.Rows), len(expected))
	}
	for index, row := range envelope.Rows {
		if row.SourceIndex != index+1 || row.SourceText != expected[index] || strings.TrimSpace(row.StoryScriptLine) == "" || len(row.CoveredSourceSpanKeys) == 0 {
			return nil, fmt.Errorf("SD 故事剧本第 %d 行未完整覆盖原文契约", index+1)
		}
	}
	return root, nil
}

func shouldRunSDStoryScript(snapshot DirectorSnapshot, config AIReasoningPromptConfig) bool {
	return selectedSDVideoPreset(config.Video) && rawString(snapshot.Effective, "storyTimelineVersion", "") == sdOriginalStoryVersion
}

func (s *DirectorService) runSDStoryScript(ctx context.Context, book Book, snapshot DirectorSnapshot, config AIReasoningPromptConfig) (json.RawMessage, error) {
	contract := buildSDStoryScriptContract(book, snapshot)
	completion, err := s.Provider.Complete(ctx, TextCompletionRequest{SystemPrompt: contract.SystemPrompt, UserPrompt: contract.UserPrompt, Temperature: contract.Temperature, MaxTokens: contract.MaxTokens})
	if err != nil {
		return nil, err
	}
	return validateSDStoryScriptRows(completion, book)
}
