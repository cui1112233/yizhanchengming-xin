package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/imagegenerationcache"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
	"qiantie/backend/internal/mergeworker"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/safehttp"
	"qiantie/backend/internal/watermark"
)

func patchShuihuoModel(current, patch map[string]any) (map[string]any, error) {
	updated := make(map[string]any, len(current)+len(patch))
	for key, value := range current {
		updated[key] = value
	}
	for _, key := range []string{"name", "kind", "adapterKind", "enabled", "parameterSchema", "credentialRef", "endpoint", "tasksEndpoint", "resultEndpoint", "workflowId", "requestTemplate", "responseMapping", "voiceCloneRequestTemplate", "voiceCloneResponseMapping", "capabilities", "credential", "apiKey"} {
		if value, exists := patch[key]; exists {
			updated[key] = value
		}
	}
	name := strings.TrimSpace(modelString(updated, "name"))
	if name == "" {
		name = strings.TrimSpace(modelString(updated, "displayName"))
	}
	if name == "" || strings.TrimSpace(modelString(updated, "kind")) == "" {
		return nil, errors.New("模型名称和能力不能为空")
	}
	return updated, nil
}

var shuihuoProductionConfigs sync.Map

// 并发边界：单本默认 4 路、最多 4 路。
// ggwk 网关的长响应对并发连接敏感，擅自放大到 4/8 路会增加断流和
// PROTOCOL_ERROR，实际反而更慢。统一配置和环境变量仍可在安全范围内调整。
const shuihuoImageConcurrencyEnv = "QIANTIE_SHUHUO_IMAGE_CONCURRENCY"
// 默认给到上限 4 路：按本串行出图后，一本书独占全部通道，4 路能把
// 「30 页 × 单张 150~200s」从约 28 分钟压到约 8 分钟；用户仍可在
// 统一配置里按通道情况下调到 1/2/3 路。
const shuihuoImageConcurrencyDefault = 4
const shuihuoImageConcurrencyMax = 4

func shuihuoClampImageConcurrency(value int) int {
	if value < 1 {
		return 1
	}
	if value > shuihuoImageConcurrencyMax {
		return shuihuoImageConcurrencyMax
	}
	return value
}

func shuihuoImageGenerationConcurrency() int {
	value := strings.TrimSpace(os.Getenv(shuihuoImageConcurrencyEnv))
	if value == "" {
		return shuihuoImageConcurrencyDefault
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return shuihuoImageConcurrencyDefault
	}
	return shuihuoClampImageConcurrency(parsed)
}

// shuihuoImageSemaphore 是可在运行时调整上限的信号量。统一配置改变后，
// 后续请求立即按新上限排队，不影响已经在跑的图片请求。
type shuihuoImageSemaphore struct {
	mu      sync.Mutex
	limit   int
	inUse   int
	waiters []*shuihuoImageSemaphoreWaiter
}

type shuihuoImageSemaphoreWaiter struct {
	ready chan struct{}
}

func newShuihuoImageSemaphore(limit int) *shuihuoImageSemaphore {
	return &shuihuoImageSemaphore{limit: shuihuoClampImageConcurrency(limit)}
}

// newShuihuoImageSemaphoreRaw 不做每书 4 路上限的裁剪，专供任务级总并发使用。
func newShuihuoImageSemaphoreRaw(limit int) *shuihuoImageSemaphore {
	if limit < 1 {
		limit = 1
	}
	return &shuihuoImageSemaphore{limit: limit}
}

func (s *shuihuoImageSemaphore) setLimit(limit int) {
	limit = shuihuoClampImageConcurrency(limit)
	s.mu.Lock()
	s.limit = limit
	s.wakeWaitersLocked()
	s.mu.Unlock()
}

func (s *shuihuoImageSemaphore) acquire(ctx context.Context) (func(), error) {
	s.mu.Lock()
	if s.inUse < s.limit && len(s.waiters) == 0 {
		s.inUse++
		s.mu.Unlock()
		return s.releaseFunc(), nil
	}
	waiter := &shuihuoImageSemaphoreWaiter{ready: make(chan struct{})}
	s.waiters = append(s.waiters, waiter)
	s.mu.Unlock()

	select {
	case <-waiter.ready:
		return s.releaseFunc(), nil
	case <-ctx.Done():
		s.mu.Lock()
		for index, candidate := range s.waiters {
			if candidate != waiter {
				continue
			}
			s.waiters = append(s.waiters[:index], s.waiters[index+1:]...)
			s.mu.Unlock()
			return nil, ctx.Err()
		}
		// 名额与取消同时发生：wake 已经把本请求计入 inUse。把名额
		// 立即还给下一位，不能因为调用方已超时而永久占住信号量。
		s.inUse--
		s.wakeWaitersLocked()
		s.mu.Unlock()
		return nil, ctx.Err()
	}
}

func (s *shuihuoImageSemaphore) releaseFunc() func() {
	var once sync.Once
	return func() {
		once.Do(func() {
			s.mu.Lock()
			s.inUse--
			s.wakeWaitersLocked()
			s.mu.Unlock()
		})
	}
}

func (s *shuihuoImageSemaphore) wakeWaitersLocked() {
	for s.inUse < s.limit && len(s.waiters) > 0 {
		waiter := s.waiters[0]
		s.waiters = s.waiters[1:]
		s.inUse++
		close(waiter.ready)
	}
}

var shuihuoImageGenerationSemaphore = newShuihuoImageSemaphore(shuihuoImageGenerationConcurrency())

// 图片网关可能很快返回 200 响应头，再用数分钟传输 base64 正文。生图名额
// 只管“请求上游生成”，正文传输由独立槽位限制；这样慢正文不会把后续生图
// 全堵死，同时最多保留 4 条大响应，避免无限堆连接和内存。
var shuihuoImageResponseTransferSemaphore = newShuihuoImageSemaphoreRaw(shuihuoImageConcurrencyMax)

func beginShuihuoImageResponseTransfer(ctx context.Context, releaseGeneration func()) (func(), error) {
	releaseTransfer, err := shuihuoImageResponseTransferSemaphore.acquire(ctx)
	if err != nil {
		if releaseGeneration != nil {
			releaseGeneration()
		}
		return nil, err
	}
	if releaseGeneration != nil {
		releaseGeneration()
	}
	return releaseTransfer, nil
}

// shuihuoProcessImageGenerationSemaphore 是整台 Go 服务共用的图片请求总闸门。
// 任务上下文里的“全局并发”只约束一个自动化任务；若同时跑两个批次，两个
// 任务级上限会叠加。这里再收口一次，避免并发大响应把上游连接压断。
// 整机上限必须和单批次安全上限一致；否则两个批次各跑 4 路时，仍会
// 叠加成网关已知不稳定的 8 路，表现为所有请求一起迟迟不返回。
var shuihuoProcessImageGenerationSemaphore = newShuihuoImageSemaphoreRaw(shuihuoImageConcurrencyMax)

type shuihuoImageConcurrencyContextKey struct{}

type shuihuoImageSemaphoreContextKey struct{}

type shuihuoGlobalImageSemaphoreContextKey struct{}

// shuihuoGlobalImageSemaphores 是任务级出图总名额。
//
// 旧实现把总名额固定切成“资产图一半 + 页图一半”。实测一轮自动化里同时并行的
// 书几乎总处在同一阶段：资产图阶段只能用满 4 路，另外 4 路页图名额一直空转；
// 页图阶段反过来空转资产名额。等于 8 路总并发实际只跑出 4 路产能。
//
// 现在改为“总名额共享 + 阶段保底”：另一阶段没有排队请求时，本阶段可以用满全部
// 名额；另一阶段一旦有等待中的请求，就为它保留 reserve 个名额，避免资产图把
// 页图饿住、或页图把资产图饿住。总量始终不超过 limit，保持对网关 RST 的收敛。
type shuihuoGlobalImageSemaphores struct {
	mu      sync.Mutex
	limit   int
	reserve int
	inUse   map[string]int
	waiting map[string]int
}

const (
	shuihuoImageStageAsset = "asset"
	shuihuoImageStagePage  = "page"
)

func newShuihuoGlobalImageSemaphores(limit int) *shuihuoGlobalImageSemaphores {
	if limit < 2 {
		limit = 2
	}
	reserve := limit / 2
	if reserve < 1 {
		reserve = 1
	}
	return &shuihuoGlobalImageSemaphores{
		limit:   limit,
		reserve: reserve,
		inUse:   make(map[string]int, 2),
		waiting: make(map[string]int, 2),
	}
}

func shuihuoOtherImageStage(stage string) string {
	if stage == shuihuoImageStageAsset {
		return shuihuoImageStagePage
	}
	return shuihuoImageStageAsset
}

// acquire 按阶段领取一个名额。capFor 只限制“本阶段还能再拿几个”：
//   - 另一阶段没有排队时，capFor = limit - 另一阶段在用量，即可以用满剩余名额；
//   - 另一阶段有排队且用不满保底名额时，先给它留出 reserve，本阶段最多用
//     limit - reserve。
func (g *shuihuoGlobalImageSemaphores) acquire(ctx context.Context, stage string) (func(), error) {
	if stage != shuihuoImageStageAsset {
		stage = shuihuoImageStagePage
	}
	other := shuihuoOtherImageStage(stage)
	g.mu.Lock()
	g.waiting[stage]++
	g.mu.Unlock()
	acquired := false
	defer func() {
		if !acquired {
			g.mu.Lock()
			g.waiting[stage]--
			g.mu.Unlock()
		}
	}()
	for {
		g.mu.Lock()
		capFor := g.limit - g.inUse[other]
		if g.waiting[other] > 0 && g.inUse[other] < g.reserve {
			capFor = g.limit - g.reserve
		}
		if g.inUse[stage] < capFor {
			g.inUse[stage]++
			g.waiting[stage]--
			g.mu.Unlock()
			acquired = true
			var once sync.Once
			return func() {
				once.Do(func() {
					g.mu.Lock()
					g.inUse[stage]--
					g.mu.Unlock()
				})
			}, nil
		}
		g.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
}

type shuihuoImageStageContextKey struct{}

// withShuihuoImageStage 标记当前出图属于哪个阶段，用于选择全局名额池。
func withShuihuoImageStage(ctx context.Context, stage string) context.Context {
	if ctx == nil || stage == "" {
		return ctx
	}
	return context.WithValue(ctx, shuihuoImageStageContextKey{}, stage)
}

func shuihuoImageStageFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	stage, _ := ctx.Value(shuihuoImageStageContextKey{}).(string)
	return stage
}

// withShuihuoGlobalImageConcurrency 给整个自动化任务加一层总并发上限。
// 图片网关在同时流式返回太多大响应时会直接 RST 连接（实测 HTTP 200 后
// wsarecv 被重置），所以“每书 4 路 × 多本书”会互相拖垮，总量必须收敛。
func withShuihuoGlobalImageConcurrency(ctx context.Context, limit int) context.Context {
	if ctx == nil {
		return ctx
	}
	return context.WithValue(ctx, shuihuoGlobalImageSemaphoreContextKey{}, newShuihuoGlobalImageSemaphores(limit))
}

func shuihuoGlobalImageSemaphoreFromContext(ctx context.Context) *shuihuoGlobalImageSemaphores {
	if ctx == nil {
		return nil
	}
	semaphores, _ := ctx.Value(shuihuoGlobalImageSemaphoreContextKey{}).(*shuihuoGlobalImageSemaphores)
	return semaphores
}

type shuihuoEffectiveSettingsContextKey struct{}

type shuihuoAssetThreeViewContextKey struct{}

// withShuihuoAssetThreeView 控制人物资产图是否套用三视图模板。
// 未显式设置时默认开启，保持原有漫画资产出图行为。
func withShuihuoAssetThreeView(ctx context.Context, enabled bool) context.Context {
	if ctx == nil {
		return ctx
	}
	return context.WithValue(ctx, shuihuoAssetThreeViewContextKey{}, enabled)
}

func shuihuoAssetThreeViewEnabled(ctx context.Context) bool {
	if ctx == nil {
		return true
	}
	value, ok := ctx.Value(shuihuoAssetThreeViewContextKey{}).(bool)
	if !ok {
		return true
	}
	return value
}

// withShuihuoEffectiveSettings 把批次生效设置注入上下文，让出图函数在保存前
// 读取水印等配置。未设置时水印功能默认关闭。
func withShuihuoEffectiveSettings(ctx context.Context, settings map[string]any) context.Context {
	if ctx == nil {
		return ctx
	}
	return context.WithValue(ctx, shuihuoEffectiveSettingsContextKey{}, settings)
}

func shuihuoEffectiveSettingsFromContext(ctx context.Context) map[string]any {
	if ctx == nil {
		return nil
	}
	if settings, ok := ctx.Value(shuihuoEffectiveSettingsContextKey{}).(map[string]any); ok {
		return settings
	}
	return nil
}

type shuihuoImageGenerationCacheContextKey struct{}
type shuihuoImageGenerationCacheBypassContextKey struct{}

func withShuihuoImageGenerationCache(ctx context.Context, cache imagegenerationcache.Repository) context.Context {
	if ctx == nil || cache == nil {
		return ctx
	}
	return context.WithValue(ctx, shuihuoImageGenerationCacheContextKey{}, cache)
}

func shuihuoImageGenerationCacheFromContext(ctx context.Context) imagegenerationcache.Repository {
	if ctx == nil {
		return nil
	}
	cache, _ := ctx.Value(shuihuoImageGenerationCacheContextKey{}).(imagegenerationcache.Repository)
	return cache
}

func withShuihuoImageGenerationCacheBypass(ctx context.Context) context.Context {
	if ctx == nil {
		return ctx
	}
	return context.WithValue(ctx, shuihuoImageGenerationCacheBypassContextKey{}, true)
}

func shuihuoImageGenerationCacheBypassed(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	bypassed, _ := ctx.Value(shuihuoImageGenerationCacheBypassContextKey{}).(bool)
	return bypassed
}

type shuihuoImageFingerprintInput struct {
	Owner            string   `json:"owner"`
	Endpoint         string   `json:"endpoint"`
	ModelID          string   `json:"modelId"`
	Prompt           string   `json:"prompt"`
	AspectRatio      string   `json:"aspectRatio"`
	References       []string `json:"references"`
	WatermarkEnabled bool     `json:"watermarkEnabled"`
	WatermarkText    string   `json:"watermarkText"`
}

func shuihuoImageGenerationFingerprint(input shuihuoImageFingerprintInput) string {
	input.Owner = strings.TrimSpace(input.Owner)
	input.Endpoint = strings.TrimRight(strings.ToLower(strings.TrimSpace(input.Endpoint)), "/")
	input.ModelID = strings.TrimSpace(input.ModelID)
	input.Prompt = strings.TrimSpace(input.Prompt)
	input.AspectRatio = strings.TrimSpace(input.AspectRatio)
	input.WatermarkText = strings.TrimSpace(input.WatermarkText)
	encoded, _ := json.Marshal(input)
	sum := sha256.Sum256(encoded)
	return fmt.Sprintf("%x", sum[:])
}

// withShuihuoImageConcurrency 把当前书生效的出图并发写进上下文，由真正的
// 图片请求读取。未设置时使用默认并发。
//
// 每次调用会同时挂一个属于本次任务（通常是一本书）的信号量实例：出图并发
// 是“每本书各自 N 路”，不是全进程共用一个 N 路上限，否则多本书并行时互相
// 抢槽位，屏幕上写着 4 路实际全批只有 4 张在跑。
func withShuihuoImageConcurrency(ctx context.Context, concurrency int) context.Context {
	if ctx == nil {
		return ctx
	}
	limit := shuihuoClampImageConcurrency(concurrency)
	if existing, ok := ctx.Value(shuihuoImageSemaphoreContextKey{}).(*shuihuoImageSemaphore); ok && existing != nil {
		existing.setLimit(limit)
		return context.WithValue(ctx, shuihuoImageConcurrencyContextKey{}, limit)
	}
	ctx = context.WithValue(ctx, shuihuoImageConcurrencyContextKey{}, limit)
	return context.WithValue(ctx, shuihuoImageSemaphoreContextKey{}, newShuihuoImageSemaphore(limit))
}

func shuihuoImageSemaphoreFromContext(ctx context.Context) *shuihuoImageSemaphore {
	if ctx == nil {
		return nil
	}
	semaphore, _ := ctx.Value(shuihuoImageSemaphoreContextKey{}).(*shuihuoImageSemaphore)
	return semaphore
}

func shuihuoImageConcurrencyFromContext(ctx context.Context) int {
	if ctx == nil {
		return shuihuoImageGenerationConcurrency()
	}
	if concurrency, ok := ctx.Value(shuihuoImageConcurrencyContextKey{}).(int); ok && concurrency > 0 {
		return shuihuoClampImageConcurrency(concurrency)
	}
	return shuihuoImageGenerationConcurrency()
}

func acquireShuihuoGPTImageGenerationSlot(ctx context.Context, modelID string) (func(), error) {
	if !shuihuoIsGPTImageModel(modelID) {
		return func() {}, nil
	}
	releaseProcess, err := shuihuoProcessImageGenerationSemaphore.acquire(ctx)
	if err != nil {
		return nil, err
	}
	// 批次任务在上下文里带有独立信号量时，优先按该任务的并发上限排队。
	task := shuihuoImageSemaphoreFromContext(ctx)
	if task == nil {
		// 单页手动生成等没有任务上下文的路径，回落到进程级上限。
		shuihuoImageGenerationSemaphore.setLimit(shuihuoImageConcurrencyFromContext(ctx))
		task = shuihuoImageGenerationSemaphore
	}
	releaseTask, err := task.acquire(ctx)
	if err != nil {
		releaseProcess()
		return nil, err
	}
	global := shuihuoGlobalImageSemaphoreFromContext(ctx)
	if global == nil {
		return func() {
			releaseTask()
			releaseProcess()
		}, nil
	}
	releaseGlobal, err := global.acquire(ctx, shuihuoImageStageFromContext(ctx))
	if err != nil {
		releaseTask()
		releaseProcess()
		return nil, err
	}
	return func() {
		releaseGlobal()
		releaseTask()
		releaseProcess()
	}, nil
}

func textProviderForShuihuo(director *batchfactoryv11.DirectorService, ownerProvider batchfactoryv11.DirectorProvider) batchfactoryv11.DirectorProvider {
	if ownerProvider != nil {
		return ownerProvider
	}
	if director != nil {
		return director.Provider
	}
	return nil
}

// shuihuoTextModelConfigured mirrors the API Config meaning of "ready": an
// enabled text model with the fields needed to make a request. The health
// endpoint must inspect the same account catalog that the model calls use;
// checking only the optional Director environment variables made a configured
// account look unconfigured in the Shuihuo workbench.
func shuihuoTextModelConfigured(config map[string]any) bool {
	return shuihuoModelKindConfigured(config, "text")
}

// shuihuoModelKindConfigured is the shared readiness rule for the account
// model catalog. It deliberately checks the same endpoint/credential fields
// that runtime adapters use, including credentialRef, so health does not
// disagree with the admin model page.
func shuihuoModelKindConfigured(config map[string]any, kind string) bool {
	for _, model := range shuihuoCollection(config, "modelCatalog", nil) {
		if !strings.EqualFold(modelString(model, "kind"), kind) || model["enabled"] != true {
			continue
		}
		if modelString(model, "modelId") == "" {
			continue
		}
		endpoint := modelString(model, "endpoint")
		if endpoint == "" {
			endpoint = modelString(model, "baseUrl")
		}
		if endpoint == "" {
			continue
		}
		if kind == "audio" && (modelString(model, "requestTemplate") == "" || modelString(model, "responseMapping") == "") {
			continue
		}
		if shuihuoModelCredential(model) == "" && !strings.EqualFold(modelString(model, "adapterKind"), "local_executor_video") {
			continue
		}
		return true
	}
	return false
}

// shuihuoAudioModelConfigured includes the built-in Go TTS HTTP adapter. A
// custom audio entry in modelCatalog can override it, but the absence of one
// does not mean that配音与时长 is unavailable: /api/tts falls back to the
// maintained tts3.121w.com adapter and the batch timeline depends on it.
const shuihuoDefaultTTSServiceURL = "http://tts3.121w.com/v1/audio/speech"

func shuihuoAudioModelConfigured(config map[string]any) bool {
	return shuihuoModelKindConfigured(config, "audio") || strings.TrimSpace(shuihuoDefaultTTSServiceURL) != ""
}

// registerShuihuoProjectCompatRoutes keeps the legacy shared-work library
// contract available while the browser still reads its project cards through
// /api/platform-projects. The source of truth is the Go batch store.
func shuihuoRedisHealth(ctx context.Context, rawURL string) (bool, string) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return false, "未配置 Redis 地址"
	}
	store, err := mergeworker.NewRedisStore(rawURL, "qiantie:health")
	if err != nil {
		return false, "Redis 地址无效"
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	value, err := store.Do(ctx, "PING")
	if err != nil {
		return false, "Redis 连接失败"
	}
	if strings.EqualFold(strings.TrimSpace(fmt.Sprint(value)), "PONG") {
		return true, ""
	}
	return false, "Redis 未返回 PONG"
}

