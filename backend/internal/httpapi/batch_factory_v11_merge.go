package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/ffmpegbin"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
	"qiantie/backend/internal/mergeworker"
)

// mergeMediaTokenTTL bounds the short-lived signed URL used by the browser's
// native <video> element, which cannot send the Authorization header.
const mergeMediaTokenTTL = 2 * time.Hour

type mergeSubmitInput struct {
	RequestID            string  `json:"requestId"`
	TimingMode           string  `json:"timingMode"`
	AudioSource          string  `json:"audioSource"`
	AudioArtifactID      string  `json:"audioArtifactId"`
	Speed                float64 `json:"speed"`
	TTSSpeed             float64 `json:"ttsSpeed"`
	AudioDurationSeconds float64 `json:"audioDurationSeconds"`
}

func registerMergeRoutes(mux *http.ServeMux, service *batchfactoryv11.MergeService, files *localartifact.Store, output mergeworker.ObjectStore, remote mergeworker.ObjectReader, writable bool) {
	if writable {
		mux.HandleFunc("POST /api/batch-factory/v11/batches/{batchId}/merge-artifacts/migrate-to-tos", func(w http.ResponseWriter, r *http.Request) {
			owner, ok := bridgeOwner(r)
			if !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			if files == nil || output == nil {
				writeStoreError(w, batchfactoryv11.ErrUnavailable)
				return
			}
			repository, ok := service.Store.(batchfactoryv11.MergeRepository)
			if !ok {
				writeStoreError(w, batchfactoryv11.ErrUnavailable)
				return
			}
			batchID := r.PathValue("batchId")
			jobs, err := repository.ListMergeJobs(r.Context(), owner, batchID)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			migrated, retained := 0, 0
			for _, job := range jobs {
				prefix := "/api/batch-factory/v12/batches/" + batchID + "/merge-media/"
				if job.Status != batchfactoryv11.MergeSucceeded || !strings.HasPrefix(strings.TrimSpace(job.OutputURL), prefix) {
					continue
				}
				artifactID := strings.TrimPrefix(strings.TrimSpace(job.OutputURL), prefix)
				path, pathErr := files.Path(artifactID + ".mp4")
				if pathErr != nil {
					continue // already remote or a separately retained legacy artifact
				}
				if _, uploadErr := output.PutMerged(r.Context(), job.ID, path); uploadErr != nil {
					retained++
					continue
				}
				updated := job
				updated.OutputURL = prefix + job.ID
				updated.ProgressPhase = "stored"
				updated.ErrorMessage = ""
				if _, updateErr := repository.UpdateMergeJob(r.Context(), owner, job.ID, updated); updateErr != nil {
					retained++
					continue
				}
				if removeErr := files.Remove(artifactID + ".mp4"); removeErr != nil {
					retained++
					continue
				}
				migrated++
			}
			writeJSON(w, http.StatusOK, map[string]any{"migrated": migrated, "retained": retained})
		})
		mux.HandleFunc("POST /api/batch-factory/v11/batches/{batchId}/merge", func(w http.ResponseWriter, r *http.Request) {
			owner, ok := bridgeOwner(r)
			if !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			var input mergeSubmitInput
			if !decodeJSON(w, r, &input) {
				return
			}
			if strings.TrimSpace(input.RequestID) == "" {
				writeStoreError(w, batchfactoryv11.ErrInvalid)
				return
			}
			job, err := service.SubmitBatchMerge(r.Context(), owner, r.PathValue("batchId"), input.RequestID, batchfactoryv11.MergeOptions{TimingMode: input.TimingMode, AudioSource: input.AudioSource, AudioArtifactID: input.AudioArtifactID, Speed: input.Speed, TTSSpeed: input.TTSSpeed, AudioDurationSeconds: input.AudioDurationSeconds})
			if err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]any{"job": job})
		})
		mux.HandleFunc("POST /api/batch-factory/v11/batches/{batchId}/books/{bookId}/merge", func(w http.ResponseWriter, r *http.Request) {
			owner, ok := bridgeOwner(r)
			if !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			var input mergeSubmitInput
			if !decodeJSON(w, r, &input) {
				return
			}
			if strings.TrimSpace(input.RequestID) == "" {
				writeStoreError(w, batchfactoryv11.ErrInvalid)
				return
			}
			jobs, err := service.SubmitBookMerge(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), input.RequestID, batchfactoryv11.MergeOptions{TimingMode: input.TimingMode, AudioSource: input.AudioSource, AudioArtifactID: input.AudioArtifactID, Speed: input.Speed, TTSSpeed: input.TTSSpeed, AudioDurationSeconds: input.AudioDurationSeconds})
			if err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]any{"jobs": jobs, "job": jobs[0]})
		})
	}
	registerMergeStatusRoute(mux, service, "v11")
	mux.HandleFunc("GET /api/batch-factory/v11/batches/{batchId}/merge-media/{artifactId}", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		repository, ok := service.Store.(batchfactoryv11.MergeRepository)
		if !ok {
			writeStoreError(w, batchfactoryv11.ErrUnavailable)
			return
		}
		serveMergeMediaArtifact(w, r, owner, r.PathValue("batchId"), r.PathValue("artifactId"), repository, files, remote)
	})
	mux.HandleFunc("GET /api/batch-factory/v11/batches/{batchId}/merge-cover/{artifactId}", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		repository, ok := service.Store.(batchfactoryv11.MergeRepository)
		if !ok || files == nil {
			writeStoreError(w, batchfactoryv11.ErrUnavailable)
			return
		}
		batchID, artifactID := r.PathValue("batchId"), r.PathValue("artifactId")
		jobs, err := repository.ListMergeJobs(r.Context(), owner, batchID)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		found := false
		for _, job := range jobs {
			if job.Status == batchfactoryv11.MergeSucceeded && mergeOutputMatchesArtifact(batchID, artifactID, job.OutputURL) {
				found = true
				break
			}
		}
		if !found {
			writeStoreError(w, batchfactoryv11.ErrNotFound)
			return
		}
		coverRef := artifactID + ".jpg"
		cover, openErr := files.Open(coverRef)
		if openErr != nil {
			coverBytes, renderErr := renderMergeCover(r.Context(), files, remote, artifactID)
			if renderErr != nil {
				log.Printf("batch factory merge cover failed: batch=%q artifact=%q error=%v", batchID, artifactID, renderErr)
				writeStoreError(w, batchfactoryv11.ErrUnavailable)
				return
			}
			if _, saveErr := files.SaveImage(artifactID, "image/jpeg", bytes.NewReader(coverBytes)); saveErr != nil && !errors.Is(saveErr, localartifact.ErrAlreadyExists) {
				writeStoreError(w, batchfactoryv11.ErrUnavailable)
				return
			}
			cover, openErr = files.Open(coverRef)
			if openErr != nil {
				writeStoreError(w, batchfactoryv11.ErrUnavailable)
				return
			}
		}
		defer cover.Close()
		w.Header().Set("Cache-Control", "private, max-age=86400")
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.ServeContent(w, r, artifactID+".jpg", time.Time{}, cover)
	})
}

