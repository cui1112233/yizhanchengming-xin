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
func (a textAPI) prepareDrafts(x textRequest, d txt.Document, snap textSnapshot, version *txt.VersionInput, reflow bool) ([]txt.PageDraft, error) {
	measure := (txt.Renderer{}).MeasureFunc(x)
	if d.ActiveVersionID != "" {
		var old textSnapshot
		if err := json.Unmarshal(d.ConfigSnapshot, &old); err != nil {
			return nil, err
		}
		v, err := a.repo.GetVersion(x, x.scope, d.ActiveVersionID)
		if err != nil {
			return nil, err
		}
		if old.SourceVersion == snap.SourceVersion {
			// Manual page edits form a derived version of this exact upstream text.
			version.ProductionText = v.ProductionText
		}
		if !reflow && v.ProductionText == version.ProductionText {
			ps, err := a.repo.ListPages(x, x.scope)
			if err != nil {
				return nil, err
			}
			drafts := make([]txt.PageDraft, 0, len(ps))
			for _, p := range ps {
				if p.VersionID != v.ID {
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
	pages, err := txt.Paginate(x, txt.PaginationInput{Text: version.ProductionText, Layout: snap.Layout}, measure)
	if err != nil {
		return nil, textLayoutError(err)
	}
	return pages, nil
}

func (a textAPI) savePageText(w http.ResponseWriter, r *http.Request, x textRequest, d txt.Document, edited txt.Page, patch txt.PagePatch) error {
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
	if len(ps) == 0 {
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
		if p.ID == edited.ID {
			if p.Revision != edited.Revision || p.Locked {
				return txt.ErrConflict
			}
			overrides, _, err := txt.ResolvePageStyle(snap.PageStyle, draft, patch)
			if err != nil {
				return err
			}
			draft.Text = edited.Text
			draft.Layout, draft.Background = overrides.Layout, overrides.Background
			draft.ClearLayout, draft.ClearBackground = patch.ClearLayout, patch.ClearBackground
		}
		joined.WriteString(draft.Text)
		draft.StartOffset = offset
		offset += utf8.RuneCountInString(draft.Text)
		draft.EndOffset = offset
		drafts = append(drafts, draft)
	}
	text := joined.String()
	breaks := txt.SentenceBreaks(text)
	for _, draft := range drafts[:len(drafts)-1] {
		if !slices.Contains(breaks, draft.EndOffset) {
			return fmt.Errorf("%w：请在页末保留完整的句子或段落", txt.ErrInvalid)
		}
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	_, err = a.repo.PreparePages(x, x.scope, txt.VersionInput{ProductionText: text, SensitiveMode: v.SensitiveMode, ChangeTrace: v.ChangeTrace}, txt.PaginationUpdate{ExpectedRevision: d.Revision, Pages: drafts, ConfigSnapshot: raw, VerifyProduction: a.productionGuard(x)})
	if err != nil {
		return err
	}
	return a.list(w, r, x)
}

// The MySQL check locks the authoritative book record and working draft in the
// same transaction as activation. Neither can change between this read and commit.
func (a textAPI) productionGuard(x textRequest) func(context.Context, *sql.Tx) error {
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
		if !hasVerifiedBatchOriginal(current) || batchfactoryv11.ProductionVersion(current) != batchfactoryv11.ProductionVersion(pinned) || current.SourceText != pinned.SourceText || current.TxtText != pinned.TxtText ||
			state["output"] != sensitiveDigest(current.SourceText) || state["raw"] != sensitiveDigest(current.TxtText) || (hasDraft && base != batchfactoryv11.WorkingDraftVersion(draft)) || (!hasDraft && strings.HasPrefix(base, "working:")) {
			return fmt.Errorf("%w：生产正文已更新，请重新准备", txt.ErrConflict)
		}
		return nil
	}
}
