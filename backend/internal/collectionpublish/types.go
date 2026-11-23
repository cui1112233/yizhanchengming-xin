package collectionpublish

import (
	"context"
	"errors"
	"time"

	"qiantie/backend/internal/n8comic"
)

type CollectionKind string

const (
	CollectionKindComic CollectionKind = "comic"
	CollectionKindText  CollectionKind = "text"
)

var (
	ErrMixedKinds = errors.New("漫画与文本不能一起上传，请分开选择。")
	ErrInvalid    = errors.New("合集上传参数无效")
	ErrIncomplete = errors.New("图片已上传，但合集尚未创建完成")
)

type Reference struct {
	Kind CollectionKind `json:"kind"`
	ID   string         `json:"id"`
}
type PublishRequest struct {
	RunID, MappingRevision                                 string
	Attempt                                                int
	Kind                                                   CollectionKind
	Owner, BatchID, BookID, CollectionID                   string
	Producer, OrganizationID, CategoryID, WebsiteProfileID string
	AudioCategoryID, AudioID                               int64
	WorkIDs                                                []int64
	Force                                                  bool
	RequestKey                                             string
	Collections                                            []Reference
}
type Page struct {
	Ordinal int
	Image   n8comic.ImageInput
}
type Collection struct {
	Kind     CollectionKind
	ID, Name string
	Pages    []Page
}
type Record struct {
	RunID             string         `json:"runId,omitempty"`
	Attempt           int            `json:"attempt,omitempty"`
	MappingRevision   string         `json:"mappingRevision,omitempty"`
	LastPublished     *Record        `json:"lastPublished,omitempty"`
	Kind              CollectionKind `json:"kind,omitempty"`
	IntentID          string         `json:"intentId,omitempty"`
	RequestKey        string         `json:"requestKey,omitempty"`
	BookID            string         `json:"bookId,omitempty"`
	CollectionID      string         `json:"collectionId,omitempty"`
	OrganizationID    string         `json:"organizationId,omitempty"`
	CategoryID        string         `json:"categoryId,omitempty"`
	WebsiteProfileID  string         `json:"websiteProfileId,omitempty"`
	AudioID           int64          `json:"audioId,omitempty"`
	Status            string         `json:"status"`
	Name              string         `json:"name"`
	ProducerAdminName string         `json:"producerAdminName"`
	AudioCategoryID   int64          `json:"audioCategoryId"`
	WorkIDs           []int64        `json:"workIds"`
	CarouselID        int64          `json:"carouselId"`
	ImageIDs          []int64        `json:"imageIds"`
	PageCount         int            `json:"pageCount,omitempty"`
	PublishedAt       time.Time      `json:"publishedAt"`
}

func (r Record) Confirmed() bool {
	if r.Status != "published" || r.CarouselID <= 0 || len(r.ImageIDs) == 0 || (r.PageCount > 0 && (len(r.ImageIDs) < r.PageCount || r.Kind == CollectionKindText && len(r.ImageIDs) != r.PageCount)) {
		return false
	}
	for _, id := range r.ImageIDs {
		if id <= 0 {
			return false
		}
	}
	return true
}

type Records interface {
	Load(context.Context, PublishRequest) (Record, bool, error)
	Save(context.Context, PublishRequest, Record) error
}
