package prompt

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestStoryboardSceneMessageMatchesCapturedZIPCalls(t *testing.T) {
	raw, err := os.ReadFile("testdata/storyboard_scene_zip.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Original string          `json:"original"`
		Scene    json.RawMessage `json:"scene"`
		Expected string          `json:"expected_user"`
	}
	if err = json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		got, err := StoryboardSceneUserMessage(row.Original, string(row.Scene))
		if err != nil {
			t.Fatal(err)
		}
		split := func(s string) (string, any, string) {
			before, rest, ok := strings.Cut(s, "【当前剧本场景JSON（仅1个场景）】\n")
			if !ok {
				t.Fatal("missing scene marker")
			}
			j, after, ok := strings.Cut(rest, "\n\n【处理要求】")
			if !ok {
				t.Fatal("missing requirements marker")
			}
			var v any
			d := json.NewDecoder(strings.NewReader(j))
			d.UseNumber()
			if err := d.Decode(&v); err != nil {
				t.Fatal(err)
			}
			return before, v, after
		}
		gb, gj, ga := split(got)
		wb, wj, wa := split(row.Expected)
		if gb != wb || ga != wa || !reflect.DeepEqual(gj, wj) {
			t.Fatalf("differs from captured ZIP request\ngot=%s\nwant=%s", got, row.Expected)
		}
	}
}

func TestStoryboardSceneMessageKeepsLargeIDsAndLiteralPlainSource(t *testing.T) {
	got, err := StoryboardSceneUserMessage("原文", `{"scene_number":9007199254740993,"content":"<窗>&文字","script_content":[{"element_id":"ID-9"}],"_internal":"hidden"}`)
	if err != nil || !strings.Contains(got, "9007199254740993") || !strings.Contains(got, "<窗>&文字") || !strings.Contains(got, "ID-9") || strings.Contains(got, "hidden") {
		t.Fatal(got, err)
	}
	for _, text := range []string{"直接输入\n第二行", `[{"content":"不是单场对象"}]`, `{未完成的普通文字`} {
		got, err = StoryboardSceneUserMessage("整章", text)
		if err != nil {
			t.Fatal(err)
		}
		raw := strings.Split(strings.Split(got, "【当前剧本场景JSON（仅1个场景）】\n")[1], "\n\n【处理要求】")[0]
		var scene map[string]string
		if err = json.Unmarshal([]byte(raw), &scene); err != nil || scene["content"] != text {
			t.Fatal("plain text changed", got, err)
		}
	}
}

func TestStoryboardSceneMessageRejectsEmptyReferenceOrMetadataOnlyScene(t *testing.T) {
	for _, in := range [][2]string{{" ", "正文"}, {"原文", " "}, {"原文", `{}`}, {"原文", `{"_source_scene_index":2,"_internal":"routing only"}`}} {
		if _, err := StoryboardSceneUserMessage(in[0], in[1]); err == nil {
			t.Fatal("empty scene/reference accepted", in)
		}
	}
}
