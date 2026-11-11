package agentchat

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ToolResult is what a tool returns to the loop.
type ToolResult struct {
	Content   string
	Summary   string
	Artifacts []string
}

// ToolEnv is the per-call environment passed to a tool.
type ToolEnv struct {
	WorkspaceRoot string
	Skills        *SkillCatalog
}

// Tool is one entry of the built-in tool registry.
type Tool struct {
	Name        string
	Label       string
	Dangerous   bool
	Description string
	Schema      string
	Run         func(ctx context.Context, env ToolEnv, args map[string]any) (ToolResult, error)
}

// Safe tool names. auto mode executes only these; everything else is planned.
var safeToolNames = map[string]bool{
	"list_workspace_files":     true,
	"read_workspace_file":      true,
	"search_files":             true,
	"skill_view":               true,
	"skills_list":              true,
	"list_workspace_artifacts": true,
}

// IsSafeTool reports whether auto mode may execute the tool.
func IsSafeTool(name string) bool {
	return safeToolNames[name]
}

// ToolRegistry returns the built-in tool set in prompt order.
func ToolRegistry() []Tool {
	return []Tool{
		{
			Name:        "list_workspace_files",
			Label:       "列出工作区文件",
			Description: "列出工作区目录和文件。参数 path 为相对路径，省略时从工作区根目录开始。",
			Schema:      `{"path":"可选，相对路径"}`,
			Run:         runListWorkspaceFiles,
		},
		{
			Name:        "read_workspace_file",
			Label:       "读取文件",
			Description: "读取工作区内的文本文件。参数 path 为相对路径；maxChars 为可选最大字符数，默认 20000。",
			Schema:      `{"path":"相对路径","maxChars":20000}`,
			Run:         runReadWorkspaceFile,
		},
		{
			Name:        "search_files",
			Label:       "搜索文件",
			Description: "在工作区内按文件名或文本内容搜索。参数 query 必填，path 可选。",
			Schema:      `{"query":"关键词","path":"可选，相对路径"}`,
			Run:         runSearchFiles,
		},
		{
			Name:        "list_workspace_artifacts",
			Label:       "列出产物",
			Description: "列出工作区 output/ 目录下的产物文件。",
			Schema:      `{}`,
			Run:         runListArtifacts,
		},
		{
			Name:        "skills_list",
			Label:       "列出技能",
			Description: "列出当前可用技能（id、名称、描述）。",
			Schema:      `{}`,
			Run:         runSkillsList,
		},
		{
			Name:        "skill_view",
			Label:       "读取技能",
			Description: "读取技能正文。参数 skillId 必填；maxChars 默认 80000，offset 用于分页。",
			Schema:      `{"skillId":"技能 id","maxChars":80000,"offset":0}`,
			Run:         runSkillView,
		},
		{
			Name:        "run_skill",
			Label:       "执行技能",
			Dangerous:   true,
			Description: "执行声明了可执行命令的 CLI 类技能。参数 skillId 必填，args 为命令参数数组或字符串。",
			Schema:      `{"skillId":"技能 id","args":["参数"]}`,
			Run:         runRunSkill,
		},
	}
}

// FindTool resolves a tool by name.
func FindTool(name string) (Tool, bool) {
	for _, tool := range ToolRegistry() {
		if tool.Name == name {
			return tool, true
		}
	}
	return Tool{}, false
}

// ToolPrompt renders the tool list for the system prompt.
func ToolPrompt() string {
	var builder strings.Builder
	for _, tool := range ToolRegistry() {
		fmt.Fprintf(&builder, "- %s（%s）：%s 参数：%s\n", tool.Name, tool.Label, tool.Description, tool.Schema)
	}
	return builder.String()
}

// ErrWorkspaceRootOutsideUserDir marks a workspace root that resolves outside
// the server-owned per-user directory.
var ErrWorkspaceRootOutsideUserDir = errors.New("workspace root outside user directory")

// ResolveWorkspacePath turns a relative path into an absolute path inside the
// workspace root. It rewrites the original "context/..." convention to
// ".luoshui/context/..." and rejects any escape from the root.
func ResolveWorkspacePath(root, relative string) (string, error) {
	relative = strings.TrimSpace(strings.ReplaceAll(relative, "\\", "/"))
	relative = strings.TrimPrefix(relative, "/")
	if relative == "" || relative == "." {
		return root, nil
	}
	if strings.HasPrefix(relative, "context/") || relative == "context" {
		relative = ".luoshui/" + relative
	}
	cleaned := filepath.Clean(filepath.FromSlash(relative))
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) || filepath.IsAbs(cleaned) {
		return "", errors.New("路径非法")
	}
	absolute := filepath.Join(root, cleaned)
	relativeToRoot, err := filepath.Rel(root, absolute)
	if err != nil || relativeToRoot == ".." || strings.HasPrefix(relativeToRoot, ".."+string(filepath.Separator)) {
		return "", errors.New("路径非法")
	}
	return absolute, nil
}

