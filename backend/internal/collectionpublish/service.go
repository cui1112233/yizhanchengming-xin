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
	Readback  func(context.Context, n8comic.ReceiptLookup) (n8comic.PublishResult, error)
	Records   Records
	Submit    func(context.Context, n8comic.PublishRequest) (n8comic.PublishResult, error)
	Now       func() time.Time
	Lock      sync.Locker
	LockFor   func(PublishRequest) sync.Locker
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
	if s.LockFor != nil {
		lock = s.LockFor(r)
	}
	if lock == nil {
		lock = &s.localLock
	}
	lock.Lock()
	defer lock.Unlock()
	old, found, err := s.Records.Load(ctx, r)
	if err != nil {
		return Record{}, err
	}
	if found && old.Kind != "" && old.Kind != r.Kind {
		return Record{}, ErrMixedKinds
	}
	if found && (old.BookID != "" && old.BookID != r.BookID || old.CollectionID != "" && old.CollectionID != r.CollectionID) {
		return Record{}, ErrInvalid
	}
	if found && old.Status == "pending" {
		old, err = s.refresh(ctx, r, old)
		if err != nil || !old.Confirmed() {
			return old, err
		}
		// This call originally observed an outstanding attempt, even on force.
		// Completing its readback must never initiate another submission.
		return s.reuse(ctx, r, old, old)
	}
	if found && old.Confirmed() && r.Force && r.RequestKey != "" && old.RequestKey == r.RequestKey {
		return old, nil
	}
	if found && old.Confirmed() && !r.Force {
		return s.reuse(ctx, r, old, old)
	}
	resuming := found && old.Status == "uploading" && old.IntentID != ""
	if resuming {
		r.Resuming = true
		r.Producer, r.OrganizationID, r.WebsiteProfileID, r.CategoryID = old.ProducerAdminName, old.OrganizationID, old.WebsiteProfileID, old.CategoryID
		r.WorkIDs, r.AudioCategoryID, r.AudioID = old.WorkIDs, old.AudioCategoryID, old.AudioID
	}
	if found && !resuming && !r.Force && old.LastPublished != nil && old.LastPublished.Confirmed() {
		return s.reuse(ctx, r, old, *old.LastPublished)
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
	rec := Record{RunID: r.RunID, Attempt: r.Attempt, MappingRevision: r.MappingRevision, Kind: r.Kind, BookID: r.BookID, CollectionID: r.CollectionID, IntentID: rand.Text(), RequestKey: r.RequestKey, DispatchID: rand.Text(), Phase: "preparing", Status: "uploading", Name: coll.Name, ProducerAdminName: r.Producer, OrganizationID: r.OrganizationID, CategoryID: r.CategoryID, WebsiteProfileID: r.WebsiteProfileID, AudioCategoryID: r.AudioCategoryID, AudioID: r.AudioID, WorkIDs: slices.Clone(r.WorkIDs), PageCount: len(images)}
	rec.PriorCarouselIDs = slices.Clone(old.PriorCarouselIDs)
	if history := old.HistoricalPublished(); history != nil {
		origin := *history
		origin.LastPublished = nil
		rec.LastPublished = &origin
		if !slices.Contains(rec.PriorCarouselIDs, origin.CarouselID) {
			rec.PriorCarouselIDs = append(rec.PriorCarouselIDs, origin.CarouselID)
		}
	}
	if resuming {
		rec.IntentID, rec.RequestKey, rec.RunID, rec.Attempt, rec.Name, rec.MappingRevision = old.IntentID, old.RequestKey, old.RunID, old.Attempt, old.Name, old.MappingRevision
		rec.ImageIDs = slices.Clone(old.ImageIDs)
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
		if exists != found || current.IntentID != old.IntentID || current.Status != old.Status || current.DispatchID != old.DispatchID {
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
	expected := rec
	checkpoint := func(next Record) error {
		current, err := s.saveCurrent(ctx, r, expected, next)
		if err != nil {
			return err
		}
		if current.IntentID != next.IntentID || current.DispatchID != next.DispatchID || current.Phase != next.Phase {
			return errors.New("另一个上传请求正在处理本合集")
		}
		expected = current
		return nil
	}
	rec.Phase, rec.Message = "uploading", "正在上传合集图片；中断后可继续上传剩余图片。"
	if err := checkpoint(rec); err != nil {
		return Record{}, err
	}
	onImage := func(index int, id int64) error {
		if len(rec.ImageIDs) == 0 {
			rec.ImageIDs = make([]int64, len(images))
		} else {
			rec.ImageIDs = slices.Clone(rec.ImageIDs)
		}
		if index < 0 || index >= len(rec.ImageIDs) || id <= 0 {
			return ErrInvalid
		}
		rec.ImageIDs[index] = id
		return checkpoint(rec)
	}
	beforeCreate := func(proof n8comic.PublishResult) error {
		rec.ImageIDs = slices.Clone(proof.ImageIDs)
		rec.ImageSignatures = slices.Clone(proof.ImageSignatures)
		rec.Phase, rec.Status, rec.Message = "create_unknown", "pending", pendingMessage
		return checkpoint(rec)
	}
	result, sendErr := s.Submit(ctx, n8comic.PublishRequest{IntentID: rec.IntentID, OnImage: onImage, BeforeCreate: beforeCreate, ProducerAdminName: r.Producer, Name: rec.Name, Images: images, WorkIDs: r.WorkIDs, AudioCategoryID: r.AudioCategoryID, AudioID: r.AudioID, UploadedImageIDs: slices.Clone(rec.ImageIDs)})
	rec.OperationID = result.OperationID
	rec.ReceiptPath = result.ReceiptPath
	if len(result.ImageSignatures) > 0 {
		rec.ImageSignatures = slices.Clone(result.ImageSignatures)
	}
	rec.CarouselID = result.CarouselID
	if len(result.ImageIDs) > 0 {
		rec.ImageIDs = slices.Clone(result.ImageIDs)
	}
	rec.Status, rec.Phase, rec.Message = "published", "confirmed", ""
	switch {
	case errors.Is(sendErr, n8comic.ErrCarouselReceiptPending):
		rec.Status, rec.Phase = "pending", "create_unknown"
		rec.Message = pendingMessage
		sendErr = nil
	case sendErr != nil:
		rec.Status, rec.Phase, rec.ErrorMessage = "failed", "failed", sendErr.Error()
		if slices.ContainsFunc(rec.ImageIDs, func(id int64) bool { return id > 0 }) {
			rec.Status = "upload_incomplete"
			sendErr = fmt.Errorf("%w：%w", ErrIncomplete, sendErr)
		}
	case !rec.Confirmed():
		rec.Status, rec.Phase = "pending", "create_unknown"
		rec.Message = pendingMessage
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
	if rec, err = s.saveCurrent(saveCtx, r, expected, rec); err != nil {
		return Record{}, fmt.Errorf("保存上传回执失败：%w", err)
	}
	return rec, sendErr
}

const pendingMessage = "等待平台返回合集回执；如长时间未更新，请到发布后台核对。"

// Status polls a known pending submission without creating a new one.
func (s *Service) Status(ctx context.Context, r PublishRequest) (Record, bool, error) {
	if err := ValidateSelection(r.Kind, r.CollectionID, r.Collections); err != nil {
		return Record{}, false, err
	}
	if r.Owner == "" || r.BatchID == "" || r.BookID == "" || r.CollectionID == "" {
		return Record{}, false, ErrInvalid
	}
	lock := s.Lock
	if s.LockFor != nil {
		lock = s.LockFor(r)
	}
	if lock == nil {
		lock = &s.localLock
	}
	lock.Lock()
	defer lock.Unlock()
	rec, found, err := s.Records.Load(ctx, r)
	if err != nil || !found || rec.Status != "pending" {
		return rec, found, err
	}
	rec, err = s.refresh(ctx, r, rec)
	return rec, true, err
}

func (s *Service) refresh(ctx context.Context, r PublishRequest, old Record) (Record, error) {
	old.Message = pendingMessage
	if s.Readback == nil || old.ReceiptPath == "" || old.IntentID == "" || len(old.ImageIDs) == 0 {
		return old, nil
	}
	proof := n8comic.ReceiptLookup{IntentID: old.IntentID, OperationID: old.OperationID, PriorCarouselIDs: old.PriorCarouselIDs, ReceiptPath: old.ReceiptPath, ImageSignatures: old.ImageSignatures, ImageIDs: old.ImageIDs}
	if old.LastPublished != nil {
		proof.PreviousCarouselID = old.LastPublished.CarouselID
	}
	remote, err := s.Readback(ctx, proof)
	if err != nil {
		var rejection *n8comic.RejectedReceipt
		if errors.As(err, &rejection) && rejection.IntentID == old.IntentID {
			next := old
			next.Status, next.Phase, next.Message, next.ErrorMessage = "failed", "failed", rejection.Reason, rejection.Reason
			return s.saveCurrent(ctx, r, old, next)
		}
		return old, nil
	}
	if remote.ReceiptIntentID != old.IntentID || remote.CarouselID <= 0 || remote.CarouselID == proof.PreviousCarouselID || slices.Contains(old.PriorCarouselIDs, remote.CarouselID) || !slices.Equal(old.ImageIDs, remote.ImageIDs) {
		return old, nil
	}
	next := old
	next.Status, next.Phase = "published", "confirmed"
	next.CarouselID = remote.CarouselID
	next.ImageIDs = slices.Clone(remote.ImageIDs)
	next.Message = ""
	if !next.Confirmed() {
		return old, nil
	}
	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	next.PublishedAt = now()
	return s.saveCurrent(ctx, r, old, next)
}

// A workflow can complete by reusing a manual receipt. The original publication
// stays in lastPublished and the derived result identifies the run that reused it.
func (s *Service) reuse(ctx context.Context, r PublishRequest, current, receipt Record) (Record, error) {
	if r.RunID == "" || r.RunID == receipt.RunID && r.Attempt == receipt.Attempt && r.RequestKey == receipt.RequestKey {
		return receipt, nil
	}
	origin := *receipt.HistoricalPublished()
	origin.LastPublished = nil
	next := receipt
	next.Reused = true
	next.LastPublished = &origin
	next.RunID, next.Attempt, next.RequestKey = r.RunID, r.Attempt, r.RequestKey
	return s.saveCurrent(ctx, r, current, next)
}

func (s *Service) saveCurrent(ctx context.Context, r PublishRequest, expected, next Record) (Record, error) {
	if s.Acquire != nil {
		release, err := s.Acquire(ctx, r)
		if err != nil {
			return Record{}, err
		}
		defer release()
	}
	latest, ok, err := s.Records.Load(ctx, r)
	if err != nil {
		return Record{}, err
	}
	if !ok || latest.IntentID != expected.IntentID || latest.Status != expected.Status || latest.RequestKey != expected.RequestKey || latest.DispatchID != expected.DispatchID || latest.Phase != expected.Phase {
		return latest, nil
	}
	if err = s.Records.Save(ctx, r, next); err != nil {
		return Record{}, err
	}
	return next, nil
}
