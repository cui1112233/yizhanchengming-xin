package httpapi

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/ffmpegbin"
	"strconv"
	"strings"
	"sync"
	"time"
)

const aicutSyncResolutionMS = 20

var aicutSyncDecodeSlots = make(chan struct{}, 2)

type aicutMultiCamSyncRequest struct {
	AssetIDs         []string `json:"assetIds"`
	ReferenceAssetID string   `json:"referenceAssetId"`
	WindowMS         int      `json:"windowMs"`
	MaxLagMS         int      `json:"maxLagMs"`
}
type aicutMultiCamSyncCamera struct {
	AssetID    string  `json:"assetId"`
	OffsetMS   int     `json:"offsetMs"`
	Confidence float64 `json:"confidence"`
}

// This read-only analysis returns reviewable offsets and never changes assets or timelines.
func RegisterAICutMultiCamSyncRoutes(mux *http.ServeMux, auth BridgeAuth, store aicut.Store, objects agentworkspace.AssetObjectStore) {
	if mux == nil || store == nil {
		return
	}
	mux.Handle("POST /api/ai-cut/projects/{id}/multicam-sync", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, 401, map[string]string{"error": "unauthorized"})
			return
		}
		owner := strings.TrimSpace(identity.Username)
		projectID := req.PathValue("id")
		if _, err := store.Project(req.Context(), owner, projectID); err != nil {
			writeAICutError(w, err)
			return
		}
		var input aicutMultiCamSyncRequest
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		if input.WindowMS == 0 {
			input.WindowMS = 30000
		}
		if input.MaxLagMS == 0 {
			input.MaxLagMS = 5000
		}
		if len(input.AssetIDs) < 2 || len(input.AssetIDs) > 9 || input.WindowMS < 2000 || input.WindowMS > 60000 || input.MaxLagMS < 20 || input.MaxLagMS > 10000 {
			writeJSON(w, 400, map[string]string{"error": "请选择2到9个机位，分析时长2000到60000毫秒，最大偏移20到10000毫秒"})
			return
		}
		assets := make([]aicut.Asset, len(input.AssetIDs))
		seen := map[string]bool{}
		reference := -1
		for i, id := range input.AssetIDs {
			if id == "" || seen[id] {
				writeJSON(w, 400, map[string]string{"error": "机位素材编号为空或重复"})
				return
			}
			seen[id] = true
			asset, err := store.Asset(req.Context(), owner, projectID, id)
			if err != nil {
				writeAICutError(w, err)
				return
			}
			if asset.ProjectID != projectID || asset.Kind != "video" {
				writeJSON(w, 400, map[string]string{"error": "机位须为当前工程的自有视频"})
				return
			}
			if asset.ByteSize < 0 || asset.ByteSize > 2<<30 {
				writeJSON(w, 400, map[string]string{"error": "机位素材超过2GB上限"})
				return
			}
			assets[i] = asset
			if id == input.ReferenceAssetID {
				reference = i
			}
		}
		if reference < 0 {
			writeJSON(w, 400, map[string]string{"error": "参考机位须在所选机位中"})
			return
		}
		if objects == nil {
			writeJSON(w, 503, map[string]string{"error": "素材存储尚未配置"})
			return
		}
		if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
			writeJSON(w, 503, map[string]string{"error": "声音同步需要本地FFmpeg"})
			return
		}
		ctx, cancel := context.WithTimeout(req.Context(), 90*time.Second)
		defer cancel()
		envelopes := make([][]float64, len(assets))
		errs := make([]error, len(assets))
		var wg sync.WaitGroup
		for i, asset := range assets {
			wg.Add(1)
			go func(i int, asset aicut.Asset) {
				defer wg.Done()
				envelopes[i], errs[i] = decodeAICutSyncEnvelope(ctx, objects, asset, input.WindowMS)
			}(i, asset)
		}
		wg.Wait()
		for i, err := range errs {
			if err != nil {
				status := 422
				if ctx.Err() != nil {
					status = 408
				}
				writeJSON(w, status, map[string]string{"error": fmt.Sprintf("机位%d无法自动同步：%s", i+1, err)})
				return
			}
		}
		cameras := make([]aicutMultiCamSyncCamera, len(assets))
		for i, asset := range assets {
			offset, confidence := 0, 1.0
			if i != reference {
				var err error
				offset, confidence, err = correlateAICutMultiCam(ctx, envelopes[reference], envelopes[i], input.MaxLagMS)
				if err != nil {
					status := 422
					if ctx.Err() != nil {
						status = 408
					}
					writeJSON(w, status, map[string]string{"error": fmt.Sprintf("机位%d同步失败：%s", i+1, err)})
					return
				}
			}
			cameras[i] = aicutMultiCamSyncCamera{asset.ID, offset, confidence}
		}
		writeJSON(w, 200, map[string]any{"method": "audio-energy-cross-correlation", "timeResolutionMs": aicutSyncResolutionMS, "windowMs": input.WindowMS, "maxLagMs": input.MaxLagMS, "referenceAssetId": input.ReferenceAssetID, "cameras": cameras})
	})))
}
func decodeAICutSyncEnvelope(ctx context.Context, objects agentworkspace.AssetObjectStore, asset aicut.Asset, windowMS int) ([]float64, error) {
	select {
	case aicutSyncDecodeSlots <- struct{}{}:
		defer func() { <-aicutSyncDecodeSlots }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	dir, err := os.MkdirTemp("", "qiantie-multicam-sync-")
	if err != nil {
		return nil, errors.New("无法准备分析目录")
	}
	defer os.RemoveAll(dir)
	// Copy through the existing object API, with a hard byte cap and cancellation.
	input := filepath.Join(dir, "source.media")
	if err := copyAICutSyncSource(ctx, objects, asset, input); err != nil {
		return nil, err
	}
	output := filepath.Join(dir, "audio.pcm")
	if err := ffmpegbin.Run(ctx, "ffmpeg", "-hide_banner", "-loglevel", "error", "-nostdin", "-y", "-threads", "1", "-i", input, "-map", "0:a:0", "-vn", "-t", strconv.FormatFloat(float64(windowMS)/1000, 'f', 3, 64), "-ac", "1", "-ar", "8000", "-c:a", "pcm_s16le", "-f", "s16le", output); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errors.New("素材没有可解码声音，请改用手动对齐")
	}
	file, err := os.Open(output)
	if err != nil {
		return nil, errors.New("声音分析读取失败")
	}
	defer file.Close()
	limit := int64(windowMS) * 16
	raw, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil || int64(len(raw)) > limit {
		return nil, errors.New("声音分析数据超过范围")
	}
	envelope := make([]float64, len(raw)/320)
	for frame := range envelope {
		var energy float64
		for j := 0; j < 160; j++ {
			sample := float64(int16(binary.LittleEndian.Uint16(raw[frame*320+j*2:]))) / 32768
			energy += sample * sample
		}
		envelope[frame] = math.Sqrt(energy / 160)
	}
	if err := usableAICutEnvelope(envelope); err != nil {
		return nil, err
	}
	return envelope, nil
}
func copyAICutSyncSource(ctx context.Context, objects agentworkspace.AssetObjectStore, asset aicut.Asset, target string) error {
	body, err := objects.Open(ctx, aicutObjectAsset(asset))
	if err != nil {
		return errors.New("源素材读取失败")
	}
	defer body.Close()
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			body.Close()
		case <-done:
		}
	}()
	defer close(done)
	file, err := os.Create(target)
	if err != nil {
		return err
	}
	defer file.Close()
	count, err := io.Copy(file, io.LimitReader(body, (2<<30)+1))
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil || count > 2<<30 {
		return errors.New("源素材读取失败或超过2GB上限")
	}
	return nil
}
func usableAICutEnvelope(values []float64) error {
	if len(values) < 100 {
		return errors.New("可用声音不足2秒，请改用手动对齐")
	}
	var sum, square float64
	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return errors.New("声音分析数据无效")
		}
		sum += v
		square += v * v
	}
	if square/float64(len(values)) < 1e-8 || square-sum*sum/float64(len(values)) < 1e-8 {
		return errors.New("声音静音或缺少变化，无法可靠同步")
	}
	return nil
}

