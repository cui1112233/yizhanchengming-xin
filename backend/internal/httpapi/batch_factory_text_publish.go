package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	txt "qiantie/backend/internal/batchfactorytext"
	"qiantie/backend/internal/batchfactoryv11"
	cp "qiantie/backend/internal/collectionpublish"
	"qiantie/backend/internal/mergeworker"
	"qiantie/backend/internal/n8comic"
)

const textPublishDraftKind = "n8-text-publish"

type collectionPublishSelection struct {
	Kind          cp.CollectionKind `json:"kind,omitempty"`
	Collections   []cp.Reference    `json:"collections,omitempty"`
	CollectionIDs []string          `json:"collectionIds,omitempty"`
	RequestKey    string            `json:"requestKey,omitempty"`
}

func (s collectionPublishSelection) check(k cp.CollectionKind, id string) error {
	if s.Kind != "" && s.Kind != k {
		return cp.ErrMixedKinds
	}
	if err := cp.ValidateSelection(k, id, s.Collections); err != nil {
		return err
	}
	for _, v := range s.CollectionIDs {
		if v != id {
			return fmt.Errorf("%w", cp.ErrMixedKinds)
		}
	}
	return nil
}
func (s collectionPublishSelection) validate(w http.ResponseWriter, k cp.CollectionKind, id string) bool {
	if err := s.check(k, id); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return false
	}
	return true
}

type collectionDraftRecords struct{ store batchfactoryv11.Store }

func publishDraftKind(k cp.CollectionKind) string {
	if k == cp.CollectionKindText {
		return textPublishDraftKind
	}
	return n8ComicPublishDraftKind
}

// Readback locations can contain a signed resource identifier. Only the durable
// store carries them; normal status responses and workflow summaries omit them.
type storedPublishRecord struct {
	cp.Record
	PrivateDispatchID  string   `json:"_dispatchId,omitempty"`
	PrivateOperationID string   `json:"_operationId,omitempty"`
	PrivateReceiptPath string   `json:"_receiptPath,omitempty"`
	PrivateSignatures  []string `json:"_imageSignatures,omitempty"`
}

