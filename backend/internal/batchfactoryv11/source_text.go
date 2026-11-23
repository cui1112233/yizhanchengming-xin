package batchfactoryv11

import "strings"

func normalizeCapturedSourceText(value string) (string, bool) {
	text := strings.TrimSpace(value)
	if text == "" || text == "[object Object]" {
		return "", false
	}
	return text, true
}

// Only a caller that actually received a full original can establish its origin.
// Existing rows with no such evidence keep their text and remain unverified.
func sourceWriteMetadata(current, patch map[string]any, original, origin string) map[string]any {
	metadata := make(map[string]any, len(current)+len(patch)+2)
	for key, value := range current {
		metadata[key] = value
	}
	for key, value := range patch {
		if key != "rawSourceStatus" && key != "rawSourceOrigin" {
			metadata[key] = value
		}
	}
	if original != "" {
		if origin == "" {
			origin = "captured_raw"
		}
		metadata["rawSourceStatus"] = "verified"
		metadata["rawSourceOrigin"] = origin
	} else if metadata["rawSourceStatus"] == nil {
		metadata["rawSourceStatus"] = "unverified"
	}
	return metadata
}

func sourceReadMetadata(metadata map[string]any) map[string]any {
	if metadata["rawSourceStatus"] != nil {
		return metadata
	}
	return sourceWriteMetadata(metadata, nil, "", "")
}

// A normal metadata edit replaces the editable fields while the original's
// provenance stays attached to its source book ID.
func replaceBookMetadata(current, replacement map[string]any, currentID, newID string) map[string]any {
	facts := map[string]any{}
	if newID == "" || newID == currentID {
		for _, key := range []string{"rawSourceStatus", "rawSourceOrigin"} {
			if value, ok := current[key]; ok {
				facts[key] = value
			}
		}
	}
	return sourceWriteMetadata(facts, replacement, "", "")
}

func (input SaveBookProductionTextInput) captureInput() CaptureBookSourceInput {
	return CaptureBookSourceInput{
		SourceText: input.SourceText, ReplaceSource: true, SourceMetadata: input.SourceMetadata,
		ExpectedRevision: input.ExpectedRevision,
	}
}

func (input CaptureBookSourceInput) updatedMetadata(current map[string]any, currentID, raw string) map[string]any {
	metadata := sourceWriteMetadata(current, input.SourceMetadata, raw, input.RawSourceOrigin)
	if id := strings.TrimSpace(input.SourceBookID); raw == "" && id != "" && id != currentID {
		delete(metadata, "rawSourceOrigin")
		metadata["rawSourceStatus"] = "unverified"
	}
	return metadata
}
