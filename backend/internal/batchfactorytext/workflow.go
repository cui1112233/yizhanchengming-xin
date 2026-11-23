package batchfactorytext

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

type EnsureRequest struct {
	Scope
	FrozenSettings                  map[string]json.RawMessage
	Target                          string
	SourceKey                       string
	CollectionID                    string // A pending upload resumes with the same immutable collection.
	PublishRunID, PublishRequestKey string
	PublishAttempt                  int
	CollectionReady                 func(Collection)
	ForceSensitive, ForcePublish    bool
	Progress                        func(stage, message string)
}

type EnsureResult struct {
	Stage      string
	Collection Collection
	Published  bool
}

var ErrPublishPending = errors.New("文本合集等待上传回执")

type WorkflowProduction struct {
	VersionInput
	Snapshot json.RawMessage
}

// Track persists each attempt and skips a matching successful attempt only when
// the caller has also checked that its output still exists.
type StageCheckpoint struct {
	Stage        string
	Fingerprint  string
	CollectionID string
}

type Workflow struct {
	Repo     Repository
	Track    func(context.Context, Scope, StageCheckpoint, bool, func() error) error
	Process  func(context.Context, EnsureRequest, bool) (WorkflowProduction, error)
	Prepare  func(context.Context, EnsureRequest, WorkflowProduction, Document, bool) error
	Render   func(context.Context, EnsureRequest, Document, Page) error // An empty page identifies the cover.
	Validate func(context.Context, EnsureRequest, Document) error
	Publish  func(context.Context, EnsureRequest, Collection) (bool, error)
}

