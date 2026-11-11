package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/prompt"
)

type AICutTextProviderResolver func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error)

func RunAICutOralClean(ctx context.Context, store aicut.Store, resolver AICutTextProviderResolver, now func() time.Time, owner, projectID, modelID string, maxPauseMS, expectedRevision int64) (aicut.Timeline, aicut.FillerTrimSummary, string, error) {
	if maxPauseMS == 0 {
		maxPauseMS = 1000
	}
	if maxPauseMS < 100 || maxPauseMS > 10000 {
		return aicut.Timeline{}, aicut.FillerTrimSummary{}, "", errors.New("停顿上限必须在 0.1 到 10 秒之间")
	}
	current, err := store.Timeline(ctx, owner, projectID)
	if err != nil {
		return aicut.Timeline{}, aicut.FillerTrimSummary{}, "", err
	}
	if current.Revision != expectedRevision {
		return current, aicut.FillerTrimSummary{}, "", aicut.ErrRevisionConflict
	}
	captions := make([]prompt.AICutCaption, 0)
	for _, track := range current.Tracks {
		if track.Kind != aicut.TrackSubtitle || track.Locked {
			continue
		}
		for _, clip := range track.Clips {
			if clip.Kind == aicut.ClipSubtitle && strings.TrimSpace(clip.Text) != "" {
				captions = append(captions, prompt.AICutCaption{ClipID: clip.ID, StartMS: clip.StartMS, EndMS: clip.StartMS + clip.DurationMS, Text: clip.Text})
			}
		}
	}
	contract, err := prompt.AICutFillerContract(captions)
	if err != nil {
		return aicut.Timeline{}, aicut.FillerTrimSummary{}, "", err
	}
	if resolver == nil {
		return aicut.Timeline{}, aicut.FillerTrimSummary{}, "", errors.New("文本模型未配置，请先到 API 配置启用文本模型")
	}
	provider, err := resolver(ctx, owner, strings.TrimSpace(modelID))
	if err != nil || provider == nil {
		return aicut.Timeline{}, aicut.FillerTrimSummary{}, "", errors.New("文本模型未配置，请先到 API 配置启用文本模型")
	}
	modelResult, err := provider.Complete(ctx, batchfactoryv11.TextCompletionRequest{SystemPrompt: contract.SystemPrompt, UserPrompt: contract.UserPrompt, Temperature: contract.Temperature, MaxTokens: contract.MaxTokens})
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return aicut.Timeline{}, aicut.FillerTrimSummary{}, "", errors.New("去口误停顿已取消")
	}
	if err != nil {
		return aicut.Timeline{}, aicut.FillerTrimSummary{}, "", errors.New("去口误停顿失败：" + err.Error())
	}
	candidates, err := aicut.ParseFillerAnalysis(modelResult)
	if err != nil {
		return aicut.Timeline{}, aicut.FillerTrimSummary{}, "", errors.New(strings.TrimSpace(strings.TrimPrefix(err.Error(), aicut.ErrInvalidFillerTrim.Error()+":")))
	}
	prepared, summary, err := aicut.ApplyOralClean(current, candidates, maxPauseMS)
	if err != nil {
		return aicut.Timeline{}, aicut.FillerTrimSummary{}, "", errors.New(strings.TrimSpace(strings.TrimPrefix(err.Error(), aicut.ErrInvalidFillerTrim.Error()+":")))
	}
	if summary.RemovedMS == 0 {
		return current, summary, "没有发现需要处理的口误或过长停顿", nil
	}
	if now == nil {
		now = time.Now
	}
	saved, err := store.SaveTimeline(ctx, owner, projectID, prepared, expectedRevision, now().UTC())
	if err != nil {
		return aicut.Timeline{}, aicut.FillerTrimSummary{}, "", err
	}
	return saved, summary, "去口误停顿完成", nil
}

