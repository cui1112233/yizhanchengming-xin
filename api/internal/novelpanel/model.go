package novelpanel

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("novelpanel: not found")
	ErrConflict = errors.New("novelpanel: revision conflict")
	ErrInvalid  = errors.New("novelpanel: invalid workspace")
)

const (
	MaxNovelRunes        = 500000
	MaxForcedRosterRunes = 30000
	MaxStyleRunes        = 6000
	MaxInstructionRunes  = 30000
	MaxCaseLearningRunes = 30000
)

type Mode string

const (
	ModeNormal             Mode = "normal"
	ModePremiumIllustrated Mode = "premium_illustrated"
)

type Density string

const (
	DensityCompact  Density = "compact"
	DensityStandard Density = "standard"
	DensityDetailed Density = "detailed"
)

type TextProcessingSettings struct {
	TrimLineWhitespace       bool `json:"trimLineWhitespace"`
	CollapseBlankLines       bool `json:"collapseBlankLines"`
	NormalizeFullWidthSpaces bool `json:"normalizeFullWidthSpaces"`
}

type InstructionSettings struct {
	GenerationRules        string `json:"generationRules"`
	MustCoverDetails       string `json:"mustCoverDetails"`
	ShotRhythmRequirements string `json:"shotRhythmRequirements"`
	NegativeInstructions   string `json:"negativeInstructions"`
}

type CaseLearningSettings struct {
	Enabled  bool   `json:"enabled"`
	Material string `json:"material"`
}

type CharacterCard struct {
	ID           string   `json:"id"`
	SourceEntry  string   `json:"sourceEntry"`
	DisplayName  string   `json:"displayName"`
	BaseName     string   `json:"baseName"`
	Aliases      []string `json:"aliases,omitempty"`
	BracketHints []string `json:"bracketHints,omitempty"`
	StageHint    string   `json:"stageHint,omitempty"`
	Gender       string   `json:"gender,omitempty"`
	Appearance   string   `json:"appearance,omitempty"`
	CardNote     string   `json:"cardNote,omitempty"`
	AssetRefs    []string `json:"assetRefs,omitempty"`
}

type CharacterRelationship struct {
	ID          string `json:"id"`
	FromID      string `json:"fromId"`
	ToID        string `json:"toId"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

type Shot struct {
	ID            string   `json:"id"`
	SourceIndex   int      `json:"sourceIndex"`
	SourceBasis   string   `json:"sourceBasis"`
	Visual        string   `json:"visual"`
	Camera        string   `json:"camera,omitempty"`
	SceneRef      string   `json:"sceneRef,omitempty"`
	CharacterRefs []string `json:"characterRefs,omitempty"`
	DurationSec   float64  `json:"durationSec,omitempty"`
	Note          string   `json:"note,omitempty"`
}

type Workspace struct {
	ProjectID      int64                   `json:"projectId"`
	Revision       int64                   `json:"revision"`
	Mode           Mode                    `json:"mode"`
	OriginalText   string                  `json:"originalText"`
	ProcessedText  string                  `json:"processedText"`
	TextProcessing TextProcessingSettings  `json:"textProcessing"`
	ForcedRoster   string                  `json:"forcedRoster"`
	Characters     []CharacterCard         `json:"characters"`
	Relationships  []CharacterRelationship `json:"relationships"`
	ContentType    string                  `json:"contentType"`
	UnifiedStyle   string                  `json:"unifiedStyle"`
	Density        Density                 `json:"density"`
	CaseLearning   CaseLearningSettings    `json:"caseLearning"`
	Instructions   InstructionSettings     `json:"instructions"`
	Shots          []Shot                  `json:"shots"`
	CreatedAt      time.Time               `json:"createdAt,omitempty"`
	UpdatedAt      time.Time               `json:"updatedAt,omitempty"`
}

type HistorySummary struct {
	SourcePreview  string  `json:"sourcePreview"`
	CharacterCount int     `json:"characterCount"`
	RelationCount  int     `json:"relationCount"`
	ShotCount      int     `json:"shotCount"`
	DurationSec    float64 `json:"durationSec"`
	Mode           Mode    `json:"mode"`
}

type HistoryRecord struct {
	ID        string         `json:"id"`
	ProjectID int64          `json:"projectId"`
	Revision  int64          `json:"revision"`
	Note      string         `json:"note,omitempty"`
	Summary   HistorySummary `json:"summary"`
	Workspace Workspace      `json:"workspace"`
	CreatedAt time.Time      `json:"createdAt"`
}

type SaveRequest struct {
	Workspace        Workspace `json:"workspace"`
	ExpectedRevision int64     `json:"expectedRevision"`
	Note             string    `json:"note,omitempty"`
}

type SaveResult struct {
	Workspace Workspace     `json:"workspace"`
	History   HistoryRecord `json:"history"`
}

type RestoreRequest struct {
	ProjectID        int64  `json:"projectId"`
	HistoryID        string `json:"historyId"`
	ExpectedRevision int64  `json:"expectedRevision"`
}

type SourceLine struct {
	Index   int    `json:"index"`
	RawLine int    `json:"rawLine"`
	Text    string `json:"text"`
}

type StoryboardRequest struct {
	ProjectID     int64                   `json:"projectId"`
	Mode          Mode                    `json:"mode"`
	ContentType   string                  `json:"contentType"`
	UnifiedStyle  string                  `json:"unifiedStyle"`
	Density       Density                 `json:"density"`
	SourceLines   []SourceLine            `json:"sourceLines"`
	Characters    []CharacterCard         `json:"characters"`
	Relationships []CharacterRelationship `json:"relationships"`
	CaseLearning  CaseLearningSettings    `json:"caseLearning"`
	Instructions  InstructionSettings     `json:"instructions"`
	CurrentShots  []Shot                  `json:"currentShots,omitempty"`
}
