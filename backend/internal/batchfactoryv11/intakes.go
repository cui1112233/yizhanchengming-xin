package batchfactoryv11

import (
	"fmt"
	"strings"
)

// AcceptExternalOriginals recomputes provenance on a newly received handoff.
// Internal intake-to-batch reads use the provenance frozen at this boundary.
func AcceptExternalOriginals(books []CreateBookInput, origin string) {
	for index := range books {
		book := &books[index]
		meta := sourceWriteMetadata(nil, book.SourceMetadata, "", "")
		if _, ok := normalizeCapturedSourceText(book.OriginalRaw); ok {
			book.TxtText = book.OriginalRaw
			meta = sourceWriteMetadata(meta, nil, book.OriginalRaw, origin)
		}
		book.SourceMetadata = meta
		book.OriginalRaw = ""
	}
}

// isExcerptSource 判断一本书的正文是不是视频截取/OCR 稿而不是完整原文。
// 这类正文允许存在（先登记、后回填），但不能当作「已核实完整原文」。
func isExcerptSource(metadata map[string]any) bool {
	if metadata == nil {
		return false
	}
	if strings.TrimSpace(fmt.Sprint(metadata["sourceMode"])) == "giant_material" {
		return true
	}
	if value, ok := metadata["sourceCompleteness"]; ok {
		switch strings.TrimSpace(fmt.Sprint(value)) {
		case "video_excerpt", "excerpt", "partial":
			return true
		}
	}
	return false
}

// normalizeNovelFetchIntake keeps the first occurrence of each explicit source
// Book ID. Empty IDs are retained because they do not provide a stable
// deduplication identity.
func normalizeNovelFetchBook(book CreateBookInput) CreateBookInput {
	book.ID = strings.TrimSpace(book.ID)
	book.BookID = strings.TrimSpace(book.BookID)
	if book.BookID == "" {
		book.BookID = book.ID
	}
	if book.ID == "" {
		book.ID = book.BookID
	}
	book.SourceTaskID = strings.TrimSpace(book.SourceTaskID)
	book.Platform = strings.TrimSpace(book.Platform)
	book.TxtFileName = strings.TrimSpace(book.TxtFileName)
	if book.SourceMetadata["sourceMode"] == "novel_fetch" {
		if _, valid := normalizeCapturedSourceText(book.OriginalRaw); valid {
			book.TxtText = book.OriginalRaw
			book.SourceMetadata = sourceWriteMetadata(book.SourceMetadata, nil, book.OriginalRaw, "novel_fetch_original_raw")
		} else {
			book.SourceMetadata = sourceReadMetadata(book.SourceMetadata)
		}
	}
	// 部分调用方（手动导入、自定义文案）直接把完整原文放在 txtText 里，而
	// 不是 originalRaw。txtText 本身就是原始正文，与 originalRaw 属于同一类
	// 证据；这类录入过去没有写 rawSourceStatus，自动化预检就一直判定
	// 「缺少完整原文」并永久阻塞（2026-11-24 jiansen 批次事故）。这里在录入
	// 边界补记来源。视频截取稿（giant_material / video_excerpt）不是完整原文，
	// 必须排除，保持待校对状态。
	if strings.TrimSpace(book.TxtText) != "" && book.SourceMetadata["rawSourceStatus"] == nil && !isExcerptSource(book.SourceMetadata) {
		book.SourceMetadata = sourceWriteMetadata(book.SourceMetadata, nil, book.TxtText, "manual_txt")
	}
	book.OriginalRaw = ""
	if book.TxtFileName == "" && book.BookID != "" {
		book.TxtFileName = book.BookID + ".txt"
	}
	return book
}

func sourceBookID(book CreateBookInput) string {
	book = normalizeNovelFetchBook(book)
	return book.BookID
}

// batchBookKey is the stable identity of a source book inside one batch.
// Book IDs are scoped by book city: a collision across platforms is not a
// duplicate, but the same platform and Book ID must never create a second
// production row in the same batch.
func batchBookKey(book CreateBookInput) string {
	book = normalizeNovelFetchBook(book)
	if book.BookID == "" {
		return ""
	}
	return book.Platform + "\x00" + book.BookID
}

func uniqueBatchBooks(books []CreateBookInput, existing map[string]bool) []CreateBookInput {
	out := make([]CreateBookInput, 0, len(books))
	for _, raw := range books {
		book := normalizeNovelFetchBook(raw)
		key := batchBookKey(book)
		if key != "" {
			if existing[key] {
				continue
			}
			existing[key] = true
		}
		out = append(out, book)
	}
	return out
}

func sourceContentVersion(book CreateBookInput) string {
	if book.SourceMetadata == nil {
		return ""
	}
	value, ok := book.SourceMetadata["sourceContentVersion"]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func normalizeNovelFetchIntake(input NovelFetchIntakeInput) NovelFetchIntakeInput {
	out := NovelFetchIntakeInput{Metadata: input.Metadata, Books: make([]CreateBookInput, 0, len(input.Books))}
	seen := map[string]struct{}{}
	for _, book := range input.Books {
		book = normalizeNovelFetchBook(book)
		sourceID := book.BookID
		if sourceID != "" {
			// 去重键纳入书城：不同书城可能复用同一 Book ID，那是两本书；
			// 同书城同 ID（含 content version 语义）仍只保留第一行。
			key := strings.TrimSpace(book.Platform) + "\x00" + sourceID
			if version := sourceContentVersion(book); version != "" {
				key += "\x00" + version
			}
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
		}
		out.Books = append(out.Books, book)
	}
	return out
}
