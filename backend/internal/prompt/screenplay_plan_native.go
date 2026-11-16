package prompt

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Text recovered by controlled calls to the installed compiled guide function.
// See reconstruction contract34 and the original-request golden fixture.
//
//go:embed screenplay_plan_guides_native.json
var screenplayNativePlanGuidesJSON []byte
var screenplayNativePlanGuides = func() map[string]string {
	var out map[string]string
	if err := json.Unmarshal(screenplayNativePlanGuidesJSON, &out); err != nil {
		panic(err)
	}
	return out
}()

const screenplayNativePlanSystem = "你为本集写六章分镜规划 Markdown。遵守已读分镜词，但禁止把条款原文、红线编号写进规划。只输出规划正文。文档可能很长；写不完会从截断处续写，请始终接着写到第六章结束，不要重开标题。"

func ScreenplayNativePlanGuide(workTitle, episodeTitle, promptName, density, continuation string) string {
	if workTitle == "" {
		workTitle = "本作品"
	}
	if episodeTitle == "" {
		episodeTitle = "本集"
	}
	if promptName == "" {
		promptName = "当前分镜词"
	}
	if density != "compact" && density != "more" && density != "sd25" {
		density = "standard"
	}
	if continuation != "tools" {
		continuation = "chat"
	}
	return strings.NewReplacer("{{WORK_TITLE}}", workTitle, "{{EPISODE_TITLE}}", episodeTitle, "{{PROMPT_NAME}}", promptName).Replace(screenplayNativePlanGuides[density+"/"+continuation])
}

func ScreenplayNativePlanningRequest(workTitle, episodeTitle, promptName, body, density, previous, script string) (string, string) {
	previous = strings.TrimSpace(previous)
	prior := "【上集规划】无。节1按本集开场建立，第一章入口承接标明待核对上集末镜。"
	if previous != "" {
		r := []rune(previous)
		if len(r) > 2200 {
			previous = string(r[len(r)-2200:])
		}
		prior = "【上集规划摘录（末段，供第六章跨集承接；不要提前兑现上集未完成的伏笔）】\n" + previous
	}
	return screenplayNativePlanSystem, fmt.Sprintf("【规划体例】\n%s\n\n【分镜词·遵守勿复述】\n%s\n\n%s\n\n【本集剧本】\n%s\n\n请输出完整六章规划。尽量一次写完；写不完会自动从断开处续写。", ScreenplayNativePlanGuide(workTitle, episodeTitle, promptName, density, "chat"), body, prior, script)
}

var screenplayContinueHeading = regexp.MustCompile(`(?m)^(?:#{1,3}\s*)?([一二三四五六七八])[、.．]\s*(\S[^\n]*)`)

func ScreenplayNativePlanContinue(accumulated string) string {
	text := strings.TrimSpace(accumulated)
	missing := []string{}
	headings := map[string]bool{}
	for _, m := range screenplayContinueHeading.FindAllStringSubmatch(text, -1) {
		headings[m[1]] = true
	}
	for _, n := range []string{"一", "二", "三", "四", "五", "六"} {
		if !headings[n] {
			missing = append(missing, "第"+n+"章")
		}
	}
	miss := strings.Join(missing, "、")
	if miss == "" {
		miss = "第六章收束"
	}
	lines := strings.Split(strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n"), "\n")
	if len(lines) > 16 {
		lines = lines[len(lines)-16:]
	}
	tail := strings.Join(lines, "\n")
	if tail == "" {
		tail = "（空）"
	}
	return fmt.Sprintf("上一轮输出被截断，规划文档还没写完。\n当前进度：已 %d 字，还缺%s。\n\n请从断开处无缝续写：\n1. 不要重复已经写过的标题、表格行或段落；\n2. 不要重开文档标题或「生成前基线」；\n3. 从上一字之后接着写，直到「## 六、跨集承接与批处理」写完并自然收束；\n4. 直接输出续写正文。\n\n【已输出末尾（请从此后继续）】\n%s", utf8.RuneCountInString(accumulated), miss, tail)
}
