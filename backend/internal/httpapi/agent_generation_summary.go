package httpapi

import "qiantie/backend/internal/agentworkspace"

// AgentGenerationSummary exposes persisted task counts without guessing from
// canvas nodes. A missing task therefore never appears as completed.
type AgentGenerationSummary struct {
	Total                int `json:"total"`
	Queued               int `json:"queued"`
	Running              int `json:"running"`
	AwaitingConfirmation int `json:"awaitingConfirmation"`
	Completed            int `json:"completed"`
	Failed               int `json:"failed"`
	Cancelled            int `json:"cancelled"`
}

func SummarizeAgentGenerations(tasks []agentworkspace.GenerationTask) AgentGenerationSummary {
	summary := AgentGenerationSummary{Total: len(tasks)}
	for _, task := range tasks {
		switch task.State {
		case agentworkspace.GenerationQueued:
			summary.Queued++
		case agentworkspace.GenerationRunning:
			summary.Running++
		case agentworkspace.GenerationAwaitingConfirmation:
			summary.AwaitingConfirmation++
		case agentworkspace.GenerationCompleted:
			summary.Completed++
		case agentworkspace.GenerationFailed:
			summary.Failed++
		case agentworkspace.GenerationCancelled:
			summary.Cancelled++
		}
	}
	return summary
}
