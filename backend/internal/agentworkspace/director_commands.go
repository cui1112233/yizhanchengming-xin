package agentworkspace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrDirectorCommandForbidden      = errors.New("director command is forbidden")
	ErrDirectorCommandUnknown        = errors.New("director command is unknown")
	ErrDirectorCommandInvalid        = errors.New("director command is invalid")
	ErrDirectorMissingConfiguration  = errors.New("director command is missing configuration")
	ErrDirectorCommandNeedsExecution = errors.New("director command requires task execution")
)

const (
	DirectorCommandCreateNode      = "create_node"
	DirectorCommandUpdateNode      = "update_node"
	DirectorCommandConnectNodes    = "connect_nodes"
	DirectorCommandDisconnectNodes = "disconnect_nodes"
	DirectorCommandMoveNodes       = "move_nodes"
	DirectorCommandRemoveNodes     = "remove_nodes"
	DirectorCommandStartGeneration = "start_generation"
	DirectorCommandApplyResult     = "apply_result"
)

type DirectorPlan struct {
	ID       string            `json:"id"`
	Mode     string            `json:"mode,omitempty"`
	Status   string            `json:"status,omitempty"`
	Summary  string            `json:"summary"`
	Commands []DirectorCommand `json:"commands"`
}

type DirectorCommand struct {
	Type         string         `json:"type"`
	NodeID       string         `json:"nodeId,omitempty"`
	NodeIDs      []string       `json:"nodeIds,omitempty"`
	SourceNodeID string         `json:"sourceNodeId,omitempty"`
	TargetNodeID string         `json:"targetNodeId,omitempty"`
	EdgeID       string         `json:"edgeId,omitempty"`
	Relation     string         `json:"relation,omitempty"`
	Node         *Node          `json:"node,omitempty"`
	Patch        map[string]any `json:"patch,omitempty"`
	Position     *Position      `json:"position,omitempty"`
	ResultID     string         `json:"resultId,omitempty"`
}

func ValidateDirectorPlan(project Project, plan DirectorPlan) error {
	if strings.TrimSpace(plan.ID) == "" || len(plan.Commands) == 0 || len(plan.Commands) > 50 {
		return ErrDirectorCommandInvalid
	}
	nodes := make(map[string]Node, len(project.Canvas.Nodes))
	for _, node := range project.Canvas.Nodes {
		if id := strings.TrimSpace(node.ID); id != "" {
			nodes[id] = node
		}
	}
	for index, command := range plan.Commands {
		if err := validateDirectorCommand(nodes, command); err != nil {
			return fmt.Errorf("command %d: %w", index+1, err)
		}
		switch command.Type {
		case DirectorCommandCreateNode:
			nodes[command.Node.ID] = *command.Node
		case DirectorCommandRemoveNodes:
			for _, id := range directorCommandNodeIDs(command) {
				delete(nodes, id)
			}
		}
	}
	return nil
}

func validateDirectorCommand(nodes map[string]Node, command DirectorCommand) error {
	switch command.Type {
	case "delete_asset":
		return ErrDirectorCommandForbidden
	case DirectorCommandCreateNode:
		if command.Node == nil || !validDirectorNode(*command.Node) {
			return ErrDirectorCommandInvalid
		}
		if _, exists := nodes[command.Node.ID]; exists {
			return ErrDirectorCommandInvalid
		}
	case DirectorCommandUpdateNode:
		if !directorNodeExists(nodes, command.NodeID) || !validDirectorNodePatch(command.Patch) {
			return ErrDirectorCommandInvalid
		}
	case DirectorCommandConnectNodes:
		if command.SourceNodeID == command.TargetNodeID || !directorNodeExists(nodes, command.SourceNodeID) || !directorNodeExists(nodes, command.TargetNodeID) || !validDirectorRelation(command.Relation) {
			return ErrDirectorCommandInvalid
		}
	case DirectorCommandDisconnectNodes:
		if strings.TrimSpace(command.EdgeID) == "" && (strings.TrimSpace(command.SourceNodeID) == "" || strings.TrimSpace(command.TargetNodeID) == "") {
			return ErrDirectorCommandInvalid
		}
	case DirectorCommandMoveNodes:
		if command.Position == nil || !directorNodeExists(nodes, command.NodeID) {
			return ErrDirectorCommandInvalid
		}
	case DirectorCommandRemoveNodes:
		ids := directorCommandNodeIDs(command)
		if len(ids) == 0 {
			return ErrDirectorCommandInvalid
		}
		for _, id := range ids {
			if !directorNodeExists(nodes, id) {
				return ErrDirectorCommandInvalid
			}
		}
	case DirectorCommandStartGeneration:
		node, exists := nodes[strings.TrimSpace(command.NodeID)]
		if !exists {
			return ErrDirectorCommandInvalid
		}
		if firstDirectorString(node.Config, "modelId", "model") == "" {
			return ErrDirectorMissingConfiguration
		}
	case DirectorCommandApplyResult:
		if !directorNodeExists(nodes, command.NodeID) || strings.TrimSpace(command.ResultID) == "" {
			return ErrDirectorCommandInvalid
		}
	default:
		return ErrDirectorCommandUnknown
	}
	return nil
}

