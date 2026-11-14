package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/safehttp"
	"qiantie/backend/internal/screenplay"
	"strconv"
	"strings"
	"time"
)

type screenplayVoiceDesignInput struct{ Prompt, PreviewText, RequestKey string }
type screenplayVoiceDesignResult struct {
	VoiceID, Engine, TargetModel, ContentType string
	Audio                                     []byte
}
type screenplayVoiceDesigner func(context.Context, string, string, screenplayVoiceDesignInput) (screenplayVoiceDesignResult, error)
type screenplayVoiceDesignTarget struct{ Engine, Endpoint, Credential, TargetModel string }

// voiceDesignModelOptions uses the same owner configuration and executor checks as generation.
// Normal TTS models are deliberately not returned as voice design services.
func (s *ScreenplayMediaWorkbench) voiceDesignModelOptions(ctx context.Context, owner string) []map[string]string {
	rows := []map[string]string{}
	if s.VoiceDesigner == nil || s.Objects == nil || s.Cut == nil {
		return rows
	}
	for _, engine := range []string{"minimax", "qwen-tts"} {
		target, err := resolveScreenplayVoiceDesignTarget(ctx, s.Catalog, owner, engine)
		if err != nil {
			continue
		}
		label := map[string]string{"minimax": "MiniMax 音色设计", "qwen-tts": "Qwen 音色设计"}[engine]
		if target.TargetModel != "" {
			label += " · " + target.TargetModel
		}
		rows = append(rows, map[string]string{"id": engine, "label": label})
	}
	return rows
}

// Only configured official audio services are used; no request-supplied URL or credential is accepted.
func resolveScreenplayVoiceDesignTarget(ctx context.Context, catalog novelfetchworkshop.Store, owner, engine string) (screenplayVoiceDesignTarget, error) {
	var zero screenplayVoiceDesignTarget
	if engine != "minimax" && engine != "qwen-tts" {
		return zero, fmt.Errorf("未知音色设计模型%s", engine)
	}
	if catalog == nil {
		return zero, fmt.Errorf("%s音色设计服务未配置", engine)
	}
	config, e := catalog.GetConfig(ctx, owner)
	if e != nil {
		return zero, fmt.Errorf("读取当前账号音色设计配置失败")
	}
	for _, m := range modelCatalog(config) {
		kind := modelString(m, "kind")
		if (kind != "audio" && kind != "voice-design") || m["enabled"] != true {
			continue
		}
		endpoint := firstText(modelString(m, "endpoint"), modelString(m, "baseUrl"))
		u, e := url.Parse(endpoint)
		if e != nil || u.Scheme != "https" || u.User != nil || (u.Port() != "" && u.Port() != "443") {
			continue
		}
		host := strings.ToLower(u.Hostname())
		target := screenplayVoiceDesignTarget{Engine: engine, Credential: shuihuoModelCredential(m), TargetModel: modelString(m, "modelId")}
		if target.Credential == "" {
			continue
		}
		if engine == "minimax" {
			if host != "api.minimax.cn" && host != "api.minimax.io" && host != "api.minimaxi.com" {
				continue
			}
			u.Path = "/v1/voice_design"
		} else {
			if host != "dashscope.aliyuncs.com" && host != "dashscope-intl.aliyuncs.com" && !strings.HasSuffix(host, ".cn-beijing.maas.aliyuncs.com") {
				continue
			}
			target.TargetModel = firstText(modelString(m, "voiceDesignTargetModel"), modelString(m, "targetModel"), target.TargetModel)
			if !strings.HasPrefix(target.TargetModel, "qwen-audio-3.1-tts-") && !strings.HasPrefix(target.TargetModel, "cosyvoice-v3") {
				continue
			}
			u.Path = "/api/v1/services/audio/tts/customization"
		}
		u.RawPath = ""
		u.RawQuery = ""
		u.Fragment = ""
		target.Endpoint = u.String()
		return target, nil
	}
	return zero, fmt.Errorf("%s音色设计服务未配置：请在当前账号配置启用的官方音频服务及凭据%s", engine, map[bool]string{true: "，Qwen还需支持音色设计的target_model", false: ""}[engine == "qwen-tts"])
}

