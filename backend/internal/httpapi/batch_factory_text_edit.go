package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"unicode/utf8"

	txt "qiantie/backend/internal/batchfactorytext"
	"qiantie/backend/internal/batchfactoryv11"
)

// Retain the editor's existing breaks whenever the upstream material still matches.
// An explicit reflow request discards manual boundaries and runs the paginator.
func (a textAPI) prepareDrafts(x textRequest, d txt.Document, snap textSnapshot, previous txt.Version, text string, reflow bool) ([]txt.PageDraft, error) {
	measure := (txt.Renderer{}).MeasureFunc(x)
	if d.ActiveVersionID != "" {
		var old textSnapshot
		if err := json.Unmarshal(d.ConfigSnapshot, &old); err != nil {
			return nil, err
		}
		if !reflow && previous.ProductionText == text {
			ps, err := a.repo.ListPages(x, x.scope)
			if err != nil {
				return nil, err
			}
			drafts := make([]txt.PageDraft, 0, len(ps))
			for _, p := range ps {
				if p.VersionID != previous.ID {
					return nil, txt.ErrConflict
				}
				_, style, err := txt.ResolvePageStyle(snap.PageStyle, p.PageDraft, txt.PagePatch{})
				if err != nil {
					return nil, err
				}
				_, previous, err := txt.ResolvePageStyle(old.PageStyle, p.PageDraft, txt.PagePatch{})
				if err != nil {
					return nil, err
				}
				oldLayout := previous.Layout
				// Both colors affect pixels only and cannot change line breaks.
				oldLayout.TextColor = style.Layout.TextColor
				oldLayout.HighlightColor = style.Layout.HighlightColor
				if oldLayout != style.Layout {
					fits, _, err := measure(p.Text, style.Layout)
					if err != nil {
						return nil, textLayoutError(err)
					}
					if !fits {
						return nil, fmt.Errorf("%w：第 %d 页文字放不下，请调整字号或选择重新分页", txt.ErrConflict, p.Ordinal)
					}
				}
				drafts = append(drafts, p.PageDraft)
			}
			return drafts, nil
		}
	}
	pages, err := txt.Paginate(x, txt.PaginationInput{Text: text, Layout: snap.Layout}, measure)
	if err != nil {
		return nil, textLayoutError(err)
	}
	return pages, nil
}

func (a textAPI) savePageText(w http.ResponseWriter, r *http.Request, x textRequest, d txt.Document, edited txt.Page, patch txt.PagePatch, reflow bool) error {
	var snap textSnapshot
	if err := json.Unmarshal(d.ConfigSnapshot, &snap); err != nil {
		return err
	}
	v, err := a.repo.GetVersion(x, x.scope, d.ActiveVersionID)
	if err != nil {
		return err
	}
	key := batchfactoryv11.ProductionVersion(x.book)
	if key == "" || (snap.SourceVersion != "" && snap.SourceVersion != key) || (snap.SourceVersion == "" && v.ContentDigest != sensitiveDigest(x.book.SourceText)) {
		return txt.ErrConflict
	}
	snap.SourceVersion = key
	ps, err := a.repo.ListPages(x, x.scope)
	if err != nil {
		return err
	}
	if len(ps) == 0 || x.book.TxtText != d.OriginalText {
		return txt.ErrConflict
	}
	var joined strings.Builder
	drafts := make([]txt.PageDraft, 0, len(ps))
	var offset int
	for _, p := range ps {
		if p.VersionID != v.ID {
			return txt.ErrConflict
		}
		draft := p.PageDraft
		replacement := []txt.PageDraft{draft}
		if p.ID == edited.ID {
			if p.Revision != edited.Revision || p.Locked {
				return txt.ErrConflict
			}
			overrides, style, err := txt.ResolvePageStyle(snap.PageStyle, draft, patch)
			if err != nil {
				return err
			}
			if reflow {
				replacement, err = txt.Paginate(x, txt.PaginationInput{Text: edited.Text, Layout: style.Layout}, (txt.Renderer{}).MeasureFunc(x))
				if err != nil {
					return textLayoutError(err)
				}
			} else {
				replacement[0].Text = edited.Text
			}
			for i := range replacement {
				replacement[i].Layout, replacement[i].Background = overrides.Layout, overrides.Background
				replacement[i].ClearLayout, replacement[i].ClearBackground = patch.ClearLayout, patch.ClearBackground
			}
		}
		for _, part := range replacement {
			joined.WriteString(part.Text)
			part.StartOffset = offset
			offset += utf8.RuneCountInString(part.Text)
			part.EndOffset = offset
			drafts = append(drafts, part)
		}
	}
	text := joined.String()
	// A form may resend the displayed production text alongside a style change.
	// Preserve the independent author draft when the wording is unchanged.
	update := txt.PaginationUpdate{ExpectedRevision: d.Revision, Pages: drafts, ConfigSnapshot: d.ConfigSnapshot, VerifyProduction: a.productionGuard(x, snap.AuthorDraftKey)}
	if text == txt.NormalizeLineEndings(v.ProductionText) {
		if patch == (txt.PagePatch{}) && !reflow {
			return a.list(w, r, x)
		}
	} else {
		breaks := txt.SentenceBreaks(text)
		for _, draft := range drafts[:len(drafts)-1] {
			if _, valid := slices.BinarySearch(breaks, draft.EndOffset); !valid {
				return fmt.Errorf("%w：请在页末保留完整的句子或段落", txt.ErrInvalid)
			}
		}
		author, err := a.authorForEdit(x, snap.AuthorDraftKey)
		if err != nil {
			return err
		}
		if snap.AuthorDraftKey == "" && txt.NormalizeLineEndings(v.ProductionText) != txt.NormalizeLineEndings(x.book.SourceText) {
			// Earlier editor versions retained a local displayed version only.
			author, err = txt.MergeAuthorPageEdit(author, txt.Version{ProductionText: x.book.SourceText, ChangeTrace: v.ChangeTrace}, 0, len([]rune(txt.NormalizeLineEndings(x.book.SourceText))), v.ProductionText)
			if err != nil {
				return err
			}
		}
		base := v
		if snap.AuthorDraftKey != "" || v.ProductionText != x.book.SourceText {
			base.ChangeTrace = nil
		}
		merged, err := txt.MergeAuthorPageEdit(author, base, edited.StartOffset, edited.EndOffset, edited.Text)
		if err != nil {
			return err
		}
		update.VerifyProduction = a.authorDraftWriter(x, snap.AuthorDraftKey, merged)
		snap.AuthorDraftKey = batchfactoryv11.WorkingDraftVersion(merged)
		update.ConfigSnapshot, err = json.Marshal(snap)
		if err != nil {
			return err
		}
		v.ProductionText = text
	}
	_, err = a.repo.PreparePages(x, x.scope, txt.VersionInput{ProductionText: v.ProductionText, SensitiveMode: v.SensitiveMode, ChangeTrace: v.ChangeTrace}, update)
	if err != nil {
		return err
	}
	return a.list(w, r, x)
}

