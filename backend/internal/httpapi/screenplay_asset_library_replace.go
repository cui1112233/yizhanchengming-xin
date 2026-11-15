package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"strings"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/screenplay"
)

const libraryReplaceKey = "_cm_library_image_replace_key"
const libraryReplaceDigest = "_cm_library_image_replace_digest"

func libraryImageObject(e screenplay.AssetLibraryEntry) agentworkspace.Asset {
	return agentworkspace.Asset{ID: e.Image.AssetID, ProjectID: screenplayLibraryObjectProject, Kind: "image", Name: e.Image.Name, ObjectKey: e.ImageObject.ObjectKey, ContentType: e.ImageObject.ContentType, ByteSize: e.ImageObject.ByteSize}
}

// ReplaceImage applies the editor draft and its new image in one row revision.
// The existing project copies and all unedited snapshot fields remain independent.
func (s *ScreenplayAssetLibrary) ReplaceImage(ctx context.Context, actor, id string, expected int64, in screenplayLibraryUploadInput) (screenplay.AssetLibraryEntry, error) {
	return s.replaceLibraryEditor(ctx, actor, id, expected, in, true)
}
func (s *ScreenplayAssetLibrary) replaceLibraryEditor(ctx context.Context, actor, id string, expected int64, in screenplayLibraryUploadInput, requireImage bool) (screenplay.AssetLibraryEntry, error) {
	fail := func(err error) (screenplay.AssetLibraryEntry, error) { return screenplay.AssetLibraryEntry{}, err }
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if s == nil || s.Repository == nil || s.Media == nil || s.Media.Objects == nil {
		return fail(screenplay.ErrInvalidState)
	}
	if expected < 1 || !validAssetLibraryRequestKey(in.RequestKey) || strings.TrimSpace(in.Name) == "" || requireImage && len(in.Bytes) == 0 || len(in.Bytes) > screenplayLibraryImageLimit {
		return fail(screenplay.ErrInvalidArgument)
	}
	var format string
	if len(in.Bytes) > 0 {
		cfg, detected, err := image.DecodeConfig(bytes.NewReader(in.Bytes))
		format = detected
		if err != nil || cfg.Width < 1 || cfg.Height < 1 {
			return fail(fmt.Errorf("%w: 请上传有效的PNG、JPEG、WEBP或GIF图片", screenplay.ErrInvalidArgument))
		}
	}
	extension := map[string]string{"png": "png", "jpeg": "jpg", "webp": "webp", "gif": "gif"}[format]
	contentType := map[string]string{"png": "image/png", "jpeg": "image/jpeg", "webp": "image/webp", "gif": "image/gif"}[format]
	if len(in.Bytes) > 0 && extension == "" {
		return fail(screenplay.ErrInvalidArgument)
	}
	entry, err := s.Repository.GetAssetLibraryEntry(ctx, actor, id)
	if err != nil {
		return fail(err)
	}
	if entry.Type != screenplay.AssetCharacter || in.Type != entry.Type {
		return fail(screenplay.ErrInvalidArgument)
	}
	audio, err := s.prepareLibraryAudio(actor, in)
	if err != nil {
		return fail(err)
	}
	rawHash := sha256.Sum256(in.Bytes)
	raw, _ := json.Marshal([]any{id, expected, in.Type, strings.TrimSpace(in.Name), in.Description, in.Category, in.Prompt, in.Voice, fmt.Sprintf("%x", rawHash)})
	if len(in.AudioBytes) > 0 {
		sum := sha256.Sum256(in.AudioBytes)
		raw = append(raw, []byte(in.AudioName+fmt.Sprintf("%x", sum))...)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(raw))
	if entry.Snapshot.Extra[libraryReplaceKey] == in.RequestKey {
		if entry.Snapshot.Extra[libraryReplaceDigest] != digest {
			return fail(screenplay.ErrRevisionConflict)
		}
		return entry, nil
	}
	if entry.Revision != expected {
		return fail(screenplay.ErrRevisionConflict)
	}
	old := entry
	oldAudio, _ := libraryAudioObject(entry)
	imageID := newAgentID("sp-library-replace")
	objectKey := ""
	if len(in.Bytes) > 0 {
		objectKey, err = agentworkspace.AgentAssetObjectKey(actor, screenplayLibraryObjectProject, imageID, extension)
		if err != nil {
			return fail(err)
		}
	}
	object := agentworkspace.Asset{ID: imageID, ProjectID: screenplayLibraryObjectProject, Kind: "image", Name: imageID + "." + extension, ObjectKey: objectKey, ContentType: contentType, ByteSize: int64(len(in.Bytes)), CreatedAt: time.Now().UTC()}
	newObjects := []agentworkspace.Asset{}
	if len(in.Bytes) > 0 {
		newObjects = append(newObjects, object)
	}
	if audio.ID != "" {
		newObjects = append(newObjects, audio)
	}
	cleanupNew := func() {
		s.cleanupLibraryDraftObjects(ctx, actor, id, newObjects)
	}
	if err = ctx.Err(); err != nil {
		return fail(err)
	}
	if len(in.Bytes) > 0 {
		if err = s.Media.Objects.Put(ctx, object, bytes.NewReader(in.Bytes)); err != nil {
			cleanupNew()
			return fail(fmt.Errorf("图片上传失败，请保留当前内容重试"))
		}
	}
	if audio.ID != "" {
		if err = s.Media.Objects.Put(ctx, audio, bytes.NewReader(in.AudioBytes)); err != nil {
			cleanupNew()
			return fail(fmt.Errorf("参考音频上传失败，请保留内容重试"))
		}
		setLibraryAudio(&entry, audio)
	}
	entry.Name = in.Name
	entry.Description = in.Description
	entry.Category = in.Category
	if entry.Snapshot.Extra == nil {
		entry.Snapshot.Extra = map[string]string{}
	}
	entry.Snapshot.Extra[screenplayLibraryPromptKey] = in.Prompt
	entry.Snapshot.Extra[screenplayLibraryVoiceKey] = in.Voice
	entry.Snapshot.Extra[libraryReplaceKey] = in.RequestKey
	entry.Snapshot.Extra[libraryReplaceDigest] = digest
	if len(in.Bytes) > 0 {
		entry.Image = screenplay.MediaItem{ID: imageID, AssetID: imageID, Name: object.Name, Kind: "image", Status: "succeeded", CreatedAt: object.CreatedAt}
		entry.ImageObject = screenplay.AssetLibraryImageObject{ObjectKey: objectKey, ContentType: contentType, ByteSize: object.ByteSize}
	}
	if err = ctx.Err(); err != nil {
		cleanupNew()
		return fail(err)
	}
	saved, err := s.Repository.SaveAssetLibraryEntry(ctx, entry, expected)
	if err != nil {
		verify, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		current, e := s.Repository.GetAssetLibraryEntry(verify, actor, id)
		if e == nil && current.Snapshot.Extra[libraryReplaceKey] == in.RequestKey && current.Snapshot.Extra[libraryReplaceDigest] == digest {
			saved = current
			err = nil
		} else {
			cleanupNew()
			return fail(err)
		}
	}
	if old.ImageObject.ObjectKey != "" && old.ImageObject.ObjectKey != saved.ImageObject.ObjectKey {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = s.Media.Objects.Delete(cleanup, libraryImageObject(old))
	}
	if audio.ID != "" && oldAudio.ID != "" && oldAudio.ID != audio.ID {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = s.Media.Objects.Delete(cleanup, oldAudio)
	}
	return saved, nil
}

func (s *ScreenplayAssetLibrary) cleanupLibraryDraftObjects(ctx context.Context, actor, id string, objects []agentworkspace.Asset) {
	verify, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	current, err := s.Repository.GetAssetLibraryEntry(verify, actor, id)
	if err != nil && !errors.Is(err, screenplay.ErrNotFound) {
		return
	}
	audio, _ := libraryAudioObject(current)
	for _, object := range objects {
		if object.ObjectKey != "" && object.ObjectKey != current.ImageObject.ObjectKey && object.ObjectKey != audio.ObjectKey {
			_ = s.Media.Objects.Delete(verify, object)
		}
	}
}