func mergeOutputMatchesArtifact(batchID, artifactID, outputURL string) bool {
	outputURL = strings.TrimSpace(outputURL)
	for _, version := range []string{"v11", "v12"} {
		if outputURL == "/api/batch-factory/"+version+"/batches/"+batchID+"/merge-media/"+artifactID {
			return true
		}
	}
	return false
}

func mergeMediaClock(now func() time.Time) time.Time {
	if now != nil {
		return now().UTC()
	}
	return time.Now().UTC()
}

func findMergeArtifactJob(ctx context.Context, repository batchfactoryv11.MergeRepository, owner, batchID, artifactID string) (batchfactoryv11.MergeJob, bool, error) {
	jobs, err := repository.ListMergeJobs(ctx, owner, batchID)
	if err != nil {
		return batchfactoryv11.MergeJob{}, false, err
	}
	for _, job := range jobs {
		if job.Status == batchfactoryv11.MergeSucceeded && mergeOutputMatchesArtifact(batchID, artifactID, job.OutputURL) {
			return job, true, nil
		}
	}
	return batchfactoryv11.MergeJob{}, false, nil
}

// mergeMediaETag uses the local file's immutable size+mtime when available.
// Remote merge artifacts keep their job revision, which is stable for a
// succeeded artifact and cheap to compute on every request.
func mergeMediaETag(job batchfactoryv11.MergeJob, fileInfo os.FileInfo) string {
	if fileInfo != nil {
		return fmt.Sprintf("\"%x-%x\"", fileInfo.Size(), fileInfo.ModTime().UTC().UnixNano())
	}
	digest := sha256.Sum256([]byte(job.ID + "|" + job.UpdatedAt.UTC().Format(time.RFC3339Nano)))
	return "\"" + hex.EncodeToString(digest[:8]) + "\""
}

