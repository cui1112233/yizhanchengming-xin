package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactorysensitive"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/prompt"
)

const missingSensitiveOriginal = "缺少完整原文，请重新获取原文或上传完整 TXT"

func (c *BatchAutomationCoordinator) classifyProductionBook(ctx context.Context, owner, batchID string, book batchfactoryv11.Book, mode string) (batchfactoryv11.Book, error) {
	metadata := book.SourceMetadata
	info := novelfetchworkshop.BookInfo{Category: firstMetadataString(metadata, "category"), Genre: firstMetadataString(metadata, "genre"), Themes: firstMetadataString(metadata, "themes"), Labels: firstMetadataString(metadata, "labels"), WorkTitle: book.Title}
	classification := c.classification
	if mode == "text" {
		classification = batchFetchClassification{}
	}
	metadata = finalizeFetchedBatchClassification(classifyFetchedBatchBook(ctx, owner, info, book.SourceText, metadata, classification))
	metadata["classificationAfterSensitive"] = false
	updated, err := c.store.UpdateBookMetadata(ctx, owner, batchID, book.ID, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata, ExpectedRevision: book.Revision})
	if err != nil {
		return book, err
	}
	return c.refreshAutomationBook(ctx, owner, batchID, updated), nil
}

func sensitiveDigest(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

type sensitiveAdapters struct {
	settings map[string]any
	stages   *batchfactoryv11.BookStageService
	models   []string
}

func (a sensitiveAdapters) ProcessRules(_ context.Context, in batchfactorysensitive.Input) (string, []batchfactorysensitive.Change, error) {
	text, _, items := novelfetchworkshop.ApplySensitiveRulesForScopeWithDetails(in.Text, a.settings, "original")
	var trace []batchfactorysensitive.Change
	for _, hit := range items {
		trace = append(trace, batchfactorysensitive.Change{Keyword: stringValue(hit["keyword"]), Before: stringValue(hit["before"]), After: stringValue(hit["after"])})
	}
	return text, trace, nil
}

func (a sensitiveAdapters) ProcessAI(ctx context.Context, in batchfactorysensitive.Input) (string, []batchfactorysensitive.Change, error) {
	var last error
	for _, model := range a.models {
		if ctx.Err() != nil {
			return "", nil, ctx.Err()
		}
		var provider batchfactoryv11.DirectorProvider
		if a.stages != nil && a.stages.TextProviderResolver != nil {
			var err error
			provider, err = a.stages.TextProviderResolver(ctx, in.Owner, model)
			if err != nil {
				last = err
				continue
			}
		}
		if provider == nil {
			last = fmt.Errorf("文本模型 %s 未配置", model)
			continue
		}
		output, err := provider.Complete(ctx, batchfactoryv11.TextCompletionRequest{SystemPrompt: prompt.BatchFactorySensitiveEditsPrompt(), UserPrompt: in.Text, Temperature: 0.1, MaxTokens: 8192})
		if err == nil {
			text, trace, decodeErr := applyBatchSensitiveEdits(in.Text, output)
			if decodeErr == nil {
				return text, trace, nil
			}
			err = decodeErr
		}
		last = err
	}
	if last == nil {
		last = errors.New("未选择可用的文本模型")
	}
	return "", nil, fmt.Errorf("敏感内容 AI 优化失败：%w", last)
}

func applyBatchSensitiveEdits(source, output string) (string, []batchfactorysensitive.Change, error) {
	var doc struct {
		Changes *[]batchfactorysensitive.Change `json:"changes"`
	}
	if err := json.Unmarshal([]byte(output), &doc); err != nil || doc.Changes == nil {
		return "", nil, errors.New("模型返回的改写列表格式不正确")
	}
	var result strings.Builder
	rest := source
	for i, change := range *doc.Changes {
		at := strings.Index(rest, change.Before)
		if at < 0 || strings.TrimSpace(change.Before) == "" || strings.TrimSpace(change.After) == "" {
			return "", nil, fmt.Errorf("第 %d 处改写没有对应的原文或替换内容", i+1)
		}
		result.WriteString(rest[:at])
		result.WriteString(change.After)
		rest = rest[at+len(change.Before):]
	}
	result.WriteString(rest)
	return result.String(), *doc.Changes, nil
}

func batchSensitiveModels(settings batchfactoryv11.SettingsPatch) []string {
	selected := resolveBatchSensitiveConfig(settings).Models
	if len(selected) > 0 {
		return selected
	}
	if primary := comicSettingString(settings, "textModelId", ""); primary != "" {
		return []string{primary}
	}
	return nil
}

func sensitiveBaseText(ctx context.Context, s batchfactoryv11.Store, owner, batchID string, book batchfactoryv11.Book, state map[string]any, limit int) (string, string, error) {
	draft, err := s.GetDraft(ctx, owner, "working-front:"+book.ID, "working-front-content", batchID)
	if err != nil && !errors.Is(err, batchfactoryv11.ErrNotFound) {
		return "", "", err
	}
	if strings.TrimSpace(draft.Content) != "" {
		return draft.Content, batchfactoryv11.WorkingDraftVersion(draft.Content), nil
	}
	digest := sensitiveDigest(book.SourceText)
	kind := firstMetadataString(book.SourceMetadata, "sourceContentVersion")
	edited := book.SourceMetadata["productionTextEdited"] == true
	// Selected rewrites and explicit edits are authored inputs. Freeze that version
	// once so changing the policy later can start from the same wording.
	baseKey := firstMetadataString(state, "authoredBase")
	if baseKey != "" && !strings.HasPrefix(baseKey, "working:") && state["raw"] == sensitiveDigest(book.TxtText) && (digest == state["output"] || digest == state["previous"]) {
		d, err := s.GetDraft(ctx, owner, baseKey, "sensitive_source", book.ID)
		if err != nil {
			return "", "", err
		}
		if sensitiveDigest(d.Content) != baseKey {
			return "", "", errors.New("生产稿快照校验失败，请重新保存生产稿")
		}
		return d.Content, baseKey, nil
	}
	selected := book.SourceMetadata["rawSourceOrigin"] == "novel_fetch_original_raw" &&
		(kind == "sensitive" || strings.HasPrefix(kind, "ai"))
	if edited || selected {
		_, err := s.SaveDraft(ctx, owner, batchfactoryv11.Draft{Key: digest, Kind: "sensitive_source", Scope: book.ID, Content: book.SourceText})
		return book.SourceText, digest, err
	}
	if isBatchCustomCopy(book) {
		return book.TxtText, "", nil
	}
	return novelfetchworkshop.ProcessRawText(book.TxtText, limit).ProcessedText, "", nil
}

func (c *BatchAutomationCoordinator) runBatchSensitiveStage(ctx context.Context, owner, batchID, bookID string, manualRetry bool) (batchfactoryv11.Book, error) {
	return c.batchSensitiveStage(ctx, owner, batchID, bookID, manualRetry, false)
}

func (c *BatchAutomationCoordinator) batchSensitiveStage(ctx context.Context, owner, batchID, bookID string, manualRetry, verifyOnly bool) (batchfactoryv11.Book, error) {
	batch, err := c.store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return batchfactoryv11.Book{}, err
	}
	book, found := findBatchBookByID(batch, bookID)
	if !found {
		return book, batchfactoryv11.ErrNotFound
	}
	return c.batchSensitiveStageFromBook(ctx, owner, batch, book, manualRetry, verifyOnly)
}

