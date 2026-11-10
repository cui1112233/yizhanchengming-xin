package aicut

import (
	"errors"
	"math"
	"strings"
	"time"
)

var (
	ErrProjectNotFound         = errors.New("ai-cut project not found")
	ErrProjectExists           = errors.New("ai-cut project already exists")
	ErrRevisionConflict        = errors.New("ai-cut timeline revision conflict")
	ErrTimelineHistoryBoundary = errors.New("ai-cut timeline history boundary")
	ErrInvalidProject          = errors.New("invalid ai-cut project")
	ErrInvalidTimeline         = errors.New("invalid ai-cut timeline")
	ErrAssetNotFound           = errors.New("ai-cut asset not found")
	ErrAssetExists             = errors.New("ai-cut asset already exists")
	ErrInvalidAsset            = errors.New("invalid ai-cut asset")
	ErrASRJobNotFound          = errors.New("ai-cut asr job not found")
	ErrInvalidASRJob           = errors.New("invalid ai-cut asr job")
	ErrASRQueueEmpty           = errors.New("ai-cut asr queue empty")
)

type ASRJobState string

const (
	ASRQueued          ASRJobState = "queued"
	ASRRunning         ASRJobState = "running"
	ASRCancelRequested ASRJobState = "cancel_requested"
	ASRSucceeded       ASRJobState = "succeeded"
	ASRFailed          ASRJobState = "failed"
	ASRCanceled        ASRJobState = "canceled"
)

type ASRJob struct {
	ID               string      `json:"id"`
	Owner            string      `json:"-"`
	ProjectID        string      `json:"projectId"`
	AssetID          string      `json:"assetId"`
	ModelID          string      `json:"modelId"`
	Language         string      `json:"language,omitempty"`
	ExpectedRevision int64       `json:"expectedRevision"`
	State            ASRJobState `json:"state"`
	Progress         int         `json:"progress"`
	Error            string      `json:"error,omitempty"`
	Imported         int         `json:"imported,omitempty"`
	TrackID          string      `json:"trackId,omitempty"`
	CreatedAt        time.Time   `json:"createdAt"`
	UpdatedAt        time.Time   `json:"updatedAt"`
}

func (job ASRJob) WithState(state ASRJobState, progress int, message string, now time.Time) ASRJob {
	job.State, job.Progress, job.Error, job.UpdatedAt = state, progress, strings.TrimSpace(message), now.UTC()
	return job
}

func validateASRJob(job ASRJob) error {
	job.ID, job.ProjectID, job.AssetID, job.ModelID = strings.TrimSpace(job.ID), strings.TrimSpace(job.ProjectID), strings.TrimSpace(job.AssetID), strings.TrimSpace(job.ModelID)
	if job.ID == "" || job.ProjectID == "" || job.AssetID == "" || job.ModelID == "" || job.ExpectedRevision < 0 || job.Progress < 0 || job.Progress > 100 {
		return ErrInvalidASRJob
	}
	switch job.State {
	case ASRQueued, ASRRunning, ASRCancelRequested, ASRSucceeded, ASRFailed, ASRCanceled:
		return nil
	default:
		return ErrInvalidASRJob
	}
}

type TrackKind string

const (
	TrackVideo    TrackKind = "video"
	TrackAudio    TrackKind = "audio"
	TrackSubtitle TrackKind = "subtitle"
	TrackEffect   TrackKind = "effect"
)

type ClipKind string

const (
	ClipVideo    ClipKind = "video"
	ClipAudio    ClipKind = "audio"
	ClipImage    ClipKind = "image"
	ClipSubtitle ClipKind = "subtitle"
	ClipEffect   ClipKind = "effect"
)

type Project struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	AspectRatio string     `json:"aspectRatio"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"-"`
}

type Timeline struct {
	Revision int64   `json:"revision"`
	Tracks   []Track `json:"tracks"`
	Script   string  `json:"script,omitempty"`
}

type TimelineHistoryState struct {
	CanUndo bool `json:"canUndo"`
	CanRedo bool `json:"canRedo"`
}

type Track struct {
	ID       string        `json:"id"`
	Kind     TrackKind     `json:"kind"`
	Name     string        `json:"name"`
	Order    int           `json:"order"`
	Muted    bool          `json:"muted"`
	Locked   bool          `json:"locked"`
	Settings TrackSettings `json:"settings,omitempty"`
	Clips    []Clip        `json:"clips"`
}

type TrackSettings struct {
	SubtitleStyleID   string `json:"subtitleStyleId,omitempty"`
	SubtitleRhythm    string `json:"subtitleRhythm,omitempty"`
	SubtitleAnimation string `json:"subtitleAnimation,omitempty"`
	PreviewVisible    bool   `json:"previewVisible,omitempty"`
}