func archivedPublishKey(id, key string) string {
	sum := sha256.Sum256([]byte(id + "\x00" + key))
	return hex.EncodeToString(sum[:])
}
func (d collectionDraftRecords) Load(ctx context.Context, r cp.PublishRequest) (cp.Record, bool, error) {
	var rec cp.Record
	if r.Force && r.RequestKey != "" {
		v, err := d.store.GetDraft(ctx, r.Owner, archivedPublishKey(r.CollectionID, r.RequestKey), publishDraftKind(r.Kind)+"-history", r.BatchID)
		if err == nil {
			if err = json.Unmarshal([]byte(v.Content), &rec); err != nil {
				return rec, false, err
			}
			if rec.BookID != r.BookID || rec.CollectionID != r.CollectionID || rec.Kind != r.Kind {
				return rec, false, batchfactoryv11.ErrNotFound
			}
			return rec, true, nil
		}
		if !errors.Is(err, batchfactoryv11.ErrNotFound) {
			return rec, false, err
		}
	}
	v, err := d.store.GetDraft(ctx, r.Owner, r.CollectionID, publishDraftKind(r.Kind), r.BatchID)
	if errors.Is(err, batchfactoryv11.ErrNotFound) {
		return rec, false, nil
	}
	if err != nil {
		return rec, false, err
	}
	var stored storedPublishRecord
	if err = json.Unmarshal([]byte(v.Content), &stored); err != nil {
		return rec, false, err
	}
	rec = stored.Record
	rec.ReceiptPath, rec.ImageSignatures = stored.PrivateReceiptPath, stored.PrivateSignatures
	rec.DispatchID, rec.OperationID = stored.PrivateDispatchID, stored.PrivateOperationID
	if rec.Kind != "" && rec.Kind != r.Kind || rec.BookID != "" && rec.BookID != r.BookID {
		return rec, false, batchfactoryv11.ErrNotFound
	}
	if rec.BookID == "" && r.Kind == cp.CollectionKindComic {
		// Older comic receipts only carried a collection ID. Verify its book before reuse.
		store, ok := d.store.(batchfactoryv11.ComicPageStore)
		if !ok {
			return rec, false, batchfactoryv11.ErrNotFound
		}
		list, err := store.ListComicCollections(ctx, r.Owner, r.BatchID, r.BookID)
		if err != nil {
			return rec, false, err
		}
		if !slices.ContainsFunc(list, func(c batchfactoryv11.ComicCollection) bool { return c.ID == r.CollectionID }) {
			return rec, false, batchfactoryv11.ErrNotFound
		}
	}
	return rec, true, nil
}
func (d collectionDraftRecords) Save(ctx context.Context, r cp.PublishRequest, rec cp.Record) error {
	var archive *batchfactoryv11.Draft
	if prev := rec.LastPublished; prev != nil && prev.RequestKey != "" && (rec.Status == "uploading" || rec.Status == "pending" || rec.RunID != "" && rec.IntentID == prev.IntentID) {
		raw, err := json.Marshal(prev)
		if err != nil {
			return err
		}
		archive = &batchfactoryv11.Draft{Kind: publishDraftKind(r.Kind) + "-history", Scope: r.BatchID, Key: archivedPublishKey(r.CollectionID, prev.RequestKey), Content: string(raw)}
	}
	raw, err := json.Marshal(storedPublishRecord{Record: rec, PrivateDispatchID: rec.DispatchID, PrivateOperationID: rec.OperationID, PrivateReceiptPath: rec.ReceiptPath, PrivateSignatures: rec.ImageSignatures})
	if err != nil {
		return err
	}
	draft := batchfactoryv11.Draft{Kind: publishDraftKind(r.Kind), Scope: r.BatchID, Key: r.CollectionID, Content: string(raw)}
	if r.Kind == cp.CollectionKindText {
		writer, ok := d.store.(interface {
			SaveTextPublishReceipt(context.Context, string, cp.Record, batchfactoryv11.Draft, *batchfactoryv11.Draft) error
		})
		if !ok {
			return batchfactoryv11.ErrUnavailable
		}
		return writer.SaveTextPublishReceipt(ctx, r.Owner, rec, draft, archive)
	}
	if archive != nil {
		if _, err = d.store.SaveDraft(ctx, r.Owner, *archive); err != nil {
			return err
		}
	}
	_, err = d.store.SaveDraft(ctx, r.Owner, draft)
	return err
}
func (p *BatchFactoryN8ComicPublisher) configureCollectionClaim(redisURL string) {
	if p == nil || redisURL == "" {
		return
	}
	redis, configErr := mergeworker.NewRedisStore(redisURL, "qiantie:collection-publish")
	p.claimCollection = func(ctx context.Context, r cp.PublishRequest) (func(), error) {
		if configErr != nil {
			return nil, configErr
		}
		key, _ := json.Marshal([]string{"collection-publish", string(r.Kind), r.BatchID, r.BookID, r.CollectionID})
		return subjectUploadLease(ctx, subjectUploadRedisDo(redis.Do), r.Owner, string(key))
	}
}
func (p *BatchFactoryN8ComicPublisher) collectionService() *cp.Service {
	svc := &cp.Service{Acquire: p.claimCollection, Records: collectionDraftRecords{p.Store}, Now: p.Now, Lock: &p.mu}
	if p.Client != nil {
		svc.Submit = p.Client.Publish
		if reader, ok := p.Client.(interface {
			Readback(context.Context, n8comic.ReceiptLookup) (n8comic.PublishResult, error)
		}); ok {
			svc.Readback = reader.Readback
		}
	}
	return svc
}
func (p *BatchFactoryN8ComicPublisher) collectionName(producer, title string) (string, error) {
	now := time.Now
	if p.Now != nil {
		now = p.Now
	}
	return n8ComicCollectionName(producer, title, now())
}
func (p *BatchFactoryN8ComicPublisher) publishPage(ctx context.Context, id string, ordinal int, ref, mediaType, digest string) (cp.Page, error) {
	var page cp.Page
	file, err := p.LocalFiles.OpenContext(ctx, ref)
	if err != nil {
		return page, fmt.Errorf("读取合集第 %d 页失败：%w", ordinal, err)
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, (64<<20)+1))
	if err != nil {
		return page, err
	}
	if len(content) == 0 || len(content) > 64<<20 {
		return page, fmt.Errorf("第 %d 页的图片为空或过大", ordinal)
	}
	if digest != "" {
		hash := sha256.Sum256(content)
		if !strings.EqualFold(digest, hex.EncodeToString(hash[:])) {
			return page, fmt.Errorf("第 %d 页与合成时的图片不一致，请重新合成", ordinal)
		}
	}
	return cp.Page{Ordinal: ordinal, Image: n8comic.ImageInput{FileName: fmt.Sprintf("%s-%02d%s", id, ordinal, comicImageExtension(mediaType)), MediaType: mediaType, Content: content}}, nil
}

