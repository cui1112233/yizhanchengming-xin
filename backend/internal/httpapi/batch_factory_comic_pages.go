package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/imagegenerationcache"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/prompt"
)

type comicPageGenerator struct {
	store          batchfactoryv11.Store
	files          *localartifact.Store
	localExecutors *localexecutor.Service
	catalogStore   novelfetchworkshop.Store
	// userConfig 保存 API 接口配置，用于解析 custom:/channel: 图片渠道。
	userConfig     *UserConfigStore
	presetCatalog  *prompt.Catalog
	generationGate *comicPageGenerationGate
	imageCache     imagegenerationcache.Repository
	// resolveModelOwner makes members execute with the manager's model/API
	// channels while all batch rows and generated media remain owned by the
	// member. This mirrors the model dropdown shown to the member.
	resolveModelOwner func(string) string
	// rewritePrompt 接统一文本模型，用于内容审核拒绝后重写该页提示词。
	// 未配置时为 nil，退回本地暴力合规改写。
	rewritePrompt func(context.Context, string, novelfetchworkshop.RewriteRequest) (string, error)
}

func (g comicPageGenerator) modelOwner(owner string) string {
	if g.resolveModelOwner != nil {
		if resolved := strings.TrimSpace(g.resolveModelOwner(owner)); resolved != "" {
			return resolved
		}
	}
	return owner
}

const comicPageGenerationLimit = 35

type comicPageWorkItem struct {
	index int
	page  batchfactoryv11.Video
}

// comicPageFailure 是单页失败的账本条目。失败只影响该页，不影响兄弟页，
// 也不中止整本。
type comicPageFailure struct {
	VideoID string
	Label   string
	Code    string
	Err     error
}

func (f comicPageFailure) Message() string {
	if f.Err == nil {
		return ""
	}
	return f.Err.Error()
}

// comicPageBusinessError 只把中文业务原因交给界面，同时保留 errors.Is 分类。
type comicPageBusinessError struct {
	message string
	cause   error
}

func (e comicPageBusinessError) Error() string { return e.message }

func (e comicPageBusinessError) Unwrap() error { return e.cause }

// comicPageFailureMessage 保证每条页失败都带页标签，但不会重复拼两遍。
func comicPageFailureMessage(f comicPageFailure) string {
	if f.Err == nil {
		return ""
	}
	message := batchfactoryv11.UserFacingErrorText(f.Err.Error())
	label := strings.TrimSpace(f.Label)
	if label != "" && !strings.HasPrefix(message, label) {
		message = label + "：" + message
	}
	return message
}

// comicPageWorkError 聚合本轮所有页级失败。调用方据此写 pageLedger，
// 而不是只看第一条错误。
type comicPageWorkError struct {
	Failures []comicPageFailure
}

func (e *comicPageWorkError) Error() string {
	if e == nil || len(e.Failures) == 0 {
		return ""
	}
	if len(e.Failures) == 1 {
		return comicPageFailureMessage(e.Failures[0])
	}
	return fmt.Sprintf("%d 个漫画页生成失败（首个：%s）", len(e.Failures), comicPageFailureMessage(e.Failures[0]))
}

func (e *comicPageWorkError) Unwrap() error {
	if e == nil || len(e.Failures) == 0 {
		return nil
	}
	return e.Failures[0].Err
}

func (e *comicPageWorkError) videoIDs() map[string]struct{} {
	ids := make(map[string]struct{}, len(e.Failures))
	for _, failure := range e.Failures {
		ids[failure.VideoID] = struct{}{}
	}
	return ids
}

// comicPagePartialError 表示本轮已有多页成功、只剩部分页失败。它保留
// comicPageWorkError 的逐页失败明细供调用方写页级账本，但语义不是"整本失败"：
// 调用方必须安排续跑，下一轮只补缺失页。
type comicPagePartialError struct {
	*comicPageWorkError
	Completed int
	Total     int
}

func (e *comicPagePartialError) Error() string {
	if e == nil {
		return ""
	}
	total := e.Total
	failed := 0
	first := ""
	if e.comicPageWorkError != nil {
		failed = len(e.Failures)
		if failed > 0 {
			first = fmt.Sprintf("（首个：%s）", comicPageFailureMessage(e.Failures[0]))
		}
	}
	if total <= 0 {
		total = e.Completed + failed
	}
	return fmt.Sprintf("已生成 %d/%d 页，%d 页失败%s", e.Completed, total, failed, first)
}

func (e *comicPagePartialError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.comicPageWorkError
}

// comicPageGenerationError 把本轮页级失败聚合成调用方可见的错误：
// 一页都没成功时保持整本失败语义；已有成功页时降级为可续跑的部分失败。
func comicPageGenerationError(failures []comicPageFailure, completed, total int) error {
	if len(failures) == 0 {
		return nil
	}
	sort.SliceStable(failures, func(i, j int) bool { return failures[i].VideoID < failures[j].VideoID })
	workErr := &comicPageWorkError{Failures: failures}
	if completed > 0 {
		return &comicPagePartialError{comicPageWorkError: workErr, Completed: completed, Total: total}
	}
	return workErr
}

// comicAssetShortfall 记录本轮资产参考图的缺失明细。资产图只是页图参考，
// 缺一张不阻断整本；调用方据此记账，并在后续轮次优先回补。
type comicAssetShortfall struct {
	MissingAssetIDs []string
	CriticalTotal   int
	CriticalMissing int
	Billed          bool
	Err             error
}

type comicAssetShortfallFunc func(comicAssetShortfall)

type comicAssetShortfallContextKey struct{}

func withComicAssetShortfall(ctx context.Context, report comicAssetShortfallFunc) context.Context {
	if report == nil {
		return ctx
	}
	return context.WithValue(ctx, comicAssetShortfallContextKey{}, report)
}

func reportComicAssetShortfall(ctx context.Context, shortfall comicAssetShortfall) {
	if ctx == nil {
		return
	}
	report, _ := ctx.Value(comicAssetShortfallContextKey{}).(comicAssetShortfallFunc)
	if report != nil {
		report(shortfall)
	}
}

type comicGenerationHeartbeatFunc func()

type comicGenerationHeartbeatContextKey struct{}

func withComicGenerationHeartbeat(ctx context.Context, beat comicGenerationHeartbeatFunc) context.Context {
	if beat == nil {
		return ctx
	}
	return context.WithValue(ctx, comicGenerationHeartbeatContextKey{}, beat)
}

// reportComicGenerationHeartbeat 只刷新停滞看门狗的心跳，不追加阶段日志。
// 「上游已返回响应头」「图片已落库」这类真实进展都应调用它，避免慢图在
// 重试预算内被判定为 stalled。
func reportComicGenerationHeartbeat(ctx context.Context) {
	if ctx == nil {
		return
	}
	if beat, ok := ctx.Value(comicGenerationHeartbeatContextKey{}).(comicGenerationHeartbeatFunc); ok && beat != nil {
		beat()
	}
}

// comicCriticalAssetCoverage 统计关键资产（漫画角色/场景）总数与缺失数。
// 它们的参考图决定页图里的“谁/在哪”，只有全部缺失才阻断整本。
func comicCriticalAssetCoverage(assets []batchfactoryv11.BookAsset, missing []string) (total, missingCount int) {
	missingSet := make(map[string]struct{}, len(missing))
	for _, id := range missing {
		missingSet[id] = struct{}{}
	}
	for _, asset := range assets {
		if strings.TrimSpace(asset.Source) != "comic" {
			continue
		}
		switch strings.TrimSpace(asset.Kind) {
		case "character", "scene":
			total++
			if _, ok := missingSet[asset.ID]; ok {
				missingCount++
			}
		}
	}
	return total, missingCount
}

// runComicPageWork 快速失败 + 跳过继续：任一页失败只记账，不阻塞兄弟页，
// 也不 cancel 队列。失败页由调用方写入页级账本，下一轮只补未完成的页。
func runComicPageWork(ctx context.Context, work []comicPageWorkItem, concurrency int, generate func(context.Context, comicPageWorkItem) error) error {
	if concurrency < 1 {
		concurrency = 1
	}
	sem := make(chan struct{}, concurrency)
	var workers sync.WaitGroup
	var resultMu sync.Mutex
	failures := make([]comicPageFailure, 0)
	for _, item := range work {
		item := item
		workers.Add(1)
		go func() {
			defer workers.Done()
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()
			if ctx.Err() != nil {
				return
			}
			if err := generate(ctx, item); err != nil {
				disposition := classifyAutomationFailure(err, "comic_images")
				resultMu.Lock()
				failures = append(failures, comicPageFailure{
					VideoID: item.page.ID,
					Label:   item.page.Label,
					Code:    disposition.Code,
					Err:     err,
				})
				resultMu.Unlock()
			}
		}()
	}
	workers.Wait()
	resultMu.Lock()
	defer resultMu.Unlock()
	if len(failures) == 0 {
		return nil
	}
	return &comicPageWorkError{Failures: append([]comicPageFailure(nil), failures...)}
}

// A single-page generation can legitimately take longer than an HTTP request
// should stay open. Keep the old fast path for local/test providers, but hand
// slow provider calls to the detached worker and let the UI read the durable
// page status instead of waiting on this request forever.
const comicPageHTTPFastPathWait = 500 * time.Millisecond

const comicBookDetachedGenerationTimeout = 6 * time.Hour

// Comic generation must not be cancelled merely because the browser tab,
// fetch request, or reverse proxy connection went away. The relay can finish
// successfully after that disconnect, and the finished image still needs to
// be written into the comic page record. Keep a bounded lifetime so a lost
// request cannot leave a worker running forever.
// 页级任务包含“等待整机出图名额 + 上游生成 + 大图正文回传”。线上一次正文
// 回传就可能接近 5 分钟；多批次同时生成时，旧的 15 分钟会让已经排在 FIFO
// 队列里的老页面提前退出，下一轮又回到队尾，形成永久饥饿。和整本任务统一为
// 6 小时有界上限，和自动化整本任务的 hard cap 对齐。它只延长等待队列的
// 生命周期，真正占用上游连接的并发数仍由整机闸门限制，不会放大并发或费用。
const comicPageDetachedGenerationTimeout = 6 * time.Hour

func detachedGenerationContext(parent context.Context, max time.Duration) (context.Context, context.CancelFunc) {
	base := context.WithoutCancel(parent)
	// 既然已经明确脱离浏览器/上层阶段，就不能只忽略 Cancel 却继续继承它的
	// 短 Deadline；否则到点后页面仍会退出 FIFO 队列，稍后重试又回到队尾。
	// 用本阶段自己的有界上限收口，防止后台任务无限驻留。
	return context.WithTimeout(base, max)
}

func comicPageGenerationContext(parent context.Context) (context.Context, context.CancelFunc) {
	return detachedGenerationContext(parent, comicPageDetachedGenerationTimeout)
}

// comicPageGenerationInFlightError 表示同一页还有一次生成没有结束。上一次
// 整本尝试被其它页的错误中止时，它的 detached 生成仍会继续跑，因此重试很
// 容易撞上这个锁。它和内容审核拒绝一样不应中止整本。
func comicPageGenerationInFlightError(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	return strings.Contains(message, "当前漫画页正在生成") || strings.Contains(message, "当前漫画页已有生成任务")
}

// comicPageClaimError 把页级条件写回的版本冲突统一转换为“该页正在生成”。
// 只有版本冲突才走这里：它表示另一个 worker 已经改了同一页，属于在途，
// 不应被上层当成内容审核失败或永久失败。
func comicPageClaimError(err error) error {
	if err == nil || !errors.Is(err, batchfactoryv11.ErrConflict) {
		return err
	}
	return fmt.Errorf("%w: 当前漫画页正在生成，请等待本次结果完成后再重试", batchfactoryv11.ErrConflict)
}

// The provider may finish after the browser request has gone away. Keep the
// whole book pipeline alive long enough to write the returned artifact and
// build the collection; otherwise a successful image can be lost during the
// final database/readback step.
func comicBookGenerationContext(parent context.Context) (context.Context, context.CancelFunc) {
	return detachedGenerationContext(parent, comicBookDetachedGenerationTimeout)
}

type comicReferenceImageResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	PreviewURL string `json:"previewUrl"`
	Source     string `json:"source"`
	Removable  bool   `json:"removable"`
}

func comicReferenceURLParts(value string) []string {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return nil
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	for index := range parts {
		if decoded, decodeErr := url.PathUnescape(parts[index]); decodeErr == nil {
			parts[index] = decoded
		}
	}
	return parts
}

func comicReferencePreviewURL(value string) string {
	parts := comicReferenceURLParts(value)
	switch {
	case len(parts) == 6 && parts[0] == "api" && parts[1] == "local-executor" && parts[3] == "batch-comic-references":
		return "/api/batch-factory/v12/batches/" + url.PathEscape(parts[4]) + "/comic-references/" + url.PathEscape(parts[5]) + "/content"
	case len(parts) == 5 && parts[0] == "api" && parts[1] == "local-executor" && parts[3] == "batch-comic-reference-scenes":
		return "/api/presets/comic-reference-scene/" + url.PathEscape(parts[4]) + "/preview"
	case len(parts) == 5 && parts[0] == "api" && parts[1] == "local-executor" && parts[3] == "batch-comic-materials":
		return "/api/presets/comic-material/" + url.PathEscape(parts[4]) + "/preview"
	case len(parts) == 8 && parts[0] == "api" && parts[1] == "local-executor" && parts[3] == "batch-book-asset-images":
		return "/api/batch-factory/v12/batches/" + url.PathEscape(parts[4]) + "/books/" + url.PathEscape(parts[5]) + "/assets/" + url.PathEscape(parts[6]) + "/images/" + url.PathEscape(parts[7]) + "/content"
	default:
		return strings.TrimSpace(value)
	}
}

