package subjectupload

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"qiantie/backend/internal/safehttp"
)

// RunError deliberately carries only safe progress and a fixed message. Provider
// response bodies and transport errors can contain credentials and are omitted.
type RunError struct {
	Kind     string
	Progress Progress
	Message  string
}

func (e *RunError) Error() string { return e.Message }
func fail(kind string, p Progress, msg string) error {
	return &RunError{Kind: kind, Progress: p, Message: msg}
}

func NewClient() *Client {
	h := safehttp.NewClient(safehttp.Options{Timeout: 45 * time.Second})
	h.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &Client{HTTP: h, FireEndpoint: "https://ark.cn-beijing.volcengineapi.com", DeyunaiEndpoint: "https://api.deyunai.com", Now: time.Now, PollInterval: 3 * time.Second, MaxWait: 180 * time.Second}
}

func (c *Client) Run(ctx context.Context, cfg Config, in Input, p Progress, checkpoint Checkpoint) (Result, error) {
	empty := Result{}
	if p.PendingAction != "" {
		return empty, fail("needs_review", p, "上次创建结果尚未确认，请先人工核对，不能重复创建")
	}
	if ctx.Err() != nil {
		return empty, fail("paused", p, "已停止等待")
	}
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	cfg.Provider = provider
	if strings.TrimSpace(in.Name) == "" {
		return empty, fail("failed", p, "请填写素材名称")
	}
	switch provider {
	case "volcano", "volcengine":
		if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" || cfg.ProjectName == "" {
			return empty, fail("failed", p, "请先配置火山项目名称和访问密钥")
		}
	case "deyunai":
		if cfg.Token == "" || (cfg.Region != "c39" && cfg.Region != "c40") {
			return empty, fail("failed", p, "请先配置得云密钥和c39或c40区域")
		}
	default:
		return empty, fail("failed", p, "当前素材上传渠道暂未实现")
	}
	if p.AssetID == "" && p.ImageURL != "" && !validURL(p.ImageURL) {
		return empty, fail("failed", p, "保存的图片地址无效")
	}
	if p.AssetID == "" && p.ImageURL == "" && in.ImageURL == "" && !validURL(cfg.ImageHostURL) {
		return empty, fail("failed", p, "缺少可用的公开HTTPS图片或图床配置")
	}
	max := c.MaxWait
	if max <= 0 || max > 180*time.Second {
		max = 180 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, max)
	defer cancel()
	save := func() error {
		if checkpoint != nil {
			if e := checkpoint(ctx, p); e != nil {
				return fail("paused", p, "进度保存失败，已停止后续请求")
			}
		}
		return nil
	}
	if p.AssetID == "" && p.ImageURL == "" {
		p.ImageURL = in.ImageURL
		if p.ImageURL != "" && !validURL(p.ImageURL) {
			return empty, fail("failed", p, "图片必须使用公开HTTPS地址")
		}
		if p.ImageURL == "" {
			s, e := c.hostImage(ctx, cfg, in)
			if e != nil {
				return empty, fail("paused", p, "图片上传失败，请核对图床配置后继续")
			}
			p.ImageURL = s
		}
		if e := save(); e != nil {
			return empty, e
		}
	}
	return c.runProvider(ctx, cfg, in, p, checkpoint)
}

