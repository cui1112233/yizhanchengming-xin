package httpapi

import (
	"encoding/json"

	"qiantie/backend/internal/batchfactoryv11"
)

// Cover background is one setting, while each title layout field can inherit
// independently. A first-time book can override just its title alignment.
func effectiveTextSettings(batch, book batchfactoryv11.SettingsPatch) batchfactoryv11.SettingsPatch {
	effective := batchfactoryv11.ResolveSettings(batch, book)
	b, c := batch["textCover"], book["textCover"]
	if len(b) == 0 || len(c) == 0 {
		return effective
	}
	var base, override map[string]json.RawMessage
	if json.Unmarshal(b, &base) != nil || base == nil || json.Unmarshal(c, &override) != nil {
		return effective
	}
	for key, val := range override {
		if key == "layout" {
			var inherited, explicit map[string]json.RawMessage
			if json.Unmarshal(base[key], &inherited) == nil && inherited != nil && json.Unmarshal(val, &explicit) == nil && explicit != nil {
				for k, v := range explicit {
					inherited[k] = v
				}
				val, _ = json.Marshal(inherited)
			}
		}
		base[key] = val
	}
	effective["textCover"], _ = json.Marshal(base)
	return effective
}
