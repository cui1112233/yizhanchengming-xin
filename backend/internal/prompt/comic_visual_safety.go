package prompt

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

// ComicVisualSafetyRetryRule is appended only after the text model rejects a
// comic-page prompt for content policy. The retry request carries dialogue as
// stable placeholders; original dialogue is restored after JSON parsing and is
// never sent to the image model.
const ComicVisualSafetyRetryRule = `【画面安全重试】禁止画面出现暴力血腥，不得出现伤口、尸体、枪支弹药、管制刀具、殴打或虐待。禁止画面出现成人或暧昧内容，不得出现裸露、性暗示、亲密身体接触或令人不适的身体描写。把危险或亲密行为改成画外发生、人物保持距离交谈、克制表情、门外声响、事后环境或象征性空镜。用户原文对白已替换为“{对白N}”占位符；screen_prompt 必须原样保留这些占位符作为页面文字，image_prompt 不得包含占位符或推测原对白。只输出原协议要求的 JSON。`

const comicVisualStrictSafetyRetryRule = `【第二轮严格安全重试】只允许日常、全年龄、无身体接触的安全画面。人物必须衣着完整并保持正常社交距离；优先使用白天公共场所、普通室内交谈、物品特写或无人空镜。不得补写被省略的动作，不得猜测“{对白N}”的真实文字。screen_prompt 只保留占位符，image_prompt 只描述安全画面。`

var comicVisualSafetyReplacer = strings.NewReplacer(
	"鲜血喷射", "紧张气氛",
	"血液喷射", "紧张气氛",
	"血流成河", "事后凌乱环境",
	"血腥", "紧张",
	"鲜血", "红色痕迹",
	"流血", "受伤后神情",
	"打斗", "隔空争执",
	"群殴", "多人激烈对峙",
	"殴打", "激烈对峙",
	"撕打", "激烈争执",
	"击杀", "制服",
	"杀死", "制止",
	"枪支", "危险物品",
	"手枪", "危险物品",
	"子弹", "金属物件",
	"管制刀具", "危险物品",
	"甩棍", "危险物品",
	"刀", "危险物品",
	"靠得很近", "保持正常距离",
	"贴得很近", "保持正常距离",
	"身体贴近", "保持正常距离",
	"亲密接触", "保持正常距离交谈",
	"暧昧气氛", "克制气氛",
	"气氛暧昧", "气氛克制",
	"暧昧", "克制",
	"裸露", "衣着完整",
	"脱衣", "整理衣着",
	"脱掉衣服", "整理衣着",
	"亲吻", "保持距离交谈",
	"接吻", "保持距离交谈",
	"拥抱", "远处点头致意",
	"床上", "普通室内",
	"性暗示", "日常交流",
)

type comicVisualDialogue struct {
	Placeholder string
	Text        string
}

// ComicVisualSafetyPlan is a request-only sanitized copy. dialogues remains
// private so callers can restore text only through RestoreScreenPrompt.
type ComicVisualSafetyPlan struct {
	SystemPrompt string
	UserPrompt   string
	dialogues    []comicVisualDialogue
}

// ComicVisualSafetyRetryPlan builds one of two bounded safety retries.
func ComicVisualSafetyRetryPlan(systemPrompt, userPrompt string, round int) ComicVisualSafetyPlan {
	system := strings.TrimSpace(systemPrompt)
	if system != "" {
		system += "\n\n"
	}
	system += ComicVisualSafetyRetryRule
	if round >= 2 {
		system += "\n" + comicVisualStrictSafetyRetryRule
	}
	safeUser, dialogues := maskComicVisualDialogue(userPrompt)
	safeUser = comicVisualSafetyReplacer.Replace(safeUser)
	if round >= 2 {
		safeUser = strictComicVisualSafetyUser(safeUser, dialogues)
	}
	return ComicVisualSafetyPlan{
		SystemPrompt: system,
		UserPrompt:   safeUser,
		dialogues:    dialogues,
	}
}