type textPublishInput struct {
	runID   string
	attempt int
	collectionPublishSelection
	Force bool `json:"force"`
}
type textCollectionPublisher struct {
	repo txt.Repository
	n8   *BatchFactoryN8ComicPublisher
}

func textPublishRequest(s txt.Scope, id string) cp.PublishRequest {
	return cp.PublishRequest{Kind: cp.CollectionKindText, Owner: s.Owner, BatchID: s.BatchID, BookID: s.BookID, CollectionID: id}
}

func (t textCollectionPublisher) route(m *http.ServeMux, options RouterOptions) {
	const route = "/api/batch-factory/v12/batches/{batchId}/books/{bookId}/text-collections/{collectionId}/publish"
	a := textAPI{store: options.Store, repo: t.repo, files: options.LocalArtifacts}
	m.HandleFunc("GET "+route, a.scoped(func(w http.ResponseWriter, r *http.Request, x textRequest) error {
		id := r.PathValue("collectionId")
		if _, err := t.repo.GetCollection(x, x.scope, id); err != nil {
			return err
		}
		svc := &cp.Service{Records: collectionDraftRecords{a.store}}
		if t.n8 != nil && t.n8.Store != nil {
			svc = t.n8.collectionService()
		}
		rec, found, err := svc.Status(x, textPublishRequest(x.scope, id))
		if err != nil {
			return err
		}
		if !found {
			writeJSON(w, 200, map[string]any{"published": false, "publish": nil, "lastPublished": nil})
			return nil
		}
		writeTextPublishResult(w, rec)
		return nil
	}))
	m.HandleFunc("POST "+route, a.scoped(func(w http.ResponseWriter, r *http.Request, x textRequest) error {
		var in textPublishInput
		if !decodeJSON(w, r, &in) {
			return nil
		}
		id := r.PathValue("collectionId")
		if !in.validate(w, cp.CollectionKindText, id) {
			return nil
		}
		if _, err := t.repo.GetCollection(x, x.scope, id); err != nil {
			return err
		}
		rec, err := t.publish(x, x.scope, x.book, x.settings, id, in)
		if err != nil {
			switch {
			case errors.Is(err, cp.ErrMixedKinds), errors.Is(err, cp.ErrInvalid):
				writeJSON(w, 400, map[string]string{"error": err.Error()})
			case errors.Is(err, txt.ErrNotFound), errors.Is(err, txt.ErrNotReady), errors.Is(err, batchfactoryv11.ErrNotFound):
				return err
			default:
				writeJSON(w, 502, map[string]any{"error": err.Error(), "publish": rec})
			}
			return nil
		}
		writeTextPublishResult(w, rec)
		return nil
	}))
}

func writeTextPublishResult(w http.ResponseWriter, rec cp.Record) {
	history := rec.HistoricalPublished()
	writeJSON(w, 200, map[string]any{"published": rec.Confirmed(), "publish": rec, "lastPublished": history})
}

