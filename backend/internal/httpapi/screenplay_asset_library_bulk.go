package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"sort"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/characterarchive"
	"qiantie/backend/internal/screenplay"
)

type libraryBulkObject struct {
	Asset agentworkspace.Asset
	Bytes []byte
}

func libraryBulkDigest(in libraryBulkFiles) string {
	h := sha256.New()
	for _, name := range libraryBulkSortedNames(in) {
		fmt.Fprintf(h, "%d:%s%d:", len(name), name, len(in.Files[name]))
		h.Write(in.Files[name])
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
func bulkLibraryID(actor, key, digest string, index int, kind string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00%s", actor, key, digest, index, kind)))
	return fmt.Sprintf("sp-bulk-%x", sum[:16])
}

func (s *ScreenplayAssetLibrary) ImportCharacters(ctx context.Context, actor string, in libraryBulkFiles) (screenplay.AssetLibraryBulkResult, error) {
	empty := screenplay.AssetLibraryBulkResult{}
	repo, ok := s.Repository.(screenplay.AssetLibraryBulkRepository)
	if !ok {
		return empty, fmt.Errorf("%w: 整库导入存储未配置", screenplay.ErrInvalidState)
	}
	if !validAssetLibraryRequestKey(in.RequestKey) {
		return empty, screenplay.ErrInvalidArgument
	}
	records, err := in.records()
	if err != nil {
		return empty, err
	}
	digest := libraryBulkDigest(in)
	if prior, lookupErr := repo.LookupAssetLibraryBulk(ctx, actor, in.RequestKey, digest); lookupErr == nil {
		return prior, nil
	} else if !errors.Is(lookupErr, screenplay.ErrNotFound) {
		return empty, lookupErr
	}
	entries := make([]screenplay.AssetLibraryEntry, 0, len(records))
	objects := []libraryBulkObject{}
	for i, record := range records {
		values := map[string]string{}
		for _, field := range []string{"id", "name", "description", "category", "prompt", "voice", "image", "imagePath", "audio", "audioPath", "audioFileName"} {
			v, e := libraryArchiveString(record, field)
			if e != nil {
				return empty, e
			}
			values[field] = v
		}
		subject, e := libraryArchiveSubject(record)
		if e != nil {
			return empty, e
		}
		id := bulkLibraryID(actor, in.RequestKey, digest, i, "entry")
		raw, _ := json.Marshal(record)
		entry := screenplay.AssetLibraryEntry{ID: id, Owner: actor, Type: screenplay.AssetCharacter, Name: values["name"], Description: values["description"], Category: values["category"], Snapshot: screenplay.Asset{OwnerUsername: actor, Type: screenplay.AssetCharacter, Name: values["name"], Description: values["description"], Extra: map[string]string{screenplay.LibraryArchiveJSON: string(raw), screenplay.LibraryArchiveID: values["id"], screenplayLibraryPromptKey: values["prompt"], screenplayLibraryVoiceKey: values["voice"]}}}
		setLibrarySubject(&entry, subject)
		if entry.Name == "" {
			return empty, libraryBulkError("每个角色必须填写名称")
		}
		imageBytes, e := in.resource(firstText(values["imagePath"], values["image"]), screenplayLibraryImageLimit)
		if e != nil {
			return empty, e
		}
		if len(imageBytes) > 0 {
			ct, ext, e := libraryBulkImageFormat(imageBytes)
			if e != nil {
				return empty, e
			}
			imageID := bulkLibraryID(actor, in.RequestKey, digest, i, "image")
			key, e := agentworkspace.AgentAssetObjectKey(actor, screenplayLibraryObjectProject, imageID, ext)
			if e != nil {
				return empty, e
			}
			object := agentworkspace.Asset{ID: imageID, ProjectID: screenplayLibraryObjectProject, Kind: "image", Name: imageID + "." + ext, ContentType: ct, ObjectKey: key, ByteSize: int64(len(imageBytes)), CreatedAt: time.Unix(0, 0).UTC()}
			entry.Image = screenplay.MediaItem{ID: imageID, AssetID: imageID, Name: object.Name, Kind: "image", Status: "succeeded", CreatedAt: object.CreatedAt}
			entry.ImageObject = screenplay.AssetLibraryImageObject{ObjectKey: key, ContentType: ct, ByteSize: object.ByteSize}
			objects = append(objects, libraryBulkObject{object, imageBytes})
		} else {
			entry.Snapshot.Extra[screenplay.LibraryMetadataOnly] = "true"
		}
		audioBytes, e := in.resource(firstText(values["audioPath"], values["audio"]), screenplayLibraryAudioLimit)
		if e != nil {
			return empty, e
		}
		if len(audioBytes) > 0 {
			ct, ext := libraryAudioFormat(audioBytes)
			if ext == "" {
				return empty, libraryBulkError("档案参考音频无效")
			}
			audioID := bulkLibraryID(actor, in.RequestKey, digest, i, "audio")
			key, e := agentworkspace.AgentAssetObjectKey(actor, screenplayLibraryObjectProject, audioID, ext)
			if e != nil {
				return empty, e
			}
			name := path.Base(values["audioFileName"])
			if name == "." || name == "" {
				name = "reference." + ext
			}
			audio := agentworkspace.Asset{ID: audioID, ProjectID: screenplayLibraryObjectProject, Kind: "audio", Name: name, ContentType: ct, ObjectKey: key, ByteSize: int64(len(audioBytes)), CreatedAt: time.Unix(0, 0).UTC()}
			setLibraryAudio(&entry, audio)
			objects = append(objects, libraryBulkObject{audio, audioBytes})
		}
		entries = append(entries, entry)
	}
	// Validate every record before any object write. Repository repeats validation
	// inside the transaction after resolving logical ID conflicts.
	if err := screenplay.ValidateAssetLibraryBulkEntries(entries); err != nil {
		return empty, libraryBulkError("档案角色名称、分类或资源资料无效，名称与分类最多191字")
	}
	if len(objects) > 0 && (s.Media == nil || s.Media.Objects == nil) {
		return empty, fmt.Errorf("%w: 档案资源存储未配置", screenplay.ErrInvalidState)
	}
	if err = repo.ReserveAssetLibraryBulk(ctx, actor, in.RequestKey, digest); err != nil {
		return empty, err
	}
	for _, object := range objects {
		if err = ctx.Err(); err != nil {
			return empty, err
		}
		if err = s.Media.Objects.Put(ctx, object.Asset, bytes.NewReader(object.Bytes)); err != nil {
			return empty, fmt.Errorf("档案资源保存失败，请保留同一请求编号重试")
		}
	}
	// Deterministic, digest-scoped object keys survive unknown commit outcomes and
	// failed DB writes. Retries reuse these bytes rather than creating duplicates.
	return repo.ApplyAssetLibraryBulk(ctx, actor, in.RequestKey, digest, "import", entries, nil)
}

func (s *ScreenplayAssetLibrary) ExportCharacters(ctx context.Context, actor string) ([]byte, error) {
	entries, err := s.Repository.ListAssetLibrary(ctx, actor, screenplay.AssetCharacter, "", "")
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, libraryBulkError("角色库为空，没有可导出的档案")
	}
	if len(entries) > libraryBulkCharacterLimit {
		return nil, libraryBulkError(fmt.Sprintf("角色库超过%d个角色，无法导出可重新导入的完整档案，请先减少角色数量", libraryBulkCharacterLimit))
	}
	records := make([]map[string]json.RawMessage, 0, len(entries))
	total := 0
	for _, entry := range entries {
		record := map[string]json.RawMessage{}
		if raw := entry.Snapshot.Extra[screenplay.LibraryArchiveJSON]; raw != "" {
			if json.Unmarshal([]byte(raw), &record) != nil || record == nil {
				return nil, fmt.Errorf("%w: 已存档案元数据损坏", screenplay.ErrInvalidState)
			}
		}
		set := func(key string, value any) { raw, _ := json.Marshal(value); record[key] = raw }
		id := entry.Snapshot.Extra[screenplay.LibraryArchiveID]
		if id == "" {
			id = entry.ID
		}
		set("id", id)
		set("type", "character")
		set("name", entry.Name)
		set("description", entry.Description)
		set("category", entry.Category)
		set("prompt", entry.Snapshot.Extra[screenplayLibraryPromptKey])
		set("voice", entry.Snapshot.Extra[screenplayLibraryVoiceKey])
		set("sora2Info", librarySubject(entry))
		exportLibrarySubjectReference(record, entry)
		if _, ok := record["sora2_info"]; ok {
			set("sora2_info", librarySubject(entry))
		}
		set("image", "")
		set("imagePath", "")
		set("audio", "")
		set("audioPath", "")
		set("audioFileName", "")
		if entry.ImageObject.ObjectKey != "" {
			value, e := s.libraryBulkDataURI(ctx, libraryImageObject(entry), screenplayLibraryImageLimit)
			if e != nil {
				return nil, e
			}
			set("image", value)
			total += len(value)
		}
		if audio, e := libraryAudioObject(entry); e == nil {
			value, e := s.libraryBulkDataURI(ctx, audio, screenplayLibraryAudioLimit)
			if e != nil {
				return nil, e
			}
			set("audio", value)
			set("audioFileName", audio.Name)
			total += len(value)
		}
		if total > int(characterarchive.DefaultMaxInputBytes) {
			return nil, libraryBulkError("导出内容超过128MB，请减少角色或资源体积")
		}
		records = append(records, record)
	}
	raw, err := json.Marshal(records)
	if err != nil {
		return nil, err
	}
	encoded, err := characterarchive.EncodeJSON(raw)
	if err != nil {
		return nil, libraryBulkError("导出内容超过档案大小限制")
	}
	if int64(len(encoded)) > characterarchive.DefaultMaxInputBytes {
		return nil, libraryBulkError("导出文件超过128MB")
	}
	return encoded, nil
}
func (s *ScreenplayAssetLibrary) libraryBulkDataURI(ctx context.Context, object agentworkspace.Asset, limit int) (string, error) {
	if s.Media == nil || s.Media.Objects == nil {
		return "", screenplay.ErrInvalidState
	}
	reader, err := s.Media.Objects.Open(ctx, object)
	if err != nil {
		return "", fmt.Errorf("档案资源读取失败，请重试")
	}
	raw, e := io.ReadAll(io.LimitReader(reader, int64(limit)+1))
	closeErr := reader.Close()
	if e != nil || closeErr != nil || len(raw) == 0 || len(raw) > limit {
		return "", fmt.Errorf("档案资源读取失败或超过限制")
	}
	return "data:" + object.ContentType + ";base64," + base64.StdEncoding.EncodeToString(raw), nil
}

type libraryClearInput struct {
	RequestKey string                            `json:"requestKey"`
	Expected   []screenplay.AssetLibraryExpected `json:"expected"`
}

func (s *ScreenplayAssetLibrary) ClearCharacters(ctx context.Context, actor string, in libraryClearInput) (screenplay.AssetLibraryBulkResult, error) {
	repo, ok := s.Repository.(screenplay.AssetLibraryBulkRepository)
	if !ok {
		return screenplay.AssetLibraryBulkResult{}, fmt.Errorf("%w: 整库清空存储未配置", screenplay.ErrInvalidState)
	}
	if !validAssetLibraryRequestKey(in.RequestKey) || in.Expected == nil {
		return screenplay.AssetLibraryBulkResult{}, screenplay.ErrInvalidArgument
	}
	expected := append([]screenplay.AssetLibraryExpected{}, in.Expected...)
	sort.Slice(expected, func(i, j int) bool { return expected[i].ID < expected[j].ID })
	raw, _ := json.Marshal(expected)
	digest := fmt.Sprintf("%x", sha256.Sum256(append([]byte("clear\x00"), raw...)))
	if err := repo.ReserveAssetLibraryBulk(ctx, actor, in.RequestKey, digest); err != nil {
		return screenplay.AssetLibraryBulkResult{}, err
	}
	return repo.ApplyAssetLibraryBulk(ctx, actor, in.RequestKey, digest, "clear", nil, expected)
}
func registerScreenplayLibraryBulk(register func(string, func(http.ResponseWriter, *http.Request, string)), service *ScreenplayAssetLibrary) {
	base := "/api/comic-screenplays/asset-library/characters"
	register("GET "+base+"/export", func(w http.ResponseWriter, r *http.Request, actor string) {
		raw, err := service.ExportCharacters(r.Context(), actor)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="characters.lschar"`)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(200)
		_, _ = w.Write(raw)
	})
	register("POST "+base+"/import", func(w http.ResponseWriter, r *http.Request, actor string) {
		in, err := parseLibraryBulkFiles(w, r)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		result, err := service.ImportCharacters(r.Context(), actor, in)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		entries := []map[string]any{}
		for _, e := range result.Entries {
			entries = append(entries, screenplayLibraryPublic(e))
		}
		writeJSON(w, 200, map[string]any{"imported_count": len(entries), "entries": entries})
	})
	register("POST "+base+"/clear", func(w http.ResponseWriter, r *http.Request, actor string) {
		var in libraryClearInput
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<20))
		decoder.DisallowUnknownFields()
		var trailing any
		if decoder.Decode(&in) != nil || decoder.Decode(&trailing) != io.EOF {
			writeScreenplayError(w, libraryBulkError("清空确认资料无效或超过16MB，请重新读取完整角色库"))
			return
		}
		result, err := service.ClearCharacters(r.Context(), actor, in)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"cleared": true, "cleared_count": result.ClearedCount})
	})
}