// A caller that already read the book can pin that exact source and settings
// through preparation. Source writes in this stage check the book revision.
func (c *BatchAutomationCoordinator) batchSensitiveStageFromBook(ctx context.Context, owner string, batch batchfactoryv11.Batch, book batchfactoryv11.Book, manualRetry, verifyOnly bool) (batchfactoryv11.Book, error) {
	batchID, bookID := batch.ID, book.ID
	if !hasVerifiedBatchOriginal(book) {
		return book, batchfactoryv11.NewConfigRequiredError(missingSensitiveOriginal)
	}
	writer, ok := c.store.(batchfactoryv11.ProductionTextStore)
	if !ok {
		return book, batchfactoryv11.ErrUnavailable
	}
	effective := c.automationEffectiveSettings(batch, book)
	mode := batchfactorysensitive.ResolveMode(effective)
	limit := bookCaptureCharacters(book)
	var capture int
	if json.Unmarshal(effective["contentCaptureCharacters"], &capture) == nil && capture >= 100 && capture <= 100000 {
		limit = capture
	}
	state := mapOrEmpty(book.SourceMetadata["batchSensitiveStage"])
	input, baseKey, err := sensitiveBaseText(ctx, c.store, owner, batchID, book, state, limit)
	if err != nil {
		return book, err
	}
	adapters := sensitiveAdapters{stages: c.stages}
	version := "batch-sensitive-v1"
	switch mode {
	case batchfactorysensitive.ModeSystem:
		if c.ttsStore != nil {
			adapters.settings, err = c.ttsStore.GetConfig(ctx, owner)
			if err != nil {
				cause := fmt.Errorf("读取本地敏感规则失败：%w", err)
				if verifyOnly {
					return book, cause
				}
				state = map[string]any{"status": "failed", "mode": string(mode), "version": version, "input": sensitiveDigest(input), "raw": sensitiveDigest(book.TxtText), "previous": sensitiveDigest(book.SourceText), "authoredBase": baseKey, "error": cause.Error()}
				updated, saveErr := writer.SaveBookProductionText(ctx, owner, batchID, bookID, batchfactoryv11.SaveBookProductionTextInput{Stage: state, KeepProductionText: true, ExpectedRevision: book.Revision})
				if saveErr != nil {
					return book, errors.Join(cause, saveErr)
				}
				return updated, cause
			}
		}
		version += "/" + novelFetchProcessingFingerprint(adapters.settings)
	case batchfactorysensitive.ModeAI:
		adapters.models = batchSensitiveModels(effective)
		version += "/" + sensitiveDigest(prompt.BatchFactorySensitiveEditsPrompt()) + "/" + strings.Join(adapters.models, ",")
	}
	rawDigest := sensitiveDigest(book.TxtText)
	signature := []string{rawDigest, sensitiveDigest(input), string(mode), version}
	if strings.HasPrefix(baseKey, "working:") {
		signature = append(signature, baseKey)
	}
	key := sensitiveDigest(strings.Join(signature, "\x00"))
	if verifyOnly {
		if state["key"] != key || state["status"] != "done" || state["output"] != sensitiveDigest(book.SourceText) {
			return book, batchfactoryv11.NewConfigRequiredError("当前素材与生产文本版本尚未核对，请先重新生成分镜或提示词")
		}
		return book, nil
	}
	if state["key"] == key {
		switch state["status"] {
		case "done":
			if state["output"] == sensitiveDigest(book.SourceText) {
				return book, nil
			}
			// A refetch can restore the same input after a successful paid edit.
			// Replay its exact substitutions before considering another model call.
			if mode == batchfactorysensitive.ModeAI {
				payload, encodeErr := json.Marshal(map[string]any{"changes": state["changes"]})
				replayed, _, replayErr := applyBatchSensitiveEdits(input, string(payload))
				if encodeErr != nil || replayErr != nil || sensitiveDigest(replayed) != state["output"] {
					return book, errors.New("已完成的 AI 改写记录与当前正文不一致，请核对生产稿")
				}
				saved, err := writer.SaveBookProductionText(ctx, owner, batchID, bookID, batchfactoryv11.SaveBookProductionTextInput{Stage: state, SourceText: replayed, ExpectedRevision: book.Revision})
				if err != nil {
					return book, err
				}
				return c.refreshAutomationBook(ctx, owner, batchID, saved), nil
			}
		case "failed":
			if !manualRetry {
				return book, errors.New(firstMetadataString(state, "error"))
			}
		case "running":
			started, _ := time.Parse(time.RFC3339Nano, firstMetadataString(state, "started"))
			if !manualRetry || time.Since(started) < 15*time.Minute {
				return book, errors.New("敏感内容处理中，若执行已中断请在 15 分钟后手动重试")
			}
		}
	}
	state = map[string]any{"status": "running", "key": key, "mode": string(mode), "version": version, "raw": rawDigest, "input": sensitiveDigest(input), "previous": sensitiveDigest(book.SourceText), "authoredBase": baseKey, "started": time.Now().UTC().Format(time.RFC3339Nano)}
	claimed, err := writer.SaveBookProductionText(ctx, owner, batchID, bookID, batchfactoryv11.SaveBookProductionTextInput{Stage: state, KeepProductionText: true, ExpectedRevision: book.Revision})
	if err != nil {
		return book, err
	}
	processor := batchfactorysensitive.Processor{Rules: adapters, AI: adapters}
	model := ""
	if len(adapters.models) > 0 {
		model = adapters.models[0]
	}
	result, processErr := processor.Process(ctx, batchfactorysensitive.Input{Mode: mode, Text: input, TextModelID: model, Owner: owner, BookID: bookID})
	if processErr == nil {
		current, err := c.store.GetDraft(ctx, owner, "working-front:"+bookID, "working-front-content", batchID)
		if err != nil && !errors.Is(err, batchfactoryv11.ErrNotFound) {
			processErr = err
		} else if strings.TrimSpace(current.Content) != "" && baseKey != batchfactoryv11.WorkingDraftVersion(current.Content) || strings.TrimSpace(current.Content) == "" && strings.HasPrefix(baseKey, "working:") {
			processErr = batchfactoryv11.NewConfigRequiredError("执行期间编辑稿已变化，请用最新稿重试")
		}
	}
	if processErr == nil && strings.TrimSpace(result.ProductionText) == "" {
		processErr = errors.New("敏感处理后的生产文本为空，请调整规则或模型后重试")
	}
	final := copyAnyMap(state)
	final["finished"] = time.Now().UTC().Format(time.RFC3339Nano)
	save := batchfactoryv11.SaveBookProductionTextInput{Stage: final, ExpectedRevision: claimed.Revision}
	if processErr != nil {
		final["status"], final["error"] = "failed", processErr.Error()
		save.KeepProductionText = true
	} else {
		final["status"], final["output"], final["changes"] = "done", sensitiveDigest(result.ProductionText), result.Changes
		save.SourceText = result.ProductionText
	}
	finishCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	updated, err := writer.SaveBookProductionText(finishCtx, owner, batchID, bookID, save)
	for attempt := 1; errors.Is(err, batchfactoryv11.ErrConflict) && attempt < automationBookWriteAttempts; attempt++ {
		fresh, readErr := c.store.GetBatch(finishCtx, owner, batchID)
		if readErr != nil {
			return claimed, readErr
		}
		latest, ok := findBatchBookByID(fresh, bookID)
		pending := mapOrEmpty(latest.SourceMetadata["batchSensitiveStage"])
		if !ok || pending["key"] != key || pending["started"] != state["started"] || pending["status"] != "running" || sensitiveDigest(latest.SourceText) != state["previous"] || sensitiveDigest(latest.TxtText) != rawDigest {
			break
		}
		save.ExpectedRevision = latest.Revision
		updated, err = writer.SaveBookProductionText(finishCtx, owner, batchID, bookID, save)
	}
	if err != nil {
		return claimed, err
	}
	// The narrow write returns the book record. Downstream stages also read
	// derived settings, video and asset state from the hydrated batch view.
	if processErr == nil {
		fresh, readErr := c.store.GetBatch(finishCtx, owner, batchID)
		if readErr != nil {
			return updated, readErr
		}
		latest, ok := findBatchBookByID(fresh, bookID)
		if !ok {
			return updated, batchfactoryv11.ErrNotFound
		}
		if latest.SourceText != updated.SourceText || !hasVerifiedBatchOriginal(latest) {
			return latest, batchfactoryv11.ErrConflict
		}
		updated = latest
	}
	return updated, processErr
}
