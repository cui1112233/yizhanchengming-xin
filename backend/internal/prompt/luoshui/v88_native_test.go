package luoshui

import (
	"strings"
	"testing"
)

// V88 内置运行词（剧本医生/分镜医生/故事概要）必须注册进默认目录，
// 管理后台「漫剧创作」分区才能展示，运行时才能经 RuntimeBody 读覆盖版本。
func TestV88NativeEntriesRegistered(t *testing.T) {
	reg := MustDefault()
	for _, id := range []string{V88DoctorScreenplayID, V88DoctorStoryboardID, V88SynopsisID} {
		entry, ok := reg.Get(id)
		if !ok {
			t.Fatalf("缺少内置运行词 %s", id)
		}
		if entry.Section != "v88" || entry.Source != "V88 内置运行词" {
			t.Fatalf("%s 元信息错误: %+v", id, entry)
		}
		if strings.TrimSpace(entry.Body()) == "" {
			t.Fatalf("%s 正文为空", id)
		}
		if !strings.Contains(entry.Body(), "JSON") {
			t.Fatalf("%s 正文缺少 JSON 输出约定: %q", id, entry.Body())
		}
	}
	if !strings.Contains(V88SynopsisBody, "300-500") {
		t.Fatal("故事概要内置词应保留洛水 SYNOPSIS_AGENT_PROMPT 规格")
	}
}

// RuntimeBody 覆盖通道：后台发布覆盖版本后运行时读到覆盖正文；
// 无覆盖时回退内置默认。
func TestV88RuntimeBodyPrefersOverride(t *testing.T) {
	SetOverrideResolver(func(id string) (string, bool) {
		if id == V88DoctorScreenplayID {
			return "管理员定制的剧本医生提示词", true
		}
		return "", false
	})
	t.Cleanup(func() { SetOverrideResolver(nil) })

	if body := V88RuntimeBody(V88DoctorScreenplayID); body != "管理员定制的剧本医生提示词" {
		t.Fatalf("覆盖未生效: %q", body)
	}
	if !strings.Contains(V88RuntimeBody(V88DoctorStoryboardID), "分镜词医生") {
		t.Fatal("未覆盖条目应回退内置默认")
	}
	if !strings.Contains(V88RuntimeBody(V88SynopsisID), "故事摘要与项目信息提取助手") {
		t.Fatal("未覆盖概要词应回退内置默认")
	}
}
