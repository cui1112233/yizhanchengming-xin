package novelfetchworkshop

import "testing"

func damagedImportedSensitive() map[string]any {
	library := readImportedSeed("sensitive.json")
	for _, group := range arrayMap(library["groups"]) {
		group["name"] = questionMask(group["name"].(string))
		for _, row := range arrayMap(group["rules"]) {
			row["find"] = questionMask(row["find"].(string))
			row["replace"] = "????"
		}
	}
	return map[string]any{"knowledgeLibraries": map[string]any{"sensitive": library}}
}

func TestRepairImportedSensitiveRequiresWholeCatalogMatchAndKeepsBackup(t *testing.T) {
	s := damagedImportedSensitive()
	group := arrayMap(knowledgeLibrary(s, "sensitive")["groups"])[0]
	arrayMap(group["rules"])[0]["enabled"] = false
	n, err := RepairImportedSensitive(s)
	if err != nil || n < 800 {
		t.Fatalf("not repaired: %d %v", n, err)
	}
	rows := arrayMap(arrayMap(knowledgeLibrary(s, "sensitive")["groups"])[0]["rules"])
	if rows[0]["find"] != "花魁" || rows[0]["enabled"] != false || onlyQuestions(rows[0]["replace"].(string)) {
		t.Fatal("content or switches lost")
	}
	if s["importedSensitiveRepairBackup"] == nil {
		t.Fatal("no backup")
	}
	if again, _ := RepairImportedSensitive(s); again != 0 {
		t.Fatal("repair is not idempotent")
	}
	custom := damagedImportedSensitive()
	arrayMap(arrayMap(knowledgeLibrary(custom, "sensitive")["groups"])[0]["rules"])[10]["find"] = "自定义词"
	if n, _ := RepairImportedSensitive(custom); n != 0 {
		t.Fatal("custom group guessed at")
	}
}
