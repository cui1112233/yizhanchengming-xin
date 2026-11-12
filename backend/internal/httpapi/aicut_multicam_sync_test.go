package httpapi

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/ffmpegbin"
	"testing"
	"time"
)

func syncFixtureEnvelope(n int) []float64 {
	x := make([]float64, n)
	var state uint32 = 37
	for i := range x {
		state = state*1664525 + 1013904223
		x[i] = float64(state%1000)/1000 + .1
	}
	return x
}
func TestAICutMultiCamCorrelationSignedOffsets(t *testing.T) {
	ref := syncFixtureEnvelope(600)
	for _, shift := range []int{-35, 0, 48} {
		camera := make([]float64, len(ref))
		for i := range camera {
			j := i - shift
			if j >= 0 && j < len(ref) {
				camera[i] = ref[j]*.4 + .02
			}
		}
		offset, confidence, err := correlateAICutMultiCam(t.Context(), ref, camera, 2000)
		if err != nil || offset != shift*20 || confidence < .99 {
			t.Fatalf("shift=%d got=%d confidence=%f err=%v", shift, offset, confidence, err)
		}
	}
}
func TestAICutMultiCamRejectsSilenceWeakAmbiguousAndCancellation(t *testing.T) {
	ref := syncFixtureEnvelope(500)
	if _, _, err := correlateAICutMultiCam(t.Context(), ref, make([]float64, 500), 2000); err == nil {
		t.Fatal("silence accepted")
	}
	other := make([]float64, 500)
	for i := range other {
		other[i] = math.Sin(float64(i)*1.31) + 2
	}
	if _, _, err := correlateAICutMultiCam(t.Context(), ref, other, 2000); err == nil {
		t.Fatal("unrelated audio accepted")
	}
	periodic := make([]float64, 500)
	for i := range periodic {
		periodic[i] = float64(i % 10)
	}
	if _, _, err := correlateAICutMultiCam(t.Context(), periodic, periodic, 2000); err == nil {
		t.Fatal("ambiguous periodic audio accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, _, err := correlateAICutMultiCam(ctx, ref, ref, 2000); err == nil {
		t.Fatal("cancel ignored")
	}
}

func TestAICutMultiCamSyncRouteValidatesOwnershipAndInputs(t *testing.T) {
	now := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "sync-project", Title: "sync", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"a", "b"} {
		if err := store.CreateAsset(t.Context(), "alice", aicut.Asset{ID: id, ProjectID: "sync-project", Kind: "video", ContentType: "video/mp4", ObjectKey: "aicut-assets/" + id, Name: id, CreatedAt: now}); err != nil {
			t.Fatal(err)
		}
	}
	mux := http.NewServeMux()
	RegisterAICutMultiCamSyncRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, nil)
	for _, body := range []map[string]any{{"assetIds": []string{"a", "a"}, "referenceAssetId": "a"}, {"assetIds": []string{"a", "b"}, "referenceAssetId": "missing"}, {"assetIds": []string{"a", "b"}, "referenceAssetId": "a", "windowMs": 60001}, {"assetIds": []string{"a", "b"}, "referenceAssetId": "a", "maxLagMs": 10001}} {
		res := signedJSONRequest(t, mux, now, "alice", "POST", "/api/ai-cut/projects/sync-project/multicam-sync", body)
		if res.Code != 400 {
			t.Fatalf("status=%d %s", res.Code, res.Body.String())
		}
	}
	response := signedJSONRequest(t, mux, now, "bob", "POST", "/api/ai-cut/projects/sync-project/multicam-sync", map[string]any{"assetIds": []string{"a", "b"}, "referenceAssetId": "a"})
	if response.Code == 200 || response.Code == 503 {
		t.Fatalf("cross owner=%d", response.Code)
	}
}
func writeAICutSyncWAV(t *testing.T, target string, envelope []float64, shift int) {
	t.Helper()
	samples := len(envelope) * 160
	data := make([]byte, 44+samples*2)
	copy(data, "RIFF")
	binary.LittleEndian.PutUint32(data[4:], uint32(len(data)-8))
	copy(data[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(data[16:], 16)
	binary.LittleEndian.PutUint16(data[20:], 1)
	binary.LittleEndian.PutUint16(data[22:], 1)
	binary.LittleEndian.PutUint32(data[24:], 8000)
	binary.LittleEndian.PutUint32(data[28:], 16000)
	binary.LittleEndian.PutUint16(data[32:], 2)
	binary.LittleEndian.PutUint16(data[34:], 16)
	copy(data[36:], "data")
	binary.LittleEndian.PutUint32(data[40:], uint32(samples*2))
	for i := 0; i < samples; i++ {
		frame := i/160 - shift
		amplitude := 0.0
		if frame >= 0 && frame < len(envelope) {
			amplitude = envelope[frame] * .5
		}
		sample := int16(amplitude * math.Sin(2*math.Pi*400*float64(i)/8000) * 32767)
		binary.LittleEndian.PutUint16(data[44+i*2:], uint16(sample))
	}
	if err := os.WriteFile(target, data, 0600); err != nil {
		t.Fatal(err)
	}
}
func TestAICutMultiCamSyncRealFFmpegFixtures(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	objects := agentworkspace.NewLocalAssetObjectStore(filepath.Join(dir, "objects"))
	now := time.Now()
	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "sync-real", Title: "sync", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	reference := syncFixtureEnvelope(400)
	for index, shift := range []int{0, -17, 23} {
		id := fmt.Sprintf("camera-%d", index)
		wav := filepath.Join(dir, id+".wav")
		media := filepath.Join(dir, id+".mkv")
		writeAICutSyncWAV(t, wav, reference, shift)
		if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-f", "lavfi", "-i", "color=c=black:s=32x32:r=5:d=8", "-i", wav, "-c:v", "mpeg4", "-c:a", "pcm_s16le", "-shortest", media); err != nil {
			t.Fatal(err)
		}
		file, err := os.Open(media)
		if err != nil {
			t.Fatal(err)
		}
		stat, _ := file.Stat()
		asset := aicut.Asset{ID: id, ProjectID: "sync-real", Kind: "video", ContentType: "video/x-matroska", Name: id + ".mkv", ObjectKey: "aicut-assets/sync-real/" + id + ".mkv", ByteSize: stat.Size(), DurationMS: 8000, CreatedAt: now}
		err = objects.Put(t.Context(), aicutObjectAsset(asset), file)
		file.Close()
		if err != nil {
			t.Fatal(err)
		}
		if err = store.CreateAsset(t.Context(), "alice", asset); err != nil {
			t.Fatal(err)
		}
	}
	mux := http.NewServeMux()
	RegisterAICutMultiCamSyncRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, objects)
	response := signedJSONRequest(t, mux, now, "alice", "POST", "/api/ai-cut/projects/sync-real/multicam-sync", map[string]any{"assetIds": []string{"camera-0", "camera-1", "camera-2"}, "referenceAssetId": "camera-0", "windowMs": 8000, "maxLagMs": 1000})
	if response.Code != 200 {
		t.Fatalf("sync=%d %s", response.Code, response.Body.String())
	}
	var payload struct {
		Cameras          []aicutMultiCamSyncCamera `json:"cameras"`
		TimeResolutionMS int                       `json:"timeResolutionMs"`
	}
	decodeTestJSON(t, response.Body.Bytes(), &payload)
	if payload.TimeResolutionMS != 20 || len(payload.Cameras) != 3 {
		t.Fatalf("response=%s", response.Body.String())
	}
	for index, want := range []int{0, -340, 460} {
		if payload.Cameras[index].OffsetMS != want || payload.Cameras[index].Confidence < .98 {
			t.Fatalf("camera%d=%+v want=%d", index, payload.Cameras[index], want)
		}
	}
	assets, _ := store.ListAssets(t.Context(), "alice", "sync-real")
	if len(assets) != 3 {
		t.Fatal("analysis mutated asset list")
	}
}