func registerShuihuoProjectCompatRoutes(root *http.ServeMux, auth BridgeAuth, store batchfactoryv11.Store, files *localartifact.Store, director *batchfactoryv11.DirectorService, textProvider batchfactoryv11.DirectorProvider, production *batchfactoryv11.ProductionService, catalogStore novelfetchworkshop.Store, presetCatalog *prompt.Catalog, redisURL string, localExecutors *localexecutor.Service, userConfig *UserConfigStore) {
	if store == nil {
		return
	}
	root.Handle("/api/shuihuo-production/health", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		enabledKinds := []string{}
		var ownerConfig map[string]any
		if config, err := loadShuihuoOwnerConfig(r.Context(), store, owner, catalogStore); err == nil {
			ownerConfig = config
		}
		textReady := director != nil && director.Provider != nil
		if !textReady && textProvider != nil && ownerConfig != nil {
			textReady = shuihuoTextModelConfigured(ownerConfig)
		}
		if textReady {
			enabledKinds = append(enabledKinds, "text")
		}
		if production != nil && production.Enabled {
			enabledKinds = append(enabledKinds, "video")
		}
		// 出图既可能来自旧 modelCatalog，也可能来自 API 接口配置里的渠道；
		// 顶部能力条必须与真实出图同口径，否则会误报“图片模型未配置”。
		if (ownerConfig != nil && shuihuoModelKindConfigured(ownerConfig, "image")) || shuihuoImageChannelReady(r.Context(), userConfig, owner) {
			enabledKinds = append(enabledKinds, "image")
		}
		if shuihuoAudioModelConfigured(ownerConfig) {
			enabledKinds = append(enabledKinds, "audio")
		}
		redisReady, redisReason := shuihuoRedisHealth(r.Context(), redisURL)
		writeJSON(w, http.StatusOK, map[string]any{
			"ok": true, "database": map[string]any{"ready": true},
			"redis":   map[string]any{"ready": redisReady, "reason": redisReason},
			"storage": map[string]any{"ready": true}, "enabledModelKinds": enabledKinds,
		})
	})))
	root.Handle("/api/shuihuo-production/models", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		models := []any{}
		if _, durableOK := store.(batchfactoryv11.ShuihuoConfigStore); durableOK {
			if config, err := loadShuihuoOwnerConfig(r.Context(), store, owner, catalogStore); err == nil {
				if catalog, catalogOK := config["modelCatalog"].([]any); catalogOK {
					for _, raw := range catalog {
						if model, modelOK := raw.(map[string]any); modelOK {
							// Never return credentials from the account-level catalog.
							copy := map[string]any{}
							for _, key := range []string{"id", "name", "displayName", "kind", "provider", "adapterKind", "modelId", "enabled", "configured"} {
								if value, exists := model[key]; exists {
									copy[key] = value
								}
							}
							models = append(models, copy)
						}
					}
				}
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"models": models})
	})))
	root.Handle("/api/shuihuo-production/preset-slots", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := bridgeOwner(r); !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		presets := make([]map[string]any, 0)
		if presetCatalog != nil {
			published, err := presetCatalog.Published("shuihuo-production")
			if err != nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "提示词预设目录不可用"})
				return
			}
			for _, item := range published {
				if item.Slot == nil || strings.TrimSpace(*item.Slot) == "" {
					continue
				}
				presets = append(presets, map[string]any{"id": item.ID, "name": item.Name, "slot": *item.Slot, "version": item.Version})
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"presets": presets})
	})))
	root.Handle("/api/shuihuo-production/admin/models", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := BridgeIdentityFromContext(r.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if !identity.IsOwner {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "无权管理模型"})
			return
		}
		config, err := loadShuihuoOwnerConfig(r.Context(), store, identity.Username, catalogStore)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		models := shuihuoCollection(config, "modelCatalog", nil)
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, map[string]any{"models": safeShuihuoAdminModels(models)})
			return
		}
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var input map[string]any
		if !decodeJSON(w, r, &input) {
			return
		}
		modelID, _ := input["modelId"].(string)
		name, _ := input["name"].(string)
		kind, _ := input["kind"].(string)
		if strings.TrimSpace(modelID) == "" || strings.TrimSpace(name) == "" || strings.TrimSpace(kind) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "modelId, name and kind are required"})
			return
		}
		for _, model := range models {
			if model["id"] == modelID || model["modelId"] == modelID {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "模型 ID 已存在"})
				return
			}
		}
		model := map[string]any{}
		for _, key := range []string{"modelId", "name", "kind", "adapterKind", "enabled", "parameterSchema", "credentialRef", "endpoint", "tasksEndpoint", "resultEndpoint", "workflowId", "requestTemplate", "responseMapping", "voiceCloneRequestTemplate", "voiceCloneResponseMapping", "capabilities", "credential", "apiKey"} {
			if value, exists := input[key]; exists {
				model[key] = value
			}
		}
		model["id"] = modelID
		models = append(models, model)
		config["modelCatalog"] = models
		if err := saveShuihuoAdminModelCatalog(r.Context(), store, catalogStore, identity.Username, config, models); err != nil {
			writeStoreError(w, err)
			return
		}
		if production != nil && production.ProviderRegistry != nil && strings.EqualFold(strings.TrimSpace(kind), "video") {
			if syncErr := syncShuihuoVideoModel(r.Context(), production.ProviderRegistry, identity.Username, model); syncErr != nil {
				writeStoreError(w, syncErr)
				return
			}
		}
		writeJSON(w, http.StatusCreated, map[string]any{"model": safeShuihuoAdminModels([]map[string]any{model})[0]})
	})))
	root.Handle("/api/shuihuo-production/admin/models/{modelId}", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := BridgeIdentityFromContext(r.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if !identity.IsOwner {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "无权管理模型"})
			return
		}
		config, err := loadShuihuoOwnerConfig(r.Context(), store, identity.Username, catalogStore)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		models := shuihuoCollection(config, "modelCatalog", nil)
		modelID := strings.TrimSpace(r.PathValue("modelId"))
		index := -1
		for i, current := range models {
			if modelString(current, "id") == modelID || modelString(current, "modelId") == modelID {
				index = i
				break
			}
		}
		if index < 0 {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "模型不存在"})
			return
		}
		if r.Method == http.MethodDelete {
			models = append(models[:index], models[index+1:]...)
			config["modelCatalog"] = models
			if err := saveShuihuoAdminModelCatalog(r.Context(), store, catalogStore, identity.Username, config, models); err != nil {
				writeStoreError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPatch {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var patch map[string]any
		if !decodeJSON(w, r, &patch) {
			return
		}
		current, patchErr := patchShuihuoModel(models[index], patch)
		if patchErr != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": patchErr.Error()})
			return
		}
		models[index] = current
		config["modelCatalog"] = models
		if err := saveShuihuoAdminModelCatalog(r.Context(), store, catalogStore, identity.Username, config, models); err != nil {
			writeStoreError(w, err)
			return
		}
		if production != nil && production.ProviderRegistry != nil && strings.EqualFold(strings.TrimSpace(modelString(current, "kind")), "video") {
			if syncErr := syncShuihuoVideoModel(r.Context(), production.ProviderRegistry, identity.Username, current); syncErr != nil {
				writeStoreError(w, syncErr)
				return
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"model": safeShuihuoAdminModels([]map[string]any{current})[0]})
	})))
	root.Handle("/api/shuihuo-production/admin/models/{modelId}/test", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := BridgeIdentityFromContext(r.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if !identity.IsOwner {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "无权测试模型"})
			return
		}
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		config, err := loadShuihuoOwnerConfig(r.Context(), store, identity.Username, catalogStore)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		modelID := strings.TrimSpace(r.PathValue("modelId"))
		var model map[string]any
		for _, candidate := range shuihuoCollection(config, "modelCatalog", nil) {
			if modelString(candidate, "id") == modelID || modelString(candidate, "modelId") == modelID {
				model = candidate
				break
			}
		}
		if model == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "模型不存在"})
			return
		}
		kind := strings.ToLower(modelString(model, "kind"))
		endpoint := modelString(model, "endpoint")
		if endpoint == "" {
			endpoint = modelString(model, "baseUrl")
		}
		credential := shuihuoModelCredential(model)
		if endpoint == "" || credential == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "kind": kind, "error": "模型地址或凭据未配置"})
			return
		}
		client := safehttp.NewClientWithTimeout(20 * time.Second)
		if kind == "image" {
			imageEndpoint, endpointErr := shuihuoImageGenerationEndpoint(endpoint)
			if endpointErr != nil {
				writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"ok": false, "kind": kind, "error": "生图服务地址格式不正确"})
				return
			}
			parsed, parseErr := url.Parse(imageEndpoint)
			if parseErr != nil || parsed.Host == "" {
				writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"ok": false, "kind": kind, "error": "生图服务地址格式不正确"})
				return
			}
			probeURL := *parsed
			probeURL.Path = strings.TrimSuffix(probeURL.Path, "/images/generations") + "/models"
			probe, requestErr := http.NewRequestWithContext(r.Context(), http.MethodGet, probeURL.String(), nil)
			if requestErr != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "kind": kind, "error": "生图服务地址无效"})
				return
			}
			probe.Header.Set("Authorization", "Bearer "+credential)
			response, requestErr := client.Do(probe)
			if requestErr != nil {
				writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "kind": kind, "error": "模型服务连接失败，请检查地址和网络", "type": "upstream_connection_error"})
				return
			}
			defer response.Body.Close()
			if response.StatusCode < 200 || response.StatusCode >= 300 {
				writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "kind": kind, "error": fmt.Sprintf("上游模型服务返回状态 %d", response.StatusCode), "type": "upstream_error"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "kind": kind, "message": "生图服务连接成功"})
			return
		}
		if kind == "text" {
			parsed, parseErr := url.Parse(endpoint)
			if parseErr != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
				writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"ok": false, "kind": kind, "error": "文本服务地址格式不正确"})
				return
			}
			textEndpoint := strings.TrimRight(endpoint, "/")
			if !strings.HasSuffix(parsed.Path, "/chat/completions") {
				textEndpoint += "/chat/completions"
			}
			payload, _ := json.Marshal(map[string]any{"model": modelString(model, "modelId"), "messages": []map[string]string{{"role": "user", "content": "Hi"}}, "max_tokens": 5})
			probe, requestErr := http.NewRequestWithContext(r.Context(), http.MethodPost, textEndpoint, bytes.NewReader(payload))
			if requestErr != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "kind": kind, "error": "文本服务地址无效"})
				return
			}
			probe.Header.Set("Content-Type", "application/json")
			probe.Header.Set("Authorization", "Bearer "+credential)
			response, requestErr := client.Do(probe)
			if requestErr != nil {
				writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "kind": kind, "error": "模型服务连接失败，请检查地址和网络", "type": "upstream_connection_error"})
				return
			}
			defer response.Body.Close()
			if response.StatusCode < 200 || response.StatusCode >= 300 {
				writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "kind": kind, "error": fmt.Sprintf("上游模型服务返回状态 %d", response.StatusCode), "type": "upstream_error"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "kind": kind, "message": "文本模型连接成功"})
			return
		}
		if kind == "audio" {
			_, contentType, probeErr := requestConfiguredTTSBytes(r.Context(), ttsRequest{Input: "连接测试", Voice: "zh-CN-XiaoxiaoNeural", Speed: 1.8, Pitch: "10", Style: "general"}, model)
			if probeErr != nil {
				writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "kind": kind, "error": "音频模型连接失败: " + probeErr.Error(), "type": "upstream_error"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "kind": kind, "contentType": contentType, "message": "音频模型连接成功，已返回音频数据"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "kind": kind, "tested": "configuration", "message": "配置完整，未发送生成任务"})
	})))
	root.Handle("/api/shuihuo-production/asset-types", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		config, err := loadShuihuoOwnerConfig(r.Context(), store, owner, catalogStore)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		types := shuihuoCollection(config, "assetTypes", []map[string]any{
			{"id": "system-character", "name": "角色", "category": "character", "system": true},
			{"id": "system-scene", "name": "场景", "category": "scene", "system": true},
			{"id": "system-prop", "name": "道具", "category": "prop", "system": true},
		})
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, map[string]any{"assetTypes": types})
			return
		}
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var input map[string]any
		if !decodeJSON(w, r, &input) {
			return
		}
		name, _ := input["name"].(string)
		category, _ := input["category"].(string)
		if strings.TrimSpace(name) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "type name is required"})
			return
		}
		id, idErr := localexecutor.NewArtifactID()
		if idErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "type id generation failed"})
			return
		}
		types = append(types, map[string]any{"id": id, "name": strings.TrimSpace(name), "category": strings.TrimSpace(category), "system": false})
		config["assetTypes"] = types
		if err := saveShuihuoOwnerConfig(r.Context(), store, owner, config); err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"assetType": types[len(types)-1]})
	})))
	root.Handle("/api/shuihuo-production/asset-types/", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/shuihuo-production/asset-types/"), "/")
		config, err := loadShuihuoOwnerConfig(r.Context(), store, owner, catalogStore)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		types := shuihuoCollection(config, "assetTypes", nil)
		for index, item := range types {
			if item["id"] != id {
				continue
			}
			if r.Method == http.MethodPut {
				if item["system"] == true {
					writeJSON(w, http.StatusConflict, map[string]string{"error": "system asset type cannot be edited"})
					return
				}
				var input map[string]any
				if !decodeJSON(w, r, &input) {
					return
				}
				name := strings.TrimSpace(modelString(input, "name"))
				if name == "" {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "type name is required"})
					return
				}
				item["name"] = name
				item["category"] = strings.TrimSpace(modelString(input, "category"))
				types[index] = item
				config["assetTypes"] = types
				if err := saveShuihuoOwnerConfig(r.Context(), store, owner, config); err != nil {
					writeStoreError(w, err)
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"assetType": item})
				return
			}
			if item["system"] == true {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "system asset type cannot be deleted"})
				return
			}
			if r.Method == http.MethodDelete {
				types = append(types[:index], types[index+1:]...)
				config["assetTypes"] = types
				if err := saveShuihuoOwnerConfig(r.Context(), store, owner, config); err != nil {
					writeStoreError(w, err)
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"ok": true})
				return
			}
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "asset type not found"})
	})))
	root.Handle("/api/shuihuo-production/asset-templates", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		config, err := loadShuihuoOwnerConfig(r.Context(), store, owner, catalogStore)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		templates := shuihuoCollection(config, "assetTemplates", nil)
		if r.Method == http.MethodGet {
			if filter := strings.TrimSpace(r.URL.Query().Get("assetTypeId")); filter != "" {
				filtered := []map[string]any{}
				for _, item := range templates {
					if item["assetTypeId"] == filter {
						filtered = append(filtered, item)
					}
				}
				templates = filtered
			}
			writeJSON(w, http.StatusOK, map[string]any{"assetTemplates": templates})
			return
		}
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var input map[string]any
		if !decodeJSON(w, r, &input) {
			return
		}
		name, _ := input["name"].(string)
		typeID, _ := input["assetTypeId"].(string)
		prompt, _ := input["prompt"].(string)
		if strings.TrimSpace(name) == "" || strings.TrimSpace(typeID) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "template name and assetTypeId are required"})
			return
		}
		id, idErr := localexecutor.NewArtifactID()
		if idErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "template id generation failed"})
			return
		}
		item := map[string]any{"id": id, "assetTypeId": strings.TrimSpace(typeID), "name": strings.TrimSpace(name), "prompt": strings.TrimSpace(prompt), "source": "manual"}
		templates = append(templates, item)
		config["assetTemplates"] = templates
		if err := saveShuihuoOwnerConfig(r.Context(), store, owner, config); err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"assetTemplate": item})
	})))
	root.Handle("/api/shuihuo-production/asset-templates/", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/shuihuo-production/asset-templates/"), "/")
		config, err := loadShuihuoOwnerConfig(r.Context(), store, owner, catalogStore)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		templates := shuihuoCollection(config, "assetTemplates", nil)
		for index, item := range templates {
			if item["id"] != id {
				continue
			}
			if r.Method == http.MethodPut {
				var input map[string]any
				if !decodeJSON(w, r, &input) {
					return
				}
				name := strings.TrimSpace(modelString(input, "name"))
				if name == "" {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "template name is required"})
					return
				}
				item["name"] = name
				item["prompt"] = strings.TrimSpace(modelString(input, "prompt"))
				if assetTypeID := strings.TrimSpace(modelString(input, "assetTypeId")); assetTypeID != "" {
					item["assetTypeId"] = assetTypeID
				}
				templates[index] = item
				config["assetTemplates"] = templates
				if err := saveShuihuoOwnerConfig(r.Context(), store, owner, config); err != nil {
					writeStoreError(w, err)
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"assetTemplate": item})
				return
			}
			if r.Method != http.MethodDelete {
				writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
				return
			}
			templates = append(templates[:index], templates[index+1:]...)
			config["assetTemplates"] = templates
			if err := saveShuihuoOwnerConfig(r.Context(), store, owner, config); err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "asset template not found"})
	})))
	root.Handle("/api/shuihuo-production/config", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if r.Method == http.MethodGet {
			if _, ok := store.(batchfactoryv11.ShuihuoConfigStore); ok {
				config, err := loadShuihuoOwnerConfig(r.Context(), store, owner, catalogStore)
				if err != nil {
					writeStoreError(w, err)
					return
				}
				merged := shuihuoDefaultConfig()
				for key, value := range config {
					merged[key] = value
				}
				writeJSON(w, http.StatusOK, safeShuihuoConfigForResponse(merged))
				return
			}
			if value, exists := shuihuoProductionConfigs.Load(owner); exists {
				if config, ok := value.(map[string]any); ok {
					writeJSON(w, http.StatusOK, safeShuihuoConfigForResponse(config))
				} else {
					writeJSON(w, http.StatusOK, value)
				}
				return
			}
			writeJSON(w, http.StatusOK, safeShuihuoConfigForResponse(shuihuoDefaultConfig()))
			return
		}
		if r.Method == http.MethodPut {
			var input map[string]any
			if !decodeJSON(w, r, &input) {
				return
			}
			config := shuihuoDefaultConfig()
			for key, value := range input {
				config[key] = value
			}
			if durable, ok := store.(batchfactoryv11.ShuihuoConfigStore); ok {
				if err := durable.SaveShuihuoConfig(r.Context(), owner, config); err != nil {
					writeStoreError(w, err)
					return
				}
				writeJSON(w, http.StatusOK, config)
				return
			}
			shuihuoProductionConfigs.Store(owner, config)
			writeJSON(w, http.StatusOK, config)
			return
		}
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	})))
	root.Handle("/api/shuihuo-production/projects/import", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var input struct {
			Name     string `json:"name"`
			Filename string `json:"filename"`
			DataURL  string `json:"dataUrl"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		name := strings.TrimSpace(input.Name)
		if name == "" || strings.TrimSpace(input.DataURL) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and dataUrl are required"})
			return
		}
		parts := strings.SplitN(strings.TrimSpace(input.DataURL), ",", 2)
		if len(parts) != 2 || !strings.HasSuffix(parts[0], ";base64") {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid source file"})
			return
		}
		raw, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil || len(raw) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid source file"})
			return
		}
		ext := strings.ToLower(filepath.Ext(strings.TrimSpace(input.Filename)))
		if ext == ".docx" {
			text, extractErr := extractShuihuoDocxText(raw)
			if extractErr != nil || strings.TrimSpace(text) == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "DOCX 正文读取失败或内容为空"})
				return
			}
			raw = []byte(text)
		}
		if ext != "" && ext != ".txt" && ext != ".srt" && ext != ".vtt" && ext != ".docx" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "仅支持 TXT、SRT、VTT 或 DOCX 文件"})
			return
		}
		batch, err := store.CreateBatch(r.Context(), owner, batchfactoryv11.CreateBatchInput{Title: name, InitialBatchSettings: batchfactoryv11.SettingsPatch{"shuihuoProject": json.RawMessage("true")}, Books: []batchfactoryv11.CreateBookInput{{Title: name}}})
		if err != nil {
			writeStoreError(w, err)
			return
		}
		if len(batch.Books) == 0 {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "project has no editable source book"})
			return
		}
		book, err := store.CaptureBookSource(r.Context(), owner, batch.ID, batch.Books[0].ID, batchfactoryv11.CaptureBookSourceInput{SourceText: string(raw), ReplaceSource: true, SourceTitle: input.Filename, ExpectedRevision: batch.Books[0].Revision})
		if err != nil {
			_ = store.DeleteBatch(r.Context(), owner, batch.ID)
			writeStoreError(w, err)
			return
		}
		batch.Books[0] = book
		writeJSON(w, http.StatusCreated, shuihuoProjectDetail(batch))
	})))
	root.Handle("/api/shuihuo-production/projects", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		switch r.Method {
		case http.MethodGet:
			if id := strings.TrimSpace(r.URL.Query().Get("id")); id != "" {
				writeShuihuoProject(w, r, store, owner, id)
				return
			}
			batches, err := store.ListBatches(r.Context(), owner)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			projects := make([]map[string]any, 0, len(batches))
			for _, batch := range batches {
				if !isShuihuoProjectBatch(batch) {
					continue
				}
				projects = append(projects, shuihuoProject(batch))
			}
			writeJSON(w, http.StatusOK, map[string]any{"projects": projects})
		case http.MethodPost:
			var input struct {
				Name           string `json:"name"`
				ProductionMode string `json:"productionMode,omitempty"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			name := strings.TrimSpace(input.Name)
			if name == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
				return
			}
			batch, err := store.CreateBatch(r.Context(), owner, batchfactoryv11.CreateBatchInput{
				Title: name,
				InitialBatchSettings: batchfactoryv11.SettingsPatch{
					"shuihuoProject": json.RawMessage("true"),
				},
				// Keep one editable source book so the legacy source PUT can update
				// a newly-created project before any segmentation work starts.
				Books: []batchfactoryv11.CreateBookInput{{Title: name}},
			})
			if err != nil {
				writeStoreError(w, err)
				return
			}
			project := shuihuoProject(batch)
			if input.ProductionMode != "" {
				project["productionMode"] = input.ProductionMode
			}
			writeJSON(w, http.StatusCreated, project)
		case http.MethodDelete:
			id := strings.TrimSpace(r.URL.Query().Get("id"))
			if id == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
				return
			}
			if err := store.DeleteBatch(r.Context(), owner, id); err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	})))

	root.Handle("/api/shuihuo-production/projects/", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/api/shuihuo-production/projects/")
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) == 0 || parts[0] == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "project id is required"})
			return
		}
		batch, err := store.GetBatch(r.Context(), owner, parts[0])
		if err != nil {
			writeStoreError(w, err)
			return
		}
		if len(parts) == 1 && r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, shuihuoProjectDetail(batch))
			return
		}
		if len(parts) == 1 && r.Method == http.MethodDelete {
			if err := store.DeleteBatch(r.Context(), owner, batch.ID); err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			return
		}
		if len(parts) == 2 && parts[1] == "segments" && r.Method == http.MethodPost {
			var input map[string]any
			if !decodeJSON(w, r, &input) {
				return
			}
			if len(batch.Books) == 0 {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "project has no editable source book"})
				return
			}
			text, _ := input["sourceText"].(string)
			if strings.TrimSpace(text) == "" {
				text, _ = input["text"].(string)
			}
			if strings.TrimSpace(text) == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "sourceText is required"})
				return
			}
			segments := storedShuihuoSegments(batch.Books[0])
			segment := normalizeShuihuoSegments([]map[string]any{input}, batch.Books[0].ID)
			segment[0]["id"] = fmt.Sprintf("segment_%s_%d", batch.Books[0].ID, len(segments)+1)
			segments = append(segments, segment[0])
			if err := saveShuihuoSegments(r, store, owner, batch, segments); err != nil {
				writeStoreError(w, err)
				return
			}
			updated, err := store.GetBatch(r.Context(), owner, batch.ID)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, shuihuoProjectDetail(updated))
			return
		}
		if len(parts) == 3 && parts[1] == "segments" && parts[2] == "order" && r.Method == http.MethodPut {
			var input struct {
				SegmentIDs []string `json:"segmentIds"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			segments := storedShuihuoSegments(batch.Books[0])
			ordered := reorderShuihuoSegments(segments, input.SegmentIDs)
			if err := saveShuihuoSegments(r, store, owner, batch, ordered); err != nil {
				writeStoreError(w, err)
				return
			}
			updated, err := store.GetBatch(r.Context(), owner, batch.ID)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, shuihuoProjectDetail(updated))
			return
		}
		if len(parts) == 2 && parts[1] == "voice-settings" && (r.Method == http.MethodGet || r.Method == http.MethodPut) {
			settings := map[string]any{}
			if len(batch.Books) > 0 {
				for _, segment := range storedShuihuoSegments(batch.Books[0]) {
					segmentID := modelString(segment, "id")
					if segmentID == "" {
						continue
					}
					settings[segmentID] = map[string]any{
						"voiceAssetId": segment["voiceAssetId"],
						"speechRate":   segment["speechRate"],
						"pitch":        segment["pitch"],
					}
				}
			}
			writeJSON(w, http.StatusOK, map[string]any{"settings": settings})
			return
		}
		if len(parts) == 2 && parts[1] == "asset-generation-config" && (r.Method == http.MethodGet || r.Method == http.MethodPut) {
			if len(batch.Books) == 0 {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "project has no editable source book"})
				return
			}
			config := storedShuihuoAssetGenerationConfig(batch)
			if r.Method == http.MethodPut {
				var input map[string]any
				if !decodeJSON(w, r, &input) {
					return
				}
				config = input
				metadata := cloneMetadata(batch.Books[0].SourceMetadata)
				encoded, _ := json.Marshal(config)
				var normalized map[string]any
				if json.Unmarshal(encoded, &normalized) == nil {
					metadata["shuihuoAssetGenerationConfig"] = normalized
				}
				if _, saveErr := store.UpdateBookMetadata(r.Context(), owner, batch.ID, batch.Books[0].ID, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata, ExpectedRevision: batch.Books[0].Revision}); saveErr != nil {
					writeStoreError(w, saveErr)
					return
				}
			}
			writeJSON(w, http.StatusOK, config)
			return
		}
		if len(parts) == 2 && parts[1] == "files" && r.Method == http.MethodGet {
			if len(batch.Books) == 0 {
				writeJSON(w, http.StatusOK, map[string]any{"hasSourceFile": false, "media": []any{}})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"hasSourceFile": strings.TrimSpace(batch.Books[0].TxtFileName) != "" || strings.TrimSpace(batch.Books[0].SourceText) != "", "media": storedShuihuoMedia(batch.Books[0])})
			return
		}
		if len(parts) == 2 && parts[1] == "media" && r.Method == http.MethodPost {
			if files == nil || len(batch.Books) == 0 {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "media storage is unavailable"})
				return
			}
			var input struct {
				Kind      string `json:"kind"`
				Filename  string `json:"filename"`
				DataURL   string `json:"dataUrl"`
				SegmentID string `json:"segmentId,omitempty"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			mediaType, payload, err := decodeShuihuoMediaDataURL(input.DataURL)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid media upload"})
				return
			}
			kind := strings.ToLower(strings.TrimSpace(input.Kind))
			if kind != "image" && kind != "video" && kind != "audio" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported media kind"})
				return
			}
			artifactID, err := localexecutor.NewArtifactID()
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "media id generation failed"})
				return
			}
			extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(strings.TrimSpace(input.Filename))), ".")
			if extension == "" {
				extension = map[string]string{"image": "png", "video": "mp4", "audio": "mp3"}[kind]
			}
			var saved localartifact.Artifact
			if kind == "image" {
				saved, err = files.SaveImage(artifactID, mediaType, bytes.NewReader(payload))
			} else {
				saved, err = files.SaveMedia(artifactID, mediaType, extension, bytes.NewReader(payload))
			}
			if err != nil {
				writeBookAssetImageStorageError(w, err)
				return
			}
			media := storedShuihuoMedia(batch.Books[0])
			media = append(media, map[string]any{"id": artifactID, "kind": kind, "filename": input.Filename, "mediaType": mediaType, "storageRef": saved.StorageRef, "segmentId": strings.TrimSpace(input.SegmentID), "isPrimary": false, "createdAt": time.Now().UTC().Format(time.RFC3339)})
			metadata := cloneMetadata(batch.Books[0].SourceMetadata)
			metadata["shuihuoMedia"] = media
			if _, err := store.UpdateBookMetadata(r.Context(), owner, batch.ID, batch.Books[0].ID, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata, ExpectedRevision: batch.Books[0].Revision}); err != nil {
				_ = files.Remove(saved.StorageRef)
				writeStoreError(w, err)
				return
			}
			updated, err := store.GetBatch(r.Context(), owner, batch.ID)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]any{"media": map[string]any{"id": artifactID, "kind": kind, "filename": input.Filename, "mediaType": mediaType, "segmentId": strings.TrimSpace(input.SegmentID), "isPrimary": false}, "project": shuihuoProjectDetail(updated)})
			return
		}
		if len(parts) == 2 && parts[1] == "tasks" && r.Method == http.MethodGet {
			tasks := make([]map[string]any, 0)
			if production != nil {
				status, statusErr := production.GetBatchStatus(r.Context(), owner, batch.ID)
				if statusErr != nil {
					writeStoreError(w, statusErr)
					return
				}
				for _, job := range status.Jobs {
					for _, task := range job.Tasks {
						tasks = append(tasks, map[string]any{"id": task.ID, "status": task.Status, "kind": "video", "mediaUrl": task.MediaURL, "errorMessage": task.ErrorMessage, "requestId": job.RequestID, "createdAt": task.CreatedAt, "updatedAt": task.UpdatedAt})
					}
				}
			}
			if len(batch.Books) > 0 {
				for _, media := range storedShuihuoMedia(batch.Books[0]) {
					mediaKind := modelString(media, "kind")
					if mediaKind != "audio" && mediaKind != "image" {
						continue
					}
					tasks = append(tasks, map[string]any{"id": modelString(media, "id"), "status": "completed", "kind": mediaKind, "segmentId": modelString(media, "segmentId"), "mediaUrl": "/api/shuihuo-production/media/" + modelString(media, "id") + "/download", "createdAt": media["createdAt"], "updatedAt": media["createdAt"]})
				}
			}
			writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
			return
		}
		if len(parts) == 2 && parts[1] == "tasks" && r.Method == http.MethodPost {
			if production == nil || len(batch.Books) == 0 {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "当前 Go 本地环境尚未配置可用的生成模型"})
				return
			}
			var input struct {
				RequestID string `json:"requestId"`
				Provider  string `json:"provider,omitempty"`
				ModelID   string `json:"modelId,omitempty"`
				Kind      string `json:"kind,omitempty"`
				SegmentID string `json:"segmentId,omitempty"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			if strings.EqualFold(strings.TrimSpace(input.Kind), "audio") {
				if files == nil {
					writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "media storage is unavailable"})
					return
				}
				task, audioErr := generateShuihuoSegmentAudio(r.Context(), store, files, owner, batch, strings.TrimSpace(input.SegmentID), catalogStore)
				if audioErr != nil {
					writeStoreError(w, audioErr)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]any{"tasks": []any{task}, "job": map[string]any{"requestId": input.RequestID, "tasks": []any{task}}})
				return
			}
			if strings.EqualFold(strings.TrimSpace(input.Kind), "image") {
				if files == nil {
					writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "media storage is unavailable"})
					return
				}
				task, imageErr := generateShuihuoSegmentImage(r.Context(), store, files, owner, batch, strings.TrimSpace(input.SegmentID), strings.TrimSpace(input.ModelID), catalogStore)
				if imageErr != nil {
					writeStoreError(w, imageErr)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]any{"tasks": []any{task}, "job": map[string]any{"requestId": input.RequestID, "tasks": []any{task}}})
				return
			}
			if kind := strings.ToLower(strings.TrimSpace(input.Kind)); kind != "" && kind != "video" {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "Go 本地迁移阶段暂未接入图片或配音任务，请先使用视频任务"})
				return
			}
			if production.ProviderRegistry != nil {
				if syncErr := syncShuihuoConfiguredVideoModels(r.Context(), store, production.ProviderRegistry, owner, catalogStore); syncErr != nil {
					writeStoreError(w, syncErr)
					return
				}
			}
			if strings.TrimSpace(input.RequestID) == "" {
				input.RequestID = "shuihuo-" + time.Now().UTC().Format("20060102T150405.000000000")
			}
			provider := input.Provider
			if strings.TrimSpace(input.ModelID) != "" {
				provider = batchfactoryv11.VideoProviderForModel(input.ModelID, provider)
			}
			job, submitErr := production.SubmitBookProductionWithProvider(r.Context(), owner, batch.ID, batch.Books[0].ID, input.RequestID, provider)
			if submitErr != nil {
				writeStoreError(w, submitErr)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]any{"tasks": job.Tasks, "job": job})
			return
		}
		if len(parts) == 3 && parts[1] == "tasks" && parts[2] == "batch" && r.Method == http.MethodPost {
			if production == nil {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "当前 Go 本地环境尚未配置可用的生成模型"})
				return
			}
			var input struct {
				RequestID string `json:"requestId"`
				Provider  string `json:"provider,omitempty"`
				ModelID   string `json:"modelId,omitempty"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			if production.ProviderRegistry != nil {
				if syncErr := syncShuihuoConfiguredVideoModels(r.Context(), store, production.ProviderRegistry, owner, catalogStore); syncErr != nil {
					writeStoreError(w, syncErr)
					return
				}
			}
			if strings.TrimSpace(input.RequestID) == "" {
				input.RequestID = "shuihuo-batch-" + time.Now().UTC().Format("20060102T150405.000000000")
			}
			provider := input.Provider
			if strings.TrimSpace(input.ModelID) != "" {
				provider = batchfactoryv11.VideoProviderForModel(input.ModelID, provider)
			}
			status, submitErr := production.SubmitBatchProductionWithProvider(r.Context(), owner, batch.ID, input.RequestID, provider)
			if submitErr != nil {
				writeStoreError(w, submitErr)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]any{"tasks": status.Jobs, "status": status})
			return
		}
		if (len(parts) == 3 && parts[1] == "analysis" && parts[2] == "assets" || len(parts) == 4 && parts[1] == "analysis" && parts[2] == "assets" && parts[3] == "and-bindings") && r.Method == http.MethodPost {
			if director == nil || len(batch.Books) == 0 {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "当前 Go 本地环境尚未配置文本模型，暂不能进行智能分析"})
				return
			}
			assets, analysisErr := director.RunAssetExtraction(r.Context(), owner, batch.ID, batch.Books[0].ID)
			if analysisErr != nil {
				writeStoreError(w, analysisErr)
				return
			}
			updated, getErr := store.GetBatch(r.Context(), owner, batch.ID)
			if getErr != nil {
				writeStoreError(w, getErr)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"assets": assets, "project": shuihuoProjectDetail(updated)})
			return
		}
		if len(parts) == 4 && parts[1] == "assets" && parts[2] == "candidates" && parts[3] == "apply" && r.Method == http.MethodPost {
			if len(batch.Books) == 0 {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "project has no editable source book"})
				return
			}
			var input struct {
				Candidates []map[string]any `json:"candidates"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			created := 0
			for _, candidate := range input.Candidates {
				name := strings.TrimSpace(modelString(candidate, "name"))
				if name == "" {
					name = strings.TrimSpace(modelString(candidate, "title"))
				}
				if name == "" {
					continue
				}
				kind := strings.TrimSpace(modelString(candidate, "category"))
				if kind == "" {
					kind = strings.TrimSpace(modelString(candidate, "kind"))
				}
				if kind == "" {
					kind = "character"
				}
				prompt := strings.TrimSpace(modelString(candidate, "prompt"))
				if prompt == "" {
					prompt = strings.TrimSpace(modelString(candidate, "description"))
				}
				formalName := modelString(candidate, "formalName")
				if formalName == "" {
					formalName = modelString(candidate, "formal_name")
				}
				if _, createErr := store.CreateBookAsset(r.Context(), owner, batch.ID, batch.Books[0].ID, batchfactoryv11.CreateBookAssetInput{Kind: kind, Name: name, FormalName: formalName, Facts: modelString(candidate, "facts"), Appearance: modelString(candidate, "appearance"), Relationships: modelString(candidate, "relationships"), Prompt: prompt}); createErr != nil {
					writeStoreError(w, createErr)
					return
				}
				created++
			}
			updated, getErr := store.GetBatch(r.Context(), owner, batch.ID)
			if getErr != nil {
				writeStoreError(w, getErr)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"created": created, "project": shuihuoProjectDetail(updated)})
			return
		}
		if len(parts) == 3 && parts[1] == "prompt-candidates" && r.Method == http.MethodPost {
			var input struct {
				ModelID        string `json:"modelId,omitempty"`
				PromptPresetID string `json:"promptPresetId,omitempty"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			kind := strings.TrimSpace(parts[2])
			candidates, strategy, generationErr := generateShuihuoPromptCandidates(r.Context(), textProviderForShuihuo(director, textProvider), batch, kind)
			response := map[string]any{"candidates": candidates, "kind": kind, "strategy": strategy}
			if strings.TrimSpace(input.ModelID) != "" {
				response["modelId"] = strings.TrimSpace(input.ModelID)
			}
			if strings.TrimSpace(input.PromptPresetID) != "" {
				response["promptPresetId"] = strings.TrimSpace(input.PromptPresetID)
			}
			if generationErr != nil {
				response["fallbackReason"] = generationErr.Error()
			}
			writeJSON(w, http.StatusOK, response)
			return
		}
		if len(parts) == 4 && parts[1] == "prompt-candidates" && r.Method == http.MethodPut {
			var input struct {
				Candidates []map[string]any `json:"candidates"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			if len(batch.Books) == 0 {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "project has no editable source book"})
				return
			}
			kind := strings.TrimSpace(parts[2])
			segments := storedShuihuoSegments(batch.Books[0])
			byID := make(map[string]map[string]any, len(segments))
			for _, segment := range segments {
				if id := modelString(segment, "id"); id != "" {
					byID[id] = segment
				}
			}
			applied, skipped := make([]string, 0), make([]string, 0)
			for _, candidate := range input.Candidates {
				segmentID := modelString(candidate, "segmentId")
				prompt := modelString(candidate, "prompt")
				segment, exists := byID[segmentID]
				if !exists || prompt == "" {
					continue
				}
				lockKey := kind + "PromptLocked"
				if locked, ok := segment[lockKey].(bool); ok && locked {
					skipped = append(skipped, segmentID)
					continue
				}
				if kind == "video" {
					segment["videoPrompt"] = prompt
				} else if kind == "image" {
					segment["imagePrompt"] = prompt
				} else {
					segment["negativePrompt"] = prompt
				}
				applied = append(applied, segmentID)
			}
			if len(applied) > 0 {
				if saveErr := saveShuihuoSegments(r, store, owner, batch, segments); saveErr != nil {
					writeStoreError(w, saveErr)
					return
				}
			}
			writeJSON(w, http.StatusOK, map[string]any{"appliedSegmentIds": applied, "skippedLockedSegmentIds": skipped})
			return
		}
		if len(parts) == 2 && parts[1] == "export" && r.Method == http.MethodGet {
			if files == nil || len(batch.Books) == 0 {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "导出素材存储不可用"})
				return
			}
			if err := writeShuihuoProjectExport(w, r, files, batch); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "项目导出失败"})
			}
			return
		}
		if len(parts) == 2 && parts[1] == "media" && r.Method == http.MethodPost {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "Go 本地迁移阶段暂未接入媒体素材存储"})
			return
		}
		if len(parts) == 3 && parts[1] == "assets" && parts[2] == "generate" && r.Method == http.MethodPost {
			if files == nil || len(batch.Books) == 0 {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "asset image storage is unavailable"})
				return
			}
			var input struct {
				AssetIDs    []string `json:"assetIds"`
				ModelID     string   `json:"modelId"`
				AspectRatio string   `json:"aspectRatio,omitempty"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			images, generationErr := generateShuihuoAssetImagesWithPresetCatalog(r.Context(), store, files, owner, batch, input.AssetIDs, input.ModelID, input.AspectRatio, "", localExecutors, presetCatalog, catalogStore)
			if generationErr != nil {
				writeStoreError(w, generationErr)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]any{"images": images})
			return
		}
		if len(parts) == 2 && parts[1] == "assets" && r.Method == http.MethodPost {
			var input struct {
				Category         string `json:"category"`
				Kind             string `json:"kind"`
				Name             string `json:"name"`
				FormalName       string `json:"formalName"`
				Facts            string `json:"facts"`
				Appearance       string `json:"appearance"`
				Relationships    string `json:"relationships"`
				Prompt           string `json:"prompt"`
				ImageModelID     string `json:"imageModelId"`
				ImageAspectRatio string `json:"imageAspectRatio"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			if len(batch.Books) == 0 || strings.TrimSpace(input.Name) == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "asset name is required"})
				return
			}
			kind := strings.TrimSpace(input.Kind)
			if kind == "" {
				kind = strings.TrimSpace(input.Category)
			}
			if kind == "" {
				kind = "character"
			}
			if _, err := store.CreateBookAsset(r.Context(), owner, batch.ID, batch.Books[0].ID, batchfactoryv11.CreateBookAssetInput{Kind: kind, Name: strings.TrimSpace(input.Name), FormalName: input.FormalName, Facts: input.Facts, Appearance: input.Appearance, Relationships: input.Relationships, Prompt: strings.TrimSpace(input.Prompt), ImageModelID: input.ImageModelID, ImageAspectRatio: input.ImageAspectRatio}); err != nil {
				writeStoreError(w, err)
				return
			}
			updated, err := store.GetBatch(r.Context(), owner, batch.ID)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, shuihuoProjectDetail(updated))
			return
		}
		if len(parts) == 3 && parts[1] == "assets" && parts[2] == "image" && r.Method == http.MethodPost {
			if files == nil || len(batch.Books) == 0 {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "asset image storage is unavailable"})
				return
			}
			var input struct {
				Category         string `json:"category"`
				Kind             string `json:"kind"`
				Name             string `json:"name"`
				FormalName       string `json:"formalName"`
				Facts            string `json:"facts"`
				Appearance       string `json:"appearance"`
				Relationships    string `json:"relationships"`
				Prompt           string `json:"prompt"`
				ImageModelID     string `json:"imageModelId"`
				ImageAspectRatio string `json:"imageAspectRatio"`
				DataURL          string `json:"dataUrl"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			mediaType, payload, err := decodeImageDataURL(input.DataURL)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid image upload"})
				return
			}
			kind := strings.TrimSpace(input.Kind)
			if kind == "" {
				kind = strings.TrimSpace(input.Category)
			}
			if kind == "" {
				kind = "character"
			}
			asset, err := store.CreateBookAsset(r.Context(), owner, batch.ID, batch.Books[0].ID, batchfactoryv11.CreateBookAssetInput{Kind: kind, Name: strings.TrimSpace(input.Name), FormalName: input.FormalName, Facts: input.Facts, Appearance: input.Appearance, Relationships: input.Relationships, Prompt: strings.TrimSpace(input.Prompt), ImageModelID: input.ImageModelID, ImageAspectRatio: input.ImageAspectRatio})
			if err != nil {
				writeStoreError(w, err)
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
			if _, err := store.CreateBookAssetImage(r.Context(), owner, batch.ID, batch.Books[0].ID, asset.ID, batchfactoryv11.CreateBookAssetImageInput{StorageRef: saved.StorageRef, MediaType: saved.MediaType, Source: "upload"}); err != nil {
				_ = files.Remove(saved.StorageRef)
				writeStoreError(w, err)
				return
			}
			updated, err := store.GetBatch(r.Context(), owner, batch.ID)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, shuihuoProjectDetail(updated))
			return
		}
		if len(parts) == 2 && parts[1] == "segmentation" && r.Method == http.MethodPost {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "segmentation action is required"})
			return
		}
		if len(parts) == 2 && parts[1] == "segmentation" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "segmentation action is required"})
			return
		}
		if len(parts) == 3 && parts[1] == "segmentation" && parts[2] == "smart" && r.Method == http.MethodPost {
			var input struct {
				Text    string `json:"text,omitempty"`
				ModelID string `json:"modelId,omitempty"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			candidates, strategy, modelErr := smartShuihuoCandidates(r.Context(), textProviderForShuihuo(director, textProvider), batch, strings.TrimSpace(input.Text))
			response := map[string]any{"candidates": candidates, "mode": "smart", "strategy": strategy}
			if strings.TrimSpace(input.ModelID) != "" {
				response["modelId"] = strings.TrimSpace(input.ModelID)
			}
			if modelErr != nil {
				response["fallbackReason"] = modelErr.Error()
			}
			writeJSON(w, http.StatusOK, response)
			return
		}
		if len(parts) == 3 && parts[1] == "segmentation" && (parts[2] == "paragraphs" || parts[2] == "fixed" || parts[2] == "import") && r.Method == http.MethodPost {
			var input struct {
				Text            string `json:"text,omitempty"`
				LinesPerSegment int    `json:"linesPerSegment,omitempty"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			if len(batch.Books) == 0 {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "project has no editable source book"})
				return
			}
			text := strings.TrimSpace(input.Text)
			if text == "" && len(batch.Books) > 0 {
				text = strings.TrimSpace(batch.Books[0].SourceText)
			}
			var candidates []map[string]any
			if parts[2] == "fixed" {
				candidates = fixedShuihuoCandidates(batch.Books[0].ID, text, input.LinesPerSegment)
			} else {
				candidates = paragraphCandidatesForText(batch.Books[0].ID, text)
			}
			writeJSON(w, http.StatusOK, map[string]any{"candidates": candidates, "mode": parts[2], "strategy": parts[2]})
			return
		}
		if len(parts) == 3 && parts[1] == "segmentation" && parts[2] == "confirm" && r.Method == http.MethodPost {
			var input struct {
				Candidates []map[string]any `json:"candidates"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			if len(batch.Books) == 0 || len(input.Candidates) == 0 {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "candidates are required"})
				return
			}
			segments := normalizeShuihuoSegments(input.Candidates, batch.Books[0].ID)
			metadata := cloneMetadata(batch.Books[0].SourceMetadata)
			metadata["shuihuoSegments"] = segments
			metadata["shuihuoSegmentationStatus"] = "confirmed"
			if _, err := store.UpdateBookMetadata(r.Context(), owner, batch.ID, batch.Books[0].ID, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata, ExpectedRevision: batch.Books[0].Revision}); err != nil {
				writeStoreError(w, err)
				return
			}
			updated, err := store.GetBatch(r.Context(), owner, batch.ID)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, shuihuoProjectDetail(updated))
			return
		}
		if len(parts) == 2 && parts[1] == "source" && (r.Method == http.MethodPut || r.Method == http.MethodPost) {
			var input struct {
				SourceText       string `json:"sourceText"`
				SegmentationMode string `json:"segmentationMode,omitempty"`
				SourceTitle      string `json:"sourceTitle,omitempty"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			if strings.TrimSpace(input.SourceText) == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "sourceText is required"})
				return
			}
			if len(batch.Books) == 0 {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "project has no editable source book"})
				return
			}
			book := batch.Books[0]
			updated, err := store.CaptureBookSource(r.Context(), owner, batch.ID, book.ID, batchfactoryv11.CaptureBookSourceInput{
				SourceText: strings.TrimSpace(input.SourceText), ReplaceSource: true, SourceTitle: input.SourceTitle, ExpectedRevision: book.Revision,
			})
			if err != nil {
				writeStoreError(w, err)
				return
			}
			batch.Books[0] = updated
			project := shuihuoProject(batch)
			project["segmentationMode"] = input.SegmentationMode
			writeJSON(w, http.StatusOK, shuihuoProjectDetailFromProject(project))
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "project route not found"})
	})))

	root.Handle("/api/shuihuo-production/segments/", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		segmentPath := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/shuihuo-production/segments/"), "/")
		voiceSettingsRoute := strings.HasSuffix(segmentPath, "/voice-settings")
		assetsRoute := strings.HasSuffix(segmentPath, "/assets")
		mergeUpRoute := strings.HasSuffix(segmentPath, "/merge-up")
		splitRoute := strings.HasSuffix(segmentPath, "/split")
		insertAfterRoute := strings.HasSuffix(segmentPath, "/insert-after")
		if voiceSettingsRoute {
			segmentPath = strings.TrimSuffix(segmentPath, "/voice-settings")
		}
		for _, suffix := range []string{"/assets", "/merge-up", "/split", "/insert-after"} {
			segmentPath = strings.TrimSuffix(segmentPath, suffix)
		}
		segmentID := strings.Trim(segmentPath, "/")
		if segmentID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "segment id is required"})
			return
		}
		batches, err := store.ListBatches(r.Context(), owner)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		for _, batch := range batches {
			// ListBatches may return compact book records without source metadata.
			// Always hydrate the batch because editable segments live in that metadata.
			full, getErr := store.GetBatch(r.Context(), owner, batch.ID)
			if getErr != nil {
				continue
			}
			batch = full
			if len(batch.Books) == 0 {
				continue
			}
			segments := storedShuihuoSegments(batch.Books[0])
			index := -1
			for i, segment := range segments {
				if id, _ := segment["id"].(string); id == segmentID {
					index = i
					break
				}
			}
			if index < 0 {
				continue
			}
			if assetsRoute && r.Method == http.MethodPut {
				var input struct {
					AssetIDs []string `json:"assetIds"`
				}
				if !decodeJSON(w, r, &input) {
					return
				}
				metadata := cloneMetadata(batch.Books[0].SourceMetadata)
				bindings := map[string][]string{}
				if raw, ok := metadata["shuihuoSegmentAssetIDs"]; ok {
					encoded, _ := json.Marshal(raw)
					_ = json.Unmarshal(encoded, &bindings)
				}
				bindings[segmentID] = append([]string(nil), input.AssetIDs...)
				metadata["shuihuoSegmentAssetIDs"] = bindings
				if _, saveErr := store.UpdateBookMetadata(r.Context(), owner, batch.ID, batch.Books[0].ID, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata, ExpectedRevision: batch.Books[0].Revision}); saveErr != nil {
					writeStoreError(w, saveErr)
					return
				}
				updated, getErr := store.GetBatch(r.Context(), owner, batch.ID)
				if getErr != nil {
					writeStoreError(w, getErr)
					return
				}
				writeJSON(w, http.StatusOK, shuihuoProjectDetail(updated))
				return
			}
			if mergeUpRoute && r.Method == http.MethodPost {
				if index == 0 {
					writeJSON(w, http.StatusConflict, map[string]string{"error": "已经是第一条分镜"})
					return
				}
				previous := segments[index-1]
				previous["sourceText"] = strings.TrimSpace(modelString(previous, "sourceText") + "\n" + modelString(segments[index], "sourceText"))
				previous["subtitleText"] = strings.TrimSpace(modelString(previous, "subtitleText") + "\n" + modelString(segments[index], "subtitleText"))
				unitIDs := storedShuihuoSegmentSourceUnitIDs(batch)
				unitIDs[modelString(previous, "id")] = append(unitIDs[modelString(previous, "id")], unitIDs[segmentID]...)
				delete(unitIDs, segmentID)
				segments = append(segments[:index], segments[index+1:]...)
				if err := saveShuihuoSegments(r, store, owner, batch, segments); err != nil {
					writeStoreError(w, err)
					return
				}
				if fresh, freshErr := store.GetBatch(r.Context(), owner, batch.ID); freshErr == nil {
					batch = fresh
				} else {
					writeStoreError(w, freshErr)
					return
				}
				if err := saveShuihuoSegmentSourceUnitIDs(r, store, owner, batch, unitIDs); err != nil {
					writeStoreError(w, err)
					return
				}
				updated, getErr := store.GetBatch(r.Context(), owner, batch.ID)
				if getErr != nil {
					writeStoreError(w, getErr)
					return
				}
				writeJSON(w, http.StatusOK, shuihuoProjectDetail(updated))
				return
			}
			if splitRoute && r.Method == http.MethodPost {
				lines := splitShuihuoSentences(modelString(segments[index], "sourceText"))
				if len(lines) < 2 {
					writeJSON(w, http.StatusConflict, map[string]string{"error": "当前分镜没有可拆分的原文"})
					return
				}
				replacement := normalizeShuihuoSegments(stringsToShuihuoCandidates(lines), batch.Books[0].ID)
				unitIDs := storedShuihuoSegmentSourceUnitIDs(batch)
				delete(unitIDs, segmentID)
				for _, item := range replacement {
					unitIDs[modelString(item, "id")] = []string{modelString(item, "id")}
				}
				segments = append(append(append([]map[string]any{}, segments[:index]...), replacement...), segments[index+1:]...)
				if err := saveShuihuoSegments(r, store, owner, batch, segments); err != nil {
					writeStoreError(w, err)
					return
				}
				if fresh, freshErr := store.GetBatch(r.Context(), owner, batch.ID); freshErr == nil {
					batch = fresh
				} else {
					writeStoreError(w, freshErr)
					return
				}
				if err := saveShuihuoSegmentSourceUnitIDs(r, store, owner, batch, unitIDs); err != nil {
					writeStoreError(w, err)
					return
				}
				updated, getErr := store.GetBatch(r.Context(), owner, batch.ID)
				if getErr != nil {
					writeStoreError(w, getErr)
					return
				}
				writeJSON(w, http.StatusOK, shuihuoProjectDetail(updated))
				return
			}
			if insertAfterRoute && r.Method == http.MethodPost {
				var input map[string]any
				if !decodeJSON(w, r, &input) {
					return
				}
				text := strings.TrimSpace(modelString(input, "sourceText"))
				if text == "" {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "sourceText is required"})
					return
				}
				created := normalizeShuihuoSegments([]map[string]any{{"text": text, "subtitleText": modelString(input, "subtitleText")}}, batch.Books[0].ID)
				segments = append(segments, nil)
				copy(segments[index+2:], segments[index+1:])
				segments[index+1] = created[0]
				unitIDs := storedShuihuoSegmentSourceUnitIDs(batch)
				unitIDs[modelString(created[0], "id")] = []string{modelString(created[0], "id")}
				if err := saveShuihuoSegments(r, store, owner, batch, segments); err != nil {
					writeStoreError(w, err)
					return
				}
				if fresh, freshErr := store.GetBatch(r.Context(), owner, batch.ID); freshErr == nil {
					batch = fresh
				} else {
					writeStoreError(w, freshErr)
					return
				}
				if err := saveShuihuoSegmentSourceUnitIDs(r, store, owner, batch, unitIDs); err != nil {
					writeStoreError(w, err)
					return
				}
				updated, getErr := store.GetBatch(r.Context(), owner, batch.ID)
				if getErr != nil {
					writeStoreError(w, getErr)
					return
				}
				writeJSON(w, http.StatusOK, shuihuoProjectDetail(updated))
				return
			}
			switch r.Method {
			case http.MethodPut:
				var input map[string]any
				if !decodeJSON(w, r, &input) {
					return
				}
				keys := []string{"sourceText", "subtitleText", "speaker", "imagePrompt", "videoPrompt"}
				if voiceSettingsRoute {
					keys = []string{"voiceAssetId", "speechRate", "pitch"}
				}
				for _, key := range keys {
					if value, exists := input[key]; exists {
						segments[index][key] = value
					}
				}
				if err := saveShuihuoSegments(r, store, owner, batch, segments); err != nil {
					writeStoreError(w, err)
					return
				}
				if voiceSettingsRoute {
					writeJSON(w, http.StatusOK, map[string]any{
						"voiceAssetId": segments[index]["voiceAssetId"],
						"speechRate":   segments[index]["speechRate"],
						"pitch":        segments[index]["pitch"],
					})
					return
				}
			case http.MethodDelete:
				segments = append(segments[:index], segments[index+1:]...)
				if err := saveShuihuoSegments(r, store, owner, batch, segments); err != nil {
					writeStoreError(w, err)
					return
				}
			default:
				writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
				return
			}
			updated, err := store.GetBatch(r.Context(), owner, batch.ID)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, shuihuoProjectDetail(updated))
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "segment not found"})
	})))

	root.Handle("/api/shuihuo-production/assets/", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/shuihuo-production/assets/"), "/")
		parts := strings.Split(path, "/")
		if len(parts) == 0 || parts[0] == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "asset id is required"})
			return
		}
		assetID := parts[0]
		batches, err := store.ListBatches(r.Context(), owner)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		for _, compact := range batches {
			batch, getErr := store.GetBatch(r.Context(), owner, compact.ID)
			if getErr != nil {
				continue
			}
			if len(batch.Books) == 0 {
				continue
			}
			var asset *batchfactoryv11.BookAsset
			for index := range batch.Books[0].AssetRecords {
				if batch.Books[0].AssetRecords[index].ID == assetID {
					asset = &batch.Books[0].AssetRecords[index]
					break
				}
			}
			if asset == nil {
				continue
			}
			if len(parts) == 2 && parts[1] == "image" && (r.Method == http.MethodGet || r.Method == http.MethodPut) {
				if r.Method == http.MethodPut {
					if files == nil {
						writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "asset image storage is unavailable"})
						return
					}
					var input struct {
						DataURL string `json:"dataUrl"`
					}
					if !decodeJSON(w, r, &input) {
						return
					}
					mediaType, payload, decodeErr := decodeShuihuoMediaDataURL(input.DataURL)
					if decodeErr != nil || !strings.HasPrefix(mediaType, "image/") {
						writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid image upload"})
						return
					}
					artifactID, idErr := localexecutor.NewArtifactID()
					if idErr != nil {
						writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "image id generation failed"})
						return
					}
					saved, saveErr := files.SaveImage(artifactID, mediaType, bytes.NewReader(payload))
					if saveErr != nil {
						writeBookAssetImageStorageError(w, saveErr)
						return
					}
					created, createErr := store.CreateBookAssetImage(r.Context(), owner, batch.ID, batch.Books[0].ID, assetID, batchfactoryv11.CreateBookAssetImageInput{StorageRef: saved.StorageRef, MediaType: saved.MediaType, Source: "upload"})
					if createErr != nil {
						_ = files.Remove(saved.StorageRef)
						writeStoreError(w, createErr)
						return
					}
					primary, primaryErr := store.SetPrimaryBookAssetImage(r.Context(), owner, batch.ID, batch.Books[0].ID, assetID, created.ID)
					if primaryErr != nil {
						writeStoreError(w, primaryErr)
						return
					}
					writeJSON(w, http.StatusOK, map[string]any{"image": bookAssetImageResponse(primary, batch.ID, batch.Books[0].ID, assetID)})
					return
				}
				if files == nil {
					writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "asset image storage is unavailable"})
					return
				}
				images, listErr := store.ListBookAssetImages(r.Context(), owner, batch.ID, batch.Books[0].ID, assetID)
				if listErr != nil {
					writeStoreError(w, listErr)
					return
				}
				var selected *batchfactoryv11.BookAssetImage
				for index := range images {
					if images[index].IsPrimary || selected == nil {
						selected = &images[index]
					}
				}
				if selected == nil || strings.TrimSpace(selected.StorageRef) == "" {
					writeJSON(w, http.StatusNotFound, map[string]string{"error": "asset image file not found"})
					return
				}
				file, openErr := files.Open(selected.StorageRef)
				if openErr != nil {
					writeJSON(w, http.StatusNotFound, map[string]string{"error": "asset image file not found"})
					return
				}
				defer file.Close()
				w.Header().Set("Content-Type", selected.MediaType)
				http.ServeContent(w, r, selected.ID, selected.CreatedAt, file)
				return
			}
			if len(parts) == 2 && parts[1] == "images" && r.Method == http.MethodGet {
				images, listErr := store.ListBookAssetImages(r.Context(), owner, batch.ID, batch.Books[0].ID, assetID)
				if listErr != nil {
					writeStoreError(w, listErr)
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"images": bookAssetImageResponses(images, batch.ID, batch.Books[0].ID, assetID)})
				return
			}
			if len(parts) == 1 && r.Method == http.MethodPut {
				var input struct {
					Name             string `json:"name"`
					FormalName       string `json:"formalName"`
					Facts            string `json:"facts"`
					Appearance       string `json:"appearance"`
					Relationships    string `json:"relationships"`
					Prompt           string `json:"prompt"`
					ImageModelID     string `json:"imageModelId"`
					ImageAspectRatio string `json:"imageAspectRatio"`
					ExpectedRevision int64  `json:"expectedRevision"`
				}
				if !decodeJSON(w, r, &input) {
					return
				}
				if input.ExpectedRevision == 0 {
					input.ExpectedRevision = asset.Revision
				}
				updated, updateErr := store.UpdateBookAsset(r.Context(), owner, batch.ID, batch.Books[0].ID, assetID, batchfactoryv11.UpdateBookAssetInput{Name: input.Name, FormalName: input.FormalName, Facts: input.Facts, Appearance: input.Appearance, Relationships: input.Relationships, Prompt: input.Prompt, ImageModelID: input.ImageModelID, ImageAspectRatio: input.ImageAspectRatio, ExpectedRevision: input.ExpectedRevision})
				if updateErr != nil {
					writeStoreError(w, updateErr)
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"asset": updated})
				return
			}
			if len(parts) == 1 && r.Method == http.MethodDelete {
				if deleteErr := store.DeleteBookAsset(r.Context(), owner, batch.ID, batch.Books[0].ID, assetID); deleteErr != nil {
					writeStoreError(w, deleteErr)
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"ok": true})
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "asset not found"})
	})))

	root.Handle("/api/shuihuo-production/asset-images/", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		imageID := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/shuihuo-production/asset-images/"), "/")
		action := "primary"
		if strings.HasSuffix(imageID, "/download") {
			action = "download"
			imageID = strings.TrimSuffix(imageID, "/download")
		} else if strings.HasSuffix(imageID, "/primary") {
			imageID = strings.TrimSuffix(imageID, "/primary")
		} else {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "asset image route not found"})
			return
		}
		batches, err := store.ListBatches(r.Context(), owner)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		for _, compact := range batches {
			batch, getErr := store.GetBatch(r.Context(), owner, compact.ID)
			if getErr != nil || len(batch.Books) == 0 {
				continue
			}
			for _, asset := range batch.Books[0].AssetRecords {
				images, listErr := store.ListBookAssetImages(r.Context(), owner, batch.ID, batch.Books[0].ID, asset.ID)
				if listErr != nil {
					continue
				}
				for _, image := range images {
					if image.ID != imageID {
						continue
					}
					if action == "download" {
						if r.Method != http.MethodGet || files == nil || strings.TrimSpace(image.StorageRef) == "" {
							writeJSON(w, http.StatusNotFound, map[string]string{"error": "asset image file not found"})
							return
						}
						file, openErr := files.Open(image.StorageRef)
						if openErr != nil {
							writeJSON(w, http.StatusNotFound, map[string]string{"error": "asset image file not found"})
							return
						}
						defer file.Close()
						w.Header().Set("Content-Type", image.MediaType)
						http.ServeContent(w, r, image.ID, image.CreatedAt, file)
						return
					}
					if r.Method != http.MethodPut {
						writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
						return
					}
					primary, primaryErr := store.SetPrimaryBookAssetImage(r.Context(), owner, batch.ID, batch.Books[0].ID, asset.ID, imageID)
					if primaryErr != nil {
						writeStoreError(w, primaryErr)
						return
					}
					writeJSON(w, http.StatusOK, map[string]any{"image": bookAssetImageResponse(primary, batch.ID, batch.Books[0].ID, asset.ID)})
					return
				}
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "asset image not found"})
	})))

	root.Handle("/api/shuihuo-production/media/", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/shuihuo-production/media/"), "/")
		parts := strings.Split(path, "/")
		if len(parts) == 0 || parts[0] == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "media id is required"})
			return
		}
		mediaID := parts[0]
		batches, err := store.ListBatches(r.Context(), owner)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		for _, compact := range batches {
			batch, getErr := store.GetBatch(r.Context(), owner, compact.ID)
			if getErr != nil || len(batch.Books) == 0 {
				continue
			}
			media := storedShuihuoMedia(batch.Books[0])
			index := -1
			for i, item := range media {
				if id, _ := item["id"].(string); id == mediaID {
					index = i
					break
				}
			}
			if index < 0 {
				continue
			}
			if len(parts) == 2 && parts[1] == "download" && r.Method == http.MethodGet {
				if files == nil {
					writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "media storage is unavailable"})
					return
				}
				storageRef, _ := media[index]["storageRef"].(string)
				file, openErr := files.Open(storageRef)
				if openErr != nil {
					writeJSON(w, http.StatusNotFound, map[string]string{"error": "media file not found"})
					return
				}
				defer file.Close()
				mediaType, _ := media[index]["mediaType"].(string)
				if mediaType == "" {
					mediaType = "application/octet-stream"
				}
				w.Header().Set("Content-Type", mediaType)
				http.ServeContent(w, r, mediaID, time.Now(), file)
				return
			}
			if len(parts) == 2 && parts[1] == "segment" && r.Method == http.MethodPut {
				var input struct {
					SegmentID string `json:"segmentId"`
				}
				if !decodeJSON(w, r, &input) {
					return
				}
				media[index]["segmentId"] = strings.TrimSpace(input.SegmentID)
			} else if len(parts) == 2 && parts[1] == "primary" && r.Method == http.MethodPut {
				for i := range media {
					media[i]["isPrimary"] = i == index
				}
			} else if len(parts) == 1 && r.Method == http.MethodDelete {
				if files != nil {
					if storageRef, _ := media[index]["storageRef"].(string); storageRef != "" {
						_ = files.Remove(storageRef)
					}
				}
				media = append(media[:index], media[index+1:]...)
			} else if len(parts) == 1 {
				writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
				return
			} else {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "media route not found"})
				return
			}
			metadata := cloneMetadata(batch.Books[0].SourceMetadata)
			metadata["shuihuoMedia"] = media
			if _, err := store.UpdateBookMetadata(r.Context(), owner, batch.ID, batch.Books[0].ID, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata, ExpectedRevision: batch.Books[0].Revision}); err != nil {
				writeStoreError(w, err)
				return
			}
			updated, err := store.GetBatch(r.Context(), owner, batch.ID)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, shuihuoProjectDetail(updated))
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "media not found"})
	})))

	// The legacy task drawer addresses tasks by task id, while the Go
	// production service cancels by the owning batch and request id. Resolve
	// that mapping here so the drawer no longer falls into a 404 hole.
	root.Handle("/api/shuihuo-production/tasks/", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if production == nil {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "当前 Go 本地环境尚未配置可用的生成模型"})
			return
		}
		path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/shuihuo-production/tasks/"), "/")
		parts := strings.Split(path, "/")
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "task route not found"})
			return
		}
		taskID, action := parts[0], parts[1]
		if action == "retry" && r.Method == http.MethodPost {
			if director == nil {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "当前 Go 本地环境尚未配置导演服务，暂不能重试任务"})
				return
			}
			batches, err := store.ListBatches(r.Context(), owner)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			for _, compact := range batches {
				status, statusErr := production.GetBatchStatus(r.Context(), owner, compact.ID)
				if statusErr != nil {
					continue
				}
				for _, job := range status.Jobs {
					for _, task := range job.Tasks {
						if task.ID != taskID {
							continue
						}
						stages := &batchfactoryv11.BookStageService{Store: store, Director: director, Production: production}
						summary, retryErr := stages.RetryLastFailed(r.Context(), owner, compact.ID, job.BookID, "shuihuo-retry-"+time.Now().UTC().Format("20060102T150405.000000000"), task.VideoID)
						if retryErr != nil {
							writeStoreError(w, retryErr)
							return
						}
						writeJSON(w, http.StatusOK, summary)
						return
					}
				}
			}
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
			return
		}
		if action != "cancel" || r.Method != http.MethodPut {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		batches, err := store.ListBatches(r.Context(), owner)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		for _, compact := range batches {
			status, statusErr := production.GetBatchStatus(r.Context(), owner, compact.ID)
			if statusErr != nil {
				continue
			}
			for _, job := range status.Jobs {
				for _, task := range job.Tasks {
					if task.ID != taskID {
						continue
					}
					result, cancelErr := production.CancelBatch(r.Context(), owner, compact.ID, job.RequestID)
					if cancelErr != nil {
						writeStoreError(w, cancelErr)
						return
					}
					writeJSON(w, http.StatusOK, result)
					return
				}
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
	})))
}

