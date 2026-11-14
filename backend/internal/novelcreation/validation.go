package novelcreation

import "strings"

const (
	defaultTargetChapterWords = 2000
	minTargetChapterWords     = 800
	maxTargetChapterWords     = 8000
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
	if to == RunFailed || to == RunCancelled {
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
	case RunDraft, RunClarifying, RunAwaitingConfirmation, RunQueued, RunPlanning,
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
