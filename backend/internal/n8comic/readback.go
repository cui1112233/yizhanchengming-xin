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
	path, operation string
	cause           error
}

func (e *pendingReceipt) Error() string {
	return fmt.Sprintf("N8 合集等待平台回执：%v", e.cause)
}
func (e *pendingReceipt) Unwrap() []error { return []error{ErrCarouselReceiptPending, e.cause} }

// ReceiptLookup carries evidence saved for this specific create attempt.
type ReceiptLookup struct {
	IntentID, OperationID string
	PriorCarouselIDs      []int64
	ReceiptPath           string
	ImageSignatures       []string
	ImageIDs              []int64
	PreviousCarouselID    int64
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

// A carousel must be confirmed by its create operation. Identical images may
// be reused across unrelated carousels, even with the same page order.
type RejectedReceipt struct{ IntentID, Reason string }

func (e *RejectedReceipt) Error() string { return e.Reason }

func (c *Client) Readback(ctx context.Context, in ReceiptLookup) (PublishResult, error) {
	if !c.Configured() {
		return PublishResult{}, errors.New("N8 发布回查服务未配置")
	}
	path := c.receiptPath(in.ReceiptPath)
	if path == "" || in.IntentID == "" || len(in.ImageIDs) == 0 || slices.ContainsFunc(in.ImageIDs, func(id int64) bool { return id <= 0 }) {
		return PublishResult{}, nil
	}
	origin, err := url.Parse(c.baseURL)
	if err != nil {
		return PublishResult{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, origin.Scheme+"://"+origin.Host+path, nil)
	if err != nil {
		return PublishResult{}, err
	}
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
		Code    string `json:"code"`
		Message string `json:"message"`
		Data    struct {
			CarouselID  int64   `json:"carousel_id"`
			ImageIDs    []int64 `json:"image_ids"`
			OperationID string  `json:"operation_id"`
			RequestID   string  `json:"request_id"`
			Status      string  `json:"status"`
			Reason      string  `json:"reason"`
			Message     string  `json:"message"`
		} `json:"data"`
	}
	if err = json.NewDecoder(io.LimitReader(res.Body, 4<<20)).Decode(&envelope); err != nil {
		return PublishResult{}, err
	}
	row := envelope.Data
	if row.OperationID == "" {
		row.OperationID = res.Header.Get("X-Operation-ID")
	}
	if row.RequestID == "" {
		row.RequestID = res.Header.Get("X-Request-ID")
	}
	if row.RequestID != "" && row.RequestID != in.IntentID || row.OperationID != "" && in.OperationID != "" && row.OperationID != in.OperationID {
		return PublishResult{}, nil
	}
	correlated := row.RequestID == in.IntentID || in.OperationID != "" && row.OperationID == in.OperationID
	if !correlated || len(row.ImageIDs) > 0 && !slices.Equal(in.ImageIDs, row.ImageIDs) {
		return PublishResult{}, nil
	}
	state := strings.ToUpper(strings.TrimSpace(row.Status))
	if state == "" {
		state = strings.ToUpper(strings.TrimSpace(envelope.Code))
	}
	switch state {
	case "FAILED", "REJECTED", "CANCELLED", "CANCELED", "DENIED":
		reason := strings.TrimSpace(row.Reason)
		if reason == "" {
			reason = strings.TrimSpace(row.Message)
		}
		if reason == "" {
			reason = strings.TrimSpace(envelope.Message)
		}
		if reason == "" {
			reason = "发布平台已拒绝本次上传"
		}
		return PublishResult{}, &RejectedReceipt{IntentID: in.IntentID, Reason: reason}
	case "SUCCESS", "SUCCEEDED", "COMPLETED", "PUBLISHED", "DONE":
		if envelope.Code == "SUCCESS" && row.CarouselID > 0 && row.CarouselID != in.PreviousCarouselID && !slices.Contains(in.PriorCarouselIDs, row.CarouselID) && slices.Equal(in.ImageIDs, row.ImageIDs) {
			return PublishResult{ReceiptIntentID: in.IntentID, CarouselID: row.CarouselID, ImageIDs: row.ImageIDs}, nil
		}
	}
	return PublishResult{}, nil
}
