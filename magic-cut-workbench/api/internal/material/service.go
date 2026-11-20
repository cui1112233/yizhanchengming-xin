package material

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	worktask "yizhanchengming/magic-cut-workbench/internal/task"
)

type Metadata struct {
	DurationMs int64   `json:"durationMs"`
	Width      int     `json:"width,omitempty"`
	Height     int     `json:"height,omitempty"`
	FrameRate  float64 `json:"frameRate,omitempty"`
	VideoCodec string  `json:"videoCodec,omitempty"`
	AudioCodec string  `json:"audioCodec,omitempty"`
}

type Material struct {
	ID           string    `json:"id"`
	ProjectID    string    `json:"projectId"`
	Owner        string    `json:"-"`
	OriginalName string    `json:"originalName"`
	MediaType    string    `json:"mediaType"`
	StorageKey   string    `json:"-"`
	SizeBytes    int64     `json:"sizeBytes"`
	Metadata     Metadata  `json:"metadata"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type ServiceError struct {
	Code string
	Err  error
}

func (e *ServiceError) Error() string {
	if e.Err == nil {
		return e.Code
	}
	return e.Err.Error()
}
func (e *ServiceError) Unwrap() error { return e.Err }

func ErrorCode(err error) string {
	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) {
		return serviceErr.Code
	}
	return ""
}

type Probe interface {
	Probe(context.Context, string) (Metadata, error)
}

type Repository interface {
	SaveImport(context.Context, Material, worktask.Task) error
}

type ServiceOptions struct {
	StorageRoot   string
	MaxBytes      int64
	Probe         Probe
	Repository    Repository
	NewMaterialID func() string
	NewTaskID     func() string
	Now           func() time.Time
}

type Service struct{ options ServiceOptions }

func NewService(options ServiceOptions) *Service {
	if options.MaxBytes <= 0 {
		options.MaxBytes = 10 << 30
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Service{options: options}
}

func ClassifyMedia(name string) (string, error) {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".mp4", ".mov", ".mkv", ".webm", ".avi", ".m4v":
		return "video", nil
	case ".mp3", ".wav", ".aac", ".m4a", ".flac", ".ogg":
		return "audio", nil
	case ".png", ".jpg", ".jpeg", ".webp", ".bmp":
		return "image", nil
	default:
		return "", &ServiceError{Code: "FORMAT_NOT_SUPPORTED", Err: fmt.Errorf("不支持的素材格式：%s", ext)}
	}
}

func (s *Service) Import(ctx context.Context, owner, projectID string, file multipart.File, header *multipart.FileHeader) (Material, worktask.Task, error) {
	if s == nil || header == nil || file == nil || strings.TrimSpace(owner) == "" || strings.TrimSpace(projectID) == "" {
		return Material{}, worktask.Task{}, errors.New("owner, project and file are required")
	}
	if header.Size > s.options.MaxBytes {
		return Material{}, worktask.Task{}, &ServiceError{Code: "FILE_TOO_LARGE", Err: errors.New("素材文件超过大小限制")}
	}
	mediaType, err := ClassifyMedia(header.Filename)
	if err != nil {
		return Material{}, worktask.Task{}, err
	}
	if s.options.NewMaterialID == nil || s.options.NewTaskID == nil || strings.TrimSpace(s.options.StorageRoot) == "" {
		return Material{}, worktask.Task{}, errors.New("material service is not configured")
	}
	materialID := s.options.NewMaterialID()
	taskID := s.options.NewTaskID()
	ext := strings.ToLower(filepath.Ext(header.Filename))
	destination := filepath.Join(s.options.StorageRoot, safeSegment(owner), safeSegment(projectID), safeSegment(materialID)+ext)
	if !insideRoot(s.options.StorageRoot, destination) {
		return Material{}, worktask.Task{}, &ServiceError{Code: "FILE_PERMISSION_DENIED", Err: errors.New("素材路径不安全")}
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return Material{}, worktask.Task{}, mapWriteError(err)
	}
	out, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return Material{}, worktask.Task{}, mapWriteError(err)
	}
	written, copyErr := io.Copy(out, io.LimitReader(file, s.options.MaxBytes+1))
	closeErr := out.Close()
	if copyErr != nil || closeErr != nil || written > s.options.MaxBytes {
		_ = os.Remove(destination)
		if written > s.options.MaxBytes {
			return Material{}, worktask.Task{}, &ServiceError{Code: "FILE_TOO_LARGE", Err: errors.New("素材文件超过大小限制")}
		}
		if copyErr != nil {
			return Material{}, worktask.Task{}, mapWriteError(copyErr)
		}
		return Material{}, worktask.Task{}, mapWriteError(closeErr)
	}
	now := s.options.Now().UTC()
	item := Material{ID: materialID, ProjectID: projectID, Owner: owner, OriginalName: filepath.Base(header.Filename), MediaType: mediaType, StorageKey: destination, SizeBytes: written, Status: "processing", CreatedAt: now, UpdatedAt: now}
	queued := worktask.Task{ID: taskID, ProjectID: projectID, Owner: owner, Type: "material_probe", Status: worktask.StatusRunning, Progress: 0.25, CreatedAt: now, UpdatedAt: now}
	if s.options.Probe != nil {
		metadata, probeErr := s.options.Probe.Probe(ctx, destination)
		if probeErr != nil {
			item.Status = "failed"
			queued.Status = worktask.StatusFailed
			queued.ErrorCode = "MEDIA_PROBE_FAILED"
			queued.ErrorMessage = "素材解析失败"
			if s.options.Repository != nil {
				_ = s.options.Repository.SaveImport(ctx, item, queued)
			}
			return item, queued, &ServiceError{Code: "MEDIA_PROBE_FAILED", Err: probeErr}
		}
		item.Metadata = metadata
	}
	item.Status = "ready"
	queued.Status = worktask.StatusSucceeded
	queued.Progress = 1
	if s.options.Repository != nil {
		if err := s.options.Repository.SaveImport(ctx, item, queued); err != nil {
			return Material{}, worktask.Task{}, err
		}
	}
	return item, queued, nil
}

func ParseProbeJSON(data []byte) (Metadata, error) {
	var payload struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
		Streams []struct {
			CodecType string `json:"codec_type"`
			CodecName string `json:"codec_name"`
			Width     int    `json:"width"`
			Height    int    `json:"height"`
			FrameRate string `json:"r_frame_rate"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return Metadata{}, err
	}
	duration, _ := strconv.ParseFloat(payload.Format.Duration, 64)
	result := Metadata{DurationMs: int64(duration*1000 + 0.5)}
	for _, stream := range payload.Streams {
		switch stream.CodecType {
		case "video":
			result.Width, result.Height, result.VideoCodec = stream.Width, stream.Height, stream.CodecName
			parts := strings.Split(stream.FrameRate, "/")
			if len(parts) == 2 {
				numerator, _ := strconv.ParseFloat(parts[0], 64)
				denominator, _ := strconv.ParseFloat(parts[1], 64)
				if denominator != 0 {
					result.FrameRate = numerator / denominator
				}
			}
		case "audio":
			result.AudioCodec = stream.CodecName
		}
	}
	return result, nil
}

func safeSegment(value string) string {
	var builder strings.Builder
	for _, char := range strings.TrimSpace(value) {
		if unicode.IsLetter(char) || unicode.IsDigit(char) || char == '-' || char == '_' || char == '.' {
			builder.WriteRune(char)
		} else {
			builder.WriteByte('_')
		}
	}
	result := strings.Trim(builder.String(), ".")
	if result == "" {
		return "unknown"
	}
	return result
}

func insideRoot(root, candidate string) bool {
	rootPath, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	candidatePath, err := filepath.Abs(candidate)
	if err != nil {
		return false
	}
	relative, err := filepath.Rel(rootPath, candidatePath)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func mapWriteError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, os.ErrPermission) {
		return &ServiceError{Code: "FILE_PERMISSION_DENIED", Err: errors.New("没有素材目录写入权限")}
	}
	return &ServiceError{Code: "DISK_SPACE_INSUFFICIENT", Err: fmt.Errorf("素材写入失败: %w", err)}
}
