package aicut

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"strings"
	"time"
)

var ErrInvalidEditPlan = errors.New("invalid ai-cut edit plan")

type EditPlanState string

const (
	EditPlanQueued   EditPlanState = "queued"
	EditPlanPlanning EditPlanState = "planning"
	EditPlanReady    EditPlanState = "ready"
	EditPlanApplied  EditPlanState = "applied"
	EditPlanFailed   EditPlanState = "failed"
	EditPlanCanceled EditPlanState = "canceled"
	EditPlanStale    EditPlanState = "stale"
)

type EditOperationType string

const (
	EditTrimClip          EditOperationType = "trim_clip"
	EditRemoveRange       EditOperationType = "remove_range"
	EditMoveClip          EditOperationType = "move_clip"
	EditAddCaption        EditOperationType = "add_caption"
	EditAddTitle          EditOperationType = "add_title"
	EditSetTransition     EditOperationType = "set_transition"
	EditSetFilter         EditOperationType = "set_filter"
	EditAddEffectLayer    EditOperationType = "add_effect_layer"
	EditAddBGM            EditOperationType = "add_bgm"
	EditSetVolumeEnvelope EditOperationType = "set_volume_envelope"
	EditSetProjectFormat  EditOperationType = "set_project_format"
)

type EditPlanRequest struct {
	AssetIDs           []string `json:"assetIds"`
	Goal               string   `json:"goal"`
	TargetDurationMS   int64    `json:"targetDurationMs,omitempty"`
	AspectRatio        string   `json:"aspectRatio"`
	IncludeCaptions    bool     `json:"includeCaptions"`
	IncludeTransitions bool     `json:"includeTransitions"`
	IncludeFilters     bool     `json:"includeFilters"`
	IncludeBGM         bool     `json:"includeBgm"`
	ExtraRequirement   string   `json:"extraRequirement,omitempty"`
}

type EditPlanSummary struct {
	SourceDurationMS    int64 `json:"sourceDurationMs"`
	EstimatedDurationMS int64 `json:"estimatedDurationMs"`
}

type EditOperation struct {
	ID         string             `json:"id"`
	Type       EditOperationType  `json:"type"`
	TrackID    string             `json:"trackId,omitempty"`
	ClipID     string             `json:"clipId,omitempty"`
	AssetID    string             `json:"assetId,omitempty"`
	ResourceID string             `json:"resourceId,omitempty"`
	StartMS    int64              `json:"startMs,omitempty"`
	EndMS      int64              `json:"endMs,omitempty"`
	DurationMS int64              `json:"durationMs,omitempty"`
	Text       string             `json:"text,omitempty"`
	Reason     string             `json:"reason"`
	Enabled    bool               `json:"enabled"`
	Volume     float64            `json:"volume,omitempty"`
	FadeInMS   int64              `json:"fadeInMs,omitempty"`
	FadeOutMS  int64              `json:"fadeOutMs,omitempty"`
	Parameters map[string]float64 `json:"parameters,omitempty"`
}

