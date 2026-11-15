package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"qiantie/backend/internal/screenplay"
	"unicode/utf8"
)

const screenplayLibrarySubjectKey = "_cm_library_sora2_info"
const screenplayLibrarySubjectLimit = 16 << 10

func validateLibrarySubject(kind screenplay.AssetType, value *string) error {
	if value == nil {
		return nil
	}
	if kind != screenplay.AssetCharacter {
		return fmt.Errorf("%w: 仅角色支持主体存档", screenplay.ErrInvalidArgument)
	}
	if len(*value) > screenplayLibrarySubjectLimit || !utf8.ValidString(*value) {
		return fmt.Errorf("%w: 主体存档须为有效文字，最多16KiB UTF-8", screenplay.ErrInvalidArgument)
	}
	return nil
}

func libraryArchiveSubject(record map[string]json.RawMessage) (*string, error) {
	var selected *string
	for _, key := range []string{"sora2Info", "sora2_info"} {
		raw, ok := record[key]
		if !ok {
			continue
		}
		var value string
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &value) != nil {
			return nil, fmt.Errorf("%w: 主体存档必须为文字", screenplay.ErrInvalidArgument)
		}
		if err := validateLibrarySubject(screenplay.AssetCharacter, &value); err != nil {
			return nil, err
		}
		if selected == nil {
			selected = &value
		}
	}
	return selected, nil
}

func librarySubject(entry screenplay.AssetLibraryEntry) string {
	if entry.Type != screenplay.AssetCharacter {
		return ""
	}
	if value, ok := entry.Snapshot.Extra[screenplayLibrarySubjectKey]; ok {
		return value
	}
	var record map[string]json.RawMessage
	if json.Unmarshal([]byte(entry.Snapshot.Extra[screenplay.LibraryArchiveJSON]), &record) == nil {
		// Historical camel values, including empty strings, take precedence.
		for _, key := range []string{"sora2Info", "sora2_info"} {
			if raw, ok := record[key]; ok {
				var value string
				if json.Unmarshal(raw, &value) == nil {
					return value
				}
			}
		}
	}
	return entry.Snapshot.Extra["sora2_info"]
}

func setLibrarySubject(entry *screenplay.AssetLibraryEntry, value *string) {
	if value == nil {
		return
	}
	// Source snapshots can share their Extra map with a project asset.
	// Copy before writing so saving to the personal library never edits a project.
	extra := make(map[string]string, len(entry.Snapshot.Extra)+1)
	for key, current := range entry.Snapshot.Extra {
		extra[key] = current
	}
	extra[screenplayLibrarySubjectKey] = *value
	entry.Snapshot.Extra = extra
}

func sourceSubjectPointer(source screenplay.Asset) *string {
	if source.Type != screenplay.AssetCharacter {
		return nil
	}
	if value, ok := source.Extra["sora2_info"]; ok {
		return &value
	}
	return nil
}
