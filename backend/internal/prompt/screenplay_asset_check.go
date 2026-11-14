package prompt

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"strings"
)

// Systems match the original batch-check GPT client boundary: PARSE_V2,
// not COMPLETE_CHECK suffixes. User templates only replace fixture values.
// Evidence: 07-parity-execution/asset-check-original/captured-client-requests.
//
//go:embed assetcheck/*-system.txt assetcheck/*-user-template.txt
var screenplayAssetCheckFiles embed.FS

const ScreenplayAssetCheckProtocolVersion = "original-episode-asset-supplement-v1"

type ScreenplayAssetCheckPrompt struct {
	SystemPrompt    string
	UserPrompt      string
	SystemDigest    string
	UserDigest      string
	OutputKey       string
	ProtocolVersion string
}

// BuildScreenplayAssetCheckPrompt builds one of the three requests per episode.
// Existing collections contain names only, as in the original supplement overlay.
// Content is kept intact; the caller owns the saved source snapshot.
func BuildScreenplayAssetCheckPrompt(assetType, promptType, content string, existingCharacters, existingScenes, existingProps []string) (ScreenplayAssetCheckPrompt, error) {
	key := map[string]string{"character": "characters", "scene": "scenes", "prop": "props"}[assetType]
	if key == "" {
		return ScreenplayAssetCheckPrompt{}, fmt.Errorf("不支持的资产检查类型：%s", assetType)
	}
	if promptType == "" {
		promptType = "general"
	}
	if promptType != "general" && promptType != "western" {
		return ScreenplayAssetCheckPrompt{}, fmt.Errorf("不支持的资产检查提示词类型：%s", promptType)
	}
	if strings.TrimSpace(content) == "" {
		return ScreenplayAssetCheckPrompt{}, fmt.Errorf("资产检查正文不能为空")
	}
	stem := "assetcheck/" + promptType + "-" + assetType
	system, err := screenplayAssetCheckFiles.ReadFile(stem + "-system.txt")
	if err != nil {
		return ScreenplayAssetCheckPrompt{}, err
	}
	userTemplate, err := screenplayAssetCheckFiles.ReadFile(stem + "-user-template.txt")
	if err != nil {
		return ScreenplayAssetCheckPrompt{}, err
	}
	// Single-pass replacement never interprets user text as another variable.
	user := strings.NewReplacer("{{CONTENT}}", content, "{{CHARACTERS}}", formatScreenplayAssetCheckNames(existingCharacters), "{{SCENES}}", formatScreenplayAssetCheckNames(existingScenes), "{{PROPS}}", formatScreenplayAssetCheckNames(existingProps)).Replace(string(userTemplate))
	return ScreenplayAssetCheckPrompt{SystemPrompt: string(system), UserPrompt: user, SystemDigest: fmt.Sprintf("%x", sha256.Sum256(system)), UserDigest: fmt.Sprintf("%x", sha256.Sum256([]byte(user))), OutputKey: key, ProtocolVersion: ScreenplayAssetCheckProtocolVersion}, nil
}

func cleanScreenplayAssetCheckName(value string) string {
	name := strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(value), "@"))
	switch strings.ToLower(name) {
	case "", "无", "未知", "暂无", "none", "null", "n/a", "-", "--":
		return ""
	}
	return name
}

func formatScreenplayAssetCheckNames(names []string) string {
	seen := map[string]bool{}
	out := make([]string, 0, len(names))
	for _, raw := range names {
		name := cleanScreenplayAssetCheckName(raw)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	if len(out) == 0 {
		return "（暂无）"
	}
	return strings.Join(out, "、")
}
