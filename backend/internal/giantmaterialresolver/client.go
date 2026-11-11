package giantmaterialresolver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"qiantie/backend/internal/safehttp"
)

const DefaultMaterialURL = "https://n8.hnqingyuwen.top/center-api/material/video/select"

var materialIDPattern = regexp.MustCompile(`^[0-9]{10,25}$`)
var trailingBookCityPattern = regexp.MustCompile(`\s*[（(]\s*([^()（）]+?)\s*[）)]\s*$`)

type CodedError struct {
	Code string
}

func (e *CodedError) Error() string { return e.Code }

func CodeOf(err error) string {
	var coded *CodedError
	if errors.As(err, &coded) {
		return coded.Code
	}
	return ""
}

type Book struct {
	PlatformBookID string `json:"platformBookId"`
	PlatformName   string `json:"platformName,omitempty"`
	// PlatformCode is the provider's stable raw code (for example CD). Keep it
	// alongside the human-readable label so a later provider rename is traceable.
	PlatformCode string `json:"platformCode,omitempty"`
	// PlatformID is the target 视频管理系统 source-platform ID. It is supplied
	// only for provider codes with a verified mapping, never guessed at upload.
	PlatformID string `json:"platformId,omitempty"`
	Title      string `json:"title"`
}

type Material struct {
	MaterialID      string   `json:"materialId"`
	GiantMaterialID string   `json:"giantMaterialId"`
	Title           string   `json:"title,omitempty"`
	PlatformBookID  string   `json:"platformBookId,omitempty"`
	PlatformName    string   `json:"platformName,omitempty"`
	VideoURL        string   `json:"videoUrl"`
	Width           *float64 `json:"width,omitempty"`
	Height          *float64 `json:"height,omitempty"`
	DurationSeconds float64  `json:"durationSeconds"`
	MaterialTitle   string   `json:"materialTitle,omitempty"`
	Books           []Book   `json:"books,omitempty"`
}

type Client struct {
	HTTPClient *http.Client
	Endpoint   string
	Token      string
}

func NewClient(httpClient *http.Client, endpoint, token string) *Client {
	if httpClient == nil {
		httpClient = safehttp.NewClientWithTimeout(20 * time.Second)
	}
	if strings.TrimSpace(endpoint) == "" {
		endpoint = DefaultMaterialURL
	}
	return &Client{HTTPClient: httpClient, Endpoint: strings.TrimSpace(endpoint), Token: strings.TrimSpace(token)}
}

func (c *Client) Resolve(ctx context.Context, giantMaterialID string) (Material, error) {
	id := strings.TrimSpace(giantMaterialID)
	if !materialIDPattern.MatchString(id) {
		return Material{}, &CodedError{Code: "INVALID_GIANT_MATERIAL_ID"}
	}
	if strings.TrimSpace(c.Token) == "" {
		return Material{}, &CodedError{Code: "QINGYU_AUTH_NOT_CONFIGURED"}
	}
	endpoint, err := url.Parse(c.Endpoint)
	if err != nil || endpoint.Scheme != "https" || endpoint.Hostname() != "n8.hnqingyuwen.top" {
		return Material{}, &CodedError{Code: "QINGYU_ENDPOINT_INVALID"}
	}
	body, _ := json.Marshal(map[string]any{"ocean_material_ids": []string{id}})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, strings.NewReader(string(body)))
	if err != nil {
		return Material{}, &CodedError{Code: "QINGYU_REQUEST_FAILED"}
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("N8-Admin-Token", c.Token)
	response, err := c.HTTPClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return Material{}, &CodedError{Code: "QINGYU_TIMEOUT"}
		}
		return Material{}, &CodedError{Code: "QINGYU_UPSTREAM_ERROR"}
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return Material{}, &CodedError{Code: "QINGYU_AUTH_FAILED"}
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Material{}, &CodedError{Code: "QINGYU_UPSTREAM_FAILED"}
	}
	var payload map[string]any
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(&payload); err != nil {
		return Material{}, &CodedError{Code: "QINGYU_UPSTREAM_INVALID_JSON"}
	}
	if code, ok := payload["code"].(string); ok && code != "" && code != "SUCCESS" {
		return Material{}, &CodedError{Code: "QINGYU_UPSTREAM_FAILED"}
	}
	material := normalize(payload)
	if material.MaterialID == "" || material.VideoURL == "" || !allowedVideoURL(material.VideoURL) {
		return Material{}, &CodedError{Code: "QINGYU_MATERIAL_RESPONSE_INVALID"}
	}
	material.GiantMaterialID = id
	return material, nil
}

func allowedVideoURL(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	return err == nil && parsed.Scheme == "https" && strings.HasSuffix(parsed.Hostname(), "material.hnqingyuwen.top") && parsed.Port() == "" && parsed.User == nil
}

