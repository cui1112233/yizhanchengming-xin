package collectionpublish

import (
	"context"
	"encoding/json"
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
	Resuming                                               bool
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

// TextKeyScope keeps the exact byte representation used by the first key registry.
func (r PublishRequest) TextKeyScope() string {
	encoded, _ := json.Marshal([]string{r.Owner, r.BatchID, r.BookID, r.CollectionID})
	return string(encoded)
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
	Reused            bool           `json:"reused,omitempty"`
	Phase             string         `json:"phase,omitempty"`
	DispatchID        string         `json:"-"`
	OperationID       string         `json:"-"`
	ErrorMessage      string         `json:"errorMessage,omitempty"`
	PriorCarouselIDs  []int64        `json:"priorCarouselIds,omitempty"`
	ReceiptPath       string         `json:"-"`
	ImageSignatures   []string       `json:"-"`
	Message           string         `json:"message,omitempty"`
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
	if r.Status != "published" || r.CarouselID <= 0 || len(r.ImageIDs) == 0 {
		return false
	}
	if r.PageCount > 0 && (len(r.ImageIDs) < r.PageCount || r.Kind == CollectionKindText && len(r.ImageIDs) != r.PageCount) {
		return false
	}
	seen := make(map[int64]bool, len(r.ImageIDs))
	for _, id := range r.ImageIDs {
		if id <= 0 || seen[id] {
			return false
		}
		seen[id] = true
	}
	return true
}

type Records interface {
	Load(context.Context, PublishRequest) (Record, bool, error)
	Save(context.Context, PublishRequest, Record) error
}

// HistoricalPublished returns the identity of the physical successful attempt.
// A workflow-derived record shares its intent ID with that original receipt.
func (r *Record) HistoricalPublished() *Record {
	if r == nil {
		return nil
	}
	if r.LastPublished != nil && r.LastPublished.Confirmed() && (!r.Confirmed() || r.Reused || r.IntentID != "" && r.LastPublished.IntentID == r.IntentID) {
		return r.LastPublished
	}
	if r.Confirmed() {
		return r
	}
	return nil
}
