package prompt

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
)

func TestStoryboardCatalogSeparates41OriginalRowsAndKeepsRealSourceDates(t *testing.T) {
	original, compat, zipOriginal := 0, 0, 0
	dates := map[int]string{137: "2026-03-04 18:21:53", 268: "2026-11-09 11:39:38", 307: "2026-11-10 15:08:11"}
	for _, item := range CreativePromptOptions() {
		if item.Type != "storyboard" {
			continue
		}
		raw, _ := json.Marshal(item)
		var fields map[string]any
		json.Unmarshal(raw, &fields)
		group, _ := fields["catalog_group"].(string)
		if group == "zip-original" {
			zipOriginal++
			if item.SourceID != 164 {
				t.Fatalf("unexpected ZIP supplement %d", item.SourceID)
			}
			continue
		}
		if item.SourceID > 0 {
			original++
			if group != "original" {
				t.Fatalf("source %d missing original catalog classification: %v", item.SourceID, fields["catalog_group"])
			}
			if fields["created_at"] == nil || fields["created_at"] == "" || fields["updated_at"] == nil || fields["updated_at"] == "" {
				t.Fatalf("source %d has no recovered dates", item.SourceID)
			}
			if want := dates[item.SourceID]; want != "" && (fields["created_at"] != want || fields["updated_at"] != want) {
				t.Fatalf("source %d date got %v/%v want original TEXT %s", item.SourceID, fields["created_at"], fields["updated_at"], want)
			}
		} else {
			compat++
			if group != "compatibility" {
				t.Fatalf("legacy %s incorrectly mixed with original catalog", item.ID)
			}
		}
	}
	if original != 41 || compat != 2 || zipOriginal != 1 {
		t.Fatalf("original=%d compatibility=%d zip-original=%d", original, compat, zipOriginal)
	}
	for _, id := range []string{"storyboard_xiaoluo_v7_15s", "storyboard_xiaoyunque_v3"} {
		p, err := ResolveCreativePrompt(id, "storyboard")
		if err != nil || p.Body == "" {
			t.Fatal("existing legacy project cannot resolve", id, err)
		}
	}
}

func TestRecoveredSourceDatesRemainPinnedAndSnapshotsRetainDates(t *testing.T) {
	raw, err := recoveredCreativeFiles.ReadFile("screenplay_recovered_dates.json")
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "02757c3a10b9a8b1c7cac98ce82d9a3bec11455c404408e8b7d961f943017f4a" {
		t.Fatal("reviewed source timestamps changed", got)
	}
	rows, err := RecoveredCreativePromptCatalog()
	if err != nil {
		t.Fatal(err)
	}
	disabled := map[int]bool{}
	for _, row := range rows {
		p, err := ResolveCreativePrompt(row.ID, row.Type)
		if err != nil {
			t.Fatal(err)
		}
		if p.CreatedAt != row.CreatedAt || p.UpdatedAt != row.UpdatedAt || p.CatalogGroup != "original" {
			t.Fatal("snapshot metadata lost", row.ID)
		}
		if row.Type == "storyboard" && !row.Supported {
			disabled[row.SourceID] = true
		}
	}
	if len(disabled) != 1 || !disabled[138] {
		t.Fatal("unsupported template guards changed", disabled)
	}
}