func resolvePublishedComicStyleReference(settings batchfactoryv11.SettingsPatch, catalog *prompt.Catalog) (batchfactoryv11.SettingsPatch, error) {
	if catalog == nil {
		return settings, nil
	}
	raw, ok := settings["aiPromptConfig"]
	if !ok || len(raw) == 0 {
		return settings, nil
	}
	var config map[string]any
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, err
	}
	comic, _ := config["comic"].(map[string]any)
	style, _ := comic["style"].(map[string]any)
	styleID := strings.TrimSpace(toString(style["presetId"]))
	if styleID == "" {
		return settings, nil
	}
	// The style picker stores the selected preset ID, while older batches may
	// not have frozen the private prompt body into aiPromptConfig yet. Resolve
	// that body here as a server-side fallback so an already selected style is
	// not reported as "请先选择画风" during prompt preview or image generation.
	if strings.TrimSpace(toString(style["body"])) == "" {
		selected, found, err := catalog.PublishedBody("batch-factory", styleID)
		if err != nil {
			return nil, err
		}
		if found && selected.Slot == "batch.comic-style" && strings.TrimSpace(selected.Body) != "" {
			style["presetId"] = selected.ID
			style["presetName"] = selected.Name
			style["presetSlot"] = selected.Slot
			style["presetVersion"] = selected.Version
			style["body"] = selected.Body
		}
	}
	global, _ := comic["globalReferenceImage"].(map[string]any)
	if !comicReferenceIsAutomatic(global) {
		return settings, nil
	}
	presets, err := catalog.Published("batch-factory")
	if err != nil {
		return nil, err
	}
	var selected *prompt.PublicPreset
	for index := range presets {
		item := &presets[index]
		if item.ID == styleID && item.Slot != nil && *item.Slot == "batch.comic-style" {
			selected = item
			break
		}
	}
	if selected == nil || strings.TrimSpace(selected.ReferenceURL) == "" {
		return settings, nil
	}
	reference := map[string]any{
		"assetId":    "style:" + selected.ID,
		"url":        selected.ReferenceURL,
		"previewUrl": selected.PreviewURL,
		"source":     "style-library",
	}
	for _, key := range []string{"enabled", "applyToAssets", "applyToPages"} {
		value, exists := global[key]
		if !exists {
			value = true
		}
		reference[key] = value
	}
	comic["globalReferenceImage"] = reference
	encoded, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	resolved := make(batchfactoryv11.SettingsPatch, len(settings))
	for key, value := range settings {
		resolved[key] = value
	}
	resolved["aiPromptConfig"] = encoded
	return resolved, nil
}

func comicReferenceIsAutomatic(reference map[string]any) bool {
	if len(reference) == 0 {
		return true
	}
	if strings.TrimSpace(toString(reference["source"])) == "manual-upload" {
		return false
	}
	source := strings.TrimSpace(toString(reference["source"]))
	if source == "builtin" || source == "scene-library" || source == "style-library" {
		return true
	}
	assetID := strings.TrimSpace(toString(reference["assetId"]))
	return strings.HasPrefix(assetID, "scene:") || strings.HasPrefix(assetID, "preset:") || strings.HasPrefix(assetID, "style:")
}

func comicReferenceImageResponses(settings batchfactoryv11.SettingsPatch, book batchfactoryv11.Book, references, pageReferences []string) []comicReferenceImageResponse {
	pageIndexes := make(map[string]int, len(pageReferences))
	for index, reference := range pageReferences {
		pageIndexes[strings.TrimSpace(reference)] = index + 1
	}
	assets := make(map[string]batchfactoryv11.BookAsset, len(book.AssetRecords))
	for _, asset := range book.AssetRecords {
		assets[asset.ID] = asset
	}
	global := batchfactoryv11.ComicGlobalReference(settings)
	result := make([]comicReferenceImageResponse, 0, len(references))
	for index, reference := range references {
		reference = strings.TrimSpace(reference)
		parts := comicReferenceURLParts(reference)
		item := comicReferenceImageResponse{URL: reference, PreviewURL: comicReferencePreviewURL(reference), Name: fmt.Sprintf("参考图 %d", index+1), Source: "reference"}
		if len(parts) > 0 {
			item.ID = parts[len(parts)-1]
		}
		if pageIndex, ok := pageIndexes[reference]; ok {
			item.Name = fmt.Sprintf("手动参考图 %d", pageIndex)
			item.Source = "manual-upload"
			item.Removable = true
		} else if reference == strings.TrimSpace(global.URL) {
			item.Source = "global-reference"
			item.Name = "漫画全局参考图"
			if strings.TrimSpace(global.Scene.Name) != "" {
				item.ID = global.Scene.ID
				item.Name = global.Scene.Name
				item.Source = "scene-library"
			}
			if strings.TrimSpace(global.PreviewURL) != "" {
				item.PreviewURL = global.PreviewURL
			}
		} else if len(parts) == 8 && parts[3] == "batch-book-asset-images" {
			item.Source = "book-asset"
			item.ID = parts[6]
			if asset, ok := assets[parts[6]]; ok && strings.TrimSpace(asset.Name) != "" {
				item.Name = asset.Name
			} else {
				item.Name = "漫画资产"
			}
		}
		result = append(result, item)
	}
	return result
}

// shuihuoRepairedPrompt 是审核被拒后由文本模型重写过的提示词。
// 缓存必须带版本与时间：人工修正提示词后，旧改写版绝不能再压住数据库版本。
type shuihuoRepairedPrompt struct {
	Prompt string
	// PromptRevision 是生成该修复版时数据库提示词（编译前原文）的内容摘要。
	// 页面行的 Revision 会因申领租约、状态写回而自增，不能直接当作提示词
	// 版本；用内容摘要才既能感知人工修改，又不会让每次重试都失效。
	PromptRevision string
	// PageRevision 只作诊断：记录写入缓存时观测到的页行版本。
	PageRevision int64
	// RepairCount 是该页已经自动改写的轮次。跨自动化续跑累计，达到上限后
	// 不再调用上游，直接转 needs_manual。
	RepairCount int
	// LastRefusalRaw 是最近一次上游拒绝原文（按字节截断，见
	// comicPromptRefusalSummary），写进页级账本供人工排查。
	LastRefusalRaw string
	LastFields     []string
	StoredAt       time.Time
}

// shuihuoRepairedPromptTTL 是修复版缓存的最长存活时间。整本重试通常发生在
// 同一天内，12 小时足够复用；超过后强制重新改写一次，才能跟上审核策略或
// 文本模型升级，避免旧改写版长期生效。
const shuihuoRepairedPromptTTL = 12 * time.Hour

// shuihuoRepairedPromptNow 可注入，测试用假时钟验证 TTL，不 sleep。
var shuihuoRepairedPromptNow = func() time.Time { return time.Now().UTC() }

// comicPromptRepairMaxRounds 是单页自动改写的轮次上限。选 3 的依据：
// 一轮改对白/文案、一轮改画面构图、一轮改场景与镜头，覆盖线上被拒的
// 主要形态；再往上继续改写边际收益很低，而且每轮都要真实出图计费。
const comicPromptRepairMaxRounds = 3

// errComicPromptRepairExhausted 表示该页自动改写轮次用尽，必须转人工。
var errComicPromptRepairExhausted = errors.New("comic prompt repair exhausted")

// shuihuoRepairedPrompts 缓存审核被拒后由文本模型重写过的提示词。
// 进程内缓存即可：后续整本自动重试直接复用修复版，不再拿原提示词撞审核。
var shuihuoRepairedPrompts sync.Map

func comicPageRepairKey(batchID, bookID, videoID string) string {
	return batchID + "|" + bookID + "|" + videoID
}

func comicAssetRepairKey(batchID, bookID, assetID string) string {
	return batchID + "|" + bookID + "|" + assetID
}

// comicPromptRevision 用数据库提示词内容派生版本号。提示词一旦被人工或
// 自动流程改写，版本就会变化，旧修复版随之失效。
func comicPromptRevision(prompt string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(prompt)))
	return fmt.Sprintf("%x", sum[:16])
}

// loadShuihuoRepairedPrompt 只返回仍然有效的修复版：版本必须与当前数据库
// 提示词一致，且未超过 TTL；否则立即删除并视为未命中。
func loadShuihuoRepairedPrompt(key, promptRevision string) (shuihuoRepairedPrompt, bool) {
	value, ok := shuihuoRepairedPrompts.Load(key)
	if !ok {
		return shuihuoRepairedPrompt{}, false
	}
	cached, ok := value.(shuihuoRepairedPrompt)
	if !ok || strings.TrimSpace(cached.Prompt) == "" {
		shuihuoRepairedPrompts.Delete(key)
		return shuihuoRepairedPrompt{}, false
	}
	if !cached.StoredAt.IsZero() && shuihuoRepairedPromptNow().Sub(cached.StoredAt) >= shuihuoRepairedPromptTTL {
		shuihuoRepairedPrompts.Delete(key)
		return shuihuoRepairedPrompt{}, false
	}
	if promptRevision != "" && cached.PromptRevision != promptRevision {
		shuihuoRepairedPrompts.Delete(key)
		return shuihuoRepairedPrompt{}, false
	}
	return cached, true
}

func storeShuihuoRepairedPrompt(key string, cached shuihuoRepairedPrompt) {
	cached.StoredAt = shuihuoRepairedPromptNow()
	shuihuoRepairedPrompts.Store(key, cached)
}

// invalidateShuihuoRepairedPrompt 是提示词写入后唯一的失效入口。任何人工
// 写入（页级 PUT、reference-images、settings_patches 补丁）都必须调用它，
// 不允许在业务代码里直接 Delete 缓存。
// 只给 batchID 时按批次清理，给到 bookID 时按书清理，三者齐全时精确清理该页。
func invalidateShuihuoRepairedPrompt(batchID, bookID, videoID string) bool {
	batchID = strings.TrimSpace(batchID)
	if batchID == "" {
		return false
	}
	bookID = strings.TrimSpace(bookID)
	videoID = strings.TrimSpace(videoID)
	if bookID != "" && videoID != "" {
		key := comicPageRepairKey(batchID, bookID, videoID)
		if _, loaded := shuihuoRepairedPrompts.LoadAndDelete(key); loaded {
			log.Printf("comic prompt repair cache invalidated batch=%s book=%s video=%s", batchID, bookID, videoID)
			return true
		}
		return false
	}
	prefix := batchID + "|"
	if bookID != "" {
		prefix = batchID + "|" + bookID + "|"
	}
	invalidated := 0
	shuihuoRepairedPrompts.Range(func(key, _ any) bool {
		if text, ok := key.(string); ok && strings.HasPrefix(text, prefix) {
			shuihuoRepairedPrompts.Delete(text)
			invalidated++
		}
		return true
	})
	if invalidated > 0 {
		log.Printf("comic prompt repair cache invalidated batch=%s book=%s entries=%d", batchID, bookID, invalidated)
	}
	return invalidated > 0
}

// comicSettingsPatchTouchesPrompt 判断 settings_patches 是否写入了页提示词。
// 只有改动提示词的补丁才需要让修复版缓存失效。
func comicSettingsPatchTouchesPrompt(patch batchfactoryv11.SettingsPatch) bool {
	for _, key := range []string{"comicPrompt", "comicImagePrompt"} {
		if _, ok := patch[key]; ok {
			return true
		}
	}
	return false
}

// comicPromptRepair 记录一次自动改写：哪一页、改了哪些字段、改写前后摘要，
// 由自动化链路写进 pageLedger，前端据此显示“这一页被自动改写过”。
type comicPromptRepair struct {
	VideoID string
	Label   string
	Fields  []string
	Before  string
	After   string
	// Attempts 是该页累计自动改写轮次（从 1 开始）。
	Attempts int
	// LastRefusalRaw 是触发本次改写的上游拒绝原文摘要。
	LastRefusalRaw string
	At             time.Time
}

type comicPromptRepairFunc func(comicPromptRepair)

type comicPromptRepairContextKey struct{}

func withComicPromptRepair(ctx context.Context, report comicPromptRepairFunc) context.Context {
	if report == nil {
		return ctx
	}
	return context.WithValue(ctx, comicPromptRepairContextKey{}, report)
}

func reportComicPromptRepair(ctx context.Context, repair comicPromptRepair) {
	if ctx == nil {
		return
	}
	report, _ := ctx.Value(comicPromptRepairContextKey{}).(comicPromptRepairFunc)
	if report != nil {
		report(repair)
	}
}

// comicPromptRefusalSummary 截断上游拒绝原文后写进页级账本。规则：
// 直接按字节截前 500 字节（上游多半是英文 JSON），再追加原长度；500 字节
// 足够看清拒绝类型和关键句，又不会把整段网关 HTML 塞进 job_json。
func comicPromptRefusalSummary(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 500 {
		return value
	}
	return value[:500] + fmt.Sprintf("…（已截断，原长 %d 字节）", len(value))
}

// comicPromptRepairFields 比较修复前后的提示词，指出自动修复改了哪些字段。
// 引号/「」内是台词，其余是画面描述（构图、主体动作、场景细节）。
func comicPromptRepairFields(before, after string) []string {
	if strings.TrimSpace(before) == strings.TrimSpace(after) {
		return nil
	}
	fields := make([]string, 0, 2)
	if comicPromptDialogue(before) != comicPromptDialogue(after) {
		fields = append(fields, "台词")
	}
	if comicPromptVisual(before) != comicPromptVisual(after) {
		fields = append(fields, "画面描述")
	}
	if len(fields) == 0 {
		return []string{"提示词"}
	}
	return fields
}