func RegisterAICutAIRoutes(mux *http.ServeMux, auth BridgeAuth, store aicut.Store, resolver AICutTextProviderResolver, now func() time.Time) {
	if mux == nil || store == nil {
		return
	}
	if now == nil {
		now = time.Now
	}
	mux.Handle("POST /api/ai-cut/projects/{id}/captions/correct", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			Script           string                    `json:"script"`
			ModelID          string                    `json:"modelId"`
			ExpectedRevision int64                     `json:"expectedRevision"`
			Preview          bool                      `json:"preview"`
			Corrections      []aicut.CaptionCorrection `json:"corrections"`
		}
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		if len(input.Corrections) == 0 && strings.TrimSpace(input.Script) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "校对剧本不能为空"})
			return
		}
		current, err := store.Timeline(req.Context(), identity.Username, req.PathValue("id"))
		if err != nil {
			writeAICutError(w, err)
			return
		}
		if current.Revision != input.ExpectedRevision {
			writeAICutError(w, aicut.ErrRevisionConflict)
			return
		}
		if len(input.Corrections) > 0 {
			prepared, applied, err := aicut.ApplyCaptionCorrections(current, input.Corrections)
			if err != nil {
				writeJSON(w, http.StatusBadGateway, map[string]string{"error": strings.TrimSpace(strings.TrimPrefix(err.Error(), aicut.ErrInvalidCaptionCorrection.Error()+":"))})
				return
			}
			saved, err := store.SaveTimeline(req.Context(), identity.Username, req.PathValue("id"), prepared, input.ExpectedRevision, now().UTC())
			if err != nil {
				writeAICutError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"timeline": saved, "applied": applied, "message": "字幕纠错已写回"})
			return
		}
		captions := make([]prompt.AICutCaption, 0)
		for _, track := range current.Tracks {
			if track.Kind != aicut.TrackSubtitle || track.Locked {
				continue
			}
			for _, clip := range track.Clips {
				if clip.Kind == aicut.ClipSubtitle && strings.TrimSpace(clip.Text) != "" {
					captions = append(captions, prompt.AICutCaption{ClipID: clip.ID, StartMS: clip.StartMS, EndMS: clip.StartMS + clip.DurationMS, Text: clip.Text})
				}
			}
		}
		contract, err := prompt.AICutCaptionCorrectionContract(input.Script, captions)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if resolver == nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "文本模型未配置，请先到 API 配置启用文本模型"})
			return
		}
		provider, err := resolver(req.Context(), identity.Username, strings.TrimSpace(input.ModelID))
		if err != nil || provider == nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "文本模型未配置，请先到 API 配置启用文本模型"})
			return
		}
		result, err := provider.Complete(req.Context(), batchfactoryv11.TextCompletionRequest{
			SystemPrompt: contract.SystemPrompt, UserPrompt: contract.UserPrompt,
			Temperature: contract.Temperature, MaxTokens: contract.MaxTokens,
		})
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			writeJSON(w, http.StatusRequestTimeout, map[string]string{"error": "字幕纠错已取消"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "字幕纠错失败：" + err.Error()})
			return
		}
		corrections, err := aicut.ParseCaptionCorrections(result)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": strings.TrimSpace(strings.TrimPrefix(err.Error(), aicut.ErrInvalidCaptionCorrection.Error()+":"))})
			return
		}
		if input.Preview {
			if len(corrections) == 0 {
				writeJSON(w, http.StatusOK, map[string]any{"preview": []any{}, "applied": 0, "revision": current.Revision, "message": "分析完成"})
				return
			}
			_, applied, err := aicut.ApplyCaptionCorrections(current, corrections)
			if err != nil {
				writeJSON(w, http.StatusBadGateway, map[string]string{"error": strings.TrimSpace(strings.TrimPrefix(err.Error(), aicut.ErrInvalidCaptionCorrection.Error()+":"))})
				return
			}
			clips := aicutSubtitleClipsByID(current)
			preview := make([]map[string]any, 0, len(corrections))
			for _, correction := range corrections {
				clip, ok := clips[correction.ClipID]
				if !ok {
					continue
				}
				preview = append(preview, map[string]any{
					"clipId": clip.ID, "startMs": clip.StartMS, "endMs": clip.StartMS + clip.DurationMS,
					"oldText": clip.Text, "newText": correction.Text, "reason": correction.Reason,
				})
			}
			writeJSON(w, http.StatusOK, map[string]any{"preview": preview, "applied": applied, "revision": current.Revision, "message": "分析完成"})
			return
		}
		if len(corrections) == 0 {
			writeJSON(w, http.StatusOK, map[string]any{"timeline": current, "applied": 0, "message": "没有发现需要修正的字幕"})
			return
		}
		prepared, applied, err := aicut.ApplyCaptionCorrections(current, corrections)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": strings.TrimSpace(strings.TrimPrefix(err.Error(), aicut.ErrInvalidCaptionCorrection.Error()+":"))})
			return
		}
		saved, err := store.SaveTimeline(req.Context(), identity.Username, req.PathValue("id"), prepared, input.ExpectedRevision, now().UTC())
		if err != nil {
			writeAICutError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"timeline": saved, "applied": applied, "message": "字幕纠错完成"})
	})))
	mux.Handle("POST /api/ai-cut/projects/{id}/filler-trim", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			ModelID          string                  `json:"modelId"`
			ExpectedRevision int64                   `json:"expectedRevision"`
			Preview          bool                    `json:"preview"`
			Fillers          []aicut.FillerCandidate `json:"fillers"`
		}
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		current, err := store.Timeline(req.Context(), identity.Username, req.PathValue("id"))
		if err != nil {
			writeAICutError(w, err)
			return
		}
		if current.Revision != input.ExpectedRevision {
			writeAICutError(w, aicut.ErrRevisionConflict)
			return
		}
		if len(input.Fillers) > 0 {
			prepared, summary, err := aicut.ApplyFillerTrim(current, input.Fillers)
			if err != nil {
				writeJSON(w, http.StatusBadGateway, map[string]string{"error": strings.TrimSpace(strings.TrimPrefix(err.Error(), aicut.ErrInvalidFillerTrim.Error()+":"))})
				return
			}
			saved, err := store.SaveTimeline(req.Context(), identity.Username, req.PathValue("id"), prepared, input.ExpectedRevision, now().UTC())
			if err != nil {
				writeAICutError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"timeline": saved, "summary": summary, "message": "水词字幕删除完成"})
			return
		}
		captions := make([]prompt.AICutCaption, 0)
		for _, track := range current.Tracks {
			if track.Kind != aicut.TrackSubtitle || track.Locked {
				continue
			}
			for _, clip := range track.Clips {
				if clip.Kind == aicut.ClipSubtitle && strings.TrimSpace(clip.Text) != "" {
					captions = append(captions, prompt.AICutCaption{ClipID: clip.ID, StartMS: clip.StartMS, EndMS: clip.StartMS + clip.DurationMS, Text: clip.Text})
				}
			}
		}
		contract, err := prompt.AICutFillerContract(captions)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if resolver == nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "文本模型未配置，请先到 API 配置启用文本模型"})
			return
		}
		provider, err := resolver(req.Context(), identity.Username, strings.TrimSpace(input.ModelID))
		if err != nil || provider == nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "文本模型未配置，请先到 API 配置启用文本模型"})
			return
		}
		result, err := provider.Complete(req.Context(), batchfactoryv11.TextCompletionRequest{SystemPrompt: contract.SystemPrompt, UserPrompt: contract.UserPrompt, Temperature: contract.Temperature, MaxTokens: contract.MaxTokens})
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			writeJSON(w, http.StatusRequestTimeout, map[string]string{"error": "去口误停顿已取消"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "去口误停顿失败：" + err.Error()})
			return
		}
		candidates, err := aicut.ParseFillerAnalysis(result)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": strings.TrimSpace(strings.TrimPrefix(err.Error(), aicut.ErrInvalidFillerTrim.Error()+":"))})
			return
		}
		prepared, summary, err := aicut.ApplyFillerTrim(current, candidates)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": strings.TrimSpace(strings.TrimPrefix(err.Error(), aicut.ErrInvalidFillerTrim.Error()+":"))})
			return
		}
		if input.Preview {
			clips := aicutSubtitleClipsByID(current)
			preview := make([]map[string]any, 0, len(candidates))
			for _, candidate := range candidates {
				clip, ok := clips[candidate.ClipID]
				if !ok {
					continue
				}
				preview = append(preview, map[string]any{
					"clipId": clip.ID, "startMs": clip.StartMS, "endMs": clip.StartMS + clip.DurationMS,
					"text": clip.Text, "reason": candidate.Reason,
				})
			}
			writeJSON(w, http.StatusOK, map[string]any{"preview": preview, "summary": summary, "revision": current.Revision, "message": "分析完成"})
			return
		}
		if summary.RemovedMS == 0 {
			writeJSON(w, http.StatusOK, map[string]any{"timeline": current, "summary": summary, "message": "没有发现水词字幕"})
			return
		}
		saved, err := store.SaveTimeline(req.Context(), identity.Username, req.PathValue("id"), prepared, input.ExpectedRevision, now().UTC())
		if err != nil {
			writeAICutError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"timeline": saved, "summary": summary, "message": "水词字幕删除完成"})
	})))
	mux.Handle("POST /api/ai-cut/projects/{id}/oral-clean", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			ModelID          string `json:"modelId"`
			ExpectedRevision int64  `json:"expectedRevision"`
			MaxPauseMS       int64  `json:"maxPauseMs"`
		}
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		timeline, summary, message, err := RunAICutOralClean(req.Context(), store, resolver, now, identity.Username, req.PathValue("id"), input.ModelID, input.MaxPauseMS, input.ExpectedRevision)
		if err != nil {
			if errors.Is(err, aicut.ErrRevisionConflict) || errors.Is(err, aicut.ErrProjectNotFound) {
				writeAICutError(w, err)
				return
			}
			status := http.StatusBadGateway
			if err.Error() == "停顿上限必须在 0.1 到 10 秒之间" || strings.Contains(err.Error(), "字幕") {
				status = http.StatusBadRequest
			} else if strings.Contains(err.Error(), "文本模型未配置") {
				status = http.StatusUnprocessableEntity
			} else if strings.Contains(err.Error(), "已取消") {
				status = http.StatusRequestTimeout
			}
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"timeline": timeline, "summary": summary, "message": message})
	})))
}

func aicutSubtitleClipsByID(timeline aicut.Timeline) map[string]aicut.Clip {
	clips := make(map[string]aicut.Clip)
	for _, track := range timeline.Tracks {
		if track.Kind != aicut.TrackSubtitle {
			continue
		}
		for _, clip := range track.Clips {
			if clip.Kind == aicut.ClipSubtitle {
				clips[clip.ID] = clip
			}
		}
	}
	return clips
}
