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
		return to == RunQueued
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