// ResolveWorkspacePathSecure resolves a relative path and then verifies the
// symlink-resolved target still lives inside the symlink-resolved root.
func ResolveWorkspacePathSecure(root, relative string) (string, error) {
	target, err := ResolveWorkspacePath(root, relative)
	if err != nil {
		return "", err
	}
	inside, err := WithinDir(root, target)
	if err != nil || !inside {
		return "", errors.New("路径非法")
	}
	return target, nil
}

// ResolveWorkspaceRoot validates a client-supplied workspace root. The root
// must resolve (after symlink expansion) inside the server-owned per-user
// directory. An empty value is returned unchanged so the caller can derive the
// default workspace directory.
func ResolveWorkspaceRoot(userRoot, requested string) (string, error) {
	requested = strings.TrimSpace(requested)
	if requested == "" {
		return "", nil
	}
	userRootReal, err := filepath.EvalSymlinks(strings.TrimSpace(userRoot))
	if err != nil {
		return "", ErrWorkspaceRootOutsideUserDir
	}
	target, err := filepath.Abs(requested)
	if err != nil {
		return "", ErrWorkspaceRootOutsideUserDir
	}
	targetReal, err := filepath.EvalSymlinks(target)
	if err != nil {
		return "", ErrWorkspaceRootOutsideUserDir
	}
	info, err := os.Stat(targetReal)
	if err != nil || !info.IsDir() {
		return "", ErrWorkspaceRootOutsideUserDir
	}
	if !isPathWithin(userRootReal, targetReal) {
		return "", ErrWorkspaceRootOutsideUserDir
	}
	return targetReal, nil
}

// WithinDir reports whether target equals or is nested under root after both
// paths are resolved through symlinks. Both paths must exist.
func WithinDir(root, target string) (bool, error) {
	rootReal, err := filepath.EvalSymlinks(strings.TrimSpace(root))
	if err != nil {
		return false, err
	}
	targetReal, err := filepath.EvalSymlinks(strings.TrimSpace(target))
	if err != nil {
		return false, err
	}
	return isPathWithin(rootReal, targetReal), nil
}

// isPathWithin compares two absolute, symlink-resolved paths.
func isPathWithin(root, target string) bool {
	relative, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)))
}

func runListWorkspaceFiles(ctx context.Context, env ToolEnv, args map[string]any) (ToolResult, error) {
	target, err := ResolveWorkspacePathSecure(env.WorkspaceRoot, stringArg(args, "path"))
	if err != nil {
		return ToolResult{}, err
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		return ToolResult{}, errors.New("获取文件树失败")
	}
	names := make([]string, 0, len(entries))
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		names = append(names, name)
	}
	return ToolResult{Content: strings.Join(names, "\n"), Summary: fmt.Sprintf("%d 项", len(names))}, nil
}

func runReadWorkspaceFile(ctx context.Context, env ToolEnv, args map[string]any) (ToolResult, error) {
	relative := stringArg(args, "path")
	if relative == "" {
		return ToolResult{}, errors.New("文件不存在")
	}
	target, err := ResolveWorkspacePathSecure(env.WorkspaceRoot, relative)
	if err != nil {
		return ToolResult{}, err
	}
	raw, err := os.ReadFile(target)
	if err != nil {
		return ToolResult{}, errors.New("文件不存在")
	}
	if strings.ContainsRune(string(raw[:minInt(len(raw), 4096)]), 0) {
		return ToolResult{}, errors.New("读取文件失败")
	}
	content := string(raw)
	limit := intArg(args, "maxChars", 20000)
	truncated := false
	if len([]rune(content)) > limit {
		content = string([]rune(content)[:limit]) + "\n…（已截断）"
		truncated = true
	}
	summary := "已读取 " + relative
	if truncated {
		summary += "（已截断）"
	}
	return ToolResult{Content: content, Summary: summary}, nil
}

func runSearchFiles(ctx context.Context, env ToolEnv, args map[string]any) (ToolResult, error) {
	query := strings.TrimSpace(stringArg(args, "query"))
	if query == "" {
		return ToolResult{}, errors.New("请输入搜索关键词")
	}
	base, err := ResolveWorkspacePathSecure(env.WorkspaceRoot, stringArg(args, "path"))
	if err != nil {
		return ToolResult{}, err
	}
	needle := strings.ToLower(query)
	matches := []string{}
	visited := 0
	_ = filepath.WalkDir(base, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || entry == nil {
			return nil
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		visited++
		if visited > 2000 || len(matches) >= 50 {
			return filepath.SkipAll
		}
		if strings.Contains(strings.ToLower(entry.Name()), needle) {
			matches = append(matches, path)
			return nil
		}
		if entry.Type().IsRegular() && entryFileSize(entry) <= 512*1024 {
			if raw, readErr := os.ReadFile(path); readErr == nil && strings.Contains(strings.ToLower(string(raw)), needle) {
				matches = append(matches, path)
			}
		}
		return nil
	})
	if len(matches) == 0 {
		return ToolResult{Content: "未找到匹配文件。", Summary: "0 项"}, nil
	}
	return ToolResult{Content: strings.Join(matches, "\n"), Summary: fmt.Sprintf("%d 项", len(matches))}, nil
}

