package httpapi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
)

const agentVideoResultMaxBytes = 512 << 20

// agentVideoAssetIngestor moves provider output into the project's private
// asset store. Browser URLs are never used as the provider-to-storage path.
type agentVideoAssetIngestor struct {
	assets  agentworkspace.AssetStore
	objects agentworkspace.AssetObjectStore
	fetch   func(context.Context, string) (io.ReadCloser, string, error)
	now     func() time.Time
}

func newAgentVideoAssetIngestor(assets agentworkspace.AssetStore, objects agentworkspace.AssetObjectStore, client *http.Client, now func() time.Time) *agentVideoAssetIngestor {
	if client == nil {
		client = http.DefaultClient
	}
	if now == nil {
		now = time.Now
	}
	return &agentVideoAssetIngestor{assets: assets, objects: objects, now: now, fetch: func(ctx context.Context, rawURL string) (io.ReadCloser, string, error) {
		parsed, err := url.Parse(strings.TrimSpace(rawURL))
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return nil, "", fmt.Errorf("视频结果地址必须是 HTTPS")
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
		if err != nil {
			return nil, "", err
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, "", err
		}
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			resp.Body.Close()
			return nil, "", fmt.Errorf("读取视频结果失败: %s", resp.Status)
		}
		if resp.Request == nil || resp.Request.URL == nil || resp.Request.URL.Scheme != "https" || resp.Request.URL.Host == "" {
			resp.Body.Close()
			return nil, "", fmt.Errorf("视频结果重定向后必须保持 HTTPS")
		}
		return resp.Body, resp.Header.Get("Content-Type"), nil
	}}
}

func (s *agentVideoAssetIngestor) SaveRemote(ctx context.Context, owner, projectID, rawURL string) (agentworkspace.Asset, error) {
	if s == nil || s.fetch == nil {
		return agentworkspace.Asset{}, fmt.Errorf("视频结果保存器不可用")
	}
	body, contentType, err := s.fetch(ctx, rawURL)
	if err != nil {
		return agentworkspace.Asset{}, err
	}
	defer body.Close()
	return s.save(ctx, owner, projectID, "生成视频.mp4", contentType, body)
}

func (s *agentVideoAssetIngestor) SaveLocal(ctx context.Context, owner, projectID, artifactID string, executor *localexecutor.Service, files *localartifact.Store) (agentworkspace.Asset, error) {
	if executor == nil || files == nil {
		return agentworkspace.Asset{}, fmt.Errorf("本地视频结果存储不可用")
	}
	artifact, err := executor.GetArtifact(ctx, owner, strings.TrimSpace(artifactID))
	if err != nil {
		return agentworkspace.Asset{}, err
	}
	body, err := files.Open(artifact.StorageRef)
	if err != nil {
		return agentworkspace.Asset{}, err
	}
	defer body.Close()
	return s.save(ctx, owner, projectID, artifact.ID+".mp4", artifact.MediaType, body)
}

func (s *agentVideoAssetIngestor) save(ctx context.Context, owner, projectID, name, contentType string, body io.Reader) (agentworkspace.Asset, error) {
	if s == nil || s.assets == nil || s.objects == nil {
		return agentworkspace.Asset{}, fmt.Errorf("项目素材存储尚未配置")
	}
	contents, err := io.ReadAll(io.LimitReader(body, agentVideoResultMaxBytes+1))
	if err != nil || len(contents) > agentVideoResultMaxBytes {
		return agentworkspace.Asset{}, fmt.Errorf("视频结果读取失败或超过大小限制")
	}
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if contentType == "" {
		contentType = http.DetectContentType(contents)
	}
	if contentType != "video/mp4" {
		return agentworkspace.Asset{}, fmt.Errorf("视频结果必须是 MP4")
	}
	assetID := newAgentID("asset")
	key, err := agentworkspace.AgentAssetObjectKey(owner, projectID, assetID, path.Ext(name))
	if err != nil {
		return agentworkspace.Asset{}, err
	}
	asset := agentworkspace.Asset{ID: assetID, ProjectID: projectID, Kind: "video", Name: path.Base(name), ContentType: contentType, ObjectKey: key, ByteSize: int64(len(contents)), CreatedAt: s.now().UTC()}
	if err := s.objects.Put(ctx, asset, bytes.NewReader(contents)); err != nil {
		return agentworkspace.Asset{}, fmt.Errorf("保存视频对象: %w", err)
	}
	if err := s.assets.CreateAsset(ctx, owner, asset); err != nil {
		_ = s.objects.Delete(ctx, asset)
		return agentworkspace.Asset{}, err
	}
	asset.URL = agentAssetContentURL(projectID, asset.ID)
	return asset, nil
}
