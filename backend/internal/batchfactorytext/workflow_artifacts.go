package batchfactorytext

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

func (w Workflow) artifactsByID(ctx context.Context, scope Scope) (map[string]PageArtifact, error) {
	artifacts, err := w.Repo.ListArtifacts(ctx, scope)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]PageArtifact, len(artifacts))
	for _, art := range artifacts {
		byID[art.ID] = art
	}
	return byID, nil
}

// A render checkpoint describes the settings. The active file can disappear or
// its contents can change independently of the document and its checkpoint.
func (w Workflow) inspectFiles(ctx context.Context, scope Scope, doc Document, pages []Page) (map[string]error, error) {
	byID, err := w.artifactsByID(ctx, scope)
	if err != nil {
		return nil, err
	}
	repairs := map[string]error{}
	check := func(pageID, artifactID, label string, valid bool) error {
		err := ErrNotReady
		if artifactID != "" && valid {
			err = w.VerifyFile(ctx, byID[artifactID], label)
		}
		if errors.Is(err, ErrNotReady) {
			repairs[pageID] = err
			return nil
		}
		return err
	}
	for _, p := range pages {
		if p.VersionID != doc.ActiveVersionID || p.PaginationVersion != doc.PaginationVersion {
			return nil, ErrConflict
		}
		art := byID[p.ActiveArtifactID]
		valid := art.Kind == "page" && art.PageID == p.ID && art.RenderVersion == p.RenderVersion && p.Status == "ready" && art.SourceKind == sourceKind(p.Locked)
		if err := check(p.ID, p.ActiveArtifactID, fmt.Sprintf("第 %d 页", p.Ordinal), valid); err != nil {
			return nil, err
		}
	}
	cover := byID[doc.CoverArtifactID]
	err = check("", doc.CoverArtifactID, "封面", cover.Kind == "cover" && cover.PageID == "" && cover.SourceKind == sourceKind(doc.CoverLocked))
	return repairs, err
}

func matchesPages(collection Collection, doc Document, manifest [][2]string) bool {
	if collection.Status != "ready" || collection.VersionID != doc.ActiveVersionID || collection.PaginationVersion != doc.PaginationVersion || collection.CoverArtifactID != doc.CoverArtifactID || len(collection.Items) != len(manifest) {
		return false
	}
	for i, item := range collection.Items {
		if item.Ordinal != i+1 || item.PageID != manifest[i][0] || item.ArtifactID != manifest[i][1] {
			return false
		}
	}
	return true
}

func workflowCollectionID(scope Scope, collection Collection) string {
	return fingerprint(scope, collection.VersionID, collection.PaginationVersion, collection.CoverArtifactID, collection.Items)
}

// Pinned uploads can refer to an earlier document version. Check the saved
// manifest against its own immutable artifacts, even after the editor has moved on.
func (w Workflow) verifyManifest(ctx context.Context, scope Scope, c Collection, pinned bool) error {
	if c.Status != "ready" || c.VersionID == "" || c.PaginationVersion < 1 || len(c.Items) == 0 {
		return ErrNotReady
	}
	// Workflow-created collections use their SHA-256 manifest ID. Manual
	// collections have opaque IDs and still require the reference checks below.
	if len(c.ID) == hex.EncodedLen(sha256.Size) && c.ID != workflowCollectionID(scope, c) {
		return fmt.Errorf("%w（合集版本或图片清单已变化）", ErrNotReady)
	}
	if pinned {
		if _, err := w.Repo.GetVersion(ctx, scope, c.VersionID); err != nil {
			return err
		}
	}
	byID, err := w.artifactsByID(ctx, scope)
	if err != nil {
		return err
	}
	cover := byID[c.CoverArtifactID]
	if cover.Kind != "cover" || cover.PageID != "" {
		return fmt.Errorf("%w（合集封面记录无效）", ErrNotReady)
	}
	if pinned {
		if err := w.VerifyFile(ctx, cover, "封面"); err != nil {
			return err
		}
	}
	prev := 0
	for _, item := range c.Items {
		art := byID[item.ArtifactID]
		if art.Kind != "page" || item.PageID == "" || art.PageID != item.PageID || art.StorageRef != item.StorageRef || art.SHA256 != item.SHA256 || art.MediaType != item.MediaType || item.Ordinal <= prev {
			return fmt.Errorf("%w（合集第 %d 页记录无效）", ErrNotReady, item.Ordinal)
		}
		if pinned {
			if err := w.VerifyFile(ctx, art, fmt.Sprintf("第 %d 页", item.Ordinal)); err != nil {
				return err
			}
		}
		prev = item.Ordinal
	}
	return nil
}
