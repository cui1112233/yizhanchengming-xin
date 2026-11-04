// Package ffmpegbin 统一解析 ffmpeg/ffprobe 可执行文件位置。
//
// 查找顺序固定为：显式配置（QIANTIE_FFMPEG_PATH / QIANTIE_FFPROBE_PATH）
// → 项目自带目录（.tools/ffmpeg 或 <二进制目录>/tools/ffmpeg，可用
// QIANTIE_TOOLS_DIR 覆盖根目录）→ 系统 PATH。
// 找不到时不再返回裸 "exec: not found"，而是给出包含期望路径的明确文案。
package ffmpegbin

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	EnvFFmpegPath  = "QIANTIE_FFMPEG_PATH"
	EnvFFprobePath = "QIANTIE_FFPROBE_PATH"
	EnvToolsDir    = "QIANTIE_TOOLS_DIR"

	defaultFFmpeg  = "ffmpeg"
	defaultFFprobe = "ffprobe"
)

// MissingError 描述未找到可执行文件的情况，Error() 直接是面向用户的中文文案。
type MissingError struct {
	Binary string
	Tried  []string
}

func (e *MissingError) Error() string {
	label := "视频合成引擎（ffmpeg）"
	env := EnvFFmpegPath
	if e.Binary == defaultFFprobe {
		label = "视频探测引擎（ffprobe）"
		env = EnvFFprobePath
	}
	tried := "无"
	if len(e.Tried) > 0 {
		tried = joinLimited(e.Tried, 6)
	}
	return fmt.Sprintf("缺少%s：未找到可执行文件；期望路径：%s；已尝试：%s；请把 %s 放到 <tools>/%s/（本机 .tools\\%s，线上 /opt/www/qiantie/tools/%s）或设置 %s",
		label, expectedPathHint(e.Binary), tried, e.Binary, e.Binary, e.Binary, e.Binary, env)
}

// joinLimited 限制错误文案里列出的路径数量，避免一行日志过长。
func joinLimited(values []string, limit int) string {
	if len(values) <= limit {
		return strings.Join(values, ", ")
	}
	return fmt.Sprintf("%s 等 %d 项", strings.Join(values[:limit], ", "), len(values))
}

// IsMissing 判断错误是否属于“可执行文件不存在”。
func IsMissing(err error) bool {
	if err == nil {
		return false
	}
	var missing *MissingError
	if errors.As(err, &missing) {
		return true
	}
	if errors.Is(err, exec.ErrNotFound) || errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ENOENT) {
		return true
	}
	// 只有进程都没起来才判定为缺失；退出码错误说明引擎已启动。
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "executable file not found") ||
		strings.Contains(message, "cannot find the file") ||
		strings.Contains(message, "no such file or directory")
}

func expectedPathHint(binary string) string {
	name := exeName(binary)
	return fmt.Sprintf("<tools>/%s/%s 或 <tools>/%s/bin/%s", binary, name, binary, name)
}

func exeName(base string) string {
	if runtime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(base), ".exe") {
		return base + ".exe"
	}
	return base
}

func toolsRoots() []string {
	roots := make([]string, 0, 6)
	seen := map[string]bool{}
	add := func(root string) {
		root = strings.TrimSpace(root)
		if root == "" || seen[root] {
			return
		}
		seen[root] = true
		roots = append(roots, root)
	}
	if value := strings.TrimSpace(os.Getenv(EnvToolsDir)); value != "" {
		add(value)
	}
	if executable, err := os.Executable(); err == nil {
		addToolRoots(add, filepath.Dir(executable))
	}
	if wd, err := os.Getwd(); err == nil {
		addToolRoots(add, wd)
	}
	if runtime.GOOS != "windows" {
		add("/opt/www/qiantie/tools")
		add("/opt/qiantie/tools")
	}
	return roots
}

