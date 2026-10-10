package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

var uploadBookIDPattern = regexp.MustCompile(`^(?:\d{1,20}|[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12})$`)

var uploadPlatformIDs = map[int]bool{1: true, 2: true, 3: true, 4: true, 6: true, 7: true, 15: true, 20: true, 26: true, 29: true, 31: true}
var uploadGenderIDs = map[string]int{"男": 1, "女": 2}
var uploadStyleIDs = map[string]int{"古风虐文": 101, "古风甜文": 102, "古风通用": 103, "年代虐文": 201, "年代甜文": 202, "年代通用": 203, "现代虐文": 301, "现代甜文": 302, "现代悬疑": 303, "现代通用": 304, "男频都市": 305, "现代女主": 306, "玄幻": 307, "历史": 308, "爆款BGM": 309, "家庭奇葩": 310, "家庭伤感": 311, "职场打脸": 312}

func RegisterNovelFetchUploadBatchCompat(root *http.ServeMux, middleware BridgeAuth, usersDir string, workshop novelfetchworkshop.Store) {
	if root == nil || middleware.AuthStore == nil || strings.TrimSpace(usersDir) == "" {
		return
	}
	root.Handle("POST /api/novel-fetch-upload/upload-batch", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			PlatformID int              `json:"platformId"`
			Advanced   map[string]any   `json:"advanced"`
			Items      []map[string]any `json:"items"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 2<<20)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "请求内容无效"})
			return
		}
		if !uploadPlatformIDs[input.PlatformID] {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "无效的平台"})
			return
		}
		if len(input.Items) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "请选择要上传的书籍"})
			return
		}
		if len(input.Items) > 50 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "一次最多上传 50 本"})
			return
		}
		session := readUploadTargetSession(usersDir, identity.Username)
		cookie := strings.TrimSpace(toString(session["cookie"]))
		if cookie == "" {
			writeJSON(w, http.StatusOK, map[string]any{"ok": false, "notLoggedIn": true, "error": "请先登录目标站"})
			return
		}
		results := make([]map[string]any, 0, len(input.Items))
		for _, item := range input.Items {
			bookID := strings.TrimSpace(toString(item["bookId"]))
			result := map[string]any{"bookId": bookID, "status": "error", "error": nil}
			if !uploadBookIDPattern.MatchString(bookID) {
				result["error"] = "书籍 ID 格式不正确"
				results = append(results, result)
				continue
			}
			meta := readUploadMeta(usersDir, identity.Username, bookID)
			var content []byte
			if strings.EqualFold(strings.TrimSpace(toString(item["source"])), "workshop") {
				if workshop == nil {
					result["error"] = "小说工作台存储未启用"
					results = append(results, result)
					continue
				}
				version := strings.TrimSpace(toString(item["version"]))
				if version == "" {
					version = "edited"
				}
				body, bodyErr := workshop.GetBody(req.Context(), identity.Username, bookID, version)
				if bodyErr != nil || strings.TrimSpace(body.Content) == "" {
					result["error"] = "未找到该版本的正文"
					results = append(results, result)
					continue
				}
				content = []byte(body.Content)
				if document, documentErr := workshop.GetDocument(req.Context(), identity.Username, bookID); documentErr == nil && document.Meta != nil {
					meta = document.Meta
				}
			} else {
				textPath := filepath.Join(usersDir, identity.Username, "novel-fetch", bookID+".txt")
				content, _ = os.ReadFile(textPath)
				if len(content) == 0 {
					result["error"] = "未找到已保存的正文"
					results = append(results, result)
					continue
				}
			}
			gender := strings.TrimSpace(toString(item["gender"]))
			if gender == "" {
				gender = strings.TrimSpace(toString(meta["gender"]))
			}
			style := strings.TrimSpace(toString(item["style"]))
			if style == "" {
				style = strings.TrimSpace(toString(meta["style"]))
			}
			fields, fieldErr := buildUploadFieldsCompat(input.PlatformID, gender, style, input.Advanced, item)
			if fieldErr != nil {
				result["error"] = fieldErr.Error()
				results = append(results, result)
				continue
			}
			body, contentType, err := buildUploadMultipart(fields, bookID+".txt", content)
			if err != nil {
				result["error"] = "上传内容组装失败"
				results = append(results, result)
				continue
			}
			request, err := http.NewRequestWithContext(req.Context(), http.MethodPost, "http://two.121w.com/tttadmin/api/zbooklist_upload.php", bytes.NewReader(body))
			if err != nil {
				result["error"] = "上传请求创建失败"
				results = append(results, result)
				continue
			}
			request.Header.Set("Cookie", cookie)
			request.Header.Set("Content-Type", contentType)
			response, requestErr := (&http.Client{Timeout: 120 * time.Second}).Do(request)
			if requestErr != nil {
				result["error"] = "上传失败"
				results = append(results, result)
				continue
			}
			responseBody, _ := io.ReadAll(io.LimitReader(response.Body, 2<<20))
			response.Body.Close()
			if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden || strings.Contains(string(responseBody), "管理员登录") {
				writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "notLoggedIn": true, "error": "目标站登录已失效，请重新登录"})
				return
			}
			var remote map[string]any
			_ = json.Unmarshal(responseBody, &remote)
			if response.StatusCode >= 200 && response.StatusCode < 300 && remote["success"] == true {
				result["status"] = "ok"
				result["error"] = nil
			} else {
				message := strings.TrimSpace(fmt.Sprint(remote["message"]))
				if message == "" || message == "<nil>" {
					message = strings.TrimSpace(fmt.Sprint(remote["msg"]))
				}
				if message == "" || message == "<nil>" {
					message = "上传失败"
				}
				result["error"] = message
			}
			results = append(results, result)
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "results": results})
	})))
}

func readUploadTargetSession(usersDir, username string) map[string]any {
	path := filepath.Join(usersDir, username, "upload-target.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return map[string]any{}
	}
	var session map[string]any
	if json.Unmarshal(raw, &session) != nil {
		return map[string]any{}
	}
	return session
}

func readUploadMeta(usersDir, username, bookID string) map[string]any {
	path := filepath.Join(usersDir, username, "novel-fetch", bookID+".meta.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return map[string]any{}
	}
	var meta map[string]any
	if json.Unmarshal(raw, &meta) != nil {
		return map[string]any{}
	}
	return meta
}

func buildUploadFieldsCompat(platformID int, gender, style string, advanced, item map[string]any) (map[string]string, error) {
	genderID, ok := uploadGenderIDs[gender]
	if !ok {
		return nil, fmt.Errorf("请为该本选择性别和风格")
	}
	styleID, ok := uploadStyleIDs[style]
	if !ok {
		return nil, fmt.Errorf("请为该本选择性别和风格")
	}
	readNumber := func(source map[string]any, key string, fallback, min, max int) int {
		value, ok := source[key]
		if !ok {
			return fallback
		}
		n, err := strconv.Atoi(fmt.Sprint(value))
		if err != nil || n < min || n > max {
			return fallback
		}
		return n
	}
	advancedValue := func(key string, fallback, min, max int) int {
		if value, exists := item[key]; exists {
			return readNumber(map[string]any{key: value}, key, fallback, min, max)
		}
		return readNumber(advanced, key, fallback, min, max)
	}
	stringValue := func(source map[string]any, key, fallback string) string {
		value, ok := source[key]
		if !ok || value == nil || fmt.Sprint(value) == "<nil>" {
			return fallback
		}
		return fmt.Sprint(value)
	}
	fontStyles := []int{1}
	if value, ok := advanced["fontColorStyles"].([]any); ok && len(value) > 0 {
		fontStyles = make([]int, 0, 4)
		for _, raw := range value {
			if n, err := strconv.Atoi(fmt.Sprint(raw)); err == nil && n >= 0 {
				fontStyles = append(fontStyles, n)
			}
			if len(fontStyles) == 4 {
				break
			}
		}
		if len(fontStyles) == 0 {
			fontStyles = []int{1}
		}
	}
	fontStylesJSON, _ := json.Marshal(fontStyles)
	return map[string]string{
		"platform_id":       strconv.Itoa(platformID),
		"organization":      "1",
		"gender":            strconv.Itoa(genderID),
		"style":             strconv.Itoa(styleID),
		"tl5":               strconv.Itoa(advancedValue("tl5", 0, 0, 1)),
		"ziti":              strconv.Itoa(advancedValue("ziti", 1, 1, 6)),
		"zitidx":            strconv.Itoa(advancedValue("zitidx", 62, 30, 162)),
		"biaohong":          stringValue(advanced, "biaohong", ""),
		"biaohong_reuse":    stringValue(advanced, "biaohongReuse", ""),
		"keywords":          stringValue(advanced, "keywords", ""),
		"jieya_num":         strconv.Itoa(advancedValue("jieyaNum", 4, 0, 8)),
		"jieya_ai_head":     strconv.Itoa(advancedValue("jieyaAiHead", 0, 0, 3)),
		"jieya_speed":       stringValue(advanced, "jieyaSpeed", "1.7"),
		"jieya_pitch":       stringValue(advanced, "jieyaPitch", "0"),
		"font_color_styles": string(fontStylesJSON),
		"gunping_num":       strconv.Itoa(advancedValue("gunpingNum", 4, 0, 20)),
		"gunping_speed":     stringValue(advanced, "gunpingSpeed", "1"),
	}, nil
}

func buildUploadMultipart(fields map[string]string, filename string, content []byte) ([]byte, string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, "", err
		}
	}
	part, err := writer.CreateFormFile("files[]", filename)
	if err != nil {
		return nil, "", err
	}
	if _, err := part.Write(content); err != nil {
		return nil, "", err
	}
	if err := writer.Close(); err != nil {
		return nil, "", err
	}
	return body.Bytes(), writer.FormDataContentType(), nil
}