func isShuihuoProjectBatch(batch batchfactoryv11.Batch) bool {
	value, ok := batch.SettingsState.Patch["shuihuoProject"]
	return ok && strings.TrimSpace(string(value)) == "true"
}

func shuihuoDefaultConfig() map[string]any {
	return map[string]any{
		"characterPrefix": "", "imagePrefix": "", "imageSuffix": "", "videoPrefix": "", "videoSuffix": "",
		"textModelId": nil, "imageModelId": nil, "videoModelId": nil, "audioModelId": nil, "jianyingDraftDirectory": "",
	}
}

func shuihuoProject(batch batchfactoryv11.Batch) map[string]any {
	createdAt := batch.CreatedAt.UTC().Format(time.RFC3339)
	updatedAt := batch.UpdatedAt.UTC().Format(time.RFC3339)
	status := "pending"
	var sourceText string
	segments := []map[string]any{}
	segmentAssetIDs := map[string][]string{}
	segmentSourceUnitIDs := map[string][]string{}
	if len(batch.Books) > 0 {
		sourceText = batch.Books[0].SourceText
		segments = storedShuihuoSegments(batch.Books[0])
		if raw, ok := batch.Books[0].SourceMetadata["shuihuoSegmentAssetIDs"]; ok {
			encoded, _ := json.Marshal(raw)
			_ = json.Unmarshal(encoded, &segmentAssetIDs)
		}
		segmentSourceUnitIDs = storedShuihuoSegmentSourceUnitIDs(batch)
		if len(segments) > 0 {
			status = "confirmed"
		}
	}
	return map[string]any{
		"id": batch.ID, "name": batch.Title, "createdAt": createdAt, "updatedAt": updatedAt,
		"sourceText": sourceText, "segmentationStatus": status, "segments": segments, "assets": shuihuoAssets(batch), "segmentAssetIDs": segmentAssetIDs, "segmentSourceUnitIDs": segmentSourceUnitIDs,
		"media": shuihuoMedia(batch),
		"books": batch.Books,
	}
}

