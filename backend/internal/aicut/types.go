package aicut

import (
	"errors"
	"math"
	"sort"
	"strconv"
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
	TimelineID       string      `json:"timelineId,omitempty"`
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
	if job.ID == "" || job.ProjectID == "" || job.AssetID == "" || job.ModelID == "" || len(job.TimelineID) > 64 || job.ExpectedRevision < 0 || job.Progress < 0 || job.Progress > 100 {
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
	TrackVideo      TrackKind = "video"
	TrackAudio      TrackKind = "audio"
	TrackSubtitle   TrackKind = "subtitle"
	TrackEffect     TrackKind = "effect"
	TrackTransition TrackKind = "transition"
)

type ClipKind string

const (
	ClipVideo      ClipKind = "video"
	ClipAudio      ClipKind = "audio"
	ClipImage      ClipKind = "image"
	ClipSubtitle   ClipKind = "subtitle"
	ClipEffect     ClipKind = "effect"
	ClipTransition ClipKind = "transition"
)

type Project struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	AspectRatio     string     `json:"aspectRatio"`
	BackgroundMode  string     `json:"backgroundMode"`
	BackgroundColor string     `json:"backgroundColor"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"-"`
}

type Timeline struct {
	Revision int64            `json:"revision"`
	Tracks   []Track          `json:"tracks"`
	Script   string           `json:"script,omitempty"`
	Markers  []TimelineMarker `json:"markers,omitempty"`
}