func TestAICutMultiCamDecoderRejectsRealSilentAndMissingAudio(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	objects := agentworkspace.NewLocalAssetObjectStore(filepath.Join(dir, "objects"))
	for _, silent := range []bool{false, true} {
		id := fmt.Sprintf("silent-%v", silent)
		media := filepath.Join(dir, id+".mkv")
		args := []string{"-hide_banner", "-loglevel", "error", "-y", "-f", "lavfi", "-i", "color=c=black:s=32x32:r=5:d=3"}
		if silent {
			args = append(args, "-f", "lavfi", "-i", "anullsrc=r=8000:cl=mono", "-c:a", "pcm_s16le")
		}
		args = append(args, "-t", "3", "-c:v", "mpeg4", media)
		if err := ffmpegbin.Run(t.Context(), "ffmpeg", args...); err != nil {
			t.Fatal(err)
		}
		asset := aicut.Asset{ID: id, ProjectID: "silent", Kind: "video", Name: id + ".mkv", ObjectKey: "aicut-assets/silent/" + id + ".mkv"}
		file, err := os.Open(media)
		if err != nil {
			t.Fatal(err)
		}
		err = objects.Put(t.Context(), aicutObjectAsset(asset), file)
		file.Close()
		if err != nil {
			t.Fatal(err)
		}
		if _, err = decodeAICutSyncEnvelope(t.Context(), objects, asset, 3000); err == nil {
			t.Fatalf("accepted silent=%v", silent)
		}
	}
}