func shuihuoAssets(batch batchfactoryv11.Batch) []map[string]any {
	if len(batch.Books) == 0 {
		return []map[string]any{}
	}
	assets := make([]map[string]any, 0, len(batch.Books[0].AssetRecords))
	for _, asset := range batch.Books[0].AssetRecords {
		assets = append(assets, map[string]any{
			"id": asset.ID, "category": asset.Kind, "kind": asset.Kind, "name": asset.Name, "prompt": asset.Prompt,
			"imageModelId": asset.ImageModelID, "imageAspectRatio": asset.ImageAspectRatio,
			"source": asset.Source, "revision": asset.Revision, "createdAt": asset.CreatedAt, "updatedAt": asset.UpdatedAt,
		})
	}
	return assets
}

// The studio page consumes a single read model rather than the compact card
// returned by the project library. Keep the envelope stable even while the
// Go migration is filling in the remaining segmentation/media stores.
func shuihuoProjectDetail(batch batchfactoryv11.Batch) map[string]any {
	return shuihuoProjectDetailFromProject(shuihuoProject(batch))
}

func shuihuoProjectDetailFromProject(project map[string]any) map[string]any {
	segments, _ := project["segments"].([]map[string]any)
	assets, _ := project["assets"].([]map[string]any)
	media, _ := project["media"].([]map[string]any)
	segmentAssetIDs, _ := project["segmentAssetIDs"].(map[string][]string)
	if segmentAssetIDs == nil {
		segmentAssetIDs = map[string][]string{}
	}
	segmentSourceUnitIDs, _ := project["segmentSourceUnitIDs"].(map[string][]string)
	if segmentSourceUnitIDs == nil {
		segmentSourceUnitIDs = map[string][]string{}
	}
	return map[string]any{
		"project":              project,
		"segments":             segments,
		"assets":               assets,
		"media":                media,
		"segmentAssetIDs":      segmentAssetIDs,
		"segmentSourceUnitIDs": segmentSourceUnitIDs,
		"settings":             map[string]any{},
	}
}

func shuihuoMedia(batch batchfactoryv11.Batch) []map[string]any {
	if len(batch.Books) == 0 {
		return []map[string]any{}
	}
	return storedShuihuoMedia(batch.Books[0])
}

func loadShuihuoOwnerConfig(ctx context.Context, store batchfactoryv11.Store, owner string, catalogStores ...novelfetchworkshop.Store) (map[string]any, error) {
	var config map[string]any
	if durable, ok := store.(batchfactoryv11.ShuihuoConfigStore); ok {
		loaded, err := durable.GetShuihuoConfig(ctx, owner)
		if err != nil {
			return nil, err
		}
		config = loaded
	} else if value, ok := shuihuoProductionConfigs.Load(owner); ok {
		if loaded, ok := value.(map[string]any); ok {
			config = loaded
		}
	}
	if config == nil {
		config = shuihuoDefaultConfig()
		shuihuoProductionConfigs.Store(owner, config)
	}
	if len(catalogStores) > 0 && catalogStores[0] != nil {
		if managed, err := catalogStores[0].GetConfig(ctx, owner); err == nil {
			if catalog, ok := managed["modelCatalog"].([]any); ok && len(catalog) > 0 {
				config["modelCatalog"] = catalog
			}
			if version, ok := managed["modelCatalogVersion"]; ok {
				config["modelCatalogVersion"] = version
			}
		}
	}
	// Older personal API settings stored the image provider in the top-level
	// `image` object instead of modelCatalog. Keep that configuration usable by
	// the newer batch/comic pipeline so the UI selection and the server check
	// cannot disagree about whether image generation is configured.
	ensureLegacyShuihuoImageModel(config)
	return config, nil
}

func ensureLegacyShuihuoImageModel(config map[string]any) {
	if config == nil {
		return
	}
	for _, model := range shuihuoCollection(config, "modelCatalog", nil) {
		if strings.EqualFold(modelString(model, "kind"), "image") {
			if enabled, exists := model["enabled"].(bool); exists && !enabled {
				continue
			}
			return
		}
	}
	legacy, ok := config["image"].(map[string]any)
	if !ok {
		return
	}
	modelID := firstText(legacy["model"], legacy["modelId"], legacy["model_id"])
	baseURL := firstText(legacy["endpoint"], legacy["baseUrl"], legacy["base_url"])
	credential := firstText(legacy["apiKey"], legacy["api_key"], legacy["credential"])
	if modelID == "" || baseURL == "" || credential == "" {
		return
	}
	model := map[string]any{
		"id":          "legacy-account-image",
		"kind":        "image",
		"displayName": firstText(legacy["displayName"], "个人 API 图片模型"),
		"modelId":     modelID,
		"baseUrl":     baseURL,
		"credential":  credential,
		"enabled":     true,
	}
	config["modelCatalog"] = append(shuihuoCollection(config, "modelCatalog", nil), model)
}

func saveShuihuoOwnerConfig(ctx context.Context, store batchfactoryv11.Store, owner string, config map[string]any) error {
	if durable, ok := store.(batchfactoryv11.ShuihuoConfigStore); ok {
		return durable.SaveShuihuoConfig(ctx, owner, config)
	}
	shuihuoProductionConfigs.Store(owner, config)
	return nil
}

// saveShuihuoAdminModelCatalog keeps the admin catalog and the account model
// catalog in sync. The latter is what the user-facing API configuration and
// runtime model selectors read, so saving only the Shuihuo compatibility blob
// would make an admin edit disappear on the next read.
func saveShuihuoAdminModelCatalog(ctx context.Context, store batchfactoryv11.Store, catalogStore novelfetchworkshop.Store, owner string, config map[string]any, models []map[string]any) error {
	if catalogStore != nil {
		managed, err := catalogStore.GetConfig(ctx, owner)
		if err != nil {
			return err
		}
		managed["modelCatalog"] = models
		if _, ok := managed["modelCatalogVersion"]; !ok {
			managed["modelCatalogVersion"] = 1
		}
		if err := catalogStore.PutConfig(ctx, owner, managed); err != nil {
			return err
		}
	}
	return saveShuihuoOwnerConfig(ctx, store, owner, config)
}

func shuihuoCollection(config map[string]any, key string, fallback []map[string]any) []map[string]any {
	if raw, ok := config[key].([]any); ok {
		items := make([]map[string]any, 0, len(raw))
		for _, value := range raw {
			if item, ok := value.(map[string]any); ok {
				items = append(items, item)
			}
		}
		return items
	}
	if raw, ok := config[key].([]map[string]any); ok {
		return raw
	}
	return append([]map[string]any(nil), fallback...)
}

func safeShuihuoModels(models []map[string]any, includeSecrets bool) []map[string]any {
	return copySafeShuihuoModels(models, []string{"id", "modelId", "name", "displayName", "kind", "provider", "adapterKind", "enabled", "configured", "parameterSchema", "capabilities"}, includeSecrets)
}

func safeShuihuoAdminModels(models []map[string]any) []map[string]any {
	return copySafeShuihuoModels(models, []string{"id", "modelId", "name", "displayName", "kind", "provider", "adapterKind", "enabled", "configured", "parameterSchema", "credentialRef", "endpoint", "baseUrl", "tasksEndpoint", "resultEndpoint", "workflowId", "requestTemplate", "responseMapping", "voiceCloneRequestTemplate", "voiceCloneResponseMapping", "capabilities"}, false)
}

func copySafeShuihuoModels(models []map[string]any, fields []string, includeSecrets bool) []map[string]any {
	result := make([]map[string]any, 0, len(models))
	for _, model := range models {
		copy := map[string]any{}
		for _, key := range fields {
			if value, exists := model[key]; exists {
				copy[key] = value
			}
		}
		if includeSecrets {
			if value, exists := model["credential"]; exists {
				copy["credential"] = value
			}
			if value, exists := model["apiKey"]; exists {
				copy["apiKey"] = value
			}
		}
		if !includeSecrets {
			credentialConfigured := shuihuoModelCredential(model) != ""
			copy["hasCredential"] = credentialConfigured
			copy["credentialConfigured"] = credentialConfigured
			endpointConfigured := modelString(model, "endpoint") != "" || modelString(model, "baseUrl") != "" || strings.EqualFold(modelString(model, "adapterKind"), "local_executor_video")
			copy["providerConfigured"] = modelString(model, "modelId") != "" && endpointConfigured
		}
		result = append(result, copy)
	}
	return result
}

func safeShuihuoConfigForResponse(config map[string]any) map[string]any {
	result := make(map[string]any, len(config))
	for key, value := range config {
		if key != "modelCatalog" {
			result[key] = value
		}
	}
	result["modelCatalog"] = safeShuihuoModels(shuihuoCollection(config, "modelCatalog", nil), false)
	return result
}

func syncShuihuoVideoModel(ctx context.Context, registry batchfactoryv11.VideoProviderRegistry, owner string, model map[string]any) error {
	modelID, _ := model["modelId"].(string)
	if strings.TrimSpace(modelID) == "" && strings.EqualFold(modelString(model, "id"), batchfactoryv11.AutoDLH3Model) {
		modelID = batchfactoryv11.AutoDLH3Model
	}
	provider, _ := model["provider"].(string)
	provider = batchfactoryv11.VideoProviderForModel(modelID, provider)
	if provider == batchfactoryv11.VideoProviderAutoDLH3 && strings.TrimSpace(modelID) == "" {
		modelID = batchfactoryv11.AutoDLH3Model
	}
	apiKey := shuihuoModelCredential(model)
	endpoint, _ := model["endpoint"].(string)
	if strings.TrimSpace(endpoint) == "" {
		endpoint = modelString(model, "baseUrl")
	}
	tasksEndpoint, _ := model["tasksEndpoint"].(string)
	resultEndpoint, _ := model["resultEndpoint"].(string)
	if provider == batchfactoryv11.VideoProviderAutoDLH3 {
		if strings.TrimSpace(endpoint) == "" {
			endpoint = batchfactoryv11.DefaultAutoDLH3CreateURL
		}
		if strings.TrimSpace(tasksEndpoint) == "" {
			tasksEndpoint = batchfactoryv11.DefaultAutoDLH3TasksURL
		}
	}
	workflowID, _ := model["workflowId"].(string)
	return registry.Put(ctx, owner, batchfactoryv11.VideoProviderConfig{Provider: provider, APIKey: apiKey, Model: modelID, CreateURL: endpoint, TasksURL: tasksEndpoint, ResultURL: resultEndpoint, WorkflowID: workflowID})
}

func syncShuihuoConfiguredVideoModels(ctx context.Context, store batchfactoryv11.Store, registry batchfactoryv11.VideoProviderRegistry, owner string, catalogStores ...novelfetchworkshop.Store) error {
	config, err := loadShuihuoOwnerConfig(ctx, store, owner, catalogStores...)
	if err != nil {
		return err
	}
	for _, model := range shuihuoCollection(config, "modelCatalog", nil) {
		kind, _ := model["kind"].(string)
		enabled, enabledSet := model["enabled"].(bool)
		if !strings.EqualFold(strings.TrimSpace(kind), "video") || (enabledSet && !enabled) {
			continue
		}
		apiKey := shuihuoModelCredential(model)
		provider, _ := model["provider"].(string)
		if strings.TrimSpace(apiKey) == "" && batchfactoryv11.VideoProviderForModel(modelString(model, "modelId"), provider) != batchfactoryv11.VideoProviderDoubaoLocal {
			continue
		}
		if err := syncShuihuoVideoModel(ctx, registry, owner, model); err != nil {
			return err
		}
	}
	return nil
}

func modelString(model map[string]any, key string) string {
	value, _ := model[key].(string)
	return strings.TrimSpace(value)
}

func shuihuoModelCredential(model map[string]any) string {
	if value := modelString(model, "apiKey"); value != "" {
		return value
	}
	if value := modelString(model, "credential"); value != "" {
		return value
	}
	if ref := modelString(model, "credentialRef"); ref != "" {
		return strings.TrimSpace(os.Getenv(ref))
	}
	return ""
}

func writeShuihuoProjectExport(w http.ResponseWriter, r *http.Request, files *localartifact.Store, batch batchfactoryv11.Batch) error {
	var archive bytes.Buffer
	zw := zip.NewWriter(&archive)
	book := batch.Books[0]
	writeEntry := func(name string, data []byte) error {
		entry, err := zw.Create(name)
		if err != nil {
			return err
		}
		_, err = entry.Write(data)
		return err
	}
	if err := writeEntry("source.txt", []byte(book.SourceText)); err != nil {
		return err
	}
	segments, err := json.MarshalIndent(storedShuihuoSegments(book), "", "  ")
	if err != nil {
		return err
	}
	if err := writeEntry("segments.json", segments); err != nil {
		return err
	}
	project, err := json.MarshalIndent(shuihuoProjectDetail(batch), "", "  ")
	if err != nil {
		return err
	}
	if err := writeEntry("project.json", project); err != nil {
		return err
	}
	for _, media := range storedShuihuoMedia(book) {
		storageRef, _ := media["storageRef"].(string)
		mediaID, _ := media["id"].(string)
		if strings.TrimSpace(storageRef) == "" || strings.TrimSpace(mediaID) == "" {
			continue
		}
		file, openErr := files.Open(storageRef)
		if openErr != nil {
			continue
		}
		name := "media/" + mediaID
		if filename, ok := media["filename"].(string); ok && strings.TrimSpace(filename) != "" {
			name = "media/" + filepath.Base(filename)
		}
		entry, createErr := zw.Create(name)
		if createErr == nil {
			_, createErr = io.Copy(entry, file)
		}
		_ = file.Close()
		if createErr != nil {
			return createErr
		}
	}
	if err := zw.Close(); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=shuihuo-project-%s.zip", batch.ID))
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(archive.Bytes())
	return err
}

func storedShuihuoMedia(book batchfactoryv11.Book) []map[string]any {
	value, ok := book.SourceMetadata["shuihuoMedia"]
	if !ok {
		return []map[string]any{}
	}
	if items, ok := value.([]map[string]any); ok {
		return items
	}
	if items, ok := value.([]any); ok {
		output := make([]map[string]any, 0, len(items))
		for _, item := range items {
			if media, ok := item.(map[string]any); ok {
				output = append(output, media)
			}
		}
		return output
	}
	return []map[string]any{}
}

func cloneMetadata(input map[string]any) map[string]any {
	output := make(map[string]any, len(input)+2)
	for key, value := range input {
		output[key] = value
	}
	return output
}

func saveShuihuoSegments(r *http.Request, store batchfactoryv11.Store, owner string, batch batchfactoryv11.Batch, segments []map[string]any) error {
	if len(batch.Books) == 0 {
		return fmt.Errorf("project has no editable source book")
	}
	for index, segment := range segments {
		segment["orderIndex"] = index + 1
	}
	metadata := cloneMetadata(batch.Books[0].SourceMetadata)
	metadata["shuihuoSegments"] = segments
	metadata["shuihuoSegmentationStatus"] = "confirmed"
	_, err := store.UpdateBookMetadata(r.Context(), owner, batch.ID, batch.Books[0].ID, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata, ExpectedRevision: batch.Books[0].Revision})
	return err
}

func storedShuihuoSegmentSourceUnitIDs(batch batchfactoryv11.Batch) map[string][]string {
	result := map[string][]string{}
	if len(batch.Books) == 0 {
		return result
	}
	if raw, ok := batch.Books[0].SourceMetadata["shuihuoSegmentSourceUnitIDs"]; ok {
		encoded, _ := json.Marshal(raw)
		_ = json.Unmarshal(encoded, &result)
	}
	for _, segment := range storedShuihuoSegments(batch.Books[0]) {
		id := modelString(segment, "id")
		if id != "" && len(result[id]) == 0 {
			result[id] = []string{id}
		}
	}
	return result
}

func saveShuihuoSegmentSourceUnitIDs(r *http.Request, store batchfactoryv11.Store, owner string, batch batchfactoryv11.Batch, unitIDs map[string][]string) error {
	if len(batch.Books) == 0 {
		return fmt.Errorf("project has no editable source book")
	}
	metadata := cloneMetadata(batch.Books[0].SourceMetadata)
	metadata["shuihuoSegmentSourceUnitIDs"] = unitIDs
	_, err := store.UpdateBookMetadata(r.Context(), owner, batch.ID, batch.Books[0].ID, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata, ExpectedRevision: batch.Books[0].Revision})
	return err
}

func reorderShuihuoSegments(segments []map[string]any, ids []string) []map[string]any {
	byID := make(map[string]map[string]any, len(segments))
	for _, segment := range segments {
		if id, _ := segment["id"].(string); id != "" {
			byID[id] = segment
		}
	}
	ordered := make([]map[string]any, 0, len(segments))
	for _, id := range ids {
		if segment, ok := byID[id]; ok {
			ordered = append(ordered, segment)
			delete(byID, id)
		}
	}
	for _, segment := range segments {
		if id, _ := segment["id"].(string); id != "" {
			if _, ok := byID[id]; ok {
				ordered = append(ordered, segment)
				delete(byID, id)
			}
		}
	}
	return ordered
}

func storedShuihuoAssetGenerationConfig(batch batchfactoryv11.Batch) map[string]any {
	defaults := map[string]any{"textModelId": nil, "imageModelId": nil, "audioModelId": nil, "assetPresetId": "shuihuo-extract-assets", "characterSheetPresetId": nil, "aspectRatio": "16:9"}
	if len(batch.Books) == 0 || batch.Books[0].SourceMetadata == nil {
		return defaults
	}
	raw, ok := batch.Books[0].SourceMetadata["shuihuoAssetGenerationConfig"]
	if !ok {
		return defaults
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return defaults
	}
	var stored map[string]any
	if json.Unmarshal(encoded, &stored) != nil {
		return defaults
	}
	for key, value := range stored {
		defaults[key] = value
	}
	return defaults
}

func storedShuihuoSegments(book batchfactoryv11.Book) []map[string]any {
	if value, ok := book.SourceMetadata["shuihuoSegments"].([]any); ok {
		segments := make([]map[string]any, 0, len(value))
		for _, item := range value {
			if segment, ok := item.(map[string]any); ok {
				segments = append(segments, segment)
			}
		}
		return segments
	}
	if value, ok := book.SourceMetadata["shuihuoSegments"].([]map[string]any); ok {
		return value
	}
	return []map[string]any{}
}

func paragraphCandidates(batch batchfactoryv11.Batch) []map[string]any {
	if len(batch.Books) == 0 {
		return []map[string]any{}
	}
	text := strings.TrimSpace(batch.Books[0].SourceText)
	if text == "" {
		return []map[string]any{}
	}
	lines := strings.FieldsFunc(text, func(r rune) bool { return r == '\n' || r == '\r' })
	if len(lines) == 1 {
		lines = splitShuihuoSentences(lines[0])
	}
	candidates := make([]map[string]any, 0, len(lines))
	for index, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		candidates = append(candidates, map[string]any{"id": fmt.Sprintf("candidate_%s_%d", batch.Books[0].ID, index+1), "text": line, "sourceText": line, "subtitleText": line, "speaker": "旁白"})
	}
	return candidates
}

// smartShuihuoCandidates asks the configured text provider for editable
// storyboard candidates. A deterministic paragraph split remains available
// when the provider is not configured or returns malformed output, so the
// user can still correct and confirm the result instead of seeing a blank
// workbench.
func smartShuihuoCandidates(ctx context.Context, provider batchfactoryv11.DirectorProvider, batch batchfactoryv11.Batch, override string) ([]map[string]any, string, error) {
	if len(batch.Books) == 0 {
		return []map[string]any{}, "paragraph_fallback", fmt.Errorf("project has no source book")
	}
	source := strings.TrimSpace(override)
	if source == "" {
		source = strings.TrimSpace(batch.Books[0].SourceText)
	}
	fallback := func(err error) ([]map[string]any, string, error) {
		return paragraphCandidatesForText(batch.Books[0].ID, source), "paragraph_fallback", err
	}
	if provider == nil {
		return fallback(fmt.Errorf("text model is not configured"))
	}
	completion, err := provider.Complete(ctx, batchfactoryv11.TextCompletionRequest{
		SystemPrompt: "你是短剧分镜编辑。把用户提供的中文原文拆成适合短视频制作的连续分镜候选。只返回 JSON，不要 Markdown。格式为 {\"candidates\":[{\"text\":\"字幕\",\"sourceText\":\"对应原文\",\"subtitleText\":\"适合口播的字幕\",\"speaker\":\"旁白\"}]}。不得编造原文之外的剧情；每条 text 要完整、可编辑。",
		UserPrompt:   source,
		Temperature:  0.2,
		MaxTokens:    4096,
	})
	if err != nil {
		return fallback(err)
	}
	candidates, err := parseSmartShuihuoCandidates(completion, batch.Books[0].ID)
	if err != nil || len(candidates) == 0 {
		if err == nil {
			err = fmt.Errorf("text model returned no candidates")
		}
		return fallback(err)
	}
	return candidates, "text_model", nil
}

func generateShuihuoPromptCandidates(ctx context.Context, provider batchfactoryv11.DirectorProvider, batch batchfactoryv11.Batch, kind string) ([]map[string]any, string, error) {
	segments := storedShuihuoSegmentsFromBatch(batch)
	fallback := func(err error) ([]map[string]any, string, error) {
		items := make([]map[string]any, 0, len(segments))
		for _, segment := range segments {
			id := modelString(segment, "id")
			text := modelString(segment, "sourceText")
			if id == "" || text == "" {
				continue
			}
			prefix := map[string]string{"image": "电影感画面，", "video": "镜头运动自然，", "negative": "避免"}[kind]
			if prefix == "" {
				prefix = "清晰呈现，"
			}
			items = append(items, map[string]any{"segmentId": id, "prompt": prefix + text})
		}
		return items, "deterministic_fallback", err
	}
	if provider == nil {
		return fallback(fmt.Errorf("text model is not configured"))
	}
	encoded, _ := json.Marshal(segments)
	completion, err := provider.Complete(ctx, batchfactoryv11.TextCompletionRequest{
		SystemPrompt: "你是短剧视觉提示词编辑。根据每条分镜原文生成可直接用于AI制作的中文提示词。只返回 JSON：{\"candidates\":[{\"segmentId\":\"原ID\",\"prompt\":\"提示词\"}]}。必须覆盖输入中的每个 segmentId，不要编造剧情。",
		UserPrompt:   fmt.Sprintf("提示词类型：%s\n分镜：%s", kind, string(encoded)),
		Temperature:  0.3,
		MaxTokens:    4096,
	})
	if err != nil {
		return fallback(err)
	}
	items, err := parseShuihuoPromptCandidates(completion)
	if err != nil || len(items) == 0 {
		if err == nil {
			err = fmt.Errorf("text model returned no prompt candidates")
		}
		return fallback(err)
	}
	return items, "text_model", nil
}

func generateShuihuoAssetImages(ctx context.Context, store batchfactoryv11.Store, files *localartifact.Store, owner string, batch batchfactoryv11.Batch, assetIDs []string, modelID, aspectRatio, provider string, localExecutors *localexecutor.Service, catalogStores ...novelfetchworkshop.Store) ([]batchfactoryv11.BookAssetImage, error) {
	return generateShuihuoAssetImagesWithPresetCatalog(ctx, store, files, owner, batch, assetIDs, modelID, aspectRatio, provider, localExecutors, nil, catalogStores...)
}

type comicPromptRewriter func(context.Context, string, novelfetchworkshop.RewriteRequest) (string, error)

func generateShuihuoAssetImagesWithPresetCatalog(ctx context.Context, store batchfactoryv11.Store, files *localartifact.Store, owner string, batch batchfactoryv11.Batch, assetIDs []string, modelID, aspectRatio, provider string, localExecutors *localexecutor.Service, presetCatalog *prompt.Catalog, catalogStores ...novelfetchworkshop.Store) ([]batchfactoryv11.BookAssetImage, error) {
	return generateShuihuoAssetImagesWithRepair(ctx, store, files, owner, batch, assetIDs, modelID, aspectRatio, provider, localExecutors, presetCatalog, nil, nil, owner, catalogStores...)
}

func generateShuihuoAssetImagesWithRepair(ctx context.Context, store batchfactoryv11.Store, files *localartifact.Store, owner string, batch batchfactoryv11.Batch, assetIDs []string, modelID, aspectRatio, provider string, localExecutors *localexecutor.Service, presetCatalog *prompt.Catalog, rewrite comicPromptRewriter, userConfig *UserConfigStore, modelConfigOwner string, catalogStores ...novelfetchworkshop.Store) ([]batchfactoryv11.BookAssetImage, error) {
	if len(batch.Books) == 0 {
		return nil, fmt.Errorf("%w: project has no source book", batchfactoryv11.ErrInvalid)
	}
	wanted := make(map[string]bool, len(assetIDs))
	for _, id := range assetIDs {
		if id = strings.TrimSpace(id); id != "" {
			wanted[id] = true
		}
	}
	if len(wanted) == 0 {
		return nil, fmt.Errorf("%w: asset selection is required", batchfactoryv11.ErrInvalid)
	}
	config, err := loadShuihuoOwnerConfig(ctx, store, owner, catalogStores...)
	if err != nil {
		return nil, err
	}
	catalog := shuihuoCollection(config, "modelCatalog", nil)
	effective := batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, batch.Books[0].SettingsState.Patch)
	type assetImageWork struct {
		asset         batchfactoryv11.BookAsset
		target        shuihuoAssetImageTarget
		prompt        string
		referenceURLs []string
	}
	work := make([]assetImageWork, 0, len(wanted))
	for _, asset := range batch.Books[0].AssetRecords {
		if !wanted[asset.ID] {
			continue
		}
		prompt := strings.TrimSpace(asset.Prompt)
		if prompt == "" {
			return nil, fmt.Errorf("%w: asset %s has no prompt", batchfactoryv11.ErrInvalid, asset.ID)
		}
		if strings.TrimSpace(modelConfigOwner) == "" {
			modelConfigOwner = owner
		}
		target, targetErr := resolveShuihuoAssetImageTargetWithChannels(ctx, userConfig, modelConfigOwner, asset, modelID, aspectRatio, provider, catalog)
		if targetErr != nil {
			return nil, targetErr
		}
		prompt, referenceImageURLs, compileErr := compileShuihuoAssetImagePrompt(asset, effective, presetCatalog, shuihuoAssetThreeViewEnabled(ctx))
		if compileErr != nil {
			return nil, compileErr
		}
		work = append(work, assetImageWork{asset: asset, target: target, prompt: prompt, referenceURLs: referenceImageURLs})
	}
	if len(work) == 0 {
		return nil, fmt.Errorf("%w: selected asset not found", batchfactoryv11.ErrNotFound)
	}
	progress := comicGenerationProgressFromContext(ctx)
	if progress != nil {
		progress("comic_asset_images", fmt.Sprintf("准备并发生成漫画资产图（已完成 0/%d 张）", len(work)))
	}
	// Asset references are independent provider requests. Generate up to five
	// at once, while keeping database writes ordered and atomic after all image
	// files have returned. This prevents one slow asset from making the whole
	// "立即执行" flow look frozen.
	assetImageConcurrency := comicAssetConcurrencyFromContext(ctx)
	// 资产图与漫画页共用统一配置的出图并发上限。
	workCtx := withShuihuoImageStage(withShuihuoImageConcurrency(ctx, comicSettingInt(effective, "imageConcurrency", shuihuoImageConcurrencyDefault)), shuihuoImageStageAsset)
	// 资产图与漫画页共用统一配置，水印设置也一并注入。
	workCtx = withShuihuoEffectiveSettings(workCtx, comicSettingsToMap(effective))
	type assetImageResult struct {
		index      int
		mediaType  string
		storageRef string
		source     string
		err        error
	}
	results := make([]assetImageResult, len(work))
	sem := make(chan struct{}, assetImageConcurrency)
	var workers sync.WaitGroup
	var resultMu sync.Mutex
	completed := 0
	skippedByPolicy := 0
	var firstErr error
	for index, item := range work {
		index, item := index, item
		workers.Add(1)
		go func() {
			defer workers.Done()
			select {
			case sem <- struct{}{}:
			case <-workCtx.Done():
				return
			}
			defer func() { <-sem }()
			generationPrompt := item.prompt
			repairKey := comicAssetRepairKey(batch.ID, batch.Books[0].ID, item.asset.ID)
			if cached, ok := shuihuoRepairedPrompts.Load(repairKey); ok {
				if repaired, textOK := cached.(string); textOK && strings.TrimSpace(repaired) != "" {
					generationPrompt = repaired
					if progress != nil {
						progress("comic_asset_images", fmt.Sprintf("资产 %s 复用已修复提示词", item.asset.Name))
					}
				}
			}
			mediaType, storageRef, source, generateErr := generateShuihuoAssetImageArtifactWithRetry(workCtx, store, files, presetCatalog, owner, item.target, generationPrompt, item.referenceURLs, localExecutors)
			if errors.Is(generateErr, errShuihuoImageContentPolicy) && rewrite != nil {
				if progress != nil {
					progress("comic_asset_images", fmt.Sprintf("资产 %s 三轮尝试后被内容审核拒绝，正在自动改写提示词", item.asset.Name))
				}
				repaired := repairComicImagePrompt(workCtx, owner, generationPrompt, generateErr.Error(), rewrite, true)
				if strings.TrimSpace(repaired) != "" && repaired != generationPrompt {
					shuihuoRepairedPrompts.Store(repairKey, repaired)
					if progress != nil {
						progress("comic_asset_images", fmt.Sprintf("资产 %s 提示词已去敏，正在重新生成", item.asset.Name))
					}
					mediaType, storageRef, source, generateErr = generateShuihuoImageArtifact(workCtx, store, files, presetCatalog, owner, item.target, repaired, item.referenceURLs, localExecutors)
				}
			} else if errors.Is(generateErr, errShuihuoImageContentPolicy) && progress != nil {
				progress("comic_asset_images", fmt.Sprintf("资产 %s 三轮尝试后仍被内容审核拒绝，跳过并继续其它资产", item.asset.Name))
			}
			if generateErr != nil {
				resultMu.Lock()
				if errors.Is(generateErr, errShuihuoImageContentPolicy) {
					// 审核拒绝的资产图直接跳过，不能让整本书的资产阶段
					// 因此中止，其它资产仍然要生成完。
					log.Printf("comic asset image skipped by content policy batch=%s asset=%s err=%v", batch.ID, item.asset.Name, generateErr)
					skippedByPolicy++
					if progress != nil {
						progress("comic_asset_images", fmt.Sprintf("%s 两次降级后仍被内容审核拒绝，本次跳过，稍后整本自动重试", item.asset.Name))
					}
				} else if firstErr == nil {
					firstErr = fmt.Errorf("%s 资产图生成失败: %w", item.asset.Name, generateErr)
				}
				results[index].err = generateErr
				resultMu.Unlock()
				return
			}
			resultMu.Lock()
			results[index] = assetImageResult{index: index, mediaType: mediaType, storageRef: storageRef, source: source}
			completed++
			if progress != nil {
				progress("comic_asset_images", fmt.Sprintf("漫画资产图 %d/%d 已完成：%s", completed, len(work), item.asset.Name))
			}
			resultMu.Unlock()
		}()
	}
	workers.Wait()
	resultMu.Lock()
	generateErr := firstErr
	resultMu.Unlock()
	generated := make([]batchfactoryv11.BookAssetImage, 0, len(work))
	for index, item := range work {
		result := results[index]
		if result.err != nil || result.storageRef == "" {
			continue
		}
		image, createErr := store.CreateBookAssetImage(ctx, owner, batch.ID, batch.Books[0].ID, item.asset.ID, batchfactoryv11.CreateBookAssetImageInput{StorageRef: result.storageRef, MediaType: result.mediaType, Source: result.source})
		if createErr != nil {
			if generateErr == nil {
				generateErr = fmt.Errorf("%s 资产图写回失败: %w", item.asset.Name, createErr)
			}
			_ = files.Remove(result.storageRef)
			continue
		}
		generated = append(generated, image)
	}
	if len(generated)+skippedByPolicy != len(wanted) {
		if generateErr == nil {
			generateErr = fmt.Errorf("%w: selected asset not found", batchfactoryv11.ErrNotFound)
		}
	}
	return generated, generateErr
}

