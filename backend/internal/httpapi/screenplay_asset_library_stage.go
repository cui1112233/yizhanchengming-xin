package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/screenplay"
)

type screenplayLibraryStageInput struct {
	RequestKey            string `json:"requestKey"`
	TargetID              string `json:"targetId"`
	ExpectedEntryRevision *int64 `json:"expectedEntryRevision"`
}

// StageImage prepares a private project copy for an editor draft. This is CM's
// storage adapter; staging never selects it or copies the library's text fields.
func (s *ScreenplayAssetLibrary) StageImage(ctx context.Context, actor, projectID, entryID string, in screenplayLibraryStageInput) (screenplay.MediaItem, error) {
	if s == nil || s.Repository == nil || s.Media == nil || s.Media.Projects == nil || s.Media.Media == nil || s.Media.Cut == nil || s.Media.Assets == nil || s.Media.Objects == nil {
		return screenplay.MediaItem{}, fmt.Errorf("%w: 图片准备服务未配置", screenplay.ErrInvalidState)
	}
	if !validAssetLibraryRequestKey(in.RequestKey) || strings.TrimSpace(in.TargetID) == "" || in.ExpectedEntryRevision == nil || *in.ExpectedEntryRevision < 1 {
		return screenplay.MediaItem{}, screenplay.ErrInvalidArgument
	}
	if err := ctx.Err(); err != nil {
		return screenplay.MediaItem{}, err
	}
	project, err := s.Media.resolve(ctx, actor, projectID, "edit")
	if err != nil {
		return screenplay.MediaItem{}, err
	}
	if project.OwnerUsername != actor {
		return screenplay.MediaItem{}, screenplay.ErrAccessDenied
	}
	var result screenplay.MediaItem
	prepare := func() error {
		current, e := s.Media.resolve(ctx, actor, projectID, "edit")
		if e != nil {
			return e
		}
		if current.OwnerUsername != actor {
			return screenplay.ErrAccessDenied
		}
		result, e = s.stageImageLocked(ctx, actor, current, entryID, in)
		return e
	}
	if s.Media.Authoring != nil {
		err = s.Media.Authoring.WithLock(ctx, actor, projectID, prepare)
	} else {
		err = prepare()
	}
	return result, err
}
func (s *ScreenplayAssetLibrary) stageImageLocked(ctx context.Context, actor string, project screenplay.Project, entryID string, in screenplayLibraryStageInput) (screenplay.MediaItem, error) {
	entry, err := s.Repository.GetAssetLibraryEntry(ctx, actor, entryID)
	if err != nil {
		return screenplay.MediaItem{}, err
	}
	if entry.Revision != *in.ExpectedEntryRevision {
		return screenplay.MediaItem{}, screenplay.ErrRevisionConflict
	}
	assets, err := s.Media.Projects.ListAssets(ctx, actor, project.ID, "")
	if err != nil {
		return screenplay.MediaItem{}, err
	}
	targetFound := false
	for _, asset := range assets {
		if asset.ID == in.TargetID {
			targetFound = true
			break
		}
	}
	if !targetFound {
		return screenplay.MediaItem{}, fmt.Errorf("%w: 目标资产不存在", screenplay.ErrNotFound)
	}
	extension := map[string]string{"image/png": "png", "image/jpeg": "jpg", "image/webp": "webp", "image/gif": "gif", "image/avif": "avif"}[entry.ImageObject.ContentType]
	if extension == "" || entry.ImageObject.ObjectKey == "" || entry.ImageObject.ByteSize <= 0 || entry.ImageObject.ByteSize > screenplayLibraryImageLimit {
		return screenplay.MediaItem{}, fmt.Errorf("%w: 库图片未持久保存或超过40MB", screenplay.ErrInvalidArgument)
	}
	digest := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%d\x00%s", actor, project.ID, in.TargetID, entry.ID, entry.Revision, in.RequestKey)))
	id := fmt.Sprintf("sp-lib-stage-%x", digest[:16])
	key, err := agentworkspace.AgentAssetObjectKey(actor, project.ID, id, extension)
	if err != nil {
		return screenplay.MediaItem{}, err
	}
	key = strings.Replace(key, "agent-assets/", "aicut-assets/", 1)
	name := strings.TrimSpace(entry.Name)
	if name == "" {
		name = "库图片"
	}
	for len(name) > 170 {
		runes := []rune(name)
		name = string(runes[:len(runes)-1])
	}
	expected := agentworkspace.Asset{ID: id, ProjectID: project.ID, Kind: "image", Name: name + "." + extension, ContentType: entry.ImageObject.ContentType, ObjectKey: key, ByteSize: entry.ImageObject.ByteSize, CreatedAt: time.Now().UTC()}
	state, err := s.Media.Media.LoadMedia(ctx, actor, project.ID)
	if err != nil {
		return screenplay.MediaItem{}, err
	}
	var item screenplay.MediaItem
	for _, current := range state.Items {
		if current.ID == id {
			if current.AssetID != id || current.TargetID != in.TargetID || current.Kind != "image" || current.Status != "succeeded" {
				return screenplay.MediaItem{}, screenplay.ErrRevisionConflict
			}
			item = current
			expected.CreatedAt = current.CreatedAt
			break
		}
	}
	cut, cutErr := s.Media.Cut.Asset(ctx, actor, project.ID, id)
	if cutErr != nil && !errors.Is(cutErr, aicut.ErrAssetNotFound) {
		return screenplay.MediaItem{}, cutErr
	}
	if cutErr == nil {
		if !sameStageAsset(expected, aicutObjectAsset(cut)) {
			return screenplay.MediaItem{}, screenplay.ErrRevisionConflict
		}
		expected.CreatedAt = cut.CreatedAt
	}
	agent, agentErr := s.Media.Assets.Asset(ctx, actor, project.ID, id)
	if agentErr != nil && !errors.Is(agentErr, agentworkspace.ErrAssetNotFound) {
		return screenplay.MediaItem{}, agentErr
	}
	if agentErr == nil {
		if !sameStageAsset(expected, agent) {
			return screenplay.MediaItem{}, screenplay.ErrRevisionConflict
		}
		expected.CreatedAt = agent.CreatedAt
	}
	// A deterministic object key permits replay after a partial write. Existing
	// object bytes are immutable for this entry revision; never delete them on an
	// ambiguous database result or on editor cancellation.
	if item.ID != "" && cutErr == nil && agentErr == nil {
		return item, nil
	}
	if err = ctx.Err(); err != nil {
		return screenplay.MediaItem{}, err
	}
	source := agentworkspace.Asset{ID: entry.Image.AssetID, ProjectID: screenplayLibraryObjectProject, Kind: "image", Name: entry.Image.Name, ContentType: entry.ImageObject.ContentType, ObjectKey: entry.ImageObject.ObjectKey, ByteSize: entry.ImageObject.ByteSize}
	reader, err := s.Media.Objects.Open(ctx, source)
	if err != nil {
		return screenplay.MediaItem{}, fmt.Errorf("库图片读取失败，请稍后重试")
	}
	raw, readErr := io.ReadAll(io.LimitReader(reader, screenplayLibraryImageLimit+1))
	closeErr := reader.Close()
	if readErr != nil {
		return screenplay.MediaItem{}, readErr
	}
	if closeErr != nil {
		return screenplay.MediaItem{}, closeErr
	}
	if len(raw) == 0 || len(raw) > screenplayLibraryImageLimit || int64(len(raw)) != expected.ByteSize {
		return screenplay.MediaItem{}, screenplay.ErrInvalidArgument
	}
	if err = ctx.Err(); err != nil {
		return screenplay.MediaItem{}, err
	}
	if err = s.Media.Objects.Put(ctx, expected, bytes.NewReader(raw)); err != nil {
		return screenplay.MediaItem{}, err
	}
	if err = ctx.Err(); err != nil {
		return screenplay.MediaItem{}, err
	}
	if err = s.Media.ensureCut(ctx, project); err != nil {
		return screenplay.MediaItem{}, err
	}
	if cutErr != nil {
		err = s.Media.Cut.CreateAsset(ctx, actor, aicut.Asset{ID: id, ProjectID: project.ID, Kind: "image", Name: expected.Name, ContentType: expected.ContentType, ObjectKey: key, ByteSize: expected.ByteSize, CreatedAt: expected.CreatedAt})
		if err != nil {
			verify, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			existing, e := s.Media.Cut.Asset(verify, actor, project.ID, id)
			cancel()
			if e != nil || !sameStageAsset(expected, aicutObjectAsset(existing)) {
				return screenplay.MediaItem{}, err
			}
		}
	}
	if err = ctx.Err(); err != nil {
		return screenplay.MediaItem{}, err
	}
	if agentErr != nil {
		err = s.Media.Assets.CreateAsset(ctx, actor, expected)
		if err != nil {
			verify, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			existing, e := s.Media.Assets.Asset(verify, actor, project.ID, id)
			cancel()
			if e != nil || !sameStageAsset(expected, existing) {
				return screenplay.MediaItem{}, err
			}
		}
	}
	if item.ID == "" {
		item = screenplay.MediaItem{ID: id, AssetID: id, Name: expected.Name, Kind: "image", Status: "succeeded", TargetID: in.TargetID, CreatedAt: expected.CreatedAt}
	}
	if err = ctx.Err(); err != nil {
		return screenplay.MediaItem{}, err
	}
	_, err = s.Media.Media.MutateMedia(ctx, actor, project.ID, func(current *screenplay.MediaState) error {
		for _, previous := range current.Items {
			if previous.ID == id {
				if previous.AssetID != id || previous.TargetID != in.TargetID || previous.Kind != "image" || previous.Status != "succeeded" {
					return screenplay.ErrRevisionConflict
				}
				item = previous
				return nil
			}
		}
		current.Items = append(current.Items, item)
		return nil
	})
	if err != nil {
		verify, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		latest, e := s.Media.Media.LoadMedia(verify, actor, project.ID)
		cancel()
		if e == nil {
			for _, saved := range latest.Items {
				if saved.ID == id && saved.AssetID == id && saved.TargetID == in.TargetID && saved.Kind == "image" && saved.Status == "succeeded" {
					return saved, nil
				}
			}
		}
		return screenplay.MediaItem{}, err
	}
	return item, nil
}
func sameStageAsset(expected, actual agentworkspace.Asset) bool {
	return expected.ID == actual.ID && expected.ProjectID == actual.ProjectID && actual.Kind == "image" && expected.ObjectKey == actual.ObjectKey && expected.ContentType == actual.ContentType && expected.ByteSize == actual.ByteSize
}
