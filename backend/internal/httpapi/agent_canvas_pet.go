package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"qiantie/backend/internal/agentworkspace"
)

// The client supplies IDs and revision only; node text comes from the owned store.
func prepareCanvasPetContext(project agentworkspace.Project, input map[string]any) (map[string]any, error) {
	raw, exists := input["canvasPet"]
	if !exists {
		return input, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var request struct {
		Revision        *int64   `json:"revision"`
		SelectedNodeIDs []string `json:"selectedNodeIds"`
	}
	if json.Unmarshal(data, &request) != nil || request.Revision == nil || len(request.SelectedNodeIDs) > 50 {
		return nil, errors.New("画布上下文无效")
	}
	if *request.Revision != project.Canvas.Revision {
		return nil, agentworkspace.ErrRevisionConflict
	}
	byID := map[string]agentworkspace.Node{}
	nodes := []any{}
	for _, n := range project.Canvas.Nodes {
		byID[n.ID] = n
		if len(nodes) < 50 {
			nodes = append(nodes, canvasPetNode(n))
		}
	}
	selected := []any{}
	seen := map[string]bool{}
	for _, id := range request.SelectedNodeIDs {
		n, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("选中卡片已失效，请重新选择")
		}
		if !seen[id] {
			selected = append(selected, canvasPetNode(n))
			seen[id] = true
		}
	}
	output := map[string]any{}
	for k, v := range input {
		output[k] = v
	}
	output["selectedNodes"] = selected
	output["canvasNodes"] = nodes
	edges := []any{}
	for _, edge := range project.Canvas.Edges {
		if len(edges) < 100 {
			edges = append(edges, map[string]any{"id": edge.ID, "source": edge.Source, "target": edge.Target, "relation": edge.Relation})
		}
	}
	output["canvasEdges"] = edges
	output["canvasPet"] = map[string]any{"revision": project.Canvas.Revision, "nodeCount": len(project.Canvas.Nodes), "includedNodeCount": len(nodes), "selectedNodeIds": request.SelectedNodeIDs}
	output["director"] = map[string]any{"enabled": true, "mode": "supervised"}
	history := []any{}
	start := len(project.Messages) - 8
	if start < 0 {
		start = 0
	}
	for _, entry := range project.Messages[start:] {
		body := []rune(entry.Content)
		if len(body) > 1200 {
			body = body[:1200]
		}
		history = append(history, map[string]any{"role": entry.Role, "content": string(body)})
	}
	output["recentConversation"] = history
	return output, nil
}

func canvasPetNode(n agentworkspace.Node) map[string]any {
	summary := []rune(n.Summary)
	if len(summary) > 1600 {
		summary = summary[:1600]
	}
	return map[string]any{"id": n.ID, "type": n.Type, "title": n.Title, "summary": string(summary), "position": n.Position}
}

func checkCanvasPetPlanRevision(project agentworkspace.Project, planID string, revision int64) error {
	for i := len(project.Messages) - 1; i >= 0; i-- {
		message := project.Messages[i]
		raw, ok := message.Context["directorPlan"]
		if !ok || message.Role != "assistant" {
			continue
		}
		bytes, _ := json.Marshal(raw)
		var plan agentworkspace.DirectorPlan
		if json.Unmarshal(bytes, &plan) != nil || plan.ID != planID {
			continue
		}
		if raw, ok := message.Context["canvasPetRevision"]; ok {
			data, _ := json.Marshal(raw)
			var base int64
			if json.Unmarshal(data, &base) != nil || base != revision || base != project.Canvas.Revision {
				return agentworkspace.ErrRevisionConflict
			}
		}
		return nil
	}
	return nil
}