func (c *Client) fire(ctx context.Context, cfg Config, action string, body any) (map[string]any, error) {
	b, e := json.Marshal(body)
	if e != nil {
		return nil, e
	}
	endpoint := c.FireEndpoint
	if endpoint == "" {
		endpoint = NewClient().FireEndpoint
	}
	endpoint = strings.TrimRight(endpoint, "/") + "/?Action=" + action + "&Version=2024-01-01"
	r, e := http.NewRequest(http.MethodPost, endpoint, nil)
	if e != nil {
		return nil, e
	}
	now := time.Now()
	if c.Now != nil {
		now = c.Now()
	}
	signFire(r, b, cfg, now)
	headers := map[string]string{}
	for k, v := range r.Header {
		headers[k] = v[0]
	}
	data, e := c.request(ctx, http.MethodPost, endpoint, b, headers)
	if e != nil {
		return nil, e
	}
	if object(object(data, "ResponseMetadata"), "Error") != nil {
		return nil, fmt.Errorf("provider rejected request")
	}
	return object(data, "Result"), nil
}
func (c *Client) runProvider(ctx context.Context, cfg Config, in Input, p Progress, cp Checkpoint) (Result, error) {
	empty := Result{}
	fire := cfg.Provider == "volcano" || cfg.Provider == "volcengine"
	save := func() error {
		if cp != nil {
			if e := cp(ctx, p); e != nil {
				return fail("paused", p, "进度保存失败，已停止后续请求")
			}
		}
		return nil
	}
	create := func(action string, call func() (map[string]any, error)) (map[string]any, error) {
		if ctx.Err() != nil {
			return nil, fail("paused", p, "已停止等待")
		}
		p.PendingAction = action
		if e := save(); e != nil {
			return nil, e
		}
		data, e := call()
		if e != nil {
			return nil, fail("needs_review", p, "创建结果无法确认，请先人工核对，不能重复创建")
		}
		if str(data, "Id") == "" {
			return nil, fail("needs_review", p, "创建响应未返回素材标识，请先人工核对")
		}
		return data, nil
	}
	if fire && p.GroupID == "" {
		p.GroupID = in.GroupID
	}
	if fire && p.AssetID == "" && p.GroupID == "" {
		data, e := create("create_group", func() (map[string]any, error) {
			return c.fire(ctx, cfg, "CreateAssetGroup", struct{ Name, Description, GroupType, ProjectName string }{in.Name, "Character asset group for " + in.Name, "AIGC", cfg.ProjectName})
		})
		if e != nil {
			return empty, e
		}
		p.GroupID = str(data, "Id")
		p.PendingAction = ""
		if e := save(); e != nil {
			return empty, e
		}
	}
	var data map[string]any
	deyEndpoint := strings.TrimRight(c.DeyunaiEndpoint, "/")
	if deyEndpoint == "" {
		deyEndpoint = NewClient().DeyunaiEndpoint
	}
	deyEndpoint += "/" + cfg.Region + "/v1/sd/assets"
	headers := map[string]string{"Content-Type": "application/json", "Accept": "application/json", "Authorization": "Bearer " + cfg.Token}
	if p.AssetID == "" {
		var e error
		data, e = create("create_asset", func() (map[string]any, error) {
			if fire {
				name := filepath.Base(strings.ReplaceAll(in.FileName, "\\", "/"))
				if name == "" || name == "." {
					name = "image.png"
				}
				return c.fire(ctx, cfg, "CreateAsset", struct{ GroupId, URL, AssetType, ProjectName, Name string }{p.GroupID, p.ImageURL, "Image", cfg.ProjectName, name})
			}
			// The shipped client reads description but its observed request omits it.
			body := struct{ URL, Name, AssetType string }{p.ImageURL, in.Name, "Image"}
			b, _ := json.Marshal(body)
			return c.request(ctx, "POST", deyEndpoint, b, headers)
		})
		if e != nil {
			return empty, e
		}
		p.AssetID = str(data, "Id")
		p.PendingAction = ""
		p.Status = str(data, "Status")
		if e := save(); e != nil {
			return empty, e
		}
	}
	for {
		if ctx.Err() != nil {
			return empty, fail("paused", p, "已停止等待，可使用现有素材标识继续查询")
		}
		if data == nil || fire {
			var e error
			if fire {
				data, e = c.fire(ctx, cfg, "GetAsset", struct{ Id, ProjectName string }{p.AssetID, cfg.ProjectName})
			} else {
				data, e = c.request(ctx, "GET", deyEndpoint+"/"+url.PathEscape(p.AssetID), nil, map[string]string{"Accept": "application/json", "Authorization": "Bearer " + cfg.Token})
			}
			if e != nil {
				return empty, fail("paused", p, "素材查询失败，可使用现有素材标识继续查询")
			}
		}
		status := str(data, "Status")
		p.Status = status
		if e := save(); e != nil {
			return empty, e
		}
		switch strings.ToLower(status) {
		case "active":
			project := cfg.ProjectName
			if !fire {
				project = "deyunai"
			}
			return Result{AssetID: p.AssetID, AssetURI: "asset://" + p.AssetID, GroupID: p.GroupID, ProjectName: project, Provider: cfg.Provider, Status: "Active"}, nil
		case "failed":
			return empty, fail("failed", p, "供应商素材处理失败")
		case "processing", "pending", "":
		default:
			return empty, fail("paused", p, "供应商返回尚未识别的素材状态，请核对后继续查询")
		}
		delay := c.PollInterval
		if delay <= 0 {
			delay = 3 * time.Second
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return empty, fail("paused", p, "已停止等待，可使用现有素材标识继续查询")
		case <-timer.C:
		}
		data = nil
	}
}

