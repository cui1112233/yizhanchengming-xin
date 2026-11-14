package prompt

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestScreenplayAssetCheckMatchesSixCapturedOriginalRequests(t *testing.T) {
	// Golden digests come from the original compiled worker at the fake GPT
	// client boundary, not from this Go builder or its templates.
	cases := []struct{ kind, mode, system, user string }{
		{"character", "general", "a967638678f77b29a9862aefb7b17860ede6bbf497c323216a8b101d697df445", "59287e3f788966918854b6795a524154ccd37b91f8713315f69044c3a8b96b71"},
		{"scene", "general", "3b90877751890870f1b67858953e030721244f8f1aa96c8fc046a17bf0e67923", "b6c132747377e84f06827eb3e26c7a87ad070faf5a9254294f4fd8f26a911452"},
		{"prop", "general", "1e1d9027a2c54037d3a0c069821486f6d910b31bb2130b39855e4fbfd32a7ede", "1f85d0252ca3161bce9f03008f310e6da3a0b1be3605721910da81b9ce7187a1"},
		{"character", "western", "f5cefd6af9efafdfc12191751ea1cdb63075a5147c741a4684df694ec97905af", "6dd7ce31a44db4dfdc11ccafd0f5f35c2e68e0c9639645710f46743e973db1be"},
		{"scene", "western", "e68e4e96f6c5571e2d85e4ef9cd49da7a951b8e4cb135cfcefd3aa8c98ed20a8", "9352d44fa53324f901d44a3b7adbc4606a690e4e31725ce3219e0146351b44ee"},
		{"prop", "western", "299ec7d4eecbcd54f272faf83a7ecb433b90b8d6e963c4e747f4c70c3abdf647", "151abbe3f56154433733522d610f9519e6bdc70a00e9604d75914506fd157e59"},
	}
	content := "CM离线剧本：已有甲在已有书房拿起已有灯。新角色乙进入茶馆并放下竹笛。"
	for _, tc := range cases {
		t.Run(tc.mode+"-"+tc.kind, func(t *testing.T) {
			got, err := BuildScreenplayAssetCheckPrompt(tc.kind, tc.mode, content, []string{"已有甲"}, []string{"已有书房"}, []string{"已有灯"})
			if err != nil {
				t.Fatal(err)
			}
			if got.SystemDigest != tc.system || got.UserDigest != tc.user {
				t.Fatalf("original request differs: system=%s user=%s", got.SystemDigest, got.UserDigest)
			}
			if got.OutputKey != map[string]string{"character": "characters", "scene": "scenes", "prop": "props"}[tc.kind] || got.ProtocolVersion != ScreenplayAssetCheckProtocolVersion {
				t.Fatal(got.OutputKey, got.ProtocolVersion)
			}
			if strings.Contains(got.SystemPrompt, "完整检查模式附加规则") {
				t.Fatal("batch check must use parse base, not complete-check suffix")
			}
		})
	}
}

func readAssetCheckFixture(t *testing.T, name string, dst any) {
	t.Helper()
	raw, err := os.ReadFile("assetcheck/testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, dst); err != nil {
		t.Fatal(err)
	}
}

func TestScreenplayAssetCheckNameCleaningMatchesOriginalPureHelper(t *testing.T) {
	var cases []struct{ Input, Output string }
	readAssetCheckFixture(t, "name-clean-cases.json", &cases)
	for _, tc := range cases {
		if got := cleanScreenplayAssetCheckName(tc.Input); got != tc.Output {
			t.Fatalf("%q => %q want %q", tc.Input, got, tc.Output)
		}
	}
}

func TestScreenplayAssetCheckNameFormattingMatchesOriginalPureHelper(t *testing.T) {
	var cases []struct {
		Variant, Mode string
		Inputs        struct{ Characters, Scenes, Props []any }
		Output        []json.RawMessage
	}
	readAssetCheckFixture(t, "name-format-cases.json", &cases)
	names := func(items []any) []string {
		out := []string{}
		for _, item := range items {
			switch value := item.(type) {
			case string:
				out = append(out, value)
			case map[string]any:
				if name, ok := value["name"].(string); ok {
					out = append(out, name)
				}
			}
		}
		return out
	}
	for _, tc := range cases {
		got, err := BuildScreenplayAssetCheckPrompt("character", tc.Mode, "CM固定合成正文", names(tc.Inputs.Characters), names(tc.Inputs.Scenes), names(tc.Inputs.Props))
		if err != nil {
			t.Fatal(err)
		}
		var want string
		if err = json.Unmarshal(tc.Output[0], &want); err != nil {
			t.Fatal(err)
		}
		if got.UserPrompt != want {
			t.Fatalf("%s/%s differs from original pure helper", tc.Variant, tc.Mode)
		}
	}
}

func TestScreenplayAssetCheckContentPreservesWhitespaceAndLiteralMarkers(t *testing.T) {
	var cases []struct {
		Mode, Content string
		Output        []json.RawMessage
	}
	readAssetCheckFixture(t, "content-format-cases.json", &cases)
	for _, tc := range cases {
		got, err := BuildScreenplayAssetCheckPrompt("character", tc.Mode, tc.Content, nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		var want string
		if err = json.Unmarshal(tc.Output[0], &want); err != nil {
			t.Fatal(err)
		}
		if got.UserPrompt != want {
			t.Fatalf("%s changed source or expanded user markers", tc.Mode)
		}
	}
}

func TestScreenplayAssetCheckRejectsInvalidInputs(t *testing.T) {
	for _, tc := range []struct{ kind, mode, content string }{{"unknown", "general", "正文"}, {"character", "unknown", "正文"}, {"character", "general", " \t\n"}} {
		if _, err := BuildScreenplayAssetCheckPrompt(tc.kind, tc.mode, tc.content, nil, nil, nil); err == nil {
			t.Fatal("invalid input accepted", tc)
		}
	}
	a, _ := BuildScreenplayAssetCheckPrompt("scene", "", "正文", nil, nil, nil)
	b, _ := BuildScreenplayAssetCheckPrompt("scene", "general", "正文", nil, nil, nil)
	if a != b {
		t.Fatal("empty mode must equal general")
	}
}
