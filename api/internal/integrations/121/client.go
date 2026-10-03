package one21

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

const defaultFetchEndpoint = "https://txt.121w.com/api.php"

type Client struct {
	endpoint string
	http     *http.Client
}

func NewClient(endpoint string) *Client {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		endpoint = defaultFetchEndpoint
	}
	return &Client{
		endpoint: endpoint,
		http:     &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) FetchBook(ctx context.Context, bookID, platformID string, maxTxt int) (BookResponse, error) {
	bookID = strings.TrimSpace(bookID)
	platformID = strings.TrimSpace(platformID)
	if bookID == "" {
		return BookResponse{}, errors.New("121 book id is required")
	}
	if platformID == "" {
		return BookResponse{}, errors.New("121 platform id is required")
	}
	if maxTxt <= 0 {
		return BookResponse{}, errors.New("121 max_txt must be positive")
	}

	parsed, err := url.Parse(c.endpoint)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return BookResponse{}, errors.New("invalid 121 fetch endpoint")
	}
	query := parsed.Query()
	query.Set("bookid", bookID)
	query.Set("platform", platformID)
	query.Set("max_txt", strconv.Itoa(maxTxt))
	parsed.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return BookResponse{}, fmt.Errorf("build 121 request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return BookResponse{}, fmt.Errorf("request 121: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return BookResponse{}, fmt.Errorf("121 returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return BookResponse{}, fmt.Errorf("read 121 response: %w", err)
	}
	var result BookResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return BookResponse{}, fmt.Errorf("decode 121 response: %w", err)
	}
	if result.Code != http.StatusOK {
		message := strings.TrimSpace(result.Message)
		if message == "" {
			message = "upstream error"
		}
		return BookResponse{}, fmt.Errorf("121 error %d: %s", result.Code, message)
	}
	if strings.TrimSpace(result.Data) == "" {
		return BookResponse{}, errors.New("121 returned empty book content")
	}
	return result, nil
}
