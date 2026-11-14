package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/screenplay"
	"strings"
)

func (s *ScreenplayMediaWorkbench) analyzeVoices(ctx context.Context, p screenplay.Project, in screenplayMediaAction) (any, error) {
	if s.PositioningText == nil {
		return nil, fmt.Errorf("音色描述分析服务未配置")
	}
	model := firstText(in.ModelID, p.ModelID)
	if model == "" {
		return nil, fmt.Errorf("请选择文字推理模型")
	}
	original, source, sourceErr := s.voiceDesignOriginal(ctx, p)
	if sourceErr != nil {
		return nil, sourceErr
	}
	if strings.TrimSpace(original) == "" {
		return nil, fmt.Errorf("%w: 音色描述分析需要项目原文", screenplay.ErrEmptyContent)
	}
	if in.IdempotencyKey == "" || len(in.IdempotencyKey) > 256 {
		return nil, fmt.Errorf("请提供有效请求标识")
	}
	key := fmt.Sprintf("voice-analysis-%x", sha256.Sum256([]byte(in.IdempotencyKey)))
	st, e := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if e != nil {
		return nil, e
	}
	if st.VoiceDesign != nil && st.VoiceDesign.RequestKey == key {
		if st.VoiceDesign.Status == "succeeded" {
			return st.VoiceDesign, nil
		}
		return nil, fmt.Errorf("此音色描述分析请求已经提交，状态：%s", st.VoiceDesign.Status)
	}
	if st.Revision != in.ExpectedRevision {
		return nil, screenplay.ErrRevisionConflict
	}
	assets, e := s.Projects.ListAssets(ctx, p.OwnerUsername, p.ID, "")
	if e != nil {
		return nil, e
	}
	reserved, e := s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(st *screenplay.MediaState) error {
		if st.Revision != in.ExpectedRevision {
			return screenplay.ErrRevisionConflict
		}
		if st.VoiceDesign != nil && st.VoiceDesign.Status == "running" {
			return screenplay.ErrInvalidState
		}
		state := screenplay.VoiceDesignState{Items: []screenplay.VoiceDesignDraft{}}
		if st.VoiceDesign != nil {
			state = *st.VoiceDesign
		}
		state.Model = model
		state.Source = source
		state.RequestKey = key
		state.Status = "running"
		state.Error = ""
		st.VoiceDesign = &state
		return nil
	})
	if e != nil {
		return nil, e
	}
	response, runErr := s.PositioningText(ctx, p.OwnerUsername, screenplay.TextRequest{Model: model, SystemPrompt: prompt.ScreenplayVoiceDesignAnalysis127, UserPrompt: original, MaxOutputTokens: 16384, Temperature: .4, JSONResponse: false})
	if ctx.Err() != nil {
		runErr = ctx.Err()
	}
	var items []screenplay.VoiceDesignDraft
	if runErr == nil {
		if e = json.Unmarshal([]byte(strings.TrimSpace(response.Content)), &items); e != nil || items == nil {
			runErr = fmt.Errorf("%w: 音色分析必须返回127约定的纯JSON数组", screenplay.ErrAIInvalidJSON)
		}
	}
	if runErr == nil {
		ids := map[int]bool{}
		names := map[string]bool{}
		for i := range items {
			item := &items[i]
			if item.ID <= 0 || ids[item.ID] || strings.TrimSpace(item.Name) == "" || names[item.Name] || strings.TrimSpace(item.Prompt) == "" || len([]rune(item.PreviewText)) < 60 {
				runErr = fmt.Errorf("%w: 音色条目ID/姓名/描述无效或试听文本少于60字", screenplay.ErrArtifactInvalid)
				break
			}
			ids[item.ID] = true
			names[item.Name] = true
			item.AssetID = ""
			for _, a := range assets {
				if a.Type == screenplay.AssetCharacter && a.Name == item.Name {
					item.AssetID = a.ID
					break
				}
			}
		}
	}
	var result *screenplay.VoiceDesignState
	if runErr == nil {
		fresh, err := s.Projects.GetProject(ctx, p.OwnerUsername, p.ID)
		if err != nil {
			runErr = err
		} else if freshOriginal, freshSource, err := s.voiceDesignOriginal(ctx, fresh); err != nil {
			runErr = err
		} else if freshOriginal != original || freshSource != source {
			runErr = screenplay.ErrRevisionConflict
		}
	}
	_, saveErr := s.Media.MutateMedia(context.WithoutCancel(ctx), p.OwnerUsername, p.ID, func(st *screenplay.MediaState) error {
		if st.VoiceDesign == nil || st.VoiceDesign.RequestKey != key || st.VoiceDesign.Status != "running" {
			return screenplay.ErrRevisionConflict
		}
		if ctx.Err() != nil {
			runErr = ctx.Err()
		}
		if runErr == nil && st.Revision != reserved.Revision {
			runErr = screenplay.ErrRevisionConflict
		}
		if runErr != nil {
			st.VoiceDesign.Status = "failed"
			st.VoiceDesign.Error = runErr.Error()
			if ctx.Err() != nil {
				st.VoiceDesign.Status = "cancelled"
			}
		} else {
			st.VoiceDesign.Items = items
			st.VoiceDesign.Status = "succeeded"
			st.VoiceDesign.Error = ""
		}
		result = st.VoiceDesign
		return nil
	})
	if runErr != nil {
		return nil, runErr
	}
	if saveErr != nil {
		return nil, saveErr
	}
	return result, nil
}