func comicPromptDialogue(value string) string {
	var builder strings.Builder
	depth := 0
	for _, r := range value {
		switch r {
		case '「', '『', '“', '‘':
			depth++
			builder.WriteRune(r)
			continue
		case '」', '』', '”', '’':
			if depth > 0 {
				depth--
			}
			builder.WriteRune(r)
			continue
		}
		if depth > 0 {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func comicPromptVisual(value string) string {
	var builder strings.Builder
	depth := 0
	for _, r := range value {
		switch r {
		case '「', '『', '“', '‘':
			depth++
			continue
		case '」', '』', '”', '’':
			if depth > 0 {
				depth--
			}
			continue
		}
		if depth == 0 {
			builder.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(builder.String()), "")
}

const comicPromptRepairSystemPrompt = "你是漫画图片提示词的合规改写器。上游图片模型审核严格，包含武器、刀具、伤害动作、血液、伤口、昏迷、殴打等描写的提示词会被拒绝。你负责在保留人物身份、画风、分格与台词原文的前提下，改写被拒绝的画面描述（构图、主体动作、视线、景别、场景细节），把冲突场面替换成可以稳定通过审核的版本。"

func repairComicImagePrompt(ctx context.Context, owner, imagePrompt, rejection string, rewrite func(context.Context, string, novelfetchworkshop.RewriteRequest) (string, error), asset bool) string {
	fallback := sanitizeShuihuoImagePromptViolence(imagePrompt)
	if rewrite == nil {
		return fallback
	}
	constraints := "保留画风描述、分格数量与阅读顺序、人物身份和台词原文；必须同时改写被拒绝的画面描述（构图、主体动作、视线、景别与场景细节），不得只改台词或文案"
	if asset {
		constraints = "保留资产身份、画风、正面/侧面/背面三视图、构图和所有参考图约束"
	}
	request := novelfetchworkshop.RewriteRequest{
		SystemPrompt: comicPromptRepairSystemPrompt,
		UserPrompt:   fmt.Sprintf("请重写下面的漫画图片提示词。硬性要求：\n1. %s；\n2. 删除武器、刀具、利刃、伤害动作、血液、伤口、昏迷、殴打、掐颈等描写；\n3. 冲突和伤害必须改写成安全的画面替代：调整构图、主体姿态、视线、景别或场景细节（如空镜、远景背影、后退躲避、被扶住、物品掉落），必要时更换场景，但不得增删剧情事实；\n4. 不新增人物，不改变资产身份和整体画风；\n5. 只输出改写后的完整提示词正文，不要解释，不要加代码块。\n\n上游拒绝原因：%s\n\n【原提示词】\n%s", constraints, strings.TrimSpace(rejection), imagePrompt),
		Temperature:  0.3, MaxTokens: 2600,
	}
	repaired, err := rewrite(ctx, owner, request)
	if err != nil {
		log.Printf("comic prompt repair failed owner=%s err=%v", owner, err)
		return fallback
	}
	repaired = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(repaired, "```text"), "```"), "```"))
	if len([]rune(repaired)) < 120 {
		return fallback
	}
	return sanitizeShuihuoImagePromptViolence(repaired)
}

func (g comicPageGenerator) repairComicPageImagePrompt(ctx context.Context, owner, imagePrompt, rejection string) string {
	return repairComicImagePrompt(ctx, owner, imagePrompt, rejection, g.rewritePrompt, false)
}

// persistRepairedComicPrompt 把自动改写结果同时写回页级提示词与 video 级
// settings_patches，保证两者一致；进程重启或后续续跑仍使用同一份提示词。
// 写回失败时保留旧提示词，调用方继续用本轮修复版生成，下一轮再尝试写回。
func (g comicPageGenerator) persistRepairedComicPrompt(ctx context.Context, comicStore batchfactoryv11.ComicPageStore, owner string, batch batchfactoryv11.Batch, book batchfactoryv11.Book, page batchfactoryv11.Video, current batchfactoryv11.ComicPage, input batchfactoryv11.UpsertComicPageInput, effective batchfactoryv11.SettingsPatch, repaired string, referenceURLs []string) (batchfactoryv11.ComicPage, string, error) {
	repairedRaw := strings.TrimSpace(batchfactoryv11.StripCompiledComicImagePrompt(effective, batchfactoryv11.ComicImagePurposePage, repaired))
	if repairedRaw == "" {
		return current, "", fmt.Errorf("repaired prompt is empty after stripping")
	}
	writeInput := input
	writeInput.ImagePrompt = repairedRaw
	writeInput.Status = current.Status
	writeInput.ExpectedRevision = current.Revision
	if contract, err := batchfactoryv11.BuildComicImageContract(effective, batchfactoryv11.ComicImagePurposePage, "", comicPagePromptContent(input.ComicPrompt, repairedRaw), nil, referenceURLs); err == nil {
		writeInput.SnapshotHash = contract.SnapshotHash
	}
	updated, err := comicStore.UpsertComicPage(ctx, owner, batch.ID, book.ID, page.ID, writeInput)
	if err != nil {
		return current, "", err
	}
	comicPromptRaw, _ := json.Marshal(input.ComicPrompt)
	imagePromptRaw, _ := json.Marshal(repairedRaw)
	if _, err := g.store.SaveSettings(ctx, owner, batchfactoryv11.ScopeRef{
		Kind: batchfactoryv11.ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: page.ID,
	}, batchfactoryv11.SettingsUpdate{
		Patch:            batchfactoryv11.SettingsPatch{"comicPrompt": comicPromptRaw, "comicImagePrompt": imagePromptRaw},
		ExpectedRevision: page.SettingsState.Revision,
	}); err != nil {
		// 页级提示词仍是权威来源；settings 写回失败只影响页面外的展示一致性。
		log.Printf("comic page repaired prompt settings writeback failed batch=%s book=%s video=%s err=%v", batch.ID, book.ID, page.ID, err)
	}
	return updated, repairedRaw, nil
}

func comicPagePromptContent(comicPrompt, imagePrompt string) string {
	comicPrompt = strings.TrimSpace(comicPrompt)
	imagePrompt = strings.TrimSpace(imagePrompt)
	if comicPrompt == "" {
		return imagePrompt
	}
	if imagePrompt == "" || imagePrompt == comicPrompt {
		return comicPrompt
	}
	return "【分格导演描述】\n" + comicPrompt + "\n\n【本页基础出图描述】\n" + imagePrompt
}

func (g comicPageGenerator) generatePage(ctx context.Context, owner, batchID, bookID, videoID string, input batchfactoryv11.UpsertComicPageInput) (page batchfactoryv11.ComicPage, image batchfactoryv11.ComicPageImage, err error) {
	log.Printf("comic page start batch=%s book=%s video=%s provider=%s model=%s", batchID, bookID, videoID, input.Provider, input.ModelID)
	comicStore, ok := g.store.(batchfactoryv11.ComicPageStore)
	if !ok || g.files == nil || g.catalogStore == nil {
		return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, fmt.Errorf("%w: 漫画图片生成服务尚未配置", batchfactoryv11.ErrUnavailable)
	}
	// The in-memory gate belongs to this Go service instance.
	ownsLocalGenerationGate := false
	if g.generationGate != nil {
		release, acquired := g.generationGate.tryAcquire(strings.Join([]string{owner, batchID, bookID, videoID}, "/"))
		if !acquired {
			return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, fmt.Errorf("%w: 当前漫画页正在生成，请等待本次结果完成后再重试", batchfactoryv11.ErrConflict)
		}
		ownsLocalGenerationGate = true
		defer release()
	}
	batch, err := g.store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, err
	}
	var selectedBook *batchfactoryv11.Book
	for index := range batch.Books {
		if batch.Books[index].ID == bookID {
			selectedBook = &batch.Books[index]
			break
		}
	}
	if selectedBook == nil {
		return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, batchfactoryv11.ErrNotFound
	}
	var selectedPage *batchfactoryv11.Video
	selectedPageIndex := -1
	for index := range selectedBook.Videos {
		if selectedBook.Videos[index].ID == videoID {
			selectedPage = &selectedBook.Videos[index]
			selectedPageIndex = index
			break
		}
	}
	if selectedPage == nil {
		return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, batchfactoryv11.ErrNotFound
	}
	if selectedPageIndex >= comicPageGenerationLimit {
		return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, fmt.Errorf("%w: 漫画生图最多执行前 %d 页", batchfactoryv11.ErrConflict, comicPageGenerationLimit)
	}
	effective := batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, selectedBook.SettingsState.Patch)
	effective, err = resolvePublishedComicStyleReference(effective, g.presetCatalog)
	if err != nil {
		return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, err
	}
	// The page editor can still hold an older provider/model after the user
	// changes the book engine settings. Comic generation must use the current
	// saved book settings, otherwise a page retry can silently fall back to the
	// old local Doubao executor.
	if configuredProvider := comicSettingString(effective, "imageProvider", ""); configuredProvider != "" {
		input.Provider = configuredProvider
	}
	if configuredModel := comicSettingString(effective, "imageModelId", ""); configuredModel != "" {
		input.ModelID = configuredModel
	}
	// A page reference is durable page data, so a retry keeps the user's selected image.
	var storedPage batchfactoryv11.ComicPage
	hasStoredPage := false
	if pages, listErr := comicStore.ListComicPages(ctx, owner, batchID, bookID); listErr != nil {
		return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, listErr
	} else {
		for _, page := range pages {
			if page.VideoID != videoID {
				continue
			}
			storedPage, hasStoredPage = page, true
			if comicPageGenerationTaskExists(page) {
				if !comicPageGenerationLeaseRecoverable(page, ownsLocalGenerationGate, time.Now().UTC()) {
					return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, fmt.Errorf("%w: 当前漫画页已有生成任务，请等待完成后再重新生成", batchfactoryv11.ErrConflict)
				}
				// The durable lease survives a service restart. Recover it only
				// after the safety window; acquiring this process-local gate does
				// not prove the remote provider stopped working.
				recovered, recoverErr := comicStore.UpsertComicPage(ctx, owner, batchID, bookID, videoID, batchfactoryv11.UpsertComicPageInput{
					ComicPrompt: page.ComicPrompt, ImagePrompt: page.ImagePrompt, ReferenceImageURLs: page.ReferenceImageURLs,
					ModelID: page.ModelID, AspectRatio: page.AspectRatio, Status: "generation_failed", SnapshotHash: page.SnapshotHash,
					ExpectedRevision: page.Revision,
				})
				if recoverErr != nil {
					if errors.Is(recoverErr, batchfactoryv11.ErrConflict) {
						return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, comicPageClaimError(recoverErr)
					}
					return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, fmt.Errorf("回收旧漫画页生成任务失败: %w", recoverErr)
				}
				storedPage = recovered
				log.Printf("comic page recovered stale generation lease batch=%s book=%s video=%s", batchID, bookID, videoID)
			}
			if len(page.ReferenceImageURLs) > 0 {
				input.ReferenceImageURLs = page.ReferenceImageURLs
			}
		}
	}
	// The comic page row is the single source of truth for a page's prompts.
	// Older callers (including the private-customization workspace before it
	// sent a full payload) may omit them; fill them from the stored row instead
	// of rejecting the request.
	if hasStoredPage {
		if strings.TrimSpace(input.ComicPrompt) == "" {
			input.ComicPrompt = storedPage.ComicPrompt
		}
		if strings.TrimSpace(input.ImagePrompt) == "" {
			input.ImagePrompt = storedPage.ImagePrompt
		}
	}
	if strings.TrimSpace(input.ImagePrompt) == "" {
		return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, fmt.Errorf("%w: %s 缺少独立的图片提示词", batchfactoryv11.ErrConflict, selectedPage.Label)
	}
	// 库里只保存原始提示词。历史数据可能已经把编译结果写回（表现为
	// 【最终画风锁定】段重复叠加），这里先还原成原文，避免再次编译时继续加长。
	input.ImagePrompt = batchfactoryv11.StripCompiledComicImagePrompt(effective, batchfactoryv11.ComicImagePurposePage, input.ImagePrompt)
	if strings.TrimSpace(input.ImagePrompt) == "" {
		return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, fmt.Errorf("%w: %s 缺少独立的图片提示词", batchfactoryv11.ErrConflict, selectedPage.Label)
	}
	input.AspectRatio = batchfactoryv11.EffectiveImageAspectRatio(effective)
	assetMatchContent := comicPagePromptContent(input.ComicPrompt, input.ImagePrompt)
	assetImageURLs, err := comicPageAssetReferenceURLsForPage(ctx, g.store, owner, batch.ID, selectedBook.ID, *selectedPage, assetMatchContent)
	if err != nil {
		return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, err
	}
	log.Printf("comic page references batch=%s book=%s video=%s assets=%d", batchID, bookID, videoID, len(assetImageURLs))
	contract, err := batchfactoryv11.BuildComicImageContract(effective, batchfactoryv11.ComicImagePurposePage, "", comicPagePromptContent(input.ComicPrompt, input.ImagePrompt), assetImageURLs, input.ReferenceImageURLs)
	if err != nil {
		return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, err
	}
	// 编译结果只用于本次出图请求，绝不写回 comic_pages；写库的仍是原始提示词。
	// SnapshotHash 记录本次实际使用的画风/参考图快照，用于判断配置是否过期。
	input.SnapshotHash = contract.SnapshotHash
	input.Status = "generating"
	// 申领页级租约时按读到的 revision 做条件更新：如果另一个 worker 已经
	// 抢先申领，这里的写回会命中版本冲突，而不是覆盖对方的 generating 状态。
	input.ExpectedRevision = 0
	if hasStoredPage {
		input.ExpectedRevision = storedPage.Revision
	}
	createdPage, err := comicStore.UpsertComicPage(ctx, owner, batchID, bookID, videoID, input)
	if err != nil {
		return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, comicPageClaimError(err)
	}
	generationFailed := true
	defer func() {
		if !generationFailed {
			return
		}
		failedInput := input
		failedInput.Status = "generation_failed"
		// 失败原因按“中文主文案 + \n上游原文：”落库，前端按分隔符切第一行做展示，
		// 与书籍级/自动化级 BatchFactoryErrorNotice 保持同一格式。
		failedInput.GenerationError = comicPagePersistedError(err)
		// 失败写回同样按 owner/revision 条件更新，避免把另一轮已经写好的
		// 页面（或新租约）覆盖成 generation_failed。
		failedInput.ExpectedRevision = createdPage.Revision
		_, _ = comicStore.UpsertComicPage(context.Background(), owner, batchID, bookID, videoID, failedInput)
	}()
	config, err := loadShuihuoOwnerConfig(ctx, g.store, owner, g.catalogStore)
	if err != nil {
		return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, err
	}
	target, err := resolveShuihuoImageTargetWithChannels(ctx, g.userConfig, g.modelOwner(owner), createdPage.ModelID, createdPage.ModelID, createdPage.AspectRatio, input.Provider, shuihuoCollection(config, "modelCatalog", nil))
	if err != nil {
		return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, err
	}
	log.Printf("comic page target batch=%s book=%s video=%s provider=%s model=%s", batchID, bookID, videoID, target.Provider, target.ModelID)
	generationCtx, cancelGeneration := comicPageGenerationContext(ctx)
	defer cancelGeneration()
	// 单页重试不能绕过统一配置的出图并发上限。
	generationCtx = withShuihuoImageStage(withShuihuoImageConcurrency(generationCtx, comicSettingInt(effective, "imageConcurrency", shuihuoImageConcurrencyDefault)), shuihuoImageStagePage)
	// 把批次生效设置注入上下文，让出图保存前读取水印配置。
	generationCtx = withShuihuoEffectiveSettings(generationCtx, comicSettingsToMap(effective))
	generationPrompt := contract.Prompt
	repairKey := comicPageRepairKey(batchID, bookID, videoID)
	// 修复版缓存用“数据库原始提示词”派生版本：人工改写会改变版本并让缓存
	// 失效；我们自己写回的修复版会带着新版本写入，跨轮次才能继续累计。
	promptRevision := comicPromptRevision(comicPagePromptContent(input.ComicPrompt, input.ImagePrompt))
	repairedCache, hasRepairedCache := loadShuihuoRepairedPrompt(repairKey, promptRevision)
	if hasRepairedCache {
		generationPrompt = repairedCache.Prompt
		log.Printf("comic page reuse repaired prompt batch=%s book=%s video=%s rounds=%d", batchID, bookID, videoID, repairedCache.RepairCount)
	}
	// 轮次用尽后不再调用上游，直接把这一页转 needs_manual，避免继续计费。
	if hasRepairedCache && repairedCache.RepairCount >= comicPromptRepairMaxRounds {
		return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, fmt.Errorf(
			"%w: %s 已自动改写 %d 轮仍被内容审核拒绝（最后拒绝：%s）",
			errComicPromptRepairExhausted, selectedPage.Label, repairedCache.RepairCount,
			comicPromptRefusalSummary(repairedCache.LastRefusalRaw))
	}
	// CM·文字插画版：把本页 current_page 原文逐字拼进最终出图 prompt，不依赖
	// 文本模型转述。其它预设原样放行。
	generationPrompt, literalTextInjected := comicPageFinalImagePrompt(generationPrompt, effective, *selectedPage)
	if batchfactoryv11.ComicPageIllustratedTextPreset(effective) {
		switch {
		case selectedPage.ComicSourceSegment == nil || strings.TrimSpace(selectedPage.ComicSourceSegment.SourceText) == "":
			log.Printf("comic page literal text skipped batch=%s book=%s video=%s reason=no_current_page", batchID, bookID, videoID)
		case literalTextInjected:
			log.Printf("comic page literal text injected batch=%s book=%s video=%s chars=%d", batchID, bookID, videoID, len(selectedPage.ComicSourceSegment.SourceText))
			// 默认关闭；只在显式审计时打印实际写入出图 prompt 的约束块，便于逐字核对。
			if os.Getenv("QIANTIE_COMIC_LITERAL_TEXT_AUDIT_LOG") == "1" {
				block := batchfactoryv11.ComicPageLiteralTextBlock(selectedPage.ComicSourceSegment.SourceText)
				log.Printf("comic page literal text audit batch=%s book=%s video=%s prompt_chars=%d block_offset=%d block=%q", batchID, bookID, videoID, len(generationPrompt), strings.LastIndex(generationPrompt, block), block)
			}
		case strings.Contains(generationPrompt, selectedPage.ComicSourceSegment.SourceText):
			log.Printf("comic page literal text deduped batch=%s book=%s video=%s", batchID, bookID, videoID)
		}
	}
	mediaType, storageRef, source, err := generateShuihuoAssetImageArtifactWithRetry(generationCtx, g.store, g.files, g.presetCatalog, owner, target, generationPrompt, contract.ReferenceImageURLs, g.localExecutors)
	if errors.Is(err, errShuihuoImageContentPolicy) {
		// 两次降级仍被审核拒绝后，不要拿原提示词反复重试，而是先用文本模型
		// 把整页提示词改写成合规版本（允许改构图/主体/场景），再生成一次。
		// 修复结果写回页级提示词与 video settings，并缓存轮次；续跑时累计，
		// 达到 comicPromptRepairMaxRounds 后转人工。
		refusal := comicPromptRefusalSummary(err.Error())
		roundsUsed := 0
		if hasRepairedCache {
			roundsUsed = repairedCache.RepairCount
		}
		if roundsUsed >= comicPromptRepairMaxRounds {
			return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, fmt.Errorf(
				"%w: %s 已自动改写 %d 轮仍被内容审核拒绝（最后拒绝：%s）",
				errComicPromptRepairExhausted, selectedPage.Label, roundsUsed, refusal)
		}
		if progress := comicGenerationProgressFromContext(ctx); progress != nil {
			progress("comic_images", fmt.Sprintf("%s 被内容审核拒绝，正在第 %d/%d 轮自动改写提示词（含画面构图）", selectedPage.Label, roundsUsed+1, comicPromptRepairMaxRounds))
		}
		repaired := g.repairComicPageImagePrompt(generationCtx, owner, generationPrompt, err.Error())
		repaired, _ = comicPageFinalImagePrompt(repaired, effective, *selectedPage)
		if repaired == "" || repaired == generationPrompt {
			return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, fmt.Errorf(
				"%w: %s 第 %d 轮改写未产生变化（最后拒绝：%s）",
				errComicPromptRepairExhausted, selectedPage.Label, roundsUsed+1, refusal)
		}
		fields := comicPromptRepairFields(generationPrompt, repaired)
		updatedRevision := promptRevision
		if updatedPage, repairedRaw, writeErr := g.persistRepairedComicPrompt(generationCtx, comicStore, owner, batch, *selectedBook, *selectedPage, createdPage, input, effective, repaired, contract.ReferenceImageURLs); writeErr != nil {
			log.Printf("comic page repaired prompt writeback failed batch=%s book=%s video=%s err=%v", batchID, bookID, videoID, writeErr)
		} else {
			createdPage = updatedPage
			input.ImagePrompt = repairedRaw
			updatedRevision = comicPromptRevision(comicPagePromptContent(input.ComicPrompt, repairedRaw))
		}
		storeShuihuoRepairedPrompt(repairKey, shuihuoRepairedPrompt{
			Prompt:         repaired,
			PromptRevision: updatedRevision,
			PageRevision:   createdPage.Revision,
			RepairCount:    roundsUsed + 1,
			LastRefusalRaw: refusal,
			LastFields:     fields,
		})
		// 页级账本记录：哪一页被自动改写、改了哪些字段、前后摘要与拒绝原文。
		reportComicPromptRepair(ctx, comicPromptRepair{
			VideoID:        videoID,
			Label:          selectedPage.Label,
			Fields:         fields,
			Before:         generationPrompt,
			After:          repaired,
			Attempts:       roundsUsed + 1,
			LastRefusalRaw: refusal,
			At:             shuihuoRepairedPromptNow(),
		})
		if progress := comicGenerationProgressFromContext(ctx); progress != nil {
			progress("comic_images", fmt.Sprintf("%s 提示词已改写为合规版本（第 %d 轮），正在重新生成", selectedPage.Label, roundsUsed+1))
		}
		mediaType, storageRef, source, err = generateShuihuoAssetImageArtifactWithRetry(generationCtx, g.store, g.files, g.presetCatalog, owner, target, repaired, contract.ReferenceImageURLs, g.localExecutors)
	}
	if err != nil {
		return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, err
	}
	createdImage, err := comicStore.CreateComicPageImage(generationCtx, owner, batchID, bookID, createdPage.ID, batchfactoryv11.CreateComicPageImageInput{StorageRef: storageRef, MediaType: mediaType, Source: source})
	if err != nil {
		_ = g.files.Remove(storageRef)
		return batchfactoryv11.ComicPage{}, batchfactoryv11.ComicPageImage{}, err
	}
	createdImage.BatchID, createdImage.BookID = createdPage.BatchID, createdPage.BookID
	createdPage.Images = append(createdPage.Images, createdImage)
	generationFailed = false
	return createdPage, createdImage, nil
}

