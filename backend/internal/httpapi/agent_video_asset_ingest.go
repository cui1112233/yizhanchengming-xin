package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
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
const agentImageResultMaxBytes = 32 << 20

// agentVideoAssetIngestor moves provider output into the project's private
// asset store. Browser URLs are never used as the provider-to-storage path.
type agentMediaAssetIngestor struct {
	assets  agentworkspace.AssetStore
	objects agentworkspace.AssetObjectStore
	fetch   func(context.Context, string) (io.ReadCloser, string, error)
	now     func() time.Time
}

type agentVideoAssetIngestor = agentMediaAssetIngestor

func newAgentMediaAssetIngestor(assets agentworkspace.AssetStore, objects agentworkspace.AssetObjectStore, client *http.Client, now func() time.Time) *agentMediaAssetIngestor {
	return newAgentVideoAssetIngestor(assets, objects, client, now)
}

func newAgentVideoAssetIngestor(assets agentworkspace.AssetStore, objects agentworkspace.AssetObjectStore, client *http.Client, now func() time.Time) *agentMediaAssetIngestor {
	if client == nil {
		client = http.DefaultClient
	}
	if now == nil {
		now = time.Now
	}
	return &agentMediaAssetIngestor{assets: assets, objects: objects, now: now, fetch: func(ctx context.Context, rawURL string) (io.ReadCloser, string, error) {
		parsed, err := url.Parse(strings.TrimSpace(rawURL))
		if err != nil {
			return nil, "", fmt.Errorf("视频结果地址无效")
		}
		if err := validateAgentVideoRemoteURL(ctx, parsed); err != nil {
			return nil, "", err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
		if err != nil {
			return nil, "", err
		}
		resp, err := agentVideoSafeHTTPClient(client).Do(req)
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

func agentVideoSafeHTTPClient(base *http.Client) *http.Client {
	copyClient := *base
	var transport *http.Transport
	if base.Transport == nil {
		transport = http.DefaultTransport.(*http.Transport).Clone()
	} else if configured, ok := base.Transport.(*http.Transport); ok {
		transport = configured.Clone()
	}
	if transport != nil {
		dialer := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
		transport.Proxy = nil
		transport.DialTLSContext = nil
		transport.DialContext = newAgentVideoSafeDialContext(net.DefaultResolver.LookupIPAddr, dialer.DialContext)
		copyClient.Transport = transport
	}
	previousCheck := base.CheckRedirect
	copyClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if err := validateAgentVideoRemoteURL(req.Context(), req.URL); err != nil {
			return err
		}
		if previousCheck != nil {
			return previousCheck(req, via)
		}
		if len(via) >= 10 {
			return fmt.Errorf("视频结果重定向次数过多")
		}
		return nil
	}
	return &copyClient
}

func newAgentVideoSafeDialContext(resolve func(context.Context, string) ([]net.IPAddr, error), dial func(context.Context, string, string) (net.Conn, error)) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, fmt.Errorf("视频结果地址无效: %w", err)
		}
		addresses, err := resolve(ctx, host)
		if err != nil || len(addresses) == 0 {
			return nil, fmt.Errorf("视频结果地址无法解析")
		}
		for _, resolved := range addresses {
			if unsafeAgentVideoIP(resolved.IP) {
				return nil, fmt.Errorf("视频结果地址不能指向内网")
			}
		}
		return dial(ctx, network, net.JoinHostPort(addresses[0].IP.String(), port))
	}
}

func validateAgentVideoRemoteURL(ctx context.Context, parsed *url.URL) error {
	if parsed == nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.Hostname() == "" {
		return fmt.Errorf("视频结果地址必须是 HTTPS")
	}
	host := parsed.Hostname()
	if ip := net.ParseIP(host); ip != nil {
		if unsafeAgentVideoIP(ip) {
			return fmt.Errorf("视频结果地址不能指向内网")
		}
		return nil
	}
	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil || len(addresses) == 0 {
		return fmt.Errorf("视频结果地址无法解析")
	}
	for _, address := range addresses {
		if unsafeAgentVideoIP(address.IP) {
			return fmt.Errorf("视频结果地址不能指向内网")
		}
	}
	return nil
}

func unsafeAgentVideoIP(ip net.IP) bool {
	return ip == nil || ip.IsUnspecified() || ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
}

func (s *agentMediaAssetIngestor) SaveRemote(ctx context.Context, owner, projectID, assetID, rawURL string) (agentworkspace.Asset, error) {
	if s == nil || s.fetch == nil {
		return agentworkspace.Asset{}, fmt.Errorf("视频结果保存器不可用")
	}
	body, contentType, err := s.fetch(ctx, rawURL)
	if err != nil {
		return agentworkspace.Asset{}, err
	}
	defer body.Close()
	return s.save(ctx, owner, projectID, assetID, "video", "生成视频.mp4", contentType, agentVideoResultMaxBytes, body)
}

