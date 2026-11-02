package agentworkspace

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrCanvasInputTargetNotFound = errors.New("canvas input target node not found")
	ErrCanvasInputUnsupported    = errors.New("canvas input is not supported by the selected model")
	ErrCanvasInputUnconfirmed    = errors.New("canvas input asset is not confirmed")
)

// CanvasInputCapability is the server-owned contract for what one selected
// model can consume. Unknown capabilities remain false so saved canvas edges
// cannot silently become model inputs.
type CanvasInputCapability struct {
	AcceptsTextPrompt      bool
	AcceptsImageReferences bool
	AcceptsVideoReferences bool
	AcceptsAudioReferences bool
	MaxReferenceAssets     int
}

type ResolvedCanvasAsset struct {
	NodeID  string `json:"nodeId"`
	AssetID string `json:"assetId"`
	Kind    string `json:"kind"`
}

type ResolvedCanvasInputs struct {
	Prompt        string                `json:"prompt"`
	Images        []ResolvedCanvasAsset `json:"images"`
	Videos        []ResolvedCanvasAsset `json:"videos"`
	Audios        []ResolvedCanvasAsset `json:"audios"`
	SourceNodeIDs []string              `json:"sourceNodeIds"`
	Warnings      []string              `json:"warnings"`
}

// ResolveCanvasInputs turns incoming canvas edges into one deterministic model
// request. It never guesses about media support: unsupported or unconfirmed
// inputs stop submission before a provider receives a partial request.
func ResolveCanvasInputs(canvas Canvas, targetNodeID string, capability CanvasInputCapability) (ResolvedCanvasInputs, error) {
	targetNodeID = strings.TrimSpace(targetNodeID)
	nodes := make(map[string]Node, len(canvas.Nodes))
	for _, node := range canvas.Nodes {
		nodes[node.ID] = node
	}
	target, found := nodes[targetNodeID]
	if !found {
		return ResolvedCanvasInputs{}, ErrCanvasInputTargetNotFound
	}

	inputs := ResolvedCanvasInputs{Images: []ResolvedCanvasAsset{}, Videos: []ResolvedCanvasAsset{}, Audios: []ResolvedCanvasAsset{}, SourceNodeIDs: []string{}, Warnings: []string{}}
	promptParts := make([]string, 0)
	seenSources := make(map[string]struct{}, len(canvas.Nodes))
	seenAssets := make(map[string]struct{}, len(canvas.Nodes))
	appendSource := func(id string) {
		if _, exists := seenSources[id]; exists {
			return
		}
		seenSources[id] = struct{}{}
		inputs.SourceNodeIDs = append(inputs.SourceNodeIDs, id)
	}

	incoming := make(map[string][]Edge, len(canvas.Nodes))
	timingAudioSources := make(map[string]struct{})
	for _, edge := range canvas.Edges {
		if source, sourceExists := nodes[edge.Source]; sourceExists {
			if target, targetExists := nodes[edge.Target]; targetExists && !canvasEdgeCanProvideInput(source.Type, target.Type) {
				if edge.Target == targetNodeID {
					inputs.Warnings = append(inputs.Warnings, fmt.Sprintf("已忽略不能连接到%s节点的%s素材 %s", canvasInputKindName(target.Type), canvasInputKindName(source.Type), source.ID))
				}
				continue
			}
		}
		incoming[edge.Target] = append(incoming[edge.Target], edge)
		if edge.Target == targetNodeID && strings.TrimSpace(edge.Relation) == "audio_source" {
			timingAudioSources[edge.Source] = struct{}{}
		}
	}
	visiting := make(map[string]struct{}, len(canvas.Nodes))
	visited := make(map[string]struct{}, len(canvas.Nodes))
	var visit func(string) error
	visit = func(nodeID string) error {
		if _, complete := visited[nodeID]; complete {
			return nil
		}
		if _, cycle := visiting[nodeID]; cycle {
			inputs.Warnings = append(inputs.Warnings, fmt.Sprintf("已忽略包含节点 %s 的循环连接", nodeID))
			return nil
		}
		visiting[nodeID] = struct{}{}
		source, sourceExists := nodes[nodeID]
		isMediaBoundary := nodeID != targetNodeID && sourceExists && (source.Type == "image" || source.Type == "video" || source.Type == "audio")
		if !isMediaBoundary {
			for _, edge := range incoming[nodeID] {
				if _, exists := nodes[edge.Source]; !exists {
					inputs.Warnings = append(inputs.Warnings, fmt.Sprintf("已忽略不存在的上游节点 %s", strings.TrimSpace(edge.Source)))
					continue
				}
				if err := visit(edge.Source); err != nil {
					return err
				}
			}
		}
		delete(visiting, nodeID)
		visited[nodeID] = struct{}{}

		if !sourceExists || nodeID == targetNodeID {
			return nil
		}
		switch strings.TrimSpace(source.Type) {
		case "text":
			if !capability.AcceptsTextPrompt {
				return fmt.Errorf("%w: 当前模型不支持文字提示词", ErrCanvasInputUnsupported)
			}
			if text := canvasInputText(source); text != "" {
				promptParts = append(promptParts, text)
				appendSource(source.ID)
			}
		case "image", "video", "audio":
			if source.Type == "audio" && !capability.AcceptsAudioReferences {
				if _, timingOnly := timingAudioSources[source.ID]; timingOnly {
					appendSource(source.ID)
					inputs.Warnings = append(inputs.Warnings, "音频连线仅用于视频时长，不会上传给当前视频模型")
					return nil
				}
			}
			asset, err := resolveCanvasMediaInput(source, capability)
			if err != nil {
				return err
			}
			assetKey := asset.Kind + "\x00" + asset.AssetID
			if _, duplicate := seenAssets[assetKey]; duplicate {
				return nil
			}
			seenAssets[assetKey] = struct{}{}
			appendSource(source.ID)
			switch asset.Kind {
			case "image":
				inputs.Images = append(inputs.Images, asset)
			case "video":
				inputs.Videos = append(inputs.Videos, asset)
			case "audio":
				inputs.Audios = append(inputs.Audios, asset)
			}
		}
		return nil
	}
	if err := visit(targetNodeID); err != nil {
		return ResolvedCanvasInputs{}, err
	}
	if prompt := canvasInputText(target); prompt != "" {
		promptParts = append(promptParts, prompt)
	}
	appendSource(target.ID)
	inputs.Prompt = strings.Join(promptParts, "\n")
	if capability.MaxReferenceAssets > 0 && len(inputs.Images)+len(inputs.Videos)+len(inputs.Audios) > capability.MaxReferenceAssets {
		return ResolvedCanvasInputs{}, fmt.Errorf("%w: 当前模型最多支持 %d 个参考素材", ErrCanvasInputUnsupported, capability.MaxReferenceAssets)
	}
	return inputs, nil
}