// comicPagePersistedError 把页级失败整理成可持久化的两段式文案：第一行中文主文案，
// 原始信息放在 "\n上游原文：" 之后，与书籍级/自动化级错误保持同一格式。
func comicPagePersistedError(err error) string {
	const fallback = "漫画页生成失败，请重试"
	if err == nil {
		return fallback
	}
	text := strings.TrimSpace(err.Error())
	if text == "" {
		return fallback
	}
	headline, raw, found := strings.Cut(text, "\n上游原文：")
	headline = batchfactoryv11.UserFacingErrorText(strings.TrimSpace(headline))
	if headline == "" {
		headline = fallback
	}
	raw = strings.TrimSpace(raw)
	if !found && text != headline {
		// 原始错误没有标准分隔符时，把完整原文放到第二段，主文案只留业务结论。
		raw = text
	}
	if raw == "" || raw == headline {
		return headline
	}
	return headline + "\n上游原文：" + raw
}

func comicPageAssetLookupText(page batchfactoryv11.Video, imagePrompt string) string {
	roleNames := make([]string, 0, len(page.ComicRoleStage)*2)
	for shortName, formalName := range page.ComicRoleStage {
		roleNames = append(roleNames, shortName, formalName)
	}
	return strings.Join([]string{imagePrompt, strings.Join(roleNames, "、"), strings.Join(page.ComicRoles, "、"), strings.Join(page.ComicScenePresets, "、")}, "\n")
}

// comicPageRefillDelay 是整本出图首轮结束后、对在途锁跳过的缺页进行补图前的
// 最长等待时长。独立变量便于测试收紧。
var comicPageRefillDelay = 60 * time.Second

// comicPageRefillPollInterval 是补图前的页状态轮询间隔。原来固定睡满
// comicPageRefillDelay 才补图，在途页提前完成时这段时间纯属空转。
var comicPageRefillPollInterval = 5 * time.Second

// comicPageSettleDelay 是补图轮之后、构建合集之前等待“仍在别的 worker 手里”
// 的页面结束的最长时间。页级生成上下文是 15 分钟，这里取 12 分钟，避免把
// 在途页误判成缺图，也避免真的死等。
var comicPageSettleDelay = 12 * time.Minute

// waitComicPageSettled 轮询页面状态，直到请求的页都不再处于未过期在途状态，
// 或等待超时。它不清理 lease，只在真正缺图且没有在途任务时立即返回。
func waitComicPageSettled(ctx context.Context, store batchfactoryv11.ComicPageStore, owner, batchID, bookID string, videos []batchfactoryv11.Video, pages []batchfactoryv11.ComicPage) ([]batchfactoryv11.ComicPage, error) {
	deadline := time.Now().Add(comicPageSettleDelay)
	for {
		if !comicPagesHaveInFlightGeneration(pages, videos) {
			return pages, nil
		}
		if time.Now().After(deadline) {
			return pages, nil
		}
		timer := time.NewTimer(comicPageRefillPollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return pages, ctx.Err()
		case <-timer.C:
		}
		if progress := comicGenerationProgressFromContext(ctx); progress != nil {
			progress("comic_images", "仍有漫画页在其它任务中生成，等待释放后继续补齐")
		}
		latest, err := store.ListComicPages(ctx, owner, batchID, bookID)
		if err != nil {
			return pages, err
		}
		pages = latest
	}
}

// comicPagesHaveInFlightGeneration 判断是否还有未过期的在途页生成任务。
// 没有在途任务时补图轮不必等待，可以立刻补齐缺失页。
func comicPagesHaveInFlightGeneration(pages []batchfactoryv11.ComicPage, videos []batchfactoryv11.Video) bool {
	now := time.Now().UTC()
	for _, page := range pages {
		if !comicVideoInList(videos, page.VideoID) {
			continue
		}
		if comicPageGenerationTaskExists(page) && !comicPageGenerationTaskIsStale(page, now) {
			return true
		}
	}
	return false
}

