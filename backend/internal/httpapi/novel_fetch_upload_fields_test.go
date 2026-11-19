package httpapi

import (
	"reflect"
	"testing"
)

func TestBuildUploadFieldsCompatMapping(t *testing.T) {
	cases := []struct {
		name     string
		platform int
		gender   string
		style    string
		advanced map[string]any
		item     map[string]any
		want     map[string]string
		wantErr  bool
	}{
		{
			name: "默认映射", platform: 3, gender: "女", style: "现代悬疑",
			want: map[string]string{
				"platform_id": "3", "organization": "1", "gender": "2", "style": "303",
				"tl5": "0", "ziti": "1", "zitidx": "62", "biaohong": "", "biaohong_reuse": "", "keywords": "",
				"jieya_num": "4", "jieya_ai_head": "0", "jieya_speed": "1.7", "jieya_pitch": "0",
				"font_color_styles": "[1]", "gunping_num": "4", "gunping_speed": "1",
			},
		},
		{
			name: "任务项覆盖配置档", platform: 6, gender: "男", style: "男频都市",
			advanced: map[string]any{"ziti": 3, "tl5": 1, "organization": "org-a", "fontColorStyles": []any{2, "3"}, "jieyaSpeed": "1.2"},
			item:     map[string]any{"ziti": 5, "organization": "org-b"},
			want: map[string]string{
				"platform_id": "6", "organization": "org-b", "gender": "1", "style": "305",
				"tl5": "1", "ziti": "5", "zitidx": "62", "biaohong": "", "biaohong_reuse": "", "keywords": "",
				"jieya_num": "4", "jieya_ai_head": "0", "jieya_speed": "1.2", "jieya_pitch": "0",
				"font_color_styles": "[2,3]", "gunping_num": "4", "gunping_speed": "1",
			},
		},
		{
			name: "越界值回退默认", platform: 1, gender: "女", style: "古风通用",
			advanced: map[string]any{"ziti": 9, "zitidx": 10, "jieyaNum": 99, "jieyaAiHead": 5, "gunpingNum": -1},
			want: map[string]string{
				"platform_id": "1", "organization": "1", "gender": "2", "style": "103",
				"tl5": "0", "ziti": "1", "zitidx": "62", "biaohong": "", "biaohong_reuse": "", "keywords": "",
				"jieya_num": "4", "jieya_ai_head": "0", "jieya_speed": "1.7", "jieya_pitch": "0",
				"font_color_styles": "[1]", "gunping_num": "4", "gunping_speed": "1",
			},
		},
		{
			name: "空字体颜色回退", platform: 2, gender: "女", style: "现代甜文",
			advanced: map[string]any{"fontColorStyles": []any{}},
			want: map[string]string{
				"platform_id": "2", "organization": "1", "gender": "2", "style": "302",
				"tl5": "0", "ziti": "1", "zitidx": "62", "biaohong": "", "biaohong_reuse": "", "keywords": "",
				"jieya_num": "4", "jieya_ai_head": "0", "jieya_speed": "1.7", "jieya_pitch": "0",
				"font_color_styles": "[1]", "gunping_num": "4", "gunping_speed": "1",
			},
		},
		{name: "未知男女频", platform: 1, gender: "未知", style: "现代悬疑", wantErr: true},
		{name: "未知风格", platform: 1, gender: "女", style: "未知", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildUploadFieldsCompat(tc.platform, tc.gender, tc.style, tc.advanced, tc.item)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %#v", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("fields mismatch\n got: %#v\nwant: %#v", got, tc.want)
			}
		})
	}
}