// Positive offset means camera[T+offset] matches reference[T].
func correlateAICutMultiCam(ctx context.Context, reference, camera []float64, maxLagMS int) (int, float64, error) {
	if err := usableAICutEnvelope(reference); err != nil {
		return 0, 0, err
	}
	if err := usableAICutEnvelope(camera); err != nil {
		return 0, 0, err
	}
	type score struct {
		lag   int
		value float64
	}
	scores := []score{}
	best := score{0, -2}
	maxLag := maxLagMS / aicutSyncResolutionMS
	for lag := -maxLag; lag <= maxLag; lag++ {
		if err := ctx.Err(); err != nil {
			return 0, 0, err
		}
		start := max(0, -lag)
		end := min(len(reference), len(camera)-lag)
		n := end - start
		if n < 100 || n < min(len(reference), len(camera))/2 {
			continue
		}
		var sx, sy, xx, yy, xy float64
		for i := start; i < end; i++ {
			x, y := reference[i], camera[i+lag]
			sx += x
			sy += y
			xx += x * x
			yy += y * y
			xy += x * y
		}
		vx, vy := xx-sx*sx/float64(n), yy-sy*sy/float64(n)
		if vx < 1e-8 || vy < 1e-8 {
			continue
		}
		value := (xy - sx*sy/float64(n)) / math.Sqrt(vx*vy)
		s := score{lag, value}
		scores = append(scores, s)
		if value > best.value {
			best = s
		}
	}
	if best.value < .65 {
		return 0, 0, errors.New("声音匹配不足，无法可靠同步，请手动对齐")
	}
	for _, s := range scores {
		if absSyncInt(s.lag-best.lag) > 5 && s.value > best.value-.06 {
			return 0, 0, errors.New("声音存在多个相似位置，无法确定唯一偏移，请手动对齐")
		}
	}
	return best.lag * aicutSyncResolutionMS, math.Min(1, math.Max(0, best.value)), nil
}
func absSyncInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