const cmCharacterThreeViewPresetID = "batch-comic-character-three-view-cm"

// generateShuihuoAssetImageArtifactWithRetry protects the asset-image flow
// from gateways that accept an image job and then reset the long-lived HTTP
// connection before returning its result. Retrying only transport-level
// failures keeps ordinary provider/configuration errors visible immediately.
func generateShuihuoAssetImageArtifactWithRetry(ctx context.Context, store batchfactoryv11.Store, files *localartifact.Store, presetCatalog *prompt.Catalog, owner string, target shuihuoAssetImageTarget, imagePrompt string, referenceImageURLs []string, localExecutors *localexecutor.Service) (string, string, string, error) {
	// 默认 3 次总尝试（首轮 + 2 次重试），可用 QIANTIE_IMAGE_PROVIDER_MAX_ATTEMPTS
	// 覆盖；上限 6 次，避免配置写出无限重试。重试只发生在页面记录/并发槽之外，
	// 不会重复写页，也不会覆盖已成功的图。
	maxAttempts := shuihuoImageProviderMaxAttempts()
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// 把「第几次外层重试」带进上下文，供出图请求的结构化日志对账使用。
		attemptCtx := withShuihuoImageOuterAttempt(ctx, attempt)
		mediaType, storageRef, source, err := generateShuihuoImageArtifact(attemptCtx, store, files, presetCatalog, owner, target, imagePrompt, referenceImageURLs, localExecutors)
		if err == nil {
			return mediaType, storageRef, source, nil
		}
		retryable := retryableShuihuoAssetImageError(err)
		var billed *shuihuoImageBilledReadFailure
		if errors.As(err, &billed) {
			// HTTP 200 表示供应商已经接单并可能计费。正文没有完整取回时
			// 只记录失败，绝不自动再发一张，避免重复生成和重复扣费。
			retryable = false
		}
		var billedSave *shuihuoImageBilledSaveFailure
		if errors.As(err, &billedSave) {
			// 图已经生成并计费，只是本地磁盘写不进去；重发只会再扣一次费。
			retryable = false
		}
		if !retryable || attempt == maxAttempts {
			return mediaType, storageRef, source, err
		}
		lastErr = err
		delay := shuihuoImageProviderRetryDelay(attempt)
		var chatErr shuihuoImageChatChannelError
		switch {
		case errors.As(err, &billed):
			log.Printf("shuihuo image body stalled but billed; retrying once attempt=%d/%d after=%s kind=%s model=%s", attempt+1, maxAttempts, delay, billed.kind, target.ModelID)
		case errors.As(err, &chatErr):
			log.Printf("shuihuo image provider returned a chat reply instead of image data; retrying attempt=%d/%d after=%s status=%d body=%q model=%s", attempt+1, maxAttempts, delay, chatErr.StatusCode, chatErr.BodySummary, target.ModelID)
		default:
			log.Printf("shuihuo image provider connection interrupted; retrying attempt=%d/%d after=%s model=%s: %v", attempt+1, maxAttempts, delay, target.ModelID, err)
		}
		select {
		case <-ctx.Done():
			return "", "", "", ctx.Err()
		case <-time.After(delay):
		}
	}
	return "", "", "", lastErr
}

func retryableShuihuoAssetImageTransportError(err error) bool {
	if err == nil {
		return false
	}
	// 排队超时说明本机正在处理其它图片，并不代表上游调用失败；本轮没有
	// 发出请求，应该交给页级续跑稍后补齐，不能在这里立即重排三次。
	if errors.Is(err, errShuihuoImageQueueBusy) {
		return false
	}
	// 等名额期间上下文已取消/到期：调用方已经不等了，立即重试没有意义。
	var queueCanceled shuihuoImageQueueCanceledError
	if errors.As(err, &queueCanceled) {
		return false
	}
	// 上游已按「请求成功」计费的读失败不再自动重发，详见
	// shuihuoImageBilledReadFailure 的注释。
	var billed *shuihuoImageBilledReadFailure
	if errors.As(err, &billed) {
		return false
	}
	// 已计费但本地磁盘写不进去：重发只会重复扣费，等磁盘恢复后再由用户重试。
	var billedSave *shuihuoImageBilledSaveFailure
	if errors.As(err, &billedSave) {
		return false
	}
	// 按错误类型判断，不再匹配错误文案：业务文案已经中文化，关键词匹配会漏判。
	// net.Error 覆盖连接重置/超时/TLS 握手超时等所有传输层抖动。
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	// 上游已返回 200 但正文断流/过慢，属于可重试的读取失败类型。
	var stalled *shuihuoImageBodyStallError
	var slow *shuihuoImageBodySlowError
	if errors.As(err, &stalled) || errors.As(err, &slow) {
		return true
	}
	// 标准库没有为这条协议消息导出哨兵错误，只能保留这一处文字兜底。
	return strings.Contains(err.Error(), "http: server closed idle connection")
}

// retryableShuihuoAssetImageError 在传输抖动之外，把“生图请求被当成聊天请求”
// 的上游 400 也视为可重试的通道错误。
func retryableShuihuoAssetImageError(err error) bool {
	return retryableShuihuoAssetImageTransportError(err) || errors.Is(err, errShuihuoImageChatChannel)
}

// shuihuoImageProviderMaxAttempts 返回单页/单资产请求图片上游的最大尝试次数。
// 默认 3 次（首轮 + 2 次重试）；环境变量可配置，上限 6 次。
func shuihuoImageProviderMaxAttempts() int {
	const defaultAttempts = 3
	raw := strings.TrimSpace(os.Getenv("QIANTIE_IMAGE_PROVIDER_MAX_ATTEMPTS"))
	if raw == "" {
		return defaultAttempts
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return defaultAttempts
	}
	if value > 6 {
		return 6
	}
	return value
}

// shuihuoImageProviderRetryDelay 返回第 attempt 次失败后的退避时间：
// 2s、4s、8s…，上限 10s。QIANTIE_IMAGE_RETRY_BASE_DELAY_MS 可调整基准值，
// 测试中设为 0 可跳过等待。
func shuihuoImageProviderRetryDelay(attempt int) time.Duration {
	base := 2 * time.Second
	if raw := strings.TrimSpace(os.Getenv("QIANTIE_IMAGE_RETRY_BASE_DELAY_MS")); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil && value >= 0 && value <= 60000 {
			base = time.Duration(value) * time.Millisecond
		}
	}
	if attempt < 1 {
		attempt = 1
	}
	delay := base
	for step := 1; step < attempt; step++ {
		delay *= 2
		if delay > 10*time.Second {
			return 10 * time.Second
		}
	}
	if delay > 10*time.Second {
		delay = 10 * time.Second
	}
	return delay
}

func compileShuihuoAssetImagePrompt(asset batchfactoryv11.BookAsset, settings batchfactoryv11.SettingsPatch, presetCatalog *prompt.Catalog, useThreeView bool) (string, []string, error) {
	content := strings.TrimSpace(asset.Prompt)
	// 人物资产默认套用 CM·角色三视图，包含正面近照与正/侧/背全身，
	// 保证后续分镜复用同一角色外观；开关关闭时退化为单张设定图。
	if asset.Kind == "character" && useThreeView && presetCatalog != nil {
		template, found, templateErr := presetCatalog.PublishedBody("batch-factory", cmCharacterThreeViewPresetID)
		if templateErr != nil {
			return "", nil, fmt.Errorf("%w: 读取CM·角色三视图失败", batchfactoryv11.ErrUnavailable)
		}
		if !found || template.Slot != "batch.comic.character-image-template" || strings.TrimSpace(template.Body) == "" {
			return "", nil, fmt.Errorf("%w: CM·角色三视图未发布", batchfactoryv11.ErrUnavailable)
		}
		content = strings.TrimSpace(content + "\n\n" + template.Body)
	}
	// Asset images belong to the selected comic project, not to the historical
	// extraction source label. Older/manual assets can still be part of a comic
	// book, and must receive the selected comic style as well.
	if !shuihuoComicProductionMode(settings) {
		return content, nil, nil
	}
	resolvedSettings, err := resolvePublishedComicStyleReference(settings, presetCatalog)
	if err != nil {
		return "", nil, err
	}
	settings = resolvedSettings
	// Never silently send a plain character description to an image model when
	// the book is in comic mode. That is how a selected comic style previously
	// degraded into photorealistic asset images.
	if !batchfactoryv11.ComicImageStyleConfigured(settings) {
		return "", nil, fmt.Errorf("%w: 当前漫画项目未加载已选择的画风，请先保存画面配置", batchfactoryv11.ErrConflict)
	}
	contract, err := batchfactoryv11.BuildComicImageContract(settings, batchfactoryv11.ComicImagePurposeAsset, asset.Kind, content, nil, nil)
	if err != nil {
		return "", nil, err
	}
	return contract.Prompt, contract.ReferenceImageURLs, nil
}

func shuihuoComicProductionMode(settings batchfactoryv11.SettingsPatch) bool {
	if raw, ok := settings["productionMode"]; ok && len(raw) > 0 {
		var mode string
		if json.Unmarshal(raw, &mode) == nil && strings.TrimSpace(mode) != "" {
			return strings.EqualFold(strings.TrimSpace(mode), "comic")
		}
	}
	var config struct {
		Comic struct {
			Enabled bool `json:"enabled"`
			Style   struct {
				Body string `json:"body"`
			} `json:"style"`
		} `json:"comic"`
	}
	if raw := settings["aiPromptConfig"]; len(raw) > 0 && json.Unmarshal(raw, &config) == nil {
		return config.Comic.Enabled && strings.TrimSpace(config.Comic.Style.Body) != ""
	}
	return false
}

func generateShuihuoImageArtifact(ctx context.Context, store batchfactoryv11.Store, files *localartifact.Store, presetCatalog *prompt.Catalog, owner string, target shuihuoAssetImageTarget, imagePrompt string, referenceImageURLs []string, localExecutors *localexecutor.Service) (mediaType, storageRef, source string, returnErr error) {
	if files == nil {
		return "", "", "", fmt.Errorf("%w: image storage is unavailable", batchfactoryv11.ErrUnavailable)
	}
	// 图片模型会按内容审核拒掉露骨或血腥描述；先做降敏替换，
	// 保证同一张漫画页不会因为几个词反复被拒。
	originalImagePrompt := imagePrompt
	imagePrompt = sanitizeShuihuoImagePrompt(imagePrompt)
	if imagePrompt != originalImagePrompt {
		reportShuihuoImageEvent(ctx, "comic_images", "图片提示词已自动去敏：已替换可能触发内容审核的敏感词")
	}
	cache := shuihuoImageGenerationCacheFromContext(ctx)
	fingerprint := ""
	cacheClaimed := false
	if cache != nil && !shuihuoImageGenerationCacheBypassed(ctx) {
		settings := shuihuoEffectiveSettingsFromContext(ctx)
		watermarkEnabled, _ := settings["watermarkEnabled"].(bool)
		watermarkText, _ := settings["watermarkText"].(string)
		endpoint := modelString(target.CatalogModel, "endpoint")
		if endpoint == "" {
			endpoint = modelString(target.CatalogModel, "baseUrl")
		}
		fingerprint = shuihuoImageGenerationFingerprint(shuihuoImageFingerprintInput{Owner: owner, Endpoint: endpoint, ModelID: target.ModelID, Prompt: imagePrompt, AspectRatio: target.AspectRatio, References: referenceImageURLs, WatermarkEnabled: watermarkEnabled, WatermarkText: watermarkText})
		entry, outcome, claimErr := cache.Claim(ctx, owner, fingerprint, target.ModelID, 10*time.Minute)
		if claimErr != nil {
			log.Printf("shuihuo image cache unavailable owner=%s model=%s: %v", owner, target.ModelID, claimErr)
		} else if outcome == imagegenerationcache.ClaimReady {
			file, openErr := files.Open(entry.StorageRef)
			if openErr == nil {
				_ = file.Close()
				return entry.MediaType, entry.StorageRef, "provider", nil
			}
			_ = cache.MarkFailed(ctx, owner, fingerprint, openErr)
		} else if outcome == imagegenerationcache.ClaimBusy {
			// 同一指纹的图片已由另一个请求（常见于超时重试、或同一页被重复调度）
			// 在上游生成中。这不是失败：等它落盘后直接复用结果。
			// 之前直接返回 conflict，会让 30 页只差 1 页时整本判失败。
			waitDeadline := time.Now().Add(shuihuoImageBusyWaitTimeout)
			for time.Now().Before(waitDeadline) {
				select {
				case <-ctx.Done():
					return "", "", "", ctx.Err()
				case <-time.After(shuihuoImageBusyPollInterval):
				}
				entry, outcome, claimErr = cache.Claim(ctx, owner, fingerprint, target.ModelID, 10*time.Minute)
				if claimErr != nil {
					break
				}
				if outcome == imagegenerationcache.ClaimReady {
					if file, openErr := files.Open(entry.StorageRef); openErr == nil {
						_ = file.Close()
						return entry.MediaType, entry.StorageRef, "provider", nil
					}
				}
				if outcome == imagegenerationcache.ClaimAcquired {
					cacheClaimed = true
					break
				}
			}
			if !cacheClaimed {
				return "", "", "", fmt.Errorf("%w: 相同图片长时间生成未完成，请稍后重试", batchfactoryv11.ErrConflict)
			}
		} else if outcome == imagegenerationcache.ClaimAcquired {
			cacheClaimed = true
		}
	}
	if cacheClaimed {
		defer func() {
			if returnErr == nil && storageRef != "" {
				if err := cache.MarkReady(context.Background(), owner, fingerprint, storageRef, mediaType, imagegenerationcache.Timing{}); err != nil {
					log.Printf("shuihuo image cache mark ready failed owner=%s model=%s: %v", owner, target.ModelID, err)
				}
				return
			}
			if err := cache.MarkFailed(context.Background(), owner, fingerprint, returnErr); err != nil {
				log.Printf("shuihuo image cache mark failed owner=%s model=%s: %v", owner, target.ModelID, err)
			}
		}()
	}
	if target.Provider == batchfactoryv11.VideoProviderDoubaoLocal {
		if baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("QIANTIE_DOUBAO_RELAY_BASE_URL")), "/"); baseURL != "" {
			resolvedReferences, err := resolveComicReferenceImageData(ctx, store, files, presetCatalog, owner, referenceImageURLs)
			if err != nil {
				return "", "", "", err
			}
			token := strings.TrimSpace(os.Getenv("QIANTIE_DOUBAO_RELAY_BEARER_TOKEN"))
			// 评审环境不再把旧的单账号变量强行传给 Relay；省略账号后由 Relay
			// 按已登录账号轮询，并由它自己的账号锁保证切换和生成不相互挤压。
			account := strings.TrimSpace(os.Getenv("QIANTIE_DOUBAO_RELAY_FIXED_ACCOUNT"))
			return generateShuihuoRelayImage(ctx, files, target, imagePrompt, resolvedReferences, baseURL, token, account)
		}
		if localExecutors == nil {
			return "", "", "", fmt.Errorf("%w: 豆包本地执行器未配置", batchfactoryv11.ErrUnavailable)
		}
		resolvedReferences, err := resolveComicReferenceImageData(ctx, store, files, presetCatalog, owner, referenceImageURLs)
		if err != nil {
			return "", "", "", err
		}
		mediaType, storageRef, err := waitForLocalExecutorImage(ctx, localExecutors, files, owner, target.ModelID, imagePrompt, shuihuoImageSize(target.AspectRatio), map[string]any{"ratio": target.AspectRatio, "referenceImages": resolvedReferences, "imageUrls": resolvedReferences})
		return mediaType, storageRef, "doubao_local_executor", err
	}
	model := target.CatalogModel
	endpointRaw := modelString(model, "endpoint")
	if endpointRaw == "" {
		endpointRaw = modelString(model, "baseUrl")
	}
	endpoint, err := shuihuoImageGenerationEndpoint(endpointRaw)
	if err != nil {
		return "", "", "", fmt.Errorf("%w: 图片模型地址无效，请在资产设置中检查模型接口地址", batchfactoryv11.ErrInvalid)
	}
	credential := shuihuoModelCredential(model)
	if credential == "" {
		return "", "", "", fmt.Errorf("%w: 图片模型缺少凭证，请在资产设置中补全", batchfactoryv11.ErrUnavailable)
	}
	resolvedReferences, err := resolveComicReferenceImageData(ctx, store, files, presetCatalog, owner, referenceImageURLs)
	if err != nil {
		return "", "", "", err
	}
	// 第一次用常规降敏提示词；如果仍被内容审核拒绝，再用更保守的降级提示词
	// 重新生成一次。跳过一张图之前至少给用户两次机会，避免整本缺页。
	attemptPrompts := []string{imagePrompt}
	if aggressive := sanitizeShuihuoImagePromptAggressive(imagePrompt); aggressive != imagePrompt {
		attemptPrompts = append(attemptPrompts, aggressive)
	}
	// 暴力/血腥类拒绝靠词表降级往往无效：提示词里仍然写着“抢刀叉、鲜血
	// 飞溅、划破脸颊”。第三次改成整页暴力合规改写，并追加合规声明，让
	// 每一次自动重试都带着修复，而不是拿原提示词反复撞审核。
	violenceSafePrompt := sanitizeShuihuoImagePromptViolence(imagePrompt)
	if violenceSafePrompt != imagePrompt && violenceSafePrompt != attemptPrompts[len(attemptPrompts)-1] {
		attemptPrompts = append(attemptPrompts, violenceSafePrompt)
	}
	var policyErr error
	for attemptIndex, attemptPrompt := range attemptPrompts {
		if attemptIndex > 0 {
			message := fmt.Sprintf("第 %d 次降级重试：改用更保守的图片提示词重新生成", attemptIndex+1)
			if attemptPrompt == violenceSafePrompt {
				message = fmt.Sprintf("第 %d 次降级重试：已按暴力合规规则改写提示词后重新生成", attemptIndex+1)
			}
			reportShuihuoImageEvent(ctx, "comic_images", message)
		}
		mediaType, storageRef, attemptErr := requestShuihuoProviderImage(ctx, files, target, model, endpoint, credential, attemptPrompt, resolvedReferences)
		if errors.Is(attemptErr, errShuihuoImageReferenceUnusable) && len(resolvedReferences) > 0 {
			// 网关读不到参考图（/edits 端点常见）时不能让整本书失败：
			// 去掉参考图，用同一段提示词走纯文生图补一张。
			log.Printf("shuihuo image provider cannot read references; retrying without references model=%s", target.ModelID)
			reportShuihuoImageEvent(ctx, "comic_images", "上游无法读取参考图，改用纯文生图重试")
			mediaType, storageRef, attemptErr = requestShuihuoProviderImage(ctx, files, target, model, endpoint, credential, attemptPrompt, nil)
		}
		if attemptErr == nil {
			if attemptIndex > 0 {
				reportShuihuoImageEvent(ctx, "comic_images", fmt.Sprintf("第 %d 次降级重试后图片生成成功", attemptIndex+1))
			}
			return mediaType, storageRef, "provider", nil
		}
		if !errors.Is(attemptErr, errShuihuoImageContentPolicy) {
			return "", "", "", attemptErr
		}
		policyErr = attemptErr
		log.Printf("shuihuo image content policy rejected attempt=%d/%d provider=%s model=%s", attemptIndex+1, len(attemptPrompts), target.Provider, target.ModelID)
		reportShuihuoImageEvent(ctx, "comic_images", fmt.Sprintf("第 %d 次图片提示词被内容审核拒绝", attemptIndex+1))
	}
	return "", "", "", policyErr
}

