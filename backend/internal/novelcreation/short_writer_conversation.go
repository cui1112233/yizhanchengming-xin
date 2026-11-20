package novelcreation

import (
	"encoding/json"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
	prompt "qiantie/backend/internal/prompt/novelcreation"
)

// This lives only for one piece. Replaying its completed draft call rebuilds
// the same context without another provider request or a second persisted copy.
type shortWriterConversation struct {
	System string
	User   string
	Draft  string
}

func shortWriterConversationCall(input RunContext, id string) bool {
	return input.ShortWriterConversationVersion == 1 && (id == prompt.ShortBodyPrompt.ID || id == prompt.ShortBodyRevision.ID)
}

func shortWriterRevisionRequest(input RunContext, request batchfactoryv11.TextCompletionRequest) (batchfactoryv11.TextCompletionRequest, error) {
	history := input.shortWriterConversation
	if history == nil || strings.TrimSpace(history.System) == "" || strings.TrimSpace(history.User) == "" || strings.TrimSpace(history.Draft) == "" {
		return request, inputError("缺少有效首稿写作对话，已停止修订")
	}
	followup, err := frozenSystemText(input, prompt.ShortWriterRevisionFollowupID, prompt.ShortWriterRevisionFollowupSystem)
	if err != nil {
		return request, err
	}
	user, err := marshalPromptPayload(map[string]any{"revision_instruction": followup, "configured_revision_instruction": request.SystemPrompt, "context": json.RawMessage(request.UserPrompt)})
	if err != nil {
		return request, err
	}
	request.SystemPrompt = history.System
	request.UserPrompt = user
	request.Messages = []batchfactoryv11.TextCompletionMessage{{Role: "system", Content: history.System}, {Role: "user", Content: history.User}, {Role: "assistant", Content: history.Draft}, {Role: "user", Content: user}}
	return request, nil
}

func rememberShortWriterDraft(input RunContext, id string, request batchfactoryv11.TextCompletionRequest, reply string) {
	if shortWriterConversationCall(input, id) && id == prompt.ShortBodyPrompt.ID && input.shortWriterConversation != nil {
		*input.shortWriterConversation = shortWriterConversation{System: request.SystemPrompt, User: request.UserPrompt, Draft: strings.TrimSpace(reply)}
	}
}
