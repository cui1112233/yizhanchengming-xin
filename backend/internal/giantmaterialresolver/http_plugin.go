package giantmaterialresolver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"qiantie/backend/internal/safehttp"
)

// HTTPPluginClient talks to an independently deployed giant-material plugin.
// The plugin is an upstream capability only: it never receives database access
// and the Go application remains responsible for users and job state.
type HTTPPluginClient struct {
	HTTPClient *http.Client
	BaseURL    string
	Token      string
}

func NewHTTPPluginClient(httpClient *http.Client, baseURL, token string) *HTTPPluginClient {
	if httpClient == nil {
		httpClient = safehttp.NewClientWithTimeout(20 * time.Second)
	}
	return &HTTPPluginClient{HTTPClient: httpClient, BaseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"), Token: strings.TrimSpace(token)}
}

func (c *HTTPPluginClient) Resolve(ctx context.Context, giantMaterialID string) (Material, error) {
	id := strings.TrimSpace(giantMaterialID)
	if !materialIDPattern.MatchString(id) {
		return Material{}, &CodedError{Code: "INVALID_GIANT_MATERIAL_ID"}
	}
	base, err := url.Parse(c.BaseURL)
	if err != nil || (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" {
		return Material{}, &CodedError{Code: "GIANT_PLUGIN_ENDPOINT_INVALID"}
	}
	payload, err := json.Marshal(struct {
		GiantMaterialID string `json:"giantMaterialId"`
	}{GiantMaterialID: id})
	if err != nil {
		return Material{}, &CodedError{Code: "GIANT_PLUGIN_REQUEST_INVALID"}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/giant-material/resolve", bytes.NewReader(payload))
	if err != nil {
		return Material{}, &CodedError{Code: "GIANT_PLUGIN_REQUEST_INVALID"}
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	response, err := c.HTTPClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return Material{}, &CodedError{Code: "GIANT_PLUGIN_TIMEOUT"}
		}
		return Material{}, &CodedError{Code: "GIANT_PLUGIN_UNAVAILABLE"}
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusServiceUnavailable {
		return Material{}, &CodedError{Code: "GIANT_PLUGIN_UNAVAILABLE"}
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return Material{}, &CodedError{Code: "GIANT_PLUGIN_UPSTREAM_FAILED"}
	}
	var result struct {
		Material Material `json:"material"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(&result); err != nil {
		return Material{}, &CodedError{Code: "GIANT_PLUGIN_INVALID_RESPONSE"}
	}
	if strings.TrimSpace(result.Material.MaterialID) == "" || strings.TrimSpace(result.Material.VideoURL) == "" {
		return Material{}, &CodedError{Code: "GIANT_PLUGIN_INVALID_RESPONSE"}
	}
	result.Material.GiantMaterialID = id
	return result.Material, nil
}
