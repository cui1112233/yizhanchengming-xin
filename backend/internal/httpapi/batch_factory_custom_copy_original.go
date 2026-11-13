package httpapi

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
)

// A selected custom copy is already authored text. Keep its saved original,
// including on retry/force-fetch; never send it through novel-fetch cleanup.
func isBatchCustomCopy(book batchfactoryv11.Book) bool {
	return book.SourceMetadata["sourceMode"] == "custom_copy" || firstMetadataString(book.SourceMetadata, "customCopyRecordId") != ""
}

func reuseCustomCopyOriginal(ctx context.Context, store batchfactoryv11.Store, owner, batchID string, book batchfactoryv11.Book) (batchfactoryv11.Book, error) {
	for attempt := 0; attempt < 3; attempt++ {
		if !isBatchCustomCopy(book) {
			return book, fmt.Errorf("文案来源已变化，请刷新后重试")
		}
		updated, err := captureCustomCopyOriginal(ctx, store, owner, batchID, book)
		if !errors.Is(err, batchfactoryv11.ErrConflict) {
			return updated, err
		}
		batch, err := store.GetBatch(ctx, owner, batchID)
		if err != nil {
			return book, err
		}
		found := false
		for _, latest := range batch.Books {
			if latest.ID == book.ID {
				book = latest
				found = true
				break
			}
		}
		if !found {
			return book, batchfactoryv11.ErrNotFound
		}
	}
	return book, fmt.Errorf("文案同时发生更新，请刷新后重试：%w", batchfactoryv11.ErrConflict)
}

func captureCustomCopyOriginal(ctx context.Context, store batchfactoryv11.Store, owner, batchID string, book batchfactoryv11.Book) (batchfactoryv11.Book, error) {
	source := book.TxtText
	if strings.TrimSpace(source) == "" {
		source = book.SourceText
	}
	if strings.TrimSpace(source) == "" {
		return book, fmt.Errorf("所选文案正文缺失，请重新通过“文案获取”选择文案")
	}
	if book.SourceText == source && book.SourceMetadata["sourceMode"] == "custom_copy" && book.SourceMetadata["sourceContentVersion"] == "original" && book.SourceMetadata["processingStatus"] == "done" && book.SourceMetadata["sensitiveStatus"] == "skipped" && book.SourceMetadata["sensitiveReviewRequired"] != true {
		return book, nil
	}
	meta := copyAnyMap(book.SourceMetadata)
	meta["sourceMode"] = "custom_copy"
	meta["sourceContentVersion"] = "original"
	meta["originalStatus"] = "done"
	meta["originalRawStatus"] = "done"
	meta["queueStage"] = "classification"
	meta["processedChars"] = len([]rune(source))
	meta["processingStatus"] = "done"
	meta["processingVersion"] = "custom-copy-original-v1"
	meta["processingStages"] = []map[string]any{}
	meta["processError"] = ""
	meta["processRetryable"] = false
	meta["sensitiveStatus"] = "skipped"
	meta["sensitiveSkipReason"] = "custom_copy"
	meta["sensitiveAIEnabled"] = false
	meta["sensitiveAIModelsApplied"] = []string{}
	meta["sensitiveReviewRequired"] = false
	meta["sensitiveReviewCount"] = 0
	meta["sensitiveReviewMessage"] = ""
	meta["sensitiveReplaced"] = 0
	meta["sensitiveHitCount"] = 0
	meta["sensitiveHits"] = nil
	meta["sensitiveFixed"] = nil
	return store.CaptureBookSource(ctx, owner, batchID, book.ID, batchfactoryv11.CaptureBookSourceInput{SourceText: source, RawSourceText: source, ReplaceSource: true, SourceMetadata: meta, ExpectedRevision: book.Revision})
}
