package novelcreation

import "strings"

const (
	defaultTargetChapterWords = 2000
	minTargetChapterWords     = 800
	maxTargetChapterWords     = 8000

	defaultMaxRevisionRounds = 2
	minMaxRevisionRounds     = 1
	maxMaxRevisionRounds     = 5

	maxKnowledgeMaterials      = 8
	maxKnowledgeMaterialRunes  = 12000
	minScheduleIntervalMinutes = 15
	maxScheduleIntervalMinutes = 10080
	minBatchChapters           = 1
	maxBatchChapters           = 50
	defaultBatchPrompt         = "继续写下一章，保持已有设定、人物关系和文风。"
)

func NormalizeCreateProjectInput(input CreateProjectInput) (CreateProjectInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Genre = strings.TrimSpace(input.Genre)
	input.World = strings.TrimSpace(input.World)
	input.Protagonist = strings.TrimSpace(input.Protagonist)
	input.CoreConflict = strings.TrimSpace(input.CoreConflict)
	input.ModelID = strings.TrimSpace(input.ModelID)

	if input.TargetChapterWords == 0 {
		input.TargetChapterWords = defaultTargetChapterWords
	}
	if input.TargetChapterWords < minTargetChapterWords || input.TargetChapterWords > maxTargetChapterWords {
		return CreateProjectInput{}, &DomainError{
			Code:    ErrInputRequired,
			Message: "target chapter words must be between 800 and 8000",
		}
	}
	return input, nil
}

func NormalizeSendMessageInput(input SendMessageInput) (SendMessageInput, error) {
	input.SessionID = strings.TrimSpace(input.SessionID)
	input.ProjectID = strings.TrimSpace(input.ProjectID)
	input.Content = strings.TrimSpace(input.Content)
	input.ModelID = strings.TrimSpace(input.ModelID)
	input.ThinkingLevel = strings.TrimSpace(input.ThinkingLevel)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)

	for _, required := range []struct {
		field string
		value string
	}{
		{field: "sessionId", value: input.SessionID},
		{field: "content", value: input.Content},
		{field: "modelId", value: input.ModelID},
		{field: "idempotencyKey", value: input.IdempotencyKey},
	} {
		if required.value == "" {
			return SendMessageInput{}, &DomainError{
				Code:    ErrInputRequired,
				Message: required.field + " is required",
			}
		}
	}
	return input, nil
}

// ModeInputField describes one mode-specific input the user must supply before
// a run can start. The frontend renders the same catalog through
// GET /api/novel-creation/modes.
type ModeInputField struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Required bool   `json:"required"`
}

var modeInputFieldCatalog = map[string][]ModeInputField{
	"long_imitation":  {{Key: "referenceText", Label: "参考文本", Required: true}},
	"short_imitation": {{Key: "referenceText", Label: "参考文本", Required: true}},
	"fanfiction":      {{Key: "existingSetting", Label: "既有人物与世界规则", Required: true}},
	"continuation":    {{Key: "previousText", Label: "被续写的前文", Required: true}},
}

// ModeInputFields returns the input contract for a creation mode. Modes
// without extra inputs return nil.
func ModeInputFields(mode string) []ModeInputField {
	fields := modeInputFieldCatalog[strings.TrimSpace(mode)]
	return append([]ModeInputField(nil), fields...)
}

// NormalizeModeInputs trims and keeps only the keys that belong to the session
// mode, then enforces required fields. Unknown keys are dropped so arbitrary
// browser payloads cannot reach the prompt snapshot.
func NormalizeModeInputs(mode string, inputs map[string]string) (map[string]string, error) {
	fields := modeInputFieldCatalog[strings.TrimSpace(mode)]
	if len(fields) == 0 {
		return nil, nil
	}
	normalized := make(map[string]string, len(fields))
	for _, field := range fields {
		if value := strings.TrimSpace(inputs[field.Key]); value != "" {
			normalized[field.Key] = value
		}
	}
	for _, field := range fields {
		if field.Required && normalized[field.Key] == "" {
			return nil, &DomainError{Code: ErrInputRequired, Message: "请填写" + field.Label}
		}
	}
	if len(normalized) == 0 {
		return nil, nil
	}
	return normalized, nil
}

// modeUsesCheckpoint reports whether a fresh blueprint must be confirmed before
// the pipeline writes chapter bodies. Continuations and short forms run
// straight through.
func modeUsesCheckpoint(mode string) bool {
	switch strings.TrimSpace(mode) {
	case "long_original", "long_imitation":
		return true
	default:
		return false
	}
}

func CanTransition(from, to RunStatus) bool {
	if !validRunStatus(from) || !validRunStatus(to) {
		return false
	}
	if from == to {
		return true
	}
	if isTerminalRunStatus(from) {
		return false
	}
	if to == RunFailed || to == RunCancelled || to == RunPaused {
		return true
	}

	switch from {
	case RunDraft:
		return to == RunClarifying || to == RunQueued
	case RunClarifying:
		return to == RunAwaitingConfirmation || to == RunQueued
	case RunAwaitingConfirmation:
		return to == RunQueued || to == RunSucceeded
	case RunQueued:
		return to == RunPlanning
	case RunPlanning:
		return to == RunWriting
	case RunWriting:
		return to == RunNormalizing
	case RunNormalizing:
		return to == RunAuditing
	case RunAuditing:
		return to == RunRevising || to == RunSettling
	case RunRevising:
		return to == RunAuditing || to == RunSettling
	case RunSettling:
		return to == RunIndexing
	case RunIndexing:
		return to == RunSucceeded || to == RunSucceededWithWarnings
	default:
		return false
	}
}

