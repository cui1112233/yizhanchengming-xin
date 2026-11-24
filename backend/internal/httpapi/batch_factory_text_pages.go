package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	txt "qiantie/backend/internal/batchfactorytext"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/renderfont"
)

type textAPI struct {
	store      batchfactoryv11.Store
	repo       txt.Repository
	files      *localartifact.Store
	production *BatchAutomationCoordinator
}
type textRequest struct {
	context.Context
	scope    txt.Scope
	batch    batchfactoryv11.Batch
	book     batchfactoryv11.Book
	settings batchfactoryv11.SettingsPatch
}

func (x textRequest) base() string {
	if x.scope.BookID == "" {
		return "/api/batch-factory/v12/text-"
	}
	return "/api/batch-factory/v12/batches/" + url.PathEscape(x.scope.BatchID) + "/books/" + url.PathEscape(x.scope.BookID) + "/text-"
}

type textHandler func(http.ResponseWriter, *http.Request, textRequest) error

func (a textAPI) scoped(fn textHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, 401, map[string]string{"error": "请登录"})
			return
		}
		s := txt.Scope{Owner: owner, BatchID: r.PathValue("batchId"), BookID: r.PathValue("bookId")}
		b, err := a.store.GetBatch(r.Context(), s.Owner, s.BatchID)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		book, ok := findBatchBookByID(b, s.BookID)
		if !ok {
			writeStoreError(w, batchfactoryv11.ErrNotFound)
			return
		}
		if batchfactoryv11.EffectiveProductionMode(b.SettingsState.Patch, book.SettingsState.Patch) != "text" {
			writeJSON(w, 409, map[string]string{"error": "当前书籍不是文本生成模式"})
			return
		}
		if a.repo == nil || a.files == nil {
			writeJSON(w, 503, map[string]string{"error": "文本生产存储尚未配置"})
			return
		}
		settings := effectiveTextSettings(b.SettingsState.Patch, book.SettingsState.Patch)
		if err = fn(w, r, textRequest{r.Context(), s, b, book, settings}); err != nil {
			writeTextError(w, err)
		}
	}
}
func writeTextError(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	var overflow *txt.SentenceOverflowError
	switch {
	case errors.Is(err, txt.ErrNotFound):
		code = 404
	case errors.Is(err, txt.ErrLocked), errors.Is(err, txt.ErrNotReady), errors.Is(err, txt.ErrConflict):
		code = 409
	case errors.Is(err, txt.ErrInvalid), errors.As(err, &overflow):
		code = 400
	case errors.Is(err, localartifact.ErrTooLarge):
		code = 413
	case errors.Is(err, localartifact.ErrDiskPressure):
		code = 507
	}
	if code == 500 {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, code, map[string]string{"error": err.Error()})
}
func registerBatchTextRoutes(m *http.ServeMux, opt RouterOptions, c *BatchAutomationCoordinator) {
	a := textAPI{opt.Store, opt.TextRepository, opt.LocalArtifacts, c}
	for pattern, fn := range map[string]textHandler{
		"GET /api/batch-factory/v12/text-backgrounds":                        a.backgrounds,
		"POST /api/batch-factory/v12/text-backgrounds":                       a.upload,
		"GET /api/batch-factory/v12/text-backgrounds/{backgroundId}/content": a.content,
	} {
		m.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			owner, ok := bridgeOwner(r)
			if !ok {
				writeJSON(w, 401, map[string]string{"error": "请登录"})
				return
			}
			if a.repo == nil || a.files == nil {
				writeJSON(w, 503, map[string]string{"error": "文本生产存储尚未配置"})
				return
			}
			if err := fn(w, r, textRequest{Context: r.Context(), scope: txt.Scope{Owner: owner}}); err != nil {
				writeTextError(w, err)
			}
		})
	}
	const base = "/api/batch-factory/v12/batches/{batchId}/books/{bookId}/text-"
	for _, route := range []struct {
		method, path string
		fn           textHandler
	}{
		{"GET", "pages", a.list}, {"POST", "prepare", a.prepare}, {"PATCH", "pages/{pageId}", a.patch},
		{"POST", "pages/{pageId}/split", a.reflow}, {"POST", "pages/{pageId}/merge", a.reflow},
		{"POST", "pages/{pageId}/preview", a.picture}, {"POST", "pages/{pageId}/render", a.picture}, {"POST", "pages/{pageId}/upload", a.upload}, {"POST", "pages/{pageId}/restore", a.restore},
		{"POST", "cover/preview", a.picture}, {"POST", "cover/render", a.picture}, {"POST", "cover/upload", a.upload}, {"POST", "cover/restore", a.restore},
		{"GET", "backgrounds", a.backgrounds}, {"POST", "backgrounds", a.upload}, {"GET", "backgrounds/{backgroundId}/content", a.content},
		{"GET", "artifacts/{artifactId}/content", a.content}, {"GET", "collections", a.collections}, {"POST", "collections", a.collections}, {"GET", "collections/{collectionId}", a.collections},
	} {
		m.HandleFunc(route.method+" "+base+route.path, a.scoped(route.fn))
	}
}

