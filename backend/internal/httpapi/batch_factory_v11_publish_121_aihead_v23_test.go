package httpapi

import "testing"

// 23 版需求 1：materialReuse 必须驱动 121 的「素材使用」
// （aiHeadVideoMode → jieya_ai_head_video_mode），true→"1"、false→"0"。
func TestV23MaterialReuseMapsToAIHeadVideoMode(t *testing.T) {
	cases := []struct {
		name string
		pub  string
		want string
	}{
		{"复用", `{"organization":"1","jieya_ai_head":"3","materialReuse":true}`, "1"},
		{"不复用", `{"organization":"1","jieya_ai_head":"3","materialReuse":false}`, "0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			batch, book := publishFieldsFixture(t, tc.pub)
			fields, err := buildPublishFields(batch, book, map[string]any{})
			if err != nil {
				t.Fatal(err)
			}
			if got := fields["jieya_ai_head_video_mode"]; got != tc.want {
				t.Fatalf("jieya_ai_head_video_mode = %q, want %q (fields=%v)", got, tc.want, fields)
			}
		})
	}
}

// 23 版需求 1：materialReuse 不再写 biaohong_reuse（标红复用有独立开关）。
func TestV23MaterialReuseNoLongerWritesBiaohongReuse(t *testing.T) {
	batch, book := publishFieldsFixture(t, `{"organization":"1","materialReuse":true}`)
	fields, err := buildPublishFields(batch, book, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if got := fields["biaohong_reuse"]; got != "" {
		t.Fatalf("biaohong_reuse = %q, want empty (materialReuse must not drive it)", got)
	}
}

// 23 版需求 1：标红复用仍然可以被独立开关打开。
func TestV23BiaohongReuseHasItsOwnSwitch(t *testing.T) {
	batch, book := publishFieldsFixture(t, `{"organization":"1","biaohongReuse":true}`)
	fields, err := buildPublishFields(batch, book, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if got := fields["biaohong_reuse"]; got != "1" {
		t.Fatalf("biaohong_reuse = %q, want \"1\"", got)
	}
}

// 23 版需求 3：配置方案 advanced 里的 aiHeadVideoMode 必须能落到 121 字段。
func TestV23AdvancedAliasMapsAIHeadVideoMode(t *testing.T) {
	publish := `{
		"websiteProfileId": "121-368",
		"organization": "1",
		"websiteProfiles": [{
			"id": "121-368", "name": "AI前贴", "config_id": "368",
			"advanced": {"aiHeadVideoMode": "1", "jieyaAiHead": "3"}
		}]
	}`
	batch, book := publishFieldsFixture(t, publish)
	fields, err := buildPublishFields(batch, book, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if got := fields["jieya_ai_head_video_mode"]; got != "1" {
		t.Fatalf("jieya_ai_head_video_mode = %q, want \"1\" (fields=%v)", got, fields)
	}
}

// 23 版需求 2：拼接方式不是 AI 前贴时，不得提交 AI 头部视频及其复用开关。
func TestV23AIHeadVideoFieldsSkippedWhenNotAIHead(t *testing.T) {
	batch, book := publishFieldsFixture(t, `{"organization":"1","jieya_ai_head":"1","materialReuse":true}`)
	fields, err := buildPublishFields(batch, book, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := fields["jieya_ai_head_video_mode"]; ok {
		t.Fatalf("jieya_ai_head_video_mode = %q, want absent when jieya_ai_head != 3 (fields=%v)", got, fields)
	}
	if got, ok := fields["jieya_ai_head_video"]; ok {
		t.Fatalf("jieya_ai_head_video = %q, want absent when jieya_ai_head != 3 (fields=%v)", got, fields)
	}
}
