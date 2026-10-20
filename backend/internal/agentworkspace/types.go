package agentworkspace

import (
	"encoding/json"
	"strings"
	"time"
)

type ProjectSummary struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	TitleMode    string    `json:"titleMode"`
	Preview      string    `json:"preview"`
	MessageCount int       `json:"messageCount"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Project struct {
	ID         string      `json:"id"`
	Title      string      `json:"title"`
	TitleMode  string      `json:"titleMode"`
	Preview    string      `json:"preview"`
	CreatedAt  time.Time   `json:"createdAt"`
	UpdatedAt  time.Time   `json:"updatedAt"`
	Messages   []Message   `json:"messages"`
	Canvas     Canvas      `json:"canvas"`
	Executions []Execution `json:"executions"`
	DeletedAt  *time.Time  `json:"-"`
}

type Message struct {
	ID        int64          `json:"id,omitempty"`
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	Context   map[string]any `json:"context,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Node struct {
	ID            string         `json:"id"`
	Type          string         `json:"type"`
	Position      Position       `json:"position"`
	Title         string         `json:"title"`
	Version       int64          `json:"version"`
	SourceNodeIDs []string       `json:"sourceNodeIds"`
	Status        string         `json:"status"`
	Summary       string         `json:"summary"`
	AssetRef      map[string]any `json:"assetRef,omitempty"`
	Config        map[string]any `json:"config,omitempty"`
	GroupID       string         `json:"groupId,omitempty"`
}

type Edge struct {
	ID       string `json:"id"`
	Source   string `json:"source"`
	Target   string `json:"target"`
	Relation string `json:"relation"`
}

type Appearance struct {
	EdgeColor       string `json:"edgeColor"`
	EdgeAnimated    bool   `json:"edgeAnimated"`
	EdgeWidth       string `json:"edgeWidth"`
	BackgroundDots  bool   `json:"backgroundDots"`
	AlignmentGuides bool   `json:"alignmentGuides"`
	EdgesVisible    bool   `json:"edgesVisible"`
}

func DefaultAppearance() Appearance {
	return Appearance{
		EdgeColor:       "coral",
		EdgeAnimated:    true,
		EdgeWidth:       "default",
		BackgroundDots:  true,
		AlignmentGuides: true,
		EdgesVisible:    true,
	}
}

func (appearance *Appearance) UnmarshalJSON(data []byte) error {
	type appearanceAlias Appearance
	decoded := appearanceAlias(DefaultAppearance())
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*appearance = Appearance(decoded)
	return nil
}

func normalizeCanvasAppearance(input Appearance, configured bool) Appearance {
	if !configured {
		return DefaultAppearance()
	}
	appearance := DefaultAppearance()
	if color := strings.ToLower(strings.TrimSpace(input.EdgeColor)); color == "coral" || color == "cyan" || color == "violet" || color == "gold" {
		appearance.EdgeColor = color
	}
	if width := strings.ToLower(strings.TrimSpace(input.EdgeWidth)); width == "thin" || width == "default" || width == "thick" {
		appearance.EdgeWidth = width
	}
	appearance.EdgeAnimated = input.EdgeAnimated
	appearance.BackgroundDots = input.BackgroundDots
	appearance.AlignmentGuides = input.AlignmentGuides
	appearance.EdgesVisible = input.EdgesVisible
	return appearance
}

type Canvas struct {
	Version              int            `json:"version"`
	Revision             int64          `json:"revision"`
	Viewport             map[string]any `json:"viewport,omitempty"`
	ViewState            map[string]any `json:"viewState,omitempty"`
	Appearance           Appearance     `json:"appearance"`
	Nodes                []Node         `json:"nodes"`
	Edges                []Edge         `json:"edges"`
	appearanceConfigured bool
}

func (canvas *Canvas) UnmarshalJSON(data []byte) error {
	var decoded struct {
		Version    int            `json:"version"`
		Revision   int64          `json:"revision"`
		Viewport   map[string]any `json:"viewport,omitempty"`
		ViewState  map[string]any `json:"viewState,omitempty"`
		Appearance Appearance     `json:"appearance"`
		Nodes      []Node         `json:"nodes"`
		Edges      []Edge         `json:"edges"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	canvas.Version = decoded.Version
	canvas.Revision = decoded.Revision
	canvas.Viewport = decoded.Viewport
	canvas.ViewState = decoded.ViewState
	canvas.Nodes = decoded.Nodes
	canvas.Edges = decoded.Edges
	canvas.Appearance = decoded.Appearance
	rawAppearance, configured := fields["appearance"]
	canvas.appearanceConfigured = configured && string(rawAppearance) != "null"
	canvas.Appearance = normalizeCanvasAppearance(canvas.Appearance, canvas.appearanceConfigured)
	return nil
}

type Execution struct {
	ID             string         `json:"id"`
	ToolName       string         `json:"toolName"`
	Status         string         `json:"status"`
	Detail         map[string]any `json:"detail,omitempty"`
	IdempotencyKey string         `json:"idempotencyKey,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}

type ApplyNodeExecutionInput struct {
	ExecutionID string
	NodeID      string
	NodeVersion int64
	NodeType    string
	Summary     string
	AssetRef    map[string]any
	Detail      map[string]any
	Now         time.Time
}

type LegacyImport struct {
	OwnerUsername    string
	SourcePath       string
	SourceSHA256     string
	ImportedProjects int
	ImportedAt       time.Time
}

func EmptyCanvas() Canvas {
	return Canvas{Version: 1, Appearance: DefaultAppearance(), Nodes: []Node{}, Edges: []Edge{}, appearanceConfigured: true}
}
