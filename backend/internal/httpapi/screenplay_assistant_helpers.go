package httpapi

import (
	"context"
	"strings"

	"qiantie/backend/internal/screenplay"
)

// assistantSuggestion 是模型返回的导航建议原始结构。
type assistantSuggestion struct {
	Label  string `json:"label"`
	Entry  string `json:"entry"`
	Reason string `json:"reason"`
}

// allowedAssistantEntries 是允许导航到的面板标识白名单，与
// screenplay.BuildAssistantCapabilities 的 Entry 保持一致。模型给出的其它
// 标识一律丢弃，避免诱导前端跳到不存在的面板。
var allowedAssistantEntries = map[string]bool{
	"source": true, "script": true, "assets": true, "storyboard": true, "media": true,
}

func sanitizeAssistantSuggestions(in []assistantSuggestion) []screenplay.AssistantNavigate {
	out := make([]screenplay.AssistantNavigate, 0, len(in))
	seen := map[string]bool{}
	for _, s := range in {
		entry := strings.TrimSpace(s.Entry)
		label := strings.TrimSpace(s.Label)
		if !allowedAssistantEntries[entry] || label == "" || seen[entry] {
			continue
		}
		seen[entry] = true
		out = append(out, screenplay.AssistantNavigate{Label: label, Entry: entry, Reason: strings.TrimSpace(s.Reason)})
		if len(out) >= 5 {
			break
		}
	}
	return out
}

// buildAssistantCapabilities 从作品真实状态构造链路能力清单与覆盖度。
// 只读，不触发任何生成；mediaStore 为 nil 时媒体类环节按未就绪处理。
func buildAssistantCapabilities(
	ctx context.Context,
	store screenplay.Store,
	media screenplay.MediaStore,
	owner, projectID string,
	project screenplay.Project,
	st screenplay.AuthoringState,
	assets []screenplay.Asset,
	shots []screenplay.Shot,
	hasDoc bool,
) ([]screenplay.AssistantCapability, screenplay.AssistantCoverage, error) {
	var hasImages, hasVideos, hasSubtitles, hasFinalCut bool
	if media != nil {
		ms, err := media.LoadMedia(ctx, owner, projectID)
		if err == nil {
			for _, item := range ms.Items {
				if item.Status != "succeeded" && item.Status != "ready" && item.Status != "done" {
					continue
				}
				switch item.Kind {
				case "image":
					hasImages = true
				case "video":
					hasVideos = true
				case "subtitle", "subtitles":
					hasSubtitles = true
				case "final", "final_cut", "merged":
					hasFinalCut = true
				}
			}
		}
	}
	input := screenplay.AssistantCapabilitiesInput{
		HasSource:     strings.TrimSpace(project.SourceText) != "",
		HasSynopsis:   strings.TrimSpace(project.Synopsis) != "" || (project.CreativeConfig != nil && project.CreativeConfig.SynopsisResult != nil),
		HasChapters:   len(st.Chapters) > 0,
		HasScript:     hasDoc,
		HasAssets:     len(assets) > 0,
		HasStoryboard: len(shots) > 0,
		HasImages:     hasImages,
		HasVideos:     hasVideos,
		HasSubtitles:  hasSubtitles,
		HasFinalCut:   hasFinalCut,
	}
	caps := screenplay.BuildAssistantCapabilities(input)
	return caps, screenplay.BuildAssistantCoverage(caps), nil
}