func fingerprint(parts ...any) string {
	data, err := json.Marshal(parts)
	if err != nil {
		panic("invalid internal text workflow fingerprint: " + err.Error())
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

type workflowSnapshot struct {
	PageStyle
	SourceVersion string
}

// ValidateWorkflowStyle checks the same geometric and color constraints as the renderer
// without allocating a full image during batch preflight.
func ValidateWorkflowStyle(s PageStyle) error {
	layout, _, err := resolveLayout(s.Layout)
	if err != nil {
		return err
	}
	if s.Layout.TargetCharacters < 0 {
		return ErrInvalid
	}
	if _, err = rgb(layout.TextColor); err != nil {
		return err
	}
	if _, err = rgb(layout.HighlightColor); err != nil {
		return err
	}
	if s.Background.Color != "" {
		if _, err = rgb(s.Background.Color); err != nil {
			return err
		}
	}
	switch s.Background.Kind {
	case "", "solid", "color":
		if s.Background.ID != "" {
			return ErrInvalid
		}
	case "library", "image", "upload":
		if s.Background.ID == "" {
			return ErrInvalid
		}
	default:
		return ErrInvalid
	}
	_, err = (Renderer{}).Measure("字", layout)
	return err
}

func geometry(layout LayoutConfig) LayoutConfig {
	layout.TextColor = ""
	layout.HighlightColor = ""
	return layout
}

func (w Workflow) Ensure(ctx context.Context, req EnsureRequest) (res EnsureResult, err error) {
	if req.Owner == "" || req.BatchID == "" || req.BookID == "" || (req.Target != "collection" && req.Target != "published") || w.Repo == nil || w.Track == nil || w.Process == nil || w.Prepare == nil || w.Render == nil || w.Validate == nil {
		return res, ErrInvalid
	}
	for _, raw := range req.FrozenSettings {
		if !json.Valid(raw) {
			return res, ErrInvalid
		}
	}
	stage := func(id, label, key string, reuse bool, run func() error) error {
		res.Stage = id
		if req.Progress != nil {
			req.Progress(id, label)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		checkpoint := StageCheckpoint{Stage: id, Fingerprint: key}
		if id == "text_publish" {
			checkpoint.CollectionID = res.Collection.ID
		}
		return w.Track(ctx, req.Scope, checkpoint, reuse, run)
	}
	publish := func() error {
		key := fingerprint(res.Collection.ID, req.FrozenSettings["textPublishSettings"], req.FrozenSettings["publishSettings"], req.FrozenSettings["comicPublishSettings"])
		return stage("text_publish", "正在上传", key, !req.ForcePublish, func() error {
			if w.Publish == nil {
				return errors.New("文本上传服务未配置")
			}
			confirmed, err := w.Publish(ctx, req, res.Collection)
			if err == nil && !confirmed {
				return ErrPublishPending
			}
			return err
		})
	}
	if req.CollectionID != "" && req.Target == "published" {
		res.Stage = "text_publish"
		res.Collection, err = w.Repo.GetCollection(ctx, req.Scope, req.CollectionID)
		if err != nil {
			return res, err
		}
		if res.Collection.ID != req.CollectionID || res.Collection.Status != "ready" {
			return res, ErrNotReady
		}
		err = publish()
		res.Published = err == nil
		return res, err
	}
	production, verifyErr := w.Process(ctx, req, true)
	err = stage("text_process", "文本处理", req.SourceKey, verifyErr == nil, func() error {
		if verifyErr == nil {
			return nil
		}
		var err error
		production, err = w.Process(ctx, req, false)
		return err
	})
	if err != nil {
		return res, err
	}
	var next workflowSnapshot
	if err = json.Unmarshal(production.Snapshot, &next); err != nil {
		return res, err
	}
	doc, err := w.Repo.GetDocument(ctx, req.Scope)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return res, err
	}
	var old workflowSnapshot
	if len(doc.ConfigSnapshot) > 0 {
		if err = json.Unmarshal(doc.ConfigSnapshot, &old); err != nil {
			return res, err
		}
	}
	pageKey := fingerprint(production.ProductionText, geometry(next.Layout))
	reflow := geometry(next.Layout) != geometry(old.Layout)
	needsPages := doc.ActiveVersionID == "" || doc.PaginationVersion == 0 || next.SourceVersion != old.SourceVersion || reflow
	err = stage("text_paginate", "正在分页", pageKey, !needsPages, func() error {
		if !needsPages {
			return nil
		}
		if err := w.Prepare(ctx, req, production, doc, reflow); err != nil {
			return err
		}
		doc, err = w.Repo.GetDocument(ctx, req.Scope)
		return err
	})
	if err != nil {
		return res, err
	}
	var current, desired map[string]any
	if err = json.Unmarshal(doc.ConfigSnapshot, &current); err != nil {
		return res, err
	}
	if err = json.Unmarshal(production.Snapshot, &desired); err != nil {
		return res, err
	}
	// An author-draft marker belongs to the editor. Its consumption is guarded
	// by the authoritative sensitive-stage input, not by the render settings.
	delete(current, "authorDraftKey")
	delete(desired, "authorDraftKey")
	changeStyle := fingerprint(current) != fingerprint(desired)
	pages, err := w.Repo.ListPages(ctx, req.Scope)
	if err != nil {
		return res, err
	}
	renderKey := fingerprint(pageKey, desired)
	rendered := !changeStyle && doc.CoverArtifactID != "" && len(pages) > 0
	for _, p := range pages {
		rendered = rendered && p.ActiveArtifactID != ""
	}
	err = stage("text_render", "正在渲染", renderKey, rendered, func() error {
		if changeStyle {
			if err := w.Prepare(ctx, req, production, doc, false); err != nil {
				return err
			}
			doc, err = w.Repo.GetDocument(ctx, req.Scope)
			if err != nil {
				return err
			}
			pages, err = w.Repo.ListPages(ctx, req.Scope)
			if err != nil {
				return err
			}
		}
		if len(pages) == 0 {
			return ErrNotReady
		}
		var failures []error
		for _, p := range pages {
			if err := ctx.Err(); err != nil {
				return err
			}
			if p.VersionID != doc.ActiveVersionID || p.PaginationVersion != doc.PaginationVersion {
				return ErrConflict
			}
			if p.Locked || p.ActiveArtifactID != "" {
				continue
			}
			if req.Progress != nil {
				req.Progress("text_render", fmt.Sprintf("正在渲染第 %d 页", p.Ordinal))
			}
			if err := w.Render(ctx, req, doc, p); err != nil {
				failures = append(failures, fmt.Errorf("第 %d 页：%w", p.Ordinal, err))
			}
		}
		head, readErr := w.Repo.GetDocument(ctx, req.Scope)
		if readErr != nil {
			return readErr
		}
		if head.ActiveVersionID != doc.ActiveVersionID || head.PaginationVersion != doc.PaginationVersion {
			return ErrConflict
		}
		if !head.CoverLocked && head.CoverArtifactID == "" {
			if err := w.Render(ctx, req, head, Page{}); err != nil {
				failures = append(failures, fmt.Errorf("封面：%w", err))
			}
		}
		return errors.Join(failures...)
	})
	if err != nil {
		return res, err
	}
	head, err := w.Repo.GetDocument(ctx, req.Scope)
	if err != nil {
		return res, err
	}
	if head.ActiveVersionID != doc.ActiveVersionID || head.PaginationVersion != doc.PaginationVersion {
		return res, ErrConflict
	}
	pages, err = w.Repo.ListPages(ctx, req.Scope)
	if err != nil {
		return res, err
	}
	var manifest [][2]string
	for _, p := range pages {
		if p.VersionID != head.ActiveVersionID || p.PaginationVersion != head.PaginationVersion {
			return res, ErrConflict
		}
		manifest = append(manifest, [2]string{p.ID, p.ActiveArtifactID})
	}
	fileKey := fingerprint(renderKey, head.ActiveVersionID, head.PaginationVersion, head.CoverArtifactID, manifest)
	err = stage("text_validate", "正在校验", fileKey, true, func() error { return w.Validate(ctx, req, head) })
	if err != nil {
		return res, err
	}
	res.Collection, err = w.matchingCollection(ctx, req.Scope, head, manifest)
	if err != nil {
		return res, err
	}
	err = stage("text_collection", "正在整理合集", fileKey, res.Collection.ID != "", func() error {
		if res.Collection.ID != "" {
			return nil
		}
		res.Collection, err = w.Repo.CreateCollection(ctx, req.Scope, CollectionInput{Reuse: true, ExpectedRevision: head.Revision, VersionID: head.ActiveVersionID, PaginationVersion: head.PaginationVersion, ConfigSnapshot: head.ConfigSnapshot})
		return err
	})
	if err != nil {
		return res, err
	}
	latest, err := w.Repo.GetDocument(ctx, req.Scope)
	if err != nil {
		return res, err
	}
	if latest.Revision != head.Revision {
		return res, ErrConflict
	}
	if req.CollectionReady != nil {
		req.CollectionReady(res.Collection)
	}
	if req.Target == "collection" {
		return res, nil
	}
	err = publish()
	res.Published = err == nil
	return res, err
}

func (w Workflow) matchingCollection(ctx context.Context, s Scope, d Document, manifest [][2]string) (Collection, error) {
	records, err := w.Repo.ListCollections(ctx, s)
	if err != nil {
		return Collection{}, err
	}
	for _, record := range records {
		if record.VersionID != d.ActiveVersionID || record.PaginationVersion != d.PaginationVersion || record.CoverArtifactID != d.CoverArtifactID || record.Status != "ready" {
			continue
		}
		full, err := w.Repo.GetCollection(ctx, s, record.ID)
		if err != nil {
			return Collection{}, err
		}
		if len(full.Items) != len(manifest) {
			continue
		}
		matched := true
		for i, item := range full.Items {
			if item.PageID != manifest[i][0] || item.ArtifactID != manifest[i][1] {
				matched = false
				break
			}
		}
		if matched {
			return full, nil
		}
	}
	return Collection{}, nil
}