func (s *agentMediaAssetIngestor) SaveImageResult(ctx context.Context, owner, projectID, assetID, result string) (agentworkspace.Asset, error) {
	result = strings.TrimSpace(result)
	if strings.HasPrefix(result, "data:") {
		comma := strings.IndexByte(result, ',')
		if comma < 0 {
			return agentworkspace.Asset{}, fmt.Errorf("图片结果数据无效")
		}
		header := strings.ToLower(result[:comma])
		if !strings.HasPrefix(header, "data:image/") || !strings.HasSuffix(header, ";base64") {
			return agentworkspace.Asset{}, fmt.Errorf("图片结果必须是 base64 图片")
		}
		contents, err := base64.StdEncoding.DecodeString(result[comma+1:])
		if err != nil || len(contents) == 0 {
			return agentworkspace.Asset{}, fmt.Errorf("图片结果数据无效")
		}
		contentType := strings.TrimSuffix(strings.TrimPrefix(header, "data:"), ";base64")
		return s.save(ctx, owner, projectID, assetID, "image", "生成图片"+imageExtension(contentType), contentType, agentImageResultMaxBytes, bytes.NewReader(contents))
	}
	if s == nil || s.fetch == nil {
		return agentworkspace.Asset{}, fmt.Errorf("图片结果保存器不可用")
	}
	body, contentType, err := s.fetch(ctx, result)
	if err != nil {
		return agentworkspace.Asset{}, err
	}
	defer body.Close()
	return s.save(ctx, owner, projectID, assetID, "image", "生成图片"+imageExtension(contentType), contentType, agentImageResultMaxBytes, body)
}

func imageExtension(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0])) {
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".png"
	}
}

func (s *agentMediaAssetIngestor) SaveLocal(ctx context.Context, owner, projectID, assetID, artifactID string, executor *localexecutor.Service, files *localartifact.Store) (agentworkspace.Asset, error) {
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
	return s.save(ctx, owner, projectID, assetID, "video", artifact.ID+".mp4", artifact.MediaType, agentVideoResultMaxBytes, body)
}

func (s *agentMediaAssetIngestor) save(ctx context.Context, owner, projectID, assetID, kind, name, contentType string, maxBytes int64, body io.Reader) (agentworkspace.Asset, error) {
	if s == nil || s.assets == nil || s.objects == nil {
		return agentworkspace.Asset{}, fmt.Errorf("项目素材存储尚未配置")
	}
	contents, err := io.ReadAll(io.LimitReader(body, maxBytes+1))
	if err != nil || int64(len(contents)) > maxBytes {
		return agentworkspace.Asset{}, fmt.Errorf("%s结果读取失败或超过大小限制", map[string]string{"image": "图片", "video": "视频"}[kind])
	}
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	detectedType := strings.ToLower(strings.TrimSpace(strings.Split(http.DetectContentType(contents), ";")[0]))
	if kind == "image" {
		if contentType == "image/jpg" {
			contentType = "image/jpeg"
		}
		if detectedType != "image/png" && detectedType != "image/jpeg" && detectedType != "image/gif" && detectedType != "image/webp" {
			return agentworkspace.Asset{}, fmt.Errorf("生成结果媒体类型无效")
		}
		if contentType != "" && contentType != detectedType {
			return agentworkspace.Asset{}, fmt.Errorf("图片结果内容与媒体类型不一致")
		}
		contentType = detectedType
	} else if contentType == "" {
		contentType = detectedType
	}
	if (kind == "video" && contentType != "video/mp4") || (kind == "image" && !strings.HasPrefix(contentType, "image/")) {
		return agentworkspace.Asset{}, fmt.Errorf("生成结果媒体类型无效")
	}
	assetID = strings.TrimSpace(assetID)
	key, err := agentworkspace.AgentAssetObjectKey(owner, projectID, assetID, path.Ext(name))
	if err != nil {
		return agentworkspace.Asset{}, err
	}
	asset := agentworkspace.Asset{ID: assetID, ProjectID: projectID, Kind: kind, Name: path.Base(name), ContentType: contentType, ObjectKey: key, ByteSize: int64(len(contents)), CreatedAt: s.now().UTC()}
	if err := s.objects.Put(ctx, asset, bytes.NewReader(contents)); err != nil {
		return agentworkspace.Asset{}, fmt.Errorf("保存生成对象: %w", err)
	}
	if err := s.assets.CreateAsset(ctx, owner, asset); err != nil {
		if errors.Is(err, agentworkspace.ErrAssetExists) {
			existing, lookupErr := s.assets.Asset(ctx, owner, projectID, assetID)
			if lookupErr == nil && existing.Kind == kind && existing.ObjectKey == asset.ObjectKey {
				existing.URL = agentAssetContentURL(projectID, existing.ID)
				return existing, nil
			}
			return agentworkspace.Asset{}, err
		}
		_ = s.objects.Delete(ctx, asset)
		return agentworkspace.Asset{}, err
	}
	asset.URL = agentAssetContentURL(projectID, asset.ID)
	return asset, nil
}
