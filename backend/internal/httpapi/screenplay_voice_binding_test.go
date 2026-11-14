package httpapi

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"qiantie/backend/internal/screenplay"
	"strings"
	"testing"
	"time"
)

func testVoiceBindingWAV() []byte {
	raw := make([]byte, 44+480)
	copy(raw, "RIFF")
	binary.LittleEndian.PutUint32(raw[4:], uint32(len(raw)-8))
	copy(raw[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(raw[16:], 16)
	binary.LittleEndian.PutUint16(raw[20:], 1)
	binary.LittleEndian.PutUint16(raw[22:], 1)
	binary.LittleEndian.PutUint32(raw[24:], 24000)
	binary.LittleEndian.PutUint32(raw[28:], 48000)
	binary.LittleEndian.PutUint16(raw[32:], 2)
	binary.LittleEndian.PutUint16(raw[34:], 16)
	copy(raw[36:], "data")
	binary.LittleEndian.PutUint32(raw[40:], 480)
	return raw
}
func voiceUploadRequest(t *testing.T, mux http.Handler, now time.Time, actor, endpoint, target string, revision int64, filename string, data []byte) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	multi := multipart.NewWriter(&body)
	_ = multi.WriteField("targetId", target)
	_ = multi.WriteField("expectedRevision", fmt.Sprint(revision))
	_ = multi.WriteField("voiceId", "forged-id-must-be-ignored")
	file, e := multi.CreateFormFile("file", filename)
	if e != nil {
		t.Fatal(e)
	}
	_, _ = file.Write(data)
	_ = multi.Close()
	request := httptest.NewRequest("POST", endpoint, &body)
	request.Header.Set("Content-Type", multi.FormDataContentType())
	SignBridgeRequest(request, actor, false, now, "secret")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func TestScreenplayVoiceUploadPersistsReferenceBindingWithoutForgedVoiceID(t *testing.T) {
	s, p := positioningFixture(t)
	ctx := context.Background()
	_, e := s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(st *screenplay.MediaState) error {
		st.VoiceDesign = &screenplay.VoiceDesignState{Items: []screenplay.VoiceDesignDraft{{ID: 1, Name: "旁白"}}}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	mux := http.NewServeMux()
	now := time.Now()
	auth := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}
	RegisterScreenplayVoiceBindingRoutes(mux, auth, s)
	RegisterScreenplayMediaWorkbenchRoutes(mux, auth, s)
	st, _ := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	response := voiceUploadRequest(t, mux, now, "alice", "/api/comic-screenplays/projects/"+p.ID+"/media/voice-upload", "voice-draft:1", st.Revision, "reference.wav", testVoiceBindingWAV())
	if response.Code != 201 {
		t.Fatal(response.Code, response.Body.String())
	}
	var out struct {
		Item    screenplay.MediaItem    `json:"item"`
		Binding screenplay.VoiceBinding `json:"binding"`
	}
	if e = json.Unmarshal(response.Body.Bytes(), &out); e != nil {
		t.Fatal(e)
	}
	if out.Binding.VoiceID != "" || out.Binding.Engine != "custom" || !out.Binding.ReferenceOnly || out.Item.VoiceID != "" || !strings.Contains(out.Item.URL, "access=") {
		t.Fatal(out)
	}
	fresh, _ := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if fresh.VoiceBindings["voice-draft:1"].MediaItemID != out.Item.ID || fresh.Voices["voice-draft:1"] != "" {
		t.Fatal(fresh)
	}
	asset, e := s.Cut.Asset(ctx, p.OwnerUsername, p.ID, out.Item.AssetID)
	if e != nil {
		t.Fatal(e)
	}
	reader, e := s.Objects.Open(ctx, aicutObjectAsset(asset))
	if e != nil {
		t.Fatal(e)
	}
	saved, _ := io.ReadAll(reader)
	reader.Close()
	if !bytes.Equal(saved, testVoiceBindingWAV()) {
		t.Fatal("actual WAV not persisted")
	}
	stale := voiceUploadRequest(t, mux, now, "alice", "/api/comic-screenplays/projects/"+p.ID+"/media/voice-upload", "voice-draft:1", st.Revision, "reference.wav", testVoiceBindingWAV())
	if stale.Code != 409 {
		t.Fatal(stale.Code, stale.Body.String())
	}
	bad := voiceUploadRequest(t, mux, now, "alice", "/api/comic-screenplays/projects/"+p.ID+"/media/voice-upload", "voice-draft:1", fresh.Revision, "fake.mp3", testImageToolPNG(t))
	if bad.Code != 400 {
		t.Fatal(bad.Code, bad.Body.String())
	}
	foreign := voiceUploadRequest(t, mux, now, "bob", "/api/comic-screenplays/projects/"+p.ID+"/media/voice-upload", "voice-draft:1", fresh.Revision, "reference.wav", testVoiceBindingWAV())
	if foreign.Code == 201 {
		t.Fatal("cross-owner upload accepted")
	}
}

func TestScreenplayVoiceLibraryOwnRealIDsAndCrossProjectBindingCopiesBytes(t *testing.T) {
	s, source := positioningFixture(t)
	ctx := context.Background()
	audio, e := s.saveBytes(ctx, source, "audio", "已设计音色.wav", "audio/wav", bytes.NewReader(testVoiceBindingWAV()), "")
	if e != nil {
		t.Fatal(e)
	}
	_, e = s.Media.MutateMedia(ctx, source.OwnerUsername, source.ID, func(st *screenplay.MediaState) error {
		for i := range st.Items {
			if st.Items[i].ID == audio.ID {
				st.Items[i].VoiceID = "actual-generated-id"
				st.Items[i].VoiceEngine = "minimax"
				st.Items[i].VoiceTargetModel = "speech-real"
			}
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	target, e := s.Projects.CreateProject(ctx, screenplay.Project{ID: "voice-target", OwnerUsername: "alice"})
	if e != nil {
		t.Fatal(e)
	}
	_, e = s.Projects.CreateAsset(ctx, screenplay.Asset{ID: "target-hero", ProjectID: target.ID, OwnerUsername: "alice", Name: "甲", Type: screenplay.AssetCharacter})
	if e != nil {
		t.Fatal(e)
	}
	mux := http.NewServeMux()
	now := time.Now()
	auth := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}
	RegisterScreenplayVoiceBindingRoutes(mux, auth, s)
	RegisterScreenplayMediaWorkbenchRoutes(mux, auth, s)
	own := signedJSONRequest(t, mux, now, "alice", "GET", "/api/screenplays/voice-library", nil)
	if own.Code != 200 || !strings.Contains(own.Body.String(), "actual-generated-id") || !strings.Contains(own.Body.String(), "access=") {
		t.Fatal(own.Code, own.Body.String())
	}
	other := signedJSONRequest(t, mux, now, "bob", "GET", "/api/screenplays/voice-library", nil)
	if other.Code != 200 || strings.Contains(other.Body.String(), "actual-generated-id") {
		t.Fatal(other.Code, other.Body.String())
	}
	response := signedJSONRequest(t, mux, now, "alice", "POST", "/api/comic-screenplays/projects/"+target.ID+"/media/actions", map[string]any{"action": "voice_bind", "targetId": "target-hero", "itemId": audio.ID, "sourceProjectId": source.ID, "expectedRevision": 0})
	if response.Code != 200 {
		t.Fatal(response.Code, response.Body.String())
	}
	state, _ := s.Media.LoadMedia(ctx, "alice", target.ID)
	binding := state.VoiceBindings["target-hero"]
	if binding.VoiceID != "actual-generated-id" || binding.Engine != "minimax" || binding.ReferenceOnly || binding.MediaItemID == audio.ID || state.Voices["target-hero"] != "" {
		t.Fatal(binding, state.Voices)
	}
	item, e := mediaItem(state, binding.MediaItemID)
	if e != nil {
		t.Fatal(e)
	}
	asset, e := s.Cut.Asset(ctx, "alice", target.ID, item.AssetID)
	if e != nil {
		t.Fatal(e)
	}
	reader, e := s.Objects.Open(ctx, aicutObjectAsset(asset))
	if e != nil {
		t.Fatal(e)
	}
	copied, _ := io.ReadAll(reader)
	reader.Close()
	if !bytes.Equal(copied, testVoiceBindingWAV()) {
		t.Fatal("cross-project copy lost bytes")
	}
}
