package n8catalog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const DefaultBaseURL = "https://n8.hnqingyuwen.top/center-api/auth"

type Client struct {
	httpClient *http.Client
	baseURL    string
	adminToken string
}

type Work struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	CPWorkID string `json:"cpWorkId,omitempty"`
	CPType   string `json:"cpType,omitempty"`
}

type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AudioCategory struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type Audio struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	OrganizationID int64   `json:"organization_id"`
	CategoryIDs    []int64 `json:"category_ids"`
	Status         string  `json:"status"`
	UpdatedAt      string  `json:"updated_at"`
}

type ProducerAccount struct {
	Name            string
	OrganizationIDs []string
}

// defaultN8WorksPath 是 N8 作品目录的默认接口路径。该路径会随 N8 后台
// 版本变化，因此允许通过环境变量覆盖，避免改接口就必须重新发版。
const defaultN8WorksPath = "/works/select"

type WorksClient struct {
	httpClient *http.Client
	baseURL    string
	adminToken string
	worksPath  string
}

type DirectoryError struct {
	Kind   string
	Status int
}

func (e *DirectoryError) Error() string {
	if e.Status != 0 {
		return fmt.Sprintf("N8 directory %s (HTTP %d)", e.Kind, e.Status)
	}
	return "N8 directory " + e.Kind
}

func NewClient(httpClient *http.Client, baseURL, adminToken string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = DefaultBaseURL
	}
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	// N8 的管理员目录属于 auth 模块。兼容此前只配置到 center-api 的地址。
	if strings.HasSuffix(baseURL, "/center-api") {
		baseURL += "/auth"
	}
	return &Client{httpClient: httpClient, baseURL: baseURL, adminToken: strings.TrimSpace(adminToken)}
}

func NewWorksClient(httpClient *http.Client, baseURL, adminToken string) *WorksClient {
	return NewWorksClientWithPath(httpClient, baseURL, adminToken, "")
}

// NewWorksClientWithPath 允许覆盖作品查询路径。worksPath 为空时使用
// defaultN8WorksPath，便于 N8 调整接口后只改配置而不改代码。
func NewWorksClientWithPath(httpClient *http.Client, baseURL, adminToken, worksPath string) *WorksClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if strings.HasSuffix(baseURL, "/center-api") {
		baseURL = strings.TrimSuffix(baseURL, "/center-api") + "/union-api"
	}
	worksPath = strings.TrimSpace(worksPath)
	if worksPath == "" {
		worksPath = defaultN8WorksPath
	}
	if !strings.HasPrefix(worksPath, "/") {
		worksPath = "/" + worksPath
	}
	return &WorksClient{httpClient: httpClient, baseURL: baseURL, adminToken: strings.TrimSpace(adminToken), worksPath: worksPath}
}

func (c *WorksClient) Configured() bool {
	return c != nil && c.baseURL != "" && c.adminToken != ""
}