func newScreenplayVoiceDesigner(catalog novelfetchworkshop.Store) screenplayVoiceDesigner {
	return func(ctx context.Context, owner, engine string, in screenplayVoiceDesignInput) (screenplayVoiceDesignResult, error) {
		target, e := resolveScreenplayVoiceDesignTarget(ctx, catalog, owner, engine)
		if e != nil {
			return screenplayVoiceDesignResult{}, e
		}
		return requestScreenplayVoiceDesign(ctx, target, in, safehttp.NewClientWithTimeout(120*time.Second))
	}
}

// Official protocol references verified 2026-11-14:
// https://platform.minimax.cn/docs/api-reference/voice-design-design
// https://www.alibabacloud.com/help/en/model-studio/voice-design-api-references
// Qwen uses the current voice-enrollment/create_voice protocol, not legacy qwen-voice-design/create.
func requestScreenplayVoiceDesign(ctx context.Context, target screenplayVoiceDesignTarget, in screenplayVoiceDesignInput, client *http.Client) (screenplayVoiceDesignResult, error) {
	zero := screenplayVoiceDesignResult{}
	if strings.TrimSpace(in.Prompt) == "" || strings.TrimSpace(in.PreviewText) == "" {
		return zero, fmt.Errorf("音色描述和试听文本不能为空")
	}
	if len([]rune(in.Prompt)) > 10000 {
		return zero, fmt.Errorf("音色描述过长")
	}
	payload := map[string]any{}
	if target.Engine == "minimax" {
		if len([]rune(in.PreviewText)) > 500 {
			return zero, fmt.Errorf("Minimax试听文本最多500字")
		}
		payload = map[string]any{"prompt": in.Prompt, "preview_text": in.PreviewText}
	} else if target.Engine == "qwen-tts" {
		if len([]rune(in.Prompt)) > 500 || len([]rune(in.PreviewText)) < 15 || len([]rune(in.PreviewText)) > 200 {
			return zero, fmt.Errorf("Qwen音色描述最多500字，试听文本须15到200字")
		}
		sum := fmt.Sprintf("%x", sha256.Sum256([]byte(in.RequestKey)))
		payload = map[string]any{"model": "voice-enrollment", "input": map[string]any{"action": "create_voice", "target_model": target.TargetModel, "voice_prompt": in.Prompt, "preview_text": in.PreviewText, "prefix": "sp" + sum[:8], "language_hints": []string{"zh"}}, "parameters": map[string]any{"sample_rate": 24000, "response_format": "wav"}}
	} else {
		return zero, fmt.Errorf("未知音色设计模型")
	}
	raw, e := json.Marshal(payload)
	if e != nil {
		return zero, e
	}
	request, e := http.NewRequestWithContext(ctx, "POST", target.Endpoint, bytes.NewReader(raw))
	if e != nil {
		return zero, fmt.Errorf("音色设计地址无效")
	}
	request.Header.Set("Authorization", "Bearer "+target.Credential)
	request.Header.Set("Content-Type", "application/json")
	if client == nil {
		return zero, fmt.Errorf("音色设计HTTP服务未配置")
	}
	owned := *client
	owned.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	response, e := owned.Do(request)
	if e != nil {
		return zero, fmt.Errorf("音色设计请求失败: %w", e)
	}
	defer response.Body.Close()
	raw, e = io.ReadAll(io.LimitReader(response.Body, (24<<20)+1))
	if e != nil || len(raw) > 24<<20 {
		return zero, fmt.Errorf("音色设计返回数据过大或无法读取")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return zero, fmt.Errorf("%s音色设计服务返回HTTP %d", target.Engine, response.StatusCode)
	}
	out := screenplayVoiceDesignResult{Engine: target.Engine, TargetModel: target.TargetModel}
	if target.Engine == "minimax" {
		var decoded struct {
			VoiceID    string `json:"voice_id"`
			TrialAudio string `json:"trial_audio"`
			Base       *struct {
				Code int `json:"status_code"`
			} `json:"base_resp"`
		}
		if e = json.Unmarshal(raw, &decoded); e != nil || decoded.Base == nil || decoded.Base.Code != 0 {
			return zero, fmt.Errorf("Minimax音色设计失败或返回格式无效")
		}
		out.VoiceID = decoded.VoiceID
		out.Audio, e = hex.DecodeString(decoded.TrialAudio)
		out.ContentType = "audio/mpeg"
	} else {
		var decoded struct {
			Output struct {
				VoiceID     string `json:"voice_id"`
				TargetModel string `json:"target_model"`
				Preview     struct {
					Data   string `json:"data"`
					Format string `json:"response_format"`
				} `json:"preview_audio"`
			} `json:"output"`
		}
		if e = json.Unmarshal(raw, &decoded); e != nil {
			return zero, fmt.Errorf("Qwen音色设计返回格式无效")
		}
		if decoded.Output.TargetModel != "" && decoded.Output.TargetModel != target.TargetModel {
			return zero, fmt.Errorf("Qwen返回的音色模型与配置不一致")
		}
		out.VoiceID = decoded.Output.VoiceID
		out.Audio, e = base64.StdEncoding.DecodeString(decoded.Output.Preview.Data)
		out.ContentType = "audio/wav"
		if decoded.Output.Preview.Format != "" && decoded.Output.Preview.Format != "wav" {
			return zero, fmt.Errorf("Qwen返回的试听音频格式与请求不一致")
		}
	}
	if e != nil || strings.TrimSpace(out.VoiceID) == "" || len(out.Audio) == 0 {
		return zero, fmt.Errorf("音色设计没有返回有效音色ID和试听音频")
	}
	if ctx.Err() != nil {
		return zero, ctx.Err()
	}
	return out, nil
}

