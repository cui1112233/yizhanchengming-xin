package novelfetchworkshop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const DefaultUpstreamEndpoint = "https://txt.121w.com/api.php"

type UpstreamFetcher struct {
	Client   *http.Client
	Endpoint string
}

type FetchResult struct {
	Code        int
	Message     string
	OriginalRaw string
	BookInfo    BookInfo
}

type FetchAndProcessResult struct {
	Code             int
	Message          string
	OriginalRaw      string
	ProcessedText    string
	BookInfo         BookInfo
	OriginalRawChars int
	ProcessedChars   int
	MaxTxt           int
}

type upstreamResponse struct {
	Code     int      `json:"code"`
	Message  string   `json:"msg"`
	Data     string   `json:"data"`
	BookInfo BookInfo `json:"bookinfo"`
}

func (f UpstreamFetcher) FetchAndProcess(ctx context.Context, bookID, platformID string, maxTxt int) (FetchAndProcessResult, error) {
	// maxTxt is a local processing limit. Sending it upstream can irreversibly
	// truncate the source before OriginalRaw is persisted.
	fetched, err := f.Fetch(ctx, bookID, platformID, 0)
	if err != nil {
		return FetchAndProcessResult{}, err
	}
	processed := ProcessRawText(fetched.OriginalRaw, maxTxt)
	return FetchAndProcessResult{
		Code:             fetched.Code,
		Message:          fetched.Message,
		OriginalRaw:      processed.OriginalRaw,
		ProcessedText:    processed.ProcessedText,
		BookInfo:         fetched.BookInfo,
		OriginalRawChars: processed.OriginalRawChars,
		ProcessedChars:   processed.ProcessedChars,
		MaxTxt:           processed.MaxTxt,
	}, nil
}

func (f UpstreamFetcher) Fetch(ctx context.Context, bookID, platformID string, maxTxt int) (FetchResult, error) {
	bookID = strings.TrimSpace(bookID)
	platformID = strings.TrimSpace(platformID)
	if bookID == "" || platformID == "" {
		return FetchResult{}, fmt.Errorf("121 upstream 参数不完整：缺少书号或书城 ID（bookid=%q platform=%q）", bookID, platformID)
	}
	endpoint := strings.TrimSpace(f.Endpoint)
	if endpoint == "" {
		endpoint = DefaultUpstreamEndpoint
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return FetchResult{}, fmt.Errorf("121 upstream endpoint invalid: %w", err)
	}
	query := parsed.Query()
	query.Set("bookid", bookID)
	query.Set("platform", platformID)
	if maxTxt > 0 {
		query.Set("max_txt", strconv.Itoa(maxTxt))
	} else {
		query.Del("max_txt")
	}
	parsed.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return FetchResult{}, fmt.Errorf("121 upstream request creation failed: %w", err)
	}
	client := f.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	response, err := client.Do(request)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return FetchResult{}, fmt.Errorf("121 upstream request timed out: %w", context.DeadlineExceeded)
		}
		return FetchResult{}, fmt.Errorf("121 upstream request failed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return FetchResult{}, fmt.Errorf("121 upstream HTTP status %d", response.StatusCode)
	}

	var payload upstreamResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, 16<<20)).Decode(&payload); err != nil {
		return FetchResult{}, fmt.Errorf("121 upstream returned invalid JSON: %w", err)
	}
	if payload.Code != http.StatusOK {
		return FetchResult{}, fmt.Errorf("121 upstream business error %d: %s（bookid=%s platform=%s）", payload.Code, strings.TrimSpace(payload.Message), bookID, platformID)
	}
	return FetchResult{Code: payload.Code, Message: payload.Message, OriginalRaw: payload.Data, BookInfo: payload.BookInfo}, nil
}