func serveMergeMediaArtifact(w http.ResponseWriter, r *http.Request, owner, batchID, artifactID string, repository batchfactoryv11.MergeRepository, files *localartifact.Store, remote mergeworker.ObjectReader) {
	job, found, err := findMergeArtifactJob(r.Context(), repository, owner, batchID, artifactID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if !found {
		writeStoreError(w, batchfactoryv11.ErrNotFound)
		return
	}
	var file io.ReadCloser
	var fileInfo os.FileInfo
	if files != nil {
		opened, localErr := files.Open(artifactID + ".mp4")
		if localErr == nil {
			file = opened
			if stat, statErr := opened.Stat(); statErr == nil {
				fileInfo = stat
			}
		} else if !errors.Is(localErr, localartifact.ErrInvalidID) && !errors.Is(localErr, os.ErrNotExist) {
			writeStoreError(w, batchfactoryv11.ErrUnavailable)
			return
		}
	}
	if file == nil && remote != nil {
		file, err = remote.Open(r.Context(), artifactID)
	}
	if file == nil || err != nil {
		writeStoreError(w, batchfactoryv11.ErrNotFound)
		return
	}
	defer file.Close()
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "private, max-age=3600")
	w.Header().Set("ETag", mergeMediaETag(job, fileInfo))
	// 浏览器播放器会先发 Range 请求。这里必须走 ServeContent 才能返回 206、
	// Accept-Ranges 与 Content-Length；直接 io.Copy 会退化成 200 + chunked，
	// 播放器无法起播，停在 0:00 黑屏并反复整包重下。ETag 也由 ServeContent
	// 参与 If-None-Match 协商，命中时直接返回 304。
	if seeker, ok := file.(io.ReadSeeker); ok {
		modTime := time.Time{}
		if fileInfo != nil {
			modTime = fileInfo.ModTime()
		}
		http.ServeContent(w, r, artifactID+".mp4", modTime, seeker)
		return
	}
	_, _ = io.Copy(w, file)
}

// registerMergeMediaTokenRoute mints a short-lived signed playback URL so the
// browser's native <video> element, which cannot send Authorization headers,
// can stream merge artifacts directly with Range requests.
func registerMergeMediaTokenRoute(mux *http.ServeMux, version string, service *batchfactoryv11.MergeService, secret string, now func() time.Time) {
	if service == nil {
		return
	}
	mux.HandleFunc("POST /api/batch-factory/"+version+"/batches/{batchId}/merge-media/{artifactId}/token", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		repository, ok := service.Store.(batchfactoryv11.MergeRepository)
		if !ok {
			writeStoreError(w, batchfactoryv11.ErrUnavailable)
			return
		}
		batchID, artifactID := r.PathValue("batchId"), r.PathValue("artifactId")
		if _, found, err := findMergeArtifactJob(r.Context(), repository, owner, batchID, artifactID); err != nil {
			writeStoreError(w, err)
			return
		} else if !found {
			writeStoreError(w, batchfactoryv11.ErrNotFound)
			return
		}
		expiresAt := mergeMediaClock(now).Add(mergeMediaTokenTTL)
		token, err := localexecutor.ArtifactPublicToken(secret, owner, artifactID, expiresAt)
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "media playback unavailable"})
			return
		}
		playbackPath := "/api/batch-factory/" + version + "/batches/" + url.PathEscape(batchID) + "/merge-media/" + url.PathEscape(artifactID)
		writeJSON(w, http.StatusOK, map[string]any{
			"url":       playbackPath + "?token=" + url.QueryEscape(token),
			"expiresAt": expiresAt.Format(time.RFC3339),
		})
	})
}

