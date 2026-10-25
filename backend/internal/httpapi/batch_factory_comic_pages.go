package httpapi

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
	"qiantie/backend/internal/novelfetchworkshop"
)

func registerComicPageRoutes(mux *http.ServeMux, store batchfactoryv11.Store, files *localartifact.Store, localExecutors *localexecutor.Service, catalogStore novelfetchworkshop.Store) {
	comicStore, ok := store.(batchfactoryv11.ComicPageStore)
	if mux == nil || !ok {
		return
	}
	const pages = "/api/batch-factory/v12/batches/{batchId}/books/{bookId}/comic-pages"
	const page = pages + "/{videoId}"
	const images = pages + "/{pageId}/images"
	const image = images + "/{imageId}"
	const collections = "/api/batch-factory/v12/batches/{batchId}/books/{bookId}/comic-collections"

	mux.HandleFunc("GET "+pages, func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		result, err := comicStore.ListComicPages(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		batch, batchErr := store.GetBatch(r.Context(), owner, r.PathValue("batchId"))
		if batchErr != nil {
			writeStoreError(w, batchErr)
			return
		}
		for index := range result {
			result[index].ConfigStale = comicPageConfigStale(r.Context(), store, owner, batch, r.PathValue("bookId"), result[index])
		}
		writeJSON(w, http.StatusOK, map[string]any{"pages": comicPageResponses(result)})
	})
	mux.HandleFunc("PUT "+page, func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchfactoryv11.UpsertComicPageInput
		if !decodeJSON(w, r, &input) {
			return
		}
		created, err := comicStore.UpsertComicPage(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("videoId"), input)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"page": comicPageResponse(created)})
	})
	mux.HandleFunc("POST "+page+"/generate", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if files == nil || catalogStore == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "漫画图片生成服务尚未配置"})
			return
		}
		var input batchfactoryv11.UpsertComicPageInput
		if !decodeJSON(w, r, &input) {
			return
		}
		batch, err := store.GetBatch(r.Context(), owner, r.PathValue("batchId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		var selectedBook *batchfactoryv11.Book
		for index := range batch.Books {
			if batch.Books[index].ID == r.PathValue("bookId") {
				selectedBook = &batch.Books[index]
				break
			}
		}
		if selectedBook == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "小说不存在"})
			return
		}
		effective := batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, selectedBook.SettingsState.Patch)
		assetImageURLs, assetErr := comicPageAssetReferenceURLs(r.Context(), store, owner, batch.ID, selectedBook.ID, input.ImagePrompt)
		if assetErr != nil {
			writeStoreError(w, assetErr)
			return
		}
		contract, compileErr := batchfactoryv11.BuildComicImageContract(effective, batchfactoryv11.ComicImagePurposePage, input.ImagePrompt, assetImageURLs, input.ReferenceImageURLs)
		if compileErr != nil {
			writeStoreError(w, compileErr)
			return
		}
		input.ImagePrompt = contract.Prompt
		input.SnapshotHash = contract.SnapshotHash
		referenceImageURLs := contract.ReferenceImageURLs
		createdPage, err := comicStore.UpsertComicPage(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("videoId"), input)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		config, err := loadShuihuoOwnerConfig(r.Context(), store, owner, catalogStore)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		target, err := resolveShuihuoAssetImageTarget(batchfactoryv11.BookAsset{}, createdPage.ModelID, createdPage.AspectRatio, input.Provider, shuihuoCollection(config, "modelCatalog", nil))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		mediaType, storageRef, source, err := generateShuihuoImageArtifact(r.Context(), files, owner, target, createdPage.ImagePrompt, referenceImageURLs, localExecutors)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		createdImage, err := comicStore.CreateComicPageImage(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), createdPage.ID, batchfactoryv11.CreateComicPageImageInput{StorageRef: storageRef, MediaType: mediaType, Source: source})
		if err != nil {
			_ = files.Remove(storageRef)
			writeStoreError(w, err)
			return
		}
		createdImage.BatchID, createdImage.BookID = createdPage.BatchID, createdPage.BookID
		createdPage.Images = append(createdPage.Images, createdImage)
		writeJSON(w, http.StatusCreated, map[string]any{"page": comicPageResponse(createdPage), "image": comicPageImageResponse(createdImage)})
	})
	mux.HandleFunc("POST "+images, func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchfactoryv11.CreateComicPageImageInput
		if !decodeJSON(w, r, &input) {
			return
		}
		if strings.EqualFold(strings.TrimSpace(input.Source), "upload") {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "uploaded image bytes must use the upload endpoint"})
			return
		}
		created, err := comicStore.CreateComicPageImage(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("pageId"), input)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"image": comicPageImageResponse(created)})
	})
	mux.HandleFunc("POST "+images+"/upload", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if files == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "comic image storage is unavailable"})
			return
		}
		var input struct {
			DataURL string `json:"dataUrl"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		mediaType, payload, err := decodeImageDataURL(input.DataURL)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid image upload"})
			return
		}
		artifactID, err := localexecutor.NewArtifactID()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "image id generation failed"})
			return
		}
		saved, err := files.SaveImage(artifactID, mediaType, bytes.NewReader(payload))
		if err != nil {
			writeBookAssetImageStorageError(w, err)
			return
		}
		created, err := comicStore.CreateComicPageImage(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("pageId"), batchfactoryv11.CreateComicPageImageInput{StorageRef: saved.StorageRef, MediaType: saved.MediaType, Source: "upload"})
		if err != nil {
			_ = files.Remove(saved.StorageRef)
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"image": comicPageImageResponse(created)})
	})
	mux.HandleFunc("PUT "+image+"/primary", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		updated, err := comicStore.SetPrimaryComicPageImage(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("pageId"), r.PathValue("imageId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"image": comicPageImageResponse(updated)})
	})
	mux.HandleFunc("GET "+image+"/content", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if files == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "comic image storage is unavailable"})
			return
		}
		pages, err := comicStore.ListComicPages(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		for _, current := range pages {
			if current.ID != r.PathValue("pageId") {
				continue
			}
			for _, candidate := range current.Images {
				if candidate.ID != r.PathValue("imageId") {
					continue
				}
				if candidate.StorageRef == "" {
					writeJSON(w, http.StatusNotFound, map[string]string{"error": "image bytes are not stored locally"})
					return
				}
				file, openErr := files.Open(candidate.StorageRef)
				if openErr != nil {
					writeJSON(w, http.StatusNotFound, map[string]string{"error": "image file not found"})
					return
				}
				defer file.Close()
				w.Header().Set("Content-Type", candidate.MediaType)
				w.Header().Set("X-Content-Type-Options", "nosniff")
				http.ServeContent(w, r, candidate.ID, candidate.CreatedAt, file)
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "image not found"})
	})
	mux.HandleFunc("POST "+collections, func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		created, missing, err := comicStore.CreateComicCollection(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"))
		if errors.Is(err, batchfactoryv11.ErrConflict) {
			writeJSON(w, http.StatusConflict, map[string]any{"error": "每个分镜都必须先选好主图", "missingVideoIds": missing})
			return
		}
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"collection": comicCollectionResponse(created)})
	})
	mux.HandleFunc("GET "+collections, func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		result, err := comicStore.ListComicCollections(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		for index := range result {
			result[index] = comicCollectionResponse(result[index])
		}
		writeJSON(w, http.StatusOK, map[string]any{"collections": result})
	})
}

func comicPageAssetReferenceURLs(ctx context.Context, store batchfactoryv11.Store, owner, batchID, bookID, promptText string) ([]string, error) {
	assets, err := store.ListBookAssets(ctx, owner, batchID, bookID)
	if err != nil {
		return nil, err
	}
	promptText = strings.TrimSpace(promptText)
	references := make([]string, 0)
	for _, asset := range assets {
		if !comicPromptMentionsAsset(promptText, asset) {
			continue
		}
		images, listErr := store.ListBookAssetImages(ctx, owner, batchID, bookID, asset.ID)
		if listErr != nil {
			return nil, listErr
		}
		for _, image := range images {
			if !image.IsPrimary {
				continue
			}
			if remote := strings.TrimSpace(image.URL); remote != "" {
				parsed, parseErr := url.Parse(remote)
				if parseErr == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != "" {
					references = append(references, remote)
				}
			} else if strings.TrimSpace(image.StorageRef) != "" {
				references = append(references, "/api/local-executor/v1/batch-book-asset-images/"+url.PathEscape(batchID)+"/"+url.PathEscape(bookID)+"/"+url.PathEscape(asset.ID)+"/"+url.PathEscape(image.ID))
			}
			break
		}
	}
	return references, nil
}

func comicPromptMentionsAsset(promptText string, asset batchfactoryv11.BookAsset) bool {
	for _, name := range []string{asset.Name, asset.FormalName} {
		if name = strings.TrimSpace(name); name != "" && strings.Contains(promptText, name) {
			return true
		}
	}
	return false
}

func comicPageConfigStale(ctx context.Context, store batchfactoryv11.Store, owner string, batch batchfactoryv11.Batch, bookID string, page batchfactoryv11.ComicPage) bool {
	if strings.TrimSpace(page.SnapshotHash) == "" || len(page.Images) == 0 {
		return false
	}
	var book *batchfactoryv11.Book
	for index := range batch.Books {
		if batch.Books[index].ID == bookID {
			book = &batch.Books[index]
			break
		}
	}
	if book == nil {
		return false
	}
	effective := batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
	assets, err := comicPageAssetReferenceURLs(ctx, store, owner, batch.ID, bookID, page.ComicPrompt)
	if err != nil {
		return false
	}
	contract, err := batchfactoryv11.BuildComicImageContract(effective, batchfactoryv11.ComicImagePurposePage, page.ComicPrompt, assets, nil)
	return err != nil || contract.SnapshotHash != page.SnapshotHash
}

func comicPageImageResponse(image batchfactoryv11.ComicPageImage) batchfactoryv11.ComicPageImage {
	if image.StorageRef != "" {
		image.URL = "/api/batch-factory/v12/batches/" + image.BatchID + "/books/" + image.BookID + "/comic-pages/" + image.ComicPageID + "/images/" + image.ID + "/content"
	}
	return image
}

func comicPageResponse(page batchfactoryv11.ComicPage) batchfactoryv11.ComicPage {
	for index := range page.Images {
		page.Images[index].BatchID = page.BatchID
		page.Images[index].BookID = page.BookID
		page.Images[index] = comicPageImageResponse(page.Images[index])
	}
	return page
}

func comicPageResponses(pages []batchfactoryv11.ComicPage) []batchfactoryv11.ComicPage {
	for index := range pages {
		pages[index] = comicPageResponse(pages[index])
	}
	return pages
}

func comicCollectionResponse(collection batchfactoryv11.ComicCollection) batchfactoryv11.ComicCollection {
	for index := range collection.Items {
		item := &collection.Items[index]
		item.URL = "/api/batch-factory/v12/batches/" + collection.BatchID + "/books/" + collection.BookID + "/comic-pages/" + item.ComicPageID + "/images/" + item.ImageID + "/content"
	}
	return collection
}