// missingComicPageWorkItems 返回还没有主图的页面任务，用于整本生成的补图轮。
// 页序 index 与原 generationVideos 对齐，保证进度文案与合集顺序不变。
func missingComicPageWorkItems(generationVideos []batchfactoryv11.Video, pages []batchfactoryv11.ComicPage) []comicPageWorkItem {
	work := make([]comicPageWorkItem, 0)
	for pageIndex, page := range generationVideos {
		if comicPageHasPrimary(pages, page.ID) {
			continue
		}
		work = append(work, comicPageWorkItem{index: pageIndex, page: page})
	}
	return work
}

func (g comicPageGenerator) generateBook(ctx context.Context, owner, batchID, bookID string) (batchfactoryv11.ComicCollection, error) {
	detachedCtx, cancel := comicBookGenerationContext(ctx)
	defer cancel()
	ctx = detachedCtx
	ctx = withShuihuoImageGenerationCache(ctx, g.imageCache)
	comicStore, ok := g.store.(batchfactoryv11.ComicPageStore)
	if !ok {
		return batchfactoryv11.ComicCollection{}, fmt.Errorf("%w: 漫画合集存储尚未配置", batchfactoryv11.ErrUnavailable)
	}
	batch, err := g.store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return batchfactoryv11.ComicCollection{}, err
	}
	var book *batchfactoryv11.Book
	for index := range batch.Books {
		if batch.Books[index].ID == bookID {
			book = &batch.Books[index]
			break
		}
	}
	if book == nil {
		return batchfactoryv11.ComicCollection{}, batchfactoryv11.ErrNotFound
	}
	effective := batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
	// 出图并发来自统一配置，漫画资产图与漫画页共用同一个上限。
	imageConcurrency := shuihuoClampImageConcurrency(comicSettingInt(effective, "imageConcurrency", shuihuoImageConcurrencyDefault))
	ctx = withShuihuoImageConcurrency(ctx, imageConcurrency)
	provider := comicSettingString(effective, "imageProvider", "")
	modelID := comicSettingString(effective, "imageModelId", "")
	aspectRatio := batchfactoryv11.EffectiveImageAspectRatio(effective)
	assets, err := g.store.ListBookAssets(ctx, owner, batchID, bookID)
	if err != nil {
		return batchfactoryv11.ComicCollection{}, err
	}
	book.AssetRecords = assets
	missingAssetIDs, err := missingComicAssetImageIDs(ctx, g.store, owner, batchID, *book)
	if err != nil {
		return batchfactoryv11.ComicCollection{}, err
	}
	if !batchfactoryv11.AssetImagesEnabled(effective) {
		if len(missingAssetIDs) > 0 {
			if report := comicGenerationProgressFromContext(ctx); report != nil {
				report("comic_asset_images", fmt.Sprintf("因资产图片开关关闭，已跳过 %d 个漫画资产参考图生成", len(missingAssetIDs)))
			}
		}
	} else if len(missingAssetIDs) > 0 {
		if report := comicGenerationProgressFromContext(ctx); report != nil {
			report("comic_asset_images", fmt.Sprintf("正在生成 %d 个漫画资产参考图", len(missingAssetIDs)))
		}
		assetBatch := batch
		assetBatch.Books = []batchfactoryv11.Book{*book}
		_, assetErr := generateShuihuoAssetImagesWithRepair(ctx, g.store, g.files, owner, assetBatch, missingAssetIDs, modelID, "16:9", provider, g.localExecutors, g.presetCatalog, g.rewritePrompt, g.userConfig, g.modelOwner(owner), g.catalogStore)
		remainingAssetIDs, verifyErr := missingComicAssetImageIDs(ctx, g.store, owner, batchID, *book)
		if verifyErr != nil {
			return batchfactoryv11.ComicCollection{}, verifyErr
		}
		if len(remainingAssetIDs) > 0 {
			criticalTotal, criticalMissing := comicCriticalAssetCoverage(book.AssetRecords, remainingAssetIDs)
			// 关键资产（角色/场景）参考图全缺时继续出图只会产出错角色/错场景，
			// 必须阻断；其余缺图降级为「用已有资产继续 + 记账回补」，
			// 不再因为一张参考图把整本书判死。
			if criticalTotal > 0 && criticalMissing >= criticalTotal {
				return batchfactoryv11.ComicCollection{}, fmt.Errorf("%w: 漫画资产图未全部写回（关键资产 %d 个全部缺失）", batchfactoryv11.ErrConflict, criticalTotal)
			}
			reportComicAssetShortfall(ctx, comicAssetShortfall{
				MissingAssetIDs: append([]string(nil), remainingAssetIDs...),
				CriticalTotal:   criticalTotal,
				CriticalMissing: criticalMissing,
				Billed:          comicBilledReadFailure(assetErr),
				Err:             assetErr,
			})
			if report := comicGenerationProgressFromContext(ctx); report != nil {
				report("comic_asset_images", fmt.Sprintf("有 %d 个资产参考图缺失（关键资产可用 %d/%d），先用现有资产继续出图，缺失项稍后回补",
					len(remainingAssetIDs), criticalTotal-criticalMissing, criticalTotal))
			}
		}
	}
	existingPages, err := comicStore.ListComicPages(ctx, owner, batchID, bookID)
	if err != nil {
		return batchfactoryv11.ComicCollection{}, err
	}
	progress := comicGenerationProgressFromContext(ctx)
	generationVideos := book.Videos
	if len(generationVideos) > comicPageGenerationLimit {
		generationVideos = generationVideos[:comicPageGenerationLimit]
	}
	totalPages := len(generationVideos)
	completedPages := 0
	for _, page := range existingPages {
		if !comicVideoInList(generationVideos, page.VideoID) {
			continue
		}
		if comicPageHasPrimary(existingPages, page.VideoID) {
			completedPages++
		}
	}
	// 已完成的页直接移出工作集：重试是续跑，不是重来。
	continuation := buildComicPageContinuation(generationVideos, existingPages)
	indexByVideoID := make(map[string]int, len(generationVideos))
	for index, page := range generationVideos {
		indexByVideoID[page.ID] = index
	}
	failedPageIDs := make(map[string]struct{})
	failures := make([]comicPageFailure, 0)
	collectPageFailures := func(err error) bool {
		var workErr *comicPageWorkError
		if !errors.As(err, &workErr) {
			return false
		}
		for _, failure := range workErr.Failures {
			if _, exists := failedPageIDs[failure.VideoID]; exists {
				continue
			}
			failedPageIDs[failure.VideoID] = struct{}{}
			failures = append(failures, failure)
		}
		return true
	}
	work := make([]comicPageWorkItem, 0, len(continuation.Pending))
	for _, page := range continuation.Pending {
		// 缺提示词的页只登记该页，不再中止整本；其余页照常出图。
		if strings.TrimSpace(page.ComicPrompt) == "" || strings.TrimSpace(page.ComicImagePrompt) == "" {
			message := fmt.Sprintf("%s 还没有漫画提示词", page.Label)
			if strings.TrimSpace(page.ComicPrompt) != "" {
				message = fmt.Sprintf("%s 缺少独立的图片提示词", page.Label)
			}
			failures = append(failures, comicPageFailure{
				VideoID: page.ID,
				Label:   page.Label,
				Code:    automationCodePromptMissing,
				Err:     comicPageBusinessError{message: message, cause: batchfactoryv11.ErrConflict},
			})
			failedPageIDs[page.ID] = struct{}{}
			continue
		}
		work = append(work, comicPageWorkItem{index: indexByVideoID[page.ID], page: page})
	}
	// A single book used to generate every page serially, which made the
	// "立即执行" flow look stuck even though the provider was working. Keep
	// page order for the final collection, while allowing several page workers
	// to queue; the provider limit is the configured image concurrency above.
	const pageWorkers = 5
	if progress != nil {
		progress("comic_images", fmt.Sprintf("准备并发生成漫画图片（已完成 %d/%d 张，出图并发 %d 路）", completedPages, totalPages, imageConcurrency))
	}
	var resultMu sync.Mutex
	runPages := func(work []comicPageWorkItem) error {
		return runComicPageWork(ctx, work, pageWorkers, func(generateCtx context.Context, item comicPageWorkItem) error {
			if progress != nil {
				resultMu.Lock()
				progress("comic_images", fmt.Sprintf("正在生成漫画第 %d/%d 张（已完成 %d/%d 张）", item.index+1, totalPages, completedPages, totalPages))
				resultMu.Unlock()
			}
			_, _, generateErr := g.generatePage(generateCtx, owner, batchID, bookID, item.page.ID, batchfactoryv11.UpsertComicPageInput{
				ComicPrompt: item.page.ComicPrompt, ImagePrompt: item.page.ComicImagePrompt, Provider: provider,
				ModelID: modelID, AspectRatio: aspectRatio,
			})
			if generateErr != nil {
				if errors.Is(generateErr, errShuihuoImageContentPolicy) {
					log.Printf("comic page skipped by content policy batch=%s book=%s page=%s err=%v", batchID, bookID, item.page.Label, generateErr)
					if progress != nil {
						progress("comic_images", fmt.Sprintf("%s 两次降级后仍被内容审核拒绝，本次跳过并继续生成其它页", item.page.Label))
					}
					return generateErr
				}
				if comicPageGenerationInFlightError(generateErr) {
					log.Printf("comic page still generating elsewhere batch=%s book=%s page=%s err=%v", batchID, bookID, item.page.Label, generateErr)
					if progress != nil {
						progress("comic_images", fmt.Sprintf("%s 上一次生成仍在跑，本次跳过并继续生成其它页，稍后自动补齐", item.page.Label))
					}
					return generateErr
				}
				return fmt.Errorf("%s 生成失败: %w", item.page.Label, generateErr)
			}
			resultMu.Lock()
			completedPages++
			if progress != nil {
				progress("comic_images", fmt.Sprintf("漫画第 %d/%d 张已完成，继续下一批（已完成 %d/%d 张）", item.index+1, totalPages, completedPages, totalPages))
			}
			resultMu.Unlock()
			return nil
		})
	}
	if err = runPages(work); err != nil && !collectPageFailures(err) {
		return batchfactoryv11.ComicCollection{}, err
	}
	// Re-read after concurrent writes so collection creation sees every page
	// that completed, including pages finishing out of order.
	existingPages, err = comicStore.ListComicPages(ctx, owner, batchID, bookID)
	if err != nil {
		return batchfactoryv11.ComicCollection{}, err
	}
	// 补图轮：只补因在途页锁跳过的页。本轮已经真实失败的页不再原地重试，
	// 它们只记账，等下一轮续跑或达到阈值后进入重试队列。
	missingWork := filterFailedComicPageWork(missingComicPageWorkItems(generationVideos, existingPages), failedPageIDs)
	if len(missingWork) > 0 {
		if progress != nil {
			progress("comic_images", fmt.Sprintf("有 %d 页因在途锁暂未生成，等待后自动补齐缺失页", len(missingWork)))
		}
		// 轮询等待而不是固定睡满：在途页一释放锁就立刻补图，全部补齐则跳过补图轮。
		deadline := time.Now().Add(comicPageRefillDelay)
		for len(missingWork) > 0 {
			if !comicPagesHaveInFlightGeneration(existingPages, generationVideos) {
				break
			}
			if time.Now().After(deadline) {
				break
			}
			timer := time.NewTimer(comicPageRefillPollInterval)
			select {
			case <-ctx.Done():
				timer.Stop()
				return batchfactoryv11.ComicCollection{}, ctx.Err()
			case <-timer.C:
			}
			if existingPages, err = comicStore.ListComicPages(ctx, owner, batchID, bookID); err != nil {
				return batchfactoryv11.ComicCollection{}, err
			}
			missingWork = filterFailedComicPageWork(missingComicPageWorkItems(generationVideos, existingPages), failedPageIDs)
		}
		if len(missingWork) > 0 {
			if err = runPages(missingWork); err != nil && !collectPageFailures(err) {
				return batchfactoryv11.ComicCollection{}, err
			}
			if existingPages, err = comicStore.ListComicPages(ctx, owner, batchID, bookID); err != nil {
				return batchfactoryv11.ComicCollection{}, err
			}
		}
		// 补图轮之后可能仍有页面在别的 worker 手里（例如上一轮 detached
		// 生成还没释放 lease）。先等它们自然结束，再判断是否真的缺图，避免
		// 把“在途锁”误报成内容审核缺图。
		existingPages, err = waitComicPageSettled(ctx, comicStore, owner, batchID, bookID, generationVideos, existingPages)
		if err != nil {
			return batchfactoryv11.ComicCollection{}, err
		}
		missingWork = filterFailedComicPageWork(missingComicPageWorkItems(generationVideos, existingPages), failedPageIDs)
		if len(missingWork) > 0 {
			if err = runPages(missingWork); err != nil && !collectPageFailures(err) {
				return batchfactoryv11.ComicCollection{}, err
			}
			if existingPages, err = comicStore.ListComicPages(ctx, owner, batchID, bookID); err != nil {
				return batchfactoryv11.ComicCollection{}, err
			}
		}
	}
	// 本轮有真实页级失败：返回聚合错误让调用方写 pageLedger，而不是把它
	// 伪装成“缺图成功”或只保留第一条错误。已有成功页时只降级为可续跑的
	// 部分失败，不因为一页把整本判死。
	if err := comicPageGenerationError(failures, completedPages, totalPages); err != nil {
		return batchfactoryv11.ComicCollection{}, err
	}
	if progress != nil {
		progress("comic_images", fmt.Sprintf("图片已全部生成（%d/%d），正在整理漫画合集", completedPages, totalPages))
	}
	requestedVideoIDs := make([]string, 0, len(generationVideos))
	for _, page := range generationVideos {
		requestedVideoIDs = append(requestedVideoIDs, page.ID)
	}
	collection, missing, err := comicStore.CreateComicCollection(ctx, owner, batchID, bookID, requestedVideoIDs)
	if err != nil && len(missing) > 0 {
		if latest, listErr := comicStore.ListComicPages(ctx, owner, batchID, bookID); listErr == nil {
			return batchfactoryv11.ComicCollection{}, classifyComicPageShortage(latest, generationVideos, missing, time.Now().UTC())
		}
		return batchfactoryv11.ComicCollection{}, fmt.Errorf("%w: %d 个漫画页缺少主图（%s）", err, len(missing), strings.Join(missing, ","))
	}
	return collection, err
}

