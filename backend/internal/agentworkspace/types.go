package agentworkspace

import "time"

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
	GroupID       string         `json:"groupId,omitempty"`
}

type Edge struct {
	ID       string `json:"id"`
	Source   string `json:"source"`
	Target   string `json:"target"`
	Relation string `json:"relation"`
}

type Canvas struct {
	Version  int            `json:"version"`
	Revision int64          `json:"revision"`
	Viewport map[string]any `json:"viewport,omitempty"`
	Nodes    []Node         `json:"nodes"`
	Edges    []Edge         `json:"edges"`
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

type LegacyImport struct {
	OwnerUsername    string
	SourcePath       string
	SourceSHA256     string
	ImportedProjects int
	ImportedAt       time.Time
}

func EmptyCanvas() Canvas {
	return Canvas{Version: 1, Nodes: []Node{}, Edges: []Edge{}}
}
