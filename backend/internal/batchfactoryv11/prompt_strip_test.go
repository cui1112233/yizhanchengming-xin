package batchfactoryv11

import (
	"strings"
	"testing"
)

func TestStripBaseSetupSection(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "有基础设定段",
			in:   "镜头画面：\n00:00-00:03 | 中景 | 女主坐在床边\n【基础设定】\n人物：女主，年轻女子\n场景：古代闺房\n\n00:03-00:06 | 近景 | 女主落泪",
			want: "镜头画面：\n00:00-00:03 | 中景 | 女主坐在床边\n\n00:03-00:06 | 近景 | 女主落泪",
		},
		{
			name: "无基础设定段",
			in:   "镜头画面：\n00:00-00:03 | 中景 | 女主坐在床边",
			want: "镜头画面：\n00:00-00:03 | 中景 | 女主坐在床边",
		},
		{
			name: "空串",
			in:   "",
			want: "",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := stripBaseSetupSection(c.in)
			if got != c.want {
				t.Errorf("stripBaseSetupSection(…) = %q; want %q", got, c.want)
			}
		})
	}
}

func TestStripStandaloneSetupSections(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "有人物卡",
			in:   "【人物卡】\n张三：年轻人\n\n镜头画面：\n00:00-00:03 | 中景 | 动作",
			want: "镜头画面：\n00:00-00:03 | 中景 | 动作",
		},
		{
			name: "有人物与场景",
			in:   "【人物与场景】\n人物：张三\n场景：街市\n\n时间轴：\n00:00开始",
			want: "时间轴：\n00:00开始",
		},
		{
			name: "无共享段",
			in:   "镜头画面：\n00:00 | 全景 | 动作",
			want: "镜头画面：\n00:00 | 全景 | 动作",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := stripStandaloneSetupSections(c.in)
			if got != c.want {
				t.Errorf("stripStandaloneSetupSections(…) = %q; want %q", got, c.want)
			}
		})
	}
}

func TestStripLegacySharedSetupSections(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "有统一风格",
			in:   "统一风格：写实古风宫廷\n\n镜头画面：\n00:00 | 中景 | 动作",
			want: "镜头画面：\n00:00 | 中景 | 动作",
		},
		{
			name: "有统一人物",
			in:   "统一人物：张三，年轻男子\n场景环境：古代街市\n\n镜头画面：\n00:00 | 近景 | 表情",
			want: "镜头画面：\n00:00 | 近景 | 表情",
		},
		{
			name: "无旧格式",
			in:   "镜头画面：\n00:00 | 全景 | 动作",
			want: "镜头画面：\n00:00 | 全景 | 动作",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := stripLegacySharedSetupSections(c.in)
			if got != c.want {
				t.Errorf("stripLegacySharedSetupSections(…) = %q; want %q", got, c.want)
			}
		})
	}
}

func TestStripConstraintLines(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "有多条约束",
			in:   "【画面前缀】写实古风\n\n正文\n\n【画质约束】4K 电影级\n负面提示词：多手多脚",
			want: "正文",
		},
		{
			name: "无约束行",
			in:   "正文\n\n正文续",
			want: "正文\n\n正文续",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := stripConstraintLines(c.in)
			if got != c.want {
				t.Errorf("stripConstraintLines(…) = %q; want %q", got, c.want)
			}
		})
	}
}

func TestStripSharedPromptSections(t *testing.T) {
	in := "统一风格：古风\n【人物卡】\n张三：年轻\n【基础设定】\n人物：张三\n\n镜头画面：\n00:00 | 中景 | 动作\n\n【画面前缀】写实\n负面提示词：模糊"
	want := "镜头画面：\n00:00 | 中景 | 动作"
	got := stripSharedPromptSections(in)
	if got != want {
		t.Errorf("stripSharedPromptSections(…) = %q; want %q", got, want)
	}
}

func TestStripSharedPromptSectionsRemovesPlainLegacyBaseSetup(t *testing.T) {
	in := "画面前缀：都市电影感\n基础设定：人物：\n我：现代女性，陆淮安的妻子。清瘦，黑色长发。\n陆淮安：女主角的丈夫。高大，短发。\n场景：医院走廊。\n道具：手机。\n\n视频提示词：\n镜头画面：两人保持距离。"
	got := stripSharedPromptSections(in)
	if strings.Contains(got, "基础设定") || strings.Contains(got, "陆淮安的妻子") || strings.Contains(got, "女主角的丈夫") {
		t.Fatalf("plain legacy base setup leaked: %q", got)
	}
	if !strings.Contains(got, "视频提示词：") || !strings.Contains(got, "两人保持距离") {
		t.Fatalf("storyboard body was removed with the legacy setup: %q", got)
	}
}
