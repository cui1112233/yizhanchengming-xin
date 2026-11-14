package screenplay

import (
	"context"
	"encoding/json"
	"fmt"
	"qiantie/backend/internal/prompt"
	"regexp"
	"sort"
	"strings"
)

func (s *StoryboardService) storyboardAuthoringState(ctx context.Context, p Project) (*AuthoringState, error) {
	authors := s.Authoring
	if authors == nil {
		if factory, ok := s.Store.(interface{ AuthoringRepository() AuthoringRepository }); ok {
			authors = &AuthoringService{Store: s.Store, Repository: factory.AuthoringRepository()}
		}
	}
	if authors == nil {
		return nil, nil
	}
	state, e := authors.Load(ctx, p.OwnerUsername, p.ID)
	return &state, e
}
func (s *StoryboardService) storyboardChapterInput(req StoryboardRequest, state *AuthoringState) (StoryboardRequest, error) {
	if req.ChapterID == "" && !req.UseOriginalContentDirectly {
		return req, nil
	}
	if state == nil {
		return req, fmt.Errorf("%w: 章节仓库未配置", ErrInvalidArgument)
	}
	if req.ChapterID == "" && req.UseOriginalContentDirectly && req.EpisodeNo <= 0 {
		originals := map[int]string{}
		for i, chapter := range state.Chapters {
			episode := i + 1
			if len(req.Episodes) > 0 {
				if _, selected := req.Episodes[episode]; !selected {
					continue
				}
			}
			if strings.TrimSpace(chapter.SourceText) == "" {
				return req, fmt.Errorf("%w: 第%d个章节原文为空", ErrEmptyContent, episode)
			}
			originals[episode] = chapter.SourceText
		}
		if len(originals) == 0 {
			return req, ErrEmptyContent
		}
		if len(req.Episodes) > 0 && len(originals) != len(req.Episodes) {
			return req, ErrNotFound
		}
		req.Episodes = originals
		req.Content = ""
		return req, nil
	}
	index := -1
	for i, c := range state.Chapters {
		if c.ID == req.ChapterID || (req.ChapterID == "" && i+1 == req.EpisodeNo) {
			index = i
			break
		}
	}
	if index < 0 {
		return req, ErrNotFound
	}
	chapter := state.Chapters[index]
	content := chapter.ScriptText
	if req.UseOriginalContentDirectly {
		content = chapter.SourceText
	}
	if strings.TrimSpace(content) == "" {
		return req, ErrEmptyContent
	}
	req.Content = content
	req.Episodes = nil
	req.ChapterID = chapter.ID
	if req.EpisodeNo <= 0 {
		req.EpisodeNo = index + 1
	}
	return req, nil
}
func storyboardPlanningText(req StoryboardRequest, state *AuthoringState, episode int) (string, error) {
	if req.UsePlanning != nil && !*req.UsePlanning {
		return "", nil
	}
	if state != nil {
		for i, c := range state.Chapters {
			match := false
			if req.PlanningChapterID != "" {
				match = c.ID == req.PlanningChapterID
			} else if req.ChapterID != "" {
				match = c.ID == req.ChapterID
			} else {
				match = i+1 == episode
			}
			if !match {
				continue
			}
			if strings.TrimSpace(c.PlanningText) == "" {
				break
			}
			source := c.ScriptText
			if c.PlanningUseOriginal {
				source = c.SourceText
			}
			if c.PlanningSourceSHA256 != "" && c.PlanningSourceSHA256 != contentDigest(source) {
				return "", fmt.Errorf("%w: 规划来源已修改，请重新规划或选择不带规划生成", ErrInvalidState)
			}
			return c.PlanningText, nil
		}
	}
	if req.UsePlanning != nil && *req.UsePlanning {
		return "", fmt.Errorf("%w: 所选章节没有已保存的规划", ErrInvalidArgument)
	}
	return "", nil
}
func (s *StoryboardService) storyboardHistory(ctx context.Context, req StoryboardRequest, engine StoryboardEngine, episode int) (string, error) {
	boards, e := s.Store.ListStoryboards(ctx, req.Owner, req.ProjectID, engine)
	if e != nil {
		return "", e
	}
	sort.SliceStable(boards, func(i, j int) bool { return boards[i].EpisodeNo < boards[j].EpisodeNo })
	history := []StoryboardShot{}
	for _, board := range boards {
		if board.EpisodeNo < episode {
			ordered := append([]StoryboardShot{}, board.Shots...)
			sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Seq < ordered[j].Seq })
			history = append(history, ordered...)
		}
	}
	if len(history) > 5 {
		history = history[len(history)-5:]
	}
	if len(history) == 0 {
		return "", nil
	}
	rows := []map[string]any{}
	for _, shot := range history {
		rows = append(rows, map[string]any{"characters": shot.Characters, "scene": shot.Scene, "props": shot.Props, "video_desc": shot.VideoDesc, "duration_sec": shot.Duration})
	}
	raw, e := json.Marshal(rows)
	return string(raw), e
}

var storyboardAssetNameSeparators = regexp.MustCompile(`[,，、;；\n|]`)