type textInput struct {
	ExpectedRevision         int64   `json:"expectedRevision"`
	ExpectedDocumentRevision int64   `json:"expectedDocumentRevision"`
	Text                     *string `json:"text,omitempty"`
	Reflow                   bool    `json:"reflow,omitempty"`
	Offset                   int     `json:"offset"`
	txt.PagePatch
	VersionID         string   `json:"versionId"`
	PaginationVersion int64    `json:"paginationVersion"`
	PageIDs           []string `json:"pageIds"`
}

type textSnapshot struct {
	txt.PageStyle
	Cover            txt.PagePatch `json:"cover,omitempty"`
	Title            string        `json:"title"`
	CoverFingerprint string        `json:"coverFingerprint"`
	SourceVersion    string        `json:"sourceVersion,omitempty"`
	AuthorDraftKey   string        `json:"authorDraftKey,omitempty"`
	HighlightWords   []string      `json:"highlightWords,omitempty"`
}

func bookTextSnapshot(x textRequest) (textSnapshot, error) {
	var s textSnapshot
	raw := map[string]json.RawMessage{}
	for k, d := range map[string]string{"textPageTarget": "targetCharacters", "imageAspectRatio": "aspectRatio", "textFontId": "fontId", "textFontSize": "fontSize", "textFontWeight": "fontWeight", "textLineHeight": "lineHeight", "textParagraphSpacing": "paragraphSpacing", "textPadding": "padding", "textAlign": "align", "textColor": "textColor", "textHighlightColor": "highlightColor", "textMaxHeightRatio": "maxHeightRatio"} {
		if v := x.settings[k]; len(v) > 0 {
			raw[d] = v
		}
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return s, txt.ErrInvalid
	}
	if err := json.Unmarshal(encoded, &s.Layout); err != nil {
		return s, fmt.Errorf("%w：排版配置无效", txt.ErrInvalid)
	}
	if v := x.settings["textBackground"]; len(v) > 0 {
		var bg struct {
			txt.Background
			Type string `json:"type"`
		}
		if err := json.Unmarshal(v, &bg); err != nil {
			return s, txt.ErrInvalid
		}
		s.Background = bg.Background
		if s.Background.Kind == "" {
			s.Background.Kind = bg.Type
		}
	}
	if v := x.settings["textCover"]; len(v) > 0 {
		if err := json.Unmarshal(v, &s.Cover); err != nil {
			return s, txt.ErrInvalid
		}
		if s.Cover.Layout != nil {
			var cover struct {
				Layout map[string]json.RawMessage `json:"layout"`
			}
			if err := json.Unmarshal(v, &cover); err != nil {
				return s, txt.ErrInvalid
			}
			for k, val := range cover.Layout {
				raw[k] = val
			}
			encoded, err := json.Marshal(raw)
			if err != nil {
				return s, txt.ErrInvalid
			}
			var merged txt.LayoutConfig
			if err = json.Unmarshal(encoded, &merged); err != nil {
				return s, txt.ErrInvalid
			}
			s.Cover.Layout = &merged
		}
	}
	// Cover and body share the book background; ignore legacy cover overrides.
	s.Cover.Background = nil
	s.Cover.ClearBackground = false
	if v := x.settings["highlightReuseEnabled"]; len(v) > 0 {
		if err := json.Unmarshal(v, &s.HighlightReuseEnabled); err != nil {
			return s, txt.ErrInvalid
		}
	} else {
		var legacy struct {
			Enabled bool `json:"biaohongReuse"`
		}
		_ = json.Unmarshal(x.settings["publishSettings"], &legacy)
		s.HighlightReuseEnabled = legacy.Enabled
	}
	if s.HighlightReuseEnabled {
		words, err := json.Marshal(x.book.SourceMetadata["illustrationHighlightKeywords"])
		if err != nil {
			return s, txt.ErrInvalid
		}
		_ = json.Unmarshal(words, &s.HighlightWords)
		if len(s.HighlightWords) > 0 {
			s.HighlightRevision = sensitiveDigest(string(words))
		}
	}
	s.Title = x.book.Title
	_, coverStyle, err := txt.ResolvePageStyle(s.PageStyle, txt.PageDraft{}, s.Cover)
	if err != nil {
		return s, err
	}
	fingerprint, err := json.Marshal([]any{s.Title, coverStyle.Layout, coverStyle.Background})
	if err != nil {
		return s, txt.ErrInvalid
	}
	s.CoverFingerprint = sensitiveDigest(string(fingerprint))
	return s, nil
}
func (a textAPI) prepare(w http.ResponseWriter, r *http.Request, x textRequest) error {
	var in textInput
	if !decodeJSON(w, r, &in) {
		return nil
	}
	if in.ExpectedRevision < 0 {
		return txt.ErrInvalid
	}
	doc, err := a.repo.GetDocument(x, x.scope)
	absent := errors.Is(err, txt.ErrNotFound)
	if err != nil && !absent {
		return err
	}
	if !absent && doc.OriginalText != x.book.TxtText {
		return txt.ErrConflict
	}
	if in.ExpectedRevision > 0 && in.ExpectedRevision != doc.Revision {
		return txt.ErrConflict
	}
	previous := txt.Version{}
	if doc.ActiveVersionID != "" {
		previous, err = a.repo.GetVersion(x, x.scope, doc.ActiveVersionID)
		if err != nil {
			return err
		}
		doc, err = a.adoptLegacyPageEdit(x, doc, previous)
		if err != nil {
			return err
		}
	}
	b, err := a.production.batchSensitiveStageFromBook(x, x.scope.Owner, x.batch, x.book, false, false)
	if err != nil {
		return err
	}
	x.book = b
	snap, err := bookTextSnapshot(x)
	if err != nil {
		return err
	}
	if _, err := a.bgData(x, snap.Background); err != nil {
		return err
	}
	if !absent && doc.OriginalText != b.TxtText {
		return txt.ErrConflict
	}
	snap.SourceVersion = batchfactoryv11.ProductionVersion(b)
	if snap.SourceVersion == "" {
		return txt.ErrConflict
	}
	state := mapOrEmpty(b.SourceMetadata["batchSensitiveStage"])
	trace, err := json.Marshal(state["changes"])
	if err != nil {
		return err
	}
	snapshot, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	version := txt.VersionInput{ProductionText: b.SourceText, SensitiveMode: stringValue(state["mode"]), ChangeTrace: trace}
	drafts, err := a.prepareDrafts(x, doc, snap, previous, version.ProductionText, in.Reflow)
	if err != nil {
		return err
	}
	_, err = a.repo.PreparePages(x, x.scope, version, txt.PaginationUpdate{OriginalText: b.TxtText, ExpectedRevision: doc.Revision, ConfigSnapshot: snapshot, Pages: drafts, VerifyProduction: a.productionGuard(x, "")})
	if err != nil {
		return err
	}
	return a.list(w, r, x)
}
func textLayoutError(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return fmt.Errorf("%w：%v", txt.ErrInvalid, err)
}
func (a textAPI) list(w http.ResponseWriter, _ *http.Request, x textRequest) error {
	return a.listAfterEdit(w, x, nil, "")
}