type EditPlan struct {
	ID              string          `json:"id"`
	ClientRequestID string          `json:"clientRequestId"`
	Owner           string          `json:"-"`
	ProjectID       string          `json:"projectId"`
	BaseRevision    int64           `json:"baseRevision"`
	ModelID         string          `json:"modelId"`
	State           EditPlanState   `json:"state"`
	Request         EditPlanRequest `json:"request"`
	Summary         EditPlanSummary `json:"summary"`
	Operations      []EditOperation `json:"operations"`
	Warnings        []string        `json:"warnings,omitempty"`
	Error           string          `json:"error,omitempty"`
	AppliedRevision int64           `json:"appliedRevision,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

type EditPlanEnvelope struct {
	ID              string
	ClientRequestID string
	Owner           string
	ProjectID       string
	BaseRevision    int64
	ModelID         string
	Request         EditPlanRequest
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type editPlanModelPayload struct {
	Summary    EditPlanSummary `json:"summary"`
	Operations []EditOperation `json:"operations"`
	Warnings   []string        `json:"warnings,omitempty"`
}

func validEditPlanState(state EditPlanState) bool {
	switch state {
	case EditPlanQueued, EditPlanPlanning, EditPlanReady, EditPlanApplied, EditPlanFailed, EditPlanCanceled, EditPlanStale:
		return true
	default:
		return false
	}
}

func validEditPlanID(value string, max int) bool {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > max {
		return false
	}
	for index, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || ((char == '-' || char == '_') && index > 0) {
			continue
		}
		return false
	}
	return true
}

func validEditPlanAspectRatio(value string) bool {
	return value == "9:16" || value == "16:9" || value == "1:1"
}

func validEffectResource(value string) bool {
	switch value {
	case "effect-blur", "effect-vignette", "effect-glow", "effect-grain", "effect-shake", "effect-breathe-zoom", "effect-flash", "effect-glitch", "effect-chromatic", "effect-mosaic", "effect-fisheye", "effect-sharpen":
		return true
	default:
		return false
	}
}

func normalizeEditPlanRequest(request EditPlanRequest) (EditPlanRequest, error) {
	request.Goal = strings.TrimSpace(request.Goal)
	request.AspectRatio = strings.TrimSpace(request.AspectRatio)
	request.ExtraRequirement = strings.TrimSpace(request.ExtraRequirement)
	if request.Goal == "" || len([]rune(request.Goal)) > 2000 || len([]rune(request.ExtraRequirement)) > 4000 || request.TargetDurationMS < 0 || request.TargetDurationMS > 7200000 || !validEditPlanAspectRatio(request.AspectRatio) || len(request.AssetIDs) == 0 || len(request.AssetIDs) > 64 {
		return EditPlanRequest{}, ErrInvalidEditPlan
	}
	seen := make(map[string]struct{}, len(request.AssetIDs))
	for index, id := range request.AssetIDs {
		id = strings.TrimSpace(id)
		if !validEditPlanID(id, 128) {
			return EditPlanRequest{}, ErrInvalidEditPlan
		}
		if _, exists := seen[id]; exists {
			return EditPlanRequest{}, ErrInvalidEditPlan
		}
		seen[id] = struct{}{}
		request.AssetIDs[index] = id
	}
	return request, nil
}

func normalizeEditOperation(operation EditOperation) (EditOperation, error) {
	operation.ID = strings.TrimSpace(operation.ID)
	operation.TrackID = strings.TrimSpace(operation.TrackID)
	operation.ClipID = strings.TrimSpace(operation.ClipID)
	operation.AssetID = strings.TrimSpace(operation.AssetID)
	operation.ResourceID = strings.TrimSpace(operation.ResourceID)
	operation.Text = strings.TrimSpace(operation.Text)
	operation.Reason = strings.TrimSpace(operation.Reason)
	if !validEditPlanID(operation.ID, 96) || operation.Reason == "" || len([]rune(operation.Reason)) > 500 || len([]rune(operation.Text)) > 5000 || operation.StartMS < 0 || operation.EndMS < 0 || operation.DurationMS < 0 || operation.FadeInMS < 0 || operation.FadeOutMS < 0 || math.IsNaN(operation.Volume) || math.IsInf(operation.Volume, 0) {
		return EditOperation{}, ErrInvalidEditPlan
	}
	for key, value := range operation.Parameters {
		if strings.TrimSpace(key) == "" || len(key) > 64 || math.IsNaN(value) || math.IsInf(value, 0) {
			return EditOperation{}, ErrInvalidEditPlan
		}
	}
	requireTrackClip := func() bool { return validEditPlanID(operation.TrackID, 128) && validEditPlanID(operation.ClipID, 128) }
	switch operation.Type {
	case EditTrimClip:
		if !requireTrackClip() || operation.EndMS <= operation.StartMS {
			return EditOperation{}, ErrInvalidEditPlan
		}
	case EditRemoveRange:
		if operation.EndMS <= operation.StartMS {
			return EditOperation{}, ErrInvalidEditPlan
		}
	case EditMoveClip:
		if !requireTrackClip() {
			return EditOperation{}, ErrInvalidEditPlan
		}
	case EditAddCaption:
		if !validEditPlanID(operation.TrackID, 128) || operation.DurationMS <= 0 || operation.Text == "" {
			return EditOperation{}, ErrInvalidEditPlan
		}
	case EditAddTitle:
		if !validEditPlanID(operation.TrackID, 128) || operation.DurationMS <= 0 || operation.Text == "" || !validTextResource(operation.ResourceID) {
			return EditOperation{}, ErrInvalidEditPlan
		}
	case EditSetTransition:
		if !requireTrackClip() || operation.DurationMS < 100 || operation.DurationMS > 3000 || !validTransition(operation.ResourceID) {
			return EditOperation{}, ErrInvalidEditPlan
		}
	case EditSetFilter:
		if !requireTrackClip() || !validCatalogVisualResource(operation.ResourceID) {
			return EditOperation{}, ErrInvalidEditPlan
		}
	case EditAddEffectLayer:
		if !validEditPlanID(operation.TrackID, 128) || operation.DurationMS <= 0 || !validEffectResource(operation.ResourceID) {
			return EditOperation{}, ErrInvalidEditPlan
		}
	case EditAddBGM:
		if !validEditPlanID(operation.TrackID, 128) || !validEditPlanID(operation.AssetID, 128) || operation.DurationMS <= 0 || operation.Volume < 0 || operation.Volume > 4 {
			return EditOperation{}, ErrInvalidEditPlan
		}
	case EditSetVolumeEnvelope:
		if !requireTrackClip() || operation.Volume < 0 || operation.Volume > 4 {
			return EditOperation{}, ErrInvalidEditPlan
		}
	case EditSetProjectFormat:
		if !validEditPlanAspectRatio(operation.ResourceID) {
			return EditOperation{}, ErrInvalidEditPlan
		}
	default:
		return EditOperation{}, ErrInvalidEditPlan
	}
	return operation, nil
}

func NormalizeEditPlan(plan EditPlan) (EditPlan, error) {
	plan.ID = strings.TrimSpace(plan.ID)
	plan.ClientRequestID = strings.TrimSpace(plan.ClientRequestID)
	plan.Owner = strings.TrimSpace(plan.Owner)
	plan.ProjectID = strings.TrimSpace(plan.ProjectID)
	plan.ModelID = strings.TrimSpace(plan.ModelID)
	plan.Error = strings.TrimSpace(plan.Error)
	if !validEditPlanID(plan.ID, 128) || !validEditPlanID(plan.ClientRequestID, 128) || plan.Owner == "" || !validEditPlanID(plan.ProjectID, 128) || !validEditPlanID(plan.ModelID, 128) || plan.BaseRevision < 0 || plan.AppliedRevision < 0 || !validEditPlanState(plan.State) || plan.Summary.SourceDurationMS < 0 || plan.Summary.EstimatedDurationMS < 0 || len(plan.Operations) > 200 {
		return EditPlan{}, ErrInvalidEditPlan
	}
	request, err := normalizeEditPlanRequest(plan.Request)
	if err != nil {
		return EditPlan{}, err
	}
	plan.Request = request
	seen := make(map[string]struct{}, len(plan.Operations))
	for index, operation := range plan.Operations {
		normalized, operationErr := normalizeEditOperation(operation)
		if operationErr != nil {
			return EditPlan{}, operationErr
		}
		if _, exists := seen[normalized.ID]; exists {
			return EditPlan{}, ErrInvalidEditPlan
		}
		seen[normalized.ID] = struct{}{}
		plan.Operations[index] = normalized
	}
	for index, warning := range plan.Warnings {
		warning = strings.TrimSpace(warning)
		if warning == "" || len([]rune(warning)) > 500 {
			return EditPlan{}, ErrInvalidEditPlan
		}
		plan.Warnings[index] = warning
	}
	return plan, nil
}

func stripEditPlanFence(raw string) string {
	value := strings.TrimSpace(raw)
	if !strings.HasPrefix(value, "```") {
		return value
	}
	firstNewline := strings.IndexByte(value, '\n')
	if firstNewline < 0 || !strings.HasSuffix(value, "```") {
		return value
	}
	return strings.TrimSpace(value[firstNewline+1 : len(value)-3])
}

func ParseEditPlanModelOutput(raw string, envelope EditPlanEnvelope) (EditPlan, error) {
	decoder := json.NewDecoder(bytes.NewBufferString(stripEditPlanFence(raw)))
	decoder.DisallowUnknownFields()
	var payload editPlanModelPayload
	if err := decoder.Decode(&payload); err != nil {
		return EditPlan{}, ErrInvalidEditPlan
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return EditPlan{}, ErrInvalidEditPlan
	}
	plan := EditPlan{
		ID: envelope.ID, ClientRequestID: envelope.ClientRequestID, Owner: envelope.Owner, ProjectID: envelope.ProjectID,
		BaseRevision: envelope.BaseRevision, ModelID: envelope.ModelID, State: EditPlanReady, Request: envelope.Request,
		Summary: payload.Summary, Operations: payload.Operations, Warnings: payload.Warnings,
		CreatedAt: envelope.CreatedAt, UpdatedAt: envelope.UpdatedAt,
	}
	return NormalizeEditPlan(plan)
}