// newShuihuoProviderImageRequest 构造一次图片生成请求：支持编辑的模型走
// /images/edits 携带参考图，其余走 /images/generations。
// responseFormat 必须明确指定 url 或 b64_json。空字符串仅用于兼容明确拒绝
// response_format 参数的旧网关，不能再把“未填写”误当成 URL 模式。
func newShuihuoProviderImageRequest(ctx context.Context, endpoint, modelID, prompt, aspectRatio string, referenceImageURLs []string, model map[string]any, responseFormat string) (*http.Request, error) {
	if len(referenceImageURLs) > 0 && shuihuoImageModelSupportsEdits(model, modelID) {
		return newShuihuoImageEditRequestWithResponseFormat(ctx, endpoint, modelID, prompt, shuihuoImageSizeForModel(modelID, aspectRatio), referenceImageURLs, responseFormat)
	}
	payload := map[string]any{"model": modelID, "prompt": prompt, "n": 1, "size": shuihuoImageSizeForModel(modelID, aspectRatio)}
	if strings.TrimSpace(responseFormat) != "" {
		payload["response_format"] = responseFormat
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: 无法向图片通道发起请求，请检查网络与模型配置", batchfactoryv11.ErrUpstream)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// shuihuoImageResponseRejectsResponseFormat 识别网关不认识 response_format
// 字段的报错，用于自动去掉该字段后重发一次。
func shuihuoImageResponseRejectsResponseFormat(raw []byte) bool {
	detail := strings.ToLower(string(raw))
	for _, keyword := range []string{"response_format", "response format", "unknown parameter", "unsupported parameter", "unrecognized", "不支持的参数", "未知参数"} {
		if strings.Contains(detail, keyword) {
			return true
		}
	}
	return false
}

// requestShuihuoProviderImage 执行一次图片供应商请求（含排队、解码与落盘）。
// 内容审核拒绝会返回 errShuihuoImageContentPolicy，由调用方决定是否降级重试。
func requestShuihuoProviderImage(ctx context.Context, files *localartifact.Store, target shuihuoAssetImageTarget, model map[string]any, endpoint, credential, imagePrompt string, referenceImageURLs []string) (string, string, error) {
	size := shuihuoImageSizeForModel(target.ModelID, target.AspectRatio)
	useEdits := len(referenceImageURLs) > 0 && shuihuoImageModelSupportsEdits(model, target.ModelID)
	outerAttempt := shuihuoImageOuterAttemptFromContext(ctx)
	// 恢复 11/03 已验证链路：普通生图不传 response_format，由网关返回
	// 紧凑 URL；只有 CDN 下载失败才在下一轮改要 base64。带参考图的编辑
	// 请求仍直接取 base64，避免编辑结果再依赖 CDN。
	// 旧网关若明确拒绝该字段，下面仍会自动去掉 response_format 做兼容。
	responseFormat := ""
	if useEdits {
		responseFormat = "b64_json"
	}
	for attempt := 0; attempt < 3; attempt++ {
		attemptStart := time.Now()
		var headerElapsed, bodyElapsed, downloadElapsed, saveElapsed time.Duration
		configuredConcurrency := shuihuoImageConcurrencyFromContext(ctx)
		parsedEndpoint, _ := url.Parse(strings.TrimSpace(endpoint))
		adaptiveHost := strings.ToLower(strings.TrimSpace(parsedEndpoint.Host))
		adaptiveLimit := shuihuoAdaptiveImageLimiter.limit(adaptiveHost, configuredConcurrency)
		slotCtx := context.WithValue(ctx, shuihuoImageConcurrencyContextKey{}, adaptiveLimit)
		if taskSemaphore := shuihuoImageSemaphoreFromContext(slotCtx); taskSemaphore != nil {
			taskSemaphore.setLimit(adaptiveLimit)
		}
		queueStart := time.Now()
		queueCtx, cancelQueue := shuihuoImageQueueContext(slotCtx)
		releaseHostSlot, _, err := shuihuoAdaptiveImageLimiter.acquire(queueCtx, endpoint, configuredConcurrency)
		if err != nil {
			cancelQueue()
			queued := time.Since(queueStart)
			logImageAttempt(shuihuoImageAttemptLog{elapsed: time.Since(attemptStart), attempt: outerAttempt, inner: attempt + 1, queued: queued, errorKind: shuihuoImageAttemptKindOfError(err, 0), model: target.ModelID, detail: err.Error()})
			return "", "", shuihuoImageQueueError(ctx, err)
		}
		releaseImageSlot, err := acquireShuihuoGPTImageGenerationSlot(queueCtx, target.ModelID)
		queued := time.Since(queueStart)
		cancelQueue()
		if err != nil {
			releaseHostSlot()
			logImageAttempt(shuihuoImageAttemptLog{elapsed: time.Since(attemptStart), attempt: outerAttempt, inner: attempt + 1, queued: queued, errorKind: shuihuoImageAttemptKindOfError(err, 0), model: target.ModelID, detail: err.Error()})
			return "", "", shuihuoImageQueueError(ctx, err)
		}
		releaseSlots := func() {
			releaseImageSlot()
			releaseHostSlot()
		}
		// 请求上下文在真正拿到名额后才绑定。排队阶段不会再向上游发请求，
		// 也不会出现“上游已经计费，但本地只剩几秒就超时”的情况。
		req, err := newShuihuoProviderImageRequest(ctx, endpoint, target.ModelID, imagePrompt, target.AspectRatio, referenceImageURLs, model, responseFormat)
		if err != nil {
			releaseSlots()
			logImageAttempt(shuihuoImageAttemptLog{elapsed: time.Since(attemptStart), attempt: outerAttempt, inner: attempt + 1, queued: queued, errorKind: "other", model: target.ModelID, detail: err.Error()})
			return "", "", err
		}
		log.Printf("shuihuo image request provider=%s model=%s endpoint=%s refs=%d size=%s response_format=%s edits=%t queued_ms=%d", target.Provider, target.ModelID, endpoint, len(referenceImageURLs), size, responseFormat, useEdits, queued.Milliseconds())
		req.Header.Set("Authorization", "Bearer "+credential)
		headerStart := time.Now()
		response, err := shuihuoImageGenerationHTTPClient(endpoint).Do(req)
		headerElapsed = time.Since(headerStart)
		if err != nil {
			releaseSlots()
			errorKind := shuihuoImageAttemptKindOfError(err, 0)
			shuihuoAdaptiveImageLimiter.report(adaptiveHost, configuredConcurrency, errorKind)
			logImageAttempt(shuihuoImageAttemptLog{elapsed: time.Since(attemptStart), attempt: outerAttempt, inner: attempt + 1, queued: queued, header: headerElapsed, errorKind: errorKind, model: target.ModelID, detail: err.Error()})
			return "", "", fmt.Errorf("%w: 图片通道请求失败：%w", batchfactoryv11.ErrUpstream, err)
		}
		// 收到响应头说明上游生成阶段已经结束。随后切换到独立的正文传输槽，
		// 释放稀缺的生图名额，让后续页可以继续提交；传输槽仍限制慢连接总数。
		releaseTransfer, transferErr := beginShuihuoImageResponseTransfer(ctx, releaseSlots)
		if transferErr != nil {
			response.Body.Close()
			return "", "", shuihuoImageQueueError(ctx, transferErr)
		}
		// 上游有响应就是真实进展：刷新停滞看门狗，避免慢图正文传输被判 stalled。
		reportComicGenerationHeartbeat(ctx)
		bodyMode := "url"
		bodyStall := shuihuoImageBodyStallTimeoutURL
		bodyTotal := shuihuoImageBodyTotalTimeoutURL
		if responseFormat == "b64_json" {
			bodyMode = "base64"
			bodyStall = shuihuoImageBodyStallTimeoutBase64
			bodyTotal = shuihuoImageBodyTotalTimeoutBase64
		}
		bodyStart := time.Now()
		raw, readErr := readShuihuoImageResponseBody(response, bodyMode, bodyStall, bodyTotal)
		bodyElapsed = time.Since(bodyStart)
		response.Body.Close()
		releaseTransfer()
		var result struct {
			Data []struct {
				B64JSON string `json:"b64_json"`
				URL     string `json:"url"`
			} `json:"data"`
		}
		resultParsed := json.Unmarshal(raw, &result) == nil && len(result.Data) > 0 && (strings.TrimSpace(result.Data[0].B64JSON) != "" || strings.TrimSpace(result.Data[0].URL) != "")
		if readErr != nil && !(response.StatusCode >= 200 && response.StatusCode < 300 && resultParsed) {
			// 上游很可能已经出图，只是响应体在传输中断（网关截断/连接重置）。
			// 用 %w 保留底层错误类型，重试判定按 net/io 类型走，不再依赖文案。
			log.Printf("shuihuo image response body read failed provider=%s model=%s status=%d refs=%d: %v", target.Provider, target.ModelID, response.StatusCode, len(referenceImageURLs), readErr)
			wrapped := fmt.Errorf("%w: 图片通道响应读取中断（HTTP %d，参考图 %d 张）：%w", batchfactoryv11.ErrUpstream, response.StatusCode, len(referenceImageURLs), readErr)
			errorKind := shuihuoImageAttemptKindOfError(readErr, response.StatusCode)
			bodyKind := ""
			var stalled *shuihuoImageBodyStallError
			var slow *shuihuoImageBodySlowError
			switch {
			case errors.As(readErr, &stalled):
				bodyKind = shuihuoImageBodyFailureStalled
			case errors.As(readErr, &slow):
				bodyKind = shuihuoImageBodyFailureSlow
			}
			if errorKind == "other" {
				errorKind = "body_read_failed"
			}
			shuihuoAdaptiveImageLimiter.report(adaptiveHost, configuredConcurrency, errorKind)
			logImageAttempt(shuihuoImageAttemptLog{elapsed: time.Since(attemptStart), attempt: outerAttempt, inner: attempt + 1, status: response.StatusCode, queued: queued, header: headerElapsed, body: bodyElapsed, errorKind: errorKind, model: target.ModelID, detail: wrapped.Error()})
			if response.StatusCode == http.StatusOK && shuihuoImageDedupeBilledRetry() {
				// 200 说明对面已经按成功计费。断流（BODY_STALLED）允许在页级/日预算内
				// 重发一次；只是慢（BODY_SLOW）不重发，避免重复扣费。
				return "", "", &shuihuoImageBilledReadFailure{message: wrapped.Error(), kind: bodyKind}
			}
			return "", "", wrapped
		}
		if readErr != nil {
			// 一些 HTTP/1.1 网关会把完整 JSON 发完后直接 RST 连接，没有发送
			// 正常的 chunk 结束标记。只要 JSON 已完整解析且包含图片，就直接
			// 使用已收到的结果；绝不因为连接尾部异常再次提交已计费的生图请求。
			log.Printf("shuihuo image response completed before connection reset provider=%s model=%s status=%d refs=%d bytes=%d err=%v", target.Provider, target.ModelID, response.StatusCode, len(referenceImageURLs), len(raw), readErr)
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			log.Printf("shuihuo image response provider=%s model=%s status=%d detail=%s", target.Provider, target.ModelID, response.StatusCode, shuihuoImageProviderErrorDetail(raw))
			logImageAttempt(shuihuoImageAttemptLog{elapsed: time.Since(attemptStart), attempt: outerAttempt, inner: attempt + 1, status: response.StatusCode, queued: queued, header: headerElapsed, body: bodyElapsed, errorKind: shuihuoImageAttemptKindOfStatus(response.StatusCode), model: target.ModelID, detail: shuihuoImageProviderErrorDetail(raw)})
			if responseFormat != "" && shuihuoImageResponseRejectsResponseFormat(raw) {
				log.Printf("shuihuo image gateway rejected response_format; retrying without it model=%s", target.ModelID)
				responseFormat = ""
				continue
			}
			// 先把“生图请求被当成聊天请求”识别成可重试的通道错误，避免被下面
			// 的内容审核关键词（“无法完成”“重新发送”）误判成永久拒绝。
			if shuihuoImageProviderChatChannelResponse(response.StatusCode, raw) {
				return "", "", shuihuoImageChatChannelError{StatusCode: response.StatusCode, BodySummary: strings.TrimPrefix(shuihuoImageProviderErrorDetail(raw), ": ")}
			}
			// 对话式软拒绝优先于硬审核：它不是标准 JSON 错误，需要单独归类，
			// 但仍走提示词改写链路（“请上传原图”类问题由改写/换通道兜底）。
			if shuihuoImageResponseSoftRefusal(response.StatusCode, raw) {
				return "", "", newShuihuoImageSoftRefusalError(shuihuoImageProviderErrorDetail(raw))
			}
			if shuihuoImageResponseContentRejected(response.StatusCode, raw) {
				return "", "", fmt.Errorf("%w: %s", errShuihuoImageContentPolicy, strings.TrimPrefix(shuihuoImageProviderErrorDetail(raw), ": "))
			}
			// /edits 端点不存在或不被网关接受时，同样按“参考图不可用”回退到
			// 纯文生图，避免整本书因为一个可选参考图而失败。
			if useEdits && (response.StatusCode == http.StatusNotFound || response.StatusCode == http.StatusMethodNotAllowed) {
				return "", "", fmt.Errorf("%w: %s", errShuihuoImageReferenceUnusable, strings.TrimPrefix(shuihuoImageProviderErrorDetail(raw), ": "))
			}
			if len(referenceImageURLs) > 0 && shuihuoImageProviderRejectedReferences(raw) {
				return "", "", fmt.Errorf("%w: %s", errShuihuoImageReferenceUnusable, strings.TrimPrefix(shuihuoImageProviderErrorDetail(raw), ": "))
			}
			return "", "", shuihuoImageUpstreamError{
				Message:    shuihuoImageStatusRejectionMessage(response.StatusCode, string(raw)),
				Raw:        shuihuoImageProviderErrorDetail(raw),
				StatusCode: response.StatusCode,
			}
		}
		if !resultParsed {
			// 200 也可能是对话式软拒绝（正文是自然语言而不是图像 JSON）。
			if shuihuoImageResponseSoftRefusalText(raw) {
				return "", "", newShuihuoImageSoftRefusalError(shuihuoImageProviderErrorDetail(raw))
			}
			return "", "", shuihuoImageUpstreamError{Message: "图片通道返回的内容不是有效图片数据，请检查该模型是否支持出图"}
		}
		downloadStart := time.Now()
		mediaType, imageBytes, err := decodeShuihuoGeneratedImage(ctx, shuihuoImageDownloadHTTPClient(), result.Data[0].B64JSON, result.Data[0].URL, credential, endpoint)
		downloadElapsed = time.Since(downloadStart)
		if err != nil {
			// 图已经生成，只是 CDN 下载不下来：改用 base64 重发一次，跳开 CDN。
			if errors.Is(err, errShuihuoImageDownloadUnusable) && responseFormat != "b64_json" && attempt < 2 {
				log.Printf("shuihuo image download unavailable; retrying with base64 model=%s", target.ModelID)
				responseFormat = "b64_json"
				continue
			}
			return "", "", fmt.Errorf("%w: %v", batchfactoryv11.ErrUpstream, err)
		}
		// 统一配置启用水印时，在保存前叠加白底方框文字。
		saveStart := time.Now()
		imageBytes, mediaType = applyShuihuoImageWatermark(ctx, imageBytes, mediaType)
		artifactID, err := localexecutor.NewArtifactID()
		if err != nil {
			return "", "", err
		}
		saved, err := files.SaveImage(artifactID, mediaType, bytes.NewReader(imageBytes))
		saveElapsed = time.Since(saveStart)
		if err != nil {
			if errors.Is(err, localartifact.ErrDiskPressure) {
				// 上游已经出图并按次计费，本地磁盘保护却拒绝落盘。此时返回普通
				// 错误会让页级/资产级重试重新提交同一张图，形成“扣钱不落盘”的
				// 重复计费；必须标记成已计费未落盘，让上层禁止自动重跑。
				logImageAttempt(shuihuoImageAttemptLog{elapsed: time.Since(attemptStart), attempt: outerAttempt, inner: attempt + 1, status: response.StatusCode, queued: queued, header: headerElapsed, body: bodyElapsed, download: downloadElapsed, save: saveElapsed, errorKind: "save_disk_pressure", model: target.ModelID, detail: err.Error()})
				return "", "", &shuihuoImageBilledSaveFailure{message: automationBilledNotRetrievedSaveMessage, cause: err}
			}
			return "", "", err
		}
		shuihuoAdaptiveImageLimiter.report(adaptiveHost, configuredConcurrency, "ok")
		logImageAttempt(shuihuoImageAttemptLog{elapsed: time.Since(attemptStart), attempt: outerAttempt, inner: attempt + 1, status: response.StatusCode, queued: queued, header: headerElapsed, body: bodyElapsed, download: downloadElapsed, save: saveElapsed, errorKind: "ok", model: target.ModelID})
		return mediaType, saved.StorageRef, nil
	}
	logImageAttempt(shuihuoImageAttemptLog{attempt: outerAttempt, inner: 3, errorKind: "other", model: target.ModelID, detail: "image provider rejected response_format"})
	return "", "", fmt.Errorf("%w: 图片通道不接受当前响应格式，请检查该模型是否支持返回图片数据", batchfactoryv11.ErrUpstream)
}

// applyShuihuoImageWatermark 读取统一配置中的水印设置，在生成的图片右下角
// 叠加白底方框文字。水印开关和文字都来自批次设置，未开启时原样返回。
func applyShuihuoImageWatermark(ctx context.Context, imageBytes []byte, mediaType string) ([]byte, string) {
	settings := shuihuoEffectiveSettingsFromContext(ctx)
	if settings == nil {
		return imageBytes, mediaType
	}
	enabled, _ := settings["watermarkEnabled"].(bool)
	text, _ := settings["watermarkText"].(string)
	if !enabled || strings.TrimSpace(text) == "" {
		return imageBytes, mediaType
	}
	return watermark.Apply(imageBytes, mediaType, watermark.Options{Enabled: true, Text: strings.TrimSpace(text)})
}

// shuihuoImageResponseHeaderTimeout 给慢图留足 3 分钟，但不能让一个完全不回
// 响应头的坏连接独占名额 12 分钟。旧的 60s 会误杀 60–120s 才返回的慢图；
// 完全不设限又会让所有并发槽位一起卡死，因此取线上慢图上限之上的 180s。
//
// shuihuoImageBodyStallTimeout 是「停滞超时」：读取循环里每读到一块字节就重置。
// 线上供应商会先返回 200，再花一到两分钟准备或分块传输正文，因此窗口取 180s；
// 旧的 30s 会主动掐掉仍会返回结果的连接，并诱发重复生成。
// shuihuoImageBodyTotalTimeout 是总上限兜底。供应商有时会忽略 URL 模式，
// 直接返回数 MB 的 base64 正文；线上实测正文持续有进展但 5 分钟仍未传完。
// 因此正文总上限取 5 分钟，停滞检测仍会在 180 秒无新字节时提前结束坏连接。
// 两组值都是变量，测试可临时覆盖。
var (
	// 单页/交互请求排队上限。通道只有 1~2 个真实并发时，多本书同时出图会让
	// 排队深度达到十几二十个；排队上限过短会让请求在轮到自己之前就到期重排，
	// 表现为整本书一直卡在“缺几页”，永远到不了发布。
	shuihuoImageQueueWaitTimeout = 15 * time.Minute
	// 正常响应头 5-6 秒返回；60 秒还不回包说明连接已经挂死，快速失败重试
	// 比一直占着出图名额更划算（原来 3 分钟会让一本卡住时整队陪等）。
	shuihuoImageResponseHeaderTimeout = 60 * time.Second
	// 正常正文 30-60 秒；120 秒内一个字节都没有说明流已断，交给重试。
	shuihuoImageBodyStallTimeoutURL    = 120 * time.Second
	shuihuoImageBodyStallTimeoutBase64 = 120 * time.Second
	// HTTP 200 后对方已可能计费。只要正文还有字节持续返回，
	// 就不再用绝对总时长丢弃已付费图片；真断流由上面的 stall 窗口拦截。
	shuihuoImageBodyTotalTimeoutURL    time.Duration
	shuihuoImageBodyTotalTimeoutBase64 time.Duration
	shuihuoImageQueueHeartbeatInterval = time.Minute
	// “相同图片正在生成”是并发/重试造成的瞬时状态，不是失败。
	// 等上游把那张图落盘后复用结果，避免 29/30 时整本被判失败。
	shuihuoImageBusyWaitTimeout  = 10 * time.Minute
	shuihuoImageBusyPollInterval = 10 * time.Second
)

var errShuihuoImageQueueBusy = fmt.Errorf("%w: 图片队列繁忙", batchfactoryv11.ErrUpstream)

// shuihuoImageQueueBusyError 是排队超时的完整错误：主文案只给中文结论，
// 底层超时原因放到「上游原文」行，由界面折叠到展开区查看。
type shuihuoImageQueueBusyError struct {
	wait  time.Duration
	cause error
}

func (e shuihuoImageQueueBusyError) Error() string {
	message := fmt.Sprintf("图片队列繁忙，等待 %s后仍没有可用名额，本轮已停止，稍后会自动补齐", shuihuoImageQueueWaitText(e.wait))
	if e.cause == nil {
		return message
	}
	return appendShuihuoUpstreamRaw(message, e.cause.Error())
}

func (e shuihuoImageQueueBusyError) Unwrap() error { return errShuihuoImageQueueBusy }

// shuihuoImageQueueWaitText 把等待时长写成中文分钟数（5m0s → 5 分钟），
// 避免英文时长格式混进主文案。
func shuihuoImageQueueWaitText(wait time.Duration) string {
	minutes := int((wait + time.Minute - 1) / time.Minute)
	if minutes < 1 {
		minutes = 1
	}
	return fmt.Sprintf("%d 分钟", minutes)
}

// shuihuoImageQueueCanceledError 表示排队等名额期间请求上下文已被取消或到期。
// 主文案只给中文结论，底层取消原因放到「上游原文」行；同时暴露 ErrUpstream
// 与原上下文错误两条错误链，上层既能沿用原有的上游瞬时错误分类，也能识别取消。
type shuihuoImageQueueCanceledError struct {
	cause error
}

func (e shuihuoImageQueueCanceledError) Error() string {
	message := "本轮已取消，图片名额等待已中止"
	if errors.Is(e.cause, context.DeadlineExceeded) {
		message = "本轮已超时，图片名额等待已中止"
	}
	return appendShuihuoUpstreamRaw(message, e.cause.Error())
}

func (e shuihuoImageQueueCanceledError) Unwrap() []error {
	return []error{batchfactoryv11.ErrUpstream, e.cause}
}

func shuihuoImageQueueError(parent context.Context, queueErr error) error {
	if parent != nil && parent.Err() != nil {
		return shuihuoImageQueueCanceledError{cause: parent.Err()}
	}
	return shuihuoImageQueueBusyError{wait: shuihuoImageQueueWaitTimeout, cause: queueErr}
}

// shuihuoImageRequestTimeout 保留为客户端可见配置值；0 表示不用「从发请
// 开始起算」的绝对总超时。建连/响应头仍限 3 分钟，正文连续 3 分钟无
// 新字节仍会断流；只是不再把「一直有数据回来」的已付费图片拦腰丢掉。
const shuihuoImageRequestTimeout time.Duration = 0

// shuihuoImageQueueBudget 限制一次排队最多 15 分钟。页级续跑会保留已完成
// 页面并只补缺页，但排队上限必须大于“排队深度 ÷ 通道并发 × 单张耗时”，
// 否则请求会在轮到自己之前反复到期重排，造成整本长期缺页。
func shuihuoImageQueueBudget(ctx context.Context) time.Duration {
	return shuihuoImageQueueWaitTimeout
}

func shuihuoImageQueueContext(ctx context.Context) (context.Context, context.CancelFunc) {
	queueCtx, cancelTimeout := context.WithTimeout(ctx, shuihuoImageQueueBudget(ctx))
	stop := make(chan struct{})
	var once sync.Once
	go func() {
		ticker := time.NewTicker(shuihuoImageQueueHeartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-queueCtx.Done():
				return
			case <-stop:
				return
			case <-ticker.C:
				reportComicGenerationHeartbeat(ctx)
			}
		}
	}()
	return queueCtx, func() {
		once.Do(func() {
			close(stop)
			cancelTimeout()
		})
	}
}

// newShuihuoImageTransport 构造出图请求的 Transport。出图链路按协议变体
// 各复用一个实例（见下方的包级共享 client）。
func newShuihuoImageTransport(forceHTTP2 bool) *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = shuihuoImageResponseHeaderTimeout
	transport.IdleConnTimeout = 90 * time.Second
	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 32
	if forceHTTP2 {
		transport.ForceAttemptHTTP2 = true
		transport.TLSNextProto = nil
	} else {
		transport.ForceAttemptHTTP2 = false
		transport.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}
	}
	return transport
}

// shuihuoLongRunningImageHTTPClient 构造一个新的 HTTP/2 client（测试用构造器；
// 生产路径走下面的共享实例）。
func shuihuoLongRunningImageHTTPClient() *http.Client {
	return safehttp.NewClientWithTransportWithoutTimeout(newShuihuoImageTransport(true))
}

// 生产路径复用的包级 client。此前每次请求都新建 Transport，等于每次重新
// TLS 握手、不复用连接，会放大网关的连接级重置与限流暴露面。客户端不设
// 绝对总超时，头部与正文停滞分别由 Transport 和 body 分层计时负责。
var (
	shuihuoImageSharedHTTP2Client = sync.OnceValue(func() *http.Client {
		return safehttp.NewClientWithTransportWithoutTimeout(newShuihuoImageTransport(true))
	})
	shuihuoImageSharedHTTP1Client = sync.OnceValue(func() *http.Client {
		return safehttp.NewClientWithTransportWithoutTimeout(newShuihuoImageTransport(false))
	})
)

// shuihuoImageBodyStallError 表示连续 stall 时间内没有读到任何新字节，
// 是真正的断流；它才计入「已计费未取回」并允许预算内重发一次。
type shuihuoImageBodyStallError struct {
	stall    time.Duration
	received int
	mode     string
}

func (err *shuihuoImageBodyStallError) Error() string {
	return fmt.Sprintf("response body stalled after=%s without new bytes received=%d mode=%s", err.stall, err.received, err.mode)
}

// shuihuoImageBodySlowError 表示直到总上限仍有字节在传，只是整体太慢。
// 它不按断流处理，也不计入 billed-not-retrieved，避免把「慢」误记成「丢」。
type shuihuoImageBodySlowError struct {
	total    time.Duration
	received int
	mode     string
}

func (err *shuihuoImageBodySlowError) Error() string {
	return fmt.Sprintf("response body slow total=%s received=%d mode=%s", err.total, err.received, err.mode)
}

// readShuihuoImageResponseBody 按「停滞超时 + 总上限」读取响应体：
//   - 每成功读到一块字节就重置 stall 计时器；
//   - 连续 stall 没有新字节 → shuihuoImageBodyStallError；
//   - 总耗时到 total 仍有数据在传 → shuihuoImageBodySlowError；
//   - 总耗时到 total 完全没数据 → 也按 slow 兜底（stall 通常先触发）。
func readShuihuoImageResponseBody(response *http.Response, mode string, stall, total time.Duration) ([]byte, error) {
	if stall <= 0 {
		stall = 30 * time.Second
	}
	type readChunk struct {
		data []byte
		err  error
	}
	chunks := make(chan readChunk, 1)
	done := make(chan struct{})
	var stopOnce sync.Once
	stop := func() { stopOnce.Do(func() { close(done) }) }
	defer stop()

	go func() {
		buf := make([]byte, 32*1024)
		for {
			n, err := response.Body.Read(buf)
			if n > 0 {
				data := append([]byte(nil), buf[:n]...)
				select {
				case chunks <- readChunk{data: data}:
				case <-done:
					return
				}
			}
			if err != nil {
				select {
				case chunks <- readChunk{err: err}:
				case <-done:
				}
				return
			}
		}
	}()

	stallTimer := time.NewTimer(stall)
	defer stallTimer.Stop()
	var totalTimer *time.Timer
	var totalTimeout <-chan time.Time
	if total > 0 {
		totalTimer = time.NewTimer(total)
		totalTimeout = totalTimer.C
		defer totalTimer.Stop()
	}
	resetStall := func() {
		if !stallTimer.Stop() {
			select {
			case <-stallTimer.C:
			default:
			}
		}
		stallTimer.Reset(stall)
	}

	var payload bytes.Buffer
	for {
		select {
		case chunk := <-chunks:
			if len(chunk.data) > 0 {
				payload.Write(chunk.data)
				if payload.Len() > 16<<20 {
					_ = response.Body.Close()
					return payload.Bytes(), fmt.Errorf("response body exceeds limit received=%d mode=%s", payload.Len(), mode)
				}
				resetStall()
			}
			if chunk.err != nil {
				if errors.Is(chunk.err, io.EOF) {
					return payload.Bytes(), nil
				}
				return payload.Bytes(), chunk.err
			}
		case <-stallTimer.C:
			_ = response.Body.Close()
			return payload.Bytes(), &shuihuoImageBodyStallError{stall: stall, received: payload.Len(), mode: mode}
		case <-totalTimeout:
			_ = response.Body.Close()
			return payload.Bytes(), &shuihuoImageBodySlowError{total: total, received: payload.Len(), mode: mode}
		}
	}
}

// shuihuoImageBilledReadFailure 表示上游已按「请求成功」计费（HTTP 200 已返回），
// 但响应体没有完整取回。kind 区分两种完全不同的原因：
//   - BODY_STALLED：连续 stall 时间没有新字节，是真断流；已计费，不自动重发。
//   - BODY_SLOW：一直有数据但到总上限仍未读完；不重发，也不计入
//     billed-not-retrieved，避免把「慢」误记成「丢」。
type shuihuoImageBilledReadFailure struct {
	message string
	kind    string
	retried bool
}

// shuihuoImageBilledSaveFailure 表示上游已返回图片（HTTP 200，按次计费），
// 但本地落盘被磁盘保护拒绝。图片字节已经丢弃，自动重发只会再扣一次费，
// 因此上层必须按「已计费未取回」处理，禁止自动重跑。
type shuihuoImageBilledSaveFailure struct {
	message string
	cause   error
}

func (err *shuihuoImageBilledSaveFailure) Error() string { return err.message }

// Unwrap 同时保留 ErrUpstream 语义与底层磁盘错误，便于分类与日志取证。
func (err *shuihuoImageBilledSaveFailure) Unwrap() []error {
	if err.cause == nil {
		return []error{batchfactoryv11.ErrUpstream}
	}
	return []error{batchfactoryv11.ErrUpstream, err.cause}
}

const (
	shuihuoImageBodyFailureStalled = "BODY_STALLED"
	shuihuoImageBodyFailureSlow    = "BODY_SLOW"
)

func (err *shuihuoImageBilledReadFailure) Error() string { return err.message }

func (err *shuihuoImageBilledReadFailure) Kind() string { return err.kind }

// Unwrap 保留 ErrUpstream 语义，上层的错误分类不受影响。
func (err *shuihuoImageBilledReadFailure) Unwrap() error { return batchfactoryv11.ErrUpstream }

func shuihuoImageDedupeBilledRetry() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("QIANTIE_IMAGE_DEDUPE_BILLED_RETRY"))) {
	case "0", "false", "off", "no":
		return false
	}
	return true
}

type shuihuoImageOuterAttemptContextKey struct{}

// withShuihuoImageOuterAttempt 记录外层重试序号，仅用于结构化日志。
func withShuihuoImageOuterAttempt(ctx context.Context, attempt int) context.Context {
	if ctx == nil {
		return ctx
	}
	return context.WithValue(ctx, shuihuoImageOuterAttemptContextKey{}, attempt)
}

func shuihuoImageOuterAttemptFromContext(ctx context.Context) int {
	if ctx == nil {
		return 1
	}
	if attempt, ok := ctx.Value(shuihuoImageOuterAttemptContextKey{}).(int); ok && attempt > 0 {
		return attempt
	}
	return 1
}

// 出图请求的结构化日志字段顺序固定：elapsed_ms → attempt → inner_attempt →
// http_status → queued_ms → error_kind → model → detail，单行便于 grep 与
// 和上游仪表盘逐条对账。
type shuihuoImageAttemptLog struct {
	elapsed   time.Duration
	attempt   int
	inner     int
	status    int
	queued    time.Duration
	header    time.Duration
	body      time.Duration
	download  time.Duration
	save      time.Duration
	errorKind string
	model     string
	detail    string
}

func logImageAttempt(entry shuihuoImageAttemptLog) {
	log.Printf("shuihuo image attempt elapsed_ms=%d attempt=%d inner_attempt=%d http_status=%d queued_ms=%d header_ms=%d body_ms=%d download_ms=%d save_ms=%d error_kind=%s model=%s detail=%q",
		entry.elapsed.Milliseconds(), entry.attempt, entry.inner, entry.status,
		entry.queued.Milliseconds(), entry.header.Milliseconds(), entry.body.Milliseconds(),
		entry.download.Milliseconds(), entry.save.Milliseconds(), entry.errorKind, entry.model, entry.detail)
}

func shuihuoImageAttemptKindOfError(err error, status int) string {
	if err != nil {
		message := strings.ToLower(err.Error())
		switch {
		case strings.Contains(message, "client.timeout"):
			return "client_timeout"
		case strings.Contains(message, "context canceled"):
			return "ctx_canceled"
		case strings.Contains(message, "context deadline exceeded"):
			return "ctx_deadline"
		case strings.Contains(message, "connection reset"),
			strings.Contains(message, "forcibly closed"),
			strings.Contains(message, "wsarecv"):
			return "reset"
		case strings.Contains(message, "unexpected eof"):
			return "unexpected_eof"
		case strings.Contains(message, "tls handshake timeout"):
			return "tls_timeout"
		case strings.Contains(message, "body stalled"):
			return "body_stalled"
		case strings.Contains(message, "body slow"):
			return "body_slow"
		case strings.Contains(message, "response body read failed"),
			strings.Contains(message, "body read timeout"):
			return "body_read_failed"
		}
		return "other"
	}
	return shuihuoImageAttemptKindOfStatus(status)
}

func shuihuoImageAttemptKindOfStatus(status int) string {
	switch {
	case status == 0:
		return "other"
	case status >= 200 && status < 300:
		return "ok"
	case status >= 400 && status < 500:
		return "http_4xx"
	case status >= 500:
		return "http_5xx"
	}
	return "other"
}

// 默认所有 https 出图端点都走 HTTP/2；只有已知会重置长流的旧网关保留
// HTTP/1.1 反向例外（shuihuoImageGenerationLegacyHTTP1Hosts）。
func shuihuoImageGenerationHTTPClient(endpoint string) *http.Client {
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	host := ""
	if err == nil {
		host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(parsed.Hostname()), "."))
	}
	if shuihuoImageGenerationLegacyHTTP1Host(host) {
		return shuihuoImageSharedHTTP1Client()
	}
	if shuihuoImageGenerationIsolatedHTTP2Host(host) {
		// This gateway only speaks HTTP/2, but one broken long response used to
		// poison every concurrent stream on the shared connection. Give each
		// paid generation its own HTTP/2 connection so failures stay isolated.
		transport := newShuihuoImageTransport(true)
		transport.DisableKeepAlives = true
		// NewClientWithTransport 在 timeout<=0 时会回落到 DefaultTimeout（120s），
		// 曾导致这个网关的慢正文在 120 秒被整条请求超时掐断（HTTP 200 已计费但图片
		// 取不回）。改用显式“不设总超时”的构造，正文仍由 stall/total 分层计时收口。
		return safehttp.NewClientWithTransportWithoutTimeout(transport)
	}
	return shuihuoImageSharedHTTP2Client()
}

var shuihuoImageGenerationLegacyHTTP1Hosts = map[string]struct{}{}

var shuihuoImageGenerationIsolatedHTTP2Hosts = map[string]struct{}{
	"ggwk1.online":     {},
	"www.ggwk1.online": {},
}

func shuihuoImageGenerationLegacyHTTP1Host(host string) bool {
	if host == "" {
		return false
	}
	// Tencent Zeabur preview 域名可能只提供 HTTP/2，明确保留在允许 h2 的范围。
	if host == "nvewvip.preview.tencent-zeabur.cn" || strings.HasSuffix(host, ".preview.tencent-zeabur.cn") {
		return false
	}
	_, ok := shuihuoImageGenerationLegacyHTTP1Hosts[host]
	return ok
}

func shuihuoImageGenerationIsolatedHTTP2Host(host string) bool {
	_, ok := shuihuoImageGenerationIsolatedHTTP2Hosts[host]
	return ok
}

func shuihuoImageDownloadHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ForceAttemptHTTP2 = true
	// 25s 超时实测是回归：img2.yydsbest.com 对大图常常需要几十秒才读完，
	// 收到 200 但 25s 内读不完会被判失败，导致整批一张都落不了库。恢复
	// 到 120s，让正常下载有机会完成；真正的抖动仍由重试和 b64 兜底处理。
	return safehttp.NewClientWithTransport(120*time.Second, transport)
}

// errShuihuoImageContentPolicy 表示图片模型按内容审核拒收了提示词。
// 调用方必须跳过这张图并继续其它图片，不能把它当成可重试的临时故障。
var errShuihuoImageContentPolicy = fmt.Errorf("图片提示词被内容审核拒绝")

// errShuihuoImageContentSoftRefusal 表示上游没有返回标准 JSON 错误，而是用
// 对话式软拒绝（“当前无法生成图片”“请上传需要编辑的原图”）。它与硬审核
// 拒绝共用改写链路，但必须单独归类，便于区分"改写提示词"还是"换通道"。
var errShuihuoImageContentSoftRefusal = fmt.Errorf("图片通道以对话方式拒绝了本次请求")

// shuihuoImageSoftRefusalError 同时满足两个 sentinel：
//   - errors.Is(err, errShuihuoImageContentPolicy) == true：仍走提示词改写链路；
//   - errors.Is(err, errShuihuoImageContentSoftRefusal) == true：归类为软拒绝。
type shuihuoImageSoftRefusalError struct {
	message string
}

func (e *shuihuoImageSoftRefusalError) Error() string { return e.message }

func (e *shuihuoImageSoftRefusalError) Unwrap() error { return errShuihuoImageContentPolicy }

func (e *shuihuoImageSoftRefusalError) Is(target error) bool {
	return target == errShuihuoImageContentSoftRefusal
}

func newShuihuoImageSoftRefusalError(message string) error {
	return &shuihuoImageSoftRefusalError{message: fmt.Sprintf("%s：%s", errShuihuoImageContentSoftRefusal, strings.TrimPrefix(strings.TrimSpace(message), ": "))}
}

