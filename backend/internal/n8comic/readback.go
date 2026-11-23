package n8comic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

type pendingReceipt struct {
	path  string
	cause error
}

func (e *pendingReceipt) Error() string {
	return fmt.Sprintf("N8 合集等待平台回执：%v", e.cause)
}
func (e *pendingReceipt) Unwrap() []error { return []error{ErrCarouselReceiptPending, e.cause} }

// ReceiptLookup carries evidence saved for this specific create attempt.
type ReceiptLookup struct {
	PriorCarouselIDs   []int64
	ReceiptPath        string
	ImageSignatures    []string
	ImageIDs           []int64
	PreviousCarouselID int64
}

func (c *Client) receiptPath(location string) string {
	if strings.TrimSpace(location) == "" {
		return ""
	}
	base, err := url.Parse(strings.TrimRight(c.baseURL, "/") + createPath)
	if err != nil {
		return ""
	}
	ref, err := url.Parse(location)
	if err != nil {
		return ""
	}
	u := base.ResolveReference(ref)
	if !strings.EqualFold(u.Scheme, base.Scheme) || !strings.EqualFold(u.Host, base.Host) || u.User != nil || u.Opaque != "" {
		return ""
	}
	path := u.RequestURI()
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return ""
	}
	return path
}

// Readback performs read-only queries. A provider that returns image IDs but
// omits their carousel ownership cannot confirm a carousel this way.
func (c *Client) Readback(ctx context.Context, in ReceiptLookup) (PublishResult, error) {
	if !c.Configured() {
		return PublishResult{}, errors.New("N8 发布回查服务未配置")
	}
	var locationErr error
	if path := c.receiptPath(in.ReceiptPath); path != "" {
		result, err := c.readReceiptLocation(ctx, path)
		locationErr = err
		if err == nil && result.CarouselID > 0 && result.CarouselID != in.PreviousCarouselID && !slices.Contains(in.PriorCarouselIDs, result.CarouselID) && len(result.ImageIDs) > 0 && (len(in.ImageIDs) == 0 || slices.Equal(in.ImageIDs, result.ImageIDs)) {
			return result, nil
		}
	}
	if len(in.ImageSignatures) == 0 || len(in.ImageSignatures) != len(in.ImageIDs) {
		return PublishResult{}, locationErr
	}
	var carouselID int64
	for i, sig := range in.ImageSignatures {
		if sig == "" || in.ImageIDs[i] <= 0 {
			return PublishResult{}, nil
		}
		var image struct {
			ImageID    int64 `json:"image_id"`
			CarouselID int64 `json:"carousel_id"`
		}
		if err := c.post(ctx, readPath, map[string]any{"signature": sig}, &image); err != nil {
			var apiErr *APIError
			if errors.As(err, &apiErr) && apiErr.Code == "IMAGE_NOT_READY" {
				return PublishResult{}, nil
			}
			return PublishResult{}, err
		}
		if image.ImageID != in.ImageIDs[i] || image.CarouselID <= 0 || image.CarouselID == in.PreviousCarouselID || slices.Contains(in.PriorCarouselIDs, image.CarouselID) || (i > 0 && image.CarouselID != carouselID) {
			return PublishResult{}, nil
		}
		carouselID = image.CarouselID
	}
	return PublishResult{CarouselID: carouselID, ImageIDs: slices.Clone(in.ImageIDs)}, nil
}

func (c *Client) readReceiptLocation(ctx context.Context, path string) (PublishResult, error) {
	origin, err := url.Parse(c.baseURL)
	if err != nil {
		return PublishResult{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, origin.Scheme+"://"+origin.Host+path, nil)
	if err != nil {
		return PublishResult{}, err
	}
	// A receipt can be signed by its URL. Never follow redirects off this resource.
	client := *c.httpClient
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	res, err := client.Do(req)
	if err != nil {
		return PublishResult{}, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusAccepted || res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusNoContent {
		return PublishResult{}, nil
	}
	if res.StatusCode != http.StatusOK {
		return PublishResult{}, fmt.Errorf("N8 回执查询 HTTP 状态码 %d", res.StatusCode)
	}
	var envelope struct {
		Code string `json:"code"`
		Data struct {
			CarouselID int64   `json:"carousel_id"`
			ImageIDs   []int64 `json:"image_ids"`
		} `json:"data"`
	}
	if err = json.NewDecoder(io.LimitReader(res.Body, 4<<20)).Decode(&envelope); err != nil {
		return PublishResult{}, err
	}
	if envelope.Code != "SUCCESS" {
		return PublishResult{}, nil
	}
	return PublishResult{CarouselID: envelope.Data.CarouselID, ImageIDs: envelope.Data.ImageIDs}, nil
}
