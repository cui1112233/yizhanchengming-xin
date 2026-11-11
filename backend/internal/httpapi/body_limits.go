package httpapi

import (
	"errors"
	"fmt"
	"net/http"
)

// 请求体上限。数值以现有业务实际可上传体积为依据，宁宽勿误伤：
//   - 剧本解析：前端限 20MB，这里给 25MB 覆盖 multipart 边界开销；
//   - 插件文档：沿用既有 20MB；
//   - 视频上传（AI 分集导出、成片/合并上传）：现有链路可传 GB 级视频，
//     分别放宽到 5GB / 2GB；
//   - JSON 接口统一 16MB（批量工厂的宽表配置与章节文本远小于该值）。
const (
	maxJSONBodyBytes           = 16 << 20
	maxScreenplayUploadBytes   = 25 << 20
	maxPluginDocumentBodyBytes = 20 << 20
	maxAISegmentUploadBytes    = 5 << 30
	maxVideoUploadBodyBytes    = 2 << 30
	maxMultipartOverheadBytes  = 64 << 20
)

// limitRequestBody 在读取前限制请求体大小；Content-Length 已超限时直接 413。
// 返回 false 表示已写出响应，调用方应立即 return。
func limitRequestBody(w http.ResponseWriter, r *http.Request, limit int64) bool {
	if r == nil || r.Body == nil || limit <= 0 {
		return true
	}
	if r.ContentLength > limit {
		writePayloadTooLarge(w, limit)
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	return true
}

// isBodyTooLarge 判断错误是否由 MaxBytesReader 触发。
func isBodyTooLarge(err error) bool {
	var maxErr *http.MaxBytesError
	return errors.As(err, &maxErr)
}

// writePayloadTooLarge 统一返回 413 与中文提示。
func writePayloadTooLarge(w http.ResponseWriter, limit int64) {
	writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{
		"error": fmt.Sprintf("请求体超过上限（%s），请压缩或拆分后重试", formatByteLimit(limit)),
	})
}

func formatByteLimit(limit int64) string {
	const (
		kb = 1 << 10
		mb = 1 << 20
		gb = 1 << 30
	)
	switch {
	case limit >= gb:
		return fmt.Sprintf("%dGB", (limit+gb-1)/gb)
	case limit >= mb:
		return fmt.Sprintf("%dMB", (limit+mb-1)/mb)
	case limit >= kb:
		return fmt.Sprintf("%dKB", (limit+kb-1)/kb)
	default:
		return fmt.Sprintf("%dB", limit)
	}
}
