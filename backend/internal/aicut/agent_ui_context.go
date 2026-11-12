package aicut

import (
	"context"
	"encoding/json"
	"regexp"
)

const AgentEventUIContext AgentEventType = "ui_context"

type AgentSelectedClip struct {
	TrackID string `json:"trackId"`
	ClipID  string `json:"clipId"`
}

// AgentUIContext is a bounded snapshot of the editor at submission time.
// It contains data only; it never changes the user's saved instruction.
type AgentUIContext struct {
	SelectedClips []AgentSelectedClip `json:"selectedClips"`
	PlayheadMS    int64               `json:"playheadMs"`
}

var agentContextID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$`)

func ValidateAgentUIContext(value AgentUIContext, timeline Timeline) error {
	if len(value.SelectedClips) > 100 || value.PlayheadMS < 0 || value.PlayheadMS > 604800000 {
		return ErrInvalidAgentRun
	}
	seen := map[AgentSelectedClip]bool{}
	for _, selected := range value.SelectedClips {
		if !agentContextID.MatchString(selected.TrackID) || !agentContextID.MatchString(selected.ClipID) || seen[selected] {
			return ErrInvalidAgentRun
		}
		seen[selected] = true
		found := false
		for _, track := range timeline.Tracks {
			if track.ID == selected.TrackID {
				for _, clip := range track.Clips {
					if clip.ID == selected.ClipID {
						found = true
						break
					}
				}
				break
			}
		}
		if !found {
			return ErrInvalidAgentRun
		}
	}
	return nil
}

func loadAgentUIContext(ctx context.Context, store AgentStore, run AgentRun) (*AgentUIContext, error) {
	persisted, err := store.AgentRun(ctx, run.Owner, run.ProjectID, run.ID)
	if err != nil {
		return nil, err
	}
	if canonicalASRTimeline(persisted.TimelineID) != canonicalASRTimeline(run.TimelineID) {
		return nil, ErrAgentRunNotFound
	}
	events, err := store.AgentEvents(ctx, run.Owner, run.ProjectID, run.ID, 0)
	if err != nil {
		return nil, err
	}
	var result *AgentUIContext
	for _, event := range events {
		if event.Type == AgentEventUIContext {
			if result != nil || event.Revision != persisted.ExpectedRevision || len(event.Message) > 40000 {
				return nil, ErrInvalidAgentRun
			}
			var value AgentUIContext
			if err := decodeToolArguments(json.RawMessage(event.Message), &value); err != nil {
				return nil, ErrInvalidAgentRun
			}
			result = &value
		}
	}
	return result, nil
}

func (e *AgentToolExecutor) initialUIContext(ctx context.Context, run AgentRun) (*AgentUIContext, error) {
	value, err := loadAgentUIContext(ctx, e.agents, run)
	if err != nil || value == nil {
		return value, err
	}
	store := e.timelines
	if canonicalASRTimeline(run.TimelineID) != DefaultTimelineID {
		if e.collection == nil {
			return nil, ErrTimelineJobUnsupported
		}
		store, err = e.collection.ScopedStore(ctx, run.Owner, run.ProjectID, run.TimelineID)
		if err != nil {
			return nil, err
		}
	}
	timeline, err := store.Timeline(ctx, run.Owner, run.ProjectID)
	if err != nil {
		return nil, err
	}
	if timeline.Revision != run.ExpectedRevision {
		return nil, ErrRevisionConflict
	}
	if err := ValidateAgentUIContext(*value, timeline); err != nil {
		return nil, err
	}
	return value, nil
}