func (s *ScreenplayMediaWorkbench) voiceDesignOriginal(ctx context.Context, p screenplay.Project) (string, string, error) {
	if s.Authoring != nil {
		state, e := s.Authoring.Load(ctx, p.OwnerUsername, p.ID)
		if e != nil {
			return "", "", e
		}
		parts := []string{}
		for _, chapter := range state.Chapters {
			if strings.TrimSpace(chapter.SourceText) != "" {
				parts = append(parts, chapter.Title+"\n"+chapter.SourceText)
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, "\n\n"), "authoring_chapters", nil
		}
	}
	return p.SourceText, "project_original", nil
}

func (s *ScreenplayMediaWorkbench) saveVoiceDesignDrafts(ctx context.Context, p screenplay.Project, in screenplayMediaAction) (any, error) {
	if in.VoiceDesignDrafts == nil || len(in.VoiceDesignDrafts) > 1000 {
		return nil, fmt.Errorf("%w: 请传入有效的items数组（最多1000条）", screenplay.ErrInvalidArgument)
	}
	assets, e := s.Projects.ListAssets(ctx, p.OwnerUsername, p.ID, "")
	if e != nil {
		return nil, e
	}
	ids := map[int]bool{}
	for i := range in.VoiceDesignDrafts {
		draft := &in.VoiceDesignDrafts[i]
		if draft.ID <= 0 || ids[draft.ID] || strings.TrimSpace(draft.Name) == "" || len([]rune(draft.Name)) > 200 || len([]rune(draft.Prompt)) > 10000 || len([]rune(draft.PreviewText)) > 5000 {
			return nil, fmt.Errorf("%w: 音色描述ID/名称无效、重复或文本过长", screenplay.ErrInvalidArgument)
		}
		ids[draft.ID] = true
		draft.AssetID = ""
		for _, asset := range assets {
			if asset.Type == screenplay.AssetCharacter && asset.Name == draft.Name {
				draft.AssetID = asset.ID
				break
			}
		}
	}
	var result *screenplay.VoiceDesignState
	_, e = s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(st *screenplay.MediaState) error {
		if st.Revision != in.ExpectedRevision {
			return screenplay.ErrRevisionConflict
		}
		state := screenplay.VoiceDesignState{Source: "manual"}
		if st.VoiceDesign != nil {
			state = *st.VoiceDesign
		}
		state.Items = in.VoiceDesignDrafts
		state.RequestKey = ""
		state.Status = "draft"
		state.Error = ""
		st.VoiceDesign = &state
		result = st.VoiceDesign
		return nil
	})
	return result, e
}