type TimelineMarker struct {
	ID      string `json:"id"`
	TimeMS  int64  `json:"timeMs"`
	Label   string `json:"label"`
	Color   string `json:"color,omitempty"`
	Source  string `json:"source,omitempty"`
	AssetID string `json:"assetId,omitempty"`
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

type TextStyleRun struct {
	Start     int     `json:"start"`
	End       int     `json:"end"`
	TextColor string  `json:"textColor,omitempty"`
	FontSize  float64 `json:"fontSize,omitempty"`
	Bold      bool    `json:"bold,omitempty"`
	Italic    bool    `json:"italic,omitempty"`
}

type TrackSettings struct {
	SubtitleStyleID   string         `json:"subtitleStyleId,omitempty"`
	SubtitleRhythm    string         `json:"subtitleRhythm,omitempty"`
	SubtitleAnimation string         `json:"subtitleAnimation,omitempty"`
	TextResourceID    string         `json:"textResourceId,omitempty"`
	PreviewVisible    bool           `json:"previewVisible,omitempty"`
	CompoundID        string         `json:"compoundId,omitempty"`
	CompoundName      string         `json:"compoundName,omitempty"`
	FontFamily        string         `json:"fontFamily,omitempty"`
	FontAssetID       string         `json:"fontAssetId,omitempty"`
	FontSize          float64        `json:"fontSize,omitempty"`
	TextCurve         float64        `json:"textCurve,omitempty"`
	TextColor         string         `json:"textColor,omitempty"`
	StrokeColor       string         `json:"strokeColor,omitempty"`
	StrokeWidth       float64        `json:"strokeWidth,omitempty"`
	BackgroundColor   string         `json:"backgroundColor,omitempty"`
	ShadowColor       string         `json:"shadowColor,omitempty"`
	ShadowOffsetX     float64        `json:"shadowOffsetX,omitempty"`
	ShadowOffsetY     float64        `json:"shadowOffsetY,omitempty"`
	ShadowBlur        float64        `json:"shadowBlur,omitempty"`
	LetterSpacing     float64        `json:"letterSpacing,omitempty"`
	LineHeight        float64        `json:"lineHeight,omitempty"`
	TextAlign         string         `json:"textAlign,omitempty"`
	TextRuns          []TextStyleRun `json:"textRuns,omitempty"`
}

type CompoundRef struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type Clip struct {
	ID                    string             `json:"id"`
	AssetID               string             `json:"assetId,omitempty"`
	Kind                  ClipKind           `json:"kind"`
	Name                  string             `json:"name"`
	StartMS               int64              `json:"startMs"`
	DurationMS            int64              `json:"durationMs"`
	SourceStartMS         int64              `json:"sourceStartMs"`
	SourceDurationMS      int64              `json:"sourceDurationMs"`
	Text                  string             `json:"text,omitempty"`
	Transition            string             `json:"transition,omitempty"`
	TransitionMS          int64              `json:"transitionMs,omitempty"`
	TransitionTrackID     string             `json:"transitionTrackId,omitempty"`
	TransitionLeftClipID  string             `json:"transitionLeftClipId,omitempty"`
	TransitionRightClipID string             `json:"transitionRightClipId,omitempty"`
	PlaybackRate          float64            `json:"playbackRate,omitempty"`
	Reversed              bool               `json:"reversed,omitempty"`
	FreezeSourceMS        *int64             `json:"freezeSourceMs,omitempty"`
	BlendMode             string             `json:"blendMode,omitempty"`
	Volume                float64            `json:"volume,omitempty"`
	FadeInMS              int64              `json:"fadeInMs,omitempty"`
	FadeOutMS             int64              `json:"fadeOutMs,omitempty"`
	ReviewStatus          string             `json:"reviewStatus,omitempty"`
	ReviewReason          string             `json:"reviewReason,omitempty"`
	LinkGroupID           string             `json:"linkGroupId,omitempty"`
	CompoundID            string             `json:"compoundId,omitempty"`
	CompoundName          string             `json:"compoundName,omitempty"`
	CompoundPath          []CompoundRef      `json:"compoundPath,omitempty"`
	VisualEffect          *VisualEffect      `json:"visualEffect,omitempty"`
	CatalogResourceID     string             `json:"catalogResourceId,omitempty"`
	TextResourceID        string             `json:"textResourceId,omitempty"`
	ResourceID            string             `json:"resourceId,omitempty"`
	Parameters            map[string]float64 `json:"parameters,omitempty"`
	CanvasTransform       *CanvasTransform   `json:"canvasTransform,omitempty"`
	Keyframes             []ClipKeyframe     `json:"keyframes,omitempty"`
	SpeedCurve            []SpeedCurvePoint  `json:"speedCurve,omitempty"`
	AudioProcessing       *AudioProcessing   `json:"audioProcessing,omitempty"`
	VisualProcessing      *VisualProcessing  `json:"visualProcessing,omitempty"`
}

type AudioProcessing struct {
	Pan             float64 `json:"pan"`
	Denoise         float64 `json:"denoise"`
	EQLowDB         float64 `json:"eqLowDB"`
	EQMidDB         float64 `json:"eqMidDB"`
	EQHighDB        float64 `json:"eqHighDB"`
	ChannelMode     string  `json:"channelMode"`
	LoudnessEnabled bool    `json:"loudnessEnabled"`
	TargetLUFS      float64 `json:"targetLUFS"`
	VoiceEffect     string  `json:"voiceEffect"`
	PitchSemitones  float64 `json:"pitchSemitones"`
}

func validAudioProcessing(processing *AudioProcessing) bool {
	if processing == nil {
		return true
	}
	values := []float64{processing.Denoise, processing.EQLowDB, processing.EQMidDB, processing.EQHighDB, processing.TargetLUFS, processing.PitchSemitones, processing.Pan}
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	if processing.Denoise < 0 || processing.Denoise > 1 || processing.EQLowDB < -12 || processing.EQLowDB > 12 || processing.EQMidDB < -12 || processing.EQMidDB > 12 || processing.EQHighDB < -12 || processing.EQHighDB > 12 {
		return false
	}
	if (processing.LoudnessEnabled || processing.TargetLUFS != 0) && (processing.TargetLUFS < -24 || processing.TargetLUFS > -8) {
		return false
	}
	if processing.PitchSemitones < -12 || processing.PitchSemitones > 12 {
		return false
	}
	if processing.Pan < -1 || processing.Pan > 1 {
		return false
	}
	switch processing.ChannelMode {
	case "", "stereo", "mono", "left", "right", "swap":
		// valid
	default:
		return false
	}
	switch processing.VoiceEffect {
	case "", "natural", "enhance", "bright", "bass", "treble", "telephone", "radio", "megaphone", "robot", "chipmunk", "deep", "sweet", "whisper", "distorted", "lofi", "echo", "room", "hall", "plate", "cathedral", "stadium", "underwater", "vinyl":
		return true
	default:
		return false
	}
}

type VisualProcessing struct {
	FrameInterpolation string             `json:"frameInterpolation,omitempty"`
	Stabilization      VideoStabilization `json:"stabilization,omitempty"`
	Color              ColorAdjustment    `json:"color"`
	ChromaKey          ChromaKey          `json:"chromaKey"`
	PersonSegmentation PersonSegmentation `json:"personSegmentation,omitempty"`
	ObjectRemoval      ObjectRemoval      `json:"objectRemoval,omitempty"`
	Mask               VisualMask         `json:"mask"`
	Geometry           VisualGeometry     `json:"geometry"`
	Details            VisualDetails      `json:"details"`
}

type ObjectRemoval struct {
	Enabled bool    `json:"enabled"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
}

type PersonSegmentation struct {
	Enabled      bool    `json:"enabled"`
	MatteAssetID string  `json:"matteAssetId,omitempty"`
	Threshold    float64 `json:"threshold"`
	Feather      float64 `json:"feather"`
	Signature    string  `json:"signature,omitempty"`
}

func PersonSegmentationSignature(clip Clip) string {
	rate := clip.PlaybackRate
	if rate <= 0 || math.IsNaN(rate) || math.IsInf(rate, 0) {
		rate = 1
	}
	freeze := "none"
	if clip.FreezeSourceMS != nil {
		freeze = strconv.FormatInt(*clip.FreezeSourceMS, 10)
	}
	reversed := "0"
	if clip.Reversed {
		reversed = "1"
	}
	duration := clip.DurationMS
	if duration < 1 {
		duration = 1
	}
	return strings.Join([]string{
		strings.TrimSpace(clip.AssetID),
		strconv.FormatInt(max(0, clip.SourceStartMS), 10),
		strconv.FormatInt(duration, 10),
		strconv.FormatFloat(rate, 'f', -1, 64),
		reversed,
		freeze,
	}, "|")
}

type VideoStabilization struct {
	Enabled    bool    `json:"enabled"`
	Smoothness int     `json:"smoothness"`
	Zoom       float64 `json:"zoom"`
	Tripod     bool    `json:"tripod"`
}

type ColorAdjustment struct {
	RGBCurves   map[string][]ColorCurvePoint    `json:"rgbCurves,omitempty"`
	Wheels      map[string]ColorWheelAdjustment `json:"wheels,omitempty"`
	HSL         map[string]HSLBandAdjustment    `json:"hsl,omitempty"`
	Exposure    float64                         `json:"exposure"`
	Temperature float64                         `json:"temperature"`
	Tint        float64                         `json:"tint"`
	Highlights  float64                         `json:"highlights"`
	Shadows     float64                         `json:"shadows"`
	Gamma       float64                         `json:"gamma"`
	Curve       []float64                       `json:"curve,omitempty"`
	CurvePoints []ColorCurvePoint               `json:"curvePoints,omitempty"`
}

type ColorCurvePoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type ChromaKey struct {
	Enabled     bool    `json:"enabled"`
	RemoveBlack bool    `json:"removeBlack"`
	Color       string  `json:"color"`
	Similarity  float64 `json:"similarity"`
	Blend       float64 `json:"blend"`
}

type VisualMask struct {
	Type         string              `json:"type"`
	X            float64             `json:"x"`
	Y            float64             `json:"y"`
	Width        float64             `json:"width"`
	Height       float64             `json:"height"`
	Feather      float64             `json:"feather"`
	Invert       bool                `json:"invert"`
	Points       []VisualMaskPoint   `json:"points,omitempty"`
	BezierPoints []VisualBezierPoint `json:"bezierPoints,omitempty"`
}

type VisualMaskPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type VisualBezierPoint struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	InX  float64 `json:"inX"`
	InY  float64 `json:"inY"`
	OutX float64 `json:"outX"`
	OutY float64 `json:"outY"`
}

func SampleVisualBezierMask(points []VisualBezierPoint) []VisualMaskPoint {
	if len(points) < 3 || len(points) > 8 {
		return nil
	}
	const samplesPerSegment = 6
	result := make([]VisualMaskPoint, 0, len(points)*samplesPerSegment)
	for index, point := range points {
		next := points[(index+1)%len(points)]
		for step := 0; step < samplesPerSegment; step++ {
			ratio := float64(step) / samplesPerSegment
			inverse := 1 - ratio
			result = append(result, VisualMaskPoint{
				X: inverse*inverse*inverse*point.X + 3*inverse*inverse*ratio*point.OutX + 3*inverse*ratio*ratio*next.InX + ratio*ratio*ratio*next.X,
				Y: inverse*inverse*inverse*point.Y + 3*inverse*inverse*ratio*point.OutY + 3*inverse*ratio*ratio*next.InY + ratio*ratio*ratio*next.Y,
			})
		}
	}
	return result
}

type VisualGeometry struct {
	CropLeft       float64 `json:"cropLeft"`
	CropTop        float64 `json:"cropTop"`
	CropRight      float64 `json:"cropRight"`
	CropBottom     float64 `json:"cropBottom"`
	FlipHorizontal bool    `json:"flipHorizontal"`
	FlipVertical   bool    `json:"flipVertical"`
}

type VisualDetails struct {
	Sharpen  float64 `json:"sharpen"`
	Vignette float64 `json:"vignette"`
	Grain    float64 `json:"grain"`
}

func validVisualProcessing(processing *VisualProcessing) bool {
	if processing == nil {
		return true
	}
	switch processing.FrameInterpolation {
	case "", "off", "blend", "optical-flow":
	default:
		return false
	}
	stabilization := processing.Stabilization
	if stabilization.Enabled && (stabilization.Smoothness < 1 || stabilization.Smoothness > 100) {
		return false
	}
	if math.IsNaN(stabilization.Zoom) || math.IsInf(stabilization.Zoom, 0) || stabilization.Zoom < 0 || stabilization.Zoom > 30 {
		return false
	}
	values := []float64{processing.Color.Exposure, processing.Color.Temperature, processing.Color.Tint, processing.Color.Highlights, processing.Color.Shadows, processing.Color.Gamma, processing.ChromaKey.Similarity, processing.ChromaKey.Blend, processing.PersonSegmentation.Threshold, processing.PersonSegmentation.Feather, processing.ObjectRemoval.X, processing.ObjectRemoval.Y, processing.ObjectRemoval.Width, processing.ObjectRemoval.Height, processing.Mask.X, processing.Mask.Y, processing.Mask.Width, processing.Mask.Height, processing.Mask.Feather, processing.Geometry.CropLeft, processing.Geometry.CropTop, processing.Geometry.CropRight, processing.Geometry.CropBottom, processing.Details.Sharpen, processing.Details.Vignette, processing.Details.Grain}
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	color := processing.Color
	if !validHSLAdjustments(color.HSL) {
		return false
	}
	if !validColorGrade(color) {
		return false
	}
	if color.Exposure < -2 || color.Exposure > 2 || color.Temperature < -1 || color.Temperature > 1 || color.Tint < -1 || color.Tint > 1 || color.Highlights < -1 || color.Highlights > 1 || color.Shadows < -1 || color.Shadows > 1 || (color.Gamma != 0 && (color.Gamma < .5 || color.Gamma > 2)) {
		return false
	}
	if len(color.Curve) != 0 && len(color.Curve) != 5 {
		return false
	}
	for _, point := range color.Curve {
		if math.IsNaN(point) || math.IsInf(point, 0) || point < 0 || point > 1 {
			return false
		}
	}
	if len(color.CurvePoints) != 0 && (len(color.CurvePoints) < 2 || len(color.CurvePoints) > 16) {
		return false
	}
	for index, point := range color.CurvePoints {
		if math.IsNaN(point.X) || math.IsInf(point.X, 0) || math.IsNaN(point.Y) || math.IsInf(point.Y, 0) || point.X < 0 || point.X > 1 || point.Y < 0 || point.Y > 1 || (index > 0 && point.X <= color.CurvePoints[index-1].X) {
			return false
		}
	}
	if len(color.CurvePoints) > 0 && (color.CurvePoints[0].X != 0 || color.CurvePoints[len(color.CurvePoints)-1].X != 1) {
		return false
	}
	key := processing.ChromaKey
	if key.Enabled && !validTextColor(key.Color) || key.Similarity < 0 || key.Similarity > 1 || key.Blend < 0 || key.Blend > 1 {
		return false
	}
	person := processing.PersonSegmentation
	if person.Threshold < 0 || person.Threshold > 1 || person.Feather < 0 || person.Feather > .5 {
		return false
	}
	if person.Enabled && (strings.TrimSpace(person.MatteAssetID) == "" || strings.TrimSpace(person.Signature) == "") {
		return false
	}
	removal := processing.ObjectRemoval
	if removal.Enabled && (removal.X < 0 || removal.X > 1 || removal.Y < 0 || removal.Y > 1 || removal.Width < .05 || removal.Width > 1 || removal.Height < .05 || removal.Height > 1) {
		return false
	}
	mask := processing.Mask
	if mask.Type != "" && mask.Type != "none" && mask.Type != "rectangle" && mask.Type != "ellipse" && mask.Type != "polygon" && mask.Type != "bezier" {
		return false
	}
	if (mask.Type == "rectangle" || mask.Type == "ellipse") && (mask.X < 0 || mask.X > 1 || mask.Y < 0 || mask.Y > 1 || mask.Width < .05 || mask.Width > 1 || mask.Height < .05 || mask.Height > 1) {
		return false
	}
	if mask.Feather < 0 || mask.Feather > .5 {
		return false
	}
	if mask.Type == "polygon" {
		if len(mask.Points) < 3 || len(mask.Points) > 12 {
			return false
		}
		area := 0.0
		for index, point := range mask.Points {
			if math.IsNaN(point.X) || math.IsInf(point.X, 0) || math.IsNaN(point.Y) || math.IsInf(point.Y, 0) || point.X < 0 || point.X > 1 || point.Y < 0 || point.Y > 1 {
				return false
			}
			next := mask.Points[(index+1)%len(mask.Points)]
			area += point.X*next.Y - next.X*point.Y
		}
		if math.Abs(area)/2 < .0001 {
			return false
		}
	} else if mask.Type == "bezier" {
		if len(mask.BezierPoints) < 3 || len(mask.BezierPoints) > 8 {
			return false
		}
		for _, point := range mask.BezierPoints {
			for _, value := range []float64{point.X, point.Y, point.InX, point.InY, point.OutX, point.OutY} {
				if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 1 {
					return false
				}
			}
		}
		sampled := SampleVisualBezierMask(mask.BezierPoints)
		area := 0.0
		for index, point := range sampled {
			next := sampled[(index+1)%len(sampled)]
			area += point.X*next.Y - next.X*point.Y
		}
		if math.Abs(area)/2 < .0001 {
			return false
		}
		if len(mask.Points) != 0 {
			if len(mask.Points) != len(sampled) {
				return false
			}
			for index := range sampled {
				if math.Abs(mask.Points[index].X-sampled[index].X) > .000001 || math.Abs(mask.Points[index].Y-sampled[index].Y) > .000001 {
					return false
				}
			}
		}
	} else if len(mask.Points) != 0 || len(mask.BezierPoints) != 0 {
		return false
	}
	geometry := processing.Geometry
	if geometry.CropLeft < 0 || geometry.CropLeft > .9 || geometry.CropTop < 0 || geometry.CropTop > .9 || geometry.CropRight < 0 || geometry.CropRight > .9 || geometry.CropBottom < 0 || geometry.CropBottom > .9 || geometry.CropLeft+geometry.CropRight >= .95 || geometry.CropTop+geometry.CropBottom >= .95 {
		return false
	}
	details := processing.Details
	if details.Sharpen < 0 || details.Sharpen > 1 || details.Vignette < 0 || details.Vignette > 1 || details.Grain < 0 || details.Grain > 1 {
		return false
	}
	return true
}

type SpeedCurvePoint struct {
	Position float64 `json:"position"`
	Rate     float64 `json:"rate"`
}

type ClipKeyframe struct {
	ID       string          `json:"id"`
	OffsetMS int64           `json:"offsetMs"`
	Values   KeyframeValues  `json:"values"`
	Easing   *KeyframeEasing `json:"easing,omitempty"`
}

type KeyframeEasing struct {
	Type string  `json:"type"`
	X1   float64 `json:"x1"`
	Y1   float64 `json:"y1"`
	X2   float64 `json:"x2"`
	Y2   float64 `json:"y2"`
}

func NormalizeKeyframeEasing(easing *KeyframeEasing) KeyframeEasing {
	if easing == nil {
		return KeyframeEasing{Type: "linear", X1: 0, Y1: 0, X2: 1, Y2: 1}
	}
	presets := map[string]KeyframeEasing{
		"linear":      {Type: "linear", X1: 0, Y1: 0, X2: 1, Y2: 1},
		"ease-in":     {Type: "ease-in", X1: .42, Y1: 0, X2: 1, Y2: 1},
		"ease-out":    {Type: "ease-out", X1: 0, Y1: 0, X2: .58, Y2: 1},
		"ease-in-out": {Type: "ease-in-out", X1: .42, Y1: 0, X2: .58, Y2: 1},
		"hold":        {Type: "hold", X1: 0, Y1: 0, X2: 1, Y2: 0},
	}
	if preset, ok := presets[easing.Type]; ok {
		return preset
	}
	return *easing
}

func validKeyframeEasing(easing *KeyframeEasing) bool {
	if easing == nil {
		return true
	}
	if easing.Type != "linear" && easing.Type != "ease-in" && easing.Type != "ease-out" && easing.Type != "ease-in-out" && easing.Type != "hold" && easing.Type != "custom" {
		return false
	}
	for _, value := range []float64{easing.X1, easing.Y1, easing.X2, easing.Y2} {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 1 {
			return false
		}
	}
	return true
}

type KeyframeValues struct {
	X        *float64 `json:"x,omitempty"`
	Y        *float64 `json:"y,omitempty"`
	ScaleX   *float64 `json:"scaleX,omitempty"`
	ScaleY   *float64 `json:"scaleY,omitempty"`
	Rotation *float64 `json:"rotation,omitempty"`
	Opacity  *float64 `json:"opacity,omitempty"`
	Volume   *float64 `json:"volume,omitempty"`
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
	ProxyURL    string    `json:"proxyUrl,omitempty"`
	ByteSize    int64     `json:"byteSize"`
	DurationMS  int64     `json:"durationMs"`
	FontFamily  string    `json:"fontFamily,omitempty"`
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
	project.BackgroundMode = strings.ToLower(strings.TrimSpace(project.BackgroundMode))
	if project.BackgroundMode == "" {
		project.BackgroundMode = "color"
	}
	project.BackgroundColor = strings.ToUpper(strings.TrimSpace(project.BackgroundColor))
	if project.BackgroundColor == "" {
		project.BackgroundColor = "#000000"
	}
	if (project.BackgroundMode != "color" && project.BackgroundMode != "transparent") || !validProjectBackgroundColor(project.BackgroundColor) {
		return Project{}, ErrInvalidProject
	}
	return project, nil
}

func validProjectBackgroundColor(value string) bool {
	if len(value) != 7 || value[0] != '#' {
		return false
	}
	for _, character := range value[1:] {
		if !((character >= '0' && character <= '9') || (character >= 'A' && character <= 'F') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}

func validateTimeline(timeline Timeline) error {
	if len([]rune(timeline.Script)) > 200000 {
		return ErrInvalidTimeline
	}
	trackIDs := make(map[string]struct{}, len(timeline.Tracks))
	clipIDs := make(map[string]struct{})
	markerIDs := make(map[string]struct{}, len(timeline.Markers))
	for _, marker := range timeline.Markers {
		if strings.TrimSpace(marker.ID) == "" || marker.TimeMS < 0 || len([]rune(marker.Label)) > 80 || len([]rune(marker.Source)) > 32 || len([]rune(marker.AssetID)) > 96 || (marker.Color != "" && (len(marker.Color) != 7 || marker.Color[0] != '#')) {
			return ErrInvalidTimeline
		}
		if _, exists := markerIDs[marker.ID]; exists {
			return ErrInvalidTimeline
		}
		markerIDs[marker.ID] = struct{}{}
	}
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
		if track.Settings.TextResourceID != "" && !validTextResource(track.Settings.TextResourceID) {
			return ErrInvalidTimeline
		}
		settings := track.Settings
		if (settings.FontAssetID == "" && settings.FontFamily != "" && !validTextFont(settings.FontFamily)) ||
			(settings.FontAssetID != "" && (len([]rune(strings.TrimSpace(settings.FontAssetID))) > 96 || strings.TrimSpace(settings.FontAssetID) == "" || len([]rune(strings.TrimSpace(settings.FontFamily))) > 128 || strings.TrimSpace(settings.FontFamily) == "")) ||
			(settings.FontSize != 0 && (settings.FontSize < 12 || settings.FontSize > 240)) ||
			math.IsNaN(settings.TextCurve) || math.IsInf(settings.TextCurve, 0) || settings.TextCurve < -180 || settings.TextCurve > 180 ||
			(settings.TextColor != "" && !validTextColor(settings.TextColor)) ||
			(settings.StrokeColor != "" && !validTextColor(settings.StrokeColor)) ||
			(settings.BackgroundColor != "" && !validTextColor(settings.BackgroundColor)) ||
			(settings.ShadowColor != "" && !validTextColor(settings.ShadowColor)) ||
			settings.StrokeWidth < 0 || settings.StrokeWidth > 12 || settings.LetterSpacing < -10 || settings.LetterSpacing > 40 ||
			settings.ShadowOffsetX < -40 || settings.ShadowOffsetX > 40 || settings.ShadowOffsetY < -40 || settings.ShadowOffsetY > 40 || settings.ShadowBlur < 0 || settings.ShadowBlur > 20 ||
			(settings.LineHeight != 0 && (settings.LineHeight < .8 || settings.LineHeight > 3)) ||
			(settings.TextAlign != "" && settings.TextAlign != "left" && settings.TextAlign != "center" && settings.TextAlign != "right") {
			return ErrInvalidTimeline
		}
		if len(settings.TextRuns) > 32 {
			return ErrInvalidTimeline
		}
		for _, run := range settings.TextRuns {
			if run.Start < 0 || run.End <= run.Start || run.End > 512 ||
				(run.TextColor != "" && !validTextColor(run.TextColor)) ||
				(run.FontSize != 0 && (run.FontSize < 12 || run.FontSize > 240)) ||
				math.IsNaN(run.FontSize) || math.IsInf(run.FontSize, 0) {
				return ErrInvalidTimeline
			}
		}
		if len([]rune(track.Settings.CompoundID)) > 128 || len([]rune(track.Settings.CompoundName)) > 128 ||
			(track.Settings.CompoundName != "" && strings.TrimSpace(track.Settings.CompoundID) == "") {
			return ErrInvalidTimeline
		}
		for _, clip := range track.Clips {
			if strings.TrimSpace(clip.ID) == "" || !validClipKind(clip.Kind) || clip.StartMS < 0 || clip.DurationMS <= 0 || clip.SourceStartMS < 0 || clip.SourceDurationMS < 0 {
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
			if clip.Reversed && (clip.Kind != ClipVideo || track.Kind != TrackVideo || clip.FreezeSourceMS != nil) {
				return ErrInvalidTimeline
			}
			if clip.FreezeSourceMS != nil && (clip.Kind != ClipVideo || track.Kind != TrackVideo || *clip.FreezeSourceMS < 0 || (clip.SourceDurationMS > 0 && *clip.FreezeSourceMS >= clip.SourceDurationMS)) {
				return ErrInvalidTimeline
			}
			if len(clip.SpeedCurve) > 0 {
				if clip.Kind != ClipVideo || len(clip.SpeedCurve) < 2 || len(clip.SpeedCurve) > 12 || clip.SpeedCurve[0].Position != 0 || clip.SpeedCurve[len(clip.SpeedCurve)-1].Position != 1 {
					return ErrInvalidTimeline
				}
				lastPosition := -1.0
				for _, point := range clip.SpeedCurve {
					if math.IsNaN(point.Position) || math.IsInf(point.Position, 0) || math.IsNaN(point.Rate) || math.IsInf(point.Rate, 0) || point.Position < 0 || point.Position > 1 || point.Position <= lastPosition || point.Rate < .25 || point.Rate > 4 {
						return ErrInvalidTimeline
					}
					lastPosition = point.Position
				}
			}
			consumedSourceMS := float64(clip.DurationMS)
			if len(clip.SpeedCurve) >= 2 {
				averageRate := 0.0
				for index := 1; index < len(clip.SpeedCurve); index++ {
					left, right := clip.SpeedCurve[index-1], clip.SpeedCurve[index]
					averageRate += (right.Position - left.Position) * (left.Rate + right.Rate) / 2
				}
				consumedSourceMS *= averageRate
			} else if clip.PlaybackRate > 0 {
				consumedSourceMS *= clip.PlaybackRate
			}
			if clip.FreezeSourceMS == nil && clip.SourceDurationMS > 0 && float64(clip.SourceStartMS)+consumedSourceMS > float64(clip.SourceDurationMS)+.5 {
				return ErrInvalidTimeline
			}
			if clip.Volume < 0 || clip.Volume > 2 || clip.FadeInMS < 0 || clip.FadeOutMS < 0 || clip.FadeInMS > clip.DurationMS || clip.FadeOutMS > clip.DurationMS {
				return ErrInvalidTimeline
			}
			if clip.AudioProcessing != nil && (clip.Kind != ClipAudio || track.Kind != TrackAudio || !validAudioProcessing(clip.AudioProcessing)) {
				return ErrInvalidTimeline
			}
			if clip.VisualProcessing != nil && ((clip.Kind != ClipVideo && clip.Kind != ClipImage) || track.Kind != TrackVideo || !validVisualProcessing(clip.VisualProcessing)) {
				return ErrInvalidTimeline
			}
			if clip.VisualProcessing != nil && clip.Kind != ClipVideo && clip.VisualProcessing.FrameInterpolation != "" && clip.VisualProcessing.FrameInterpolation != "off" {
				return ErrInvalidTimeline
			}
			if clip.VisualProcessing != nil && clip.Kind != ClipVideo && clip.VisualProcessing.Stabilization.Enabled {
				return ErrInvalidTimeline
			}
			if !validBlendMode(clip.BlendMode) || (clip.BlendMode != "" && clip.BlendMode != "normal" && ((clip.Kind != ClipVideo && clip.Kind != ClipImage) || track.Kind != TrackVideo)) {
				return ErrInvalidTimeline
			}
			if clip.ReviewStatus != "" && clip.ReviewStatus != "delete" && clip.ReviewStatus != "review" && clip.ReviewStatus != "keep" {
				return ErrInvalidTimeline
			}
			if len([]rune(clip.ReviewReason)) > 1000 {
				return ErrInvalidTimeline
			}
			if len(clip.CompoundPath) > 8 {
				return ErrInvalidTimeline
			}
			compoundIDs := make(map[string]struct{}, len(clip.CompoundPath))
			for _, ref := range clip.CompoundPath {
				id := strings.TrimSpace(ref.ID)
				if id == "" || len([]rune(id)) > 128 || len([]rune(ref.Name)) > 128 {
					return ErrInvalidTimeline
				}
				if _, exists := compoundIDs[id]; exists {
					return ErrInvalidTimeline
				}
				compoundIDs[id] = struct{}{}
			}
			if len(clip.CompoundPath) > 0 {
				outer := clip.CompoundPath[len(clip.CompoundPath)-1]
				if strings.TrimSpace(clip.CompoundID) != strings.TrimSpace(outer.ID) || clip.CompoundName != outer.Name {
					return ErrInvalidTimeline
				}
			}
			if !validVisualEffect(clip.VisualEffect) {
				return ErrInvalidTimeline
			}
			if clip.CatalogResourceID != "" && !validCatalogVisualResource(clip.CatalogResourceID) {
				return ErrInvalidTimeline
			}
			if clip.TextResourceID != "" && !validTextResource(clip.TextResourceID) {
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
			if clip.Kind == ClipTransition {
				if track.Kind != TrackTransition || !strings.HasPrefix(clip.ResourceID, "transition-") || clip.Transition != clip.ResourceID || clip.TransitionMS != clip.DurationMS ||
					strings.TrimSpace(clip.TransitionTrackID) == "" || strings.TrimSpace(clip.TransitionLeftClipID) == "" || strings.TrimSpace(clip.TransitionRightClipID) == "" ||
					clip.TransitionLeftClipID == clip.TransitionRightClipID || clip.DurationMS < 100 || clip.DurationMS > 5000 {
					return ErrInvalidTimeline
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
			keyframeIDs := make(map[string]struct{}, len(clip.Keyframes))
			lastOffset := int64(-1)
			for _, frame := range clip.Keyframes {
				if strings.TrimSpace(frame.ID) == "" || frame.OffsetMS < 0 || frame.OffsetMS > clip.DurationMS || frame.OffsetMS <= lastOffset {
					return ErrInvalidTimeline
				}
				if _, exists := keyframeIDs[frame.ID]; exists {
					return ErrInvalidTimeline
				}
				keyframeIDs[frame.ID] = struct{}{}
				if !validKeyframeEasing(frame.Easing) {
					return ErrInvalidTimeline
				}
				lastOffset = frame.OffsetMS
				values := frame.Values
				for _, value := range []*float64{values.X, values.Y, values.ScaleX, values.ScaleY, values.Rotation, values.Opacity, values.Volume} {
					if value != nil && (math.IsNaN(*value) || math.IsInf(*value, 0)) {
						return ErrInvalidTimeline
					}
				}
				if (values.X != nil && (*values.X < -1 || *values.X > 1)) || (values.Y != nil && (*values.Y < -1 || *values.Y > 1)) ||
					(values.ScaleX != nil && (*values.ScaleX < .05 || *values.ScaleX > 20)) || (values.ScaleY != nil && (*values.ScaleY < .05 || *values.ScaleY > 20)) ||
					(values.Rotation != nil && (*values.Rotation < -180 || *values.Rotation > 180)) || (values.Opacity != nil && (*values.Opacity < 0 || *values.Opacity > 1)) ||
					(values.Volume != nil && (*values.Volume < 0 || *values.Volume > 2)) {
					return ErrInvalidTimeline
				}
			}
			if _, exists := clipIDs[clip.ID]; exists {
				return ErrInvalidTimeline
			}
			clipIDs[clip.ID] = struct{}{}
		}
	}
	transitionTargets := make(map[string]struct{})
	for _, transitionTrack := range timeline.Tracks {
		if transitionTrack.Kind != TrackTransition {
			continue
		}
		for _, transitionClip := range transitionTrack.Clips {
			var videoTrack *Track
			for index := range timeline.Tracks {
				if timeline.Tracks[index].ID == transitionClip.TransitionTrackID && timeline.Tracks[index].Kind == TrackVideo {
					videoTrack = &timeline.Tracks[index]
					break
				}
			}
			if videoTrack == nil {
				return ErrInvalidTimeline
			}
			ordered := make([]Clip, 0, len(videoTrack.Clips))
			for _, clip := range videoTrack.Clips {
				if clip.Kind == ClipVideo || clip.Kind == ClipImage {
					ordered = append(ordered, clip)
				}
			}
			sort.Slice(ordered, func(i, j int) bool { return ordered[i].StartMS < ordered[j].StartMS })
			matched := false
			for index := 1; index < len(ordered); index++ {
				left, right := ordered[index-1], ordered[index]
				if left.ID != transitionClip.TransitionLeftClipID || right.ID != transitionClip.TransitionRightClipID {
					continue
				}
				if left.StartMS+left.DurationMS != right.StartMS || right.Transition != transitionClip.ResourceID || right.TransitionMS != transitionClip.DurationMS ||
					transitionClip.StartMS != max(int64(0), right.StartMS-transitionClip.DurationMS/2) {
					return ErrInvalidTimeline
				}
				matched = true
				break
			}
			if !matched {
				return ErrInvalidTimeline
			}
			targetKey := transitionClip.TransitionTrackID + "\x00" + transitionClip.TransitionRightClipID
			if _, exists := transitionTargets[targetKey]; exists {
				return ErrInvalidTimeline
			}
			transitionTargets[targetKey] = struct{}{}
		}
	}
	return nil
}

func validBlendMode(value string) bool {
	switch value {
	case "", "normal", "multiply", "screen", "overlay", "soft-light", "difference", "add":
		return true
	default:
		return false
	}
}

func validCatalogVisualResource(value string) bool {
	switch value {
	case "filter-original", "filter-bright", "filter-clear", "filter-natural", "filter-warm-sun", "filter-cool-blue", "filter-cinema", "filter-retro",
		"filter-monochrome", "filter-high-contrast", "filter-low-saturation", "filter-teal-orange", "filter-night", "filter-portrait", "filter-vivid", "filter-soft",
		"filter-golden-hour", "filter-noir", "filter-vintage", "filter-cyberpunk", "filter-vaporwave", "filter-mint-haze", "filter-bleach-bypass", "filter-winter-cool", "filter-rosy", "filter-film", "filter-super-8", "filter-instant-print":
		return true
	default:
		return false
	}
}

func validTextResource(value string) bool {
	switch value {
	case "text-basic-caption", "text-bold-title", "text-outline-caption", "text-shadow-caption", "text-panel-caption", "text-word-highlight",
		"text-news-bar", "text-lower-third", "text-comic", "text-handwriting", "text-vertical", "text-gradient-title":
		return true
	default:
		return false
	}
}

func validTextFont(value string) bool {
	switch value {
	case "Microsoft YaHei", "SimHei", "SimSun", "KaiTi", "Arial":
		return true
	default:
		return false
	}
}

func validTextColor(value string) bool {
	if len(value) != 7 && len(value) != 9 {
		return false
	}
	if value[0] != '#' {
		return false
	}
	for _, character := range value[1:] {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f') || (character >= 'A' && character <= 'F')) {
			return false
		}
	}
	return true
}

func validTrackKind(kind TrackKind) bool {
	return kind == TrackVideo || kind == TrackAudio || kind == TrackSubtitle || kind == TrackEffect || kind == TrackTransition
}

func validClipKind(kind ClipKind) bool {
	return kind == ClipVideo || kind == ClipAudio || kind == ClipImage || kind == ClipSubtitle || kind == ClipEffect || kind == ClipTransition
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
		(asset.Kind == "subtitle" && asset.ContentType != "application/x-subrip" && asset.ContentType != "text/plain") ||
		(asset.Kind == "font" && asset.ContentType != "font/ttf" && asset.ContentType != "font/otf") {
		return ErrInvalidAsset
	}
	if asset.Kind != "image" && asset.Kind != "video" && asset.Kind != "audio" && asset.Kind != "subtitle" && asset.Kind != "font" {
		return ErrInvalidAsset
	}
	return nil
}
