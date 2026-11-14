package screenplay

import (
	"encoding/json"
	"fmt"
	"strings"

	"qiantie/backend/internal/prompt"
)

// BuildAssetCheckSteps snapshots the exact recovered request contract using the
// project's own configured model. No discovery request writes live assets.
func BuildAssetCheckSteps(project Project, episodes []AssetCheckEpisode, existing []Asset, model string) ([]AssetCheckStep, error) {
	if strings.TrimSpace(model) == "" {
		return nil, ErrInvalidArgument
	}
	names := map[AssetType][]string{}
	for _, asset := range existing {
		names[asset.Type] = append(names[asset.Type], asset.Name)
	}
	steps := make([]AssetCheckStep, 0, len(episodes)*3)
	for _, episode := range episodes {
		for _, kind := range []AssetType{AssetCharacter, AssetScene, AssetProp} {
			text, err := prompt.BuildScreenplayAssetCheckPrompt(string(kind), EffectiveCreativeConfig(project).AssetPromptType, episode.Content, names[AssetCharacter], names[AssetScene], names[AssetProp])
			if err != nil {
				return nil, err
			}
			request := TextRequest{Model: model, SystemPrompt: text.SystemPrompt, UserPrompt: text.UserPrompt, MaxOutputTokens: 32768, Temperature: 0.3, JSONResponse: true}
			digest := generationFingerprint(map[string]any{"request": request}, "", "", "")
			steps = append(steps, AssetCheckStep{ID: fmt.Sprintf("%d-%s-%s", episode.ChapterIndex, kind, digest[:16]), ChapterIndex: episode.ChapterIndex, EpisodeNo: episode.EpisodeNo, Type: kind, Status: "pending", Request: request, RequestDigest: digest})
		}
	}
	return steps, nil
}

func parseAssetCheckCandidates(job AssetCheckJob, step AssetCheckStep, raw string) ([]AssetCheckCandidate, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, ErrAIEmptyContent
	}
	object, ok := extractJSONObject(raw)
	if !ok {
		return nil, ErrAIInvalidJSON
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal([]byte(object), &envelope); err != nil {
		return nil, fmt.Errorf("%w %v", ErrAIJSONParse, err)
	}
	key := map[AssetType]string{AssetCharacter: "characters", AssetScene: "scenes", AssetProp: "props"}[step.Type]
	if key == "" {
		return nil, ErrInvalidArgument
	}
	value, ok := envelope[key]
	// Original silently treats malformed output as an empty success. CM records a
	// visible failure instead; an explicit [] is the only valid empty collection.
	if !ok || strings.TrimSpace(string(value)) == "null" {
		return nil, ErrAIInvalidJSON
	}
	var items []json.RawMessage
	if err := json.Unmarshal(value, &items); err != nil {
		return nil, fmt.Errorf("%w %v", ErrAIJSONParse, err)
	}
	out := []AssetCheckCandidate{}
	seen := map[string]bool{}
	for _, asset := range job.Existing {
		seen[assetCheckNameKey(asset)] = true
	}
	for _, original := range items {
		var item map[string]any
		if err := json.Unmarshal(original, &item); err != nil || item == nil {
			continue
		}
		kind := step.Type
		if typed := stringField(item, "type"); typed != "" {
			kind = AssetType(typed)
			if kind != AssetCharacter && kind != AssetScene && kind != AssetProp {
				continue
			}
		}
		asset, valid := assetFromMap(kind, item)
		if !valid {
			continue
		}
		asset.Prompt = firstNonEmpty(asset.Prompt, stringField(item, "image_prompt"))
		asset.Description = firstNonEmpty(asset.Description, stringField(item, "description"))
		asset.OwnerUsername = job.Owner
		asset.ProjectID = job.ProjectID
		asset.Degree = job.Degree
		identity := assetCheckNameKey(asset)
		if seen[identity] {
			continue
		}
		seen[identity] = true
		out = append(out, AssetCheckCandidate{ID: contentDigest(step.ID + "\x00" + identity), Asset: asset, Original: append(json.RawMessage{}, original...)})
	}
	return out, nil
}
