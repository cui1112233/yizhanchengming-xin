package httpapi

import (
	"context"
	"errors"
	"sort"
	"strings"

	"qiantie/backend/internal/novelfetchworkshop"
)

// hydrateNovelFetchDocument loads正文 that was moved out of document_json into
// the body store. The history record intentionally stays lightweight, but all
// handlers that need the actual text must see the same complete document.
func hydrateNovelFetchDocument(ctx context.Context, store novelfetchworkshop.Store, owner string, document novelfetchworkshop.Document) (novelfetchworkshop.Document, error) {
	if store == nil {
		return document, nil
	}
	if document.BodyRefs == nil {
		document.BodyRefs = map[string]novelfetchworkshop.BodyRef{}
	}
	refs, err := store.ListBodyRefs(ctx, owner, document.BookID)
	if err != nil {
		return document, err
	}
	for _, ref := range refs {
		document.BodyRefs[ref.VersionID] = ref
	}
	versionIDs := make([]string, 0, len(document.BodyRefs))
	for versionID := range document.BodyRefs {
		versionIDs = append(versionIDs, versionID)
	}
	sort.Strings(versionIDs)
	for _, versionID := range versionIDs {
		body, bodyErr := store.GetBody(ctx, owner, document.BookID, versionID)
		if errors.Is(bodyErr, novelfetchworkshop.ErrNotFound) {
			continue
		}
		if bodyErr != nil {
			return document, bodyErr
		}
		document.BodyRefs[versionID] = body.BodyRef
		switch strings.ToLower(strings.TrimSpace(versionID)) {
		case "original_raw", "originalraw":
			document.OriginalRaw = body.Content
			document.OriginalRawChars = len([]rune(body.Content))
		case "original":
			document.Original = body.Content
			document.ProcessedText = body.Content
			document.ProcessedChars = len([]rune(body.Content))
		default:
			if document.Versions == nil {
				document.Versions = map[string]any{}
			}
			if metadata, ok := document.Versions[versionID].(map[string]any); ok {
				copy := map[string]any{}
				for key, value := range metadata {
					copy[key] = value
				}
				copy["text"] = body.Content
				document.Versions[versionID] = copy
			} else {
				document.Versions[versionID] = body.Content
			}
		}
	}
	return document, nil
}

func persistNovelFetchBodies(ctx context.Context, store novelfetchworkshop.Store, owner string, document *novelfetchworkshop.Document, originalRaw, processed string) error {
	if document == nil || store == nil {
		return nil
	}
	if document.BodyRefs == nil {
		document.BodyRefs = map[string]novelfetchworkshop.BodyRef{}
	}
	if strings.TrimSpace(originalRaw) != "" {
		ref, err := store.PutBody(ctx, owner, novelfetchworkshop.BodyRecord{
			BookID:  document.BookID,
			BodyRef: novelfetchworkshop.BodyRef{VersionID: "original_raw", State: "ready"},
			Content: originalRaw,
		})
		if err != nil {
			return err
		}
		document.BodyRefs[ref.VersionID] = ref
	}
	if strings.TrimSpace(processed) != "" {
		ref, err := store.PutBody(ctx, owner, novelfetchworkshop.BodyRecord{
			BookID:  document.BookID,
			BodyRef: novelfetchworkshop.BodyRef{VersionID: "original", State: "ready"},
			Content: processed,
		})
		if err != nil {
			return err
		}
		document.BodyRefs[ref.VersionID] = ref
	}
	return nil
}
