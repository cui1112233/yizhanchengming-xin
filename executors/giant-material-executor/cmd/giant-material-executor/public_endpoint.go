package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const defaultPublicAPIURL = "http://127.0.0.1:4000"

// bakedPublicAPIURL 是正式发布包自带的控制服务地址（打包时通过
// -ldflags "-X main.bakedPublicAPIURL=..." 注入）。普通用户拿到执行器后
// 无需知道服务器地址，直接填配对码即可绑定。
var bakedPublicAPIURL = ""

// bakedUpdatePublicKey is an Ed25519 public key encoded as base64. The
// release build injects it with -ldflags; without it, self-update stays off.
var bakedUpdatePublicKey = ""

func normalizePublicAPIURL(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", errors.New("控制服务地址不能为空")
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("控制服务地址无效：%w", err)
	}
	if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return "", errors.New("控制服务地址必须是带 http:// 或 https:// 的地址")
	}
	if parsed.User != nil || parsed.Fragment != "" || parsed.RawQuery != "" {
		return "", errors.New("控制服务地址不能包含账号、查询参数或片段")
	}
	return strings.TrimRight(parsed.String(), "/"), nil
}

func publicAPIURLPath() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "YizhanShengming", "GiantMaterialExecutor", "public-api-url"), nil
}

func loadPublicAPIURL() string {
	// 优先级：环境变量 > 用户保存的地址 > 打包烧入的地址 > 本机调试兜底。
	candidates := []string{strings.TrimSpace(os.Getenv("GIANT_MATERIAL_PUBLIC_API_URL"))}
	if path, err := publicAPIURLPath(); err == nil {
		if data, readErr := os.ReadFile(path); readErr == nil {
			candidates = append(candidates, strings.TrimSpace(string(data)))
		}
	}
	candidates = append(candidates, strings.TrimSpace(bakedPublicAPIURL))
	for _, candidate := range candidates {
		if normalized, err := normalizePublicAPIURL(candidate); err == nil {
			return normalized
		}
	}
	return defaultPublicAPIURL
}

func savePublicAPIURL(value string) error {
	path, err := publicAPIURLPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(value+"\n"), 0o600)
}
