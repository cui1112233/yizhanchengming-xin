package httpapi

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"qiantie/backend/internal/batchfactorysensitive"
	txt "qiantie/backend/internal/batchfactorytext"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
)

// The publisher saves an owner-scoped receipt under the immutable collection ID.
// Read the current attempt and its last confirmed predecessor independently.
func addWorkflowPublishSummary(ctx context.Context, store batchfactoryv11.Store, owner, batchID, bookID, manualCollectionID string, runs []batchfactoryv11.BookStageRun, summary map[string]any) error {
	var uploads []batchfactoryv11.BookStageRun
	for _, r := range runs {
		if r.Stage == "text_publish" && r.CollectionID != "" {
			uploads = append(uploads, r)
		}
	}
	slices.SortFunc(uploads, func(a, b batchfactoryv11.BookStageRun) int { return cmp.Compare(b.Attempt, a.Attempt) })
	manual := len(uploads) == 0 && manualCollectionID != ""
	if manual {
		uploads = append(uploads, batchfactoryv11.BookStageRun{CollectionID: manualCollectionID})
	}
	seen := map[string]bool{}
	scoped := func(rec map[string]any, id string) bool {
		return rec["kind"] == "text" && rec["bookId"] == bookID && rec["collectionId"] == id
	}
	confirmed := func(rec map[string]any) bool {
		images, _ := rec["imageIds"].([]any)
		if rec["status"] != "published" || intValue(rec["carouselId"]) <= 0 || len(images) == 0 {
			return false
		}
		if n := intValue(rec["pageCount"]); n > 0 && n != len(images) {
			return false
		}
		for _, id := range images {
			if intValue(id) <= 0 {
				return false
			}
		}
		return true
	}
	for i, run := range uploads {
		if seen[run.CollectionID] {
			continue
		}
		seen[run.CollectionID] = true
		d, err := store.GetDraft(ctx, owner, run.CollectionID, "n8-text-publish", batchID)
		if errors.Is(err, batchfactoryv11.ErrNotFound) {
			continue
		}
		if err != nil {
			return err
		}
		var rec map[string]any
		if err = json.Unmarshal([]byte(d.Content), &rec); err != nil {
			return err
		}
		if !scoped(rec, run.CollectionID) {
			continue
		}
		if i == 0 {
			runID := stringValue(rec["runId"])
			attempt := intValue(rec["attempt"])
			key := stringValue(rec["requestKey"])
			matches := (runID != "" && runID == run.ID) || (runID == "" && attempt > 0 && attempt == run.Attempt) || (runID == "" && attempt == 0 && key != "" && key == run.RequestID)
			if manual || matches && (attempt == 0 || attempt == run.Attempt) {
				if rec["status"] == "pending" {
					rec["status"] = "upload_pending_confirm"
				}
				summary["textPublishRecord"] = rec
			}
		}
		history := rec
		if !confirmed(rec) {
			history, _ = rec["lastPublished"].(map[string]any)
		}
		if scoped(history, run.CollectionID) && confirmed(history) {
			summary["textLastPublishedRecord"] = history
			return nil
		}
	}
	return nil
}

func freezeAutomationTextSettings(entry map[string]any, batch batchfactoryv11.SettingsPatch, book batchfactoryv11.Book) error {
	if batchfactoryv11.EffectiveProductionMode(batch, book.SettingsState.Patch) != "text" {
		return nil
	}
	settings := effectiveTextSettings(batch, book.SettingsState.Patch)
	settings["productionMode"] = json.RawMessage(`"text"`)
	raw, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	entry["frozenTextSettings"] = string(raw)
	return nil
}

func automationTextSettings(entry map[string]any, batch batchfactoryv11.Batch, book batchfactoryv11.Book) (batchfactoryv11.SettingsPatch, bool, error) {
	if raw, exists := entry["frozenTextSettings"]; exists {
		var frozen batchfactoryv11.SettingsPatch
		data, ok := raw.(string)
		if !ok || json.Unmarshal([]byte(data), &frozen) != nil || frozen == nil {
			return nil, true, fmt.Errorf("任务保存的文本设置无效")
		}
		return frozen, true, nil
	}
	if batchfactoryv11.EffectiveProductionMode(batch.SettingsState.Patch, book.SettingsState.Patch) != "text" {
		return nil, false, nil
	}
	return effectiveTextSettings(batch.SettingsState.Patch, book.SettingsState.Patch), true, nil
}

func validateAutomationTextSettings(x textRequest) error {
	snap, err := bookTextSnapshot(x)
	if err != nil {
		return err
	}
	_, cover, err := txt.ResolvePageStyle(snap.PageStyle, txt.PageDraft{}, snap.Cover)
	if err != nil {
		return err
	}
	for _, style := range []txt.PageStyle{snap.PageStyle, cover} {
		if err := txt.ValidateWorkflowStyle(style); err != nil {
			return err
		}
	}
	return nil
}