// comicPageShortageError 把“缺主图”拆成三种来源：仍在生成、lease 已过期、
// 真正生成失败/被内容审核拒绝。自动化只对最后一种消耗内容审核重试预算，
// 在途页和过期 lease 不应被写成“两次降级后仍被内容审核拒绝”。
type comicPageShortageError struct {
	Total    int
	InFlight int
	Stale    int
	Failed   int
	Missing  []string
}

func (e *comicPageShortageError) Error() string {
	if e == nil {
		return ""
	}
	parts := make([]string, 0, 3)
	if e.InFlight > 0 {
		parts = append(parts, fmt.Sprintf("%d 页仍在生成", e.InFlight))
	}
	if e.Stale > 0 {
		parts = append(parts, fmt.Sprintf("%d 页生成租约已过期", e.Stale))
	}
	if e.Failed > 0 {
		parts = append(parts, fmt.Sprintf("%d 页生成失败", e.Failed))
	}
	return fmt.Sprintf("%d 个漫画页缺少主图（%s）", e.Total, strings.Join(parts, "，"))
}

func (e *comicPageShortageError) Unwrap() error { return batchfactoryv11.ErrConflict }

func classifyComicPageShortage(pages []batchfactoryv11.ComicPage, videos []batchfactoryv11.Video, missing []string, now time.Time) *comicPageShortageError {
	shortage := &comicPageShortageError{Total: len(missing), Missing: append([]string(nil), missing...)}
	for _, videoID := range missing {
		if !comicVideoInList(videos, videoID) {
			shortage.Failed++
			continue
		}
		var matched *batchfactoryv11.ComicPage
		for index := range pages {
			if pages[index].VideoID == videoID {
				matched = &pages[index]
				break
			}
		}
		switch {
		case matched == nil:
			shortage.Failed++
		case comicPageGenerationTaskIsStale(*matched, now):
			shortage.Stale++
		case comicPageGenerationTaskExists(*matched):
			shortage.InFlight++
		default:
			shortage.Failed++
		}
	}
	return shortage
}

func comicVideoInList(videos []batchfactoryv11.Video, videoID string) bool {
	for _, video := range videos {
		if video.ID == videoID {
			return true
		}
	}
	return false
}

func missingComicAssetImageIDs(ctx context.Context, store batchfactoryv11.Store, owner, batchID string, book batchfactoryv11.Book) ([]string, error) {
	missing := make([]string, 0, len(book.AssetRecords))
	for _, asset := range book.AssetRecords {
		if strings.TrimSpace(asset.Source) != "comic" {
			continue
		}
		images, err := store.ListBookAssetImages(ctx, owner, batchID, book.ID, asset.ID)
		if err != nil {
			return nil, err
		}
		ready := false
		for _, image := range images {
			if image.IsPrimary {
				ready = true
				break
			}
		}
		if !ready {
			missing = append(missing, asset.ID)
		}
	}
	return missing, nil
}

func comicPageHasPrimary(pages []batchfactoryv11.ComicPage, videoID string) bool {
	for _, page := range pages {
		if page.VideoID != videoID {
			continue
		}
		for _, image := range page.Images {
			if image.IsPrimary {
				return true
			}
		}
		return false
	}
	return false
}

// comicPageFinalImagePrompt injects the page's current_page source text verbatim
// into the final image prompt for CM·文字插画版 only. It returns whether the
// block was appended; empty source leaves the prompt untouched so generation
// still proceeds.
func comicPageFinalImagePrompt(prompt string, settings batchfactoryv11.SettingsPatch, page batchfactoryv11.Video) (string, bool) {
	if !batchfactoryv11.ComicPageIllustratedTextPreset(settings) {
		return prompt, false
	}
	if page.ComicSourceSegment == nil || strings.TrimSpace(page.ComicSourceSegment.SourceText) == "" {
		return prompt, false
	}
	return batchfactoryv11.AppendComicPageLiteralText(prompt, page.ComicSourceSegment.SourceText)
}

