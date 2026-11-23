package batchfactoryv11

import (
	"encoding/json"
	"sort"
	"strings"
)

func clonePatch(input SettingsPatch) SettingsPatch {
	out := SettingsPatch{}
	for k, v := range input {
		out[k] = append([]byte(nil), v...)
	}
	return out
}

// normalizeLegacyNestedPatch reads the accidental V12 UI wrapper written by
// earlier unified-setting saves. The persisted document was {"patch": {...}}
// instead of the settings object itself. The wrapper is not a supported
// settings key, so it can be removed while retaining later sparse outer keys.
func normalizeLegacyNestedPatch(input SettingsPatch) SettingsPatch {
	out := clonePatch(input)
	// expectedRevision belongs to the HTTP request envelope. Older clients
	// accidentally stored it alongside the real sparse settings.
	delete(out, "expectedRevision")
	rawPatch, ok := out["patch"]
	if !ok {
		return out
	}
	nested := SettingsPatch{}
	if err := json.Unmarshal(rawPatch, &nested); err != nil || len(nested) == 0 {
		return out
	}
	delete(out, "patch")
	// Only the canonical top-level marker was written by the project creator.
	for key := range nested {
		if IsProjectOwnershipKey(key) {
			delete(nested, key)
		}
	}
	// The nested patch is the older full configuration; outer keys are later
	// sparse updates and therefore win when both provide the same setting.
	for key, value := range out {
		nested[key] = append([]byte(nil), value...)
	}
	return normalizeLegacyNestedPatch(nested)
}

const shuihuoProjectKey = "shuihuoProject"

// Project ownership is set once by the creating module and is never a user setting.
func IsProjectOwnershipKey(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "shuihuoproject", "projectkind", "project-kind", "project_kind":
		return true
	}
	return false
}

// Strip ownership from a saved config or legacy patch envelope while
// preserving every unrelated value, including non-object configs.
func userConfigJSON(raw json.RawMessage) (json.RawMessage, bool) {
	var object SettingsPatch
	if json.Unmarshal(raw, &object) != nil || object == nil {
		return raw, false
	}
	changed := false
	for key := range object {
		if IsProjectOwnershipKey(key) {
			delete(object, key)
			changed = true
		}
	}
	if nested, ok := object["patch"]; ok {
		if clean, modified := userConfigJSON(nested); modified {
			object["patch"] = clean
			changed = true
		}
	}
	if !changed {
		return raw, false
	}
	cleaned, _ := json.Marshal(object)
	return cleaned, true
}

func (p SettingsPatch) IsShuihuoProject() bool {
	return strings.TrimSpace(string(p[shuihuoProjectKey])) == "true"
}

func (input CreateBatchInput) initialSettings() SettingsPatch {
	patch := ApplySparseUpdate(nil, SettingsUpdate{Patch: input.InitialBatchSettings})
	if input.ShuihuoProject {
		patch[shuihuoProjectKey] = json.RawMessage("true")
	}
	return patch
}

func ApplySparseUpdate(current SettingsPatch, update SettingsUpdate) SettingsPatch {
	out := clonePatch(current)
	for _, key := range update.RestoreKeys {
		if IsProjectOwnershipKey(key) {
			continue
		}
		delete(out, key)
	}
	for key, value := range update.Patch {
		if IsProjectOwnershipKey(key) {
			continue
		}
		if key == "patch" {
			value, _ = userConfigJSON(value)
		}
		out[key] = append([]byte(nil), value...)
	}
	return out
}

func mergeAIPromptConfig(base, next json.RawMessage) json.RawMessage {
	if len(base) == 0 {
		return append([]byte(nil), next...)
	}
	merged := map[string]json.RawMessage{}
	if err := json.Unmarshal(base, &merged); err != nil {
		return append([]byte(nil), next...)
	}
	patch := map[string]json.RawMessage{}
	if err := json.Unmarshal(next, &patch); err != nil {
		return append([]byte(nil), next...)
	}
	for key, value := range patch {
		// A book-level partial write (for example only the comic page preset)
		// must not erase the batch-level sibling selections under the same
		// module. Merge plain JSON objects one level deeper; arrays and scalars
		// keep the newer layer's value.
		merged[key] = mergeJSONObject(merged[key], value)
	}
	encoded, err := json.Marshal(merged)
	if err != nil {
		return append([]byte(nil), next...)
	}
	return encoded
}

func mergeJSONObject(base, next json.RawMessage) json.RawMessage {
	if len(base) == 0 {
		return append([]byte(nil), next...)
	}
	merged := map[string]json.RawMessage{}
	if err := json.Unmarshal(base, &merged); err != nil {
		return append([]byte(nil), next...)
	}
	patch := map[string]json.RawMessage{}
	if err := json.Unmarshal(next, &patch); err != nil {
		return append([]byte(nil), next...)
	}
	for key, value := range patch {
		merged[key] = append([]byte(nil), value...)
	}
	encoded, err := json.Marshal(merged)
	if err != nil {
		return append([]byte(nil), next...)
	}
	return encoded
}

func ResolveSettings(layers ...SettingsPatch) SettingsPatch {
	out := SettingsPatch{}
	for _, layer := range layers {
		for key, value := range layer {
			switch key {
			case "aiPromptConfig":
				out[key] = mergeAIPromptConfig(out[key], value)
				continue
			case "publishSettings", "comicPublishSettings":
				out[key] = mergeJSONObject(out[key], value)
				continue
			}
			out[key] = append([]byte(nil), value...)
		}
	}
	return out
}

// EffectiveProductionMode returns the batch-wide production switch. The batch
// layer is authoritative because the workbench renders one mode for the whole
// batch; a stale single-book productionMode (left behind by an older switch)
// must not drag a video batch into the comic stage sequence. Batches that only
// stored the mode on the book keep working when the batch layer is silent.
func EffectiveProductionMode(batchPatch, bookPatch SettingsPatch) string {
	if mode := rawString(batchPatch, "productionMode", ""); mode == "comic" || mode == "video" || mode == "text" {
		return mode
	}
	return normalizeProductionMode(ResolveSettings(batchPatch, bookPatch))
}

// AssetImagesEnabled reports whether automation should generate asset
// reference images. Defaults to false: batches and books that never wrote the
// key keep the assets stage off without blocking the storyboard, video or comic
// stages. Only an explicit true in the batch or book patch turns generation on;
// when enabled, downstream prompts use the reference images and must not repeat
// the stored asset descriptions as text.
func AssetImagesEnabled(patch SettingsPatch) bool {
	return rawBool(patch, "assetImagesEnabled", false)
}

// AssetTextFallbackEnabled reports whether downstream prompts should merge the
// stored asset descriptions because asset reference images are disabled. It is
// the inverse of AssetImagesEnabled: the two paths are mutually exclusive.
func AssetTextFallbackEnabled(patch SettingsPatch) bool {
	return !AssetImagesEnabled(patch)
}

func ChangedKeys(update SettingsUpdate) []string {
	set := map[string]struct{}{}
	for k := range update.Patch {
		set[k] = struct{}{}
	}
	for _, k := range update.RestoreKeys {
		set[k] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
