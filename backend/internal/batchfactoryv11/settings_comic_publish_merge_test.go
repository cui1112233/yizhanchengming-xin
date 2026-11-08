package batchfactoryv11

import (
	"encoding/json"
	"testing"
)

func TestResolveSettingsDeepMergesComicPublishSettings(t *testing.T) {
	batchPatch := SettingsPatch{"comicPublishSettings": json.RawMessage(`{"organizationId":"1","producerAdminName":"制作者","audioCategorySelectionMode":"auto"}`)}
	bookPatch := SettingsPatch{"comicPublishSettings": json.RawMessage(`{"audioCategorySelectionMode":"manual","audioCategoryId":14,"specifiedAudioId":199}`)}
	resolved := ResolveSettings(batchPatch, bookPatch)
	var publish map[string]any
	if err := json.Unmarshal(resolved["comicPublishSettings"], &publish); err != nil {
		t.Fatal(err)
	}
	if publish["organizationId"] != "1" || publish["producerAdminName"] != "制作者" {
		t.Fatalf("batch fields were lost: %+v", publish)
	}
	if publish["audioCategorySelectionMode"] != "manual" || publish["audioCategoryId"] != float64(14) || publish["specifiedAudioId"] != float64(199) {
		t.Fatalf("book manual override was not preserved: %+v", publish)
	}
}
