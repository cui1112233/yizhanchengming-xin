package httpapi

import (
	"context"
	"encoding/base64"
	"fmt"
	"math"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/safehttp"
)

const (
	nativeScriptConfigIDMaxBytes = 512
	nativeScriptTaskIDMaxBytes   = 2048
	scriptVideoProviderArk       = "seedance_ark"
	scriptVideoProviderSora      = "sora_json"
)

func isNativeScriptVideoProvider(p string) bool {
	return p == scriptVideoProviderArk || p == scriptVideoProviderSora
}
func nativeScriptTaskPrefix(provider, key string) string {
	return "native:" + provider + ":" + base64.RawURLEncoding.EncodeToString([]byte(key)) + ":"
}
func nativeScriptTaskParts(task string) (prefix, provider, key string) {
	parts := strings.SplitN(task, ":", 4)
	if len(parts) != 4 || parts[0] != "native" || !isNativeScriptVideoProvider(parts[1]) || parts[3] == "" {
		return
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(raw) == 0 || len(raw) > nativeScriptConfigIDMaxBytes {
		return
	}
	provider, key = parts[1], string(raw)
	prefix = strings.Join(parts[:3], ":") + ":"
	return
}

var nativeAssetID = regexp.MustCompile(`^asset-[A-Za-z0-9-]{1,250}$`)

func validateNativeMedia(raw string) error {
	if strings.HasPrefix(raw, "asset://") {
		if !nativeAssetID.MatchString(strings.TrimPrefix(raw, "asset://")) {
			return fmt.Errorf("主体ID格式无效")
		}
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil || len(raw) > 8192 || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
		return fmt.Errorf("参考图片必须是有效的公网HTTPS地址")
	}
	host := strings.TrimRight(strings.ToLower(u.Hostname()), ".")
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return fmt.Errorf("参考图片不能使用本机地址")
	}
	if ip := net.ParseIP(strings.SplitN(host, "%", 2)[0]); ip != nil && (safehttp.IsBlockedIP(ip)) {
		return fmt.Errorf("参考图片不能使用内网地址")
	}
	return nil
}
func (s *scriptVideoService) ValidateInput(ctx context.Context, owner string, input scriptVideoLocalInput) error {
	_, _, err := s.resolveInput(ctx, owner, input)
	return err
}
func (s *scriptVideoService) resolveInput(ctx context.Context, owner string, input scriptVideoLocalInput) (scriptVideoModel, batchfactoryv11.VideoProviderConfig, error) {
	var cfg batchfactoryv11.VideoProviderConfig
	fail := func(message string) (scriptVideoModel, batchfactoryv11.VideoProviderConfig, error) {
		return scriptVideoModel{}, cfg, fmt.Errorf("%s", message)
	}
	if err := ctx.Err(); err != nil {
		return scriptVideoModel{}, cfg, err
	}
	if strings.TrimSpace(input.Prompt) == "" || len([]rune(input.Prompt)) > 12000 {
		return fail("分镜视频提示词不能为空或超过12000个字符")
	}
	if input.GenerationMode != "" && input.GenerationMode != "text_to_video" && input.GenerationMode != "image_to_video" {
		return fail("视频生成模式无效")
	}
	if input.GenerationMode == "text_to_video" && (len(input.ImageURLs) > 0 || len(input.ReferenceAssets) > 0) {
		return fail("文生视频不能携带普通图片")
	}
	for _, raw := range input.ImageURLs {
		if strings.HasPrefix(strings.TrimSpace(raw), "asset://") {
			return fail("主体URI不能作为普通图片")
		}
	}
	spec, known := scriptVideoModelSpec(strings.TrimSpace(input.ModelKey))
	if !known || spec.Provider != batchfactoryv11.VideoProviderDoubaoLocal {
		if s.resolver == nil {
			return fail("视频模型配置不可用")
		}
		var err error
		cfg, err = s.resolver(ctx, owner, strings.TrimSpace(input.ModelKey))
		if err != nil {
			return scriptVideoModel{}, cfg, err
		}
		if isNativeScriptVideoProvider(cfg.Provider) {
			spec = scriptVideoModel{Provider: cfg.Provider, AcceptsReferenceImages: true}
			known = true
		}
	}
	if !known {
		return fail("unsupported video provider")
	}
	if input.GenerationMode == "image_to_video" && len(input.ImageURLs) == 0 && len(input.SubjectReferenceURIs) == 0 && !(spec.Provider == batchfactoryv11.VideoProviderDoubaoLocal && len(input.ReferenceAssets) > 0) {
		return fail("图生视频需要参考图片")
	}
	switch input.SubjectStrategy {
	case "":
		if len(input.SubjectReferenceURIs) > 0 || len(input.SubjectReferenceIDs) > 0 {
			return fail("主体引用需要开启主体模式")
		}
	case "seedance_subject":
		if spec.Provider != scriptVideoProviderArk {
			return fail("主体模式仅支持seedance_ark协议")
		}
		if len(input.SubjectReferenceURIs) == 0 {
			return fail("主体模式缺少主体引用")
		}
		if len(input.ImageURLs) > 0 {
			return fail("主体模式不能混用普通参考图片")
		}
	case "sora_text":
		if spec.Provider != scriptVideoProviderSora || input.GenerationMode != "text_to_video" {
			return fail("自动替换主体存档仅支持sora_json文生视频")
		}
		if len(input.SubjectReferenceURIs) > 0 {
			return fail("Sora文字替换不能携带主体图片引用")
		}
	default:
		return fail("主体策略无效")
	}
	if isNativeScriptVideoProvider(spec.Provider) {
		if _, err := nativeScriptDuration(input.Duration); err != nil {
			return scriptVideoModel{}, cfg, err
		}
		a := nativeScriptVideoAdapter{cfg: cfg}
		if err := a.Validate(); err != nil {
			return scriptVideoModel{}, cfg, err
		}
		if strings.TrimSpace(cfg.ConfigID) == "" || len(cfg.ConfigID) > nativeScriptConfigIDMaxBytes {
			return fail("原生视频配置缺少有效配置条目ID")
		}
		refs := append(append([]string{}, input.SubjectReferenceURIs...), input.ImageURLs...)
		unique := map[string]bool{}
		for _, raw := range refs {
			if err := validateNativeMedia(raw); err != nil {
				return scriptVideoModel{}, cfg, err
			}
			unique[raw] = true
		}
		if len(unique) > 9 {
			return fail("参考图片数量不能超过9张")
		}
		for _, raw := range input.ImageURLs {
			if strings.HasPrefix(raw, "asset://") {
				return fail("主体URI不能作为普通图片")
			}
		}
	}
	if spec.Provider == batchfactoryv11.VideoProviderDoubaoLocal {
		if s.localExecutor == nil {
			return fail("本地执行器服务不可用")
		}
	} else if !isNativeScriptVideoProvider(spec.Provider) {
		adapter, err := newScriptVideoAdapter(cfg, spec.Provider)
		if err != nil {
			return scriptVideoModel{}, cfg, err
		}
		if validator, ok := adapter.(interface{ Validate() error }); ok {
			if err = validator.Validate(); err != nil {
				return scriptVideoModel{}, cfg, err
			}
		}
	}
	return spec, cfg, nil
}

func nativeScriptDuration(raw any) (int, error) {
	value := 10
	switch n := raw.(type) {
	case nil:
	case int:
		value = n
	case float64:
		if math.IsNaN(n) || math.IsInf(n, 0) || math.Trunc(n) != n {
			return 0, fmt.Errorf("视频时长必须是整数秒")
		}
		value = int(n)
	case string:
		var err error
		value, err = strconv.Atoi(strings.TrimSpace(n))
		if err != nil {
			return 0, fmt.Errorf("视频时长格式无效")
		}
	default:
		return 0, fmt.Errorf("视频时长格式无效")
	}
	if value < 1 || value > 15 {
		return 0, fmt.Errorf("视频时长需在1至15秒内")
	}
	return value, nil
}
