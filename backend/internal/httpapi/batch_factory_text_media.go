package httpapi

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	txt "qiantie/backend/internal/batchfactorytext"
	"qiantie/backend/internal/localartifact"
)

const textPictureLimit = 16 << 20

type textMediaView struct {
	ID            string `json:"id"`
	PageID        string `json:"pageId,omitempty"`
	Name          string `json:"name,omitempty"`
	Kind          string `json:"kind"`
	SourceKind    string `json:"sourceKind,omitempty"`
	RenderVersion int64  `json:"renderVersion,omitempty"`
	MediaType     string `json:"mediaType"`
	ByteSize      int64  `json:"byteSize"`
	SHA256        string `json:"sha256"`
	ContentURL    string `json:"contentUrl"`
}

func artifactView(x textRequest, a txt.PageArtifact) textMediaView {
	return textMediaView{ID: a.ID, PageID: a.PageID, Kind: a.Kind, SourceKind: a.SourceKind, RenderVersion: a.RenderVersion, MediaType: a.MediaType, ByteSize: a.ByteSize, SHA256: a.SHA256, ContentURL: x.base() + "artifacts/" + url.PathEscape(a.ID) + "/content"}
}
func backgroundView(x textRequest, b txt.LibraryBackground) textMediaView {
	return textMediaView{ID: b.ID, Name: b.Name, Kind: "background", MediaType: b.MediaType, ByteSize: b.ByteSize, SHA256: b.SHA256, ContentURL: x.base() + "backgrounds/" + url.PathEscape(b.ID) + "/content"}
}
func (a textAPI) bgData(x textRequest, b txt.Background) ([]byte, error) {
	switch b.Kind {
	case "", "solid", "color":
		if b.ID != "" {
			return nil, txt.ErrInvalid
		}
		if b.Color != "" {
			if len(b.Color) != 7 || b.Color[0] != '#' {
				return nil, txt.ErrInvalid
			}
			if _, err := strconv.ParseUint(b.Color[1:], 16, 24); err != nil {
				return nil, txt.ErrInvalid
			}
		}
		return nil, nil
	case "upload", "library", "image":
		bg, err := a.repo.GetBackground(x, x.scope.Owner, b.ID)
		if err != nil {
			return nil, err
		}
		file, err := a.files.OpenContext(x, bg.StorageRef)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		return readTextImage(file)
	default:
		return nil, fmt.Errorf("%w：未知的背景类型", txt.ErrInvalid)
	}
}
func readTextImage(src io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(src, textPictureLimit+1))
	if err != nil {
		return nil, err
	}
	if len(data) > textPictureLimit {
		return nil, localartifact.ErrTooLarge
	}
	return data, nil
}
func validateTextImage(data []byte) (string, error) {
	cfg, f, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || cfg.Height <= 0 || cfg.Width <= 0 || int64(cfg.Width) > 32_000_000/int64(cfg.Height) {
		return "", fmt.Errorf("%w：图片损坏或尺寸过大", txt.ErrInvalid)
	}
	switch f {
	case "png", "jpeg", "webp":
	default:
		return "", fmt.Errorf("%w：请选择 PNG、JPEG 或 WebP 图片", txt.ErrInvalid)
	}
	_, _, err = image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("%w：图片数据不完整", txt.ErrInvalid)
	}
	return "image/" + f, nil
}
func (a textAPI) renderInput(x textRequest, d txt.Document, p txt.Page, cover bool, patch txt.PagePatch) (txt.RenderInput, txt.Renderer, error) {
	var snap textSnapshot
	if err := json.Unmarshal(d.ConfigSnapshot, &snap); err != nil {
		return txt.RenderInput{}, txt.Renderer{}, err
	}
	prior := p.PageDraft
	content := p.Text
	if cover {
		prior = txt.PageDraft{Layout: snap.Cover.Layout, Background: snap.Cover.Background}
		content = snap.Title
		if content == "" {
			content = x.book.Title
		}
	}
	_, style, err := txt.ResolvePageStyle(snap.PageStyle, prior, patch)
	if err != nil {
		return txt.RenderInput{}, txt.Renderer{}, err
	}
	bg, err := a.bgData(x, style.Background)
	if err != nil {
		return txt.RenderInput{}, txt.Renderer{}, err
	}
	in := txt.RenderInput{Owner: x.scope.Owner, BookID: x.scope.BookID, VersionID: d.ActiveVersionID, Text: content, Layout: style.Layout, Background: style.Background, BackgroundImage: bg, HighlightReuseEnabled: style.HighlightReuseEnabled && !cover}
	renderer := txt.Renderer{Highlights: txt.KeywordHighlightSource(func(context.Context, string, string, string) ([]string, error) { return snap.HighlightWords, nil })}
	return in, renderer, nil
}
func (a textAPI) picture(w http.ResponseWriter, r *http.Request, x textRequest) error {
	var in textInput
	if !decodeJSON(w, r, &in) {
		return nil
	}
	d, err := a.repo.GetDocument(x, x.scope)
	if err != nil {
		return err
	}
	cover := r.PathValue("pageId") == ""
	preview := strings.HasSuffix(r.URL.Path, "/preview")
	var p txt.Page
	revision := d.Revision
	locked := d.CoverLocked
	if !cover {
		p, err = a.page(x, r.PathValue("pageId"))
		if err != nil {
			return err
		}
		if p.VersionID != d.ActiveVersionID {
			return txt.ErrConflict
		}
		revision = p.Revision
		locked = p.Locked
	}
	if preview && ((in.ExpectedRevision != 0 && in.ExpectedRevision != revision) || (in.ExpectedDocumentRevision != 0 && in.ExpectedDocumentRevision != d.Revision)) {
		return txt.ErrConflict
	}
	if in.Text != nil {
		if locked {
			return txt.ErrLocked
		}
		if !preview || cover || strings.TrimSpace(*in.Text) == "" {
			return fmt.Errorf("%w：草稿正文只可用于单页预览", txt.ErrInvalid)
		}
		p.Text = txt.NormalizeLineEndings(*in.Text)
	}
	if !preview {
		if revision != in.ExpectedRevision {
			return txt.ErrConflict
		}
		if locked {
			return txt.ErrLocked
		}
		if in.PagePatch != (txt.PagePatch{}) {
			return fmt.Errorf("%w：请先保存排版调整再生成图片", txt.ErrInvalid)
		}
	}
	input, renderer, err := a.renderInput(x, d, p, cover, in.PagePatch)
	if err != nil {
		return err
	}
	data, mime, err := renderer.Render(x, input)
	if err != nil {
		return textLayoutError(err)
	}
	if preview {
		w.Header().Set("Content-Type", mime)
		w.Header().Set("Cache-Control", "private, no-store")
		_, err = w.Write(data)
		return err
	}
	view, err := a.saveImage(x, p.ID, in.ExpectedRevision, false, "", data, mime)
	if err != nil {
		return err
	}
	writeJSON(w, 201, view)
	return nil
}
func (a textAPI) saveImage(x textRequest, pageID string, rev int64, upload bool, backgroundName string, data []byte, mime string) (view textMediaView, err error) {
	f, err := a.files.SaveImage("text_"+rand.Text(), mime, bytes.NewReader(data))
	if err != nil {
		return view, err
	}
	artifact := txt.ArtifactFile{StorageRef: f.StorageRef, MediaType: f.MediaType, ByteSize: f.ByteSize, SHA256: f.SHA256}
	// Only remove bytes on a definite rejection. An ambiguous DB commit must keep the bytes available.
	defer func() {
		if errors.Is(err, txt.ErrConflict) || errors.Is(err, txt.ErrLocked) || errors.Is(err, txt.ErrNotFound) || errors.Is(err, txt.ErrInvalid) {
			_ = a.files.Remove(f.StorageRef)
		}
	}()
	if backgroundName != "" {
		b, err := a.repo.SaveBackground(x, x.scope.Owner, backgroundName, artifact)
		return backgroundView(x, b), err
	}
	var saved txt.PageArtifact
	switch {
	case pageID == "":
		saved, err = a.repo.SaveCover(x, x.scope, rev, artifact, upload)
	case upload:
		saved, err = a.repo.LockPageArtifact(x, x.scope, pageID, rev, artifact)
	default:
		saved, err = a.repo.SaveSystemPageArtifact(x, x.scope, pageID, rev, artifact)
	}
	return artifactView(x, saved), err
}
func (a textAPI) upload(w http.ResponseWriter, r *http.Request, x textRequest) error {
	var name string
	var rev int64
	var err error
	if strings.HasSuffix(r.URL.Path, "/text-backgrounds") {
		name = strings.TrimSpace(r.URL.Query().Get("name"))
		if name == "" {
			name = "自定义背景"
		}
	} else {
		rev, err = strconv.ParseInt(r.URL.Query().Get("expectedRevision"), 10, 64)
		if err != nil || rev < 1 {
			return txt.ErrInvalid
		}
	}
	data, err := readTextImage(r.Body)
	if err != nil {
		return err
	}
	mime, err := validateTextImage(data)
	if err != nil {
		return err
	}
	view, err := a.saveImage(x, r.PathValue("pageId"), rev, true, name, data, mime)
	if err != nil {
		return err
	}
	writeJSON(w, 201, view)
	return nil
}
func (a textAPI) restore(w http.ResponseWriter, r *http.Request, x textRequest) error {
	var in textInput
	if !decodeJSON(w, r, &in) {
		return nil
	}
	var err error
	if id := r.PathValue("pageId"); id != "" {
		_, err = a.repo.RestoreSystemPage(x, x.scope, id, in.ExpectedRevision)
	} else {
		err = a.repo.RestoreCover(x, x.scope, in.ExpectedRevision)
	}
	if err != nil {
		return err
	}
	return a.list(w, r, x)
}
func (a textAPI) backgrounds(w http.ResponseWriter, _ *http.Request, x textRequest) error {
	bs, err := a.repo.ListBackgrounds(x, x.scope.Owner)
	if err != nil {
		return err
	}
	views := make([]textMediaView, 0, len(bs))
	for _, b := range bs {
		views = append(views, backgroundView(x, b))
	}
	writeJSON(w, 200, map[string]any{"items": views})
	return nil
}
func (a textAPI) content(w http.ResponseWriter, r *http.Request, x textRequest) error {
	var ref, mime string
	if id := r.PathValue("backgroundId"); id != "" {
		bg, err := a.repo.GetBackground(x, x.scope.Owner, id)
		if err != nil {
			return err
		}
		ref, mime = bg.StorageRef, bg.MediaType
	} else {
		artifacts, err := a.repo.ListArtifacts(x, x.scope)
		if err != nil {
			return err
		}
		for _, art := range artifacts {
			if art.ID == r.PathValue("artifactId") {
				ref, mime = art.StorageRef, art.MediaType
				break
			}
		}
	}
	if ref == "" {
		return txt.ErrNotFound
	}
	f, err := a.files.OpenContext(x, ref)
	if err != nil {
		return err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Content-Type", mime)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, "image", stat.ModTime(), f)
	return nil
}

// Check the immutable files before freezing their IDs. The document revision also
// changes on page/cover replacement, so the final transaction cannot select other files.
func (a textAPI) verifyCollectionFiles(x textRequest, in textInput) (txt.Document, error) {
	d, err := a.repo.GetDocument(x, x.scope)
	if err != nil {
		return d, err
	}
	if d.ActiveVersionID != in.VersionID || d.PaginationVersion != in.PaginationVersion || in.ExpectedRevision > 0 && in.ExpectedRevision != d.Revision {
		return d, txt.ErrConflict
	}
	if d.CoverArtifactID == "" {
		return d, fmt.Errorf("%w（封面）", txt.ErrNotReady)
	}
	artifacts, err := a.repo.ListArtifacts(x, x.scope)
	if err != nil {
		return d, err
	}
	byID := make(map[string]txt.PageArtifact, len(artifacts))
	for _, art := range artifacts {
		byID[art.ID] = art
	}
	verify := func(id, label string) error {
		art, ok := byID[id]
		if !ok {
			return fmt.Errorf("%w（%s）", txt.ErrNotReady, label)
		}
		f, err := a.files.OpenContext(x, art.StorageRef)
		if err != nil {
			return fmt.Errorf("%w（%s图片不可读取）", txt.ErrNotReady, label)
		}
		defer f.Close()
		h := sha256.New()
		n, err := io.Copy(h, io.LimitReader(f, art.ByteSize+1))
		if err != nil || n != art.ByteSize || hex.EncodeToString(h.Sum(nil)) != art.SHA256 {
			return fmt.Errorf("%w（%s图片已损坏）", txt.ErrNotReady, label)
		}
		return x.Err()
	}
	if err := verify(d.CoverArtifactID, "封面"); err != nil {
		return d, err
	}
	pages, err := a.repo.ListPages(x, x.scope)
	if err != nil {
		return d, err
	}
	selected := make(map[string]bool, len(in.PageIDs))
	for _, id := range in.PageIDs {
		selected[id] = true
	}
	for _, p := range pages {
		if len(selected) == 0 || selected[p.ID] {
			if err := verify(p.ActiveArtifactID, fmt.Sprintf("第 %d 页", p.Ordinal)); err != nil {
				return d, err
			}
		}
	}
	return d, nil
}
func collectionView(x textRequest, c txt.Collection) map[string]any {
	urlFor := func(id string) string { return x.base() + "artifacts/" + url.PathEscape(id) + "/content" }
	items := make([]map[string]any, 0, len(c.Items))
	for _, i := range c.Items {
		items = append(items, map[string]any{"ordinal": i.Ordinal, "pageId": i.PageID, "artifactId": i.ArtifactID, "mediaType": i.MediaType, "sha256": i.SHA256, "contentUrl": urlFor(i.ArtifactID)})
	}
	return map[string]any{"id": c.ID, "versionId": c.VersionID, "paginationVersion": c.PaginationVersion, "coverArtifactId": c.CoverArtifactID, "coverContentUrl": urlFor(c.CoverArtifactID), "status": c.Status, "items": items}
}
func (a textAPI) collections(w http.ResponseWriter, r *http.Request, x textRequest) error {
	if r.Method == http.MethodPost {
		var in textInput
		if !decodeJSON(w, r, &in) {
			return nil
		}
		d, err := a.verifyCollectionFiles(x, in)
		if err != nil {
			return err
		}
		c, err := a.repo.CreateCollection(x, x.scope, txt.CollectionInput{ExpectedRevision: d.Revision, VersionID: in.VersionID, PaginationVersion: in.PaginationVersion, PageIDs: in.PageIDs, ConfigSnapshot: d.ConfigSnapshot})
		if err != nil {
			return err
		}
		writeJSON(w, 201, collectionView(x, c))
		return nil
	}
	if id := r.PathValue("collectionId"); id != "" {
		c, err := a.repo.GetCollection(x, x.scope, id)
		if err != nil {
			return err
		}
		writeJSON(w, 200, collectionView(x, c))
		return nil
	}
	cs, err := a.repo.ListCollections(x, x.scope)
	if err != nil {
		return err
	}
	views := make([]map[string]any, 0, len(cs))
	for _, c := range cs {
		views = append(views, collectionView(x, c))
	}
	writeJSON(w, 200, map[string]any{"items": views})
	return nil
}
