package batchfactoryv11

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

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
		if key != "rawSourceStatus" && key != "rawSourceOrigin" && key != "batchSensitiveStage" {
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
		for _, key := range []string{"rawSourceStatus", "rawSourceOrigin", "batchSensitiveStage"} {
			if value, ok := current[key]; ok {
				facts[key] = value
			}
		}
	}
	return sourceWriteMetadata(facts, replacement, "", "")
}

func (input SaveBookProductionTextInput) captureInput() CaptureBookSourceInput {
	return CaptureBookSourceInput{
		ProductionStage: input.Stage, KeepProductionText: input.KeepProductionText,
		SourceText: input.SourceText, ReplaceSource: true, SourceMetadata: input.SourceMetadata,
		ExpectedRevision: input.ExpectedRevision,
	}
}

func (input CaptureBookSourceInput) updatedMetadata(current map[string]any, currentID, raw string) map[string]any {
	metadata := sourceWriteMetadata(current, input.SourceMetadata, raw, input.RawSourceOrigin)
	if raw != "" {
		delete(metadata, "productionTextEdited")
	}
	if input.ProductionStage != nil {
		metadata["batchSensitiveStage"] = input.ProductionStage
	}
	if id := strings.TrimSpace(input.SourceBookID); raw == "" && id != "" && id != currentID {
		delete(metadata, "rawSourceOrigin")
		metadata["rawSourceStatus"] = "unverified"
	}
	return metadata
}

type preparedProductionContextKey struct{}
type preparedProductionVersion struct{ BookID, Key string }

func ProductionVersion(book Book) string {
	state, _ := book.SourceMetadata["batchSensitiveStage"].(map[string]any)
	if state["status"] != "done" {
		return ""
	}
	key, _ := state["key"].(string)
	return key
}

// WithPreparedProduction binds the version approved for this call. It must be
// carried to the service that actually contacts the model.
func WithPreparedProduction(ctx context.Context, book Book) context.Context {
	if _, pinned := ctx.Value(preparedProductionContextKey{}).(preparedProductionVersion); pinned {
		return ctx
	}
	return context.WithValue(ctx, preparedProductionContextKey{}, preparedProductionVersion{book.ID, ProductionVersion(book)})
}

func checkedProductionSource(ctx context.Context, store Store, owner, batchID string, book Book) (string, bool, error) {
	draft, err := store.GetDraft(ctx, owner, "working-front:"+book.ID, "working-front-content", batchID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return "", false, err
	}
	hasDraft := strings.TrimSpace(draft.Content) != ""
	version := ProductionVersion(book)
	expected, pinned := ctx.Value(preparedProductionContextKey{}).(preparedProductionVersion)
	state, _ := book.SourceMetadata["batchSensitiveStage"].(map[string]any)
	stale := func() (string, bool, error) {
		return "", false, fmt.Errorf("%w: 生产正文已更新，请重新执行", ErrConflict)
	}
	if pinned && (expected.BookID != book.ID || version == "" || expected.Key != version) {
		return stale()
	}
	if len(state) > 0 {
		base, _ := state["authoredBase"].(string)
		if version == "" || state["raw"] != sourceDigest(book.TxtText) || state["output"] != sourceDigest(book.SourceText) || hasDraft && base != WorkingDraftVersion(draft.Content) || !hasDraft && strings.HasPrefix(base, "working:") {
			return stale()
		}
		return book.SourceText, hasDraft, nil
	}
	if hasDraft {
		return draft.Content, true, nil
	}
	return book.SourceText, false, nil
}