// Both the manual text endpoints and the automation worker use the same prepared
// drafts, render inputs, artifact writer and cryptographic file validation.
func (c *BatchAutomationCoordinator) configureTextWorkflow(repo txt.Repository, files *localartifact.Store) {
	if repo == nil || files == nil || c.textEnsure != nil {
		return
	}
	a := textAPI{c.store, repo, files, c}
	c.textEnsure = func(ctx context.Context, req txt.EnsureRequest) (txt.EnsureResult, error) {
		unlock := c.lockBatchRun(req.Owner, "text-workflow/"+req.BatchID+"/"+req.BookID)
		defer unlock()
		batch, err := c.store.GetBatch(ctx, req.Owner, req.BatchID)
		if err != nil {
			return txt.EnsureResult{Stage: "text_process"}, err
		}
		b, ok := findBatchBookByID(batch, req.BookID)
		if !ok {
			return txt.EnsureResult{Stage: "text_process"}, batchfactoryv11.ErrNotFound
		}
		batch.SettingsState.Patch = req.FrozenSettings
		b.SettingsState.Patch = nil
		x := textRequest{ctx, req.Scope, batch, b, req.FrozenSettings}
		sourceID, _ := json.Marshal([]any{sensitiveDigest(b.TxtText), sensitiveDigest(b.WorkingFrontContent), sensitiveCaptureLimit(b, req.FrozenSettings), batchfactorysensitive.ResolveMode(req.FrozenSettings), batchSensitiveModels(req.FrozenSettings)})
		req.SourceKey = sensitiveDigest(string(sourceID))
		w := txt.Workflow{Repo: repo}
		var publishAttempt batchfactoryv11.BookStageRun
		if c.textPublish != nil {
			w.Publish = func(ctx context.Context, r txt.EnsureRequest, collection txt.Collection) (bool, error) {
				r.PublishRunID, r.PublishAttempt, r.PublishRequestKey = publishAttempt.ID, publishAttempt.Attempt, publishAttempt.RequestID
				return c.textPublish(ctx, r, collection)
			}
		}
		w.Track = func(ctx context.Context, s txt.Scope, checkpoint txt.StageCheckpoint, reuse bool, run func() error) error {
			return c.stages.TrackTextStage(ctx, s.Owner, s.BatchID, s.BookID, checkpoint.Stage, checkpoint.Fingerprint, checkpoint.CollectionID, reuse, func(attempt batchfactoryv11.BookStageRun) error {
				if checkpoint.Stage == "text_publish" {
					publishAttempt = attempt
				}
				return run()
			})
		}
		// A legacy manual edit may exist only in the text document. Import its
		// authored draft once before approving or reusing upstream production text.
		doc, err := repo.GetDocument(ctx, req.Scope)
		if err != nil && !errors.Is(err, txt.ErrNotFound) {
			return txt.EnsureResult{Stage: "text_process"}, err
		}
		if req.CollectionID == "" && doc.ActiveVersionID != "" {
			v, err := repo.GetVersion(ctx, req.Scope, doc.ActiveVersionID)
			if err != nil {
				return txt.EnsureResult{Stage: "text_process"}, err
			}
			if _, err = a.adoptLegacyPageEdit(x, doc, v); err != nil {
				return txt.EnsureResult{Stage: "text_process"}, err
			}
		}
		w.Process = func(ctx context.Context, req txt.EnsureRequest, verify bool) (out txt.WorkflowProduction, err error) {
			processed, err := c.batchSensitiveStageFromBook(ctx, req.Owner, x.batch, x.book, req.ForceSensitive, verify)
			if err != nil {
				return out, err
			}
			if processed.SourceMetadata["classificationAfterSensitive"] == true {
				processed, err = c.classifyProductionBook(ctx, req.Owner, req.BatchID, processed, "text")
				if err != nil {
					return out, err
				}
			}
			x.book = processed
			x.book.SettingsState.Patch = nil
			snap, err := bookTextSnapshot(x)
			if err != nil {
				return out, err
			}
			if _, err = a.bgData(x, snap.Background); err != nil {
				return out, err
			}
			snap.SourceVersion = batchfactoryv11.ProductionVersion(processed)
			if snap.SourceVersion == "" {
				return out, txt.ErrConflict
			}
			state := mapOrEmpty(processed.SourceMetadata["batchSensitiveStage"])
			trace, err := json.Marshal(state["changes"])
			if err != nil {
				return out, err
			}
			out.VersionInput = txt.VersionInput{ProductionText: processed.SourceText, SensitiveMode: stringValue(state["mode"]), ChangeTrace: trace}
			out.Snapshot, err = json.Marshal(snap)
			return out, err
		}
		w.Prepare = func(ctx context.Context, req txt.EnsureRequest, prod txt.WorkflowProduction, d txt.Document, reflow bool) error {
			if d.ID != "" && d.OriginalText != x.book.TxtText {
				return txt.ErrConflict
			}
			var snap textSnapshot
			if err := json.Unmarshal(prod.Snapshot, &snap); err != nil {
				return err
			}
			var prev txt.Version
			if d.ActiveVersionID != "" {
				var err error
				prev, err = repo.GetVersion(ctx, req.Scope, d.ActiveVersionID)
				if err != nil {
					return err
				}
			}
			drafts, err := a.prepareDrafts(x, d, snap, prev, prod.ProductionText, reflow)
			if err != nil {
				return err
			}
			_, err = repo.PreparePages(ctx, req.Scope, prod.VersionInput, txt.PaginationUpdate{OriginalText: x.book.TxtText, ExpectedRevision: d.Revision, ConfigSnapshot: prod.Snapshot, Pages: drafts, VerifyProduction: a.productionGuard(x, "")})
			return err
		}
		w.Render = func(ctx context.Context, req txt.EnsureRequest, d txt.Document, p txt.Page) error {
			input, renderer, err := a.renderInput(x, d, p, p.ID == "", txt.PagePatch{})
			if err != nil {
				return err
			}
			data, mime, err := renderer.Render(ctx, input)
			if err != nil {
				return textLayoutError(err)
			}
			rev := p.Revision
			if p.ID == "" {
				rev = d.Revision
			}
			_, err = a.saveImage(x, p.ID, rev, false, "", data, mime)
			return err
		}
		w.Validate = func(_ context.Context, _ txt.EnsureRequest, d txt.Document) error {
			_, err := a.verifyCollectionFiles(x, textInput{VersionID: d.ActiveVersionID, ExpectedRevision: d.Revision, PaginationVersion: d.PaginationVersion})
			return err
		}
		return w.Ensure(ctx, req)
	}
}