func (s *ScreenplayMediaWorkbench) designVoice(ctx context.Context, p screenplay.Project, in screenplayMediaAction) (any, error) {
	if s.VoiceDesigner == nil || s.Objects == nil || s.Cut == nil {
		return nil, fmt.Errorf("音色设计服务或素材存储未配置")
	}
	if strings.TrimSpace(in.IdempotencyKey) == "" || len(in.IdempotencyKey) > 256 {
		return nil, fmt.Errorf("请提供有效请求标识")
	}
	st, e := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if e != nil {
		return nil, e
	}
	assets, e := s.Projects.ListAssets(ctx, p.OwnerUsername, p.ID, "")
	if e != nil {
		return nil, e
	}
	name := ""
	if st.VoiceDesign != nil {
		for _, draft := range st.VoiceDesign.Items {
			if in.TargetID == "voice-draft:"+strconv.Itoa(draft.ID) || (draft.AssetID != "" && draft.AssetID == in.TargetID) {
				name = draft.Name
				if strings.TrimSpace(in.Prompt) == "" {
					in.Prompt = draft.Prompt
				}
				if strings.TrimSpace(in.Text) == "" {
					in.Text = draft.PreviewText
				}
				break
			}
		}
	}
	for _, a := range assets {
		if a.ID == in.TargetID && a.Type == screenplay.AssetCharacter {
			name = a.Name
			break
		}
	}
	if name == "" {
		return nil, fmt.Errorf("音色设计须选择当前项目的角色")
	}
	engine := screenplay.EffectiveCreativeConfig(p).VoiceDesignModel
	id := fmt.Sprintf("sp-voice-%x", sha256.Sum256([]byte(engine+":"+in.TargetID+":"+in.IdempotencyKey)))
	for _, old := range st.Items {
		if old.ID == id {
			if old.Status == "succeeded" {
				return mediaItem(st, old.ProviderID)
			}
			return nil, fmt.Errorf("此音色设计请求已提交，状态：%s", old.Status)
		}
	}
	if e = screenplayVoiceAlreadyRunning(st, in.TargetID, engine); e != nil {
		return nil, e
	}
	if st.Revision != in.ExpectedRevision {
		return nil, screenplay.ErrRevisionConflict
	}
	// Resolve availability before creating an attempt or making any provider request.
	if _, e = resolveScreenplayVoiceDesignTarget(ctx, s.Catalog, p.OwnerUsername, engine); e != nil {
		return nil, e
	}
	_, e = s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(st *screenplay.MediaState) error {
		// Recheck under the same storage lock as insertion: changing the request key
		// and refreshing revision cannot submit a second billable request for this row.
		if e := screenplayVoiceAlreadyRunning(*st, in.TargetID, engine); e != nil {
			return e
		}
		if st.Revision != in.ExpectedRevision {
			return screenplay.ErrRevisionConflict
		}
		st.Items = append(st.Items, screenplay.MediaItem{ID: id, Kind: "voice_design", Status: "running", TargetID: in.TargetID, VoiceEngine: engine, CreatedAt: time.Now().UTC()})
		return nil
	})
	if e != nil {
		return nil, e
	}
	result, runErr := s.VoiceDesigner(ctx, p.OwnerUsername, engine, screenplayVoiceDesignInput{Prompt: in.Prompt, PreviewText: in.Text, RequestKey: id})
	if ctx.Err() != nil {
		runErr = ctx.Err()
	}
	if runErr == nil && (strings.TrimSpace(result.VoiceID) == "" || len(result.Audio) == 0 || !strings.HasPrefix(result.ContentType, "audio/")) {
		runErr = fmt.Errorf("音色设计未返回有效音色ID或试听音频")
	}
	var item screenplay.MediaItem
	if runErr == nil {
		item, runErr = s.saveBytes(ctx, p, "audio", name+"音色试听"+aicutTTSExtension(result.ContentType), result.ContentType, bytes.NewReader(result.Audio), in.TargetID)
	}
	if ctx.Err() != nil {
		runErr = ctx.Err()
	}
	_, saveErr := s.Media.MutateMedia(context.WithoutCancel(ctx), p.OwnerUsername, p.ID, func(st *screenplay.MediaState) error {
		if ctx.Err() != nil {
			runErr = ctx.Err()
		}
		if runErr != nil && item.ID != "" {
			kept := st.Items[:0]
			for _, old := range st.Items {
				if old.ID != item.ID {
					kept = append(kept, old)
				}
			}
			st.Items = kept
		}
		for i := range st.Items {
			if st.Items[i].ID == id {
				st.Items[i].Status = "succeeded"
				st.Items[i].ProviderID = item.ID
				if runErr != nil {
					st.Items[i].Status = "failed"
					if ctx.Err() != nil {
						st.Items[i].Status = "cancelled"
					}
					st.Items[i].Error = runErr.Error()
				}
			}
			if item.ID != "" && st.Items[i].ID == item.ID && runErr == nil {
				st.Items[i].VoiceID = result.VoiceID
				st.Items[i].VoiceEngine = result.Engine
				st.Items[i].VoiceTargetModel = result.TargetModel
				st.Items[i].VoiceDescription = in.Prompt
				st.Items[i].VoicePreviewText = in.Text
				item = st.Items[i]
			}
		}
		return nil
	})
	if runErr != nil {
		return nil, runErr
	}
	if saveErr != nil {
		return nil, saveErr
	}
	return item, nil
}

func screenplayVoiceAlreadyRunning(st screenplay.MediaState, target, engine string) error {
	targetKey := func(value string) string {
		if st.VoiceDesign != nil {
			for _, draft := range st.VoiceDesign.Items {
				if value == "voice-draft:"+strconv.Itoa(draft.ID) && draft.AssetID != "" {
					return draft.AssetID
				}
			}
		}
		return value
	}
	for _, item := range st.Items {
		if item.Kind == "voice_design" && item.Status == "running" && targetKey(item.TargetID) == targetKey(target) && (item.VoiceEngine == engine || item.VoiceEngine == "") {
			return fmt.Errorf("%w: 当前角色的%s音色设计正在进行（%s），请等待结束后刷新；此请求未重复提交", screenplay.ErrInvalidState, engine, item.ID)
		}
	}
	return nil
}