// registerMergeMediaPublicRoute keeps the authenticated merge-media handler as
// the fallback for normal API clients, but lets token-signed requests from a
// native player stream the same artifact without the Authorization header.
func registerMergeMediaPublicRoute(root *http.ServeMux, version string, fallback http.Handler, service *batchfactoryv11.MergeService, files *localartifact.Store, remote mergeworker.ObjectReader, secret string, now func() time.Time) {
	if root == nil || fallback == nil || service == nil {
		return
	}
	root.HandleFunc("GET /api/batch-factory/"+version+"/batches/{batchId}/merge-media/{artifactId}", func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		if token == "" {
			fallback.ServeHTTP(w, r)
			return
		}
		owner, err := localexecutor.VerifyArtifactPublicToken(secret, r.PathValue("artifactId"), token, mergeMediaClock(now))
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired media token"})
			return
		}
		repository, ok := service.Store.(batchfactoryv11.MergeRepository)
		if !ok {
			writeStoreError(w, batchfactoryv11.ErrUnavailable)
			return
		}
		serveMergeMediaArtifact(w, r, owner, r.PathValue("batchId"), r.PathValue("artifactId"), repository, files, remote)
	})
}

func registerMergeStatusRoute(mux *http.ServeMux, service *batchfactoryv11.MergeService, version string) {
	mux.HandleFunc("GET /api/batch-factory/"+version+"/batches/{batchId}/merge-status", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		jobs, err := service.GetBatchStatus(r.Context(), owner, r.PathValue("batchId"))
		if err != nil {
			log.Printf("batch factory merge status failed: batch=%q owner=%q error=%v", r.PathValue("batchId"), owner, err)
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"batchId": r.PathValue("batchId"), "jobs": jobs})
	})
}

func renderMergeCover(ctx context.Context, files *localartifact.Store, remote mergeworker.ObjectReader, artifactID string) ([]byte, error) {
	inputPath, cleanup, err := mergeCoverInput(ctx, files, remote, artifactID)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	var stderr bytes.Buffer
	command, binaryPath, cmdErr := ffmpegbin.Command(ctx, "ffmpeg", "-v", "error", "-ss", "0.1", "-i", inputPath, "-frames:v", "1", "-f", "image2pipe", "-vcodec", "mjpeg", "pipe:1")
	if cmdErr != nil {
		return nil, cmdErr
	}
	command.Stderr = &stderr
	cover, err := command.Output()
	if err != nil {
		if ffmpegbin.IsMissing(err) {
			return nil, &ffmpegbin.MissingError{Binary: "ffmpeg", Tried: []string{binaryPath}}
		}
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			return nil, fmt.Errorf("render merge cover: %w", err)
		}
		return nil, fmt.Errorf("render merge cover: %s", message)
	}
	if len(cover) < 3 || cover[0] != 0xff || cover[1] != 0xd8 || cover[2] != 0xff {
		return nil, errors.New("render merge cover returned an invalid jpeg")
	}
	return cover, nil
}

func mergeCoverInput(ctx context.Context, files *localartifact.Store, remote mergeworker.ObjectReader, artifactID string) (string, func(), error) {
	if localPath, err := files.Path(artifactID + ".mp4"); err == nil {
		return localPath, func() {}, nil
	} else if !errors.Is(err, os.ErrNotExist) && !errors.Is(err, localartifact.ErrInvalidID) {
		return "", nil, err
	}
	if remote == nil {
		return "", nil, os.ErrNotExist
	}
	reader, err := remote.Open(ctx, artifactID)
	if err != nil {
		return "", nil, err
	}
	defer reader.Close()
	file, err := os.CreateTemp("", "qiantie-merge-cover-*.mp4")
	if err != nil {
		return "", nil, err
	}
	path := file.Name()
	committed := false
	defer func() {
		if !committed {
			_ = file.Close()
			_ = os.Remove(path)
		}
	}()
	if _, err := io.Copy(file, reader); err != nil {
		return "", nil, err
	}
	if err := file.Close(); err != nil {
		return "", nil, err
	}
	if info, err := os.Stat(path); err != nil || info.Size() == 0 {
		if err == nil {
			err = errors.New("remote merged video is empty")
		}
		return "", nil, err
	}
	committed = true
	return filepath.Clean(path), func() { _ = os.Remove(path) }, nil
}

func jobTime(jobs []batchfactoryv11.MergeJob, output string) (at time.Time) {
	for _, job := range jobs {
		if strings.TrimSpace(job.OutputURL) == output {
			return job.UpdatedAt
		}
	}
	return at
}
