package novelcreation

import (
	catalog "qiantie/backend/internal/prompt"
	novelprompt "qiantie/backend/internal/prompt/novelcreation"
	"strings"
)

type FrozenSystemPrompt struct {
	ID       string `json:"id"`
	Version  int    `json:"version"`
	Revision int    `json:"revision"`
	Body     string `json:"body"`
}
type SystemPromptSnapshot struct {
	Version int                           `json:"version"`
	Entries map[string]FrozenSystemPrompt `json:"entries"`
}

func (s *Service) ConfigureSystemPrompts(c *catalog.Catalog) { s.promptCatalog = c }
func (s *Service) snapshotSystemPrompts() (*SystemPromptSnapshot, error) {
	if s.promptCatalog == nil {
		return nil, nil
	}
	rows, e := s.promptCatalog.PublishedBodies(novelprompt.AdminModule)
	if e != nil {
		return nil, &DomainError{Code: ErrPersistenceFailed, Message: "系统预设词读取失败，未创建生成任务", Cause: e}
	}
	known := map[string]bool{}
	for _, p := range novelprompt.AdminPrompts() {
		known[p.ID] = true
	}
	snap := &SystemPromptSnapshot{Version: 1, Entries: map[string]FrozenSystemPrompt{}}
	for _, p := range rows {
		if known[p.ID] {
			snap.Entries[p.ID] = FrozenSystemPrompt{p.ID, p.Version, p.Revision, p.Body}
		}
	}
	return snap, nil
}
func frozenSystemText(in RunContext, key, fallback string) (string, error) {
	if in.SystemPrompts == nil {
		return fallback, nil
	}
	if in.SystemPrompts.Version != 1 {
		return "", &DomainError{Code: ErrPersistenceFailed, Message: "系统预设词快照版本无效，已停止生成"}
	}
	p, ok := in.SystemPrompts.Entries[key]
	if !ok || p.ID != key || p.Version < 1 || p.Revision < 1 || strings.TrimSpace(p.Body) == "" {
		return "", &DomainError{Code: ErrModelUnavailable, Message: "系统预设词未发布或为空：" + key + "，未调用该步骤模型"}
	}
	return p.Body, nil
}
func frozenSystemDefinition(in RunContext, d catalog.Definition, key string) (catalog.Definition, error) {
	body, e := frozenSystemText(in, key, d.SystemPrompt)
	if e != nil {
		return d, e
	}
	d.SystemPrompt = body
	return d, nil
}
func systemPresetKey(d catalog.Definition) string {
	if d.ID == novelprompt.TruthSettlement.ID && d.Version >= 2 {
		return novelprompt.IncrementalSettlementPresetID
	}
	return d.ID
}
