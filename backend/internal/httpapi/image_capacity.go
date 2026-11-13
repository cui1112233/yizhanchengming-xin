package httpapi

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"net/url"
	"strings"
	"time"

	_ "golang.org/x/image/webp"
	"qiantie/backend/internal/imagecapacity"
)

const imageCapacityQueueHint = "所有用户和书籍共享全站出图额度，名额不足时排队，图片完整取回后归还"

func validateCapacityImageBytes(data []byte) error {
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("上游返回的内容不是可识别的图片")
	}
	if config.Width < 1 || config.Height < 1 || int64(config.Width)*int64(config.Height) > 64000000 {
		return fmt.Errorf("上游图片尺寸无效或过大")
	}
	if _, _, err = image.Decode(bytes.NewReader(data)); err != nil {
		return fmt.Errorf("上游图片内容不完整或已损坏")
	}
	return nil
}
func validateCapacityImageURL(value string) error {
	if !imagecapacity.Enabled() {
		return nil
	}
	if strings.HasPrefix(strings.ToLower(value), "data:") {
		_, data, err := decodeImageDataURL(value)
		if err != nil {
			return err
		}
		return validateCapacityImageBytes(data)
	}
	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
		return fmt.Errorf("上游未返回有效图片地址")
	}
	return nil
}

func confirmImageRequestResult(ctx context.Context) { imagecapacity.ConfirmResult(ctx) }

func beginImageRequestOperation(ctx context.Context) (context.Context, func(error)) {
	if identity, ok := BridgeIdentityFromContext(ctx); ok {
		ctx = imagecapacity.WithIdentity(ctx, identity.Username, "")
	}
	ctx = imagecapacity.WithWaitHeartbeat(ctx, func() {
		reportComicGenerationHeartbeat(ctx)
		reportShuihuoImageEvent(ctx, "comic_images", "正在等待全站共享出图名额，尚未向上游提交")
	})
	return imagecapacity.BeginOperation(ctx)
}
func imageRequestIdentity(ctx context.Context, owner, book string) context.Context {
	return imagecapacity.WithIdentity(ctx, owner, book)
}
func imageRequestWorkerCount(legacy int) int {
	if imagecapacity.Enabled() {
		return 20
	}
	return legacy
}
func acquireLegacyHostImageSlot(ctx context.Context, endpoint string, configured int) (func(), error) {
	if imagecapacity.Enabled() {
		return func() {}, nil
	}
	release, _, err := shuihuoAdaptiveImageLimiter.acquire(ctx, endpoint, configured)
	return release, err
}

func validateImageCapacityLimits(account, model, initial int) error {
	if account < 1 || account > 20 || model < 1 || model > 20 || initial < 1 || initial > model || initial > account {
		return fmt.Errorf("并发值须在 1–20 之间，初始并发不能超过账号或模型上限")
	}
	return nil
}

func imageCapacityCanManage(ctx context.Context, options RouterOptions, identity BridgeIdentity) bool {
	if options.AuthStore == nil {
		return identity.IsOwner
	}
	account, err := options.AuthStore.FindAccount(ctx, identity.Username)
	return err == nil && account.IsOwner
}

func resolveImageCapacityRequest(ctx context.Context, options RouterOptions, identity BridgeIdentity, modelID, ratio, size string, references bool) (imagecapacity.Request, error) {
	if strings.TrimSpace(modelID) == "" {
		return imagecapacity.Request{}, fmt.Errorf("请先选择图片模型")
	}
	owner := effectiveModelCatalogOwner(identity, options.MemberProfileStore, options.AuthStore)
	config, err := loadShuihuoOwnerConfig(ctx, options.Store, owner, options.NovelFetchStore)
	if err != nil {
		return imagecapacity.Request{}, err
	}
	target, err := resolveShuihuoImageTargetWithChannels(ctx, options.UserConfigStore, owner, modelID, modelID, ratio, "personal_api", shuihuoCollection(config, "modelCatalog", nil))
	if err != nil {
		return imagecapacity.Request{}, err
	}
	model := target.CatalogModel
	endpoint := firstText(model["endpoint"], model["baseUrl"])
	if endpoint == "" {
		return imagecapacity.Request{}, fmt.Errorf("此模型由本地执行器管理并发，不使用图片 API 共享池")
	}
	if size == "" {
		if ratio == "" {
			ratio = "9:16"
		}
		size = shuihuoImageSizeForModel(target.ModelID, ratio)
	}
	credential := shuihuoModelCredential(model)
	if credential == "" {
		return imagecapacity.Request{}, fmt.Errorf("当前图片通道未配置密钥")
	}
	params, _ := model["params"].(map[string]any)
	return imagecapacity.Request{Endpoint: endpoint, Credential: credential, Model: target.ModelID, Size: size, Quality: modelString(params, "quality"), Resolution: modelString(params, "resolution"), References: references, Owner: identity.Username}, nil
}

func registerImageCapacityRoutes(root *http.ServeMux, auth BridgeAuth, options RouterOptions) {
	get := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := BridgeIdentityFromContext(r.Context())
		if !ok {
			writeJSON(w, 401, map[string]string{"error": "未登录"})
			return
		}
		if options.ImageCapacity == nil {
			writeJSON(w, 503, map[string]string{"error": "全站图片并发控制尚未初始化"})
			return
		}
		target, err := resolveImageCapacityRequest(r.Context(), options, identity, r.URL.Query().Get("modelId"), r.URL.Query().Get("aspectRatio"), r.URL.Query().Get("size"), r.URL.Query().Get("references") == "true")
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		status, err := options.ImageCapacity.Snapshot(ctx, target)
		if err != nil {
			writeJSON(w, 503, map[string]string{"error": "共享队列或并发策略暂不可用，请稍后刷新"})
			return
		}
		writeImageCapacityStatus(w, status, imageCapacityCanManage(r.Context(), options, identity))
	})
	root.Handle("GET /api/config/image-capacity", auth.Middleware(get))
}

func writeImageCapacityStatus(w http.ResponseWriter, status imagecapacity.Status, canManage bool) {
	writeJSON(w, 200, map[string]any{"policy": status.Policy, "running": status.Running, "modelRunning": status.ModelRunning, "modelTotalRunning": status.ModelTotalRunning, "uncertain": status.Uncertain, "waiting": status.Waiting, "effectiveLimit": status.EffectiveLimit, "coolingUntil": status.CoolingUntil, "canManage": canManage, "queueHint": imageCapacityQueueHint})
}
