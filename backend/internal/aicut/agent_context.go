package aicut

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

func (s *AgentService) ConfigureSkills(resolve func(string) (string, bool)) { s.resolveSkill = resolve }

// Only prior user/assistant messages from this timeline are conversation context.
// Tool snapshots are intentionally fetched fresh rather than reused across runs.
func (s *AgentService) conversationHistory(ctx context.Context, run AgentRun) ([]string, error) {
	var messages []AgentMessage
	var err error
	if reader, ok := s.store.(interface {
		AgentConversationMessages(context.Context, string, string, string, int) ([]AgentMessage, error)
	}); ok {
		messages, err = reader.AgentConversationMessages(ctx, run.Owner, run.ProjectID, run.TimelineID, 40)
	} else {
		messages, err = s.store.AgentMessages(ctx, run.Owner, run.ProjectID, 200)
	}
	if err != nil {
		return nil, err
	}
	history := []string{}
	remaining := 12000
	for i := len(messages) - 1; i >= 0 && len(history) < 20 && remaining > 0; i-- {
		item := messages[i]
		if item.RunID == run.ID || (item.Role != AgentRoleUser && item.Role != AgentRoleAssistant) {
			continue
		}
		previous, err := s.store.AgentRun(ctx, run.Owner, run.ProjectID, item.RunID)
		if err != nil || canonicalASRTimeline(previous.TimelineID) != canonicalASRTimeline(run.TimelineID) {
			continue
		}
		content := []rune(item.Content)
		if len(content) > remaining {
			content = content[:remaining]
		}
		remaining -= len(content)
		history = append(history, "历史"+string(item.Role)+"："+string(content))
	}
	for l, r := 0, len(history)-1; l < r; l, r = l+1, r-1 {
		history[l], history[r] = history[r], history[l]
	}
	return history, nil
}

type AgentFollowup struct {
	Question string               `json:"question"`
	Fields   []AgentFollowupField `json:"fields"`
}
type AgentFollowupField struct {
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	Required bool     `json:"required"`
	Options  []string `json:"options,omitempty"`
}

func normalizeAgentFollowup(value any) (AgentFollowup, error) {
	encoded, err := json.Marshal(value)
	if err != nil || len(encoded) > 16000 {
		return AgentFollowup{}, fmt.Errorf("invalid followup")
	}
	var question AgentFollowup
	if err := json.Unmarshal(encoded, &question); err != nil {
		return question, err
	}
	question.Question = strings.TrimSpace(question.Question)
	if question.Question == "" || len(question.Fields) > 6 {
		return question, fmt.Errorf("invalid followup")
	}
	if len(question.Fields) == 0 {
		question.Fields = []AgentFollowupField{{ID: "answer", Label: "补充说明", Required: true}}
	}
	seen := map[string]bool{}
	for _, field := range question.Fields {
		if strings.TrimSpace(field.ID) == "" || strings.TrimSpace(field.Label) == "" || seen[field.ID] || len(field.Options) > 12 {
			return question, fmt.Errorf("invalid followup field")
		}
		seen[field.ID] = true
	}
	return question, nil
}
