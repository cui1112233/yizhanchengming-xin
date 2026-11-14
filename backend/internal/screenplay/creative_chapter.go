package screenplay

import (
	"context"
	"encoding/json"
	"fmt"
	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/prompt/luoshui"
	"strings"
)

func BuildChapterTextRequest(p Project, c Chapter, instruction, contextJSON string) (TextRequest, error) {
	id := selectedScriptPrompt(p)
	system := prompt.ScreenplayAuthoringChapter
	jsonResponse := false
	if id != "" {
		snap, err := ResolveProjectPrompt(p, id, "script")
		if err != nil {
			return TextRequest{}, fmt.Errorf("%w: %v", ErrInvalidArgument, err)
		}
		if !snap.Supported {
			return TextRequest{}, fmt.Errorf("%w: 所选剧本提示词的输出协议尚未支持：%s", ErrInvalidArgument, snap.OutputProtocol)
		}
		system = snap.Body
		// 144 requires JSON followed by an optional plain-text conversion report.
		// JSON-only provider mode would contradict that documented output contract.
		jsonResponse = false
	}
	if strings.TrimSpace(c.SourceText) == "" {
		return TextRequest{}, ErrEmptyContent
	}
	vars, err := chapterLibraryVars(contextJSON)
	if err != nil {
		return TextRequest{}, err
	}
	system = luoshui.Render(system, vars)
	user := contextJSON + "\n章节原文：\n" + c.SourceText + "\n用户要求：" + instruction
	if style := strings.TrimSpace(p.StyleDesc); style != "" {
		user += "\n画风文字：\n" + style
	}
	return TextRequest{Model: p.ModelID, SystemPrompt: system, UserPrompt: user, JSONResponse: jsonResponse, MaxOutputTokens: 16000, Temperature: 0.4}, nil
}
func ApplyChapterGeneration(c *Chapter, raw string) error {
	if c == nil {
		return ErrInvalidArgument
	}
	if strings.TrimSpace(raw) == "" {
		return ErrAIEmptyContent
	}
	text, err := professionalScriptMarkdown(raw)
	if err != nil {
		return err
	}
	c.RawResult = raw
	c.ScriptText = text
	return nil
}
func professionalScriptMarkdown(raw string) (string, error) {
	cleaned := strings.TrimSpace(raw)
	if !strings.HasPrefix(cleaned, "{") && !strings.HasPrefix(cleaned, "```json") {
		// Legacy Markdown remains valid, but a prefixed professional JSON response
		// must not silently become an editable plain-text script.
		if strings.Contains(cleaned, "\"scenes\"") || strings.Contains(cleaned, "【转换自检】") {
			return "", ErrAIInvalidJSON
		}
		return cleaned, nil
	}
	payload, err := decodeProfessionalScriptObject(cleaned)
	if err != nil {
		return "", err
	}
	var doc struct {
		Episode int    `json:"episode_index"`
		Title   string `json:"title"`
		Scenes  []struct {
			Number     string            `json:"scene_number"`
			Location   string            `json:"location"`
			Time       string            `json:"time"`
			Indoor     string            `json:"indoor_outdoor"`
			Characters []string          `json:"characters"`
			Content    []json.RawMessage `json:"script_content"`
		} `json:"scenes"`
	}
	if json.Unmarshal(payload, &doc) != nil || len(doc.Scenes) == 0 {
		return "", ErrAIInvalidJSON
	}
	if doc.Episode <= 0 {
		doc.Episode = 1
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# 第%d集：%s\n\n", doc.Episode, doc.Title)
	for i, scene := range doc.Scenes {
		if strings.TrimSpace(scene.Location) == "" || len(scene.Content) == 0 {
			return "", fmt.Errorf("%w: 场景地点或正文为空", ErrAIInvalidJSON)
		}
		number := scene.Number
		if number == "" {
			number = fmt.Sprintf("%d-%d", doc.Episode, i+1)
		}
		fmt.Fprintf(&b, "## %s：%s %s %s\n人物：%s\n\n", number, scene.Time, scene.Indoor, scene.Location, strings.Join(scene.Characters, " "))
		for _, rawElement := range scene.Content {
			var e struct {
				Type        string `json:"element_type"`
				Description string `json:"description"`
				Character   string `json:"character"`
				Content     string `json:"content"`
				Emotion     string `json:"emotion"`
				Voiceover   bool   `json:"is_voiceover"`
			}
			if json.Unmarshal(rawElement, &e) != nil {
				return "", ErrAIInvalidJSON
			}
			switch e.Type {
			case "action":
				if strings.TrimSpace(e.Description) == "" {
					return "", fmt.Errorf("%w: 动作正文为空", ErrAIInvalidJSON)
				}
				fmt.Fprintf(&b, "△ %s\n", e.Description)
			case "dialogue":
				if strings.TrimSpace(e.Character) == "" || strings.TrimSpace(e.Content) == "" {
					return "", fmt.Errorf("%w: 对白人物或正文为空", ErrAIInvalidJSON)
				}
				emotion := e.Emotion
				if e.Voiceover {
					emotion = strings.Trim(emotion+"、OS", "、")
				}
				if emotion != "" {
					fmt.Fprintf(&b, "%s（%s）：%s\n", e.Character, emotion, e.Content)
				} else {
					fmt.Fprintf(&b, "%s：%s\n", e.Character, e.Content)
				}
			default:
				return "", fmt.Errorf("%w: 未知剧本元素 %s", ErrAIInvalidJSON, e.Type)
			}
		}
		b.WriteString("\n")
	}
	return b.String(), nil
}
func chapterLibraryVars(contextJSON string) (map[string]string, error) {
	vars := map[string]string{"ROLES_PLACEHOLDER": "", "SCENES_PLACEHOLDER": "", "PROPS_PLACEHOLDER": ""}
	if strings.TrimSpace(contextJSON) == "" {
		return vars, nil
	}
	var context struct {
		Assets []Asset `json:"assets"`
	}
	if err := json.Unmarshal([]byte(contextJSON), &context); err != nil {
		return vars, fmt.Errorf("%w: 章节上下文不是有效JSON", ErrInvalidArgument)
	}
	grouped := map[AssetType][]Asset{}
	for _, a := range context.Assets {
		grouped[a.Type] = append(grouped[a.Type], a)
	}
	vars["ROLES_PLACEHOLDER"] = formatAssetLibrary(grouped[AssetCharacter])
	vars["SCENES_PLACEHOLDER"] = formatAssetLibrary(grouped[AssetScene])
	vars["PROPS_PLACEHOLDER"] = formatAssetLibrary(grouped[AssetProp])
	return vars, nil
}
func renderScriptLibrary(ctx context.Context, store Store, owner, projectID, body string) (string, error) {
	if !strings.Contains(body, "{ROLES_PLACEHOLDER}") && !strings.Contains(body, "{SCENES_PLACEHOLDER}") && !strings.Contains(body, "{PROPS_PLACEHOLDER}") {
		return body, nil
	}
	assets, err := store.ListAssets(ctx, owner, projectID, "")
	if err != nil {
		return "", err
	}
	raw, err := json.Marshal(map[string]any{"assets": assets})
	if err != nil {
		return "", err
	}
	vars, err := chapterLibraryVars(string(raw))
	if err != nil {
		return "", err
	}
	return luoshui.Render(body, vars), nil
}

// Each provider response is parsed independently. Segment boundaries are never
// inferred from self-check prose because that could swallow the next episode.
func editableScreenplayOutput(raw string) (string, error) { return professionalScriptMarkdown(raw) }

func selectedScreenplayOutput(raw, promptID string) (string, error) {
	if promptID == prompt.CreativeScript144 {
		if _, err := decodeProfessionalScriptObject(raw); err != nil {
			return "", err
		}
	}
	return editableScreenplayOutput(raw)
}

func decodeProfessionalScriptObject(raw string) (json.RawMessage, error) {
	cleaned := strings.TrimSpace(raw)
	fenced := strings.HasPrefix(cleaned, "```json")
	if fenced {
		cleaned = strings.TrimSpace(strings.TrimPrefix(cleaned, "```json"))
	}
	if !strings.HasPrefix(cleaned, "{") {
		return nil, ErrAIInvalidJSON
	}
	decoder := json.NewDecoder(strings.NewReader(cleaned))
	var payload json.RawMessage
	if err := decoder.Decode(&payload); err != nil {
		return nil, ErrAIInvalidJSON
	}
	tail := strings.TrimSpace(cleaned[decoder.InputOffset():])
	if fenced {
		if !strings.HasPrefix(tail, "```") {
			return nil, ErrAIInvalidJSON
		}
		tail = strings.TrimSpace(strings.TrimPrefix(tail, "```"))
	}
	if tail != "" && !validConversionSelfCheck(tail) {
		return nil, ErrAIInvalidJSON
	}
	return payload, nil
}

func validConversionSelfCheck(tail string) bool {
	lines := strings.Split(strings.ReplaceAll(tail, "\r\n", "\n"), "\n")
	if strings.TrimSpace(lines[0]) != "【转换自检】" {
		return false
	}
	allowed := map[string]bool{"原文对话总数": true, "action元素总数": true, "场景总数": true, "element_id 是否连续无重复": true, "每集结构": true}
	seen := map[string]bool{}
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// The report is plain text with the five documented field labels. It is
		// not an arbitrary tail container and cannot contain a second JSON object.
		if strings.ContainsAny(line, "{}[]") || strings.Contains(line, "```") {
			return false
		}
		label, value, found := strings.Cut(line, "：")
		if !found {
			label, value, found = strings.Cut(line, ":")
		}
		label = strings.TrimSpace(label)
		if !found || !allowed[label] || seen[label] || strings.TrimSpace(value) == "" {
			return false
		}
		seen[label] = true
	}
	return len(seen) > 0
}