func (a textAPI) listAfterEdit(w http.ResponseWriter, x textRequest, previous []txt.Page, editedID string) error {
	d, err := a.repo.GetDocument(x, x.scope)
	if errors.Is(err, txt.ErrNotFound) {
		writeJSON(w, 200, map[string]any{"document": nil, "sentenceBreaks": []int{}, "pages": []txt.Page{}, "artifacts": []textMediaView{}, "fonts": renderfont.Catalog()})
		return nil
	}
	if err != nil {
		return err
	}
	ps, err := a.repo.ListPages(x, x.scope)
	if err != nil {
		return err
	}
	as, err := a.repo.ListArtifacts(x, x.scope)
	if err != nil {
		return err
	}
	views := make([]textMediaView, 0, len(as))
	for _, a := range as {
		views = append(views, artifactView(x, a))
	}
	var version *txt.Version
	breaks := []int{}
	if d.ActiveVersionID != "" {
		v, err := a.repo.GetVersion(x, x.scope, d.ActiveVersionID)
		if err != nil {
			return err
		}
		version = &v
		breaks = append(breaks, txt.SentenceBreaks(v.ProductionText)...)
	}
	latest, err := a.repo.GetDocument(x, x.scope)
	if err != nil {
		return err
	}
	if d.Revision != latest.Revision {
		return txt.ErrConflict
	}
	result := map[string]any{"document": d, "version": version, "sentenceBreaks": breaks, "pages": ps, "artifacts": views, "fonts": renderfont.Catalog()}
	if previous != nil {
		before := make(map[string]txt.Page, len(previous))
		for _, p := range previous {
			before[p.ID] = p
		}
		affected := []string{}
		for _, p := range ps {
			old, exists := before[p.ID]
			if !p.Locked && p.Status != "ready" && (!exists || p.ID == editedID || old.Status == "ready" || old.ActiveArtifactID != "") {
				affected = append(affected, p.ID)
			}
		}
		result["affectedPageIds"] = affected
	}
	writeJSON(w, 200, result)
	return nil
}
func (a textAPI) page(x textRequest, id string) (txt.Page, error) {
	ps, err := a.repo.ListPages(x, x.scope)
	if err != nil {
		return txt.Page{}, err
	}
	for _, p := range ps {
		if p.ID == id {
			return p, nil
		}
	}
	return txt.Page{}, txt.ErrNotFound
}
func (a textAPI) patch(w http.ResponseWriter, r *http.Request, x textRequest) error {
	var in textInput
	if !decodeJSON(w, r, &in) {
		return nil
	}
	doc, err := a.repo.GetDocument(x, x.scope)
	if err != nil {
		return err
	}
	p, err := a.page(x, r.PathValue("pageId"))
	if err != nil {
		return err
	}
	if p.Revision != in.ExpectedRevision || (in.ExpectedDocumentRevision != 0 && in.ExpectedDocumentRevision != doc.Revision) {
		return txt.ErrConflict
	}
	if in.Text != nil {
		if in.ExpectedDocumentRevision != doc.Revision || p.VersionID != doc.ActiveVersionID {
			return txt.ErrConflict
		}
		if p.Locked {
			return txt.ErrLocked
		}
		if strings.TrimSpace(*in.Text) == "" {
			return txt.ErrInvalid
		}
		p.Text = txt.NormalizeLineEndings(*in.Text)
	}
	input, _, err := a.renderInput(x, doc, p, false, in.PagePatch)
	if err != nil {
		return err
	}
	if in.Reflow {
		if in.Text == nil {
			return fmt.Errorf("%w：局部分页需要本页正文", txt.ErrInvalid)
		}
		return a.savePageText(w, r, x, doc, p, in.PagePatch, true)
	}
	fits, _, err := (txt.Renderer{}).MeasureFunc(x)(input.Text, input.Layout)
	if err != nil {
		return textLayoutError(err)
	}
	if !fits {
		return fmt.Errorf("%w：当前文字放不下，请调整字号或分页", txt.ErrInvalid)
	}
	if in.Text != nil {
		return a.savePageText(w, r, x, doc, p, in.PagePatch, false)
	}
	in.PagePatch.DocumentRevision = in.ExpectedDocumentRevision
	result, err := a.repo.PatchPage(x, x.scope, p.ID, in.ExpectedRevision, in.PagePatch)
	if err != nil {
		return err
	}
	writeJSON(w, 200, result)
	return nil
}
func (a textAPI) reflow(w http.ResponseWriter, r *http.Request, x textRequest) error {
	var in textInput
	if !decodeJSON(w, r, &in) {
		return nil
	}
	doc, err := a.repo.GetDocument(x, x.scope)
	if err != nil {
		return err
	}
	if in.ExpectedRevision != doc.Revision {
		return txt.ErrConflict
	}
	pages, err := a.repo.ListPages(x, x.scope)
	if err != nil {
		return err
	}
	i := slices.IndexFunc(pages, func(p txt.Page) bool { return p.ID == r.PathValue("pageId") })
	if i < 0 {
		return txt.ErrNotFound
	}
	p := pages[i]
	if p.Locked {
		return txt.ErrLocked
	}
	drafts := make([]txt.PageDraft, 0, len(pages)+1)
	for _, item := range pages[:i] {
		drafts = append(drafts, item.PageDraft)
	}
	next := i + 1
	if strings.HasSuffix(r.URL.Path, "/split") {
		cut := in.Offset - p.StartOffset
		v, err := a.repo.GetVersion(x, x.scope, doc.ActiveVersionID)
		if err != nil {
			return err
		}
		if cut <= 0 || in.Offset >= p.EndOffset || !slices.Contains(txt.SentenceBreaks(v.ProductionText), in.Offset) {
			return fmt.Errorf("%w：请选择完整句子的末尾", txt.ErrInvalid)
		}
		chars := []rune(p.Text)
		left, right := p.PageDraft, p.PageDraft
		left.Text, left.EndOffset = string(chars[:cut]), in.Offset
		right.Text, right.StartOffset = string(chars[cut:]), in.Offset
		if strings.TrimSpace(left.Text) == "" || strings.TrimSpace(right.Text) == "" {
			return fmt.Errorf("%w：两页都需要保留正文", txt.ErrInvalid)
		}
		drafts = append(drafts, left, right)
	} else {
		if next >= len(pages) {
			return txt.ErrInvalid
		}
		if pages[next].Locked {
			return txt.ErrLocked
		}
		merged := p.PageDraft
		merged.Text += pages[next].Text
		merged.EndOffset = pages[next].EndOffset
		style, err := txt.PageStyleFromSnapshot(doc.ConfigSnapshot)
		if err != nil {
			return err
		}
		_, effective, err := txt.ResolvePageStyle(style, merged, txt.PagePatch{})
		if err != nil {
			return err
		}
		fits, _, err := (txt.Renderer{}).MeasureFunc(x)(merged.Text, effective.Layout)
		if err != nil {
			return textLayoutError(err)
		}
		if !fits {
			return fmt.Errorf("%w：合并后文字超出页面", txt.ErrInvalid)
		}
		drafts = append(drafts, merged)
		next++
	}
	for _, item := range pages[next:] {
		drafts = append(drafts, item.PageDraft)
	}
	if _, err = a.repo.ReplaceSystemPages(x, x.scope, txt.PaginationUpdate{ExpectedRevision: doc.Revision, VersionID: doc.ActiveVersionID, Pages: drafts}); err != nil {
		return err
	}
	return a.list(w, r, x)
}
