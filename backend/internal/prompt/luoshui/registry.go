// Package luoshui 保存漫剧剧本与分镜链路所需的逐字提示词语料。
//
// 语料来源：洛水 2026 v2.9.0 逆向取证规格 §3（离线解密/常量池抽取）。
// 本包只做「可枚举、可版本化、可校验」的集中管理，不包含任何洛水二进制、
// DLL 或运行时代码。提示词正文逐字保留，运行时不改写、不润色。
package luoshui

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

//go:embed prompts/manifest.json prompts/*.txt
var corpus embed.FS

// Entry 是一条提示词的版本登记信息。摘要与长度用于防止语料被无意篡改。
type Entry struct {
	ID                 string `json:"id"`
	Section            string `json:"section"`
	Name               string `json:"name"`
	Source             string `json:"source"`
	Length             int    `json:"length"`
	SHA256_16          string `json:"sha256_16"`
	ExpectedSHA256_16  string `json:"expected_sha256_16"`
	File               string `json:"file"`
	ConstantPoolPrefix string `json:"constant_pool_prefix"`
	Verified           bool   `json:"verified"`

	body string
}

// Body 返回提示词正文。常量池前缀（Nuitka 的 "u" 标记）不是提示词内容，
// 按规格 §3.24 注释在运行时剥离；其余字符逐字保留。
func (e Entry) Body() string {
	if e.ConstantPoolPrefix == "" {
		return e.body
	}
	return strings.TrimPrefix(e.body, e.ConstantPoolPrefix)
}

// RuntimeLength 是剥离常量池前缀后的字符数。
func (e Entry) RuntimeLength() int { return len([]rune(e.Body())) }

// HiddenFromUser 标记禁止向用户展示正文的提示词（规格 §1.1 / §1.6 明确要求）。
func (e Entry) HiddenFromUser() bool {
	switch e.ID {
	case "a1_novel_to_screenplay", "storyboard_xiaoluo_v7_15s":
		return true
	}
	return false
}

// Registry 是只读的提示词目录。
type Registry struct {
	entries map[string]Entry
	order   []string
}

var (
	defaultOnce sync.Once
	defaultReg  *Registry
	defaultErr  error
)

// Default 返回进程内共享的语料目录，并在首次加载时逐条校验 SHA256。
func Default() (*Registry, error) {
	defaultOnce.Do(func() {
		defaultReg, defaultErr = load()
	})
	return defaultReg, defaultErr
}

// MustDefault 仅供测试与初始化路径使用；语料损坏时应视为构建错误。
func MustDefault() *Registry {
	reg, err := Default()
	if err != nil {
		panic(err)
	}
	return reg
}

func load() (*Registry, error) {
	raw, err := corpus.ReadFile("prompts/manifest.json")
	if err != nil {
		return nil, fmt.Errorf("读取提示词清单失败: %w", err)
	}
	var entries []Entry
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, fmt.Errorf("解析提示词清单失败: %w", err)
	}
	reg := &Registry{entries: make(map[string]Entry, len(entries)), order: make([]string, 0, len(entries))}
	for _, entry := range entries {
		body, readErr := corpus.ReadFile("prompts/" + entry.File)
		if readErr != nil {
			return nil, fmt.Errorf("读取提示词 %s 失败: %w", entry.ID, readErr)
		}
		text := string(body)
		sum := sha256.Sum256([]byte(text))
		digest := hex.EncodeToString(sum[:])
		if entry.SHA256_16 != "" && !strings.HasPrefix(digest, entry.SHA256_16) {
			return nil, fmt.Errorf("提示词 %s 校验失败: got=%s expect=%s", entry.ID, digest[:16], entry.SHA256_16)
		}
		if entry.Length > 0 && len([]rune(text)) != entry.Length {
			return nil, fmt.Errorf("提示词 %s 长度不一致: got=%d expect=%d", entry.ID, len([]rune(text)), entry.Length)
		}
		entry.body = text
		if _, exists := reg.entries[entry.ID]; exists {
			return nil, fmt.Errorf("提示词 ID 重复: %s", entry.ID)
		}
		reg.entries[entry.ID] = entry
		reg.order = append(reg.order, entry.ID)
	}
	return reg, nil
}

// Get 返回一条提示词；不存在时 found=false。
func (r *Registry) Get(id string) (Entry, bool) {
	if r == nil {
		return Entry{}, false
	}
	entry, ok := r.entries[strings.TrimSpace(id)]
	return entry, ok
}

// List 按清单顺序返回全部提示词元信息（不含正文）。
func (r *Registry) List() []Entry {
	if r == nil {
		return nil
	}
	out := make([]Entry, 0, len(r.order))
	for _, id := range r.order {
		entry := r.entries[id]
		entry.body = ""
		entry.Verified = true
		out = append(out, entry)
	}
	return out
}

// IDs 返回全部提示词 ID，用于枚举与版本登记。
func (r *Registry) IDs() []string {
	if r == nil {
		return nil
	}
	return append([]string(nil), r.order...)
}

// Body 返回提示词正文。
func (r *Registry) Body(id string) (string, bool) {
	entry, ok := r.Get(id)
	if !ok {
		return "", false
	}
	return entry.Body(), true
}

// overrideResolver 由 prompt 包在装配管理后台预设词目录时注入。
// 它让运行时优先读取已发布的管理员覆盖正文，避免后台改了不生效。
var overrideResolver atomic.Pointer[func(string) (string, bool)]

// SetOverrideResolver 注册覆盖正文解析器；传 nil 时恢复只读内嵌语料。
func SetOverrideResolver(resolve func(string) (string, bool)) {
	if resolve == nil {
		overrideResolver.Store(nil)
		return
	}
	overrideResolver.Store(&resolve)
}

// RuntimeBody 返回运行时正文：已发布覆盖优先，读不到时回退内嵌原文。
// 校验只在加载内嵌语料时进行，覆盖版本按管理后台发布的内容原样使用。
func (r *Registry) RuntimeBody(id string) (string, bool) {
	if resolve := overrideResolver.Load(); resolve != nil {
		if body, ok := (*resolve)(strings.TrimSpace(id)); ok {
			return body, true
		}
	}
	return r.Body(id)
}

// Render 用 vars 替换 {NAME} 形式的占位符。未提供的占位符原样保留，
// 以便调用方与规格逐字对照时能看出缺口。
func Render(body string, vars map[string]string) string {
	if len(vars) == 0 || body == "" {
		return body
	}
	pairs := make([]string, 0, len(vars)*2)
	for key, value := range vars {
		pairs = append(pairs, "{"+key+"}", value)
	}
	return strings.NewReplacer(pairs...).Replace(body)
}

// RenderByID 是 Body + Render 的便捷组合。
func (r *Registry) RenderByID(id string, vars map[string]string) (string, bool) {
	body, ok := r.Body(id)
	if !ok {
		return "", false
	}
	return Render(body, vars), true
}
