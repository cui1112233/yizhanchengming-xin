package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

func RegisterWebSubmitSubmitCompat(root *http.ServeMux, middleware BridgeAuth, usersDir string, store novelfetchworkshop.Store) {
	if root == nil || store == nil || strings.TrimSpace(usersDir) == "" {
		return
	}
	submitHandler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var body struct {
			TaskIDs  []string `json:"taskIds"`
			Versions []string `json:"versions"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<10)).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取提交配置失败"})
			return
		}
		app := mapOrEmpty(settings["appConfig"])
		web := mapOrEmpty(app["web_submit"])
		if web["enabled"] != true {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "网站提交尚未启用；请先保存账号并勾选启用网站提交"})
			return
		}
		session := readUploadTargetSession(usersDir, identity.Username)
		cookie := strings.TrimSpace(toString(session["cookie"]))
		if cookie == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请先在网站提交页保存账号密码并登录目标站"})
			return
		}
		versions := cleanStrings(body.Versions)
		if len(versions) == 0 {
			versions = cleanStrings(anyStrings(web["submit_versions"]))
		}
		if len(versions) == 0 {
			versions = []string{"ai1"}
		}
		if len(body.TaskIDs) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "没有可提交的任务，请先勾选任务"})
			return
		}
		results := []any{}
		success := 0
		failed := 0
		for _, rawID := range body.TaskIDs {
			bookID := strings.TrimSpace(rawID)
			doc, docErr := store.GetDocument(req.Context(), identity.Username, bookID)
			if docErr != nil {
				results = append(results, map[string]any{"bookId": bookID, "status": "error", "error": "任务不存在"})
				failed++
				continue
			}
			meta := doc.Meta
			if meta == nil {
				meta = map[string]any{}
			}
			done := anyStrings(meta["siteSubmitDoneVersions"])
			if len(done) == 0 {
				done = anyStrings(meta["site_submit_done_versions"])
			}
			for _, version := range versions {
				row := map[string]any{"bookId": bookID, "version": version, "status": "error"}
				if containsVersion(done, version) {
					row["status"] = "skipped"
					row["error"] = "该版本已确认提交"
					results = append(results, row)
					continue
				}
				content, bodyErr := store.GetBody(req.Context(), identity.Username, bookID, version)
				if bodyErr != nil || strings.TrimSpace(content.Content) == "" {
					row["error"] = "未找到正文版本"
					results = append(results, row)
					failed++
					continue
				}
				platform, _ := strconv.Atoi(firstText(meta["platformId"], meta["platform_id"]))
				gender, style := submissionGenderStyle(doc)
				fields, fieldErr := buildUploadFieldsCompat(platform, gender, style, mapOrEmpty(web["advanced"]), map[string]any{})
				if fieldErr != nil {
					row["error"] = fieldErr.Error()
					results = append(results, row)
					failed++
					continue
				}
				multipartBody, contentType, buildErr := buildUploadMultipart(fields, bookID+".txt", []byte(content.Content))
				if buildErr != nil {
					row["error"] = "上传内容组装失败"
					results = append(results, row)
					failed++
					continue
				}
				uploadReq, reqErr := http.NewRequestWithContext(req.Context(), http.MethodPost, "http://two.121w.com/tttadmin/api/zbooklist_upload.php", bytes.NewReader(multipartBody))
				if reqErr != nil {
					row["error"] = "上传请求创建失败"
					results = append(results, row)
					failed++
					continue
				}
				uploadReq.Header.Set("Cookie", cookie)
				uploadReq.Header.Set("Content-Type", contentType)
				response, requestErr := (&http.Client{Timeout: 120 * time.Second}).Do(uploadReq)
				if requestErr != nil {
					row["error"] = "上传失败"
					results = append(results, row)
					failed++
					continue
				}
				responseBody, _ := io.ReadAll(io.LimitReader(response.Body, 2<<20))
				response.Body.Close()
				textBody := string(responseBody)
				if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden || strings.Contains(textBody, "管理员登录") {
					writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "notLoggedIn": true, "error": "目标站登录已失效，请重新登录"})
					return
				}
				var remote map[string]any
				_ = json.Unmarshal(responseBody, &remote)
				if response.StatusCode < 200 || response.StatusCode >= 300 || remote["success"] != true {
					row["error"] = firstText(remote["message"], remote["msg"], "上传失败")
					results = append(results, row)
					failed++
					continue
				}
				row["status"] = "submitted"
				row["error"] = nil
				results = append(results, row)
				success++
				meta["siteSubmitStatus"] = "submitted"
				done = append(done, version)
				meta["siteSubmitDoneVersions"] = uniqueStrings(done)
				meta["siteSubmitFailedVersions"] = removeString(anyStrings(meta["siteSubmitFailedVersions"]), version)
			}
			doc.Meta = meta
			_ = store.PutDocument(req.Context(), identity.Username, doc)
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": failed == 0, "results": results, "success_groups": success, "failed_groups": failed})
	}))
	// The legacy workbench calls /api/web-submit/submit, while the canonical
	// compatibility route lives under /api/batch-rewrite/web-submit/submit.
	// Keep both paths on the same handler so the Go 4001 page cannot 404.
	root.Handle("POST /api/batch-rewrite/web-submit/submit", submitHandler)
	root.Handle("POST /api/web-submit/submit", submitHandler)
}

func submissionGenderStyle(doc novelfetchworkshop.Document) (string, string) {
	meta := doc.Meta
	gender := strings.ReplaceAll(firstText(meta["gender"], meta["gender_name"], doc.Gender), "频", "")
	style := firstText(meta["style"], meta["styleName"], meta["style_name"])
	if _, ok := uploadStyleIDs[style]; !ok {
		style = "现代通用"
	}
	return strings.TrimSpace(gender), style
}

func uniqueStrings(values []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	return out
}
func removeString(values []string, target string) []string {
	out := []string{}
	for _, value := range values {
		if strings.TrimSpace(value) != target {
			out = append(out, value)
		}
	}
	return out
}
