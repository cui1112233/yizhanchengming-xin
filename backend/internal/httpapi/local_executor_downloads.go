package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const localExecutorReleaseVersion = "1.0.3"

// The native Go executor is now the only installer linked from the UI.  The
// older Electron installer remains addressable for rollback only and is not
// advertised by the manifest.
const goLocalExecutorReleaseVersion = "0.1.0"

var localExecutorUpdateFilePattern = regexp.MustCompile(`^yizhan-local-executor-v88-[0-9]+\.[0-9]+\.[0-9]+-win-x64\.exe$`)
var goLocalExecutorFilePattern = regexp.MustCompile(`^qiantie-executor-v88-[0-9]+\.[0-9]+\.[0-9]+-win-x64\.exe$`)

// RegisterLocalExecutorDownloadRoutes keeps the local executor download API
// available on Go 4001. The actual installer remains an operator-provided
// file under downloadsDir; Go never fabricates an executable when it is absent.
func RegisterLocalExecutorDownloadRoutes(root *http.ServeMux, downloadsDir string) {
	root.HandleFunc("GET /downloads/local-executor/manifest.json", func(w http.ResponseWriter, req *http.Request) {
		origin := "http://" + req.Host
		writeJSON(w, http.StatusOK, map[string]any{
			"version": goLocalExecutorReleaseVersion,
			"runtime": "go",
			"downloads": map[string]any{
				"mac":     nil,
				"windows": strings.TrimRight(origin, "/") + "/downloads/local-executor/qiantie-executor-v88-" + goLocalExecutorReleaseVersion + "-win-x64.exe",
			},
		})
	})
	root.HandleFunc("GET /downloads/local-executor/updates/{channel}/manifest.json", func(w http.ResponseWriter, req *http.Request) {
		channel := req.PathValue("channel")
		if !validLocalExecutorChannel(channel) {
			writeLocalExecutorDownloadError(w, http.StatusNotFound, "更新通道不存在")
			return
		}
		serveLocalExecutorFile(w, req, filepath.Join(downloadsDir, "local-executor-updates", channel, "manifest.json"), "更新清单正在发布，请稍后重试", "application/json; charset=utf-8", false)
	})
	root.HandleFunc("GET /downloads/local-executor/updates/{channel}/{file}", func(w http.ResponseWriter, req *http.Request) {
		channel, file := req.PathValue("channel"), req.PathValue("file")
		if !validLocalExecutorChannel(channel) || !validLocalExecutorUpdateFile(file) {
			writeLocalExecutorDownloadError(w, http.StatusNotFound, "更新文件不存在")
			return
		}
		serveLocalExecutorFile(w, req, filepath.Join(downloadsDir, "local-executor-updates", channel, file), "更新安装包正在发布，请稍后重试", "application/vnd.microsoft.portable-executable", true)
	})
	root.HandleFunc("GET /downloads/local-executor/{file}", func(w http.ResponseWriter, req *http.Request) {
		file := req.PathValue("file")
		canonical := "qiantie-executor-v88-" + goLocalExecutorReleaseVersion + "-win-x64.exe"
		legacy := "yizhan-local-executor-v88-" + localExecutorReleaseVersion + "-win-x64.exe"
		if file != canonical && file != legacy {
			writeLocalExecutorDownloadError(w, http.StatusNotFound, "下载文件不存在")
			return
		}
		serveLocalExecutorFile(w, req, filepath.Join(downloadsDir, file), "安装包正在发布，请稍后重试", "application/vnd.microsoft.portable-executable", false)
	})
}

func validLocalExecutorChannel(value string) bool {
	return value == "beta" || value == "stable"
}

func validLocalExecutorUpdateFile(value string) bool {
	return filepath.Base(value) == value && localExecutorUpdateFilePattern.MatchString(value)
}

func serveLocalExecutorFile(w http.ResponseWriter, req *http.Request, filePath, missingMessage, contentType string, immutable bool) {
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		writeLocalExecutorDownloadError(w, http.StatusServiceUnavailable, missingMessage)
		return
	}
	if immutable {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}
	w.Header().Set("Content-Type", contentType)
	http.ServeFile(w, req, filePath)
}

func writeLocalExecutorDownloadError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