func validRunStatus(status RunStatus) bool {
	switch status {
	case RunPaused, RunDraft, RunClarifying, RunAwaitingConfirmation, RunQueued, RunPlanning,
		RunWriting, RunNormalizing, RunAuditing, RunRevising, RunSettling, RunIndexing,
		RunSucceeded, RunSucceededWithWarnings, RunFailed, RunCancelled:
		return true
	default:
		return false
	}
}

func isTerminalRunStatus(status RunStatus) bool {
	switch status {
	case RunSucceeded, RunSucceededWithWarnings, RunFailed, RunCancelled:
		return true
	default:
		return false
	}
}

// NormalizeKnowledgeIDs trims, deduplicates, and bounds selected materials.
func NormalizeKnowledgeIDs(values []string) ([]string, error) {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		id := strings.TrimSpace(value)
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
	}
	if len(normalized) > maxKnowledgeMaterials {
		return nil, &DomainError{Code: ErrInputRequired, Message: "单次最多选择 8 个素材"}
	}
	return normalized, nil
}

// NormalizeAgentSettings clamps the project-level agent controls. A fallback
// target word count keeps older projects aligned with their creation setting.
func NormalizeAgentSettings(input AgentSettings, fallbackTargetWords int) AgentSettings {
	target := input.TargetChapterWords
	if target == 0 {
		target = fallbackTargetWords
	}
	if target == 0 {
		target = defaultTargetChapterWords
	}
	if target < minTargetChapterWords {
		target = minTargetChapterWords
	}
	if target > maxTargetChapterWords {
		target = maxTargetChapterWords
	}
	rounds := input.MaxRevisionRounds
	if rounds == 0 {
		rounds = defaultMaxRevisionRounds
	}
	if rounds < minMaxRevisionRounds {
		rounds = minMaxRevisionRounds
	}
	if rounds > maxMaxRevisionRounds {
		rounds = maxMaxRevisionRounds
	}
	return AgentSettings{TargetChapterWords: target, MaxRevisionRounds: rounds, ThinkingLevel: normalizeThinkingLevel(input.ThinkingLevel)}
}

func normalizeThinkingLevel(value string) string {
	switch strings.TrimSpace(value) {
	case "fast", "standard", "deep":
		return strings.TrimSpace(value)
	default:
		return "standard"
	}
}

func NormalizeMaterialInput(input CreateMaterialInput) (CreateMaterialInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Content = strings.TrimSpace(input.Content)
	if input.Content == "" {
		return CreateMaterialInput{}, &DomainError{Code: ErrInputRequired, Message: "素材内容不能为空"}
	}
	if input.Title == "" {
		input.Title = truncateRunes(input.Content, 40)
	}
	return input, nil
}

func NormalizeNoteInput(input ProjectNoteInput) (ProjectNoteInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Content = strings.TrimSpace(input.Content)
	if input.Title == "" && input.Content == "" {
		return ProjectNoteInput{}, &DomainError{Code: ErrInputRequired, Message: "备忘录标题或内容不能为空"}
	}
	if input.Title == "" {
		input.Title = truncateRunes(input.Content, 40)
	}
	return input, nil
}

func NormalizeBatchInput(input BatchInput) (BatchInput, error) {
	input.SessionID = strings.TrimSpace(input.SessionID)
	input.Prompt = strings.TrimSpace(input.Prompt)
	input.ModelID = strings.TrimSpace(input.ModelID)
	input.ThinkingLevel = normalizeThinkingLevel(input.ThinkingLevel)
	if input.SessionID == "" {
		return BatchInput{}, &DomainError{Code: ErrInputRequired, Message: "请选择批量续写使用的会话"}
	}
	if input.TotalChapters < minBatchChapters || input.TotalChapters > maxBatchChapters {
		return BatchInput{}, &DomainError{Code: ErrInputRequired, Message: "批量续写章节数必须在 1 到 50 之间"}
	}
	if input.Prompt == "" {
		input.Prompt = defaultBatchPrompt
	}
	return input, nil
}

func NormalizeScheduleInput(input ScheduleInput) (ScheduleInput, error) {
	input.SessionID = strings.TrimSpace(input.SessionID)
	input.Prompt = strings.TrimSpace(input.Prompt)
	input.ModelID = strings.TrimSpace(input.ModelID)
	input.ThinkingLevel = normalizeThinkingLevel(input.ThinkingLevel)
	if input.SessionID == "" {
		return ScheduleInput{}, &DomainError{Code: ErrInputRequired, Message: "请选择定时任务使用的会话"}
	}
	if input.Prompt == "" {
		return ScheduleInput{}, &DomainError{Code: ErrInputRequired, Message: "定时写作指令不能为空"}
	}
	if input.IntervalMinutes == 0 {
		input.IntervalMinutes = 60
	}
	if input.IntervalMinutes < minScheduleIntervalMinutes || input.IntervalMinutes > maxScheduleIntervalMinutes {
		return ScheduleInput{}, &DomainError{Code: ErrInputRequired, Message: "定时间隔必须在 15 到 10080 分钟之间"}
	}
	return input, nil
}

// truncateRunes keeps the snapshot and prompt bounded without splitting UTF-8.
func truncateRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= limit {
		return string(runes)
	}
	return string(runes[:limit])
}
