package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func characterText(value any) string { return strings.TrimSpace(toString(value)) }

func characterSlotToken(entry string, index int) string {
	hash := sha256.Sum256([]byte(strings.TrimSpace(entry) + "\n" + toString(index)))
	return hex.EncodeToString(hash[:])[:24]
}

func parseCharacterRoster(value string) []map[string]any {
	bracketPattern := regexp.MustCompile(`[（(]([^()（）]{1,120})[）)]`)
	entries := []string{}
	current := strings.Builder{}
	depth := 0
	flush := func() {
		item := strings.TrimSpace(current.String())
		if item != "" {
			entries = append(entries, item)
		}
		current.Reset()
	}
	for _, r := range value {
		switch r {
		case '(', '（', '[':
			depth++
			current.WriteRune(r)
		case ')', '）', ']':
			if depth > 0 {
				depth--
			}
			current.WriteRune(r)
		default:
			if depth == 0 && strings.ContainsRune("\n,，、;；|", r) {
				flush()
			} else {
				current.WriteRune(r)
			}
		}
	}
	flush()
	result := make([]map[string]any, 0, len(entries))
	seen := map[string]bool{}
	for _, entry := range entries {
		if seen[entry] {
			continue
		}
		seen[entry] = true
		base := strings.TrimSpace(entry)
		hints := []string{}
		for _, match := range bracketPattern.FindAllStringSubmatch(base, -1) {
			if len(match) > 1 && strings.TrimSpace(match[1]) != "" {
				hints = append(hints, strings.TrimSpace(match[1]))
			}
		}
		base = strings.TrimSpace(bracketPattern.ReplaceAllString(base, ""))
		if base == "" {
			continue
		}
		token := characterSlotToken(entry, len(result)+1)
		stage := ""
		if len(hints) > 0 {
			stage = hints[0]
		}
		result = append(result, map[string]any{"slot_id": token, "slot_token": token, "person_id": "person_" + token, "source_entry": entry, "display_name": entry, "base_name": base, "canonical_name_hint": base, "aliases": []any{}, "bracket_hints": hints, "identity_hints": []any{}, "gender": "待确认", "species": "", "age": map[string]any{"chronological_age": "", "visual_age_stage": stage, "life_stage": stage, "timeline_stage": "", "stage_source": map[bool]string{true: "forced_roster_bracket", false: "unset"}[stage != ""]}, "manual_values": map[string]any{}, "inferred_values": map[string]any{}, "appearance": "", "appearance_revision_note": "", "disabled": false})
	}
	return result
}

