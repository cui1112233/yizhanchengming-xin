package luoshui

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

// 语料必须逐条可复核：规格 §3 的 35 条 + 插件市场 4 条文字类工具补齐的 13 条洛水原文，
// SHA256[:16] 可复核、长度一致。
func TestRegistryMatchesSpecCorpus(t *testing.T) {
	reg, err := Default()
	if err != nil {
		t.Fatalf("加载语料失败: %v", err)
	}
	entries := reg.List()
	// 53 条取证语料 + 3 条 V88 内置运行词（v88_doctor_screenplay / v88_doctor_storyboard / v88_synopsis，
	// 见 v88_native.go，Section=v88，不参与取证 SHA 校验）。
	if len(entries) != 56 {
		t.Fatalf("提示词条数 = %d, want 56", len(entries))
	}
	v88Count := 0
	for _, entry := range entries {
		if entry.Section == "v88" {
			v88Count++
			continue
		}
		raw := reg.entries[entry.ID].body
		if raw == "" {
			t.Fatalf("缺少提示词 %s", entry.ID)
		}
		// 规格给出的 SHA256 是对常量池原串（含前缀）计算的，因此校验原始正文。
		sum := sha256.Sum256([]byte(raw))
		got := hex.EncodeToString(sum[:])[:16]
		want := entry.ExpectedSHA256_16
		if want != "" && got != want {
			t.Fatalf("%s SHA256[:16] = %s, want %s", entry.ID, got, want)
		}
		if entry.Length > 0 && len([]rune(raw)) != entry.Length {
			t.Fatalf("%s 长度 = %d, want %d", entry.ID, len([]rune(raw)), entry.Length)
		}
	}
	if v88Count != 3 {
		t.Fatalf("V88 内置运行词条数 = %d, want 3", v88Count)
	}
}

func TestHiddenPrompts(t *testing.T) {
	reg := MustDefault()
	for _, id := range []string{"a1_novel_to_screenplay", "storyboard_xiaoluo_v7_15s"} {
		entry, ok := reg.Get(id)
		if !ok {
			t.Fatalf("缺少提示词 %s", id)
		}
		if !entry.HiddenFromUser() {
			t.Fatalf("%s 必须禁止向用户展示", id)
		}
	}
	if entry, _ := reg.Get("b_director_screenplay"); entry.HiddenFromUser() {
		t.Fatalf("B 档提示词不应被标记为禁止展示")
	}
}

func TestConstantPoolPrefixStripped(t *testing.T) {
	reg := MustDefault()
	entry, ok := reg.Get("batch_plan_steps")
	if !ok {
		t.Fatal("缺少 batch_plan_steps")
	}
	if entry.ConstantPoolPrefix != "u" {
		t.Fatalf("常量池前缀 = %q, want u", entry.ConstantPoolPrefix)
	}
	body := entry.Body()
	if strings.HasPrefix(body, "u") {
		t.Fatalf("运行时应剥离常量池前缀，got %q", body[:8])
	}
	if !strings.HasPrefix(body, "# 短剧批次 8 步") {
		t.Fatalf("剥离后正文不符: %q", body[:24])
	}
}

func TestRender(t *testing.T) {
	body, ok := MustDefault().RenderByID("batch_step_5", map[string]string{"N": "3"})
	if !ok {
		t.Fatal("缺少 batch_step_5")
	}
	if !strings.Contains(body, "第 3 集") {
		t.Fatalf("占位符未替换: %s", body)
	}
	if strings.Contains(body, "{N}") {
		t.Fatalf("仍残留占位符: %s", body)
	}
	// 未提供的占位符必须原样保留，便于暴露缺口。
	partial := Render("第 {N} 集 · {UNSET}", map[string]string{"N": "2"})
	if partial != "第 2 集 · {UNSET}" {
		t.Fatalf("Render = %q", partial)
	}
}