// Older page edits created a local version without saving the working draft. Adopt
// it once, only when both its upstream version and its underlying input are still current.
func (a textAPI) adoptLegacyPageEdit(x textRequest, doc txt.Document, v txt.Version) (txt.Document, error) {
	if txt.NormalizeLineEndings(v.ProductionText) == txt.NormalizeLineEndings(x.book.SourceText) {
		return doc, nil
	}
	var snap textSnapshot
	if err := json.Unmarshal(doc.ConfigSnapshot, &snap); err != nil {
		return doc, err
	}
	if snap.AuthorDraftKey != "" || snap.SourceVersion == "" || snap.SourceVersion != batchfactoryv11.ProductionVersion(x.book) {
		return doc, nil
	}
	draft, err := a.store.GetDraft(x, x.scope.Owner, "working-front:"+x.scope.BookID, "working-front-content", x.scope.BatchID)
	if err != nil && !errors.Is(err, batchfactoryv11.ErrNotFound) {
		return doc, err
	}
	base := stringValue(mapOrEmpty(x.book.SourceMetadata["batchSensitiveStage"])["authoredBase"])
	if strings.TrimSpace(draft.Content) != "" && base != batchfactoryv11.WorkingDraftVersion(draft.Content) || strings.TrimSpace(draft.Content) == "" && strings.HasPrefix(base, "working:") {
		return doc, nil
	}
	ps, err := a.repo.ListPages(x, x.scope)
	if err != nil {
		return doc, err
	}
	pages := make([]txt.PageDraft, 0, len(ps))
	for _, p := range ps {
		if p.VersionID != v.ID {
			return doc, txt.ErrConflict
		}
		pages = append(pages, p.PageDraft)
	}
	author, err := a.authorForEdit(x, "")
	if err != nil {
		return doc, err
	}
	merged, err := txt.MergeAuthorPageEdit(author, txt.Version{ProductionText: x.book.SourceText, ChangeTrace: v.ChangeTrace}, 0, len([]rune(txt.NormalizeLineEndings(x.book.SourceText))), v.ProductionText)
	if err != nil {
		return doc, err
	}
	snap.AuthorDraftKey = batchfactoryv11.WorkingDraftVersion(merged)
	cfg, err := json.Marshal(snap)
	if err != nil {
		return doc, err
	}
	_, err = a.repo.PreparePages(x, x.scope, txt.VersionInput{ProductionText: v.ProductionText, SensitiveMode: v.SensitiveMode, ChangeTrace: v.ChangeTrace}, txt.PaginationUpdate{ExpectedRevision: doc.Revision, Pages: pages, ConfigSnapshot: cfg, VerifyProduction: a.authorDraftWriter(x, "", merged)})
	if err != nil {
		return doc, err
	}
	// This transaction reused the immutable version, changed pagination once
	// and compared the initially captured document revision.
	doc.Revision++
	doc.PaginationVersion++
	doc.ConfigSnapshot = cfg
	return doc, nil
}

