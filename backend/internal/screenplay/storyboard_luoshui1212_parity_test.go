package screenplay

import (
	"encoding/json"
	"os"
	"testing"
)

// TestLuoshui1212StoryboardNativeParse 用洛水真实项目 1212 的分镜 JSON
// 喂给 V88 原生解析器，验证字段能否无损解析。纯离线，不调用任何模型。
func TestLuoshui1212StoryboardNativeParse(t *testing.T) {
	raw, err := os.ReadFile("testdata/luoshui1212_storyboard.json")
	if err != nil {
		t.Fatalf("读取洛水真实分镜失败: %v", err)
	}
	var src struct {
		Storyboard []map[string]any `json:"storyboard"`
	}
	if err := json.Unmarshal(raw, &src); err != nil {
		t.Fatalf("解析源 JSON 失败: %v", err)
	}
	if len(src.Storyboard) != 3 {
		t.Fatalf("源分镜数=%d, want 3", len(src.Storyboard))
	}

	shots, err := parseStoryboardJSON(string(raw))
	if err != nil {
		t.Fatalf("V88 原生解析器拒绝洛水真实数据: %v", err)
	}
	if len(shots) != len(src.Storyboard) {
		t.Fatalf("解析后分镜数=%d, want %d", len(shots), len(src.Storyboard))
	}

	// 核心字段比对：洛水原始 JSON 的关键字段值必须能在解析结果里找到对应
	for i, want := range src.Storyboard {
		got := shots[i]
		checks := []struct {
			field string
			want  any
			got   any
		}{
			{"id/seq", want["id"], got.Seq},
			{"scene_id", want["scene_id"], got.SceneID},
			{"duration_sec", want["duration_sec"], got.Duration},
			{"characters", want["characters"], got.Characters},
			{"scene", want["scene"], got.Scene},
			{"props", want["props"], got.Props},
			{"video_desc", want["video_desc"], got.VideoDesc},
			{"_scene_index", want["_scene_index"], got.SourceSceneIndex},
			{"_scene_local_order", want["_scene_local_order"], got.SourceSceneLocalOrder},
		}
		for _, c := range checks {
			if !jsonEqual(c.want, c.got) {
				t.Errorf("shot[%d].%s: 洛水=%v (type %T), V88=%v (type %T)", i, c.field, c.want, c.want, c.got, c.got)
			}
		}
		// 整个原始分组应被完整保存在 ProtocolFields（扩展字段不丢）
		var protocol map[string]any
		if err := json.Unmarshal(got.ProtocolFields, &protocol); err != nil {
			t.Errorf("shot[%d].ProtocolFields 无法解析: %v", i, err)
			continue
		}
		for k, v := range want {
			if _, ok := protocol[k]; !ok {
				t.Errorf("shot[%d].ProtocolFields 丢失字段 %q", i, k)
			}
			if !jsonEqual(protocol[k], v) {
				t.Errorf("shot[%d].ProtocolFields[%s]: 洛水=%v, V88=%v", i, k, v, protocol[k])
			}
		}
	}

	// 形态差异（非数据丢失）：image_prompt —— 洛水 307 提示词只用 video_desc，
	// 原始 JSON 无该字段，V88 解析为 ""，nil 与空串语义等价。
	t.Logf("形态差异: 洛水分镜无 image_prompt 字段（307 提示词仅用 video_desc），V88 解析为空串，语义等价")
	t.Logf("结论: 洛水 1212 真实分镜 %d 条核心字段全部通过 V88 原生解析且无损；_scene_local_order 已映射专用字段；扩展字段经 ProtocolFields 完整保留", len(shots))
}

// jsonEqual 把数值统一成 float64 后比较，避免 int/float64 类型差异误报。
func jsonEqual(a, b any) bool {
	af, aok := toFloat(a)
	bf, bok := toFloat(b)
	if aok && bok {
		return af == bf
	}
	return deepEqualJSON(a) == deepEqualJSON(b)
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, _ := n.Float64()
		return f, true
	}
	return 0, false
}

func deepEqualJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