func (c *WorksClient) SearchWorks(ctx context.Context, query string) ([]Work, error) {
	if !c.Configured() {
		return nil, fmt.Errorf("N8 作品目录尚未配置")
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	queries := []string{query}
	// 本地书名和 N8 作品名可能存在标点或个别字差异。整句查不到时，
	// 再用标题中的有效片段查询，避免要求用户手动填写作品 ID。
	for _, part := range strings.FieldsFunc(query, func(r rune) bool {
		return strings.ContainsRune("，,。；;：:、|/\\", r)
	}) {
		part = strings.TrimSpace(part)
		if len([]rune(part)) >= 4 && !containsString(queries, part) {
			queries = append(queries, part)
		}
	}
	seen := make(map[int64]struct{})
	works := make([]Work, 0)
	for _, keyword := range queries {
		found, err := c.searchWorksOnce(ctx, keyword)
		if err != nil {
			return nil, err
		}
		for _, work := range found {
			if _, exists := seen[work.ID]; exists {
				continue
			}
			seen[work.ID] = struct{}{}
			works = append(works, work)
		}
		if len(works) > 0 {
			break
		}
	}
	return works, nil
}

func (c *WorksClient) searchWorksOnce(ctx context.Context, query string) ([]Work, error) {
	// N8 作品目录只支持 filtering 条件；传 keyword 会被忽略并返回整库数据，
	// 因此这里固定使用 name LIKE 精确到书名。
	payload := map[string]any{
		"page_size": 100,
		"filtering": []map[string]any{
			{"field": "name", "operator": "LIKE", "value": query},
		},
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+c.worksPath, bytes.NewReader(encoded))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("N8-Admin-Token", c.adminToken)
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, &DirectoryError{Kind: "connection"}
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, &DirectoryError{Kind: "http", Status: response.StatusCode}
	}
	var envelope struct {
		Code string `json:"code"`
		Data struct {
			List []struct {
				ID       int64  `json:"id"`
				Name     string `json:"name"`
				CPWorkID string `json:"cp_work_id"`
				CPType   string `json:"cp_type"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(&envelope); err != nil {
		return nil, &DirectoryError{Kind: "invalid_response"}
	}
	if envelope.Code != "SUCCESS" {
		return nil, &DirectoryError{Kind: "service_rejected"}
	}
	works := make([]Work, 0, len(envelope.Data.List))
	for _, item := range envelope.Data.List {
		name := strings.TrimSpace(item.Name)
		if item.ID <= 0 || name == "" {
			continue
		}
		works = append(works, Work{ID: item.ID, Name: name, CPWorkID: strings.TrimSpace(item.CPWorkID), CPType: strings.TrimSpace(item.CPType)})
	}
	return works, nil
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func (c *Client) Configured() bool {
	return c != nil && c.baseURL != "" && c.adminToken != ""
}

func (c *Client) ListProducerAccounts(ctx context.Context) ([]string, error) {
	entries, err := c.ListProducerAccountDirectory(ctx)
	if err != nil {
		return nil, err
	}
	accounts := make([]string, 0, len(entries))
	for _, entry := range entries {
		accounts = append(accounts, entry.Name)
	}
	return accounts, nil
}

func (c *Client) ListProducerAccountDirectory(ctx context.Context) ([]ProducerAccount, error) {
	if !c.Configured() {
		return nil, fmt.Errorf("N8 管理员目录尚未配置")
	}
	seen := make(map[string]struct{})
	accounts := make([]ProducerAccount, 0)
	const pageSize = 100
	for page := 1; page <= 100; page++ {
		payload := map[string]any{
			"page": page,
			// The publish API is authoritative for the producer's organization binding.
			// Do not filter by organization_info here: N8's account directory may omit
			// that field for the currently logged-in producer account.
			"page_size": pageSize,
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/admin_user/select", bytes.NewReader(encoded))
		if err != nil {
			return nil, err
		}
		request.Header.Set("Accept", "application/json")
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("N8-Admin-Token", c.adminToken)
		response, err := c.httpClient.Do(request)
		if err != nil {
			return nil, &DirectoryError{Kind: "connection"}
		}
		var envelope struct {
			Code string `json:"code"`
			Data struct {
				List []struct {
					Name             string          `json:"name"`
					Status           string          `json:"status"`
					OrganizationInfo json.RawMessage `json:"organization_info"`
				} `json:"list"`
			} `json:"data"`
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			response.Body.Close()
			return nil, &DirectoryError{Kind: "http", Status: response.StatusCode}
		}
		decodeErr := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(&envelope)
		response.Body.Close()
		if decodeErr != nil {
			return nil, &DirectoryError{Kind: "invalid_response"}
		}
		if envelope.Code != "SUCCESS" {
			return nil, &DirectoryError{Kind: "service_rejected"}
		}
		for _, item := range envelope.Data.List {
			name := strings.TrimSpace(item.Name)
			if name == "" || strings.EqualFold(strings.TrimSpace(item.Status), "DISABLE") {
				continue
			}
			if _, exists := seen[name]; exists {
				continue
			}
			seen[name] = struct{}{}
			accounts = append(accounts, ProducerAccount{Name: name, OrganizationIDs: organizationIDs(item.OrganizationInfo)})
		}
		if len(envelope.Data.List) < pageSize {
			break
		}
	}
	log.Printf("N8 producer directory returned %d enabled accounts", len(accounts))
	return accounts, nil
}

func organizationIDs(raw json.RawMessage) []string {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return nil
	}
	seen := map[string]struct{}{}
	result := make([]string, 0)
	var walk func(any)
	walk = func(node any) {
		switch item := node.(type) {
		case []any:
			for _, child := range item {
				walk(child)
			}
		case map[string]any:
			if id, ok := item["id"]; ok {
				value := strings.TrimSpace(fmt.Sprint(id))
				if value != "" && value != "<nil>" {
					if _, ok := seen[value]; !ok {
						seen[value] = struct{}{}
						result = append(result, value)
					}
				}
			}
			for _, child := range item {
				walk(child)
			}
		}
	}
	walk(value)
	return result
}

func (c *Client) ListOrganizations(ctx context.Context) ([]Organization, error) {
	if !c.Configured() {
		return nil, fmt.Errorf("N8 组织目录尚未配置")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/organization/tree", bytes.NewReader([]byte(`{}`)))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("N8-Admin-Token", c.adminToken)
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, &DirectoryError{Kind: "connection"}
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, &DirectoryError{Kind: "http", Status: response.StatusCode}
	}
	var envelope struct {
		Code string          `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(&envelope); err != nil {
		return nil, &DirectoryError{Kind: "invalid_response"}
	}
	if envelope.Code != "SUCCESS" {
		return nil, &DirectoryError{Kind: "service_rejected"}
	}
	var roots any
	if err := json.Unmarshal(envelope.Data, &roots); err != nil {
		return nil, &DirectoryError{Kind: "invalid_response"}
	}
	seen := make(map[string]struct{})
	result := make([]Organization, 0)
	var walk func(any)
	walk = func(value any) {
		switch node := value.(type) {
		case []any:
			for _, child := range node {
				walk(child)
			}
		case map[string]any:
			id := strings.TrimSpace(fmt.Sprint(node["id"]))
			name := strings.TrimSpace(fmt.Sprint(node["name"]))
			if id != "" && id != "<nil>" && name != "" && name != "<nil>" {
				if _, ok := seen[id]; !ok {
					seen[id] = struct{}{}
					result = append(result, Organization{ID: id, Name: name})
				}
			}
			for _, key := range []string{"children", "list", "items", "data"} {
				if child, ok := node[key]; ok {
					walk(child)
				}
			}
		}
	}
	walk(roots)
	return result, nil
}

func (c *Client) ListAudio(ctx context.Context, organizationID string) ([]Audio, error) {
	if !c.Configured() {
		return nil, fmt.Errorf("N8 音频目录尚未配置")
	}
	categoryBaseURL := strings.TrimSuffix(c.baseURL, "/auth")
	payload, _ := json.Marshal(map[string]any{"organization_id": strings.TrimSpace(organizationID), "status": "ENABLE", "page": 1, "page_size": 100})
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, categoryBaseURL+"/material/audio/select", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("N8-Admin-Token", c.adminToken)
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, &DirectoryError{Kind: "connection"}
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, &DirectoryError{Kind: "http", Status: response.StatusCode}
	}
	var envelope struct {
		Code string `json:"code"`
		Data struct {
			List []Audio `json:"list"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(&envelope); err != nil {
		return nil, &DirectoryError{Kind: "invalid_response"}
	}
	if envelope.Code != "SUCCESS" {
		return nil, &DirectoryError{Kind: "service_rejected"}
	}
	result := make([]Audio, 0, len(envelope.Data.List))
	for _, item := range envelope.Data.List {
		if item.ID > 0 && strings.TrimSpace(item.Name) != "" && strings.EqualFold(item.Status, "ENABLE") {
			item.Name = strings.TrimSpace(item.Name)
			result = append(result, item)
		}
	}
	return result, nil
}

func (c *Client) ListAudioCategories(ctx context.Context) ([]AudioCategory, error) {
	if !c.Configured() {
		return nil, fmt.Errorf("N8 音频分类目录尚未配置")
	}
	categoryBaseURL := strings.TrimSuffix(c.baseURL, "/auth")
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, categoryBaseURL+"/material/category/select", bytes.NewReader([]byte(`{"page":1,"page_size":100}`)))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("N8-Admin-Token", c.adminToken)
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, &DirectoryError{Kind: "connection"}
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, &DirectoryError{Kind: "http", Status: response.StatusCode}
	}
	var envelope struct {
		Code string `json:"code"`
		Data struct {
			List []AudioCategory `json:"list"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(&envelope); err != nil {
		return nil, &DirectoryError{Kind: "invalid_response"}
	}
	if envelope.Code != "SUCCESS" {
		return nil, &DirectoryError{Kind: "service_rejected"}
	}
	result := make([]AudioCategory, 0, len(envelope.Data.List))
	for _, item := range envelope.Data.List {
		if item.ID > 0 && strings.TrimSpace(item.Name) != "" && strings.EqualFold(strings.TrimSpace(item.Status), "ENABLE") {
			item.Name = strings.TrimSpace(item.Name)
			result = append(result, item)
		}
	}
	return result, nil
}