type Clip struct {
	ID               string             `json:"id"`
	AssetID          string             `json:"assetId,omitempty"`
	Kind             ClipKind           `json:"kind"`
	Name             string             `json:"name"`
	StartMS          int64              `json:"startMs"`
	DurationMS       int64              `json:"durationMs"`
	SourceStartMS    int64              `json:"sourceStartMs"`
	SourceDurationMS int64              `json:"sourceDurationMs"`
	Text             string             `json:"text,omitempty"`
	Transition       string             `json:"transition,omitempty"`
	TransitionMS     int64              `json:"transitionMs,omitempty"`
	PlaybackRate     float64            `json:"playbackRate,omitempty"`
	Volume           float64            `json:"volume,omitempty"`
	FadeInMS         int64              `json:"fadeInMs,omitempty"`
	FadeOutMS        int64              `json:"fadeOutMs,omitempty"`
	ReviewStatus     string             `json:"reviewStatus,omitempty"`
	ReviewReason     string             `json:"reviewReason,omitempty"`
	VisualEffect     *VisualEffect      `json:"visualEffect,omitempty"`
	ResourceID       string             `json:"resourceId,omitempty"`
	Parameters       map[string]float64 `json:"parameters,omitempty"`
	CanvasTransform  *CanvasTransform   `json:"canvasTransform,omitempty"`
}

type CanvasTransform struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	ScaleX   float64 `json:"scaleX"`
	ScaleY   float64 `json:"scaleY"`
	Rotation float64 `json:"rotation"`
	Opacity  float64 `json:"opacity"`
}

type VisualEffect struct {
	Preset     string  `json:"preset"`
	Brightness float64 `json:"brightness"`
	Contrast   float64 `json:"contrast"`
	Saturation float64 `json:"saturation"`
	Blur       float64 `json:"blur"`
	Grayscale  bool    `json:"grayscale"`
	Intensity  float64 `json:"intensity,omitempty"`
	Speed      float64 `json:"speed,omitempty"`
}

func validVisualEffect(effect *VisualEffect) bool {
	if effect == nil {
		return true
	}
	switch effect.Preset {
	case "original", "bright", "cool-cinema", "warm", "grayscale", "soft-focus", "vignette", "glow", "shake", "breathe", "flash", "mosaic", "custom":
	default:
		return false
	}
	values := []float64{effect.Brightness, effect.Contrast, effect.Saturation, effect.Blur, effect.Intensity, effect.Speed}
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return effect.Brightness >= -1 && effect.Brightness <= 1 && effect.Contrast >= .5 && effect.Contrast <= 2 && effect.Saturation >= 0 && effect.Saturation <= 3 && effect.Blur >= 0 && effect.Blur <= 20 && effect.Intensity >= 0 && effect.Intensity <= 1 && effect.Speed >= 0 && effect.Speed <= 5 && (effect.Speed == 0 || effect.Speed >= .1)
}

func validTransition(value string) bool {
	switch value {
	case "", "cross-dissolve", "fade", "flash-black", "flash-white", "slide-left", "slide-right", "slide-up", "slide-down", "zoom", "blur",
		"transition-cross-dissolve", "transition-fade", "transition-flash-black", "transition-flash-white", "transition-push-left", "transition-push-right", "transition-push-up", "transition-push-down", "transition-wipe-left", "transition-wipe-right", "transition-wipe-up", "transition-wipe-down", "transition-zoom", "transition-blur", "transition-glitch", "transition-light-leak":
		return true
	default:
		return false
	}
}

