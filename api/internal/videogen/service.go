package videogen

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/media"
)

type ProviderFactory interface {
	ForModel(context.Context, string, string, string) (Provider, string, error)
}

type ReferenceResolver interface {
	Resolve(context.Context, string, string) (media.ResolvedAsset, error)
}

type DownloadedMedia struct {
	Body          io.ReadCloser
	ContentType   string
	ContentLength int64
}

type MediaDownloader interface {
	Download(context.Context, string) (DownloadedMedia, error)
}

type MediaIngestor interface {
	Ingest(context.Context, media.IngestInput) (media.Asset, error)
}

type GenerateInput struct {
	ToolCallID            string
	Model                 string
	FallbackProvider      string
	Prompt                string
	Duration              int
	AspectRatio           string
	Resolution            string
	ReferenceMediaAssetIDs []string
}

type Service struct {
	Providers    ProviderFactory
	References   ReferenceResolver
	Downloader   MediaDownloader
	Ingestor     MediaIngestor
	PollInterval time.Duration
	MaxPolls     int
	Sleep        func(context.Context, time.Duration) error
}

func (s Service) Generate(ctx context.Context, owner string, input GenerateInput) (media.Asset, error) {
	if s.Providers == nil || s.Downloader == nil || s.Ingestor == nil {
		return media.Asset{}, errors.New("video generation service unavailable")
	}
	owner = strings.TrimSpace(owner)
	input.Model = strings.TrimSpace(input.Model)
	input.Prompt = strings.TrimSpace(input.Prompt)
	input.ToolCallID = strings.TrimSpace(input.ToolCallID)
	if owner == "" || input.Model == "" || input.Prompt == "" { return media.Asset{}, errors.New("owner, model and prompt are required") }
	if input.Duration <= 0 { return media.Asset{}, errors.New("video duration is required") }

	referenceURLs := make([]string, 0, len(input.ReferenceMediaAssetIDs))
	if len(input.ReferenceMediaAssetIDs) > 0 && s.References == nil { return media.Asset{}, errors.New("video reference resolver unavailable") }
	for _, id := range input.ReferenceMediaAssetIDs {
		id = strings.TrimSpace(id)
		if id == "" { continue }
		resolved, err := s.References.Resolve(ctx, owner, id)
		if err != nil { return media.Asset{}, fmt.Errorf("resolve reference media %s: %w", id, err) }
		if resolved.Asset.MediaType != media.TypeImage { return media.Asset{}, fmt.Errorf("reference media %s is not an image", id) }
		if strings.TrimSpace(resolved.PreviewURL) == "" { return media.Asset{}, fmt.Errorf("reference media %s has no preview URL", id) }
		referenceURLs = append(referenceURLs, resolved.PreviewURL)
	}

	provider, providerName, err := s.Providers.ForModel(ctx, owner, input.Model, input.FallbackProvider)
	if err != nil { return media.Asset{}, err }
	task, err := provider.Submit(ctx, Request{Model:input.Model,Prompt:input.Prompt,Duration:input.Duration,AspectRatio:input.AspectRatio,Resolution:input.Resolution,ReferenceURLs:referenceURLs})
	if err != nil { return media.Asset{}, fmt.Errorf("submit video generation: %w", err) }
	if task.State == StateFailed { return media.Asset{}, errors.New("video provider rejected generation") }
	if task.State != StateSucceeded {
		maxPolls := s.MaxPolls
		if maxPolls <= 0 { maxPolls = 180 }
		interval := s.PollInterval
		if interval <= 0 { interval = 3 * time.Second }
		for attempt := 0; attempt < maxPolls; attempt++ {
			if err := s.sleep(ctx, interval); err != nil { return media.Asset{}, err }
			task, err = provider.Poll(ctx, task)
			if err != nil { return media.Asset{}, fmt.Errorf("poll video generation: %w", err) }
			if task.State == StateSucceeded { break }
			if task.State == StateFailed { return media.Asset{}, errors.New("video provider generation failed") }
		}
	}
	if task.State != StateSucceeded { return media.Asset{}, errors.New("video generation polling limit reached") }
	if strings.TrimSpace(task.MediaURL) == "" { return media.Asset{}, errors.New("video provider succeeded without media URL") }

	downloaded, err := s.Downloader.Download(ctx, task.MediaURL)
	if err != nil { return media.Asset{}, fmt.Errorf("download generated video: %w", err) }
	if downloaded.Body == nil { return media.Asset{}, errors.New("generated video download has no body") }
	defer downloaded.Body.Close()
	metadata, _ := json.Marshal(map[string]any{
		"tool_call_id": input.ToolCallID,
		"provider": providerName,
		"model": input.Model,
		"provider_task_id": task.ID,
		"reference_media_asset_ids": input.ReferenceMediaAssetIDs,
	})
	return s.Ingestor.Ingest(ctx, media.IngestInput{
		Owner: owner,
		MediaType: media.TypeVideo,
		ContentType: downloaded.ContentType,
		ContentLength: downloaded.ContentLength,
		Body: downloaded.Body,
		DurationMs: task.DurationMs,
		Metadata: metadata,
	})
}

func (s Service) sleep(ctx context.Context, duration time.Duration) error {
	if s.Sleep != nil { return s.Sleep(ctx, duration) }
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done(): return ctx.Err()
	case <-timer.C: return nil
	}
}