func normalize(payload map[string]any) Material {
	source := object(payload)
	if data, ok := payload["data"].(map[string]any); ok {
		source = data
		if list, ok := data["list"].([]any); ok && len(list) > 0 {
			if first, ok := list[0].(map[string]any); ok {
				source = first
			}
		}
	}
	books := make([]Book, 0)
	seen := map[string]bool{}
	if works, ok := source["works"].([]any); ok {
		for _, raw := range works {
			work, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			id := text(work["cp_work_id"])
			name := text(work["name"])
			platformCode := text(work["cp_type"])
			if id == "" || name == "" || seen[platformCode+":"+id] {
				continue
			}
			seen[platformCode+":"+id] = true
			name, taggedPlatformName := splitTrailingBookCityTag(name)
			platformName, platformID := normalizePlatform(platformCode)
			if taggedPlatformName != "" {
				platformName = taggedPlatformName
				platformID = platformIDForBookCity(taggedPlatformName)
			}
			books = append(books, Book{PlatformBookID: id, PlatformName: platformName, PlatformCode: platformCode, PlatformID: platformID, Title: name})
		}
	}
	material := Material{
		MaterialID:      firstText(source, "id", "material_id", "materialId"),
		Title:           firstText(source, "title", "book_title", "bookTitle", "name"),
		VideoURL:        firstText(source, "video_url", "videoUrl", "url", "play_url", "playUrl", "path"),
		DurationSeconds: number(source, "duration", "duration_seconds", "durationSeconds"),
		Width:           numberPtr(source, "width", "video_width", "videoWidth"),
		Height:          numberPtr(source, "height", "video_height", "videoHeight"),
		Books:           books,
	}
	if len(books) > 0 {
		titles, platforms := make([]string, 0, len(books)), make([]string, 0, len(books))
		platformSeen := map[string]bool{}
		for _, book := range books {
			titles = append(titles, book.Title)
			if book.PlatformName != "" && !platformSeen[book.PlatformName] {
				platforms = append(platforms, book.PlatformName)
				platformSeen[book.PlatformName] = true
			}
		}
		material.Title = strings.Join(titles, " / ")
		if len(books) == 1 {
			material.PlatformBookID = books[0].PlatformBookID
		}
		material.PlatformName = strings.Join(platforms, " / ")
	} else {
		material.PlatformBookID = firstText(source, "book_id", "bookId", "platform_book_id", "platformBookId")
		material.PlatformName = firstText(source, "platform_name", "platformName", "platform_label", "platformLabel")
	}
	if _, ok := source["works"].([]any); ok {
		material.MaterialTitle = text(source["name"])
	}
	return material
}

// normalizePlatform translates only codes whose display name and target
// platform ID are verified by Qingyu's own platform configuration. Unknown
// values remain visible but deliberately have no target ID, so callers cannot
// silently submit a book to the wrong source platform.
func normalizePlatform(code string) (name, platformID string) {
	switch strings.TrimSpace(code) {
	case "CD":
		// Qingyu: 常读; its own automatic-text configuration uses platform_id 2.
		return "常读", "2"
	case "QM":
		return "七猫", "3"
	case "ZH":
		return "知乎", "15"
	case "YG":
		return "点众付费", "4"
	case "HY":
		return "黑岩", "1"
	default:
		return strings.TrimSpace(code), ""
	}
}

// splitTrailingBookCityTag separates Qingyu's display convention
// "book title (book city)". The tag is source metadata, not part of the
// title that we later show as the novel name or send to 视频管理系统.
func splitTrailingBookCityTag(raw string) (title, bookCity string) {
	value := strings.TrimSpace(raw)
	match := trailingBookCityPattern.FindStringSubmatchIndex(value)
	if len(match) == 0 {
		return value, ""
	}
	title = strings.TrimSpace(value[:match[0]])
	bookCity = strings.TrimSpace(value[match[2]:match[3]])
	if title == "" || bookCity == "" {
		return value, ""
	}
	return title, bookCity
}

// platformIDForBookCity includes only values confirmed by Qingyu's own source
// platform configuration. Unknown labels are intentionally left unmapped so
// the caller cannot silently submit a work to the wrong 121 source platform.
func platformIDForBookCity(name string) string {
	switch strings.TrimSpace(name) {
	case "黑岩":
		return "1"
	case "常读":
		return "2"
	case "七猫":
		return "3"
	case "知乎":
		return "15"
	default:
		return ""
	}
}

func object(value any) map[string]any {
	if result, ok := value.(map[string]any); ok {
		return result
	}
	return map[string]any{}
}

func firstText(source map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := text(source[key]); value != "" {
			return value
		}
	}
	return ""
}

func text(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func number(source map[string]any, keys ...string) float64 {
	for _, key := range keys {
		switch value := source[key].(type) {
		case float64:
			if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
				continue
			}
			return value
		case json.Number:
			if parsed, err := value.Float64(); err == nil && parsed >= 0 {
				return parsed
			}
		}
	}
	return 0
}

func numberPtr(source map[string]any, keys ...string) *float64 {
	value := number(source, keys...)
	if value <= 0 {
		return nil
	}
	return &value
}
