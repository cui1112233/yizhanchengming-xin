package batchfactoryv11

import (
	"encoding/json"
	"testing"
)

// 23 版需求 4：单书只覆盖 publishSettings 的一个键时，不得丢掉批次层的
// websiteProfileId / organization 等键（否则发布回落 4/4 默认）。
func TestV23ResolveSettingsDeepMergesPublishSettings(t *testing.T) {
	batchPatch := SettingsPatch{
		"publishSettings": json.RawMessage(`{"websiteProfileId":"121-420","organization":"1","materialReuse":false}`),
	}
	bookPatch := SettingsPatch{
		"publishSettings": json.RawMessage(`{"materialReuse":true}`),
	}

	resolved := ResolveSettings(batchPatch, bookPatch)
	var publish map[string]any
	if err := json.Unmarshal(resolved["publishSettings"], &publish); err != nil {
		t.Fatalf("publishSettings unmarshal: %v (raw=%s)", err, resolved["publishSettings"])
	}
	if got := publish["websiteProfileId"]; got != "121-420" {
		t.Fatalf("websiteProfileId = %v, want 121-420", got)
	}
	if got := publish["organization"]; got != "1" {
		t.Fatalf("organization = %v, want 1", got)
	}
	if got := publish["materialReuse"]; got != true {
		t.Fatalf("materialReuse = %v, want true (book must win)", got)
	}
}

// 23 版需求 4：深合并不能改变既有优先级——同键单书覆盖批次，
// 非 publishSettings 键仍然整键覆盖。
func TestV23ResolveSettingsKeepsBookPriority(t *testing.T) {
	batchPatch := SettingsPatch{
		"publishSettings": json.RawMessage(`{"organization":"1","materialReuse":false}`),
		"productionMode":  json.RawMessage(`"video"`),
	}
	bookPatch := SettingsPatch{
		"publishSettings": json.RawMessage(`{"organization":"2"}`),
	}

	resolved := ResolveSettings(batchPatch, bookPatch)
	var publish map[string]any
	if err := json.Unmarshal(resolved["publishSettings"], &publish); err != nil {
		t.Fatalf("publishSettings unmarshal: %v", err)
	}
	if got := publish["organization"]; got != "2" {
		t.Fatalf("organization = %v, want 2 (book wins)", got)
	}
	if got := publish["materialReuse"]; got != false {
		t.Fatalf("materialReuse = %v, want false (batch kept)", got)
	}
	if got := string(resolved["productionMode"]); got != `"video"` {
		t.Fatalf("productionMode = %s, want \"video\"", got)
	}
}
