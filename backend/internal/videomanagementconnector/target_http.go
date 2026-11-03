package videomanagementconnector

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// 目标站的会话失效不会返回 401/403：zidingyi.php 会 302 跳转到 login.php，
// 而 http.Client 默认跟随跳转，最终仍是 HTTP 200 的登录页。因此会话状态
// 必须同时看“响应正文是不是登录页”，否则状态接口会把已过期会话当成已登录。
var (
	ErrSessionMissing     = errors.New("target site session missing")
	ErrSessionExpired     = errors.New("target site session expired")
	ErrSessionUnavailable = errors.New("target site session unavailable")
)

// VideoUploadService is the optional media step used by the batch publisher.
// Keeping it optional avoids changing the smaller connector test doubles while
// making the production TargetHTTPService the single owner of target uploads.
type VideoUploadService interface {
	UploadVideo(context.Context, Session, string, []byte) (string, error)
}

type StyleCatalogService interface {
	RefreshStyles(context.Context, Session) (string, error)
}

// BookReadbackService is the Go-owned equivalent of the legacy browser-worker
// "book_list" action. It is deliberately optional so existing connector
// fakes and integrations remain source-compatible while the migration is
// rolled out.
type BookReadbackService interface {
	GetBook(context.Context, Session, string, int, int) (any, error)
}

// TargetHTTPService is the Go-owned transport for the video-management site.
// It deliberately contains no browser or Node dependency. A CDP implementation
// can be added behind the same Service interface for sites that require a UI.
type TargetHTTPService struct {
	BaseURL string
	Client  *http.Client
}

func (s *TargetHTTPService) Login(ctx context.Context, input LoginInput) (Session, error) {
	base, err := s.base()
	if err != nil {
		return Session{}, err
	}
	client := s.client()
	landing, err := s.do(ctx, client, http.MethodGet, base+"/tttadmin/login.php", "", nil)
	if err != nil {
		return Session{}, fmt.Errorf("target site login page: %w", err)
	}
	initial := mergeCookies(landing.Header.Values("Set-Cookie"))
	landing.Body.Close()
	payload, _ := json.Marshal(input)
	response, err := s.do(ctx, client, http.MethodPost, base+"/tttadmin/api/login.php", initial, bytes.NewReader(payload))
	if err != nil {
		return Session{}, fmt.Errorf("target site login: %w", err)
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Session{}, fmt.Errorf("target site login returned HTTP %d", response.StatusCode)
	}
	var result map[string]any
	if json.Unmarshal(body, &result) != nil || result["success"] != true {
		return Session{}, fmt.Errorf("target site login rejected")
	}
	cookie := mergeCookies(append([]string{initial}, response.Header.Values("Set-Cookie")...))
	if cookie == "" {
		return Session{}, fmt.Errorf("target site did not return a session")
	}
	return Session{Cookie: cookie, TargetUsername: strings.TrimSpace(input.Username)}, nil
}

func (s *TargetHTTPService) RefreshCatalogs(ctx context.Context, session Session) (Catalogs, error) {
	base, err := s.base()
	if err != nil {
		return Catalogs{}, err
	}
	response, err := s.do(ctx, s.client(), http.MethodGet, base+"/tttadmin/api/organization.php", session.Cookie, nil)
	if err != nil {
		return Catalogs{}, fmt.Errorf("target site catalogs: %w", err)
	}
	defer response.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Catalogs{}, fmt.Errorf("target site catalogs returned HTTP %d", response.StatusCode)
	}
	var value any
	if json.Unmarshal(raw, &value) != nil {
		value = string(raw)
	}
	return Catalogs{Raw: value}, nil
}

// RefreshProfiles reads the video-management system's saved configuration
// profiles. It is intentionally separate from organization catalogs because
// the two endpoints have different payload contracts, while both remain
// owned by the same Go connector.
func (s *TargetHTTPService) RefreshProfiles(ctx context.Context, session Session) (any, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	response, err := s.do(ctx, s.client(), http.MethodGet, base+"/tttadmin/api/zdy_config.php?action=list", session.Cookie, nil)
	if err != nil {
		return nil, fmt.Errorf("target site profiles: %w", err)
	}
	defer response.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("target site profiles returned HTTP %d", response.StatusCode)
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, fmt.Errorf("target site profiles returned invalid JSON: %w", err)
	}
	return value, nil
}

func (s *TargetHTTPService) RefreshStyles(ctx context.Context, session Session) (string, error) {
	base, err := s.base()
	if err != nil {
		return "", err
	}
	response, err := s.do(ctx, s.client(), http.MethodGet, base+"/tttadmin/zidingyi.php", session.Cookie, nil)
	if err != nil {
		return "", fmt.Errorf("target site styles: %w", err)
	}
	defer response.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("target site styles returned HTTP %d", response.StatusCode)
	}
	return string(raw), nil
}

