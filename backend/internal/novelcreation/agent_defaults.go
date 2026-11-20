package novelcreation

import (
	"context"
	"strings"
)

// Defaults are account scoped. Each new project copies them once; existing
// projects and queued task snapshots never read mutable defaults at execution.
func (s *Service) GetAgentDefaults(ctx context.Context, owner string) (AgentSettings, error) {
	owner, e := requiredOwner(owner)
	if e != nil {
		return AgentSettings{}, e
	}
	a, e := s.store.GetAgentDefaults(ctx, owner)
	if e != nil {
		return a, e
	}
	return withDefaultReviewDimensions(NormalizeAgentSettings(a, 2000)), nil
}
func (s *Service) UpdateAgentDefaults(ctx context.Context, owner string, a AgentSettings) (AgentSettings, error) {
	owner, e := requiredOwner(owner)
	if e != nil {
		return AgentSettings{}, e
	}
	if e = validateAgentRequirements(a); e != nil {
		return AgentSettings{}, e
	}
	// Explicit zero/false and an empty dimensions list are meaningful.
	a.Version = 1
	a = withDefaultReviewDimensions(NormalizeAgentSettings(a, 2000))
	return s.store.PutAgentDefaults(ctx, owner, a)
}
func inheritAgentDefaults(ctx context.Context, st Store, owner string, p Project) error {
	a, e := st.GetAgentDefaults(ctx, strings.TrimSpace(owner))
	if e != nil {
		return e
	}
	if a.Version == 0 {
		return nil
	}
	a = withDefaultReviewDimensions(NormalizeAgentSettings(a, p.TargetChapterWords))
	// Work-specific chapter lengths remain project inputs, not account settings.
	a.TargetChapterWords = p.TargetChapterWords
	_, e = st.UpdateProjectSettings(ctx, owner, p.ID, ProjectSettings{Agent: a})
	return e
}