func canvasEdgeCanProvideInput(sourceKind, targetKind string) bool {
	sourceKind = strings.TrimSpace(sourceKind)
	targetKind = strings.TrimSpace(targetKind)
	if sourceKind == "text" {
		return true
	}
	switch targetKind {
	case "image", "director":
		return sourceKind == "image"
	case "video":
		return sourceKind == "image" || sourceKind == "video" || sourceKind == "audio"
	case "audio":
		return sourceKind == "audio"
	case "text":
		return sourceKind == "text"
	default:
		return false
	}
}

func canvasInputText(node Node) string {
	if text := strings.TrimSpace(node.Summary); text != "" {
		return text
	}
	if node.Config == nil {
		return ""
	}
	text, _ := node.Config["prompt"].(string)
	return strings.TrimSpace(text)
}

func resolveCanvasMediaInput(node Node, capability CanvasInputCapability) (ResolvedCanvasAsset, error) {
	kind := strings.TrimSpace(node.Type)
	if !canvasInputCapabilityAccepts(kind, capability) {
		return ResolvedCanvasAsset{}, fmt.Errorf("%w: 当前模型不支持%s参考素材", ErrCanvasInputUnsupported, canvasInputKindName(kind))
	}
	if strings.TrimSpace(node.Status) != "confirmed" {
		return ResolvedCanvasAsset{}, fmt.Errorf("%w: %s 节点尚未确认", ErrCanvasInputUnconfirmed, canvasInputKindName(kind))
	}
	assetID := ""
	assetKind := ""
	if node.AssetRef != nil {
		assetID, _ = node.AssetRef["id"].(string)
		if assetID == "" {
			assetID, _ = node.AssetRef["assetId"].(string)
		}
		assetKind, _ = node.AssetRef["kind"].(string)
	}
	assetID = strings.TrimSpace(assetID)
	assetKind = strings.TrimSpace(assetKind)
	if assetID == "" || assetKind != kind {
		return ResolvedCanvasAsset{}, fmt.Errorf("%w: %s 节点缺少已确认素材", ErrCanvasInputUnconfirmed, canvasInputKindName(kind))
	}
	return ResolvedCanvasAsset{NodeID: node.ID, AssetID: assetID, Kind: kind}, nil
}

func canvasInputCapabilityAccepts(kind string, capability CanvasInputCapability) bool {
	switch kind {
	case "image":
		return capability.AcceptsImageReferences
	case "video":
		return capability.AcceptsVideoReferences
	case "audio":
		return capability.AcceptsAudioReferences
	default:
		return false
	}
}

func canvasInputKindName(kind string) string {
	switch kind {
	case "image":
		return "图片"
	case "video":
		return "视频"
	case "audio":
		return "音频"
	default:
		return kind
	}
}
