package agentworkspace

import (
	"errors"
	"reflect"
	"testing"
)

func TestResolveCanvasInputsComposesUpstreamTextBeforeNodePrompt(t *testing.T) {
	canvas := Canvas{Nodes: []Node{
		{ID: "text-1", Type: "text", Status: "confirmed", Summary: "雨夜的街道"},
		{ID: "text-2", Type: "text", Status: "confirmed", Config: map[string]any{"prompt": "一个孤独的旅人"}},
		{ID: "image-1", Type: "image", Status: "draft", Config: map[string]any{"prompt": "电影感，冷色调"}},
	}, Edges: []Edge{
		{ID: "edge-1", Source: "text-1", Target: "image-1"},
		{ID: "edge-2", Source: "text-2", Target: "image-1"},
	}}

	inputs, err := ResolveCanvasInputs(canvas, "image-1", CanvasInputCapability{AcceptsTextPrompt: true})
	if err != nil {
		t.Fatalf("ResolveCanvasInputs() error = %v", err)
	}
	if got, want := inputs.Prompt, "雨夜的街道\n一个孤独的旅人\n电影感，冷色调"; got != want {
		t.Fatalf("Prompt = %q, want %q", got, want)
	}
	if got, want := inputs.SourceNodeIDs, []string{"text-1", "text-2", "image-1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SourceNodeIDs = %#v, want %#v", got, want)
	}
}

func TestResolveCanvasInputsUsesTopologicalOrderForChainedText(t *testing.T) {
	canvas := Canvas{Nodes: []Node{
		{ID: "text-1", Type: "text", Status: "confirmed", Summary: "第一段"},
		{ID: "text-2", Type: "text", Status: "confirmed", Summary: "第二段"},
		{ID: "image-1", Type: "image", Status: "draft", Config: map[string]any{"prompt": "生成封面"}},
	}, Edges: []Edge{
		{ID: "edge-1", Source: "text-1", Target: "text-2"},
		{ID: "edge-2", Source: "text-2", Target: "image-1"},
	}}

	inputs, err := ResolveCanvasInputs(canvas, "image-1", CanvasInputCapability{AcceptsTextPrompt: true})
	if err != nil {
		t.Fatalf("ResolveCanvasInputs() error = %v", err)
	}
	if got, want := inputs.Prompt, "第一段\n第二段\n生成封面"; got != want {
		t.Fatalf("Prompt = %q, want %q", got, want)
	}
	if got, want := inputs.SourceNodeIDs, []string{"text-1", "text-2", "image-1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SourceNodeIDs = %#v, want %#v", got, want)
	}
}

func TestResolveCanvasInputsKeepsConfirmedImageReference(t *testing.T) {
	canvas := Canvas{Nodes: []Node{
		{ID: "image-source", Type: "image", Status: "confirmed", AssetRef: map[string]any{"id": "asset-image-1", "kind": "image"}},
		{ID: "image-target", Type: "image", Status: "draft", Config: map[string]any{"prompt": "扩展画面"}},
	}, Edges: []Edge{{ID: "edge-1", Source: "image-source", Target: "image-target"}}}

	inputs, err := ResolveCanvasInputs(canvas, "image-target", CanvasInputCapability{AcceptsTextPrompt: true, AcceptsImageReferences: true})
	if err != nil {
		t.Fatalf("ResolveCanvasInputs() error = %v", err)
	}
	if got, want := inputs.Images, []ResolvedCanvasAsset{{NodeID: "image-source", AssetID: "asset-image-1", Kind: "image"}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Images = %#v, want %#v", got, want)
	}
}

func TestResolveCanvasInputsStopsAtConfirmedMediaBoundary(t *testing.T) {
	canvas := Canvas{Nodes: []Node{
		{ID: "audio-source", Type: "audio", Status: "confirmed", AssetRef: map[string]any{"id": "asset-audio-1", "kind": "audio"}},
		{ID: "image-source", Type: "image", Status: "confirmed", AssetRef: map[string]any{"id": "asset-image-1", "kind": "image"}},
		{ID: "video-target", Type: "video", Status: "draft", Config: map[string]any{"prompt": "小鸟转头"}},
	}, Edges: []Edge{
		{ID: "edge-audio-image", Source: "audio-source", Target: "image-source", Relation: "reference"},
		{ID: "edge-image-video", Source: "image-source", Target: "video-target", Relation: "first_frame"},
	}}

	inputs, err := ResolveCanvasInputs(canvas, "video-target", CanvasInputCapability{AcceptsTextPrompt: true, AcceptsImageReferences: true})
	if err != nil {
		t.Fatalf("ResolveCanvasInputs() error = %v", err)
	}
	if got, want := inputs.Images, []ResolvedCanvasAsset{{NodeID: "image-source", AssetID: "asset-image-1", Kind: "image"}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Images = %#v, want %#v", got, want)
	}
	if len(inputs.Audios) != 0 {
		t.Fatalf("Audios = %#v, want media inputs to stop at the confirmed image", inputs.Audios)
	}
	if got, want := inputs.SourceNodeIDs, []string{"image-source", "video-target"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SourceNodeIDs = %#v, want %#v", got, want)
	}
}

