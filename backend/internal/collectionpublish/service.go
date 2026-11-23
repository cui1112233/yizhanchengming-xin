package collectionpublish

import (
	"cmp"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/n8comic"
)

type Service struct {
	Records   Records
	Submit    func(context.Context, n8comic.PublishRequest) (n8comic.PublishResult, error)
	Now       func() time.Time
	Lock      sync.Locker
	localLock sync.Mutex
	Acquire   func(context.Context, PublishRequest) (func(), error)
}

func ValidateSelection(kind CollectionKind, id string, refs []Reference) error {
	if kind != CollectionKindComic && kind != CollectionKindText {
		return ErrInvalid
	}
	for _, r := range refs {
		if r.Kind != CollectionKindComic && r.Kind != CollectionKindText {
			return ErrInvalid
		}
		if r.Kind != kind {
			return ErrMixedKinds
		}
	}
	for _, r := range refs {
		if r.ID != id {
			return fmt.Errorf("%w：每次请选择当前合集", ErrInvalid)
		}
	}
	return nil
}

func ValidateMapping(r PublishRequest) error {
	var missing []string
	if strings.TrimSpace(r.Producer) == "" {
		missing = append(missing, "制作者")
	}
	if r.Kind == CollectionKindText {
		if strings.TrimSpace(r.OrganizationID) == "" {
			missing = append(missing, "组织")
		}
		if strings.TrimSpace(r.CategoryID) == "" {
			missing = append(missing, "书籍分类")
		}
	}
	if r.AudioCategoryID <= 0 {
		missing = append(missing, "音频分类")
	}
	if r.AudioID <= 0 {
		missing = append(missing, "音频")
	}
	if len(r.WorkIDs) == 0 {
		missing = append(missing, "关联作品")
	}
	if len(missing) > 0 {
		return fmt.Errorf("%w：缺少%s", ErrInvalid, strings.Join(missing, "、"))
	}
	return nil
}

func (s *Service) Publish(ctx context.Context, r PublishRequest, prepare func(context.Context, *PublishRequest) (Collection, error)) (Record, error) {
	if err := ValidateSelection(r.Kind, r.CollectionID, r.Collections); err != nil {
		return Record{}, err
	}
	if r.Owner == "" || r.BatchID == "" || r.BookID == "" || r.CollectionID == "" {
		return Record{}, ErrInvalid
	}
	lock := s.Lock
	if lock == nil {
		lock = &s.localLock
	}
	lock.Lock()
	defer lock.Unlock()
	old, found, err := s.Records.Load(ctx, r)
	if err != nil {
		return Record{}, err
	}
	if found && old.BookID != "" && old.BookID != r.BookID {
		return Record{}, ErrInvalid
	}
	sameMapping := r.MappingRevision == "" || old.MappingRevision == "" || old.MappingRevision == r.MappingRevision
	if found && (old.Status == "pending" || old.Confirmed() && sameMapping && (!r.Force || r.RequestKey != "" && r.RequestKey == old.RequestKey)) {
		return old, nil
	}
	coll, err := prepare(ctx, &r)
	if err != nil {
		return Record{}, err
	}
	if coll.Kind != r.Kind {
		return Record{}, ErrMixedKinds
	}
	if coll.ID != r.CollectionID || coll.Name == "" || len(coll.Pages) == 0 {
		return Record{}, ErrInvalid
	}
	if err = ValidateMapping(r); err != nil {
		return Record{}, err
	}
	pages := slices.Clone(coll.Pages)
	slices.SortStableFunc(pages, func(a, b Page) int { return cmp.Compare(a.Ordinal, b.Ordinal) })
	images := make([]n8comic.ImageInput, len(pages))
	for i, p := range pages {
		if i > 0 && p.Ordinal == pages[i-1].Ordinal {
			return Record{}, fmt.Errorf("%w：合集中存在重复页码", ErrInvalid)
		}
		images[i] = p.Image
	}
	rec := Record{RunID: r.RunID, Attempt: r.Attempt, MappingRevision: r.MappingRevision, Kind: r.Kind, BookID: r.BookID, CollectionID: r.CollectionID, IntentID: rand.Text(), RequestKey: r.RequestKey, Status: "pending", Name: coll.Name, ProducerAdminName: r.Producer, OrganizationID: r.OrganizationID, CategoryID: r.CategoryID, WebsiteProfileID: r.WebsiteProfileID, AudioCategoryID: r.AudioCategoryID, AudioID: r.AudioID, WorkIDs: slices.Clone(r.WorkIDs), PageCount: len(images)}
	if old.Confirmed() {
		old.LastPublished = nil
		rec.LastPublished = &old
	} else {
		rec.LastPublished = old.LastPublished
	}
	// Historical incomplete receipts contain image IDs in the frozen collection order.
	if found && (old.Status == "upload_incomplete" || old.Status == "uploaded") && len(old.ImageIDs) == len(images) && old.ProducerAdminName == r.Producer {
		rec.ImageIDs = slices.Clone(old.ImageIDs)
		if old.IntentID != "" {
			rec.IntentID = old.IntentID
		}
	}
	// Only the durable claim needs the inter-process lease. Preparation can be
	// slow, so compare the receipt again after obtaining it.
	if s.Acquire != nil {
		release, err := s.Acquire(ctx, r)
		if err != nil {
			return Record{}, err
		}
		current, exists, err := s.Records.Load(ctx, r)
		if err != nil {
			release()
			return Record{}, err
		}
		if exists != found || current.IntentID != old.IntentID || current.Status != old.Status {
			release()
			return current, nil
		}
		err = s.Records.Save(ctx, r, rec)
		release()
		if err != nil {
			return Record{}, err
		}
	} else if err = s.Records.Save(ctx, r, rec); err != nil {
		return Record{}, err
	}
	result, sendErr := s.Submit(ctx, n8comic.PublishRequest{ProducerAdminName: r.Producer, Name: coll.Name, Images: images, WorkIDs: r.WorkIDs, AudioCategoryID: r.AudioCategoryID, AudioID: r.AudioID, UploadedImageIDs: rec.ImageIDs})
	rec.CarouselID = result.CarouselID
	if len(result.ImageIDs) > 0 {
		rec.ImageIDs = slices.Clone(result.ImageIDs)
	}
	rec.Status = "published"
	switch {
	case errors.Is(sendErr, n8comic.ErrCarouselReceiptPending):
		rec.Status = "pending"
		sendErr = nil
	case sendErr != nil:
		rec.Status = "failed"
		if slices.ContainsFunc(rec.ImageIDs, func(id int64) bool { return id > 0 }) {
			rec.Status = "upload_incomplete"
			sendErr = fmt.Errorf("%w：%w", ErrIncomplete, sendErr)
		}
	case !rec.Confirmed():
		rec.Status = "pending"
	default:
		now := time.Now
		if s.Now != nil {
			now = s.Now
		}
		rec.PublishedAt = now()
	}
	// A caller may disconnect after the remote reply. Keep the receipt durable.
	saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	if err = s.Records.Save(saveCtx, r, rec); err != nil {
		return Record{}, fmt.Errorf("保存上传回执失败：%w", err)
	}
	return rec, sendErr
}
