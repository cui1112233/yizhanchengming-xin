package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"qiantie/backend/internal/novelfetchworkshop"
)

// Bind a manual confirmation to the saved configuration and source versions
// that were previewed. Never store or expose the raw configuration in the token.
func webSubmitPreviewRevision(ctx context.Context, store novelfetchworkshop.Store, owner string, ids, versions []string) (string, error) {
	settings, err := store.GetConfig(ctx, owner)
	if err != nil {
		return "", err
	}
	web := mapOrEmpty(mapOrEmpty(settings["appConfig"])["web_submit"])
	if len(versions) == 0 {
		versions = cleanStrings(anyStrings(web["submit_versions"]))
	}
	if len(versions) == 0 {
		versions = []string{"ai1"}
	}
	rows := []any{}
	for _, id := range cleanStrings(ids) {
		doc, err := store.GetDocument(ctx, owner, id)
		if err != nil {
			rows = append(rows, map[string]any{"id": id, "missing": true})
			continue
		}
		bodies := map[string]any{}
		for _, version := range versions {
			body, bodyErr := store.GetBody(ctx, owner, id, version)
			if bodyErr != nil {
				bodies[version] = nil
			} else {
				bodies[version] = body.Content
			}
		}
		rows = append(rows, map[string]any{"id": id, "meta": doc.Meta, "gender": doc.Gender, "info": doc.BookInfo, "bodies": bodies})
	}
	encoded, err := json.Marshal(map[string]any{"owner": owner, "config": web, "versions": versions, "tasks": rows})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}