// shuihuoImageResponseSoftRefusal 识别对话式软拒绝。部分网关不返回标准
// JSON 错误，而是让模型回一句自然语言；这类响应不能混进 http_4xx 重试。
func shuihuoImageResponseSoftRefusal(statusCode int, raw []byte) bool {
	if statusCode != http.StatusBadRequest && statusCode != http.StatusForbidden &&
		statusCode != http.StatusUnprocessableEntity && statusCode != http.StatusOK {
		return false
	}
	return shuihuoImageResponseSoftRefusalText(raw)
}

func shuihuoImageResponseSoftRefusalText(raw []byte) bool {
	detail := strings.ToLower(strings.TrimSpace(string(raw)))
	if detail == "" {
		return false
	}
	for _, keyword := range []string{
		"请上传需要编辑的原图", "上传需要编辑的原图", "上传原图",
		"当前无法生成图片", "目前无法生成图片", "暂时无法生成图片", "无法生成图片",
		"未能成功生成图片", "未能完成", "无法完成", "重新发送",
		"unable to generate", "could not generate", "please resend", "upload the original image",
	} {
		if strings.Contains(detail, keyword) {
			return true
		}
	}
	return false
}

// errShuihuoImageReferenceUnusable 表示上游网关读不到我们上传的参考图
// （常见于 /edits 端点只接受公网图片或字段格式不兼容）。这不是内容审核
// 拒绝，也不该让整本书失败：调用方应去掉参考图，改用纯文生图再试一次。
var errShuihuoImageReferenceUnusable = fmt.Errorf("图片通道无法读取参考图")

// errShuihuoImageDownloadUnusable 表示图已经生成成功，但返回的 CDN 地址
// 下载不下来。调用方可以改用 base64 响应重发一次，绕开这个不稳定的 CDN。
var errShuihuoImageDownloadUnusable = fmt.Errorf("生成的图片下载失败")

// errShuihuoImageChatChannel 表示上游把生图请求当成聊天请求，返回 400 +
// 聊天式自然语言（不是图像通道 JSON）。这是可重试的通道错误，不应把页面
// 永久标失败；调用方应在有限次数内退避重试。
var errShuihuoImageChatChannel = fmt.Errorf("图片通道返回的是聊天回复而不是图片")

// shuihuoImageChatChannelError 保留上游状态码与原文摘要，使重试日志和最终
// 失败信息都能看出是哪一次重试、上游回了什么。Error() 只给中文结论，
// 上游原文另起一行，交给界面的「展开原始信息」呈现。
type shuihuoImageChatChannelError struct {
	StatusCode  int
	BodySummary string
}

func (e shuihuoImageChatChannelError) Error() string {
	message := fmt.Sprintf("图片通道返回的是聊天回复而不是图片（HTTP %d），当前配置的模型可能不是绘图模型", e.StatusCode)
	return appendShuihuoUpstreamRaw(message, e.BodySummary)
}

func (e shuihuoImageChatChannelError) Unwrap() error { return errShuihuoImageChatChannel }

// shuihuoImageUpstreamError 用于图片通道的通用拒绝/无效响应：界面主文案只放
// 中文结论，上游原文作为第二行保留给「展开原始信息」。
type shuihuoImageUpstreamError struct {
	Message string
	Raw     string
	// StatusCode 是上游 HTTP 状态码。分类“确定性失败/是否回补”必须看它，
	// 不能再解析错误文案（文案会中文化，关键词匹配会漏判）。
	StatusCode int
}

func (e shuihuoImageUpstreamError) Error() string {
	return appendShuihuoUpstreamRaw(e.Message, e.Raw)
}

func (e shuihuoImageUpstreamError) Unwrap() error { return batchfactoryv11.ErrUpstream }

// shuihuoImageStatusRejectionMessage 按上游状态码给出中文结论：401/403 是渠道
// 凭证问题（Key 无效或没有出图权限），不能再说成“模型不支持出图”；其余非 2xx
// 状态才归到通道拒绝这一类。
func shuihuoImageStatusRejectionMessage(status int, raw string) string {
	// 余额不足通常也以 403 返回。必须先单独识别，否则会被误报成“Key 无效”，
	// 用户按 Key 问题排查半天，实际只是账号没钱了。
	if message, ok := shuihuoImageQuotaShortfallMessage(raw); ok {
		return message
	}
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return fmt.Sprintf("图片通道鉴权失败（HTTP %d），请检查渠道 Key 是否有效、是否具备出图权限", status)
	}
	return fmt.Sprintf("图片通道拒绝了本次请求（HTTP %d），当前配置的模型可能不支持出图", status)
}

var shuihuoImageQuotaAmountPattern = regexp.MustCompile(`[¥￥]\s*([0-9]+(?:\.[0-9]+)?)`)

// shuihuoImageQuotaShortfallMessage 把上游的余额不足原文翻译成一句中文结论，
// 并尽量带上「剩余 / 需要」两个金额，方便用户判断要充多少。
func shuihuoImageQuotaShortfallMessage(raw string) (string, bool) {
	lower := strings.ToLower(raw)
	isQuota := strings.Contains(lower, "not enough quota") ||
		strings.Contains(lower, "insufficient quota") ||
		strings.Contains(lower, "insufficient balance") ||
		strings.Contains(lower, "pre-consume") ||
		strings.Contains(raw, "余额不足") ||
		strings.Contains(raw, "额度不足")
	if !isQuota {
		return "", false
	}
	amounts := shuihuoImageQuotaAmountPattern.FindAllStringSubmatch(raw, -1)
	if len(amounts) >= 2 {
		return fmt.Sprintf("上游账号余额不足：剩余 ¥%s，本次出图需要 ¥%s，请充值后重试", amounts[0][1], amounts[1][1]), true
	}
	return "上游账号余额不足，请充值后重试", true
}

func appendShuihuoUpstreamRaw(message, raw string) string {
	message = strings.TrimSpace(message)
	raw = strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
	// 上游原文自己可能以冒号开头（例如渠道返回 “: Incorrect API key...”），
	// 直接拼在“上游原文：”后面会变成双冒号，拼接前先去掉开头的冒号与空白。
	raw = strings.TrimSpace(strings.TrimLeft(raw, ":："))
	if raw == "" {
		return message
	}
	if runes := []rune(raw); len(runes) > 240 {
		raw = string(runes[:240]) + "..."
	}
	return message + "\n上游原文：" + raw
}

var shuihuoImagePromptSofteningRules = []struct {
	from string
	to   string
}{
	{"鲜血淋漓", "衣襟破损"},
	{"血泊", "地面痕迹"},
	{"血迹", "暗色痕迹"},
	{"鲜血", "暗红"},
	{"血腥", "暗色"},
	{"血流", "痕迹"},
	{"伤口", "痕迹"},
	{"抓伤", "痕迹"},
	{"尸体", "倒地的人"},
	{"残肢", "倒地的人"},
	{"断手", "手部"},
	{"刀刺", "逼近"},
	{"刺入", "触及"},
	{"砍", "挥"},
	{"劈", "挥"},
	{"斩", "挥"},
	{"捅", "推"},
	{"殴打", "冲突"},
	{"扇耳光", "争执"},
	{"掐住", "扣住"},
	{"掐死", "制住"},
	{"勒死", "制住"},
	{"勒住", "扣住"},
	{"掐", "扣"},
	{"踢", "碰"},
	{"踹", "碰"},
	{"摔在", "跌在"},
	{"鞭打", "斥责"},
	{"酷刑", "威压"},
	{"杀戮", "冲突"},
	{"杀人", "加害"},
	{"杀死", "制住"},
	{"杀了", "制住"},
	{"自尽", "绝望"},
	{"上吊", "绝望"},
	{"割腕", "绝望"},
	{"坠楼", "跌落"},
	{"囚禁", "限制"},
	{"绑架", "带走"},
	{"撕扯", "拉扯"},
	{"裸露", "衣衫单薄"},
	{"赤裸", "衣衫单薄"},
	{"胴体", "身影"},
	{"肉体", "身形"},
	{"私处", "衣襟"},
	{"大腿内侧", "腿侧"},
	{"胸部", "肩颈"},
	{"胸脯", "肩颈"},
	{"臀部", "背影"},
	{"臀", "背影"},
	{"呻吟", "低声喘息"},
	{"娇喘", "呼吸急促"},
	{"情欲", "情绪"},
	{"淫", "轻浮"},
	{"媚肉", "衣料"},
	{"抚摸", "轻触"},
	{"亲吻", "靠近"},
	{"吻", "贴近"},
	{"床上", "榻边"},
	{"衣不蔽体", "衣衫单薄"},
	{"衣衫尽褪", "衣衫凌乱"},
	{"肌肤", "肤色"},
	{"白皙", "浅淡"},
	{"春光", "景象"},
	{"裸", "着衣"},
	{"鲜血", "暗红"},
	{"血腥", "暗色"},
	{"尸体", "倒地的人"},
	{"杀戮", "冲突"},
	{"掐死", "制住"},
	{"勒死", "制住"},
}

// sanitizeShuihuoImagePrompt 在送图前把容易触发内容审核的词换成含蓄表达。
// 只做词语替换并追加一句画面尺度说明，不改变剧情走向和构图要求。
func sanitizeShuihuoImagePrompt(prompt string) string {
	result := prompt
	changed := false
	for _, rule := range shuihuoImagePromptSofteningRules {
		if !strings.Contains(result, rule.from) {
			continue
		}
		result = strings.ReplaceAll(result, rule.from, rule.to)
		changed = true
	}
	if strings.TrimSpace(result) == "" {
		return prompt
	}
	// 没有任何命中时保持原提示词不变，避免无关提示词被追加说明后改变出图效果。
	if !changed {
		return result
	}
	const note = "画面以含蓄、艺术化的方式呈现，不表现露骨情色、暴力冲突或血腥细节。"
	if !strings.Contains(result, note) {
		result = strings.TrimSpace(result) + "\n\n" + note
	}
	return result
}

// shuihuoImagePromptAggressiveRules 是第二轮降级用的更保守替换表。
// 只在第一轮降敏后仍被内容审核拒绝时使用，尽量保留构图和情绪，去掉一切
// 可能被判定为暴力或情色的细节描述。
var shuihuoImagePromptAggressiveRules = []struct {
	from string
	to   string
}{
	// 纸质道具（合同、声明、信件）被“撕开/破损”的描述经常被网关软拒绝，
	// 保留内容但改成完整的纸张，避免整张资产图被跳过。
	{"撕成两截", "分成两份"},
	{"撕碎", "分成碎片"},
	{"撕裂", "分开"},
	{"衣襟破损", "衣襟整齐"},
	{"破损", "完整"},
	{"残破", "完整"},
	{"地面痕迹", "地面"},
	{"暗色痕迹", "地面"},
	{"暗红", "深色"},
	{"暗色", "昏暗"},
	{"痕迹", "细节"},
	{"倒地的人", "站立的人物"},
	{"制住", "拦下"},
	{"逼近", "靠近"},
	{"冲突", "对峙"},
	{"争执", "交谈"},
	{"斥责", "劝说"},
	{"威压", "注视"},
	{"加害", "阻拦"},
	{"绝望", "沉默"},
	{"跌落", "重心不稳"},
	{"限制", "陪同"},
	{"带走", "离开"},
	{"拉扯", "相扶"},
	{"衣衫单薄", "日常衣着"},
	{"衣衫凌乱", "衣着整齐"},
	{"身影", "人物"},
	{"身形", "人物"},
	{"衣襟", "衣领"},
	{"肩颈", "肩部"},
	{"背影", "侧影"},
	{"低声喘息", "呼吸"},
	{"呼吸急促", "呼吸"},
	{"情欲", "情绪"},
	{"轻浮", "轻快"},
	{"衣料", "衣物"},
	{"轻触", "相扶"},
	{"相邻", "并肩"},
	{"榻边", "屋内"},
	{"肤色", "光影"},
	{"浅淡", "柔和"},
	{"景象", "画面"},
}

const shuihuoImagePromptAggressiveNote = "请以完全含蓄、克制的方式表现：画面干净，人物衣着整齐；不出现任何血腥、伤口、暴力动作、情色暗示或敏感身体描写；只保留人物姿态、场景与情绪氛围。"

// sanitizeShuihuoImagePromptAggressive 在常规降敏之上再做一轮更保守的替换，
// 用于被内容审核拒绝后的二次生成尝试。
func sanitizeShuihuoImagePromptAggressive(prompt string) string {
	result := prompt
	for _, rule := range shuihuoImagePromptSofteningRules {
		if !strings.Contains(result, rule.from) {
			continue
		}
		result = strings.ReplaceAll(result, rule.from, rule.to)
	}
	for _, rule := range shuihuoImagePromptAggressiveRules {
		if !strings.Contains(result, rule.from) {
			continue
		}
		result = strings.ReplaceAll(result, rule.from, rule.to)
	}
	if strings.TrimSpace(result) == "" {
		return prompt
	}
	return strings.TrimSpace(result) + "\n\n" + shuihuoImagePromptAggressiveNote
}

// shuihuoImagePromptViolenceRules 专门处理“暴力内容”类拒绝。上游对武器、
// 流血、伤口、昏迷、殴打这几类最敏感，普通降敏词表覆盖不到，所以单独
// 做一轮改写：保留分格、构图、台词和情绪，把伤害动作换成安全等价表达。
var shuihuoImagePromptViolenceRules = []struct {
	from string
	to   string
}{
	{"伸手抢夺失控", "情绪激动"},
	{"餐叉尖端狠狠擦过夏夏脸颊", "餐具被放回桌面，夏夏侧身避开"},
	{"抢刀叉", "伸手去拿餐具"},
	{"刀叉", "餐具"},
	{"餐叉", "餐具"},
	{"叉尖", "餐具边缘"},
	{"刀尖", "餐盘边缘"},
	{"匕首", "餐具"},
	{"剪刀", "文具"},
	{"狠狠擦过", "轻轻掠过"},
	{"狠狠", "轻轻"},
	{"擦过", "掠过"},
	{"划破", "掠过"},
	{"划伤", "掠过"},
	{"鲜血飞溅而出", "餐巾扬起"},
	{"鲜血飞溅", "餐巾扬起"},
	{"鲜血", "餐巾"},
	{"血迹渗开", "纸巾褶皱"},
	{"血迹", "餐巾"},
	{"血污", "污渍"},
	{"血痕", "泪痕"},
	{"伤口", "脸颊"},
	{"流血", "发丝凌乱"},
	{"殷红", "淡红"},
	{"刺入", "碰触"},
	{"刺穿", "碰触"},
	{"捅入", "碰触"},
	{"砍向", "挥向"},
	{"劈下", "落下"},
	{"掐住", "扶住"},
	{"掐着", "扶着"},
	{"扇耳光", "挡在身前"},
	{"殴打", "阻拦"},
	{"拳打", "推挡"},
	{"踹开", "避开"},
	{"踢开", "避开"},
	{"口吐白沫", "脸色苍白"},
	{"不省人事", "神情恍惚"},
	{"昏迷", "虚弱"},
	{"痛苦倒下", "站立不稳"},
	{"倒地", "坐下"},
	{"抢夺", "伸手去接"},
	{"夺扯", "递出"},
}

const shuihuoImagePromptViolenceNote = "【内容合规约束】请以人物表情、自然姿态、适当距离变化和对话气泡呈现情绪转折；保持原有画风、分格布局、台词含义、场景元素与人物身份连续。"

// sanitizeShuihuoImagePromptViolence 是针对暴力审核的整页合规改写。它只在
// 提示词确实命中暴力词时生效，返回空串以外的结果会作为第三次生成尝试。
func sanitizeShuihuoImagePromptViolence(prompt string) string {
	result := prompt
	changed := false
	for _, rule := range shuihuoImagePromptViolenceRules {
		if !strings.Contains(result, rule.from) {
			continue
		}
		result = strings.ReplaceAll(result, rule.from, rule.to)
		changed = true
	}
	if !changed || strings.TrimSpace(result) == "" {
		return prompt
	}
	return strings.TrimSpace(result) + "\n\n" + shuihuoImagePromptViolenceNote
}

type shuihuoImageEventContextKey struct{}

// withShuihuoImageEvent 把出图过程中的去敏、审核拒绝、降级重试事件交给调用方
// 写进任务日志。没有绑定日志时为空操作。
func withShuihuoImageEvent(ctx context.Context, report func(stage, message string)) context.Context {
	if report == nil {
		return ctx
	}
	return context.WithValue(ctx, shuihuoImageEventContextKey{}, report)
}

func reportShuihuoImageEvent(ctx context.Context, stage, message string) {
	if ctx == nil || strings.TrimSpace(message) == "" {
		return
	}
	if report, ok := ctx.Value(shuihuoImageEventContextKey{}).(func(stage, message string)); ok && report != nil {
		report(stage, message)
	}
}

func shuihuoImageResponseContentRejected(statusCode int, raw []byte) bool {
	if statusCode != http.StatusBadRequest && statusCode != http.StatusForbidden && statusCode != http.StatusUnprocessableEntity {
		return false
	}
	detail := strings.ToLower(string(raw))
	for _, keyword := range []string{"违规", "违反", "内容政策", "content policy", "content_policy", "moderation", "safety", "unsafe", "blocked"} {
		if strings.Contains(detail, keyword) {
			return true
		}
	}
	return false
}

// shuihuoImageProviderChatChannelResponse 识别“生图请求被上游当成聊天请求”的
// 400：响应体不是图像通道 JSON，而是聊天式自然语言。该错误可重试，不能当作
// 页面永久失败。
func shuihuoImageProviderChatChannelResponse(statusCode int, raw []byte) bool {
	if statusCode != http.StatusBadRequest {
		return false
	}
	body := strings.TrimSpace(string(raw))
	if body == "" {
		return false
	}
	// 聊天模型经常把拒绝理由包在 JSON 里（error.message / message / choices），
	// 只判断“响应体不是 JSON”会把这类回复漏成通用上游错误，所以连同解析后的
	// 自然语言一起匹配。
	detail := strings.ToLower(body + "\n" + shuihuoImageProviderErrorDetail(raw))
	for _, keyword := range []string{
		"在这个聊天", "在此聊天", "聊天中直接生成", "聊天窗口", "聊天里",
		"in this chat", "this chat", "chat interface", "chat window",
	} {
		if strings.Contains(detail, keyword) {
			return true
		}
	}
	return false
}

// shuihuoImageProviderRejectedReferences reports whether a 4xx response means
// the gateway could not read the reference images we uploaded, rather than the
// prompt itself being rejected. The wording differs between gateways, so match
// the common reference-image phrases.
func shuihuoImageProviderRejectedReferences(raw []byte) bool {
	detail := strings.ToLower(string(raw))
	for _, keyword := range []string{"参考图", "参考图片", "reference image", "reference_image", "original image", "上传的图片"} {
		if strings.Contains(detail, keyword) {
			return true
		}
	}
	return false
}

func shuihuoImageModelSupportsEdits(model map[string]any, modelID string) bool {
	if value, ok := model["supportsImageEdits"].(bool); ok {
		return value
	}
	// Several OpenAI-compatible gateways expose gpt-image models only through
	// /images/generations. Sending their comic asset references to /edits makes
	// the request hang or return an unsupported-operation response. The full
	// comic prompt still carries the visual descriptions, so generation remains
	// useful while the model is configured without explicit edit support.
	return !shuihuoIsGPTImageModel(modelID)
}

func shuihuoIsGPTImageModel(modelID string) bool {
	return strings.Contains(strings.ToLower(strings.TrimSpace(modelID)), "gpt-image")
}

func shuihuoImageProviderErrorDetail(raw []byte) string {
	message := strings.TrimSpace(string(raw))
	if message == "" {
		return ""
	}
	var payload map[string]any
	if json.Unmarshal(raw, &payload) == nil {
		if nested, ok := payload["error"].(map[string]any); ok {
			for _, key := range []string{"message", "detail", "code"} {
				if value := strings.TrimSpace(fmt.Sprint(nested[key])); value != "" && value != "<nil>" {
					message = value
					break
				}
			}
		} else if value, ok := payload["message"].(string); ok && strings.TrimSpace(value) != "" {
			message = strings.TrimSpace(value)
		}
	}
	message = strings.Join(strings.Fields(message), " ")
	if len([]rune(message)) > 240 {
		message = string([]rune(message)[:240]) + "..."
	}
	return ": " + message
}

func newShuihuoImageEditRequest(ctx context.Context, generationEndpoint, modelID, prompt, size string, references []string) (*http.Request, error) {
	return newShuihuoImageEditRequestWithResponseFormat(ctx, generationEndpoint, modelID, prompt, size, references, "b64_json")
}

func newShuihuoImageEditRequestWithResponseFormat(ctx context.Context, generationEndpoint, modelID, prompt, size string, references []string, responseFormat string) (*http.Request, error) {
	editEndpoint := strings.TrimSuffix(generationEndpoint, "/generations") + "/edits"
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("model", modelID)
	_ = writer.WriteField("prompt", prompt)
	_ = writer.WriteField("n", "1")
	_ = writer.WriteField("size", size)
	if strings.TrimSpace(responseFormat) != "" {
		_ = writer.WriteField("response_format", responseFormat)
	}
	client := safehttp.NewClientWithTimeout(60 * time.Second)
	for index, reference := range references {
		payload, mediaType, err := readShuihuoReferenceImage(ctx, client, reference)
		if err != nil {
			_ = writer.Close()
			return nil, err
		}
		part, err := writer.CreateFormFile("image[]", fmt.Sprintf("reference-%d.%s", index+1, shuihuoImageExtension(mediaType)))
		if err != nil {
			_ = writer.Close()
			return nil, err
		}
		if _, err := part.Write(payload); err != nil {
			_ = writer.Close()
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, editEndpoint, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

func readShuihuoReferenceImage(ctx context.Context, client *http.Client, reference string) ([]byte, string, error) {
	value := strings.TrimSpace(reference)
	if strings.HasPrefix(value, "data:") {
		comma := strings.IndexByte(value, ',')
		if comma <= len("data:") {
			return nil, "", fmt.Errorf("invalid reference image data")
		}
		meta := value[len("data:"):comma]
		parts := strings.Split(meta, ";")
		mediaType := parts[0]
		if len(parts) < 2 || parts[1] != "base64" || mediaType == "" {
			return nil, "", fmt.Errorf("unsupported reference image data")
		}
		payload, err := base64.StdEncoding.DecodeString(value[comma+1:])
		if err != nil || len(payload) == 0 {
			return nil, "", fmt.Errorf("invalid reference image data")
		}
		return payload, mediaType, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, value, nil)
	if err != nil {
		return nil, "", err
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, "", fmt.Errorf("reference image download returned HTTP %d", response.StatusCode)
	}
	payload, err := io.ReadAll(io.LimitReader(response.Body, 16<<20))
	if err != nil || len(payload) == 0 {
		return nil, "", fmt.Errorf("reference image download returned empty content")
	}
	mediaType := response.Header.Get("Content-Type")
	if index := strings.IndexByte(mediaType, ';'); index >= 0 {
		mediaType = mediaType[:index]
	}
	if mediaType == "" {
		mediaType = detectedShuihuoImageMediaType(payload)
	}
	return payload, mediaType, nil
}

func shuihuoImageExtension(mediaType string) string {
	switch strings.ToLower(strings.TrimSpace(mediaType)) {
	case "image/jpeg", "image/jpg":
		return "jpg"
	case "image/webp":
		return "webp"
	default:
		return "png"
	}
}

const shuihuoRelayImageRequestTimeout = 12 * time.Minute

func generateShuihuoRelayImage(ctx context.Context, files *localartifact.Store, target shuihuoAssetImageTarget, prompt string, references []string, baseURL, token, account string) (string, string, string, error) {
	if token == "" {
		return "", "", "", fmt.Errorf("%w: 中继通道缺少访问凭证，请在配置中补全", batchfactoryv11.ErrUnavailable)
	}
	payload := map[string]any{"model": relayImageModel(target.ModelID), "prompt": prompt, "ratio": target.AspectRatio}
	if account != "" {
		payload["account"] = account
	}
	if len(references) > 0 {
		payload["ref_image"] = references
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/images/generations", bytes.NewReader(body))
	if err != nil {
		return "", "", "", fmt.Errorf("%w: 无法向图片中继通道发起请求，请检查网络配置", batchfactoryv11.ErrUpstream)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	response, err := safehttp.NewClientWithTimeout(shuihuoRelayImageRequestTimeout).Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("%w: 图片中继通道请求失败：%w", batchfactoryv11.ErrUpstream, err)
	}
	defer response.Body.Close()
	raw, readErr := io.ReadAll(io.LimitReader(response.Body, 32<<20))
	if readErr != nil || response.StatusCode < 200 || response.StatusCode >= 300 {
		detail := string(raw)
		if len(detail) > 500 {
			detail = detail[:500]
		}
		return "", "", "", shuihuoImageUpstreamError{
			Message:    shuihuoImageStatusRejectionMessage(response.StatusCode, detail),
			Raw:        detail,
			StatusCode: response.StatusCode,
		}
	}
	var result struct {
		Data []struct {
			URL      string `json:"url"`
			LocalURL string `json:"local_url"`
		} `json:"data"`
	}
	if json.Unmarshal(raw, &result) != nil || len(result.Data) == 0 {
		return "", "", "", fmt.Errorf("%w: 图片通道返回的内容不是有效图片数据，请检查该模型是否支持出图", batchfactoryv11.ErrUpstream)
	}
	imageURL := result.Data[0].LocalURL
	if imageURL == "" {
		imageURL = result.Data[0].URL
	}
	if imageURL == "" {
		return "", "", "", fmt.Errorf("%w: 图片通道没有返回图片地址", batchfactoryv11.ErrUpstream)
	}
	imageReq, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("%w: 图片通道返回的图片地址无效", batchfactoryv11.ErrUpstream)
	}
	imageReq.Header.Set("Authorization", "Bearer "+token)
	imageResponse, err := safehttp.NewClientWithTimeout(60 * time.Second).Do(imageReq)
	if err != nil {
		return "", "", "", fmt.Errorf("%w: 图片下载失败：%v", batchfactoryv11.ErrUpstream, err)
	}
	defer imageResponse.Body.Close()
	if imageResponse.StatusCode < 200 || imageResponse.StatusCode >= 300 {
		return "", "", "", fmt.Errorf("%w: 图片下载失败（HTTP %d），请稍后重试", batchfactoryv11.ErrUpstream, imageResponse.StatusCode)
	}
	mediaType := imageResponse.Header.Get("Content-Type")
	if mediaType == "" {
		mediaType = "image/png"
	}
	imageBytes, err := io.ReadAll(io.LimitReader(imageResponse.Body, 32<<20))
	if err != nil || len(imageBytes) == 0 {
		return "", "", "", fmt.Errorf("%w: 图片通道返回的图片内容无效", batchfactoryv11.ErrUpstream)
	}
	artifactID, err := localexecutor.NewArtifactID()
	if err != nil {
		return "", "", "", err
	}
	saved, err := files.SaveImage(artifactID, mediaType, bytes.NewReader(imageBytes))
	if err != nil {
		return "", "", "", err
	}
	// The relay is the HTTP transport, but it is still the local Doubao
	// provider from the asset store's perspective. Keep the source value within
	// the existing allow-list so generated images can be persisted and selected
	// as the primary asset image.
	return mediaType, saved.StorageRef, batchfactoryv11.VideoProviderDoubaoLocal, nil
}

func relayImageModel(model string) string {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "seedream 4.0":
		return "doubao-image-4.0"
	case "seedream 5.0 flash":
		return "doubao-image-5.0-flash"
	case "seedream 5.0 pro":
		return "doubao-image-5.0-flash"
	default:
		return "doubao-image-4.5"
	}
}

type shuihuoAssetImageTarget struct {
	Provider     string
	ModelID      string
	AspectRatio  string
	CatalogModel map[string]any
}

func resolveShuihuoAssetImageTarget(asset batchfactoryv11.BookAsset, fallbackModelID, fallbackAspectRatio, fallbackProvider string, catalog []map[string]any) (shuihuoAssetImageTarget, error) {
	selectedModelID, resolvedAspectRatio := resolveShuihuoAssetImageSettings(asset, fallbackModelID, fallbackAspectRatio)
	return resolveShuihuoImageTarget(selectedModelID, fallbackModelID, resolvedAspectRatio, fallbackProvider, catalog)
}

// resolveShuihuoAssetImageTargetWithChannels 在资产级重试时也要支持
// custom:/channel: 渠道，否则单个资产的模型覆盖会被旧目录静默忽略。
func resolveShuihuoAssetImageTargetWithChannels(ctx context.Context, userConfig *UserConfigStore, owner string, asset batchfactoryv11.BookAsset, fallbackModelID, fallbackAspectRatio, fallbackProvider string, catalog []map[string]any) (shuihuoAssetImageTarget, error) {
	selectedModelID, resolvedAspectRatio := resolveShuihuoAssetImageSettings(asset, fallbackModelID, fallbackAspectRatio)
	return resolveShuihuoImageTargetWithChannels(ctx, userConfig, owner, selectedModelID, fallbackModelID, resolvedAspectRatio, fallbackProvider, catalog)
}

// resolveShuihuoImageTargetWithChannels 先按旧 modelCatalog 解析；当选中的是
// custom:/channel: 渠道时，用 API 接口配置展开成临时目录条目，避免“页面选了
// 新渠道、实际仍打旧地址/旧 Key”。
func resolveShuihuoImageTargetWithChannels(ctx context.Context, userConfig *UserConfigStore, owner, selectedModelID, fallbackModelID, aspectRatio, fallbackProvider string, catalog []map[string]any) (shuihuoAssetImageTarget, error) {
	if entry, matched, err := resolveIntegrationImageCatalogEntry(ctx, userConfig, owner, selectedModelID); matched {
		if err != nil {
			return shuihuoAssetImageTarget{}, withErrorCode(err, ErrorCodeImageModelNotConfigured)
		}
		catalog = append([]map[string]any{entry}, catalog...)
	} else if entry, legacyMatched, legacyErr := resolveLegacyIntegrationImageCatalogEntry(ctx, userConfig, owner, selectedModelID); legacyMatched {
		if legacyErr != nil {
			return shuihuoAssetImageTarget{}, withErrorCode(legacyErr, ErrorCodeImageModelNotConfigured)
		}
		catalog = append([]map[string]any{entry}, catalog...)
	}
	return resolveShuihuoImageTarget(selectedModelID, fallbackModelID, aspectRatio, fallbackProvider, catalog)
}

func resolveShuihuoImageTarget(selectedModelID, fallbackModelID, aspectRatio, fallbackProvider string, catalog []map[string]any) (shuihuoAssetImageTarget, error) {
	selectedModelID = strings.TrimSpace(selectedModelID)
	aspectRatio = strings.TrimSpace(aspectRatio)
	provider := batchfactoryv11.NormalizeVideoProviderForHTTP(fallbackProvider)
	if provider == batchfactoryv11.VideoProviderDoubaoLocal && selectedModelID != "" {
		return shuihuoAssetImageTarget{Provider: provider, ModelID: selectedModelID, AspectRatio: aspectRatio}, nil
	}
	for _, candidate := range catalog {
		if !strings.EqualFold(modelString(candidate, "kind"), "image") {
			continue
		}
		if selectedModelID != "" && !strings.EqualFold(modelString(candidate, "id"), selectedModelID) && !strings.EqualFold(modelString(candidate, "modelId"), selectedModelID) {
			continue
		}
		if enabled, exists := candidate["enabled"].(bool); exists && !enabled {
			continue
		}
		candidateProvider := modelString(candidate, "provider")
		if candidateProvider == "" && strings.EqualFold(selectedModelID, strings.TrimSpace(fallbackModelID)) {
			candidateProvider = fallbackProvider
		}
		providerModelID := modelString(candidate, "modelId")
		if providerModelID == "" {
			providerModelID = modelString(candidate, "id")
		}
		return shuihuoAssetImageTarget{
			Provider:     batchfactoryv11.NormalizeVideoProviderForHTTP(candidateProvider),
			ModelID:      providerModelID,
			AspectRatio:  aspectRatio,
			CatalogModel: candidate,
		}, nil
	}
	return shuihuoAssetImageTarget{}, withErrorCode(fmt.Errorf("%w: 图片模型未配置；%s", batchfactoryv11.ErrUnavailable, setupHintImage), ErrorCodeImageModelNotConfigured)
}

func resolveShuihuoAssetImageSettings(asset batchfactoryv11.BookAsset, fallbackModelID, _ string) (modelID, aspectRatio string) {
	modelID = strings.TrimSpace(asset.ImageModelID)
	if modelID == "" {
		modelID = strings.TrimSpace(fallbackModelID)
	}
	aspectRatio = strings.TrimSpace(asset.ImageAspectRatio)
	if aspectRatio == "" {
		aspectRatio = "16:9"
	}
	return modelID, aspectRatio
}

func waitForLocalExecutorImage(ctx context.Context, service *localexecutor.Service, files *localartifact.Store, owner, model, prompt, size string, extra ...map[string]any) (string, string, error) {
	model = strings.TrimSpace(model)
	if model == "" {
		model = "Seedream 4.5"
	}
	payload := map[string]any{"type": "image", "model": model, "prompt": prompt, "size": size}
	if len(extra) > 0 {
		for key, value := range extra[0] {
			payload[key] = value
		}
	}
	job, err := service.CreateJob(ctx, owner, localexecutor.CreateJobInput{
		SourceTaskID: "batch-factory:image:" + time.Now().UTC().Format("20060102150405.000000000"),
		Platform:     localexecutor.PlatformDoubao,
		Payload:      payload,
	})
	if err != nil {
		return "", "", fmt.Errorf("%w: 创建豆包图片任务失败: %v", batchfactoryv11.ErrUpstream, err)
	}
	// 豆包图片任务在带参考图和长提示词时可能超过三分钟；只要执行器仍在
	// generating，就继续等待，避免后端先判超时而丢掉即将返回的结果。
	deadline := time.NewTimer(10 * time.Minute)
	defer deadline.Stop()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		current, getErr := service.GetJob(ctx, owner, job.ID)
		if getErr != nil {
			return "", "", fmt.Errorf("%w: 读取豆包图片任务失败: %v", batchfactoryv11.ErrUpstream, getErr)
		}
		switch current.State {
		case localexecutor.JobSucceeded:
			if current.ArtifactID == "" {
				return "", "", fmt.Errorf("%w: 豆包图片任务没有返回图片", batchfactoryv11.ErrUpstream)
			}
			artifact, artifactErr := service.GetArtifact(ctx, owner, current.ArtifactID)
			if artifactErr != nil {
				return "", "", fmt.Errorf("%w: 读取豆包图片结果失败: %v", batchfactoryv11.ErrUpstream, artifactErr)
			}
			if files == nil || artifact.StorageRef == "" {
				return "", "", fmt.Errorf("%w: 豆包图片结果文件不存在", batchfactoryv11.ErrUpstream)
			}
			return artifact.MediaType, artifact.StorageRef, nil
		case localexecutor.JobFailed, localexecutor.JobCancelled:
			return "", "", fmt.Errorf("%w: 豆包图片任务失败: %s", batchfactoryv11.ErrUpstream, artifactErrorMessage(current.ErrorMessage))
		}
		select {
		case <-ctx.Done():
			return "", "", ctx.Err()
		case <-deadline.C:
			return "", "", fmt.Errorf("%w: 豆包图片任务超时", batchfactoryv11.ErrUpstream)
		case <-ticker.C:
		}
	}
}

func artifactErrorMessage(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "未提供错误信息"
	}
	if len(value) > 300 {
		return value[:300]
	}
	return value
}

