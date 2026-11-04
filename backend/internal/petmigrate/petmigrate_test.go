package petmigrate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, usersDir, username, body string) string {
	t.Helper()
	dir := filepath.Join(usersDir, username)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, ConfigFileName)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func readPet(t *testing.T, path string) any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := json.Unmarshal(raw, &config); err != nil {
		t.Fatal(err)
	}
	return config["pet"]
}

func TestScanClassifiesPetShapes(t *testing.T) {
	usersDir := t.TempDir()
	writeConfig(t, usersDir, "legacy-string", `{"pet":"stacky"}`)
	writeConfig(t, usersDir, "legacy-object", `{"pet":{"id":"pixiu","displayName":"貔貅","spriteVersionNumber":2}}`)
	writeConfig(t, usersDir, "unset", `{"provider":"openai"}`)
	writeConfig(t, usersDir, "already", `{"pet":"little-companion"}`)
	writeConfig(t, usersDir, "broken", `{not-json`)

	records, err := Scan("local", usersDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 5 {
		t.Fatalf("records = %d, want 5", len(records))
	}
	byUser := map[string]Record{}
	for _, record := range records {
		byUser[record.Username] = record
	}
	cases := []struct {
		user    string
		kind    string
		id      string
		changed bool
		after   string
	}{
		{"legacy-string", KindString, "stacky", false, "stacky"},
		{"legacy-object", KindObject, "pixiu", false, "pixiu"},
		{"unset", KindMissing, "", false, DefaultPetID},
		{"already", KindString, DefaultPetID, false, DefaultPetID},
		{"broken", KindInvalid, "", false, ""},
	}
	for _, item := range cases {
		record := byUser[item.user]
		if record.BeforeKind != item.kind || record.BeforeID != item.id || record.Changed != item.changed || record.AfterID != item.after {
			t.Fatalf("%s = %#v, want kind=%s id=%s changed=%v after=%s", item.user, record, item.kind, item.id, item.changed, item.after)
		}
	}
	if byUser["broken"].Err == "" {
		t.Fatal("invalid JSON must be reported instead of migrated")
	}

	legacy, err := ScanWithPolicy("local", usersDir, PolicyRewriteAll)
	if err != nil {
		t.Fatal(err)
	}
	legacyByUser := map[string]Record{}
	for _, record := range legacy {
		legacyByUser[record.Username] = record
	}
	if !legacyByUser["legacy-string"].Changed || !legacyByUser["legacy-object"].Changed {
		t.Fatalf("historical rewrite policy must still flag explicit choices: %#v", legacyByUser)
	}
	if legacyByUser["unset"].Changed || legacyByUser["already"].Changed {
		t.Fatalf("historical rewrite policy must skip unset/already-default: %#v", legacyByUser)
	}
}

func TestDefaultPolicyNeedsNoMigration(t *testing.T) {
	usersDir := t.TempDir()
	backupRoot := t.TempDir()
	stackyPath := writeConfig(t, usersDir, "stacky-user", `{"pet":"stacky","model":"gpt-4o-mini"}`)
	objectPath := writeConfig(t, usersDir, "object-user", `{"pet":{"id":"pixiu","displayName":"貔貅"}}`)
	originalStacky, _ := os.ReadFile(stackyPath)

	records, err := Scan("local", usersDir)
	if err != nil {
		t.Fatal(err)
	}
	backupPath, migrated, err := Migrate(records, backupRoot, "run-1", false)
	if err != nil {
		t.Fatal(err)
	}
	if migrated != 0 || backupPath != "" {
		t.Fatalf("preserve policy migrated=%d backup=%q, want 0 and empty", migrated, backupPath)
	}
	if !bytesEqual(t, mustRead(t, stackyPath), originalStacky) {
		t.Fatal("preserve policy must not rewrite stacky config")
	}
	if pet := readPet(t, objectPath); pet == DefaultPetID {
		t.Fatal("preserve policy must not rewrite object config")
	}
}

func bytesEqual(t *testing.T, left, right []byte) bool {
	t.Helper()
	return string(left) == string(right)
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestMigrateRewritesPetAndKeepsBackup(t *testing.T) {
	usersDir := t.TempDir()
	backupRoot := t.TempDir()
	stringPath := writeConfig(t, usersDir, "legacy-string", `{"pet":"stacky","model":"gpt-4o-mini"}`)
	objectPath := writeConfig(t, usersDir, "legacy-object", `{"pet":{"id":"pixiu","displayName":"貔貅"}}`)
	alreadyPath := writeConfig(t, usersDir, "already", `{"pet":"little-companion"}`)
	unsetPath := writeConfig(t, usersDir, "unset", `{"model":"gpt-4o-mini"}`)
	originalString, _ := os.ReadFile(stringPath)

	records, err := ScanWithPolicy("local", usersDir, PolicyRewriteAll)
	if err != nil {
		t.Fatal(err)
	}
	backupPath, migrated, err := Migrate(records, backupRoot, "20261104-120000", false)
	if err != nil {
		t.Fatal(err)
	}
	if migrated != 2 {
		t.Fatalf("migrated = %d, want 2", migrated)
	}
	if backupPath != filepath.Join(backupRoot, "20261104-120000") {
		t.Fatalf("backupPath = %s", backupPath)
	}
	if pet := readPet(t, stringPath); pet != DefaultPetID {
		t.Fatalf("string pet = %#v, want %s", pet, DefaultPetID)
	}
	if pet := readPet(t, objectPath); pet != DefaultPetID {
		t.Fatalf("object pet = %#v, want %s", pet, DefaultPetID)
	}
	if pet := readPet(t, alreadyPath); pet != DefaultPetID {
		t.Fatalf("already pet = %#v", pet)
	}
	if pet := readPet(t, unsetPath); pet != nil {
		t.Fatalf("unset pet must stay unset, got %#v", pet)
	}
	backedUp, err := os.ReadFile(filepath.Join(backupPath, "legacy-string", ConfigFileName))
	if err != nil {
		t.Fatal(err)
	}
	if string(backedUp) != string(originalString) {
		t.Fatalf("backup bytes differ from original")
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	usersDir := t.TempDir()
	backupRoot := t.TempDir()
	writeConfig(t, usersDir, "legacy", `{"pet":"pixiu"}`)
	records, err := ScanWithPolicy("local", usersDir, PolicyRewriteAll)
	if err != nil {
		t.Fatal(err)
	}
	if _, migrated, err := Migrate(records, backupRoot, "run-1", false); err != nil || migrated != 1 {
		t.Fatalf("first run migrated=%d err=%v", migrated, err)
	}
	again, err := ScanWithPolicy("local", usersDir, PolicyRewriteAll)
	if err != nil {
		t.Fatal(err)
	}
	backupPath, migrated, err := Migrate(again, backupRoot, "run-2", false)
	if err != nil {
		t.Fatal(err)
	}
	if migrated != 0 || backupPath != "" {
		t.Fatalf("second run migrated=%d backup=%q, want 0 and empty", migrated, backupPath)
	}
	if pet := readPet(t, filepath.Join(usersDir, "legacy", ConfigFileName)); pet != DefaultPetID {
		t.Fatalf("pet after repeated run = %#v", pet)
	}
}

func TestMarkdownAndCSVDescribeMigration(t *testing.T) {
	usersDir := t.TempDir()
	path := writeConfig(t, usersDir, "legacy", `{"pet":"stacky"}`)
	records, err := ScanWithPolicy("local", usersDir, PolicyRewriteAll)
	if err != nil {
		t.Fatal(err)
	}
	markdown := Markdown(records, "/tmp/backup/20261104-120000", false, 1)
	for _, want := range []string{"2026-11-04 桌宠默认值迁移清单", "已作废：用户决定仅对未设置用户生效，本次迁移已回滚，生产未执行", "本次实际迁移条数：1", "/tmp/backup/20261104-120000", "回滚", "幂等"} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("markdown missing %q", want)
		}
	}
	csvText := CSV(records)
	if !strings.Contains(csvText, "source,username,path,before_kind,before_pet_id,after_pet_id,action,error") {
		t.Fatalf("csv header missing: %s", csvText)
	}
	if !strings.Contains(csvText, "legacy") || !strings.Contains(csvText, path) {
		t.Fatalf("csv row missing: %s", csvText)
	}
}
