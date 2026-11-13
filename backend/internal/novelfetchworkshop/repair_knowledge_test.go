package novelfetchworkshop

import (
	"encoding/json"
	"strings"
	"testing"
)

func corruptImportedText(value any) any {
	switch v := value.(type) {
	case string:
		return questionMask(v)
	case []any:
		for i := range v {
			v[i] = corruptImportedText(v[i])
		}
	case map[string]any:
		for key := range v {
			if key != "id" && key != "source" && key != "imported_at" && key != "updated_at" {
				v[key] = corruptImportedText(v[key])
			}
		}
	}
	return value
}

func TestRepairImportedKnowledgePreservesCustomContentFlagsAndCreatesPrivateBackup(t *testing.T) {
	s := DefaultWorkshopSettings()
	corruptImportedText(s["knowledgeLibraries"])
	high := knowledgeLibrary(s, "high_imitation")
	row := arrayMap(high["prompts"])[0]
	row["enabled"] = false
	row["custom_note"] = "用户自己的内容?"
	row["system_prompt"] = "用户独立编写的提示词?"
	before, _ := json.Marshal(s)
	report := RepairImportedKnowledge(s, false)
	after, _ := json.Marshal(s)
	if string(before) != string(after) || report.Repaired < 100 {
		t.Fatalf("dry run mutated or did not detect: %+v", report)
	}
	report = RepairImportedKnowledge(s, true)
	row = arrayMap(knowledgeLibrary(s, "high_imitation")["prompts"])[0]
	if row["name"] != "高仿文章" || row["enabled"] != false || row["system_prompt"] != "用户独立编写的提示词?" || row["custom_note"] != "用户自己的内容?" {
		t.Fatal("custom content or switches overwritten")
	}
	if s["importedKnowledgeRepairBackup"] == nil {
		t.Fatal("missing backup")
	}
	if next := RepairImportedKnowledge(s, true); next.Repaired != 0 {
		t.Fatalf("not idempotent: %+v", next)
	}
}

func TestRepairImportedKnowledgeRejectsForeignSourceAndChangedRecordIdentity(t *testing.T) {
	s := DefaultWorkshopSettings()
	corruptImportedText(s["knowledgeLibraries"])
	high := knowledgeLibrary(s, "high_imitation")
	high["source"] = "用户自建"
	opening := knowledgeLibrary(s, "opening_phrases")
	rows := arrayMap(opening["items"])
	rows[0]["id"] = "user-other-record"
	content, _ := rows[0]["content"].(string)
	RepairImportedKnowledge(s, true)
	if arrayMap(high["prompts"])[0]["name"] == "高仿文章" {
		t.Fatal("foreign source changed")
	}
	if rows[0]["content"] != content {
		t.Fatal("changed identity overwritten")
	}
}

func TestDegradedImportedTextRequiresExactPositionAndLength(t *testing.T) {
	if !degradedImportedText("�???A", "“中文”A") {
		t.Fatal("replacement character not recovered")
	}
	for _, value := range []string{"?A", "???B", "自定义?", strings.Repeat("?", 20)} {
		if degradedImportedText(value, "中文A") {
			t.Fatalf("guessed at %q", value)
		}
	}
}
