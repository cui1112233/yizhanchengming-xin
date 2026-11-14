package prompt

import (
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
)

// All 12 character and 6 scene entries, isolated from the original database on 2026-11-14.
//
//go:embed screenplay_typed_prompts.json
var screenplayTypedCorpus []byte

type screenplayTypedRecord struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"prompt_type"`
	Content string `json:"content"`
	SHA256  string `json:"sha256"`
}

var screenplayTypedOnce sync.Once
var screenplayTypedRecords []screenplayTypedRecord
var screenplayTypedError error

func screenplayTypedEntries() ([]screenplayTypedRecord, error) {
	screenplayTypedOnce.Do(func() {
		if screenplayTypedError = json.Unmarshal(screenplayTypedCorpus, &screenplayTypedRecords); screenplayTypedError != nil {
			return
		}
		for _, r := range screenplayTypedRecords {
			if fmt.Sprintf("%x", sha256.Sum256([]byte(r.Content))) != r.SHA256 {
				screenplayTypedError = fmt.Errorf("原版出图模板%d摘要不一致", r.ID)
				return
			}
		}
	})
	return screenplayTypedRecords, screenplayTypedError
}
func typedCreativeID(r screenplayTypedRecord) string {
	if r.Type == "人设设计" {
		return "luoshui_character_design_" + strconv.Itoa(r.ID)
	}
	return "luoshui_scene_design_" + strconv.Itoa(r.ID)
}
func CreativePromptAlias(id string) string {
	if recovered := recoveredCreativeAlias(id); recovered != id {
		return recovered
	}
	records, e := screenplayTypedEntries()
	if e == nil {
		for _, r := range records {
			if strconv.Itoa(r.ID) == id {
				return typedCreativeID(r)
			}
		}
	}
	switch id {
	case "144":
		return CreativeScript144
	case "307":
		return CreativeStoryboard307
	}
	return id
}
func resolveTypedCreativePrompt(id, purpose string) (CreativePromptSnapshot, bool, error) {
	records, e := screenplayTypedEntries()
	if e != nil {
		return CreativePromptSnapshot{}, false, e
	}
	for _, r := range records {
		if typedCreativeID(r) != id {
			continue
		}
		kind := "character-design"
		if r.Type == "场景设计" {
			kind = "scene-design"
		}
		if kind != purpose {
			return CreativePromptSnapshot{}, true, fmt.Errorf("提示词%s不适用于%s", id, purpose)
		}
		return CreativePromptSnapshot{ID: id, Name: r.Name, Type: kind, Version: r.SHA256[:16], Body: r.Content, Digest: r.SHA256}, true, nil
	}
	return CreativePromptSnapshot{}, false, nil
}
