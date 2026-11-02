package httpapi

import (
	"encoding/json"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
)

func TestBatchFactoryAgentProjectTypeFollowsComicProductionMode(t *testing.T) {
	comic, _ := json.Marshal("comic")
	video, _ := json.Marshal("video")
	if got := batchFactoryAgentProjectType(batchfactoryv11.Batch{SettingsState: batchfactoryv11.SettingsState{Patch: batchfactoryv11.SettingsPatch{"productionMode": comic}}}); got != "comic" {
		t.Fatalf("comic batch project type=%q", got)
	}
	if got := batchFactoryAgentProjectType(batchfactoryv11.Batch{SettingsState: batchfactoryv11.SettingsState{Patch: batchfactoryv11.SettingsPatch{"productionMode": video}}}); got != "video" {
		t.Fatalf("video batch project type=%q", got)
	}
	if got := batchFactoryAgentProjectType(batchfactoryv11.Batch{Books: []batchfactoryv11.Book{{SettingsState: batchfactoryv11.SettingsState{Patch: batchfactoryv11.SettingsPatch{"productionMode": comic}}}}}); got != "comic" {
		t.Fatalf("comic book override project type=%q", got)
	}
}
