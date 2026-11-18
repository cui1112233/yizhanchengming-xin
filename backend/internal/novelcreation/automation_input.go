package novelcreation

import (
	"context"
	"encoding/json"
	"strings"
)

// AutomationInput is shared by manual submissions and saved automation. Saved
// plans retain selections; each new run resolves current content once. Retries
// reuse the run snapshot, including content removed since the original run.
type AutomationInput struct {
	Creation           *AutomationCreation `json:"creation,omitempty"`
	State              *AutomationState    `json:"automationState,omitempty"`
	ModeInputs         map[string]string   `json:"modeInputs"`
	KnowledgeIDs       []string            `json:"knowledgeIds"`
	NoteIDs            []string            `json:"noteIds"`
	StyleProfileID     string              `json:"styleProfileId"`
	ConfigurationError string              `json:"configurationError,omitempty"`
}

type generationMaterials struct {
	AutomationInput
	Knowledge    []KnowledgeMaterial
	Notes        []ProjectNote
	StyleProfile string
}

func (s *Service) prepareGenerationInput(ctx context.Context, owner, projectID, sessionID string, input AutomationInput) (generationMaterials, error) {
	result := generationMaterials{}
	session, err := writableSession(ctx, s.store, owner, projectID, sessionID)
	if err != nil {
		return result, err
	}
	if _, err = s.store.GetProject(ctx, owner, projectID); err != nil {
		return result, err
	}
	input.StyleProfileID = strings.TrimSpace(input.StyleProfileID)
	settings, err := s.store.GetProjectSettings(ctx, owner, projectID)
	if err != nil {
		return result, err
	}
	result.StyleProfile = settings.StyleProfile
	if input.StyleProfileID != "" {
		profile, e := s.store.GetStyleProfile(ctx, owner, input.StyleProfileID)
		if e != nil {
			return result, e
		}
		if strings.TrimSpace(profile.Content) == "" {
			return result, &DomainError{Code: ErrInputRequired, Message: "所选文风内容为空，请重新选择"}
		}
		result.StyleProfile = profile.Content
	}
	input.KnowledgeIDs, err = NormalizeKnowledgeIDs(input.KnowledgeIDs)
	if err != nil {
		return result, err
	}
	materials, err := s.store.ListMaterials(ctx, owner, projectID)
	if err != nil {
		return result, err
	}
	legacyIDs := []string{}
	for _, id := range input.KnowledgeIDs {
		if !documentID(id) {
			legacyIDs = append(legacyIDs, id)
		}
	}
	result.Knowledge, err = selectKnowledgeMaterials(materials, legacyIDs)
	if err != nil {
		return result, err
	}
	for _, id := range input.KnowledgeIDs {
		if documentID(id) {
			j, e := getDocument(ctx, s.store, owner, projectID, strings.TrimPrefix(id, "docref-"))
			if e != nil {
				return result, e
			}
			if j.Selection == nil {
				return result, inputError("请先选择资料章节范围")
			}
			selected, e := selectedDocument(j, DocumentSelection{ExpectedVersion: j.Selection.SourceVersion, ChapterIDs: j.Selection.SelectedChapterIDs})
			if e != nil {
				return result, e
			}
			result.Knowledge = append(result.Knowledge, selected)
		}
	}
	byKnowledgeID := map[string]KnowledgeMaterial{}
	for _, k := range result.Knowledge {
		byKnowledgeID[k.ID] = k
	}
	result.Knowledge = nil
	for _, id := range input.KnowledgeIDs {
		result.Knowledge = append(result.Knowledge, byKnowledgeID[id])
	}
	input.NoteIDs, err = NormalizeKnowledgeIDs(input.NoteIDs)
	if err != nil {
		return result, &DomainError{Code: ErrInputRequired, Message: "单次最多选择 8 条备忘录"}
	}
	if len(input.NoteIDs) > 0 {
		notes, e := s.store.ListNotes(ctx, owner, projectID)
		if e != nil {
			return result, e
		}
		byID := map[string]ProjectNote{}
		for _, n := range notes {
			byID[n.ID] = n
		}
		for _, id := range input.NoteIDs {
			n, ok := byID[id]
			if !ok {
				return result, &DomainError{Code: ErrInputRequired, Message: "所选备忘录已删除或不属于当前作品，请重新选择"}
			}
			result.Notes = append(result.Notes, n)
		}
	}
	// An explicitly selected style or reference material satisfies imitation's
	// alternative input path. Other modes keep their own required fields.
	modeInputs := input.ModeInputs
	if session.Mode == longImitationKind {
		creation, e := loadLongImitation(ctx, s.store, owner, projectID)
		if e != nil {
			return result, e
		}
		if creation != nil {
			modeInputs = map[string]string{"referenceText": creation.EffectiveReferenceText}
		}
	}
	if (session.Mode == "long_imitation" || session.Mode == "short_imitation") && strings.TrimSpace(modeInputs["referenceText"]) == "" {
		if input.StyleProfileID != "" || len(result.Knowledge) > 0 {
			input.ModeInputs = nil
		} else {
			return result, &DomainError{Code: ErrInputRequired, Message: "请导入参考文本或选择文风"}
		}
	} else {
		input.ModeInputs, err = NormalizeModeInputs(session.Mode, modeInputs)
		if err != nil {
			return result, err
		}
	}
	input.ConfigurationError = ""
	result.AutomationInput = input
	return result, nil
}

