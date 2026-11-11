package httpapi

import (
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	"qiantie/backend/internal/upscale"
)

// 与火山 AI MediaKit 本地上传限制保持一致：单文件最大 5 GB。
const upscaleMaxUploadBytes = int64(5 << 30)

// 浏览器上传的 multipart 片段常带 application/octet-stream，按扩展名补全真实类型。
var upscaleVideoTypes = map[string]string{
	".mp4":  "video/mp4",
	".m4v":  "video/x-m4v",
	".mov":  "video/quicktime",
	".mkv":  "video/x-matroska",
	".avi":  "video/x-msvideo",
	".flv":  "video/x-flv",
	".ts":   "video/mp2t",
	".wmv":  "video/x-ms-wmv",
	".webm": "video/webm",
	".mpg":  "video/mpeg",
	".mpeg": "video/mpeg",
}

func upscaleContentType(fileName, declared string) (string, bool) {
	declared = strings.ToLower(strings.TrimSpace(declared))
	if mimeType, ok := upscaleVideoTypes[strings.ToLower(filepath.Ext(strings.TrimSpace(fileName)))]; ok {
		if strings.HasPrefix(declared, "video/") {
			return declared, true
		}
		return mimeType, true
	}
	if strings.HasPrefix(declared, "video/") {
		return declared, true
	}
	return "", false
}

func RegisterUpscaleRoutes(mux *http.ServeMux, auth BridgeAuth, service *upscale.Service) {
	if mux == nil || service == nil {
		return
	}
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, string)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity, ok := BridgeIdentityFromContext(req.Context())
			if !ok || strings.TrimSpace(identity.Username) == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			fn(w, req, strings.TrimSpace(identity.Username))
		})))
	}

	handle("GET /api/ai-upscale/config", func(w http.ResponseWriter, _ *http.Request, _ string) {
		writeJSON(w, http.StatusOK, map[string]any{
			"configured": service.Configured(),
			"resolutions": []map[string]string{
				{"value": "1080p", "label": "1080p"},
				{"value": "2k", "label": "2K"},
			},
		})
	})

	handle("POST /api/ai-upscale/jobs", func(w http.ResponseWriter, req *http.Request, owner string) {
		if !service.Configured() {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": upscale.ErrNotConfigured.Error()})
			return
		}
		req.Body = http.MaxBytesReader(w, req.Body, upscaleMaxUploadBytes+(8<<20))
		if err := req.ParseMultipartForm(32 << 20); err != nil {
			if isBodyTooLarge(err) {
				writePayloadTooLarge(w, upscaleMaxUploadBytes+(8<<20))
				return
			}
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "读取上传视频失败：" + err.Error()})
			return
		}
		if req.MultipartForm != nil {
			defer req.MultipartForm.RemoveAll()
		}
		file, header, err := req.FormFile("file")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择要超分的视频文件"})
			return
		}
		defer file.Close()
		contentType, ok := upscaleContentType(header.Filename, header.Header.Get("Content-Type"))
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "只支持视频文件"})
			return
		}
		job, err := service.Submit(req.Context(), owner, file, header.Size, contentType, header.Filename, req.FormValue("resolution"))
		if err != nil {
			writeUpscaleError(w, err)
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"job": job})
	})

	handle("GET /api/ai-upscale/jobs/{id}", func(w http.ResponseWriter, req *http.Request, owner string) {
		job, err := service.Job(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			if errors.Is(err, upscale.ErrJobNotFound) {
				writeUpscaleError(w, err)
				return
			}
			// 上游临时查询失败时保留本地状态，前端可以继续轮询。
			if job.ID != "" {
				writeJSON(w, http.StatusOK, map[string]any{"job": job, "warning": err.Error()})
				return
			}
			writeUpscaleError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"job": job})
	})

	handle("POST /api/ai-upscale/jobs/{id}/cancel", func(w http.ResponseWriter, req *http.Request, owner string) {
		job, err := service.Cancel(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			if errors.Is(err, upscale.ErrJobNotFound) {
				writeUpscaleError(w, err)
				return
			}
			status := http.StatusBadGateway
			if job.ID != "" {
				// 火山拒绝取消（例如任务已经进入执行阶段）属于可解释的业务冲突。
				status = http.StatusConflict
			}
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"job": job})
	})
}

func writeUpscaleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, upscale.ErrNotConfigured):
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": upscale.ErrNotConfigured.Error()})
	case errors.Is(err, upscale.ErrInvalidResolution):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": upscale.ErrInvalidResolution.Error()})
	case errors.Is(err, upscale.ErrJobNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": upscale.ErrJobNotFound.Error()})
	default:
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
	}
}