func storyboardAssetNames(value string) []string {
	out := []string{}
	for _, part := range storyboardAssetNameSeparators.Split(value, -1) {
		name := strings.TrimSpace(part)
		if name != "" && name != "无" && name != "无道具" && name != "none" && name != "-" {
			out = append(out, name)
		}
	}
	return out
}
func (s *StoryboardService) syncStoryboardMissingAssets(ctx context.Context, p Project, req StoryboardRequest, shots []StoryboardShot) ([]Asset, error) {
	if req.generationSession != nil {
		if ep := req.generationSession.attempt.Episodes[req.EpisodeNo]; ep != nil && len(ep.PendingAssets) > 0 {
			return s.persistStoryboardAssetPlan(ctx, req, ep.PendingAssets)
		}
	}
	current, e := s.Store.ListAssets(ctx, req.Owner, req.ProjectID, "")
	if e != nil {
		return nil, e
	}
	known := map[AssetType]map[string]bool{AssetCharacter: {}, AssetScene: {}, AssetProp: {}}
	for _, a := range current {
		if names, ok := known[a.Type]; ok {
			names[a.Name] = true
		}
	}
	missing := map[AssetType][]string{AssetCharacter: {}, AssetScene: {}, AssetProp: {}}
	for _, shot := range shots {
		for kind, field := range map[AssetType]string{AssetCharacter: shot.Characters, AssetScene: shot.Scene, AssetProp: shot.Props} {
			for _, name := range storyboardAssetNames(field) {
				if !known[kind][name] {
					known[kind][name] = true
					missing[kind] = append(missing[kind], name)
				}
			}
		}
	}
	count := len(missing[AssetCharacter]) + len(missing[AssetScene]) + len(missing[AssetProp])
	if count == 0 {
		return []Asset{}, nil
	}
	raw, _ := json.Marshal(map[string]any{"missing_names": missing, "storyboard": storyboardSemanticRows(shots)})
	system := prompt.ScreenplayStoryboardMissingAssets
	if EffectiveCreativeConfig(p).AssetPromptType == "western" {
		system += "\n\n" + prompt.ScreenplayWesternAssetExtraction
	}
	response, e := s.generationCall(ctx, req, req.EpisodeNo, "asset_sync", TextRequest{Model: firstNonEmpty(req.Model, p.ModelID, "gpt-4o-mini"), SystemPrompt: system, UserPrompt: string(raw), JSONResponse: true, MaxOutputTokens: 12000, Temperature: .2})
	if e != nil {
		return nil, e
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	var descriptions struct {
		Characters []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"characters"`
	}
	_ = json.Unmarshal([]byte(response.Content), &descriptions)
	pending := []Asset{}
	for _, kind := range []AssetType{AssetCharacter, AssetScene, AssetProp} {
		if len(missing[kind]) == 0 {
			continue
		}
		parsed, e := parseAssetJSON(kind, response.Content)
		if e != nil {
			return nil, e
		}
		requested := map[string]bool{}
		for _, name := range missing[kind] {
			requested[name] = true
		}
		for _, asset := range parsed {
			if kind == AssetCharacter && asset.Description == "" {
				for _, row := range descriptions.Characters {
					if row.Name == asset.Name {
						asset.Description = row.Description
						break
					}
				}
			}
			if !requested[asset.Name] {
				return nil, fmt.Errorf("%w: 补充资产返回未请求的名称%s", ErrArtifactInvalid, asset.Name)
			}
			delete(requested, asset.Name)
			asset.Type = kind
			asset.Degree = p.AssetDegree
			if !validDegree(asset.Degree) {
				asset.Degree = DegreeMaster
			}
			asset.ID = newScreenplayID("asset")
			asset.ProjectID = req.ProjectID
			asset.OwnerUsername = req.Owner
			if strings.TrimSpace(asset.Description+asset.Prompt) == "" {
				return nil, fmt.Errorf("%w: 补充资产描述为空", ErrArtifactInvalid)
			}
			pending = append(pending, asset)
		}
		if len(requested) > 0 {
			return nil, fmt.Errorf("%w: 补充资产遗漏指定名称", ErrArtifactInvalid)
		}
	}
	if req.generationSession != nil {
		if e := req.generationSession.recordAssetPlan(ctx, req.EpisodeNo, pending); e != nil {
			return nil, e
		}
	}
	return s.persistStoryboardAssetPlan(ctx, req, pending)
}

func (s *StoryboardService) persistStoryboardAssetPlan(ctx context.Context, req StoryboardRequest, pending []Asset) ([]Asset, error) {
	current, e := s.Store.ListAssets(ctx, req.Owner, req.ProjectID, "")
	if e != nil {
		return nil, e
	}
	exists := map[string]bool{}
	for _, asset := range current {
		exists[string(asset.Type)+"\x00"+asset.Name] = true
	}
	added := []Asset{}
	for _, asset := range pending {
		if asset.OwnerUsername != req.Owner || asset.ProjectID != req.ProjectID || !validAssetType(asset.Type) {
			return added, ErrInvalidArgument
		}
		if exists[string(asset.Type)+"\x00"+asset.Name] {
			continue
		}
		if ctx.Err() != nil {
			return added, ctx.Err()
		}
		saved, e := s.Store.CreateAsset(ctx, asset)
		if e != nil {
			fresh, listErr := s.Store.ListAssets(ctx, req.Owner, req.ProjectID, asset.Type)
			found := false
			for _, old := range fresh {
				if old.Name == asset.Name {
					found = true
				}
			}
			if listErr == nil && found {
				continue
			}
			return added, e
		}
		added = append(added, saved)
		exists[string(asset.Type)+"\x00"+asset.Name] = true
	}
	return added, nil
}
