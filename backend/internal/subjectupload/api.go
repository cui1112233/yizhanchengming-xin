// Package subjectupload implements provider asset registration, not video generation.
package subjectupload

import (
	"context"
	"net/http"
	"time"
)

type Config struct {
	Provider, AccessKeyID, SecretAccessKey, Token, ProjectName, Region string
	ImageHostURL, ImageHostToken                                       string
}
type Input struct {
	Name, FileName, Description, GroupID, ImageURL, ContentType string
	Image                                                       []byte
}
type Progress struct {
	ImageURL      string `json:"image_url,omitempty"`
	GroupID       string `json:"group_id,omitempty"`
	AssetID       string `json:"asset_id,omitempty"`
	PendingAction string `json:"pending_action,omitempty"`
	Status        string `json:"status,omitempty"`
}
type Result struct {
	AssetID     string `json:"assetId"`
	AssetURI    string `json:"assetUri"`
	GroupID     string `json:"groupId,omitempty"`
	ProjectName string `json:"projectName,omitempty"`
	Provider    string `json:"provider"`
	Status      string `json:"status"`
}
type Checkpoint func(context.Context, Progress) error
type Runner interface {
	Run(context.Context, Config, Input, Progress, Checkpoint) (Result, error)
}

// Client endpoints are injectable for tests; production constructor fixes known HTTPS origins.
type Client struct {
	HTTP                          *http.Client
	FireEndpoint, DeyunaiEndpoint string
	Now                           func() time.Time
	PollInterval                  time.Duration
	MaxWait                       time.Duration
}