func runListArtifacts(ctx context.Context, env ToolEnv, args map[string]any) (ToolResult, error) {
	outputDir := filepath.Join(env.WorkspaceRoot, "output")
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return ToolResult{Content: "当前工作区还没有产物。", Summary: "0 项"}, nil
	}
	lines := make([]string, 0, len(entries))
	artifacts := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		lines = append(lines, "output/"+entry.Name())
		artifacts = append(artifacts, "output/"+entry.Name())
	}
	if len(lines) == 0 {
		return ToolResult{Content: "当前工作区还没有产物。", Summary: "0 项"}, nil
	}
	return ToolResult{Content: strings.Join(lines, "\n"), Summary: fmt.Sprintf("%d 项", len(lines)), Artifacts: artifacts}, nil
}

func runSkillsList(ctx context.Context, env ToolEnv, args map[string]any) (ToolResult, error) {
	skills := env.Skills.List()
	if len(skills) == 0 {
		return ToolResult{Content: "当前没有可用技能。", Summary: "0 项"}, nil
	}
	lines := make([]string, 0, len(skills))
	for _, skill := range skills {
		lines = append(lines, fmt.Sprintf("- %s：%s", skill.ID, skill.Description))
	}
	return ToolResult{Content: strings.Join(lines, "\n"), Summary: fmt.Sprintf("%d 项", len(skills))}, nil
}

func runSkillView(ctx context.Context, env ToolEnv, args map[string]any) (ToolResult, error) {
	skillID := stringArg(args, "skillId")
	if skillID == "" {
		skillID = stringArg(args, "skill_id")
	}
	skill, ok := env.Skills.Get(skillID)
	if !ok {
		return ToolResult{}, errors.New("指定的 Skill 未开启，请在技能页开启后重试。")
	}
	if strings.TrimSpace(skill.Body) == "" {
		return ToolResult{}, errors.New("Skill 正文为空")
	}
	body := []rune(skill.Body)
	offset := intArg(args, "offset", 0)
	if offset < 0 || offset > len(body) {
		offset = 0
	}
	limit := intArg(args, "maxChars", 80000)
	if limit <= 0 {
		limit = 80000
	}
	end := offset + limit
	if end > len(body) {
		end = len(body)
	}
	content := string(body[offset:end])
	if end < len(body) {
		content += fmt.Sprintf("\n…（已截断，offset=%d 继续读取）", end)
	}
	return ToolResult{Content: content, Summary: "已读取技能 " + skill.ID}, nil
}

func runRunSkill(ctx context.Context, env ToolEnv, args map[string]any) (ToolResult, error) {
	skillID := stringArg(args, "skillId")
	skill, ok := env.Skills.Get(skillID)
	if !ok {
		return ToolResult{}, errors.New("指定的 Skill 未开启，请在技能页开启后重试。")
	}
	if strings.TrimSpace(skill.Command) == "" {
		return ToolResult{}, errors.New("Skill 未配置可执行命令")
	}
	commandArgs := stringListArg(args, "args")
	runCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	command := exec.CommandContext(runCtx, skill.Command, commandArgs...)
	command.Dir = env.WorkspaceRoot
	output, err := command.CombinedOutput()
	result := ToolResult{Content: string(output), Summary: "执行 Skill：" + skill.ID}
	if err != nil {
		return result, fmt.Errorf("执行失败：%s", err.Error())
	}
	return result, nil
}

func stringArg(args map[string]any, key string) string {
	if args == nil {
		return ""
	}
	value, ok := args[key]
	if !ok {
		return ""
	}
	text, _ := value.(string)
	return strings.TrimSpace(text)
}

func intArg(args map[string]any, key string, fallback int) int {
	if args == nil {
		return fallback
	}
	switch value := args[key].(type) {
	case float64:
		return int(value)
	case int:
		return value
	}
	return fallback
}

func stringListArg(args map[string]any, key string) []string {
	if args == nil {
		return nil
	}
	switch value := args[key].(type) {
	case []any:
		out := make([]string, 0, len(value))
		for _, item := range value {
			if text, ok := item.(string); ok {
				out = append(out, text)
			}
		}
		return out
	case string:
		if strings.TrimSpace(value) == "" {
			return nil
		}
		return strings.Fields(value)
	}
	return nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func entryFileSize(entry os.DirEntry) int64 {
	info, err := entry.Info()
	if err != nil {
		return 0
	}
	return info.Size()
}