func (t textCollectionPublisher) automation(ctx context.Context, r txt.EnsureRequest, c txt.Collection) (bool, error) {
	if t.n8 == nil || t.n8.Store == nil {
		return false, fmt.Errorf("文本发布服务尚未配置")
	}
	b, err := t.n8.Store.GetBatch(ctx, r.Owner, r.BatchID)
	if err != nil {
		return false, err
	}
	book, ok := findBatchBookByID(b, r.BookID)
	if !ok {
		return false, batchfactoryv11.ErrNotFound
	}
	input := textPublishInput{runID: r.PublishRunID, attempt: r.PublishAttempt, Force: r.ForcePublish}
	input.RequestKey = r.PublishRequestKey
	rec, err := t.publish(ctx, r.Scope, book, r.FrozenSettings, c.ID, input)
	if err == nil && rec.Status == "failed" {
		return false, errors.New(rec.ErrorMessage)
	}
	return rec.Confirmed(), err
}

func (t textCollectionPublisher) publish(ctx context.Context, s txt.Scope, book batchfactoryv11.Book, settings batchfactoryv11.SettingsPatch, id string, in textPublishInput) (cp.Record, error) {
	if err := in.check(cp.CollectionKindText, id); err != nil {
		return cp.Record{}, err
	}
	p := t.n8
	if p == nil || p.Store == nil {
		return cp.Record{}, fmt.Errorf("文本发布服务尚未配置")
	}
	req := textPublishRequest(s, id)
	req.RunID, req.Attempt = in.runID, in.attempt
	req.Force = in.Force
	req.RequestKey = in.RequestKey
	req.Collections = in.Collections
	return p.collectionService().Publish(ctx, req, func(ctx context.Context, r *cp.PublishRequest) (cp.Collection, error) {
		var out cp.Collection
		if t.repo == nil || p.LocalFiles == nil || p.Client == nil || !p.Client.Configured() {
			return out, fmt.Errorf("文本发布服务尚未配置")
		}
		c, err := t.repo.GetCollection(ctx, s, id)
		if err != nil {
			return out, err
		}
		if c.ID != id || c.Status != "ready" || len(c.Items) == 0 || c.CoverArtifactID == "" {
			return out, txt.ErrNotReady
		}
		if !r.Resuming {
			if err = t.resolveMapping(ctx, r, book, settings); err != nil {
				return out, err
			}
		}
		// Hash only the selected routing values after resolving the configuration.
		data, err := json.Marshal([]any{r.Producer, r.OrganizationID, r.WebsiteProfileID, r.CategoryID, r.AudioCategoryID, r.AudioID, r.WorkIDs})
		if err != nil {
			return out, err
		}
		sum := sha256.Sum256(data)
		r.MappingRevision = hex.EncodeToString(sum[:])
		out.Kind = cp.CollectionKindText
		out.ID = c.ID
		out.Name, err = p.collectionName(r.Producer, book.Title)
		if err != nil {
			return out, err
		}
		files, err := t.repo.ListArtifacts(ctx, s)
		if err != nil {
			return out, err
		}
		index := slices.IndexFunc(files, func(a txt.PageArtifact) bool { return a.ID == c.CoverArtifactID && a.Kind == "cover" })
		if index < 0 {
			return out, fmt.Errorf("%w（封面）", txt.ErrNotReady)
		}
		cover := files[index]
		page, err := p.publishPage(ctx, c.ID, 0, cover.StorageRef, cover.MediaType, cover.SHA256)
		if err != nil {
			return out, err
		}
		out.Pages = append(out.Pages, page)
		for _, item := range c.Items {
			page, err = p.publishPage(ctx, c.ID, item.Ordinal, item.StorageRef, item.MediaType, item.SHA256)
			if err != nil {
				return out, err
			}
			out.Pages = append(out.Pages, page)
		}
		return out, nil
	})
}