// GetBook reads one book's current target-site metadata. The target site uses
// the same endpoint for the novel upload readback and the old browser-worker
// book_list action; keeping it here makes Go the single owner of that request.
func (s *TargetHTTPService) GetBook(ctx context.Context, session Session, bookID string, page, pageSize int) (any, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	bookID = strings.TrimSpace(bookID)
	if bookID == "" || !numericBookID(bookID) {
		return nil, fmt.Errorf("target site book id is invalid")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	query := url.Values{
		"bookid":   []string{bookID},
		"page":     []string{fmt.Sprint(page)},
		"pageSize": []string{fmt.Sprint(pageSize)},
	}.Encode()
	response, err := s.do(ctx, s.client(), http.MethodGet, base+"/tttadmin/api/zbooklist_get.php?"+query, session.Cookie, nil)
	if err != nil {
		return nil, fmt.Errorf("target site book readback: %w", err)
	}
	defer response.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("target site book readback returned HTTP %d", response.StatusCode)
	}
	var value any
	if json.Unmarshal(raw, &value) != nil {
		value = string(raw)
	}
	return value, nil
}

// CheckSession 是会话状态的唯一判定入口：同时校验 HTTP 状态码和响应正文。
// 目标站在会话失效时会 302 到 login.php（最终 200），只判断状态码会把失效
// 会话误判为“已登录并验证”，与随后同步风格目录的“登录已失效”互相矛盾。
func (s *TargetHTTPService) CheckSession(ctx context.Context, session Session) error {
	if strings.TrimSpace(session.Cookie) == "" {
		return ErrSessionMissing
	}
	base, err := s.base()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSessionUnavailable, err)
	}
	response, err := s.do(ctx, s.client(), http.MethodGet, base+"/tttadmin/zidingyi.php", session.Cookie, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSessionUnavailable, err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return ErrSessionExpired
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("%w: HTTP %d", ErrSessionUnavailable, response.StatusCode)
	}
	body, _ := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if LooksLikeLoginPage(string(body)) {
		return ErrSessionExpired
	}
	return nil
}

// LooksLikeLoginPage 判断目标站响应正文是否为管理员登录页。会话失效时
// zidingyi.php 会跳转到 login.php，正文包含“管理员登录”和登录表单。
func LooksLikeLoginPage(page string) bool {
	text := strings.TrimSpace(page)
	if text == "" {
		return false
	}
	if strings.Contains(text, "管理员登录") {
		return true
	}
	lower := strings.ToLower(text)
	return strings.Contains(lower, `name="password"`) && strings.Contains(lower, `name="username"`)
}

func (s *TargetHTTPService) Submit(ctx context.Context, session Session, input SubmissionInput) (Submission, error) {
	base, err := s.base()
	if err != nil {
		return Submission{}, err
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range input.FormFields {
		if err := writer.WriteField(key, value); err != nil {
			return Submission{}, err
		}
	}
	if input.BookID != "" {
		_ = writer.WriteField("bookid", input.BookID)
	}
	filename := input.Filename
	if filename == "" {
		filename = input.BookID + ".txt"
	}
	// The target site accepts novel uploads under the same files[] field used
	// by the existing production contract. Keep this exact field name so the
	// Go connector and the former client cannot diverge.
	part, err := writer.CreateFormFile("files[]", filename)
	if err != nil {
		return Submission{}, err
	}
	if _, err := part.Write(input.Content); err != nil {
		return Submission{}, err
	}
	if err := writer.Close(); err != nil {
		return Submission{}, err
	}
	response, err := s.doWithContentType(ctx, s.client(), http.MethodPost, base+"/tttadmin/api/zbooklist_upload.php", session.Cookie, &body, writer.FormDataContentType())
	if err != nil {
		return Submission{}, fmt.Errorf("target site submit: %w", err)
	}
	defer response.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Submission{}, fmt.Errorf("target site submit returned HTTP %d", response.StatusCode)
	}
	var result map[string]any
	_ = json.Unmarshal(raw, &result)
	if result["success"] != true {
		return Submission{}, fmt.Errorf("target site rejected submission")
	}
	id := strings.TrimSpace(fmt.Sprint(result["id"]))
	if id == "<nil>" {
		id = ""
	}
	return Submission{ID: id, Status: "submitted", ReadbackRaw: result}, nil
}