func ApplyDirectorPlan(ctx context.Context, store Store, owner, projectID string, expectedRevision int64, plan DirectorPlan, now time.Time) (Canvas, error) {
	project, err := store.Project(ctx, owner, projectID)
	if err != nil {
		return Canvas{}, err
	}
	if project.Canvas.Revision != expectedRevision {
		return Canvas{}, ErrRevisionConflict
	}
	if err := ValidateDirectorPlan(project, plan); err != nil {
		return Canvas{}, err
	}
	for _, command := range plan.Commands {
		if command.Type == DirectorCommandStartGeneration || command.Type == DirectorCommandApplyResult {
			return Canvas{}, ErrDirectorCommandNeedsExecution
		}
	}
	canvas, err := cloneDirectorCanvas(project.Canvas)
	if err != nil {
		return Canvas{}, err
	}
	for index, command := range plan.Commands {
		applyDirectorStructuralCommand(&canvas, command, index)
	}
	return store.SaveCanvas(ctx, owner, projectID, canvas, expectedRevision, now)
}

func applyDirectorStructuralCommand(canvas *Canvas, command DirectorCommand, index int) {
	switch command.Type {
	case DirectorCommandCreateNode:
		node := *command.Node
		if node.Version < 1 {
			node.Version = 1
		}
		if strings.TrimSpace(node.Status) == "" {
			node.Status = "draft"
		}
		canvas.Nodes = append(canvas.Nodes, node)
	case DirectorCommandUpdateNode:
		for nodeIndex := range canvas.Nodes {
			if canvas.Nodes[nodeIndex].ID != command.NodeID {
				continue
			}
			patchDirectorNode(&canvas.Nodes[nodeIndex], command.Patch)
			canvas.Nodes[nodeIndex].Version++
			break
		}
	case DirectorCommandConnectNodes:
		for _, edge := range canvas.Edges {
			if edge.Source == command.SourceNodeID && edge.Target == command.TargetNodeID && edge.Relation == command.Relation {
				return
			}
		}
		edgeID := strings.TrimSpace(command.EdgeID)
		if edgeID == "" {
			edgeID = fmt.Sprintf("director-edge-%s-%s-%d", command.SourceNodeID, command.TargetNodeID, index+1)
		}
		canvas.Edges = append(canvas.Edges, Edge{ID: edgeID, Source: command.SourceNodeID, Target: command.TargetNodeID, Relation: command.Relation})
	case DirectorCommandDisconnectNodes:
		edges := canvas.Edges[:0]
		for _, edge := range canvas.Edges {
			matchesID := command.EdgeID != "" && edge.ID == command.EdgeID
			matchesPair := command.EdgeID == "" && edge.Source == command.SourceNodeID && edge.Target == command.TargetNodeID
			if !matchesID && !matchesPair {
				edges = append(edges, edge)
			}
		}
		canvas.Edges = edges
	case DirectorCommandMoveNodes:
		for nodeIndex := range canvas.Nodes {
			if canvas.Nodes[nodeIndex].ID == command.NodeID {
				canvas.Nodes[nodeIndex].Position = *command.Position
				break
			}
		}
	case DirectorCommandRemoveNodes:
		removed := make(map[string]bool)
		for _, id := range directorCommandNodeIDs(command) {
			removed[id] = true
		}
		nodes := canvas.Nodes[:0]
		for _, node := range canvas.Nodes {
			if !removed[node.ID] {
				nodes = append(nodes, node)
			}
		}
		canvas.Nodes = nodes
		edges := canvas.Edges[:0]
		for _, edge := range canvas.Edges {
			if !removed[edge.Source] && !removed[edge.Target] {
				edges = append(edges, edge)
			}
		}
		canvas.Edges = edges
	}
}

func patchDirectorNode(node *Node, patch map[string]any) {
	if title, ok := patch["title"].(string); ok && strings.TrimSpace(title) != "" {
		node.Title = strings.TrimSpace(title)
	}
	if summary, ok := patch["summary"].(string); ok {
		node.Summary = summary
	}
	if status, ok := patch["status"].(string); ok && strings.TrimSpace(status) != "" {
		node.Status = strings.TrimSpace(status)
	}
	if groupID, ok := patch["groupId"].(string); ok {
		node.GroupID = strings.TrimSpace(groupID)
	}
	if config, ok := patch["config"].(map[string]any); ok {
		if node.Config == nil {
			node.Config = map[string]any{}
		}
		for key, value := range config {
			node.Config[key] = value
		}
	}
}

func validDirectorNode(node Node) bool {
	if strings.TrimSpace(node.ID) == "" || strings.TrimSpace(node.Title) == "" {
		return false
	}
	switch strings.TrimSpace(node.Type) {
	case "text", "image", "video", "audio":
		return true
	default:
		return false
	}
}

func validDirectorNodePatch(patch map[string]any) bool {
	if len(patch) == 0 {
		return false
	}
	for key := range patch {
		switch key {
		case "title", "summary", "status", "config", "groupId":
		default:
			return false
		}
	}
	return true
}

func validDirectorRelation(value string) bool {
	switch strings.TrimSpace(value) {
	case "reference", "first_frame", "last_frame", "audio_source", "script_source", "sequence":
		return true
	default:
		return false
	}
}

func directorNodeExists(nodes map[string]Node, id string) bool {
	_, exists := nodes[strings.TrimSpace(id)]
	return exists
}

func directorCommandNodeIDs(command DirectorCommand) []string {
	ids := append([]string(nil), command.NodeIDs...)
	if strings.TrimSpace(command.NodeID) != "" {
		ids = append(ids, command.NodeID)
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

func firstDirectorString(input map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := input[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func cloneDirectorCanvas(canvas Canvas) (Canvas, error) {
	raw, err := json.Marshal(canvas)
	if err != nil {
		return Canvas{}, err
	}
	var cloned Canvas
	if err := json.Unmarshal(raw, &cloned); err != nil {
		return Canvas{}, err
	}
	return cloned, nil
}