func RegisterNovelPanelCharacterCompat(root *http.ServeMux, middleware BridgeAuth) {
	if root == nil || middleware.AuthStore == nil {
		return
	}
	handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if _, ok := BridgeIdentityFromContext(req.Context()); !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input map[string]any
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 8<<20)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		switch req.URL.Path {
		case "/api/novel-panel/character-core/parse-slots":
			guide := characterText(input["guide_text"])
			slots := parseCharacterRoster(guide)
			sourceHash := characterText(input["source_hash"])
			if sourceHash == "" {
				hash := sha256.Sum256([]byte(guide))
				sourceHash = hex.EncodeToString(hash[:])[:24]
			}
			writeJSON(w, http.StatusOK, map[string]any{"slots": slots, "slot_count": len(slots), "source_hash": sourceHash})
			return
		case "/api/novel-panel/character-core/resolve-scene-cast":
			source := characterText(input["source_text"])
			selectedIDs := map[string]bool{}
			if values, ok := input["manual_selected_slot_ids"].([]any); ok {
				for _, value := range values {
					selectedIDs[characterText(value)] = true
				}
			}
			selected, mentioned := []any{}, []any{}
			if slots, ok := input["slots"].([]any); ok {
				for _, raw := range slots {
					slot, ok := raw.(map[string]any)
					if !ok {
						continue
					}
					id := characterText(slot["slot_id"])
					if id == "" {
						id = characterText(slot["slot_token"])
					}
					names := []string{characterText(slot["base_name"]), characterText(slot["display_name"]), characterText(slot["canonical_name_hint"])}
					matched := ""
					for _, name := range names {
						if name != "" && strings.Contains(source, name) {
							matched = name
							break
						}
					}
					item := map[string]any{"slot_id": id, "slot_token": characterText(slot["slot_token"]), "person_id": characterText(slot["person_id"]), "display_name": characterText(slot["display_name"]), "scene_role": "mentioned_only", "evidence": []string{}, "confidence": 0}
					if matched != "" || selectedIDs[id] {
						item["scene_role"] = "visible"
						item["confidence"] = 1
						if matched != "" {
							item["evidence"] = []string{"原文包含「" + matched + "」"}
						}
						selected = append(selected, item)
					} else {
						mentioned = append(mentioned, item)
					}
				}
			}
			writeJSON(w, http.StatusOK, map[string]any{"scene_id": characterText(input["scene_id"]), "selected": selected, "mentioned_only": mentioned, "offscreen_voice": []any{}, "ambiguous": []any{}})
			return
		case "/api/novel-panel/character-core/build-scene-context":
			cast, _ := input["cast"].(map[string]any)
			selected, _ := cast["selected"].([]any)
			slots, _ := input["slots"].([]any)
			byID := map[string]map[string]any{}
			for _, raw := range slots {
				if slot, ok := raw.(map[string]any); ok {
					id := characterText(slot["slot_id"])
					if id == "" {
						id = characterText(slot["slot_token"])
					}
					byID[id] = slot
				}
			}
			characters := []any{}
			for _, raw := range selected {
				item, ok := raw.(map[string]any)
				if !ok {
					continue
				}
				id := characterText(item["slot_id"])
				slot := byID[id]
				name := characterText(item["display_name"])
				if name == "" {
					name = characterText(slot["display_name"])
					if name == "" {
						name = characterText(slot["base_name"])
					}
				}
				if name != "" {
					characters = append(characters, map[string]any{"slot_id": id, "display_name": name, "gender": characterText(slot["gender"]), "age_stage": characterText(slot["age_stage"]), "appearance": characterText(slot["appearance"])})
				}
			}
			relationships, _ := input["relationships"].([]any)
			writeJSON(w, http.StatusOK, map[string]any{"scene_id": characterText(cast["scene_id"]), "characters": characters, "relationships": relationships})
			return
		default:
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "人物核心接口不存在"})
		}
	}))
	for _, path := range []string{"/api/novel-panel/character-core/parse-slots", "/api/novel-panel/character-core/resolve-scene-cast", "/api/novel-panel/character-core/build-scene-context"} {
		root.Handle("POST "+path, handler)
	}
}

func RegisterNovelPanelCharacterTraceCompat(root *http.ServeMux, middleware BridgeAuth, usersDir string) {
	if root == nil || middleware.AuthStore == nil || strings.TrimSpace(usersDir) == "" {
		return
	}
	root.Handle("POST /api/novel-panel/character-core/trace", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var body map[string]any
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 2<<20)).Decode(&body) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		path := filepath.Join(usersDir, identity.Username, "novel-panel", "character-core-traces.json")
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "追踪记录保存失败"})
			return
		}
		var records []map[string]any
		if raw, err := os.ReadFile(path); err == nil {
			_ = json.Unmarshal(raw, &records)
		}
		if records == nil {
			records = []map[string]any{}
		}
		entry := map[string]any{"id": characterText(body["request_id"]), "at": time.Now().UTC().Format(time.RFC3339Nano), "trace_stage": characterText(body["trace_stage"])}
		if entry["id"] == "" {
			entry["id"] = "trace-" + toString(time.Now().UnixNano())
		}
		for key, value := range body {
			entry[key] = value
		}
		records = append([]map[string]any{entry}, records...)
		if len(records) > 50 {
			records = records[:50]
		}
		raw, _ := json.MarshalIndent(records, "", "  ")
		tmp := path + ".tmp"
		if err := os.WriteFile(tmp, raw, 0o600); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "追踪记录保存失败"})
			return
		}
		if err := os.Rename(tmp, path); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "追踪记录保存失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "logged": true})
	})))
}
