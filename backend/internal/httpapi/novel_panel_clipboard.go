package httpapi

import (
	"net/http"
)

func RegisterNovelPanelClipboardCompat(root *http.ServeMux, middleware BridgeAuth) {
	root.Handle("POST /api/novel-panel/native-clipboard/copy-rich", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "网页版不支持原生系统图文剪贴板；请改用浏览器的图文复制（系统已自动切换）。",
			"code":  "NATIVE_CLIPBOARD_UNAVAILABLE",
		})
	})))
}