func strictComicVisualSafetyUser(safeUser string, dialogues []comicVisualDialogue) string {
	placeholders := make([]string, 0, len(dialogues))
	for _, dialogue := range dialogues {
		placeholders = append(placeholders, dialogue.Placeholder)
	}
	fallback := "请生成全年龄日常安全漫画画面。人物衣着完整、保持正常社交距离，不出现身体接触；可使用普通室内交谈、公共场所或无人空镜。"
	if len(placeholders) > 0 {
		fallback += "screen_prompt 页面文字只保留这些占位符：" + strings.Join(placeholders, "、") + "；image_prompt 不得包含占位符。"
	}

	var source map[string]any
	if json.Unmarshal([]byte(safeUser), &source) != nil {
		return fallback
	}
	strict := map[string]any{
		"book_title":    "安全替代漫画页",
		"current_page":  fallback,
		"source_text":   fallback,
		"previous_page": "",
		"next_page":     "",
	}
	for _, key := range []string{"page_label", "aspect_ratio"} {
		if value, ok := source[key]; ok {
			strict[key] = value
		}
	}
	if assets, ok := source["comic_assets"].([]any); ok {
		identities := make([]map[string]any, 0, len(assets))
		for _, rawAsset := range assets {
			asset, ok := rawAsset.(map[string]any)
			if !ok {
				continue
			}
			identity := make(map[string]any)
			for _, key := range []string{"asset_id", "kind", "name", "formal_name"} {
				if value, exists := asset[key]; exists {
					identity[key] = value
				}
			}
			identities = append(identities, identity)
		}
		strict["comic_assets"] = identities
	}
	encoded, err := json.Marshal(strict)
	if err != nil {
		return fallback
	}
	return string(encoded)
}

// RestoreScreenPrompt puts the original dialogue back only after the text
// model has returned valid JSON. ImagePrompt deliberately has no equivalent.
func (p ComicVisualSafetyPlan) RestoreScreenPrompt(screenPrompt string) string {
	restored := screenPrompt
	missing := make([]string, 0, len(p.dialogues))
	for _, dialogue := range p.dialogues {
		if strings.Contains(restored, dialogue.Placeholder) {
			restored = strings.ReplaceAll(restored, dialogue.Placeholder, dialogue.Text)
			continue
		}
		missing = append(missing, "“"+dialogue.Text+"”")
	}
	if len(missing) > 0 {
		restored = strings.TrimSpace(restored) + "\n页面文字（保留原文对白）：" + strings.Join(missing, "；")
	}
	return strings.TrimSpace(restored)
}

func maskComicVisualDialogue(text string) (string, []comicVisualDialogue) {
	var out strings.Builder
	remaining := text
	dialogues := make([]comicVisualDialogue, 0)
	known := make(map[string]string)
	for remaining != "" {
		openAt, closeMark := nextComicDialogueQuote(remaining)
		if openAt < 0 {
			out.WriteString(remaining)
			break
		}
		out.WriteString(remaining[:openAt])
		quoted := remaining[openAt:]
		openRune, openBytes := utf8.DecodeRuneInString(quoted)
		closeAt := strings.Index(quoted[openBytes:], closeMark)
		if closeAt < 0 {
			out.WriteString(quoted)
			break
		}
		closeAt += openBytes
		textValue := quoted[openBytes:closeAt]
		placeholder, exists := known[textValue]
		if !exists {
			placeholder = "{对白" + fmt.Sprint(len(dialogues)+1) + "}"
			known[textValue] = placeholder
			dialogues = append(dialogues, comicVisualDialogue{Placeholder: placeholder, Text: textValue})
		}
		out.WriteRune(openRune)
		out.WriteString(placeholder)
		out.WriteString(closeMark)
		remaining = quoted[closeAt+len(closeMark):]
	}
	return out.String(), dialogues
}

// ComicVisualSafetyRetry returns a backend-owned policy retry. It never writes
// the sanitized copy back to the novel or page source.
func ComicVisualSafetyRetry(systemPrompt, userPrompt string) (string, string) {
	plan := ComicVisualSafetyRetryPlan(systemPrompt, userPrompt, 1)
	return plan.SystemPrompt, plan.UserPrompt
}

func sanitizeComicVisualTextPreservingDialogue(text string) string {
	var out strings.Builder
	remaining := text
	for remaining != "" {
		openAt, closeMark := nextComicDialogueQuote(remaining)
		if openAt < 0 {
			out.WriteString(comicVisualSafetyReplacer.Replace(remaining))
			break
		}
		out.WriteString(comicVisualSafetyReplacer.Replace(remaining[:openAt]))
		quoted := remaining[openAt:]
		_, openBytes := utf8.DecodeRuneInString(quoted)
		closeAt := strings.Index(quoted[openBytes:], closeMark)
		if closeAt < 0 {
			out.WriteString(comicVisualSafetyReplacer.Replace(quoted))
			break
		}
		closeAt += openBytes
		end := closeAt + len(closeMark)
		out.WriteString(quoted[:end])
		remaining = quoted[end:]
	}
	return out.String()
}

func nextComicDialogueQuote(text string) (int, string) {
	best := -1
	closeMark := ""
	for _, pair := range [][2]string{{"“", "”"}, {"「", "」"}, {"『", "』"}} {
		if at := strings.Index(text, pair[0]); at >= 0 && (best < 0 || at < best) {
			best = at
			closeMark = pair[1]
		}
	}
	return best, closeMark
}