func comicSettingString(settings batchfactoryv11.SettingsPatch, key, fallback string) string {
	raw, ok := settings[key]
	if !ok {
		return fallback
	}
	var value string
	if json.Unmarshal(raw, &value) != nil || strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func comicSettingInt(settings batchfactoryv11.SettingsPatch, key string, fallback int) int {
	raw, ok := settings[key]
	if !ok {
		return fallback
	}
	var value int
	if json.Unmarshal(raw, &value) != nil || value <= 0 {
		return fallback
	}
	return value
}

func comicSettingBool(settings batchfactoryv11.SettingsPatch, key string, fallback bool) bool {
	raw, ok := settings[key]
	if !ok {
		return fallback
	}
	var value bool
	if json.Unmarshal(raw, &value) != nil {
		return fallback
	}
	return value
}

// comicSettingsToMap 把 SettingsPatch 转成 map[string]any，供上下文传递。
func comicSettingsToMap(settings batchfactoryv11.SettingsPatch) map[string]any {
	if len(settings) == 0 {
		return nil
	}
	result := make(map[string]any, len(settings))
	for key, raw := range settings {
		var value any
		if json.Unmarshal(raw, &value) == nil {
			result[key] = value
		}
	}
	return result
}

func comicPromptPreviewDeveloper(ctx context.Context, owner string, accounts auth.Store, profiles *MemberProfileStore) bool {
	if accounts == nil {
		return false
	}
	account, err := accounts.FindAccount(ctx, owner)
	if err != nil || !account.Active {
		return false
	}
	return toString(authAccountPayload(account, profiles)["role"]) == "dev"
}

func registerComicPageRoutes(mux *http.ServeMux, store batchfactoryv11.Store, files *localartifact.Store, localExecutors *localexecutor.Service, catalogStore novelfetchworkshop.Store, userConfig *UserConfigStore, presetCatalog *prompt.Catalog, accounts auth.Store, profiles *MemberProfileStore, director *batchfactoryv11.DirectorService, generationGate *comicPageGenerationGate, imageCache imagegenerationcache.Repository) {
	registerComicSourceSegmentRoutes(mux, store, director)
	comicStore, ok := store.(batchfactoryv11.ComicPageStore)
	if mux == nil || !ok {
		return
	}
	const pages = "/api/batch-factory/v12/batches/{batchId}/books/{bookId}/comic-pages"
	const page = pages + "/{videoId}"
	const promptZip = pages + "/download-zip"
	const images = pages + "/{pageId}/images"
	const image = images + "/{imageId}"
	const collections = "/api/batch-factory/v12/batches/{batchId}/books/{bookId}/comic-collections"
	const collectionDownload = collections + "/{collectionId}/download"
	generator := comicPageGenerator{store: store, files: files, localExecutors: localExecutors, catalogStore: catalogStore, userConfig: userConfig, presetCatalog: presetCatalog, generationGate: generationGate, imageCache: imageCache, resolveModelOwner: func(owner string) string {
		return effectiveModelCatalogOwner(BridgeIdentity{Username: owner}, profiles, accounts)
	}}

	mux.HandleFunc("GET "+pages, func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		// view=lite 供列表行缩略图与状态渲染使用；提示词正文单本约 300KB，
		// 列表行完全不用，只有编辑弹窗需要全量视图。
		liteView := r.URL.Query().Get("view") == "lite"
		result, err := listComicPagesForView(r.Context(), comicStore, owner, r.PathValue("batchId"), r.PathValue("bookId"), liteView)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		if indexStore, ok := store.(batchfactoryv11.ComicSourceIndexStore); ok {
			// Only the requested book's source text and segment digests are needed
			// for the stale check. Loading the full batch here made every book row
			// refetch the whole batch when the comic library rendered.
			index, found, idxErr := indexStore.GetBookComicSourceIndex(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"))
			if idxErr != nil {
				writeStoreError(w, idxErr)
				return
			}
			for pageIndex := range result {
				result[pageIndex].ConfigStale = comicPageConfigStaleFromIndex(found, index, result[pageIndex])
			}
			readiness := buildComicReadiness(nil, result, comicPageGenerationLimit)
			writeJSON(w, http.StatusOK, map[string]any{
				"pages":     comicPageResponsesLite(comicPageResponses(result), liteView),
				"readiness": readiness,
			})
			return
		}
		batch, batchErr := store.GetBatch(r.Context(), owner, r.PathValue("batchId"))
		if batchErr != nil {
			writeStoreError(w, batchErr)
			return
		}
		for index := range result {
			result[index].ConfigStale = comicPageConfigStale(r.Context(), store, presetCatalog, owner, batch, r.PathValue("bookId"), result[index])
		}
		readiness := buildComicReadiness(nil, result, comicPageGenerationLimit)
		writeJSON(w, http.StatusOK, map[string]any{
			"pages":     comicPageResponsesLite(comicPageResponses(result), liteView),
			"readiness": readiness,
		})
	})
	mux.HandleFunc("GET "+promptZip, func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		plan, err := planComicPromptZip(r.Context(), store, comicStore, files, presetCatalog, owner, r.PathValue("batchId"), r.PathValue("bookId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		serveComicZip(w, files, plan, "comic-prompts.zip")
	})
	mux.HandleFunc("PUT "+page, func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchfactoryv11.PatchComicPageInput
		if !decodeJSON(w, r, &input) {
			return
		}
		var created batchfactoryv11.ComicPage
		var err error
		if patcher, ok := comicStore.(batchfactoryv11.ComicPagePatcher); ok {
			created, err = patcher.PatchComicPage(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("videoId"), input)
		} else {
			// Compatibility fallback for stores that only implement the full-row
			// upsert contract: convert the provided patch fields into a full input.
			full := batchfactoryv11.UpsertComicPageInput{}
			if input.ComicPrompt != nil {
				full.ComicPrompt = *input.ComicPrompt
			}
			if input.ImagePrompt != nil {
				full.ImagePrompt = *input.ImagePrompt
			}
			if input.ReferenceImageURLs != nil {
				full.ReferenceImageURLs = *input.ReferenceImageURLs
			}
			if input.ModelID != nil {
				full.ModelID = *input.ModelID
			}
			if input.AspectRatio != nil {
				full.AspectRatio = *input.AspectRatio
			}
			if input.Status != nil {
				full.Status = *input.Status
			}
			created, err = comicStore.UpsertComicPage(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("videoId"), full)
		}
		if err != nil {
			writeStoreError(w, err)
			return
		}
		// 人工改了页提示词就必须让修复版缓存立即失效，否则旧的自动改写版
		// 会继续压住数据库里的新提示词。
		if input.ComicPrompt != nil || input.ImagePrompt != nil {
			invalidateShuihuoRepairedPrompt(r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("videoId"))
		}
		writeJSON(w, http.StatusOK, map[string]any{"page": comicPageResponse(created)})
	})
	mux.HandleFunc("POST "+page+"/prompt-preview", func(w http.ResponseWriter, r *http.Request) {
		owner, authorized := bridgeOwner(r)
		if !authorized {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if !comicPromptPreviewDeveloper(r.Context(), owner, accounts, profiles) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "developer access required"})
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
		var targetBook *batchfactoryv11.Book
		var targetVideo *batchfactoryv11.Video
		for bookIndex := range batch.Books {
			if batch.Books[bookIndex].ID != r.PathValue("bookId") {
				continue
			}
			targetBook = &batch.Books[bookIndex]
			for videoIndex := range targetBook.Videos {
				if targetBook.Videos[videoIndex].ID == r.PathValue("videoId") {
					targetVideo = &targetBook.Videos[videoIndex]
					break
				}
			}
			break
		}
		if targetBook == nil || targetVideo == nil {
			writeStoreError(w, batchfactoryv11.ErrNotFound)
			return
		}
		if strings.TrimSpace(input.ComicPrompt) == "" {
			input.ComicPrompt = targetVideo.ComicPrompt
		}
		if strings.TrimSpace(input.ImagePrompt) == "" {
			input.ImagePrompt = targetVideo.ComicImagePrompt
		}
		if len(input.ReferenceImageURLs) == 0 {
			existingPages, listErr := comicStore.ListComicPages(r.Context(), owner, batch.ID, targetBook.ID)
			if listErr != nil {
				writeStoreError(w, listErr)
				return
			}
			for _, existingPage := range existingPages {
				if existingPage.VideoID == targetVideo.ID {
					input.ReferenceImageURLs = existingPage.ReferenceImageURLs
					break
				}
			}
		}
		assetMatchContent := comicPagePromptContent(input.ComicPrompt, input.ImagePrompt)
		assetImageURLs, err := comicPageAssetReferenceURLsForPage(r.Context(), store, owner, batch.ID, targetBook.ID, *targetVideo, assetMatchContent)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		effective := batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, targetBook.SettingsState.Patch)
		effective, err = resolvePublishedComicStyleReference(effective, presetCatalog)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		contract, err := batchfactoryv11.BuildComicImageContract(effective, batchfactoryv11.ComicImagePurposePage, "", comicPagePromptContent(input.ComicPrompt, input.ImagePrompt), assetImageURLs, input.ReferenceImageURLs)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"comicPrompt": input.ComicPrompt, "imagePrompt": input.ImagePrompt,
			"compiledPrompt": contract.Prompt, "referenceImageUrls": contract.ReferenceImageURLs,
			"referenceImages": comicReferenceImageResponses(effective, *targetBook, contract.ReferenceImageURLs, input.ReferenceImageURLs),
			"stylePresetId":   contract.StylePresetID, "snapshotHash": contract.SnapshotHash,
		})
	})
	mux.HandleFunc("PUT "+page+"/reference-images", func(w http.ResponseWriter, r *http.Request) {
		owner, authorized := bridgeOwner(r)
		if !authorized {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			ReferenceImageURLs []string `json:"referenceImageUrls"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		updated, err := comicStore.SetComicPageReferenceImages(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("videoId"), input.ReferenceImageURLs)
		if errors.Is(err, batchfactoryv11.ErrNotFound) {
			batch, batchErr := store.GetBatch(r.Context(), owner, r.PathValue("batchId"))
			if batchErr != nil {
				writeStoreError(w, batchErr)
				return
			}
			var targetBook *batchfactoryv11.Book
			var targetVideo *batchfactoryv11.Video
			for bookIndex := range batch.Books {
				if batch.Books[bookIndex].ID != r.PathValue("bookId") {
					continue
				}
				targetBook = &batch.Books[bookIndex]
				for videoIndex := range targetBook.Videos {
					if targetBook.Videos[videoIndex].ID == r.PathValue("videoId") {
						targetVideo = &targetBook.Videos[videoIndex]
						break
					}
				}
				break
			}
			if targetBook == nil || targetVideo == nil || strings.TrimSpace(targetVideo.ComicPrompt) == "" || strings.TrimSpace(targetVideo.ComicImagePrompt) == "" {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "请先提取当前分镜的漫画提示词"})
				return
			}
			effective := batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, targetBook.SettingsState.Patch)
			_, err = comicStore.UpsertComicPage(r.Context(), owner, batch.ID, targetBook.ID, targetVideo.ID, batchfactoryv11.UpsertComicPageInput{
				ComicPrompt: targetVideo.ComicPrompt, ImagePrompt: targetVideo.ComicImagePrompt,
				ModelID: comicSettingString(effective, "imageModelId", ""), AspectRatio: batchfactoryv11.EffectiveImageAspectRatio(effective),
			})
			if err == nil {
				invalidateShuihuoRepairedPrompt(batch.ID, targetBook.ID, targetVideo.ID)
				updated, err = comicStore.SetComicPageReferenceImages(r.Context(), owner, batch.ID, targetBook.ID, targetVideo.ID, input.ReferenceImageURLs)
			}
		}
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"page": comicPageResponse(updated)})
	})
	mux.HandleFunc("POST "+page+"/reference-images/apply-to-all", func(w http.ResponseWriter, r *http.Request) {
		owner, authorized := bridgeOwner(r)
		if !authorized {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		batch, err := store.GetBatch(r.Context(), owner, r.PathValue("batchId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		var book *batchfactoryv11.Book
		for index := range batch.Books {
			if batch.Books[index].ID == r.PathValue("bookId") {
				book = &batch.Books[index]
				break
			}
		}
		if book == nil {
			writeStoreError(w, batchfactoryv11.ErrNotFound)
			return
		}
		pages, err := comicStore.ListComicPages(r.Context(), owner, batch.ID, book.ID)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		existing := make(map[string]bool, len(pages))
		for _, existingPage := range pages {
			existing[existingPage.VideoID] = true
		}
		effective := batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
		for _, video := range book.Videos {
			if existing[video.ID] {
				continue
			}
			if _, err := comicStore.UpsertComicPage(r.Context(), owner, batch.ID, book.ID, video.ID, batchfactoryv11.UpsertComicPageInput{
				ComicPrompt: video.ComicPrompt, ImagePrompt: video.ComicImagePrompt,
				ModelID: comicSettingString(effective, "imageModelId", ""), AspectRatio: batchfactoryv11.EffectiveImageAspectRatio(effective),
			}); err != nil {
				writeStoreError(w, err)
				return
			}
			invalidateShuihuoRepairedPrompt(batch.ID, book.ID, video.ID)
		}
		updated, err := comicStore.ApplyComicPageReferenceImages(r.Context(), owner, batch.ID, book.ID, r.PathValue("videoId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"updated": updated})
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
		batchID, bookID, videoID := r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("videoId")
		type generationResult struct {
			page  batchfactoryv11.ComicPage
			image batchfactoryv11.ComicPageImage
			err   error
		}
		resultCh := make(chan generationResult, 1)
		go func() {
			// Do not inherit the browser request context. A browser refresh or
			// disconnect must not discard a provider result that can still be
			// written back to the comic page.
			generationCtx, cancel := comicPageGenerationContext(context.Background())
			defer cancel()
			generationCtx = withShuihuoImageGenerationCacheBypass(withShuihuoImageGenerationCache(generationCtx, imageCache))
			createdPage, createdImage, err := generator.generatePage(generationCtx, owner, batchID, bookID, videoID, input)
			resultCh <- generationResult{page: createdPage, image: createdImage, err: err}
			if err != nil {
				log.Printf("comic page background generation failed batch=%s book=%s video=%s error=%v", batchID, bookID, videoID, err)
			}
		}()
		select {
		case result := <-resultCh:
			if result.err != nil {
				writeStoreError(w, result.err)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]any{"page": comicPageResponse(result.page), "image": comicPageImageResponse(result.image)})
		case <-time.After(comicPageHTTPFastPathWait):
			writeJSON(w, http.StatusAccepted, map[string]any{
				"status":  "generating",
				"batchId": batchID,
				"bookId":  bookID,
				"videoId": videoID,
				"message": "漫画页已提交，生成完成后会自动回填。",
			})
		}
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
				// 页图只在登录态下可见：只允许浏览器私有缓存，不引入共享缓存。
				// 7 天与缩略图一致；同 ref 原地重新生成会改变 size/mtime，ETag 随之失效。
				if info, statErr := file.Stat(); statErr == nil {
					w.Header().Set("ETag", fmt.Sprintf("\"%x-%x\"", info.Size(), info.ModTime().UTC().UnixNano()))
				}
				w.Header().Set("Cache-Control", "private, max-age=604800")
				// ServeContent 继续负责 Accept-Ranges、Content-Length、Range→206
				// 与 If-None-Match/If-Modified-Since→304。
				http.ServeContent(w, r, candidate.ID, candidate.CreatedAt, file)
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "image not found"})
	})
	mux.HandleFunc("GET "+image+"/thumbnail", handleComicPageThumbnail(comicStore, files))
	mux.HandleFunc("POST "+collections, func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var payload struct {
			VideoIDs []string `json:"videoIds"`
		}
		if r.Body != nil {
			if !limitRequestBody(w, r, maxJSONBodyBytes) {
				return
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
				if isBodyTooLarge(err) {
					writePayloadTooLarge(w, maxJSONBodyBytes)
					return
				}
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "合集页码选择格式不正确"})
				return
			}
		}
		missingPrompts, stalePages, preflightErr := comicCollectionPreflight(r.Context(), store, comicStore, presetCatalog, owner, r.PathValue("batchId"), r.PathValue("bookId"), payload.VideoIDs)
		if preflightErr != nil {
			writeStoreError(w, preflightErr)
			return
		}
		if len(missingPrompts) > 0 {
			writeJSON(w, http.StatusConflict, map[string]any{"error": "每个漫画页必须先完成分页和分镜提示词", "missingVideoIds": missingPrompts})
			return
		}
		if len(stalePages) > 0 {
			writeJSON(w, http.StatusConflict, map[string]any{"error": "漫画页配置已变化，请先重新生成对应漫画页", "staleVideoIds": stalePages})
			return
		}
		created, missing, err := comicStore.CreateComicCollection(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), payload.VideoIDs)
		if errors.Is(err, batchfactoryv11.ErrConflict) {
			writeJSON(w, http.StatusConflict, map[string]any{"error": "所选每个分镜都必须先选好主图，且页码必须来自当前书籍", "missingVideoIds": missing})
			return
		}
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"collection": comicCollectionResponse(created)})
	})
	mux.HandleFunc("POST "+collections+"/generate", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if files == nil || catalogStore == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "漫画图片生成服务尚未配置"})
			return
		}
		collection, err := generator.generateBook(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"collection": comicCollectionResponse(collection)})
	})
	mux.HandleFunc("GET "+collectionDownload, func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		plan, err := planComicCollectionImageZip(r.Context(), store, comicStore, files, owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("collectionId"))
		if err != nil {
			if errors.Is(err, batchfactoryv11.ErrNotFound) && strings.Contains(err.Error(), "第 ") {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
				return
			}
			writeStoreError(w, err)
			return
		}
		serveComicZip(w, files, plan, "comic-collection.zip")
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

// planComicCollectionImageZip 先解析合集并逐页校验主图文件可打开，再返回
// 流式打包计划。缺图仍返回「第 N 页…」的可读错误，绝不会写半截 ZIP。
func planComicCollectionImageZip(ctx context.Context, store batchfactoryv11.Store, comicStore batchfactoryv11.ComicPageStore, files *localartifact.Store, owner, batchID, bookID, collectionID string) (comicZipPlan, error) {
	if files == nil {
		return comicZipPlan{}, fmt.Errorf("%w: 漫画图片存储不可用", batchfactoryv11.ErrUnavailable)
	}
	batch, err := store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return comicZipPlan{}, err
	}
	var title string
	for _, book := range batch.Books {
		if book.ID == bookID {
			title = book.Title
			break
		}
	}
	if strings.TrimSpace(title) == "" {
		return comicZipPlan{}, batchfactoryv11.ErrNotFound
	}
	collections, err := comicStore.ListComicCollections(ctx, owner, batchID, bookID)
	if err != nil {
		return comicZipPlan{}, err
	}
	var collection *batchfactoryv11.ComicCollection
	for index := range collections {
		if collections[index].ID == collectionID {
			collection = &collections[index]
			break
		}
	}
	if collection == nil {
		return comicZipPlan{}, batchfactoryv11.ErrNotFound
	}
	if len(collection.Items) == 0 {
		return comicZipPlan{}, fmt.Errorf("%w: 漫画合集没有可下载的图片", batchfactoryv11.ErrConflict)
	}

	entries := make([]comicZipEntry, 0, len(collection.Items))
	for _, item := range collection.Items {
		pageNumber := item.Ordinal
		if pageNumber < 1 {
			pageNumber = 1
		}
		if strings.TrimSpace(item.StorageRef) == "" {
			return comicZipPlan{}, fmt.Errorf("%w: 第 %d 页主图不可下载", batchfactoryv11.ErrNotFound, pageNumber)
		}
		file, openErr := files.Open(item.StorageRef)
		if openErr != nil {
			return comicZipPlan{}, fmt.Errorf("%w: 第 %d 页主图文件不存在", batchfactoryv11.ErrNotFound, pageNumber)
		}
		_ = file.Close()
		entries = append(entries, comicZipEntry{
			name:       fmt.Sprintf("%s-%03d%s", comicDownloadFileStem(title), pageNumber, comicDownloadImageExtension(item.MediaType, item.StorageRef)),
			storageRef: item.StorageRef,
		})
	}
	return comicZipPlan{filename: comicDownloadFileStem(title) + ".zip", entries: entries}, nil
}

func comicDownloadFileStem(title string) string {
	clean := strings.TrimSpace(title)
	clean = strings.NewReplacer("\\", "_", "/", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_").Replace(clean)
	clean = strings.Trim(clean, " .")
	if clean == "" {
		return "漫画合集"
	}
	return clean
}

func comicDownloadImageExtension(mediaType, storageRef string) string {
	switch strings.ToLower(strings.TrimSpace(mediaType)) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "image/png":
		return ".png"
	}
	if extension := strings.ToLower(filepath.Ext(storageRef)); extension == ".png" || extension == ".jpg" || extension == ".jpeg" || extension == ".webp" {
		return extension
	}
	return ".png"
}

func comicCollectionPreflight(ctx context.Context, store batchfactoryv11.Store, comicStore batchfactoryv11.ComicPageStore, presetCatalog *prompt.Catalog, owner, batchID, bookID string, requestedVideoIDs []string) ([]string, []string, error) {
	batch, err := store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return nil, nil, err
	}
	var book *batchfactoryv11.Book
	for index := range batch.Books {
		if batch.Books[index].ID == bookID {
			book = &batch.Books[index]
			break
		}
	}
	if book == nil {
		return nil, nil, batchfactoryv11.ErrNotFound
	}
	selectedVideoIDs, unknownVideoIDs := batchfactoryv11.SelectComicCollectionVideoIDs(book.Videos, requestedVideoIDs)
	if len(unknownVideoIDs) > 0 || len(selectedVideoIDs) == 0 {
		return unknownVideoIDs, nil, batchfactoryv11.ErrConflict
	}
	selected := make(map[string]struct{}, len(selectedVideoIDs))
	for _, videoID := range selectedVideoIDs {
		selected[videoID] = struct{}{}
	}
	pages, err := comicStore.ListComicPages(ctx, owner, batchID, bookID)
	if err != nil {
		return nil, nil, err
	}
	byVideo := make(map[string]batchfactoryv11.ComicPage, len(pages))
	for _, page := range pages {
		byVideo[page.VideoID] = page
	}
	missingPrompts := make([]string, 0)
	stalePages := make([]string, 0)
	for _, video := range book.Videos {
		if _, ok := selected[video.ID]; !ok {
			continue
		}
		page, found := byVideo[video.ID]
		if !found || strings.TrimSpace(page.ComicPrompt) == "" || strings.TrimSpace(page.ImagePrompt) == "" {
			missingPrompts = append(missingPrompts, video.ID)
			continue
		}
		if comicPageConfigStale(ctx, store, presetCatalog, owner, batch, bookID, page) {
			stalePages = append(stalePages, video.ID)
		}
	}
	return missingPrompts, stalePages, nil
}

func comicPageAssetReferenceURLs(ctx context.Context, store batchfactoryv11.Store, owner, batchID, bookID, promptText string) ([]string, error) {
	assets, err := store.ListBookAssets(ctx, owner, batchID, bookID)
	if err != nil {
		return nil, err
	}
	promptText = strings.TrimSpace(promptText)
	references := make([]string, 0)
	for _, asset := range assets {
		if strings.TrimSpace(asset.Kind) != "character" {
			continue
		}
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

// comicPageAssetReferenceURLsForPage resolves references from the page's
// structured bindings only. The generated image prompt is deliberately not
// used for name matching because ordinary words in prose can look like asset
// names (for example, the first-person pronoun "我").
func comicPageAssetReferenceURLsForPage(ctx context.Context, store batchfactoryv11.Store, owner, batchID, bookID string, page batchfactoryv11.Video, imagePrompt string) ([]string, error) {
	_ = imagePrompt
	return comicPageAssetReferenceURLs(ctx, store, owner, batchID, bookID, comicPageAssetLookupText(page, ""))
}

func comicPromptMentionsAsset(promptText string, asset batchfactoryv11.BookAsset) bool {
	for _, name := range []string{asset.Name, asset.FormalName} {
		if name = strings.TrimSpace(name); name != "" && strings.Contains(promptText, name) {
			return true
		}
	}
	return false
}

func comicPageConfigStale(ctx context.Context, store batchfactoryv11.Store, presetCatalog *prompt.Catalog, owner string, batch batchfactoryv11.Batch, bookID string, page batchfactoryv11.ComicPage) bool {
	// 页面提示词和图片属于生产结果；统一配置（画风、模型、参考图）
	// 的变化不再把已有主图判成过期。只有页面绑定的原文分段明确不属于
	// 当前正文时，才提示用户先重新分析分页和提示词。
	_ = ctx
	_ = store
	_ = presetCatalog
	_ = owner
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
	var video *batchfactoryv11.Video
	for index := range book.Videos {
		if book.Videos[index].ID == page.VideoID {
			video = &book.Videos[index]
			break
		}
	}
	if video == nil {
		return true
	}
	if video.ComicSourceSegment == nil || strings.TrimSpace(video.ComicSourceSegment.SourceDigest) == "" {
		return false
	}
	digest := sha256.Sum256([]byte(book.SourceText))
	return video.ComicSourceSegment.SourceDigest != fmt.Sprintf("%x", digest)
}

// comicPageConfigStaleFromIndex mirrors comicPageConfigStale using the
// lightweight per-book source index instead of a hydrated batch.
func comicPageConfigStaleFromIndex(bookFound bool, index batchfactoryv11.ComicSourceIndex, page batchfactoryv11.ComicPage) bool {
	if !bookFound {
		return false
	}
	digest, ok := index.VideoDigests[page.VideoID]
	if !ok {
		return true
	}
	if strings.TrimSpace(digest) == "" {
		return false
	}
	sum := sha256.Sum256([]byte(index.SourceText))
	return digest != fmt.Sprintf("%x", sum)
}

func comicPageImageResponse(image batchfactoryv11.ComicPageImage) batchfactoryv11.ComicPageImage {
	if image.StorageRef != "" {
		image.URL = "/api/batch-factory/v12/batches/" + image.BatchID + "/books/" + image.BookID + "/comic-pages/" + image.ComicPageID + "/images/" + image.ID + "/content"
		image.ThumbnailURL = comicThumbnailURL(image)
	}
	return image
}

// comicZipEntry 是一条待写入 ZIP 的图片条目。plan 阶段只校验文件可打开，
// 真正的字节复制留到响应体流式写出，避免整包驻留内存。
type comicZipEntry struct {
	name       string
	storageRef string
}

type comicZipPlan struct {
	filename string
	markdown string
	entries  []comicZipEntry
}

// planComicPromptZip 只做数据库读取、提示词拼装和条目收集，不生成 ZIP 字节。
// 单本漫画 35 页原图约 150MB，必须让压缩发生在响应流里而不是内存里。
func planComicPromptZip(ctx context.Context, store batchfactoryv11.Store, comicStore batchfactoryv11.ComicPageStore, files *localartifact.Store, presetCatalog *prompt.Catalog, owner, batchID, bookID string) (comicZipPlan, error) {
	batch, err := store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return comicZipPlan{}, err
	}
	var book *batchfactoryv11.Book
	for index := range batch.Books {
		if batch.Books[index].ID == bookID {
			book = &batch.Books[index]
			break
		}
	}
	if book == nil {
		return comicZipPlan{}, batchfactoryv11.ErrNotFound
	}
	pages, err := comicStore.ListComicPages(ctx, owner, batchID, bookID)
	if err != nil {
		return comicZipPlan{}, err
	}
	pagesByVideo := make(map[string]batchfactoryv11.ComicPage, len(pages))
	for _, page := range pages {
		pagesByVideo[page.VideoID] = page
	}
	effective := batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
	effective, err = resolvePublishedComicStyleReference(effective, presetCatalog)
	if err != nil {
		return comicZipPlan{}, err
	}

	var promptMarkdown strings.Builder
	fmt.Fprintf(&promptMarkdown, "# %s｜最终完整出图提示词\n\n", book.Title)
	for index, video := range book.Videos {
		current, hasPage := pagesByVideo[video.ID]
		directorPrompt := strings.TrimSpace(video.ComicPrompt)
		imagePrompt := strings.TrimSpace(video.ComicImagePrompt)
		if hasPage {
			if strings.TrimSpace(current.ComicPrompt) != "" {
				directorPrompt = strings.TrimSpace(current.ComicPrompt)
			}
			if strings.TrimSpace(current.ImagePrompt) != "" {
				imagePrompt = strings.TrimSpace(current.ImagePrompt)
			}
		}
		compiled := ""
		if imagePrompt == "" {
			compiled = "当前页面没有基础出图描述。"
		} else {
			references := []string(nil)
			if hasPage {
				references, _ = comicPageAssetReferenceURLsForPage(ctx, store, owner, batchID, bookID, video, imagePrompt)
			}
			contract, contractErr := batchfactoryv11.BuildComicImageContract(effective, batchfactoryv11.ComicImagePurposePage, "", comicPagePromptContent(directorPrompt, imagePrompt), references, current.ReferenceImageURLs)
			if contractErr != nil {
				compiled = "完整提示词暂不可生成：" + contractErr.Error()
			} else {
				compiled = contract.Prompt
			}
		}
		label := safeComicZipPart(video.Label)
		if label == "" {
			label = "漫画页"
		}
		fmt.Fprintf(&promptMarkdown, "## 第 %d 页｜%s\n\n", index+1, label)
		promptMarkdown.WriteString("### 最终完整出图提示词\n\n")
		promptMarkdown.WriteString(compiled)
		promptMarkdown.WriteString("\n\n")
	}

	entries := make([]comicZipEntry, 0, len(pages))
	appendEntry := func(name, storageRef string) {
		if files == nil || strings.TrimSpace(storageRef) == "" {
			return
		}
		// 只校验可打开；读不到的图片沿用原有“跳过”语义，不让整包失败。
		file, openErr := files.Open(storageRef)
		if openErr != nil {
			return
		}
		_ = file.Close()
		entries = append(entries, comicZipEntry{name: name, storageRef: storageRef})
	}

	// 复用同一个打包接口，把已生成的漫画页主图一并放入压缩包；没有图片时
	// 仍然只导出提示词与资产图片，保持原有调用方的行为。
	for pageIndex, page := range pages {
		for imageIndex, image := range page.Images {
			extension := filepath.Ext(image.StorageRef)
			if extension == "" {
				extension = comicZipImageExtension(image.MediaType)
			}
			appendEntry(fmt.Sprintf("漫画页/第%02d页_%02d%s", pageIndex+1, imageIndex+1, extension), image.StorageRef)
		}
	}

	assets, err := store.ListBookAssets(ctx, owner, batchID, bookID)
	if err != nil {
		return comicZipPlan{}, err
	}
	for assetIndex, asset := range assets {
		images, listErr := store.ListBookAssetImages(ctx, owner, batchID, bookID, asset.ID)
		if listErr != nil {
			return comicZipPlan{}, listErr
		}
		kind := comicZipAssetKindLabel(asset.Kind)
		assetName := safeComicZipPart(asset.Name)
		if assetName == "" {
			assetName = fmt.Sprintf("资产_%03d", assetIndex+1)
		}
		for imageIndex, image := range images {
			extension := filepath.Ext(image.StorageRef)
			if extension == "" {
				extension = comicZipImageExtension(image.MediaType)
			}
			appendEntry(fmt.Sprintf("资产图片/%s/%s_%02d%s", kind, assetName, imageIndex+1, extension), image.StorageRef)
		}
	}

	filename := safeComicZipPart(book.Title)
	if filename == "" {
		filename = "漫画提示词"
	}
	return comicZipPlan{filename: filename + "_最终提示词_资产图片.zip", markdown: promptMarkdown.String(), entries: entries}, nil
}

// writeComicZipStream 边压边写：zip.NewWriter 直接落到响应 writer，条目按
// 顺序 io.Copy，内存占用只跟单个读缓冲有关。
func writeComicZipStream(w io.Writer, files *localartifact.Store, plan comicZipPlan) error {
	writer := zip.NewWriter(w)
	if plan.markdown != "" {
		if err := writeComicZipText(writer, "最终完整出图提示词.md", plan.markdown); err != nil {
			_ = writer.Close()
			return err
		}
	}
	for _, entry := range plan.entries {
		file, openErr := files.Open(entry.storageRef)
		if openErr != nil {
			_ = writer.Close()
			return openErr
		}
		target, createErr := writer.Create(entry.name)
		if createErr == nil {
			_, createErr = io.Copy(target, file)
		}
		_ = file.Close()
		if createErr != nil {
			_ = writer.Close()
			return createErr
		}
	}
	return writer.Close()
}

// serveComicZip 在写出任何 ZIP 字节前先发响应头；plan 阶段已经拦掉了可读
// 错误（缺图、权限、不存在），所以这里出现 I/O 错误只可能是流中断。此时
// 状态码已经发出，主动断开连接让客户端明确看到下载失败，而不是收到一个
// 看似成功但内容截断的 ZIP。
func serveComicZip(w http.ResponseWriter, files *localartifact.Store, plan comicZipPlan, asciiName string) {
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+asciiName+"\"; filename*=UTF-8''"+url.QueryEscape(plan.filename))
	w.WriteHeader(http.StatusOK)
	if err := writeComicZipStream(w, files, plan); err != nil {
		log.Printf("stream comic zip %s: %v", plan.filename, err)
		panic(http.ErrAbortHandler)
	}
}

func comicZipAssetKindLabel(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "character", "person", "人物":
		return "人物"
	case "scene", "场景":
		return "场景"
	case "prop", "object", "道具":
		return "道具"
	default:
		return "其他资产"
	}
}

func writeComicZipText(writer *zip.Writer, name, content string) error {
	entry, err := writer.Create(name)
	if err != nil {
		return err
	}
	_, err = io.WriteString(entry, content)
	return err
}

func safeComicZipPart(value string) string {
	value = strings.TrimSpace(value)
	value = strings.NewReplacer("\\", "_", "/", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_").Replace(value)
	return strings.Trim(value, " .")
}

func comicZipImageExtension(mediaType string) string {
	switch strings.ToLower(strings.TrimSpace(mediaType)) {
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	default:
		return ".png"
	}
}

func comicPageResponse(page batchfactoryv11.ComicPage) batchfactoryv11.ComicPage {
	for index := range page.Images {
		page.Images[index].BatchID = page.BatchID
		page.Images[index].BookID = page.BookID
		page.Images[index] = comicPageImageResponse(page.Images[index])
	}
	hasPrimary, hasAny := comicPageImageFlags(page)
	page.PromptStatus, page.ImageStatus, page.GenerationState, page.PrimaryFallback = deriveComicPageStatus(page, hasPrimary, hasAny)
	return page
}

func comicPageResponses(pages []batchfactoryv11.ComicPage) []batchfactoryv11.ComicPage {
	for index := range pages {
		pages[index] = comicPageResponse(pages[index])
	}
	return pages
}

// comicPageResponsesLite 剥离列表行不需要的提示词正文与参考图。派生状态已经
// 在 comicPageResponses 阶段算好，这里只做输出裁剪，避免影响 readiness 统计。
func comicPageResponsesLite(pages []batchfactoryv11.ComicPage, lite bool) []batchfactoryv11.ComicPage {
	if !lite {
		return pages
	}
	for index := range pages {
		pages[index].ComicPrompt = ""
		pages[index].ImagePrompt = ""
		pages[index].ReferenceImageURLs = nil
	}
	return pages
}

func comicCollectionResponse(collection batchfactoryv11.ComicCollection) batchfactoryv11.ComicCollection {
	for index := range collection.Items {
		item := &collection.Items[index]
		base := "/api/batch-factory/v12/batches/" + collection.BatchID + "/books/" + collection.BookID + "/comic-pages/" + item.ComicPageID + "/images/" + item.ImageID
		item.URL = base + "/content"
		// 合集九宫格同样先取缩略图；点开大图/下载仍走 URL 原图。
		item.ThumbnailURL = base + "/thumbnail"
	}
	return collection
}