// addToolRoots 从给定目录向上最多三层寻找 .tools / tools，
// 兼容 backend/ 子目录与 .release/ 嵌套目录启动的情况。
func addToolRoots(add func(string), base string) {
	dir := base
	for depth := 0; depth < 4; depth++ {
		add(filepath.Join(dir, ".tools"))
		add(filepath.Join(dir, "tools"))
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
}

// bundledCandidates 返回项目自带目录下所有可能放置的位置。
func bundledCandidates(binary string) []string {
	name := exeName(binary)
	candidates := make([]string, 0, 16)
	for _, root := range toolsRoots() {
		candidates = append(candidates,
			filepath.Join(root, binary, name),
			filepath.Join(root, binary, "bin", name),
			filepath.Join(root, name),
			filepath.Join(root, "bin", name),
		)
	}
	return candidates
}

func existingFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// resolveCandidate 依次尝试显式路径、项目自带目录与 PATH。
func resolveCandidate(binary, explicitEnv string) (string, error) {
	tried := make([]string, 0, 16)
	if explicit := strings.TrimSpace(os.Getenv(explicitEnv)); explicit != "" {
		if existingFile(explicit) {
			return explicit, nil
		}
		if resolved, err := exec.LookPath(explicit); err == nil {
			return resolved, nil
		}
		tried = append(tried, explicit)
	}
	for _, candidate := range bundledCandidates(binary) {
		tried = append(tried, candidate)
		if existingFile(candidate) {
			return candidate, nil
		}
	}
	if resolved, err := exec.LookPath(binary); err == nil {
		return resolved, nil
	}
	tried = append(tried, binary+" (PATH)")
	return "", &MissingError{Binary: binary, Tried: tried}
}

// ResolveFFmpeg 返回可用的 ffmpeg 路径。
func ResolveFFmpeg() (string, error) {
	return resolveCandidate(defaultFFmpeg, EnvFFmpegPath)
}

// ResolveFFprobe 返回可用的 ffprobe 路径。
func ResolveFFprobe() (string, error) {
	return resolveCandidate(defaultFFprobe, EnvFFprobePath)
}

// ResolveProbe 优先返回 ffprobe；没有 ffprobe 时退回 ffmpeg，isFFmpeg 标记退回情况。
func ResolveProbe(explicit string) (path string, isFFmpeg bool, err error) {
	if explicit = strings.TrimSpace(explicit); explicit != "" {
		if existingFile(explicit) {
			return explicit, false, nil
		}
		if resolved, lookupErr := exec.LookPath(explicit); lookupErr == nil {
			return resolved, false, nil
		}
	}
	if resolved, probeErr := ResolveFFprobe(); probeErr == nil {
		return resolved, false, nil
	}
	if resolved, ffmpegErr := ResolveFFmpeg(); ffmpegErr == nil {
		return resolved, true, nil
	}
	return "", false, &MissingError{Binary: defaultFFprobe, Tried: append(bundledCandidates(defaultFFprobe), bundledCandidates(defaultFFmpeg)...)}
}

func isBareName(binary string) bool {
	base := strings.ToLower(filepath.Base(strings.TrimSpace(binary)))
	return base == "ffmpeg" || base == "ffmpeg.exe" || base == "ffprobe" || base == "ffprobe.exe" || strings.TrimSpace(binary) == ""
}

// Command 返回已解析路径的 exec.Cmd；binary 为空或裸名称时按 ffmpeg 统一解析。
func Command(ctx context.Context, binary string, args ...string) (*exec.Cmd, string, error) {
	requested := strings.TrimSpace(binary)
	if requested == "" {
		requested = defaultFFmpeg
	}
	path := requested
	if isBareName(requested) {
		kind := defaultFFmpeg
		if strings.HasPrefix(strings.ToLower(filepath.Base(requested)), "ffprobe") {
			kind = defaultFFprobe
		}
		resolved, err := resolveCandidate(kind, envForKind(kind))
		if err != nil {
			return nil, requested, err
		}
		path = resolved
	} else if !existingFile(path) {
		resolved, err := exec.LookPath(path)
		if err != nil {
			return nil, requested, &MissingError{Binary: requested, Tried: []string{requested}}
		}
		path = resolved
	}
	return exec.CommandContext(ctx, path, args...), path, nil
}

func envForKind(kind string) string {
	if kind == defaultFFprobe {
		return EnvFFprobePath
	}
	return EnvFFmpegPath
}

// Run 执行并返回合并输出错误；可执行文件缺失时返回明确的中文文案。
func Run(ctx context.Context, binary string, args ...string) error {
	command, path, err := Command(ctx, binary, args...)
	if err != nil {
		return err
	}
	output, err := command.CombinedOutput()
	if err == nil {
		return nil
	}
	if IsMissing(err) {
		return &MissingError{Binary: filepath.Base(path), Tried: []string{path}}
	}
	message := strings.TrimSpace(string(output))
	if len(message) > 512 {
		message = message[:512]
	}
	if message == "" {
		return fmt.Errorf("ffmpeg 执行失败（%s）：%w", path, err)
	}
	return fmt.Errorf("ffmpeg 执行失败（%s）：%s", path, message)
}

// Snapshot 描述当前媒体引擎状态，供启动日志与 runtime-build-info 复用。
type Snapshot struct {
	FFmpegPath     string `json:"ffmpeg_path"`
	FFmpegVersion  string `json:"ffmpeg_version"`
	FFprobePath    string `json:"ffprobe_path"`
	FFprobeVersion string `json:"ffprobe_version"`
	Available      bool   `json:"available"`
	Message        string `json:"message,omitempty"`
}

var (
	snapshotMu     sync.Mutex
	cachedSnapshot *Snapshot
)

// SnapshotNow 重新探测并缓存媒体引擎状态。探测超时 5 秒，不会阻塞启动。
func SnapshotNow() Snapshot {
	snapshot := Snapshot{}
	if path, err := ResolveFFmpeg(); err != nil {
		snapshot.Message = err.Error()
	} else {
		snapshot.FFmpegPath = path
		snapshot.FFmpegVersion = versionOf(path)
		snapshot.Available = true
	}
	if path, err := ResolveFFprobe(); err == nil {
		snapshot.FFprobePath = path
		snapshot.FFprobeVersion = versionOf(path)
	} else if snapshot.Message == "" {
		snapshot.Message = err.Error()
	}
	snapshotMu.Lock()
	cachedSnapshot = &snapshot
	snapshotMu.Unlock()
	return snapshot
}

// Current 返回缓存的媒体引擎状态；尚未探测时立即探测一次。
func Current() Snapshot {
	snapshotMu.Lock()
	cached := cachedSnapshot
	snapshotMu.Unlock()
	if cached != nil {
		return *cached
	}
	return SnapshotNow()
}

func versionOf(path string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, path, "-version").Output()
	if err != nil {
		return ""
	}
	line := strings.TrimSpace(strings.SplitN(string(output), "\n", 2)[0])
	if len(line) > 200 {
		line = line[:200]
	}
	return line
}

// LogStartup 在启动阶段输出一行明确的自检结论，缺失时给出放置说明。
func LogStartup() {
	snapshot := SnapshotNow()
	if snapshot.Available {
		log.Printf("媒体引擎自检：ffmpeg=%s（版本：%s）; ffprobe=%s（版本：%s）",
			snapshot.FFmpegPath, snapshot.FFmpegVersion, snapshot.FFprobePath, snapshot.FFprobeVersion)
		return
	}
	log.Printf("WARNING: %s；视频合成相关阶段会失败，请先放置 ffmpeg/ffprobe", snapshot.Message)
}