func TestResolveCanvasInputsRejectsUnsupportedOrUnconfirmedMedia(t *testing.T) {
	tests := []struct {
		name    string
		source  Node
		cap     CanvasInputCapability
		wantErr error
	}{
		{
			name:    "audio reference is unsupported",
			source:  Node{ID: "audio-source", Type: "audio", Status: "confirmed", AssetRef: map[string]any{"id": "asset-audio-1", "kind": "audio"}},
			cap:     CanvasInputCapability{AcceptsTextPrompt: true},
			wantErr: ErrCanvasInputUnsupported,
		},
		{
			name:    "unconfirmed image is stale",
			source:  Node{ID: "image-source", Type: "image", Status: "draft", AssetRef: map[string]any{"id": "asset-image-1", "kind": "image"}},
			cap:     CanvasInputCapability{AcceptsTextPrompt: true, AcceptsImageReferences: true},
			wantErr: ErrCanvasInputUnconfirmed,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			canvas := Canvas{Nodes: []Node{test.source, {ID: "video-target", Type: "video", Status: "draft", Config: map[string]any{"prompt": "测试"}}}, Edges: []Edge{{ID: "edge-1", Source: test.source.ID, Target: "video-target"}}}
			_, err := ResolveCanvasInputs(canvas, "video-target", test.cap)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("ResolveCanvasInputs() error = %v, want errors.Is(_, %v)", err, test.wantErr)
			}
		})
	}
}

func TestResolveCanvasInputsIgnoresStaleIncompatibleTypedEdge(t *testing.T) {
	canvas := Canvas{Nodes: []Node{
		{ID: "audio-source", Type: "audio", Status: "confirmed", AssetRef: map[string]any{"id": "asset-audio-1", "kind": "audio"}},
		{ID: "image-target", Type: "image", Status: "draft", Config: map[string]any{"prompt": "生成一只蓝色小鸟"}},
	}, Edges: []Edge{{ID: "legacy-audio-image", Source: "audio-source", Target: "image-target", Relation: "reference"}}}

	inputs, err := ResolveCanvasInputs(canvas, "image-target", CanvasInputCapability{AcceptsTextPrompt: true, AcceptsImageReferences: true})
	if err != nil {
		t.Fatalf("ResolveCanvasInputs() error = %v, want stale audio-to-image edge ignored", err)
	}
	if len(inputs.Audios) != 0 {
		t.Fatalf("Audios = %#v, want no audio input for an image target", inputs.Audios)
	}
	if got, want := inputs.SourceNodeIDs, []string{"image-target"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SourceNodeIDs = %#v, want %#v", got, want)
	}
}

func TestResolveCanvasInputsUsesAudioSourceAsTimingMetadata(t *testing.T) {
	canvas := Canvas{Nodes: []Node{
		{ID: "audio-source", Type: "audio", Status: "confirmed", AssetRef: map[string]any{"id": "asset-audio-1", "kind": "audio"}, Config: map[string]any{"durationSeconds": 7.4}},
		{ID: "video-target", Type: "video", Status: "draft", Config: map[string]any{"prompt": "小鸟转头"}},
	}, Edges: []Edge{{ID: "edge-audio", Source: "audio-source", Target: "video-target", Relation: "audio_source"}}}

	inputs, err := ResolveCanvasInputs(canvas, "video-target", CanvasInputCapability{AcceptsTextPrompt: true, AcceptsImageReferences: true})
	if err != nil {
		t.Fatalf("ResolveCanvasInputs() error = %v", err)
	}
	if len(inputs.Audios) != 0 {
		t.Fatalf("Audios = %#v, want no provider audio references", inputs.Audios)
	}
	if got, want := inputs.SourceNodeIDs, []string{"audio-source", "video-target"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SourceNodeIDs = %#v, want %#v", got, want)
	}
	if len(inputs.Warnings) != 1 || inputs.Warnings[0] != "音频连线仅用于视频时长，不会上传给当前视频模型" {
		t.Fatalf("Warnings = %#v", inputs.Warnings)
	}
}
