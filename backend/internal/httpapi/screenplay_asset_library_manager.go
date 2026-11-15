package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/screenplay"
)

const screenplayLibraryPromptKey = "_cm_library_prompt"
const screenplayLibraryVoiceKey = "_cm_library_voice"

type screenplayLibraryUploadInput struct {
	RequestKey                                           string
	Type                                                 screenplay.AssetType
	Name, Description, Category, Prompt, Voice, FileName string
	Bytes                                                []byte
	Sora2Info                                            *string
	AudioName                                            string
	AudioBytes                                           []byte
}
type screenplayLibraryMetadataInput struct {
	Sora2Info   *string `json:"sora2Info"`
	Expected    *int64  `json:"expectedRevision"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	Prompt      *string `json:"prompt"`
	Voice       *string `json:"voice"`
}

func ordinaryLibraryType(kind screenplay.AssetType) bool {
	return kind == screenplay.AssetCharacter || kind == screenplay.AssetScene || kind == screenplay.AssetProp
}
func parseScreenplayLibraryUpload(w http.ResponseWriter, r *http.Request) (screenplayLibraryUploadInput, error) {
	return parseScreenplayLibraryEditor(w, r, true)
}
func parseScreenplayLibraryEditor(w http.ResponseWriter, r *http.Request, requireImage bool) (screenplayLibraryUploadInput, error) {
	var in screenplayLibraryUploadInput
	r.Body = http.MaxBytesReader(w, r.Body, screenplayLibraryImageLimit+screenplayLibraryAudioLimit+(128<<10))
	reader, err := r.MultipartReader()
	if err != nil {
		return in, screenplay.ErrInvalidArgument
	}
	fields := map[string]string{}
	hasFile := false
	hasAudio := false
	for {
		part, e := reader.NextPart()
		if e == io.EOF {
			break
		}
		if e != nil {
			return in, screenplay.ErrInvalidArgument
		}
		name := part.FormName()
		if name == "audio" {
			if hasAudio {
				part.Close()
				return in, screenplay.ErrInvalidArgument
			}
			hasAudio = true
			in.AudioName = path.Base(strings.ReplaceAll(part.FileName(), "\\", "/"))
			in.AudioBytes, e = io.ReadAll(io.LimitReader(part, screenplayLibraryAudioLimit+1))
			part.Close()
			if e != nil || len(in.AudioBytes) == 0 || len(in.AudioBytes) > screenplayLibraryAudioLimit {
				return in, fmt.Errorf("%w: 参考音频不能为空且最大32MB", screenplay.ErrInvalidArgument)
			}
			continue
		}
		if name == "file" {
			if hasFile {
				part.Close()
				return in, fmt.Errorf("%w: 每次上传一张图片", screenplay.ErrInvalidArgument)
			}
			hasFile = true
			in.FileName = path.Base(strings.ReplaceAll(part.FileName(), "\\", "/"))
			in.Bytes, e = io.ReadAll(io.LimitReader(part, screenplayLibraryImageLimit+1))
			part.Close()
			if e != nil || len(in.Bytes) == 0 || len(in.Bytes) > screenplayLibraryImageLimit {
				return in, fmt.Errorf("%w: 图片不能为空且最大40MB", screenplay.ErrInvalidArgument)
			}
			continue
		}
		switch name {
		case "requestKey", "type", "name", "description", "category", "prompt", "voice", "sora2Info":
		default:
			part.Close()
			return in, screenplay.ErrInvalidArgument
		}
		if _, duplicate := fields[name]; duplicate {
			part.Close()
			return in, screenplay.ErrInvalidArgument
		}
		value, e := io.ReadAll(io.LimitReader(part, (16<<10)+1))
		part.Close()
		if e != nil || len(value) > 16<<10 {
			return in, screenplay.ErrInvalidArgument
		}
		fields[name] = string(value)
	}
	if requireImage && !hasFile || !requireImage && (!hasAudio || hasFile) {
		return in, screenplay.ErrInvalidArgument
	}
	in.RequestKey = fields["requestKey"]
	in.Type = screenplay.AssetType(fields["type"])
	in.Name = fields["name"]
	in.Description = fields["description"]
	in.Category = fields["category"]
	in.Prompt = fields["prompt"]
	in.Voice = fields["voice"]
	if value, ok := fields["sora2Info"]; ok {
		in.Sora2Info = &value
	}
	return in, nil
}
func parseLibraryExpectedRevision(r *http.Request) (int64, error) {
	value, err := strconv.ParseInt(r.URL.Query().Get("expectedRevision"), 10, 64)
	if err != nil || value < 1 {
		return 0, screenplay.ErrInvalidArgument
	}
	return value, nil
}
func (s *ScreenplayAssetLibrary) UploadEntry(ctx context.Context, actor string, in screenplayLibraryUploadInput) (screenplay.AssetLibraryEntry, error) {
	if s == nil || s.Repository == nil || s.Media == nil || s.Media.Objects == nil {
		return screenplay.AssetLibraryEntry{}, screenplay.ErrInvalidState
	}
	if err := ctx.Err(); err != nil {
		return screenplay.AssetLibraryEntry{}, err
	}
	if !ordinaryLibraryType(in.Type) || !validAssetLibraryRequestKey(in.RequestKey) || len(in.Bytes) == 0 || len(in.Bytes) > screenplayLibraryImageLimit {
		return screenplay.AssetLibraryEntry{}, screenplay.ErrInvalidArgument
	}
	if err := validateLibrarySubject(in.Type, in.Sora2Info); err != nil {
		return screenplay.AssetLibraryEntry{}, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" && in.Type != screenplay.AssetCharacter {
		name = strings.TrimSpace(in.FileName)
	}
	if name == "" || name == "." {
		return screenplay.AssetLibraryEntry{}, fmt.Errorf("%w: 请填写名称", screenplay.ErrInvalidArgument)
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(in.Bytes))
	if err != nil || cfg.Width < 1 || cfg.Height < 1 {
		return screenplay.AssetLibraryEntry{}, fmt.Errorf("%w: 请上传有效的PNG、JPEG、WEBP或GIF图片", screenplay.ErrInvalidArgument)
	}
	extension := map[string]string{"png": "png", "jpeg": "jpg", "webp": "webp", "gif": "gif"}[format]
	contentType := map[string]string{"png": "image/png", "jpeg": "image/jpeg", "webp": "image/webp", "gif": "image/gif"}[format]
	if extension == "" {
		return screenplay.AssetLibraryEntry{}, screenplay.ErrInvalidArgument
	}
	audio, err := s.prepareLibraryAudio(actor, in)
	if err != nil {
		return screenplay.AssetLibraryEntry{}, err
	}
	rawHash := sha256.Sum256(in.Bytes)
	metadata, _ := json.Marshal([]string{string(in.Type), name, in.Description, in.Category, in.Prompt, in.Voice, fmt.Sprintf("%x", rawHash)})
	if len(in.AudioBytes) > 0 {
		sum := sha256.Sum256(in.AudioBytes)
		metadata = append(metadata, []byte(in.AudioName+fmt.Sprintf("%x", sum))...)
	}
	if in.Sora2Info != nil {
		subject, _ := json.Marshal([]string{"sora2Info", *in.Sora2Info})
		metadata = append(metadata, subject...)
	}
	payloadHash := sha256.Sum256(metadata)
	digest := fmt.Sprintf("%x", payloadHash)
	keyHash := sha256.Sum256([]byte(actor + "\x00" + in.RequestKey))
	id := fmt.Sprintf("sp-lib-upload-%x", keyHash[:16])
	previous, err := s.Repository.GetAssetLibraryEntry(ctx, actor, id)
	if err == nil {
		if previous.Snapshot.Extra["library_upload_digest"] != digest {
			return screenplay.AssetLibraryEntry{}, screenplay.ErrRevisionConflict
		}
		return previous, nil
	}
	if !errors.Is(err, screenplay.ErrNotFound) {
		return screenplay.AssetLibraryEntry{}, err
	}
	imageID := newAgentID("sp-library-image")
	objectKey, err := agentworkspace.AgentAssetObjectKey(actor, screenplayLibraryObjectProject, imageID, extension)
	if err != nil {
		return screenplay.AssetLibraryEntry{}, err
	}
	object := agentworkspace.Asset{ID: imageID, ProjectID: screenplayLibraryObjectProject, Kind: "image", Name: imageID + "." + extension, ContentType: contentType, ObjectKey: objectKey, ByteSize: int64(len(in.Bytes)), CreatedAt: time.Now().UTC()}
	if err = ctx.Err(); err != nil {
		return screenplay.AssetLibraryEntry{}, err
	}
	if err = s.Media.Objects.Put(ctx, object, bytes.NewReader(in.Bytes)); err != nil {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = s.Media.Objects.Delete(cleanup, object)
		return screenplay.AssetLibraryEntry{}, fmt.Errorf("图片上传失败，请重试")
	}
	entry := screenplay.AssetLibraryEntry{ID: id, Owner: actor, Type: in.Type, Name: name, Description: in.Description, Category: in.Category, Image: screenplay.MediaItem{ID: imageID, AssetID: imageID, Name: object.Name, Kind: "image", Status: "succeeded", CreatedAt: object.CreatedAt}, ImageObject: screenplay.AssetLibraryImageObject{ObjectKey: objectKey, ContentType: contentType, ByteSize: object.ByteSize}, Snapshot: screenplay.Asset{OwnerUsername: actor, Type: in.Type, Name: name, Description: in.Description, Prompt: in.Prompt, Voice: in.Voice, Extra: map[string]string{"library_upload_digest": digest, screenplayLibraryPromptKey: in.Prompt, screenplayLibraryVoiceKey: in.Voice}}}
	setLibrarySubject(&entry, in.Sora2Info)
	if audio.ID != "" {
		if err = s.Media.Objects.Put(ctx, audio, bytes.NewReader(in.AudioBytes)); err != nil {
			s.cleanupLibraryDraftObjects(ctx, actor, id, []agentworkspace.Asset{object, audio})
			return screenplay.AssetLibraryEntry{}, fmt.Errorf("参考音频上传失败，请保留内容重试")
		}
		setLibraryAudio(&entry, audio)
	}
	if err = ctx.Err(); err != nil {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = s.Media.Objects.Delete(cleanup, object)
		if audio.ID != "" {
			_ = s.Media.Objects.Delete(cleanup, audio)
		}
		return screenplay.AssetLibraryEntry{}, err
	}
	saved, err := s.Repository.SaveAssetLibraryEntry(ctx, entry, 0)
	if err == nil {
		return saved, nil
	}
	verify, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	winner, readErr := s.Repository.GetAssetLibraryEntry(verify, actor, id)
	if errors.Is(readErr, screenplay.ErrNotFound) || readErr == nil && winner.ImageObject.ObjectKey != objectKey {
		_ = s.Media.Objects.Delete(verify, object)
		if audio.ID != "" {
			_ = s.Media.Objects.Delete(verify, audio)
		}
	}
	if readErr == nil && winner.Snapshot.Extra["library_upload_digest"] == digest {
		return winner, nil
	}
	if readErr != nil && !errors.Is(readErr, screenplay.ErrNotFound) {
		return screenplay.AssetLibraryEntry{}, fmt.Errorf("上传保存结果待核对，请使用同一请求标识重试")
	}
	return screenplay.AssetLibraryEntry{}, err
}
func (s *ScreenplayAssetLibrary) UpdateEntry(ctx context.Context, actor, id string, in screenplayLibraryMetadataInput) (screenplay.AssetLibraryEntry, error) {
	if err := ctx.Err(); err != nil {
		return screenplay.AssetLibraryEntry{}, err
	}
	if in.Expected == nil || *in.Expected < 1 || in.Name == nil && in.Description == nil && in.Category == nil && in.Prompt == nil && in.Voice == nil && in.Sora2Info == nil {
		return screenplay.AssetLibraryEntry{}, screenplay.ErrInvalidArgument
	}
	entry, err := s.Repository.GetAssetLibraryEntry(ctx, actor, id)
	if err != nil {
		return screenplay.AssetLibraryEntry{}, err
	}
	if err := validateLibrarySubject(entry.Type, in.Sora2Info); err != nil {
		return screenplay.AssetLibraryEntry{}, err
	}
	setLibrarySubject(&entry, in.Sora2Info)
	if in.Name != nil {
		entry.Name = *in.Name
	}
	if in.Description != nil {
		entry.Description = *in.Description
	}
	if in.Category != nil {
		entry.Category = *in.Category
	}
	if in.Prompt != nil {
		if entry.Snapshot.Extra == nil {
			entry.Snapshot.Extra = map[string]string{}
		}
		entry.Snapshot.Extra[screenplayLibraryPromptKey] = *in.Prompt
	}
	if in.Voice != nil {
		if entry.Snapshot.Extra == nil {
			entry.Snapshot.Extra = map[string]string{}
		}
		entry.Snapshot.Extra[screenplayLibraryVoiceKey] = *in.Voice
	}
	if err := ctx.Err(); err != nil {
		return screenplay.AssetLibraryEntry{}, err
	}
	saved, err := s.Repository.SaveAssetLibraryEntry(ctx, entry, *in.Expected)
	if err == nil {
		return saved, nil
	}
	verify, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	current, readErr := s.Repository.GetAssetLibraryEntry(verify, actor, id)
	if readErr == nil && sameLibraryMetadata(current, entry) {
		return current, nil
	}
	return screenplay.AssetLibraryEntry{}, err
}
func sameLibraryMetadata(a, b screenplay.AssetLibraryEntry) bool {
	return librarySubject(a) == librarySubject(b) && a.Name == strings.TrimSpace(b.Name) && a.Description == b.Description && a.Category == b.Category && a.Snapshot.Extra[screenplayLibraryPromptKey] == b.Snapshot.Extra[screenplayLibraryPromptKey] && a.Snapshot.Extra[screenplayLibraryVoiceKey] == b.Snapshot.Extra[screenplayLibraryVoiceKey]
}
func (s *ScreenplayAssetLibrary) DeleteEntry(ctx context.Context, actor, id string, expected int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if expected < 1 {
		return screenplay.ErrInvalidArgument
	}
	entry, err := s.Repository.GetAssetLibraryEntry(ctx, actor, id)
	if errors.Is(err, screenplay.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	err = s.Repository.DeleteAssetLibraryEntry(ctx, actor, id, expected)
	if err != nil {
		verify, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_, readErr := s.Repository.GetAssetLibraryEntry(verify, actor, id)
		if !errors.Is(readErr, screenplay.ErrNotFound) {
			return err
		}
	}
	// Project staging uses independent copies. Only remove this deleted entry's
	// private original after its durable row is known absent; uncertain DB state
	// always retains bytes. A failed object cleanup does not undo record deletion.
	if entry.ImageObject.ObjectKey != "" && s.Media != nil && s.Media.Objects != nil {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = s.Media.Objects.Delete(cleanup, agentworkspace.Asset{ID: entry.Image.AssetID, ProjectID: screenplayLibraryObjectProject, Kind: "image", Name: entry.Image.Name, ObjectKey: entry.ImageObject.ObjectKey, ContentType: entry.ImageObject.ContentType, ByteSize: entry.ImageObject.ByteSize})
		if audio, err := libraryAudioObject(entry); err == nil {
			_ = s.Media.Objects.Delete(cleanup, audio)
		}
	}
	return nil
}

// These endpoints are CM persistence adapters. Native file-copy, deletion and
// duplicate-ID rules remain unrecovered; they are not original parity claims.
func registerScreenplayLibraryManager(register func(string, func(http.ResponseWriter, *http.Request, string)), service *ScreenplayAssetLibrary) {
	registerScreenplayLibraryCategories(register, service)
	base := "/api/comic-screenplays/asset-library"
	registerScreenplayLibraryAudio(register, service)
	register("PUT "+base+"/{entryID}/image", func(w http.ResponseWriter, r *http.Request, actor string) {
		expected, err := strconv.ParseInt(r.URL.Query().Get("expectedRevision"), 10, 64)
		if err != nil || expected < 1 {
			writeScreenplayError(w, screenplay.ErrInvalidArgument)
			return
		}
		in, err := parseScreenplayLibraryUpload(w, r)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		entry, err := service.ReplaceImage(r.Context(), actor, r.PathValue("entryID"), expected, in)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"entry": screenplayLibraryPublic(entry)})
	})
	register("POST "+base, func(w http.ResponseWriter, r *http.Request, actor string) {
		in, err := parseScreenplayLibraryUpload(w, r)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		entry, err := service.UploadEntry(r.Context(), actor, in)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, 201, map[string]any{"entry": screenplayLibraryPublic(entry)})
	})
	register("PATCH "+base+"/{entryID}", func(w http.ResponseWriter, r *http.Request, actor string) {
		var in screenplayLibraryMetadataInput
		if !decodeAssetCheckRequest(w, r, &in, false) {
			return
		}
		entry, err := service.UpdateEntry(r.Context(), actor, r.PathValue("entryID"), in)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"entry": screenplayLibraryPublic(entry)})
	})
	register("DELETE "+base+"/{entryID}", func(w http.ResponseWriter, r *http.Request, actor string) {
		revision, err := strconv.ParseInt(r.URL.Query().Get("expectedRevision"), 10, 64)
		if err != nil || revision < 1 {
			writeScreenplayError(w, screenplay.ErrInvalidArgument)
			return
		}
		if err = service.DeleteEntry(r.Context(), actor, r.PathValue("entryID"), revision); err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"deleted": true})
	})
}
