package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"qiantie/backend/internal/screenplay"
	"qiantie/backend/internal/subjectupload"
	"regexp"
	"strings"
	"time"
)

type screenplayVideoSubjects struct {
	Prompt    string
	URIs, IDs []string
}

var screenplaySoraToken = regexp.MustCompile(`@([^,|，\t\n\v\f\r \x{00A0}\x{1680}\x{2000}-\x{200A}\x{2028}\x{2029}\x{202F}\x{205F}\x{3000}\x{FEFF}@]+)`)
var screenplaySubjectID = regexp.MustCompile(`^asset-[A-Za-z0-9-]+$`)

// The typed upload snapshot is only an explicit adapter when sora2_info has
// never been set. An explicit empty archive remains empty.
func screenplaySubjectArchive(a screenplay.Asset) string {
	if value, exists := a.Extra["sora2_info"]; exists {
		return value
	}
	var result subjectupload.Result
	if json.Unmarshal([]byte(a.Extra["subject_upload_result"]), &result) == nil && screenplaySubjectID.MatchString(result.AssetID) {
		return result.AssetID
	}
	return ""
}
func screenplaySubjectURI(raw string) string {
	raw = strings.TrimSpace(raw)
	normalize := func(v string) string {
		v = strings.TrimPrefix(strings.TrimSpace(v), "asset://")
		if screenplaySubjectID.MatchString(v) {
			return "asset://" + v
		}
		return ""
	}
	if uri := normalize(raw); uri != "" {
		return uri
	}
	// Original helper accepts a JSON object whose string values contain a
	// valid asset ID. Decode it first: never turn opaque JSON into a URI.
	var object map[string]json.RawMessage
	if json.Unmarshal([]byte(raw), &object) != nil || object == nil {
		return ""
	}
	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	_, _ = decoder.Token()
	for decoder.More() {
		_, err := decoder.Token()
		if err != nil {
			return ""
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil {
			return ""
		}
		var text string
		if json.Unmarshal(value, &text) == nil {
			if uri := normalize(text); uri != "" {
				return uri
			}
		}
	}
	return ""
}

func screenplaySubjectNames(value string) []string {
	return strings.FieldsFunc(value, func(r rune) bool { return strings.ContainsRune(",，、;；/|\n@", r) })
}

// Explicit IDs from the original storyboard fields win over name matching.
func screenplayBoundSubjects(shot screenplay.StoryboardShot, assets []screenplay.Asset, includeProps bool) ([]screenplay.Asset, error) {
	var fields struct {
		Characters []json.RawMessage `json:"character_ids"`
		Scene      json.RawMessage   `json:"scene_id"`
		Props      []json.RawMessage `json:"prop_ids"`
	}
	if len(shot.ProtocolFields) > 0 && json.Unmarshal(shot.ProtocolFields, &fields) != nil {
		return nil, fmt.Errorf("分镜主体绑定格式无效")
	}
	parseID := func(raw json.RawMessage) string {
		var v string
		if json.Unmarshal(raw, &v) == nil {
			return v
		}
		var n json.Number
		if json.Unmarshal(raw, &n) == nil {
			return string(n)
		}
		return ""
	}
	ids := map[screenplay.AssetType][]json.RawMessage{screenplay.AssetCharacter: fields.Characters, screenplay.AssetProp: fields.Props}
	if len(fields.Scene) > 0 && string(fields.Scene) != "null" && string(fields.Scene) != "0" && string(fields.Scene) != `""` {
		ids[screenplay.AssetScene] = []json.RawMessage{fields.Scene}
	}
	names := map[screenplay.AssetType]string{screenplay.AssetCharacter: shot.Characters, screenplay.AssetScene: shot.Scene, screenplay.AssetProp: shot.Props}
	targets := screenplayReferenceTargets(shot, assets)
	allowed := map[string]bool{}
	for _, id := range targets {
		allowed[id] = true
	}
	out := []screenplay.Asset{}
	seen := map[string]bool{}
	for _, kind := range []screenplay.AssetType{screenplay.AssetCharacter, screenplay.AssetScene, screenplay.AssetProp} {
		if kind == screenplay.AssetProp && !includeProps {
			continue
		}
		appendOne := func(a screenplay.Asset) {
			if !seen[a.ID] {
				out = append(out, a)
				seen[a.ID] = true
			}
		}
		if len(ids[kind]) > 0 {
			for _, raw := range ids[kind] {
				id := parseID(raw)
				found := false
				for _, a := range assets {
					if a.ID == id && a.Type == kind {
						appendOne(a)
						found = true
						break
					}
				}
				if !found {
					return nil, fmt.Errorf("找不到当前作品的绑定主体 %s", id)
				}
			}
			continue
		}
		for _, part := range screenplaySubjectNames(names[kind]) {
			name := strings.TrimSpace(part)
			if name == "" || name == "无" || name == "无道具" || name == "none" || name == "-" {
				continue
			}
			var matches []screenplay.Asset
			for _, a := range assets {
				if a.Type == kind && strings.TrimSpace(a.Name) == name {
					matches = append(matches, a)
				}
			}
			if len(matches) != 1 || !allowed[matches[0].ID] {
				return nil, fmt.Errorf("主体 %s 缺失或重名，请明确绑定资产", name)
			}
			appendOne(matches[0])
		}
	}
	return out, nil
}
func (s *ScreenplayMediaWorkbench) VideoSubjects(ctx context.Context, p screenplay.Project, shot screenplay.StoryboardShot) (screenplayVideoSubjects, error) {
	out := screenplayVideoSubjects{Prompt: strings.TrimSpace(shot.VideoDesc)}
	assets, err := s.Projects.ListAssets(ctx, p.OwnerUsername, p.ID, "")
	if err != nil {
		return out, err
	}
	// Store access is owner/project scoped; reject malformed cross-project rows too.
	scoped := assets[:0]
	for _, a := range assets {
		if a.OwnerUsername == p.OwnerUsername && a.ProjectID == p.ID {
			scoped = append(scoped, a)
		}
	}
	subjectMode := p.CreativeConfig != nil && p.CreativeConfig.UseSubjectMode
	bound, err := screenplayBoundSubjects(shot, scoped, subjectMode)
	if err != nil {
		return out, err
	}
	if !subjectMode {
		replacements := map[string]string{}
		for _, a := range bound {
			if v := screenplaySubjectArchive(a); v != "" {
				if _, exists := replacements[a.Name]; !exists {
					replacements[a.Name] = v
				}
			}
		}
		out.Prompt = screenplaySoraToken.ReplaceAllStringFunc(out.Prompt, func(token string) string {
			if value := replacements[strings.TrimPrefix(token, "@")]; value != "" {
				return "@" + value
			}
			return token
		})
		return out, nil
	}
	seen := map[string]bool{}
	for _, a := range bound {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		uri := screenplaySubjectURI(screenplaySubjectArchive(a))
		stable := uri
		if uri == "" {
			if s.Media == nil {
				return out, fmt.Errorf("主体 %s 缺少有效存档或已选图片", a.Name)
			}
			state, e := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
			if e != nil {
				return out, e
			}
			selected := state.Selected[a.ID]
			if selected == "" {
				return out, fmt.Errorf("主体 %s 缺少有效存档或已选图片", a.Name)
			}
			item, e := mediaItem(state, selected)
			if e != nil {
				return out, e
			}
			if item.TargetID != a.ID || item.Kind != "image" {
				return out, fmt.Errorf("主体 %s 已选媒体不是该主体图片", a.Name)
			}
			if s.Cut == nil || s.ReferencesProvider == nil {
				return out, fmt.Errorf("主体 %s 图片签名服务不可用", a.Name)
			}
			mediaAsset, e := s.Cut.Asset(ctx, p.OwnerUsername, p.ID, item.AssetID)
			if e != nil {
				return out, e
			}
			uri, e = s.ReferencesProvider.ReferenceURL(ctx, aicutObjectAsset(mediaAsset), time.Hour)
			if e != nil {
				return out, e
			}
			u, e := url.Parse(uri)
			if e != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
				return out, fmt.Errorf("主体 %s 图片引用不可用", a.Name)
			}
			stable = "media:" + item.AssetID
		}
		if !seen[stable] {
			seen[stable] = true
			out.URIs = append(out.URIs, uri)
			out.IDs = append(out.IDs, stable)
		}
	}
	return out, nil
}
