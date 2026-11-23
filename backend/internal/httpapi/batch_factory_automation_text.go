package httpapi

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"qiantie/backend/internal/batchfactorysensitive"
	txt "qiantie/backend/internal/batchfactorytext"
	"qiantie/backend/internal/batchfactoryv11"
	cp "qiantie/backend/internal/collectionpublish"
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
	var latest batchfactoryv11.BookStageRun
	if len(uploads) > 0 {
		latest = uploads[0]
	}
	ids := []string{manualCollectionID}
	for _, r := range uploads {
		ids = append(ids, r.CollectionID)
	}
	seen := map[string]bool{}
	manualCurrent := false
	var historyTime time.Time
	for _, id := range ids {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		draft, err := store.GetDraft(ctx, owner, id, textPublishDraftKind, batchID)
		if errors.Is(err, batchfactoryv11.ErrNotFound) {
			continue
		}
		if err != nil {
			return err
		}
		var rec cp.Record
		if err = json.Unmarshal([]byte(draft.Content), &rec); err != nil {
			return err
		}
		scoped := func(record *cp.Record) bool {
			return record != nil && record.Kind == cp.CollectionKindText && record.BookID == bookID && record.CollectionID == id
		}
		if !scoped(&rec) {
			continue
		}
		if id == manualCollectionID && rec.RunID == "" && rec.Attempt == 0 {
			finished := latest.Status == batchfactoryv11.ProductionSucceeded || latest.Status == batchfactoryv11.ProductionFailed
			knownKey := rec.RequestKey != "" && slices.ContainsFunc(uploads, func(r batchfactoryv11.BookStageRun) bool { return rec.RequestKey == r.RequestID })
			manualCurrent = !knownKey && (latest.ID == "" || finished && draft.UpdatedAt.After(latest.UpdatedAt))
		}
		if id == manualCollectionID && manualCurrent || !manualCurrent && matchingWorkflowPublishAttempt(rec, latest) {
			if rec.Status == "pending" {
				rec.Status = "upload_pending_confirm"
			}
			summary["textPublishRecord"] = rec
			if manualCurrent {
				summary["textPublishSource"] = "manual"
			}
		}
		history := &rec
		if !rec.Confirmed() {
			history = rec.LastPublished
		}
		if scoped(history) && history.Confirmed() {
			published := history.PublishedAt
			if published.IsZero() {
				published = draft.UpdatedAt
			}
			if summary["textLastPublishedRecord"] == nil || published.After(historyTime) {
				summary["textLastPublishedRecord"] = *history
				historyTime = published
			}
		}
	}
	return nil
}

func matchingWorkflowPublishAttempt(rec cp.Record, run batchfactoryv11.BookStageRun) bool {
	if rec.CollectionID != run.CollectionID || run.ID == "" {
		return false
	}
	hasRun := rec.RunID != ""
	hasAttempt := rec.Attempt > 0 && run.Attempt > 0
	hasKey := rec.RequestKey != "" && run.RequestID != ""
	if hasRun && rec.RunID != run.ID || hasAttempt && rec.Attempt != run.Attempt || hasKey && rec.RequestKey != run.RequestID {
		return false
	}
	return hasRun || hasAttempt || hasKey
}

// A directory sync can change many profiles. Only the selected profile and
// the choices consumed by the text publisher identify this upload attempt.
func workflowPublishSettingsKey(settings batchfactoryv11.SettingsPatch) (string, error) {
	var config n8ComicPublishSettings
	var err error
	if len(settings["comicPublishSettings"]) > 0 {
		config, err = parseN8ComicPublishSettings(settings)
		if err != nil {
			return "", err
		}
	}
	var website map[string]any
	if raw := settings["publishSettings"]; len(raw) > 0 {
		if err := json.Unmarshal(raw, &website); err != nil {
			return "", err
		}
	}
	profile := selectedPublishProfile(website)
	selected, err := json.Marshal([]any{
		website["websiteProfileId"], website["bookCategory"], profile["id"], profile["config_id"],
		config.ProducerAdminName, config.OrganizationID, config.AudioCategoryID, config.AudioCategoryName,
		config.AudioCategorySelectionMode, config.AudioID, config.SpecifiedAudioID,
		config.WorkIDs, config.WorkIDsText, config.WorkMatchMode,
	})
	if err != nil {
		return "", err
	}
	return sensitiveDigest(string(selected)), nil
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
		if req.Target == "published" {
			req.PublishSettingsKey, err = workflowPublishSettingsKey(req.FrozenSettings)
			if err != nil {
				return txt.EnsureResult{Stage: "text_publish"}, err
			}
		}
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
		w.VerifyFile = func(_ context.Context, artifact txt.PageArtifact, label string) error {
			return a.verifyArtifactFile(x, artifact, label)
		}
		w.Validate = func(_ context.Context, _ txt.EnsureRequest, d txt.Document) error {
			_, err := a.verifyCollectionFiles(x, textInput{VersionID: d.ActiveVersionID, ExpectedRevision: d.Revision, PaginationVersion: d.PaginationVersion})
			return err
		}
		return w.Ensure(ctx, req)
	}
}