func generateShuihuoSegmentAudio(ctx context.Context, store batchfactoryv11.Store, files *localartifact.Store, owner string, batch batchfactoryv11.Batch, segmentID string, catalogStores ...novelfetchworkshop.Store) (map[string]any, error) {
	if len(batch.Books) == 0 {
		return nil, fmt.Errorf("%w: project has no source book", batchfactoryv11.ErrInvalid)
	}
	var text string
	for _, segment := range storedShuihuoSegments(batch.Books[0]) {
		if modelString(segment, "id") != segmentID {
			continue
		}
		text = modelString(segment, "subtitleText")
		if text == "" {
			text = modelString(segment, "sourceText")
		}
		break
	}
	if text == "" {
		return nil, fmt.Errorf("%w: segment not found or empty", batchfactoryv11.ErrNotFound)
	}
	input := ttsRequest{Input: text, Voice: "zh-CN-XiaoxiaoNeural", Speed: 1.8, Pitch: "10", Style: "general"}
	var payload []byte
	var mediaType string
	var err error
	if len(catalogStores) > 0 {
		if model, found := configuredTTSModel(ctx, catalogStores[0], owner); found {
			payload, mediaType, err = requestConfiguredTTSBytes(ctx, input, model)
		}
	}
	if payload == nil {
		body, _ := json.Marshal(map[string]any{"input": text, "voice": input.Voice, "speed": input.Speed, "pitch": input.Pitch, "style": input.Style})
		req, requestErr := http.NewRequestWithContext(ctx, http.MethodPost, shuihuoDefaultTTSServiceURL, bytes.NewReader(body))
		if requestErr != nil {
			return nil, fmt.Errorf("%w: TTS request failed", batchfactoryv11.ErrUpstream)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Host", "tts3.121w.com")
		req.Header.Set("Referer", "http://tts3.121w.com/")
		req.Header.Set("Origin", "http://tts3.121w.com")
		response, requestErr := safehttp.NewClientWithTimeout(120 * time.Second).Do(req)
		if requestErr != nil {
			return nil, fmt.Errorf("%w: TTS request failed: %v", batchfactoryv11.ErrUpstream, requestErr)
		}
		defer response.Body.Close()
		payload, err = io.ReadAll(io.LimitReader(response.Body, 32<<20))
		if err != nil || response.StatusCode < 200 || response.StatusCode >= 300 || len(payload) == 0 {
			return nil, fmt.Errorf("%w: TTS provider returned HTTP %d", batchfactoryv11.ErrUpstream, response.StatusCode)
		}
		mediaType = strings.TrimSpace(strings.Split(response.Header.Get("Content-Type"), ";")[0])
	}
	if err != nil {
		return nil, fmt.Errorf("%w: TTS request failed: %v", batchfactoryv11.ErrUpstream, err)
	}
	if mediaType == "" || !strings.HasPrefix(mediaType, "audio/") {
		mediaType = "audio/mpeg"
	}
	artifactID, err := localexecutor.NewArtifactID()
	if err != nil {
		return nil, err
	}
	saved, err := files.SaveMedia(artifactID, mediaType, "mp3", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	media := storedShuihuoMedia(batch.Books[0])
	entry := map[string]any{"id": artifactID, "kind": "audio", "filename": artifactID + ".mp3", "mediaType": mediaType, "storageRef": saved.StorageRef, "segmentId": segmentID, "isPrimary": true, "createdAt": time.Now().UTC().Format(time.RFC3339)}
	media = append(media, entry)
	metadata := cloneMetadata(batch.Books[0].SourceMetadata)
	metadata["shuihuoMedia"] = media
	if _, err := store.UpdateBookMetadata(ctx, owner, batch.ID, batch.Books[0].ID, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata, ExpectedRevision: batch.Books[0].Revision}); err != nil {
		_ = files.Remove(saved.StorageRef)
		return nil, err
	}
	return map[string]any{"id": artifactID, "status": "completed", "kind": "audio", "segmentId": segmentID, "mediaUrl": "/api/shuihuo-production/media/" + artifactID + "/download", "createdAt": entry["createdAt"]}, nil
}

func generateShuihuoSegmentImage(ctx context.Context, store batchfactoryv11.Store, files *localartifact.Store, owner string, batch batchfactoryv11.Batch, segmentID, modelID string, catalogStores ...novelfetchworkshop.Store) (map[string]any, error) {
	if len(batch.Books) == 0 {
		return nil, fmt.Errorf("%w: project has no source book", batchfactoryv11.ErrInvalid)
	}
	var prompt string
	for _, segment := range storedShuihuoSegments(batch.Books[0]) {
		if modelString(segment, "id") != segmentID {
			continue
		}
		prompt = modelString(segment, "imagePrompt")
		if prompt == "" {
			prompt = modelString(segment, "sourceText")
		}
		break
	}
	if prompt == "" {
		return nil, fmt.Errorf("%w: segment not found or has no image prompt", batchfactoryv11.ErrNotFound)
	}
	config, err := loadShuihuoOwnerConfig(ctx, store, owner, catalogStores...)
	if err != nil {
		return nil, err
	}
	var model map[string]any
	for _, candidate := range shuihuoCollection(config, "modelCatalog", nil) {
		if !strings.EqualFold(modelString(candidate, "kind"), "image") || (modelID != "" && modelString(candidate, "modelId") != modelID) {
			continue
		}
		enabled, hasEnabled := candidate["enabled"].(bool)
		if hasEnabled && !enabled {
			continue
		}
		model = candidate
		break
	}
	if model == nil {
		return nil, withErrorCode(fmt.Errorf("%w: 图片模型未配置；%s", batchfactoryv11.ErrUnavailable, setupHintImage), ErrorCodeImageModelNotConfigured)
	}
	endpoint, err := shuihuoImageGenerationEndpoint(modelString(model, "endpoint"))
	if err != nil {
		return nil, fmt.Errorf("%w: 图片模型地址无效，请在资产设置中检查模型接口地址", batchfactoryv11.ErrInvalid)
	}
	apiKey := shuihuoModelCredential(model)
	if apiKey == "" {
		return nil, fmt.Errorf("%w: 图片模型缺少凭证，请在资产设置中补全", batchfactoryv11.ErrUnavailable)
	}
	body, _ := json.Marshal(map[string]any{"model": modelString(model, "modelId"), "prompt": prompt, "n": 1, "size": shuihuoImageSize("16:9"), "response_format": "b64_json"})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: 无法向图片通道发起请求，请检查网络与模型配置", batchfactoryv11.ErrUpstream)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	// 同上：b64 大图正文传输可能超过 120 秒，不能在上游已计费后掐断读取。
	response, err := safehttp.NewClientWithTimeout(10 * time.Minute).Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: 图片通道请求失败：%w", batchfactoryv11.ErrUpstream, err)
	}
	defer response.Body.Close()
	raw, readErr := io.ReadAll(io.LimitReader(response.Body, 16<<20))
	if readErr != nil || response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, shuihuoImageUpstreamError{
			Message:    shuihuoImageStatusRejectionMessage(response.StatusCode, string(raw)),
			Raw:        shuihuoImageProviderErrorDetail(raw),
			StatusCode: response.StatusCode,
		}
	}
	var result struct {
		Data []struct {
			B64JSON string `json:"b64_json"`
			URL     string `json:"url"`
		} `json:"data"`
	}
	if json.Unmarshal(raw, &result) != nil || len(result.Data) == 0 {
		return nil, shuihuoImageUpstreamError{Message: "图片通道返回的内容不是有效图片数据，请检查该模型是否支持出图"}
	}
	mediaType, imageBytes, decodeErr := decodeShuihuoGeneratedImage(ctx, shuihuoImageDownloadHTTPClient(), result.Data[0].B64JSON, result.Data[0].URL, apiKey, endpoint)
	if decodeErr != nil {
		return nil, fmt.Errorf("%w: %v", batchfactoryv11.ErrUpstream, decodeErr)
	}
	artifactID, err := localexecutor.NewArtifactID()
	if err != nil {
		return nil, err
	}
	saved, err := files.SaveImage(artifactID, mediaType, bytes.NewReader(imageBytes))
	if err != nil {
		return nil, err
	}
	media := storedShuihuoMedia(batch.Books[0])
	createdAt := time.Now().UTC().Format(time.RFC3339)
	media = append(media, map[string]any{"id": artifactID, "kind": "image", "filename": artifactID + ".png", "mediaType": mediaType, "storageRef": saved.StorageRef, "segmentId": segmentID, "isPrimary": false, "createdAt": createdAt})
	metadata := cloneMetadata(batch.Books[0].SourceMetadata)
	metadata["shuihuoMedia"] = media
	if _, err := store.UpdateBookMetadata(ctx, owner, batch.ID, batch.Books[0].ID, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata, ExpectedRevision: batch.Books[0].Revision}); err != nil {
		_ = files.Remove(saved.StorageRef)
		return nil, err
	}
	return map[string]any{"id": artifactID, "status": "completed", "kind": "image", "segmentId": segmentID, "mediaUrl": "/api/shuihuo-production/media/" + artifactID + "/download", "createdAt": createdAt}, nil
}

func shuihuoImageGenerationEndpoint(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid image endpoint")
	}
	path := strings.TrimRight(u.Path, "/")
	if !strings.HasSuffix(strings.ToLower(path), "/images/generations") {
		if path == "" {
			path = "/v1/images/generations"
		} else if strings.HasSuffix(path, "/v1") {
			path += "/images/generations"
		} else {
			path += "/v1/images/generations"
		}
	}
	u.Path = path
	return u.String(), nil
}

func shuihuoImageSize(aspectRatio string) string {
	switch strings.TrimSpace(aspectRatio) {
	case "9:16":
		return "1024x1792"
	case "3:4":
		return "1536x2048"
	case "2:3":
		return "1024x1536"
	case "1:1":
		return "1024x1024"
	case "4:3":
		return "1024x768"
	case "3:2":
		return "1536x1024"
	default:
		return "1792x1024"
	}
}

func shuihuoImageSizeForModel(modelID, aspectRatio string) string {
	if !shuihuoIsGPTImageModel(modelID) {
		return shuihuoImageSize(aspectRatio)
	}
	// OpenAI-compatible gpt-image endpoints commonly support 1024x1024,
	// 1024x1536 and 1536x1024, rather than the older 1024x1792 sizes.
	switch strings.TrimSpace(aspectRatio) {
	case "9:16", "3:4", "2:3":
		return "1024x1536"
	case "16:9", "4:3", "3:2":
		return "1536x1024"
	default:
		return "1024x1024"
	}
}

func decodeShuihuoGeneratedImage(ctx context.Context, client *http.Client, b64, imageURL, credential, providerEndpoint string) (string, []byte, error) {
	encoded := strings.TrimSpace(b64)
	if encoded != "" {
		if strings.HasPrefix(strings.ToLower(encoded), "data:") {
			mediaType, payload, err := decodeImageDataURL(encoded)
			if err != nil {
				return "", nil, fmt.Errorf("图片通道返回的图片数据无效：%v", err)
			}
			return mediaType, payload, nil
		}
		payload, err := decodeShuihuoBase64(encoded)
		if err != nil || len(payload) == 0 {
			return "", nil, fmt.Errorf("图片通道返回的图片数据不是有效的 base64")
		}
		return detectedShuihuoImageMediaType(payload), payload, nil
	}

	imageURL = strings.TrimSpace(imageURL)
	if imageURL == "" {
		return "", nil, fmt.Errorf("图片通道没有返回图片数据")
	}
	if strings.HasPrefix(strings.ToLower(imageURL), "data:") {
		mediaType, payload, err := decodeImageDataURL(imageURL)
		if err != nil {
			return "", nil, fmt.Errorf("图片通道返回的图片数据无效：%v", err)
		}
		return mediaType, payload, nil
	}

	parsedImageURL, err := url.Parse(imageURL)
	if err != nil || parsedImageURL.Scheme == "" || parsedImageURL.Host == "" {
		return "", nil, fmt.Errorf("图片通道返回的图片地址无效")
	}
	allowCredential := sameShuihuoImageOrigin(imageURL, providerEndpoint) && strings.TrimSpace(credential) != ""
	mediaType, payload, statusCode, downloadErr := downloadShuihuoGeneratedImageWithRetry(ctx, client, imageURL, "", false)
	if downloadErr == nil {
		return mediaType, payload, nil
	}
	if allowCredential {
		if retryMediaType, retryPayload, retryStatus, retryErr := downloadShuihuoGeneratedImageWithRetry(ctx, client, imageURL, credential, true); retryErr == nil {
			log.Printf("shuihuo image URL required provider authorization host=%s", parsedImageURL.Host)
			return retryMediaType, retryPayload, nil
		} else if retryStatus != 0 {
			statusCode = retryStatus
			downloadErr = retryErr
		}
	}
	if downloadErr != nil && statusCode != 0 {
		return "", nil, fmt.Errorf("%w: %v", errShuihuoImageDownloadUnusable, downloadErr)
	}
	if downloadErr != nil {
		return "", nil, fmt.Errorf("%w: %v", errShuihuoImageDownloadUnusable, downloadErr)
	}
	return "", nil, fmt.Errorf("%w: 图片通道没有返回图片数据", errShuihuoImageDownloadUnusable)
}

func decodeShuihuoBase64(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	decoders := []*base64.Encoding{base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding}
	for _, decoder := range decoders {
		if payload, err := decoder.DecodeString(value); err == nil && len(payload) > 0 {
			return payload, nil
		}
	}
	return nil, errors.New("invalid base64")
}

func sameShuihuoImageOrigin(imageURL, providerEndpoint string) bool {
	imageParsed, imageErr := url.Parse(strings.TrimSpace(imageURL))
	providerParsed, providerErr := url.Parse(strings.TrimSpace(providerEndpoint))
	if imageErr != nil || providerErr != nil || imageParsed == nil || providerParsed == nil {
		return false
	}
	return strings.EqualFold(imageParsed.Scheme, providerParsed.Scheme) && strings.EqualFold(imageParsed.Host, providerParsed.Host)
}

func downloadShuihuoGeneratedImage(ctx context.Context, client *http.Client, imageURL, credential string, withCredential bool) (string, []byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", nil, 0, err
	}
	req.Header.Set("Accept", "image/*")
	if withCredential {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(credential))
	}
	response, err := client.Do(req)
	if err != nil {
		return "", nil, 0, err
	}
	defer response.Body.Close()
	payload, readErr := io.ReadAll(io.LimitReader(response.Body, 16<<20))
	if readErr != nil {
		return "", nil, response.StatusCode, readErr
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", nil, response.StatusCode, fmt.Errorf("图片下载失败（HTTP %d）", response.StatusCode)
	}
	if len(payload) == 0 {
		return "", nil, response.StatusCode, fmt.Errorf("图片下载返回空内容")
	}
	return detectedShuihuoImageMediaType(payload), payload, response.StatusCode, nil
}

// shuihuoImageDownloadSemaphore 单独给 CDN 下载限流：生成阶段已经受
// gpt-image 信号量约束，但下载大图同样会抢占同一出口带宽，多路同时下载
// 会把 CDN 挤到读不完 body。这里固定 2 路，避免下载互相拖垮。
var shuihuoImageDownloadSemaphore = newShuihuoImageSemaphoreRaw(2)

func downloadShuihuoGeneratedImageWithRetry(ctx context.Context, client *http.Client, imageURL, credential string, withCredential bool) (string, []byte, int, error) {
	// 三次：CDN 抖动经常连续两次；第二次就放弃会让已生成好的图被丢弃。
	const maxAttempts = 3
	parsed, _ := url.Parse(imageURL)
	host := "unknown"
	if parsed != nil && parsed.Host != "" {
		host = parsed.Host
	}

	releaseDownload, acquireErr := shuihuoImageDownloadSemaphore.acquire(ctx)
	if acquireErr != nil {
		return "", nil, 0, acquireErr
	}
	defer releaseDownload()

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		mediaType, payload, status, err := downloadShuihuoGeneratedImage(ctx, client, imageURL, credential, withCredential)
		if err == nil {
			return mediaType, payload, status, nil
		}

		log.Printf("shuihuo image download failed host=%s attempt=%d/%d status=%d withCredential=%t error=%v", host, attempt, maxAttempts, status, withCredential, err)
		if attempt == maxAttempts || (status >= 400 && status < 500) {
			return mediaType, payload, status, err
		}

		delay := time.Duration(attempt) * time.Second
		select {
		case <-ctx.Done():
			return mediaType, payload, status, ctx.Err()
		case <-time.After(delay):
		}
	}

	return "", nil, 0, nil
}

func detectedShuihuoImageMediaType(payload []byte) string {
	if len(payload) >= 8 && bytes.Equal(payload[:8], []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}) {
		return "image/png"
	}
	if len(payload) >= 3 && payload[0] == 0xff && payload[1] == 0xd8 && payload[2] == 0xff {
		return "image/jpeg"
	}
	if len(payload) >= 12 && string(payload[:4]) == "RIFF" && string(payload[8:12]) == "WEBP" {
		return "image/webp"
	}
	return "image/png"
}

func extractShuihuoDocxText(payload []byte) (string, error) {
	reader, err := openDocxDocumentXML(payload)
	if err != nil {
		if errors.Is(err, errDocxMissingDoc) {
			return "", fmt.Errorf("缺少 word/document.xml")
		}
		return "", err
	}
	defer reader.Close()
	decoder := xml.NewDecoder(reader)
	var paragraph strings.Builder
	var output strings.Builder
	for {
		token, tokenErr := decoder.Token()
		if tokenErr == io.EOF {
			break
		}
		if tokenErr != nil {
			return "", tokenErr
		}
		switch value := token.(type) {
		case xml.StartElement:
			switch value.Name.Local {
			case "t":
				var text string
				if decodeErr := decoder.DecodeElement(&text, &value); decodeErr != nil {
					return "", decodeErr
				}
				paragraph.WriteString(text)
			case "tab":
				paragraph.WriteByte('\t')
			case "br", "cr":
				paragraph.WriteByte('\n')
			}
		case xml.EndElement:
			if value.Name.Local == "p" && strings.TrimSpace(paragraph.String()) != "" {
				output.WriteString(strings.TrimSpace(paragraph.String()))
				output.WriteByte('\n')
				paragraph.Reset()
			}
		}
	}
	if strings.TrimSpace(paragraph.String()) != "" {
		output.WriteString(strings.TrimSpace(paragraph.String()))
	}
	return strings.TrimSpace(output.String()), nil
}

func storedShuihuoSegmentsFromBatch(batch batchfactoryv11.Batch) []map[string]any {
	if len(batch.Books) == 0 {
		return []map[string]any{}
	}
	return storedShuihuoSegments(batch.Books[0])
}

func parseShuihuoPromptCandidates(raw string) ([]map[string]any, error) {
	clean := strings.TrimSpace(raw)
	if strings.HasPrefix(clean, "```") {
		clean = strings.TrimSpace(strings.TrimPrefix(clean, "```json"))
		clean = strings.TrimSpace(strings.TrimPrefix(clean, "```"))
		clean = strings.TrimSpace(strings.TrimSuffix(clean, "```"))
	}
	start, end := strings.Index(clean, "{"), strings.LastIndex(clean, "}")
	if start < 0 || end <= start {
		return nil, fmt.Errorf("text model returned invalid prompt JSON")
	}
	var envelope struct {
		Candidates []map[string]any `json:"candidates"`
	}
	if err := json.Unmarshal([]byte(clean[start:end+1]), &envelope); err != nil {
		return nil, fmt.Errorf("prompt JSON parse failed: %w", err)
	}
	result := make([]map[string]any, 0, len(envelope.Candidates))
	for _, candidate := range envelope.Candidates {
		if modelString(candidate, "segmentId") == "" || modelString(candidate, "prompt") == "" {
			continue
		}
		result = append(result, map[string]any{"segmentId": modelString(candidate, "segmentId"), "prompt": modelString(candidate, "prompt")})
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("text model returned no usable prompts")
	}
	return result, nil
}

func paragraphCandidatesForText(bookID, text string) []map[string]any {
	text = strings.TrimSpace(text)
	if text == "" {
		return []map[string]any{}
	}
	lines := strings.FieldsFunc(text, func(r rune) bool { return r == '\n' || r == '\r' })
	if len(lines) == 1 {
		lines = splitShuihuoSentences(lines[0])
	}
	candidates := make([]map[string]any, 0, len(lines))
	for index, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		candidates = append(candidates, map[string]any{"id": fmt.Sprintf("candidate_%s_%d", bookID, index+1), "text": line, "sourceText": line, "subtitleText": line, "speaker": "旁白"})
	}
	return candidates
}

func fixedShuihuoCandidates(bookID, text string, linesPerSegment int) []map[string]any {
	if linesPerSegment < 1 {
		linesPerSegment = 1
	}
	lines := strings.FieldsFunc(strings.TrimSpace(text), func(r rune) bool { return r == '\n' || r == '\r' })
	if len(lines) == 0 && strings.TrimSpace(text) != "" {
		lines = []string{strings.TrimSpace(text)}
	}
	candidates := make([]map[string]any, 0, (len(lines)+linesPerSegment-1)/linesPerSegment)
	for start, index := 0, 1; start < len(lines); start, index = start+linesPerSegment, index+1 {
		end := start + linesPerSegment
		if end > len(lines) {
			end = len(lines)
		}
		line := strings.TrimSpace(strings.Join(lines[start:end], "\n"))
		if line == "" {
			continue
		}
		candidates = append(candidates, map[string]any{"id": fmt.Sprintf("candidate_%s_%d", bookID, index), "text": line, "sourceText": line, "subtitleText": line, "speaker": "旁白"})
	}
	return candidates
}

func parseSmartShuihuoCandidates(raw, bookID string) ([]map[string]any, error) {
	clean := strings.TrimSpace(raw)
	if strings.HasPrefix(clean, "```") {
		clean = strings.TrimSpace(strings.TrimPrefix(clean, "```json"))
		clean = strings.TrimSpace(strings.TrimPrefix(clean, "```"))
		clean = strings.TrimSpace(strings.TrimSuffix(clean, "```"))
	}
	start, end := strings.Index(clean, "{"), strings.LastIndex(clean, "}")
	if start < 0 || end <= start {
		return nil, fmt.Errorf("text model returned invalid JSON")
	}
	var envelope struct {
		Candidates []map[string]any `json:"candidates"`
		Segments   []map[string]any `json:"segments"`
	}
	if err := json.Unmarshal([]byte(clean[start:end+1]), &envelope); err != nil {
		return nil, fmt.Errorf("text model JSON parse failed: %w", err)
	}
	items := envelope.Candidates
	if len(items) == 0 {
		items = envelope.Segments
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("text model returned no candidates")
	}
	return normalizeShuihuoSegments(items, bookID), nil
}

func splitShuihuoSentences(text string) []string {
	var out []string
	start := 0
	for index, r := range text {
		if r == '。' || r == '！' || r == '？' || r == '!' || r == '?' || r == '；' || r == ';' {
			if line := strings.TrimSpace(text[start : index+len(string(r))]); line != "" {
				out = append(out, line)
			}
			start = index + len(string(r))
		}
	}
	if line := strings.TrimSpace(text[start:]); line != "" {
		out = append(out, line)
	}
	return out
}

func decodeShuihuoMediaDataURL(value string) (string, []byte, error) {
	parts := strings.SplitN(strings.TrimSpace(value), ",", 2)
	if len(parts) != 2 || !strings.HasPrefix(parts[0], "data:") || !strings.HasSuffix(parts[0], ";base64") {
		return "", nil, fmt.Errorf("invalid data url")
	}
	mediaType, _, err := mime.ParseMediaType(strings.TrimSuffix(strings.TrimPrefix(parts[0], "data:"), ";base64"))
	if err != nil {
		return "", nil, err
	}
	payload, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil || len(payload) == 0 {
		return "", nil, fmt.Errorf("invalid base64")
	}
	return strings.ToLower(mediaType), payload, nil
}

func normalizeShuihuoSegments(candidates []map[string]any, bookID string) []map[string]any {
	segments := make([]map[string]any, 0, len(candidates))
	for index, candidate := range candidates {
		text, _ := candidate["text"].(string)
		if strings.TrimSpace(text) == "" {
			text, _ = candidate["sourceText"].(string)
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		id, _ := candidate["id"].(string)
		if strings.TrimSpace(id) == "" {
			id = fmt.Sprintf("segment_%s_%d", bookID, index+1)
		}
		segment := map[string]any{"id": id, "orderIndex": index + 1, "sourceText": text, "subtitleText": text, "speaker": "旁白", "confirmed": true, "imagePrompt": "", "videoPrompt": ""}
		for _, key := range []string{"subtitleText", "speaker", "imagePrompt", "videoPrompt"} {
			if value, ok := candidate[key].(string); ok && strings.TrimSpace(value) != "" {
				segment[key] = strings.TrimSpace(value)
			}
		}
		segments = append(segments, segment)
	}
	return segments
}

func stringsToShuihuoCandidates(lines []string) []map[string]any {
	items := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		if text := strings.TrimSpace(line); text != "" {
			items = append(items, map[string]any{"text": text})
		}
	}
	return items
}

func writeShuihuoProject(w http.ResponseWriter, r *http.Request, store batchfactoryv11.Store, owner, id string) {
	batch, err := store.GetBatch(r.Context(), owner, id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, shuihuoProject(batch))
}
