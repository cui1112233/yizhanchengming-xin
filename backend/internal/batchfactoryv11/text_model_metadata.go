package batchfactoryv11

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// TextModelMetadata contains only public identifiers and advertised controls.
// The digest excludes credentials, URL user information, paths and queries.
type TextModelMetadata struct {
	CatalogOwner        string `json:"catalogOwner,omitempty"`
	ActualModel         string `json:"actualModel,omitempty"`
	ChannelID           string `json:"channelId,omitempty"`
	ConfigRevision      string `json:"configRevision,omitempty"`
	SafeConfigDigest    string `json:"safeConfigDigest,omitempty"`
	SupportsTemperature *bool  `json:"supportsTemperature,omitempty"`
	SupportsMaxTokens   *bool  `json:"supportsMaxTokens,omitempty"`
}
type TextModelMetadataProvider interface{ TextModelMetadata() TextModelMetadata }

func NewTextModelMetadata(model, channel, endpoint string, config map[string]any) TextModelMetadata {
	m := TextModelMetadata{ActualModel: model, ChannelID: channel}
	for _, key := range []string{"configRevision", "updatedAt"} {
		if v, ok := config[key]; ok && v != nil {
			m.ConfigRevision = fmt.Sprint(v)
			break
		}
	}
	if caps, ok := config["capabilities"].(map[string]any); ok {
		if v, ok := caps["supportsTemperature"].(bool); ok {
			m.SupportsTemperature = &v
		}
		if v, ok := caps["supportsMaxTokens"].(bool); ok {
			m.SupportsMaxTokens = &v
		}
	}
	host := ""
	if u, err := url.Parse(endpoint); err == nil {
		host = strings.ToLower(u.Hostname())
	}
	raw, _ := json.Marshal(struct {
		Model, Channel, Host, Revision string
		Temperature, Tokens            *bool
	}{model, channel, host, m.ConfigRevision, m.SupportsTemperature, m.SupportsMaxTokens})
	m.SafeConfigDigest = fmt.Sprintf("%x", sha256.Sum256(raw))
	return m
}

func ApplyTextParameterOmissions(payload map[string]any, r TextCompletionRequest) {
	if r.OmitTemperature {
		delete(payload, "temperature")
	}
	if r.OmitMaxTokens {
		delete(payload, "max_tokens")
	}
}
