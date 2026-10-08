package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
)

func copyGenerationTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func projectBookRun(value generation.BookRun) generation.BookRun {
	outcome := generation.OutcomeForPersistedMessage(value.ErrorMessage)
	value.ErrorMessage, value.ErrorCode = outcome.Message, outcome.Code
	value.StartedAt, value.FinishedAt = copyGenerationTime(value.StartedAt), copyGenerationTime(value.FinishedAt)
	return value
}

func projectStageRun(value generation.StageRun) generation.StageRun {
	outcome := generation.OutcomeForPersistedMessage(value.ErrorMessage)
	value.ErrorMessage, value.ErrorCode = outcome.Message, outcome.Code
	value.StartedAt, value.FinishedAt = copyGenerationTime(value.StartedAt), copyGenerationTime(value.FinishedAt)
	value.InputSnapshot = projectGenerationSnapshot(value.InputSnapshot)
	value.ValidationResult = projectGenerationSnapshot(value.ValidationResult)
	return value
}

func projectStageMap(values map[generation.Stage]generation.StageRun) map[generation.Stage]generation.StageRun {
	if values == nil {
		return nil
	}
	projected := make(map[generation.Stage]generation.StageRun, len(values))
	for stage, value := range values {
		projected[stage] = projectStageRun(value)
	}
	return projected
}

func projectBookGeneration(value generation.BookGenerationResult) generation.BookGenerationResult {
	value.Run = projectBookRun(value.Run)
	if value.Stages != nil {
		stages := make([]generation.StageRun, len(value.Stages))
		for i, stage := range value.Stages {
			stages[i] = projectStageRun(stage)
		}
		value.Stages = stages
	}
	value.Latest = projectStageMap(value.Latest)
	outcome := generation.OutcomeForPersistedMessage(value.Error)
	value.Error, value.ErrorCode = outcome.Message, outcome.Code
	return value
}

func projectBatchGeneration(value generation.BatchGenerationResult) generation.BatchGenerationResult {
	if value.Books != nil {
		books := make([]generation.BatchBookResult, len(value.Books))
		for i, book := range value.Books {
			if book.Run != nil {
				run := projectBookRun(*book.Run)
				book.Run = &run
			}
			outcome := generation.OutcomeForPersistedMessage(book.Error)
			book.Error, book.ErrorCode = outcome.Message, outcome.Code
			books[i] = book
		}
		value.Books = books
	}
	return value
}

func projectGenerationSummary(value generation.ProjectSummary) generation.ProjectSummary {
	if value.Books != nil {
		books := make([]generation.BookGenerationSummary, len(value.Books))
		for i, book := range value.Books {
			if book.Run != nil {
				run := projectBookRun(*book.Run)
				book.Run = &run
			}
			book.Stages = projectStageMap(book.Stages)
			books[i] = book
		}
		value.Books = books
	}
	return value
}

// Snapshots remain intact in storage for retry. Only inspect structured public
// copies, retaining numbers as JSON tokens and leaving source/script/prompt
// string contents uninterpreted.
func projectGenerationSnapshot(snapshot string) string {
	raw := json.RawMessage(strings.TrimSpace(snapshot))
	if len(raw) == 0 || !json.Valid(raw) || (raw[0] != '{' && raw[0] != '[') {
		return ""
	}
	return string(projectDiagnosticJSON(raw))
}

func projectDiagnosticJSON(raw json.RawMessage) json.RawMessage {
	raw = json.RawMessage(strings.TrimSpace(string(raw)))
	switch raw[0] {
	case '{':
		var fields map[string]json.RawMessage
		_ = json.Unmarshal(raw, &fields)
		var outcome generation.Outcome
		hasError := false
		for _, key := range []string{"error", "errorMessage"} {
			if diagnostic, ok := fields[key]; ok {
				hasError = true
				if strings.TrimSpace(string(diagnostic)) == "null" {
					continue // No diagnostic; preserve the historical null shape.
				}
				var message string
				if json.Unmarshal(diagnostic, &message) != nil {
					message = "unknown diagnostic"
				}
				current := generation.OutcomeForPersistedMessage(message)
				fields[key], _ = json.Marshal(current.Message)
				if outcome.Code == "" {
					outcome = current
				} else if current.Code != "" && current.Code != outcome.Code {
					outcome = generation.OutcomeForPersistedMessage("unknown diagnostic")
				}
			}
		}
		for key, value := range fields {
			switch {
			case key == "errorCode" || (key == "code" && hasError):
				// Never trust a historical/provider-supplied diagnostic code.
				fields[key], _ = json.Marshal(outcome.Code)
			case key != "error" && key != "errorMessage":
				fields[key] = projectDiagnosticJSON(value)
			}
		}
		encoded, _ := json.Marshal(fields)
		return encoded
	case '[':
		var values []json.RawMessage
		_ = json.Unmarshal(raw, &values)
		for i, value := range values {
			values[i] = projectDiagnosticJSON(value)
		}
		encoded, _ := json.Marshal(values)
		return encoded
	default:
		return raw
	}
}

func writeBookGenerationOutcome(w http.ResponseWriter, value generation.BookGenerationResult, cause error) {
	value = projectBookGeneration(value)
	if cause == nil {
		writeJSON(w, http.StatusOK, value)
		return
	}
	outcome := generation.OutcomeForError(cause)
	value.Error, value.ErrorCode = outcome.Message, outcome.Code
	writeJSON(w, generationHTTPStatus(cause), struct {
		generation.BookGenerationResult
		generation.Outcome
	}{value, outcome})
}

func writeBatchGenerationOutcome(w http.ResponseWriter, value generation.BatchGenerationResult, cause error) {
	value = projectBatchGeneration(value)
	if cause == nil {
		writeJSON(w, http.StatusOK, value)
		return
	}
	outcome := generation.OutcomeForError(cause)
	if errors.Is(cause, generation.ErrProjectArchived) {
		writeJSON(w, http.StatusConflict, struct {
			generation.BatchGenerationResult
			generation.Outcome
			Error string `json:"error"`
		}{value, outcome, outcome.Message})
		return
	}
	writeJSON(w, http.StatusMultiStatus, struct {
		generation.BatchGenerationResult
		generation.Outcome
		Error string `json:"error"`
	}{value, outcome, outcome.Message})
}
