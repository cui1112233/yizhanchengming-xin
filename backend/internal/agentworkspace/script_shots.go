package agentworkspace

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"
)

var (
	ErrScriptShotNotFound = errors.New("agent script shot not found")
	ErrScriptShotInvalid  = errors.New("invalid agent script shot")
)

type ScriptShot struct {
	ID              string    `json:"id"`
	ProjectID       string    `json:"projectId"`
	ScriptNodeID    string    `json:"scriptNodeId"`
	Sequence        int       `json:"sequence"`
	Scene           string    `json:"scene"`
	Characters      string    `json:"characters"`
	Dialogue        string    `json:"dialogue"`
	Narration       string    `json:"narration"`
	ImagePrompt     string    `json:"imagePrompt"`
	VideoPrompt     string    `json:"videoPrompt"`
	TTSText         string    `json:"ttsText"`
	DurationSeconds int       `json:"durationSeconds"`
	SourceShotIDs   []string  `json:"sourceShotIds,omitempty"`
	Revision        int64     `json:"revision"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type ScriptSplit struct {
	ProjectID    string       `json:"projectId"`
	ScriptNodeID string       `json:"scriptNodeId"`
	Shots        []ScriptShot `json:"shots"`
}

type ShotGenerationLink struct {
	ProjectID    string `json:"projectId"`
	ScriptNodeID string `json:"scriptNodeId"`
	ShotID       string `json:"shotId"`
	NodeID       string `json:"nodeId"`
	GenerationID string `json:"generationId"`
	Attempt      int    `json:"attempt"`
	Active       bool   `json:"active"`
}

type ScriptShotStore interface {
	ReplaceShots(context.Context, string, string, string, []ScriptShot) error
	ListShots(context.Context, string, string, string) ([]ScriptShot, error)
	UpdateShot(context.Context, string, ScriptShot) (ScriptShot, error)
	AttachGeneration(context.Context, string, ShotGenerationLink) error
	ListGenerationLinks(context.Context, string, string, string) ([]ShotGenerationLink, error)
	CreateShotGeneration(context.Context, string, CreateShotGenerationInput) (Canvas, GenerationTask, ShotGenerationLink, error)
}

type CreateShotGenerationInput struct {
	ExpectedCanvasRevision int64
	Node                   Node
	Edge                   Edge
	Task                   CreateGenerationTaskInput
	Link                   ShotGenerationLink
	Now                    time.Time
}

func normalizeScriptShot(owner, projectID, scriptNodeID string, shot ScriptShot, now time.Time) (ScriptShot, error) {
	owner = strings.TrimSpace(owner)
	projectID = strings.TrimSpace(projectID)
	scriptNodeID = strings.TrimSpace(scriptNodeID)
	shot.ID = strings.TrimSpace(shot.ID)
	if shot.ProjectID == "" {
		shot.ProjectID = projectID
	}
	if shot.ScriptNodeID == "" {
		shot.ScriptNodeID = scriptNodeID
	}
	shot.ProjectID = strings.TrimSpace(shot.ProjectID)
	shot.ScriptNodeID = strings.TrimSpace(shot.ScriptNodeID)
	shot.Scene = strings.TrimSpace(shot.Scene)
	shot.Characters = strings.TrimSpace(shot.Characters)
	shot.Dialogue = strings.TrimSpace(shot.Dialogue)
	shot.Narration = strings.TrimSpace(shot.Narration)
	shot.ImagePrompt = strings.TrimSpace(shot.ImagePrompt)
	shot.VideoPrompt = strings.TrimSpace(shot.VideoPrompt)
	shot.TTSText = strings.TrimSpace(shot.TTSText)
	if shot.DurationSeconds < 0 {
		return ScriptShot{}, ErrScriptShotInvalid
	}
	seenSources := make(map[string]struct{}, len(shot.SourceShotIDs))
	sources := make([]string, 0, len(shot.SourceShotIDs))
	for _, sourceID := range shot.SourceShotIDs {
		sourceID = strings.TrimSpace(sourceID)
		if sourceID == "" {
			continue
		}
		if _, exists := seenSources[sourceID]; exists {
			continue
		}
		seenSources[sourceID] = struct{}{}
		sources = append(sources, sourceID)
	}
	shot.SourceShotIDs = sources
	if owner == "" || projectID == "" || scriptNodeID == "" || shot.ID == "" || shot.ProjectID != projectID || shot.ScriptNodeID != scriptNodeID || shot.Sequence < 1 {
		return ScriptShot{}, ErrScriptShotInvalid
	}
	if shot.Revision < 1 {
		shot.Revision = 1
	}
	if shot.CreatedAt.IsZero() {
		shot.CreatedAt = now
	}
	shot.CreatedAt = shot.CreatedAt.UTC().Truncate(time.Microsecond)
	shot.UpdatedAt = now.UTC().Truncate(time.Microsecond)
	return shot, nil
}

func scriptShotKey(owner, projectID, scriptNodeID, shotID string) string {
	return strings.TrimSpace(owner) + "\x00" + strings.TrimSpace(projectID) + "\x00" + strings.TrimSpace(scriptNodeID) + "\x00" + strings.TrimSpace(shotID)
}

func shotGenerationLinkKey(owner string, link ShotGenerationLink) string {
	return strings.TrimSpace(owner) + "\x00" + link.ProjectID + "\x00" + link.ShotID + "\x00" + link.NodeID + "\x00" + link.GenerationID
}

func normalizeShotGenerationLink(owner string, link ShotGenerationLink) (ShotGenerationLink, error) {
	owner = strings.TrimSpace(owner)
	link.ProjectID = strings.TrimSpace(link.ProjectID)
	link.ScriptNodeID = strings.TrimSpace(link.ScriptNodeID)
	link.ShotID = strings.TrimSpace(link.ShotID)
	link.NodeID = strings.TrimSpace(link.NodeID)
	link.GenerationID = strings.TrimSpace(link.GenerationID)
	if owner == "" || link.ProjectID == "" || link.ScriptNodeID == "" || link.ShotID == "" || link.NodeID == "" || link.GenerationID == "" || link.Attempt < 1 {
		return ShotGenerationLink{}, ErrScriptShotInvalid
	}
	return link, nil
}

func (s *MemoryStore) ReplaceShots(_ context.Context, owner, projectID, scriptNodeID string, shots []ScriptShot) error {
	now := time.Now().UTC()
	normalized := make([]ScriptShot, 0, len(shots))
	ids := make(map[string]struct{}, len(shots))
	sequences := make(map[int]struct{}, len(shots))
	for _, shot := range shots {
		shot, err := normalizeScriptShot(owner, projectID, scriptNodeID, shot, now)
		if err != nil {
			return err
		}
		if _, exists := ids[shot.ID]; exists {
			return ErrScriptShotInvalid
		}
		if _, exists := sequences[shot.Sequence]; exists {
			return ErrScriptShotInvalid
		}
		ids[shot.ID] = struct{}{}
		sequences[shot.Sequence] = struct{}{}
		normalized = append(normalized, shot)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	project, exists := s.projects[projectKey(owner, projectID)]
	if !exists || project.DeletedAt != nil {
		return ErrProjectNotFound
	}
	prefix := scriptShotKey(owner, projectID, scriptNodeID, "")
	for key := range s.scriptShots {
		if strings.HasPrefix(key, prefix) {
			delete(s.scriptShots, key)
		}
	}
	for _, shot := range normalized {
		s.scriptShots[scriptShotKey(owner, projectID, scriptNodeID, shot.ID)] = shot
	}
	return nil
}

func (s *MemoryStore) ListShots(_ context.Context, owner, projectID, scriptNodeID string) ([]ScriptShot, error) {
	owner, projectID, scriptNodeID = strings.TrimSpace(owner), strings.TrimSpace(projectID), strings.TrimSpace(scriptNodeID)
	if owner == "" || projectID == "" || scriptNodeID == "" {
		return nil, ErrScriptShotInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	prefix := scriptShotKey(owner, projectID, scriptNodeID, "")
	shots := make([]ScriptShot, 0)
	for key, shot := range s.scriptShots {
		if strings.HasPrefix(key, prefix) {
			shots = append(shots, shot)
		}
	}
	sort.Slice(shots, func(i, j int) bool {
		if shots[i].Sequence == shots[j].Sequence {
			return shots[i].ID < shots[j].ID
		}
		return shots[i].Sequence < shots[j].Sequence
	})
	return shots, nil
}

func (s *MemoryStore) UpdateShot(_ context.Context, owner string, shot ScriptShot) (ScriptShot, error) {
	owner = strings.TrimSpace(owner)
	if owner == "" || shot.Revision < 1 {
		return ScriptShot{}, ErrScriptShotInvalid
	}
	key := scriptShotKey(owner, shot.ProjectID, shot.ScriptNodeID, shot.ID)
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, found := s.scriptShots[key]
	if !found {
		return ScriptShot{}, ErrScriptShotNotFound
	}
	if existing.Revision != shot.Revision {
		return ScriptShot{}, ErrRevisionConflict
	}
	normalized, err := normalizeScriptShot(owner, existing.ProjectID, existing.ScriptNodeID, shot, time.Now().UTC())
	if err != nil {
		return ScriptShot{}, err
	}
	normalized.CreatedAt = existing.CreatedAt
	normalized.Revision = existing.Revision + 1
	s.scriptShots[key] = normalized
	return normalized, nil
}

func (s *MemoryStore) AttachGeneration(_ context.Context, owner string, link ShotGenerationLink) error {
	owner = strings.TrimSpace(owner)
	link, err := normalizeShotGenerationLink(owner, link)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	project, exists := s.projects[projectKey(owner, link.ProjectID)]
	if !exists || project.DeletedAt != nil {
		return ErrProjectNotFound
	}
	if _, exists := s.scriptShots[scriptShotKey(owner, link.ProjectID, link.ScriptNodeID, link.ShotID)]; !exists {
		return ErrScriptShotNotFound
	}
	s.shotGenerationLinks[shotGenerationLinkKey(owner, link)] = link
	return nil
}

func (s *MemoryStore) ListGenerationLinks(_ context.Context, owner, projectID, shotID string) ([]ShotGenerationLink, error) {
	owner, projectID, shotID = strings.TrimSpace(owner), strings.TrimSpace(projectID), strings.TrimSpace(shotID)
	if owner == "" || projectID == "" || shotID == "" {
		return nil, ErrScriptShotInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	prefix := owner + "\x00" + projectID + "\x00" + shotID + "\x00"
	links := make([]ShotGenerationLink, 0)
	for key, link := range s.shotGenerationLinks {
		if strings.HasPrefix(key, prefix) {
			links = append(links, link)
		}
	}
	sort.Slice(links, func(i, j int) bool {
		if links[i].Attempt != links[j].Attempt {
			return links[i].Attempt < links[j].Attempt
		}
		if links[i].NodeID != links[j].NodeID {
			return links[i].NodeID < links[j].NodeID
		}
		return links[i].GenerationID < links[j].GenerationID
	})
	return links, nil
}

func normalizeCreateShotGeneration(owner string, input CreateShotGenerationInput) (CreateShotGenerationInput, ShotGenerationLink, error) {
	owner = strings.TrimSpace(owner)
	input.Node.ID = strings.TrimSpace(input.Node.ID)
	input.Node.Type = strings.TrimSpace(input.Node.Type)
	input.Edge.ID = strings.TrimSpace(input.Edge.ID)
	input.Edge.Source = strings.TrimSpace(input.Edge.Source)
	input.Edge.Target = strings.TrimSpace(input.Edge.Target)
	input.Edge.Relation = strings.TrimSpace(input.Edge.Relation)
	if owner == "" || input.ExpectedCanvasRevision < 0 || input.Now.IsZero() || input.Node.ID == "" || input.Node.Type == "" || input.Node.Version < 1 || input.Edge.ID == "" || input.Edge.Source == "" || input.Edge.Target != input.Node.ID || input.Edge.Relation == "" {
		return CreateShotGenerationInput{}, ShotGenerationLink{}, ErrScriptShotInvalid
	}
	task, err := normalizeCreateGenerationTask(input.Task)
	if err != nil || task.NodeID != input.Node.ID || task.NodeVersion != input.Node.Version || task.Kind != input.Node.Type {
		return CreateShotGenerationInput{}, ShotGenerationLink{}, ErrGenerationInvalid
	}
	input.Task = task
	link, err := normalizeShotGenerationLink(owner, input.Link)
	if err != nil || link.ProjectID != task.ProjectID || link.NodeID != input.Node.ID || link.GenerationID != task.ID || link.ScriptNodeID != input.Edge.Source {
		return CreateShotGenerationInput{}, ShotGenerationLink{}, ErrScriptShotInvalid
	}
	input.Now = normalizeGenerationTime(input.Now)
	return input, link, nil
}

func generationTaskFromInput(owner string, input CreateGenerationTaskInput) GenerationTask {
	return GenerationTask{
		ID: input.ID, ProjectID: input.ProjectID, OwnerUsername: strings.TrimSpace(owner), NodeID: input.NodeID, NodeVersion: input.NodeVersion,
		Kind: input.Kind, Provider: input.Provider, ModelID: input.ModelID, Mode: input.Mode, State: GenerationQueued,
		Settings: cloneMap(input.Settings), ReferenceAssetIDs: append([]string(nil), input.ReferenceAssetIDs...), Results: []GenerationResult{},
		IdempotencyKey: input.IdempotencyKey, RetryOfTaskID: input.RetryOfTaskID, CreatedAt: input.Now, UpdatedAt: input.Now,
	}
}

func (s *MemoryStore) CreateShotGeneration(_ context.Context, owner string, input CreateShotGenerationInput) (Canvas, GenerationTask, ShotGenerationLink, error) {
	input, link, err := normalizeCreateShotGeneration(owner, input)
	if err != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, err
	}
	owner = strings.TrimSpace(owner)
	s.mu.Lock()
	defer s.mu.Unlock()
	key := projectKey(owner, input.Task.ProjectID)
	project, exists := s.projects[key]
	if !exists || project.DeletedAt != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, ErrProjectNotFound
	}
	if project.Canvas.Revision != input.ExpectedCanvasRevision {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, ErrRevisionConflict
	}
	if _, exists := s.scriptShots[scriptShotKey(owner, link.ProjectID, link.ScriptNodeID, link.ShotID)]; !exists {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, ErrScriptShotNotFound
	}
	for _, node := range project.Canvas.Nodes {
		if node.ID == input.Node.ID {
			return Canvas{}, GenerationTask{}, ShotGenerationLink{}, ErrGenerationInvalid
		}
	}
	for _, edge := range project.Canvas.Edges {
		if edge.ID == input.Edge.ID {
			return Canvas{}, GenerationTask{}, ShotGenerationLink{}, ErrGenerationInvalid
		}
	}
	idempotencyScope := generationIdempotencyScope(owner, input.Task.ProjectID, input.Task.IdempotencyKey)
	if s.generationIdempotency[idempotencyScope] != "" {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, ErrGenerationInvalid
	}
	taskKey := generationIDKey(owner, input.Task.ProjectID, input.Task.ID)
	if _, exists := s.generationTasks[taskKey]; exists {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, ErrGenerationInvalid
	}
	linkKey := shotGenerationLinkKey(owner, link)
	if _, exists := s.shotGenerationLinks[linkKey]; exists {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, ErrScriptShotInvalid
	}

	project.Canvas.Nodes = append(project.Canvas.Nodes, input.Node)
	project.Canvas.Edges = append(project.Canvas.Edges, input.Edge)
	project.Canvas.Revision++
	project.UpdatedAt = input.Now
	task := generationTaskFromInput(owner, input.Task)
	s.projects[key] = project
	s.generationTasks[taskKey] = cloneGenerationTask(task)
	s.generationIdempotency[idempotencyScope] = task.ID
	s.shotGenerationLinks[linkKey] = link
	return cloneProject(project).Canvas, cloneGenerationTask(task), link, nil
}