// Read exactly the authored input used by the verified policy stage. A pending
// page edit instead names its working draft by digest. This path never writes.
func (a textAPI) authorForEdit(x textRequest, pending string) (string, error) {
	state := mapOrEmpty(x.book.SourceMetadata["batchSensitiveStage"])
	key := stringValue(state["authoredBase"])
	if pending != "" {
		key = pending
	}
	var content string
	if key == "" {
		effective := a.production.automationEffectiveSettings(x.batch, x.book)
		content = capturedSensitiveInput(x.book, sensitiveCaptureLimit(x.book, effective))
	} else {
		draftKey, kind, scope := key, "sensitive_source", x.scope.BookID
		if strings.HasPrefix(key, "working:") {
			draftKey, kind, scope = "working-front:"+x.scope.BookID, "working-front-content", x.scope.BatchID
		}
		d, err := a.store.GetDraft(x, x.scope.Owner, draftKey, kind, scope)
		if errors.Is(err, batchfactoryv11.ErrNotFound) {
			return "", txt.ErrConflict
		}
		if err != nil {
			return "", err
		}
		content = d.Content
		if strings.TrimPrefix(key, "working:") != sensitiveDigest(content) {
			return "", txt.ErrConflict
		}
	}
	if pending == "" && state["input"] != sensitiveDigest(content) {
		return "", fmt.Errorf("%w：作者稿已变化，请重新准备", txt.ErrConflict)
	}
	return content, nil
}

func (a textAPI) authorDraftWriter(x textRequest, pending, content string) func(context.Context, *sql.Tx) error {
	return func(ctx context.Context, tx *sql.Tx) error {
		if err := a.productionGuard(x, pending)(ctx, tx); err != nil {
			return err
		}
		writer, ok := a.store.(interface {
			SaveDraftTx(context.Context, *sql.Tx, string, batchfactoryv11.Draft) (batchfactoryv11.Draft, error)
		})
		if !ok {
			return batchfactoryv11.ErrUnavailable
		}
		_, err := writer.SaveDraftTx(ctx, tx, x.scope.Owner, batchfactoryv11.Draft{Key: "working-front:" + x.scope.BookID, Kind: "working-front-content", Scope: x.scope.BatchID, Content: content})
		return err
	}
}

// The MySQL check locks the authoritative book record and working draft in the
// same transaction as activation. Neither can change between this read and commit.
func (a textAPI) productionGuard(x textRequest, pendingDraft string) func(context.Context, *sql.Tx) error {
	pinned := x.book
	return func(ctx context.Context, tx *sql.Tx) error {
		current := pinned
		var draft string
		if store, ok := a.store.(interface {
			LockProductionSource(context.Context, *sql.Tx, string, string, string) (batchfactoryv11.Book, string, error)
		}); ok {
			var err error
			current, draft, err = store.LockProductionSource(ctx, tx, x.scope.Owner, x.scope.BatchID, x.scope.BookID)
			if errors.Is(err, batchfactoryv11.ErrNotFound) {
				return txt.ErrConflict
			}
			if err != nil {
				return err
			}
		} else {
			b, err := a.store.GetBatch(ctx, x.scope.Owner, x.scope.BatchID)
			if err != nil {
				return err
			}
			var found bool
			current, found = findBatchBookByID(b, x.scope.BookID)
			if !found {
				return txt.ErrConflict
			}
			d, err := a.store.GetDraft(ctx, x.scope.Owner, "working-front:"+x.scope.BookID, "working-front-content", x.scope.BatchID)
			if err != nil && !errors.Is(err, batchfactoryv11.ErrNotFound) {
				return err
			}
			draft = d.Content
		}
		state := mapOrEmpty(current.SourceMetadata["batchSensitiveStage"])
		base := stringValue(state["authoredBase"])
		hasDraft := strings.TrimSpace(draft) != ""
		currentWorking := batchfactoryv11.WorkingDraftVersion(draft)
		expectedWorking := base
		if pendingDraft != "" {
			expectedWorking = pendingDraft
		}
		approvedWorking := expectedWorking == currentWorking
		if !hasVerifiedBatchOriginal(current) || batchfactoryv11.ProductionVersion(current) != batchfactoryv11.ProductionVersion(pinned) || current.SourceText != pinned.SourceText || current.TxtText != pinned.TxtText ||
			state["output"] != sensitiveDigest(current.SourceText) || state["raw"] != sensitiveDigest(current.TxtText) || (hasDraft && !approvedWorking) || (!hasDraft && (pendingDraft != "" || strings.HasPrefix(base, "working:"))) {
			return fmt.Errorf("%w：生产正文已更新，请重新准备", txt.ErrConflict)
		}
		return nil
	}
}
