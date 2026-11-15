package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/screenplay"
)

const screenplayLibraryAudioLimit = 32 << 20
const libraryAudioKey = "_cm_library_reference_audio"

type libraryStoredAudio struct {
	Asset     agentworkspace.Asset `json:"asset"`
	ObjectKey string               `json:"object_key"`
}

func libraryAudioFormat(raw []byte) (string, string) {
	ct := http.DetectContentType(raw)
	switch ct {
	case "audio/wave", "audio/wav", "audio/x-wav":
		if len(raw) >= 44 {
			return "audio/wav", "wav"
		}
	case "audio/mpeg":
		return ct, "mp3"
	case "application/ogg", "audio/ogg":
		return "audio/ogg", "ogg"
	case "audio/flac", "audio/x-flac":
		return "audio/flac", "flac"
	}
	if bytes.HasPrefix(raw, []byte("fLaC")) && len(raw) >= 42 {
		return "audio/flac", "flac"
	}
	if len(raw) >= 24 && string(raw[4:8]) == "ftyp" && (string(raw[8:12]) == "M4A " || string(raw[8:12]) == "M4B ") {
		return "audio/mp4", "m4a"
	}
	return "", ""
}
func (s *ScreenplayAssetLibrary) prepareLibraryAudio(actor string, in screenplayLibraryUploadInput) (agentworkspace.Asset, error) {
	if len(in.AudioBytes) == 0 && in.AudioName == "" {
		return agentworkspace.Asset{}, nil
	}
	if in.Type != screenplay.AssetCharacter || len(in.AudioBytes) == 0 || len(in.AudioBytes) > screenplayLibraryAudioLimit {
		return agentworkspace.Asset{}, fmt.Errorf("%w: 参考音频不能为空且最大32MB", screenplay.ErrInvalidArgument)
	}
	ct, ext := libraryAudioFormat(in.AudioBytes)
	if ext == "" {
		return agentworkspace.Asset{}, fmt.Errorf("%w: 请上传MP3、WAV、M4A、OGG或FLAC参考音频", screenplay.ErrInvalidArgument)
	}
	name := path.Base(strings.ReplaceAll(in.AudioName, "\\", "/"))
	if strings.TrimSpace(name) == "" || name == "." {
		return agentworkspace.Asset{}, screenplay.ErrInvalidArgument
	}
	id := newAgentID("sp-library-audio")
	key, err := agentworkspace.AgentAssetObjectKey(actor, screenplayLibraryObjectProject, id, ext)
	if err != nil {
		return agentworkspace.Asset{}, err
	}
	return agentworkspace.Asset{ID: id, ProjectID: screenplayLibraryObjectProject, Kind: "audio", Name: name, ContentType: ct, ObjectKey: key, ByteSize: int64(len(in.AudioBytes)), CreatedAt: time.Now().UTC()}, nil
}
func setLibraryAudio(entry *screenplay.AssetLibraryEntry, audio agentworkspace.Asset) {
	if audio.ID == "" {
		return
	}
	if entry.Snapshot.Extra == nil {
		entry.Snapshot.Extra = map[string]string{}
	}
	audio.URL = ""
	raw, _ := json.Marshal(libraryStoredAudio{Asset: audio, ObjectKey: audio.ObjectKey})
	entry.Snapshot.Extra[libraryAudioKey] = string(raw)
}
func libraryAudioObject(entry screenplay.AssetLibraryEntry) (agentworkspace.Asset, error) {
	var audio agentworkspace.Asset
	var stored libraryStoredAudio
	if entry.Type != screenplay.AssetCharacter || json.Unmarshal([]byte(entry.Snapshot.Extra[libraryAudioKey]), &stored) != nil {
		return audio, screenplay.ErrNotFound
	}
	audio = stored.Asset
	audio.ObjectKey = stored.ObjectKey
	ext := map[string]string{"audio/wav": "wav", "audio/mpeg": "mp3", "audio/ogg": "ogg", "audio/flac": "flac", "audio/mp4": "m4a"}[audio.ContentType]
	expected, err := agentworkspace.AgentAssetObjectKey(entry.Owner, screenplayLibraryObjectProject, audio.ID, ext)
	if err != nil || ext == "" || audio.Kind != "audio" || audio.ProjectID != screenplayLibraryObjectProject || audio.ObjectKey != expected || audio.ByteSize < 1 || audio.ByteSize > screenplayLibraryAudioLimit {
		return agentworkspace.Asset{}, screenplay.ErrNotFound
	}
	return audio, nil
}
func (s *ScreenplayAssetLibrary) ReplaceReferenceAudio(ctx context.Context, actor, id string, expected int64, in screenplayLibraryUploadInput) (screenplay.AssetLibraryEntry, error) {
	if len(in.AudioBytes) == 0 || len(in.Bytes) != 0 {
		return screenplay.AssetLibraryEntry{}, screenplay.ErrInvalidArgument
	}
	return s.replaceLibraryEditor(ctx, actor, id, expected, in, false)
}
func registerScreenplayLibraryAudio(register func(string, func(http.ResponseWriter, *http.Request, string)), service *ScreenplayAssetLibrary) {
	base := "/api/comic-screenplays/asset-library/{entryID}/audio"
	register("GET "+base, func(w http.ResponseWriter, r *http.Request, actor string) {
		if service.Media == nil || service.Media.Objects == nil {
			writeScreenplayError(w, screenplay.ErrInvalidState)
			return
		}
		entry, err := service.Repository.GetAssetLibraryEntry(r.Context(), actor, r.PathValue("entryID"))
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		audio, err := libraryAudioObject(entry)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		reader, err := service.Media.Objects.Open(r.Context(), audio)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		defer reader.Close()
		w.Header().Set("Content-Type", audio.ContentType)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
		w.Header().Set("Content-Length", fmt.Sprint(audio.ByteSize))
		_, _ = io.Copy(w, reader)
	})
	register("PUT "+base, func(w http.ResponseWriter, r *http.Request, actor string) {
		expected, err := parseLibraryExpectedRevision(r)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		in, err := parseScreenplayLibraryEditor(w, r, false)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		entry, err := service.ReplaceReferenceAudio(r.Context(), actor, r.PathValue("entryID"), expected, in)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"entry": screenplayLibraryPublic(entry)})
	})
}
