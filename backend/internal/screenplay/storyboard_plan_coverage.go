package screenplay

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// StoryboardPlanCoverage describes structural completion, not artistic quality.
// Rules are independently implemented from controlled original evaluator probes;
// provenance and exact original responses are in the compressed golden fixture.
type StoryboardPlanCoverage struct {
	OK       bool                         `json:"ok"`
	Chars    int                          `json:"chars"`
	Headings []string                     `json:"headings"`
	Missing  []string                     `json:"missing"`
	Wrong    []StoryboardPlanWrongHeading `json:"wrong"`
	Rows     int                          `json:"rows"`
	Declared int                          `json:"declared"`
	Reason   string                       `json:"reason"`
}
type StoryboardPlanWrongHeading struct {
	Heading  string `json:"heading"`
	Actual   string `json:"actual"`
	Expected string `json:"expected"`
}

var planHeadingPattern = regexp.MustCompile(`(?m)^(?:#{1,3}\s*)?([一二三四五六七八])[、.．]\s*(\S[^\n]*)`)
var planRowPattern = regexp.MustCompile(`(?m)^\|\s*(\d+)\s*\|`)
var planDeclaredPattern = regexp.MustCompile(`总规划[^\n]*?(\d+)\s*节`)
var planChapterNumbers = []string{"一", "二", "三", "四", "五", "六"}
var planChapterLabels = []string{"总览", "资产绑定核对", "分节明细", "生成前必须处理的风险点", "待确认项", "跨集承接与批处理"}
var planChapterKeys = [][]string{{"总览"}, {"资产", "命名"}, {"分节", "节表", "内容落点", "落点"}, {"风险", "高危"}, {"待确认"}, {"跨集", "承接", "批处理"}}

func EvaluateStoryboardPlan(text, density string) StoryboardPlanCoverage {
	text = strings.TrimSpace(text)
	out := StoryboardPlanCoverage{Chars: utf8.RuneCountInString(text), Headings: []string{}, Missing: []string{}, Wrong: []StoryboardPlanWrongHeading{}}
	matches := planHeadingPattern.FindAllStringSubmatchIndex(text, -1)
	seen := map[string]bool{}
	for _, m := range matches {
		number := text[m[2]:m[3]]
		if seen[number] {
			continue
		}
		seen[number] = true
		out.Headings = append(out.Headings, number)
		for i, n := range planChapterNumbers {
			if n != number {
				continue
			}
			title := strings.TrimSpace(text[m[4]:m[5]])
			valid := false
			for _, key := range planChapterKeys[i] {
				valid = valid || strings.Contains(title, key)
			}
			if !valid {
				out.Wrong = append(out.Wrong, StoryboardPlanWrongHeading{number, title, planChapterLabels[i]})
			}
		}
	}
	for _, n := range planChapterNumbers {
		if !seen[n] {
			out.Missing = append(out.Missing, n)
		}
	}
	rows := map[int]bool{}
	for _, m := range planRowPattern.FindAllStringSubmatch(text, -1) {
		n, err := strconv.Atoi(m[1])
		if err == nil {
			rows[n] = true
		}
	}
	out.Rows = len(rows)
	if m := planDeclaredPattern.FindStringSubmatch(text); len(m) > 1 {
		out.Declared, _ = strconv.Atoi(m[1])
	}
	switch {
	case text == "":
		out.Reason = "没有规划文档"
	case len(out.Wrong) > 0 && len(out.Missing) == 0:
		descriptions := []string{}
		for _, w := range out.Wrong {
			descriptions = append(descriptions, fmt.Sprintf("第%s章写成「%s」，应为「%s」", w.Heading, w.Actual, w.Expected))
		}
		out.Reason = "章标题不是体例规定的六章：" + strings.Join(descriptions, "；") + "。注意「章」是本文档的六个部分，「节」是几十个分镜小节（时长跟当前规划档走），不要混为一谈。请先 read_plan_guide 对齐六章标题，再 write_file target=plan 按体例重写，禁止自创章名。"
	case len(out.Missing) > 0:
		last := "未分章"
		if len(out.Headings) > 0 {
			last = out.Headings[len(out.Headings)-1]
		}
		out.Reason = fmt.Sprintf("规划已有 %d 字，写到第%s章，还缺第%s章。请 read_file target=plan view=tail，再用 append_file 从断开处续写，不要 write_file 覆盖已写部分。", out.Chars, last, strings.Join(out.Missing, "、"))
	default:
		minimum := 2500
		if NormalizePlanDensity(density) == "sd25" {
			minimum = 1800
		}
		if out.Chars < minimum {
			detail := "对白密集的一集通常会有较多节；"
			if NormalizePlanDensity(density) == "sd25" {
				detail = "SD2.5 专项按 26–30 秒一节（硬框 25.0–30.0），三千字对白大约八到十节，不要按 15 秒时代的 30–50 节来写；"
			}
			out.Reason = fmt.Sprintf("六章标题齐了但正文过短（%d 字）。%s请 append_file 把第三章的节表按场次逐场补齐，每节一行。", out.Chars, detail)
			break
		}
		budget := false
		for _, key := range []string{"剧本字数", "剧本约", "参考节数", "N_ref", "节数下限", "N_min", "字数账", "台词账", "发声条数", "口播"} {
			budget = budget || strings.Contains(text, key)
		}
		if !budget {
			out.Reason = "第一章缺字数账结论（本集剧本字数 / 综合分镜词后的总节数）。请按体例写一句：本集约 C 字，综合分镜词后拆成 N 节。字数只是参考，总节数不是下限，以分镜词为准。"
			break
		}
		if out.Declared > 0 && out.Rows < int(float64(out.Declared)*.9) {
			out.Reason = fmt.Sprintf("总览自报 %d 节，但第三章节表只列出 %d 行。第三章必须一节一行列满，禁止省略号和「其余同上」。请 read_file target=plan view=tail，再 append_file 把缺的场次逐场补上。", out.Declared, out.Rows)
			break
		}
		tail := ""
		if len(matches) > 0 {
			tail = strings.TrimSpace(text[matches[len(matches)-1][1]:])
		}
		if utf8.RuneCountInString(tail) < 80 {
			out.Reason = "第六章正文过短，像写到一半被截断。请 append_file 接着写完，不要重开标题。"
			break
		}
		out.OK = true
		out.Reason = fmt.Sprintf("规划六章已齐（%d 字 · 第三章 %d 节）", out.Chars, out.Rows)
	}
	return out
}