// The image platform associates work and audio IDs with each carousel. Its
// organization is resolved by the producer and the organization's audio directory.
func (t textCollectionPublisher) resolveMapping(ctx context.Context, r *cp.PublishRequest, book batchfactoryv11.Book, settings batchfactoryv11.SettingsPatch) error {
	config, err := parseN8ComicPublishSettings(settings)
	if err != nil {
		return fmt.Errorf("%w：请先配置文本上传人和组织", cp.ErrInvalid)
	}
	var website map[string]any
	if raw := settings["publishSettings"]; len(raw) > 0 {
		if err = json.Unmarshal(raw, &website); err != nil {
			return cp.ErrInvalid
		}
	}
	r.Producer = strings.TrimSpace(config.ProducerAdminName)
	r.OrganizationID = strings.TrimSpace(config.OrganizationID)
	r.CategoryID = strings.TrimSpace(toString(website["bookCategory"]))
	r.WebsiteProfileID = strings.TrimSpace(toString(website["websiteProfileId"]))
	if r.Producer == "" || r.OrganizationID == "" || r.CategoryID == "" || r.WebsiteProfileID == "" {
		return fmt.Errorf("%w：请在文本发布配置中填写上传人、组织、网站配置档和书籍分类", cp.ErrInvalid)
	}
	if profile := selectedPublishProfile(website); profile == nil {
		return fmt.Errorf("%w：所选网站配置档不存在，请重新同步并选择", cp.ErrInvalid)
	}
	p := t.n8
	names, _, warning, orgs := p.producerDirectorySnapshot(ctx)
	if warning != "" && len(names) == 0 {
		return errors.New(warning)
	}
	if len(names) > 0 && !slices.Contains(names, r.Producer) {
		return fmt.Errorf("%w：上传人不在可用账号中", cp.ErrInvalid)
	}
	if orgs[r.Producer] != nil && !slices.Contains(orgs[r.Producer], r.OrganizationID) {
		return fmt.Errorf("%w：上传人不属于所选组织", cp.ErrInvalid)
	}
	r.WorkIDs = config.WorkIDs
	for _, token := range strings.FieldsFunc(config.WorkIDsText, func(c rune) bool { return c == ',' || c == '，' || c == ' ' || c == '\t' || c == '\n' }) {
		id, e := strconv.ParseInt(token, 10, 64)
		if e != nil || id <= 0 {
			return fmt.Errorf("%w：关联作品 ID 无效", cp.ErrInvalid)
		}
		if !slices.Contains(r.WorkIDs, id) {
			r.WorkIDs = append(r.WorkIDs, id)
		}
	}
	if len(r.WorkIDs) == 0 {
		if !comicPublishUsesTitleMatch(config.WorkMatchMode) {
			return fmt.Errorf("%w：请先确认关联作品", cp.ErrInvalid)
		}
		r.WorkIDs, err = p.resolveComicWorkIDs(ctx, r.Owner, r.BatchID, r.BookID)
		if err != nil {
			return err
		}
	}
	// Text books never persist a comic audio selection. A failed upload reuses
	// the sorted first available track when the directory has not changed.
	categories, warning := p.audioCategoriesWithWarning(ctx)
	if warning != "" {
		return errors.New(warning)
	}
	audioID := config.AudioID
	if audioID == 0 && config.AudioCategorySelectionMode == "manual" {
		audioID = config.SpecifiedAudioID
	}
	categoryName := config.AudioCategoryName
	categoryID := config.AudioCategoryID
	if audioID <= 0 {
		categoryName, err = comicAudioCategoryName(book)
		if err != nil {
			return err
		}
		categoryID = 0
	}
	category, err := enabledComicAudioCategory(categories, categoryID, categoryName)
	if err != nil {
		return err
	}
	tracks, err := p.audios(ctx, strconv.FormatInt(category.ID, 10), r.OrganizationID)
	if err != nil {
		return err
	}
	if len(tracks) == 0 && audioID <= 0 {
		category, tracks, err = p.pickComicAudioFallback(ctx, r.OrganizationID, category, categories)
		if err != nil {
			return err
		}
	}
	if len(tracks) == 0 {
		return fmt.Errorf("%w：当前组织的分类没有启用的音频", cp.ErrInvalid)
	}
	if audioID <= 0 {
		audioID = tracks[0].ID
	}
	if _, err = enabledComicAudio(tracks, audioID); err != nil {
		return err
	}
	r.AudioCategoryID = category.ID
	r.AudioID = audioID
	return cp.ValidateMapping(*r)
}