func scheduleInput(v Schedule) AutomationInput {
	return AutomationInput{ModeInputs: v.ModeInputs, KnowledgeIDs: v.KnowledgeIDs, NoteIDs: v.NoteIDs, StyleProfileID: v.StyleProfileID, ConfigurationError: v.ConfigurationError}
}
func batchInput(v Batch) AutomationInput {
	return AutomationInput{ModeInputs: v.ModeInputs, KnowledgeIDs: v.KnowledgeIDs, NoteIDs: v.NoteIDs, StyleProfileID: v.StyleProfileID}
}
func scheduleStoredInput(v Schedule) AutomationInput {
	in := scheduleInput(v)
	in.Creation = v.Creation
	in.State = &v.AutomationState
	return in
}
func batchStoredInput(v Batch) AutomationInput {
	in := batchInput(v)
	in.Creation = v.Creation
	in.State = &v.AutomationState
	return in
}
func inputJSON(v AutomationInput) string { b, _ := json.Marshal(v); return string(b) }
func cloneSchedule(v Schedule) Schedule {
	var in AutomationInput
	_ = json.Unmarshal([]byte(inputJSON(scheduleStoredInput(v))), &in)
	if in.State != nil {
		v.AutomationState = *in.State
	}
	v.ModeInputs = in.ModeInputs
	v.Creation = in.Creation
	v.KnowledgeIDs = in.KnowledgeIDs
	v.NoteIDs = in.NoteIDs
	if v.LastRunAt != nil {
		t := *v.LastRunAt
		v.LastRunAt = &t
	}
	return v
}
func cloneBatch(v Batch) Batch {
	var in AutomationInput
	_ = json.Unmarshal([]byte(inputJSON(batchStoredInput(v))), &in)
	if in.State != nil {
		v.AutomationState = *in.State
	}
	v.ModeInputs = in.ModeInputs
	v.Creation = in.Creation
	v.KnowledgeIDs = in.KnowledgeIDs
	v.NoteIDs = in.NoteIDs
	return v
}

func (s *MemoryStore) GetStyleProfile(_ context.Context, owner, id string) (StyleProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.styles[memoryKey(owner, strings.TrimSpace(id))]
	if !ok {
		return StyleProfile{}, &DomainError{Code: ErrInputRequired, Message: "所选文风不存在或无权使用，请重新选择"}
	}
	return p, nil
}
func (s *MemoryStore) SaveStyleProfile(_ context.Context, owner string, p StyleProfile) (StyleProfile, error) {
	if _, err := requiredOwner(owner); err != nil {
		return StyleProfile{}, err
	}
	p, err := normalizeStyleProfile(p)
	if err != nil {
		return p, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if p.ID == "" {
		p.ID = s.newID("style")
	}
	s.styles[memoryKey(owner, p.ID)] = p
	return p, nil
}
func normalizeStyleProfile(p StyleProfile) (StyleProfile, error) {
	p.ID = strings.TrimSpace(p.ID)
	p.Title = strings.TrimSpace(p.Title)
	p.Content = strings.TrimSpace(p.Content)
	if p.Content == "" {
		return p, &DomainError{Code: ErrInputRequired, Message: "文风内容不能为空"}
	}
	return p, nil
}
func (s *Service) SaveStyleProfile(ctx context.Context, owner string, p StyleProfile) (StyleProfile, error) {
	p, err := normalizeStyleProfile(p)
	if err != nil {
		return p, err
	}
	if p.ID != "" {
		if _, err = s.store.GetStyleProfile(ctx, owner, p.ID); err != nil {
			return p, err
		}
	}
	return s.store.SaveStyleProfile(ctx, strings.TrimSpace(owner), p)
}
func (s *Service) GetStyleProfile(ctx context.Context, owner, id string) (StyleProfile, error) {
	return s.store.GetStyleProfile(ctx, strings.TrimSpace(owner), strings.TrimSpace(id))
}