func validURL(raw string) bool {
	u, e := url.Parse(raw)
	if e != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
		return false
	}
	if ip := net.ParseIP(u.Hostname()); ip != nil && safehttp.IsBlockedIP(ip) {
		return false
	}
	h := strings.ToLower(u.Hostname())
	return h != "localhost" && !strings.HasSuffix(h, ".localhost") && !strings.HasSuffix(h, ".local")
}
func (c *Client) request(ctx context.Context, method, target string, body []byte, headers map[string]string) (map[string]any, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("request stopped")
	}
	req, err := http.NewRequestWithContext(ctx, method, target, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("invalid target")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	h := c.HTTP
	if h == nil {
		h = NewClient().HTTP
	}
	copyClient := *h
	copyClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := copyClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("provider request failed")
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20+1))
	if err != nil || len(raw) > 1<<20 {
		return nil, fmt.Errorf("provider response exceeds limit or cannot be read")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("provider returned HTTP %d", resp.StatusCode)
	}
	var out map[string]any
	if json.Unmarshal(raw, &out) != nil || out == nil {
		return nil, fmt.Errorf("provider returned invalid JSON")
	}
	return out, nil
}
func str(m map[string]any, k string) string            { s, _ := m[k].(string); return strings.TrimSpace(s) }
func object(m map[string]any, k string) map[string]any { v, _ := m[k].(map[string]any); return v }
func (c *Client) hostImage(ctx context.Context, cfg Config, in Input) (string, error) {
	if len(in.Image) == 0 || len(in.Image) > 40<<20 {
		return "", fmt.Errorf("图片为空或超过40MiB")
	}
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	name := filepath.Base(strings.ReplaceAll(in.FileName, "\\", "/"))
	if name == "." || name == "" {
		name = "image.png"
	}
	ct := in.ContentType
	if !strings.HasPrefix(ct, "image/") {
		ct = http.DetectContentType(in.Image)
	}
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, strings.NewReplacer("\\", "_", "\"", "_", "\r", "_", "\n", "_").Replace(name)))
	header.Set("Content-Type", ct)
	part, e := w.CreatePart(header)
	if e != nil {
		return "", e
	}
	_, _ = part.Write(in.Image)
	_ = w.Close()
	headers := map[string]string{"Content-Type": w.FormDataContentType()}
	if cfg.ImageHostToken != "" {
		headers["Authorization"] = "Bearer " + cfg.ImageHostToken
	}
	data, e := c.request(ctx, http.MethodPost, cfg.ImageHostURL, b.Bytes(), headers)
	if e != nil {
		return "", e
	}
	result := str(data, "url")
	if result == "" {
		result = str(object(data, "data"), "url")
	}
	if !validURL(result) {
		return "", fmt.Errorf("图床未返回公开HTTPS图片地址")
	}
	return result, nil
}
func digest(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func mac(key []byte, s string) []byte {
	h := hmac.New(sha256.New, key)
	_, _ = h.Write([]byte(s))
	return h.Sum(nil)
}
func signFire(req *http.Request, body []byte, cfg Config, now time.Time) {
	stamp := now.UTC().Format("20060102T150405Z")
	day := stamp[:8]
	hash := digest(body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Date", stamp)
	req.Header.Set("X-Content-Sha256", hash)
	signed := "content-type;host;x-content-sha256;x-date"
	canonical := "POST\n/\n" + req.URL.Query().Encode() + "\ncontent-type:application/json\nhost:" + req.URL.Host + "\nx-content-sha256:" + hash + "\nx-date:" + stamp + "\n\n" + signed + "\n" + hash
	scope := day + "/cn-beijing/ark/request"
	key := mac(mac(mac(mac([]byte(cfg.SecretAccessKey), day), "cn-beijing"), "ark"), "request")
	sig := hex.EncodeToString(mac(key, "HMAC-SHA256\n"+stamp+"\n"+scope+"\n"+digest([]byte(canonical))))
	req.Header.Set("Authorization", "HMAC-SHA256 Credential="+cfg.AccessKeyID+"/"+scope+", SignedHeaders="+signed+", Signature="+sig)
}