type Asset struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"projectId"`
	Kind        string    `json:"kind"`
	Name        string    `json:"name"`
	ContentType string    `json:"contentType"`
	ObjectKey   string    `json:"-"`
	URL         string    `json:"url,omitempty"`
	ByteSize    int64     `json:"byteSize"`
	DurationMS  int64     `json:"durationMs"`
	CreatedAt   time.Time `json:"createdAt"`
}

func normalizeProject(project Project) (Project, error) {
	project.ID = strings.TrimSpace(project.ID)
	project.Title = strings.TrimSpace(project.Title)
	project.AspectRatio = strings.TrimSpace(project.AspectRatio)
	if project.ID == "" {
		return Project{}, ErrInvalidProject
	}
	if project.Title == "" {
		project.Title = "未命名剪辑"
	}
	if project.AspectRatio == "" {
		project.AspectRatio = "9:16"
	}
	if project.AspectRatio != "9:16" && project.AspectRatio != "16:9" && project.AspectRatio != "1:1" {
		return Project{}, ErrInvalidProject
	}
	return project, nil
}

func validateTimeline(timeline Timeline) error {
	if len([]rune(timeline.Script)) > 200000 {
		return ErrInvalidTimeline
	}
	trackIDs := make(map[string]struct{}, len(timeline.Tracks))
	clipIDs := make(map[string]struct{})
	for _, track := range timeline.Tracks {
		if strings.TrimSpace(track.ID) == "" || !validTrackKind(track.Kind) {
			return ErrInvalidTimeline
		}
		if _, exists := trackIDs[track.ID]; exists {
			return ErrInvalidTimeline
		}
		trackIDs[track.ID] = struct{}{}
		if len([]rune(track.Settings.SubtitleStyleID)) > 64 || len([]rune(track.Settings.SubtitleAnimation)) > 32 ||
			(track.Settings.SubtitleRhythm != "" && track.Settings.SubtitleRhythm != "sentence" && track.Settings.SubtitleRhythm != "word") {
			return ErrInvalidTimeline
		}
		for _, clip := range track.Clips {
			if strings.TrimSpace(clip.ID) == "" || !validClipKind(clip.Kind) || clip.StartMS < 0 || clip.DurationMS <= 0 || clip.SourceStartMS < 0 || clip.SourceDurationMS < 0 {
				return ErrInvalidTimeline
			}
			if clip.SourceDurationMS > 0 && clip.SourceStartMS+clip.DurationMS > clip.SourceDurationMS {
				return ErrInvalidTimeline
			}
			if !validTransition(clip.Transition) {
				return ErrInvalidTimeline
			}
			if clip.TransitionMS < 0 || clip.TransitionMS > 5000 || (clip.Transition != "" && clip.TransitionMS == 0) {
				return ErrInvalidTimeline
			}
			if clip.PlaybackRate != 0 && (clip.PlaybackRate < 0.25 || clip.PlaybackRate > 4) {
				return ErrInvalidTimeline
			}
			if clip.Volume < 0 || clip.Volume > 2 || clip.FadeInMS < 0 || clip.FadeOutMS < 0 || clip.FadeInMS > clip.DurationMS || clip.FadeOutMS > clip.DurationMS {
				return ErrInvalidTimeline
			}
			if clip.ReviewStatus != "" && clip.ReviewStatus != "delete" && clip.ReviewStatus != "review" && clip.ReviewStatus != "keep" {
				return ErrInvalidTimeline
			}
			if len([]rune(clip.ReviewReason)) > 1000 {
				return ErrInvalidTimeline
			}
			if !validVisualEffect(clip.VisualEffect) {
				return ErrInvalidTimeline
			}
			if clip.Kind == ClipEffect {
				if track.Kind != TrackEffect || !strings.HasPrefix(clip.ResourceID, "effect-") {
					return ErrInvalidTimeline
				}
				for _, value := range clip.Parameters {
					if math.IsNaN(value) || math.IsInf(value, 0) || value < -100 || value > 100 {
						return ErrInvalidTimeline
					}
				}
			}
			if clip.CanvasTransform != nil {
				transform := clip.CanvasTransform
				values := []float64{transform.X, transform.Y, transform.ScaleX, transform.ScaleY, transform.Rotation, transform.Opacity}
				for _, value := range values {
					if math.IsNaN(value) || math.IsInf(value, 0) {
						return ErrInvalidTimeline
					}
				}
				if transform.X < -1 || transform.X > 1 || transform.Y < -1 || transform.Y > 1 || transform.ScaleX < .05 || transform.ScaleX > 20 || transform.ScaleY < .05 || transform.ScaleY > 20 || transform.Rotation < -180 || transform.Rotation > 180 || transform.Opacity < 0 || transform.Opacity > 1 {
					return ErrInvalidTimeline
				}
			}
			if _, exists := clipIDs[clip.ID]; exists {
				return ErrInvalidTimeline
			}
			clipIDs[clip.ID] = struct{}{}
		}
	}
	return nil
}

func validTrackKind(kind TrackKind) bool {
	return kind == TrackVideo || kind == TrackAudio || kind == TrackSubtitle || kind == TrackEffect
}

func validClipKind(kind ClipKind) bool {
	return kind == ClipVideo || kind == ClipAudio || kind == ClipImage || kind == ClipSubtitle || kind == ClipEffect
}

func validateAsset(asset Asset) error {
	asset.ID = strings.TrimSpace(asset.ID)
	asset.ProjectID = strings.TrimSpace(asset.ProjectID)
	asset.Name = strings.TrimSpace(asset.Name)
	asset.ContentType = strings.ToLower(strings.TrimSpace(asset.ContentType))
	asset.ObjectKey = strings.TrimSpace(asset.ObjectKey)
	if asset.ID == "" || asset.ProjectID == "" || asset.Name == "" || asset.ObjectKey == "" || asset.ByteSize < 0 || asset.DurationMS < 0 {
		return ErrInvalidAsset
	}
	if (asset.Kind == "image" && !strings.HasPrefix(asset.ContentType, "image/")) ||
		(asset.Kind == "video" && !strings.HasPrefix(asset.ContentType, "video/")) ||
		(asset.Kind == "audio" && !strings.HasPrefix(asset.ContentType, "audio/")) ||
		(asset.Kind == "subtitle" && asset.ContentType != "application/x-subrip" && asset.ContentType != "text/plain") {
		return ErrInvalidAsset
	}
	if asset.Kind != "image" && asset.Kind != "video" && asset.Kind != "audio" && asset.Kind != "subtitle" {
		return ErrInvalidAsset
	}
	return nil
}
