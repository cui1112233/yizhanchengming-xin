package httpapi

import (
	"context"
	"strings"

	"qiantie/backend/internal/novelfetchworkshop"
)

func novelFetchStopped(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "stopped", "cancelled", "canceled":
		return true
	}
	return false
}

func novelFetchRequiredVersions(versions []string) []string {
	if len(versions) == 0 {
		return []string{"original"}
	}
	return versions
}

// A task-level submitted label cannot prove that every requested version was
// accepted. Unknown outcomes retain their existing verification path.
func novelFetchVersionsSubmitted(doc novelfetchworkshop.Document, versions []string) bool {
	done := scheduleStringSlice(firstConfiguredValue(doc.Meta["siteSubmitDoneVersions"], doc.Meta["site_submit_done_versions"]))
	for _, version := range novelFetchRequiredVersions(versions) {
		if !seenString(done, version) {
			return false
		}
	}
	return true
}

func novelFetchVersionsGenerated(ctx context.Context, store novelfetchworkshop.Store, owner string, doc novelfetchworkshop.Document, versions []string) bool {
	for _, version := range novelFetchRequiredVersions(versions) {
		if version == "original" && strings.TrimSpace(firstText(doc.ProcessedText, doc.Original)) != "" {
			continue
		}
		if version != "original" && strings.TrimSpace(toString(doc.Versions[version])) != "" {
			continue
		}
		if ref, ok := doc.BodyRefs[version]; ok && ref.State == "ready" {
			continue
		}
		body, err := store.GetBody(ctx, owner, doc.BookID, version)
		if err == nil && strings.TrimSpace(body.Content) != "" {
			continue
		}
		if version == "original" {
			body, err = store.GetBody(ctx, owner, doc.BookID, "processed_text")
			if err == nil && strings.TrimSpace(body.Content) != "" {
				continue
			}
		}
		return false
	}
	return true
}