func (s *TargetHTTPService) UploadVideo(ctx context.Context, session Session, filename string, data []byte) (string, error) {
	base, err := s.base()
	if err != nil {
		return "", err
	}
	digest := md5.Sum(data)
	payload, _ := json.Marshal(map[string]any{
		"files":  []map[string]string{{"name": filename, "md5": hex.EncodeToString(digest[:])}},
		"kind":   "ai_head_video",
		"source": "zbooklist_upload_ai_head",
	})
	response, err := s.doWithContentType(ctx, s.client(), http.MethodPost, base+"/tttadmin/api/music_put_url.php", session.Cookie, bytes.NewReader(payload), "application/json")
	if err != nil {
		return "", fmt.Errorf("target site video presign: %w", err)
	}
	defer response.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	var result map[string]any
	_ = json.Unmarshal(raw, &result)
	if response.StatusCode < 200 || response.StatusCode >= 300 || result["success"] != true {
		return "", fmt.Errorf("target site did not return video upload URL")
	}
	dataObject := mapOrEmpty(result["data"])
	files := arrayOfMaps(dataObject["files"])
	if len(files) == 0 {
		return "", fmt.Errorf("target site video upload URL is incomplete")
	}
	item := files[0]
	uploadURL := strings.TrimSpace(fmt.Sprint(item["upload_url"]))
	objectKey := strings.TrimSpace(fmt.Sprint(item["object_key"]))
	bucketURL := strings.TrimSpace(fmt.Sprint(item["bucket_url"]))
	if uploadURL == "" || objectKey == "" {
		return "", fmt.Errorf("target site video upload URL is incomplete")
	}
	put, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	for key, value := range mapString(item["headers"]) {
		put.Header.Set(key, value)
	}
	putResponse, err := s.client().Do(put)
	if err != nil {
		return "", fmt.Errorf("target site video upload: %w", err)
	}
	putResponse.Body.Close()
	if putResponse.StatusCode < 200 || putResponse.StatusCode >= 300 {
		return "", fmt.Errorf("target site video upload returned HTTP %d", putResponse.StatusCode)
	}
	return fmt.Sprintf(`{"name":%q,"original_name":%q,"url":%q,"object_key":%q}`, filename, filename, bucketURL, objectKey), nil
}

func (s *TargetHTTPService) GetSubmission(ctx context.Context, session Session, id string) (Submission, error) {
	base, err := s.base()
	if err != nil {
		return Submission{}, err
	}
	query := url.Values{
		"bookid":   []string{strings.TrimSpace(id)},
		"page":     []string{"1"},
		"pageSize": []string{"10"},
	}.Encode()
	response, err := s.do(ctx, s.client(), http.MethodGet, base+"/tttadmin/api/zbooklist_get.php?"+query, session.Cookie, nil)
	if err != nil {
		return Submission{}, fmt.Errorf("target site readback: %w", err)
	}
	defer response.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Submission{}, fmt.Errorf("target site readback returned HTTP %d", response.StatusCode)
	}
	var value any
	if json.Unmarshal(raw, &value) != nil {
		value = string(raw)
	}
	return Submission{ID: id, Status: "readback", ReadbackRaw: value}, nil
}

func (s *TargetHTTPService) base() (string, error) {
	raw := strings.TrimRight(strings.TrimSpace(s.BaseURL), "/")
	if raw == "" {
		return "", fmt.Errorf("target site base URL is required")
	}
	if _, err := url.ParseRequestURI(raw); err != nil {
		return "", fmt.Errorf("target site base URL: %w", err)
	}
	return raw, nil
}
func (s *TargetHTTPService) client() *http.Client {
	if s.Client != nil {
		return s.Client
	}
	proxyURL := strings.TrimSpace(os.Getenv("QIANTIE_VIDEO_CONNECTOR_PROXY"))
	if proxyURL == "" {
		return &http.Client{Timeout: 120 * time.Second}
	}
	proxy, err := url.Parse(proxyURL)
	if err != nil || proxy.Scheme == "" || proxy.Host == "" {
		return &http.Client{Timeout: 120 * time.Second}
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = http.ProxyURL(proxy)
	return &http.Client{Timeout: 120 * time.Second, Transport: transport}
}
func (s *TargetHTTPService) do(ctx context.Context, client *http.Client, method, endpoint, cookie string, body io.Reader) (*http.Response, error) {
	return s.doWithContentType(ctx, client, method, endpoint, cookie, body, "application/json")
}
func (s *TargetHTTPService) doWithContentType(ctx context.Context, client *http.Client, method, endpoint, cookie string, body io.Reader, contentType string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	if body != nil {
		req.Header.Set("Content-Type", contentType)
	}
	return client.Do(req)
}
func mergeCookies(values []string) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ";") {
			part = strings.TrimSpace(part)
			if strings.Contains(part, "=") {
				parts = append(parts, part)
				break
			}
		}
	}
	return strings.Join(parts, "; ")
}

func mapOrEmpty(value any) map[string]any {
	if result, ok := value.(map[string]any); ok {
		return result
	}
	return map[string]any{}
}

func arrayOfMaps(value any) []map[string]any {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if row, ok := item.(map[string]any); ok {
			out = append(out, row)
		}
	}
	return out
}

func mapString(value any) map[string]string {
	row := mapOrEmpty(value)
	out := make(map[string]string, len(row))
	for key, item := range row {
		out[key] = fmt.Sprint(item)
	}
	return out
}

func numericBookID(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(value) <= 20
}
