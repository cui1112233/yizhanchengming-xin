package n8comic

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/safehttp"
)

const (
	uploadPath = "/material/open/carousel/image_upload_param"
	readPath   = "/material/open/carousel/image/read"
	createPath = "/material/open/carousel/create"
)

type Config struct {
	BaseURL      string
	APIKey       string
	HTTPClient   *http.Client
	Now          func() time.Time
	PollInterval time.Duration
	PollAttempts int
}

type Client struct {
	baseURL      string
	apiKey       string
	httpClient   *http.Client
	now          func() time.Time
	pollInterval time.Duration
	pollAttempts int
}

type ImageInput struct {
	FileName  string
	MediaType string
	Content   []byte
}

type PublishRequest struct {
	ProducerAdminName string
	Name              string
	Images            []ImageInput
	WorkIDs           []int64
	AudioCategoryID   int64
	AudioID           int64
	// UploadedImageIDs 用于“图片已上传、合集未创建完成”的续建场景。长度必须
	// 与 Images 一致，下标对应同一页；大于 0 的位置直接复用，0 的位置重新上传。
	UploadedImageIDs []int64
}

// UploadImagesRequest 是分片上传的入参。分片只负责把图片送入 N8 并拿回
// image_id，不创建合集；合集由 CreateCarousel 用全部 image_id 一次创建。
type UploadImagesRequest struct {
	ProducerAdminName string
	Images            []ImageInput
	// UploadedImageIDs 与 Images 等长，大于 0 的位置表示该页已有图片 ID，
	// 本轮跳过重传。用于同一分片重复提交时的幂等。
	UploadedImageIDs []int64
}

// CreateCarouselRequest 复用 N8 的合集创建能力，不包含图片上传。
type CreateCarouselRequest struct {
	ProducerAdminName string
	Name              string
	ImageIDs          []int64
	WorkIDs           []int64
	AudioCategoryID   int64
	AudioID           int64
}

type PublishResult struct {
	CarouselID int64   `json:"carouselId"`
	ImageIDs   []int64 `json:"imageIds"`
}

type APIError struct {
	Code    string
	Message string
	Body    string
}

func (e *APIError) Error() string {
	detail := ""
	if strings.TrimSpace(e.Body) != "" {
		detail = "；响应JSON：" + e.Body
	}
	if strings.TrimSpace(e.Message) == "" {
		return "N8 接口返回错误：" + e.Code + detail
	}
	return "N8 接口返回错误：" + e.Message + "（" + e.Code + "）" + detail
}

func NewClient(config Config) *Client {
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = safehttp.NewClientWithTimeout(2 * time.Minute)
	}
	now := config.Now
	if now == nil {
		now = time.Now
	}
	pollInterval := config.PollInterval
	if pollInterval <= 0 {
		pollInterval = time.Second
	}
	pollAttempts := config.PollAttempts
	if pollAttempts <= 0 {
		// TOS 回调到 N8 入库是异步的，30 秒在图片较大或回调拥堵时容易误报。
		pollAttempts = 90
	}
	return &Client{
		baseURL:      strings.TrimRight(strings.TrimSpace(config.BaseURL), "/"),
		apiKey:       strings.TrimSpace(config.APIKey),
		httpClient:   httpClient,
		now:          now,
		pollInterval: pollInterval,
		pollAttempts: pollAttempts,
	}
}

func (c *Client) Configured() bool {
	return c != nil && c.baseURL != "" && c.apiKey != ""
}

func (c *Client) Publish(ctx context.Context, input PublishRequest) (PublishResult, error) {
	if !c.Configured() {
		return PublishResult{}, fmt.Errorf("N8 漫画发布接口未配置")
	}
	input.ProducerAdminName = strings.TrimSpace(input.ProducerAdminName)
	input.Name = strings.TrimSpace(input.Name)
	if input.ProducerAdminName == "" || len([]rune(input.ProducerAdminName)) > 64 {
		return PublishResult{}, fmt.Errorf("制作者账号不能为空且不能超过 64 个字符")
	}
	if input.Name == "" || len([]rune(input.Name)) > 64 {
		return PublishResult{}, fmt.Errorf("漫画合集名称不能为空且不能超过 64 个字符")
	}
	if len(input.Images) < 2 || len(input.Images) > 35 {
		return PublishResult{}, fmt.Errorf("漫画合集必须包含 2 到 35 张图片")
	}
	workIDs, err := uniquePositiveIDs(input.WorkIDs, "关联作品")
	if err != nil {
		return PublishResult{}, err
	}
	if len(workIDs) == 0 {
		return PublishResult{}, fmt.Errorf("至少选择一个关联作品")
	}
	if input.AudioCategoryID <= 0 || input.AudioID <= 0 {
		return PublishResult{}, fmt.Errorf("音频分类 ID 和指定音频 ID 必须大于 0")
	}

	if len(input.UploadedImageIDs) > 0 && len(input.UploadedImageIDs) != len(input.Images) {
		return PublishResult{}, fmt.Errorf("已上传图片记录与当前漫画页数不一致，不能复用")
	}

	// N8 文档要求：多张图先全部申请并上传，再统一查询 image/read。
	// 如果上传一张就立刻查询，TOS 回调还没完成时会把后续流程卡在
	// IMAGE_NOT_READY，导致合集发布看起来一直失败。
	imageIDs, uploadErr := c.UploadImages(ctx, UploadImagesRequest{ProducerAdminName: input.ProducerAdminName, Images: input.Images, UploadedImageIDs: input.UploadedImageIDs})
	if uploadErr != nil {
		// 返回部分成功的 imageIDs：调用方可以把已上传的页持久化，下一次只
		// 重试缺图和建合集，避免整本书 35 张图全部重传。
		return PublishResult{ImageIDs: imageIDs}, uploadErr
	}
	carouselID, err := c.CreateCarousel(ctx, CreateCarouselRequest{
		ProducerAdminName: input.ProducerAdminName,
		Name:              input.Name,
		ImageIDs:          imageIDs,
		WorkIDs:           workIDs,
		AudioCategoryID:   input.AudioCategoryID,
		AudioID:           input.AudioID,
	})
	if err != nil {
		return PublishResult{ImageIDs: imageIDs}, err
	}
	return PublishResult{CarouselID: carouselID, ImageIDs: imageIDs}, nil
}

// UploadImages 只执行图片上传与入库查询，返回与 Images 下标一致的 image_id。
// 上传失败时仍返回已成功的位置，0 表示该页未拿到 ID。
func (c *Client) UploadImages(ctx context.Context, input UploadImagesRequest) ([]int64, error) {
	if !c.Configured() {
		return nil, fmt.Errorf("N8 漫画发布接口未配置")
	}
	producer := strings.TrimSpace(input.ProducerAdminName)
	if producer == "" || len([]rune(producer)) > 64 {
		return nil, fmt.Errorf("制作者账号不能为空且不能超过 64 个字符")
	}
	if len(input.Images) < 1 || len(input.Images) > 35 {
		return nil, fmt.Errorf("漫画合集必须包含 2 到 35 张图片")
	}
	if len(input.UploadedImageIDs) > 0 && len(input.UploadedImageIDs) != len(input.Images) {
		return nil, fmt.Errorf("已上传图片记录与当前漫画页数不一致，不能复用")
	}
	return c.uploadImages(ctx, producer, input.Images, input.UploadedImageIDs)
}

// CreateCarousel 用已经上传完成的 image_id 创建漫画合集，返回合集 ID。
func (c *Client) CreateCarousel(ctx context.Context, input CreateCarouselRequest) (int64, error) {
	if !c.Configured() {
		return 0, fmt.Errorf("N8 漫画发布接口未配置")
	}
	producer := strings.TrimSpace(input.ProducerAdminName)
	name := strings.TrimSpace(input.Name)
	if producer == "" || len([]rune(producer)) > 64 {
		return 0, fmt.Errorf("制作者账号不能为空且不能超过 64 个字符")
	}
	if name == "" || len([]rune(name)) > 64 {
		return 0, fmt.Errorf("漫画合集名称不能为空且不能超过 64 个字符")
	}
	imageIDs, err := uniquePositiveIDs(input.ImageIDs, "合集图片")
	if err != nil {
		return 0, err
	}
	if len(imageIDs) < 2 || len(imageIDs) > 35 {
		return 0, fmt.Errorf("漫画合集必须包含 2 到 35 张图片")
	}
	workIDs, err := uniquePositiveIDs(input.WorkIDs, "关联作品")
	if err != nil {
		return 0, err
	}
	if len(workIDs) == 0 {
		return 0, fmt.Errorf("至少选择一个关联作品")
	}
	if input.AudioCategoryID <= 0 || input.AudioID <= 0 {
		return 0, fmt.Errorf("音频分类 ID 和指定音频 ID 必须大于 0")
	}

	payload := map[string]any{
		"producer_admin_name": producer,
		"name":                name,
		"image_ids":           imageIDs,
		"work_ids":            workIDs,
		"audio_category_id":   input.AudioCategoryID,
		"audio_id":            input.AudioID,
	}
	var created struct {
		CarouselID int64 `json:"carousel_id"`
	}
	if err := c.post(ctx, createPath, payload, &created); err != nil {
		return 0, err
	}
	if created.CarouselID <= 0 {
		return 0, fmt.Errorf("N8 创建漫画合集成功但未返回合集 ID")
	}
	return created.CarouselID, nil
}

type pendingImage struct {
	fileName  string
	signature string
	imageID   int64
	// target 是该页在 Images 原始数组中的下标；跳过的已上传页不占用 pending，
	// 因此必须显式记录，否则 imageIDs 的顺序会错位。
	target int
}

func (c *Client) uploadImages(ctx context.Context, producer string, inputs []ImageInput, uploaded []int64) ([]int64, error) {
	imageIDs := make([]int64, len(inputs))
	pending := make([]pendingImage, 0, len(inputs))
	for index, input := range inputs {
		if index < len(uploaded) && uploaded[index] > 0 {
			imageIDs[index] = uploaded[index]
			continue
		}
		item, err := c.prepareImageUpload(ctx, producer, input)
		if err != nil {
			// 已上传的页仍然保留在 imageIDs 里，便于调用方持久化后只补缺图。
			return imageIDs, err
		}
		item.target = index
		pending = append(pending, item)
	}

	for index := range pending {
		if pending[index].imageID > 0 {
			imageIDs[pending[index].target] = pending[index].imageID
		}
	}
	// N8 入库回调是异步的。逐张等待会让 35 张图最多串行等待 52 分钟，
	// 而自动发布阶段只有 10 分钟，最终会被误判成发布失败。并发等待但
	// 保留原数组下标，确保合集里的图片顺序不变。
	pendingIndexes := make([]int, 0, len(pending))
	for index, item := range pending {
		if item.imageID == 0 {
			pendingIndexes = append(pendingIndexes, index)
		}
	}
	workerCount := len(pendingIndexes)
	if workerCount > 5 {
		workerCount = 5
	}
	// Keep tiny publishes deterministic and avoid creating more work than the
	// request contains; large comic collections get the bounded parallelism.
	if workerCount <= 2 {
		workerCount = 1
	}
	if workerCount == 0 {
		return imageIDs, nil
	}
	jobs := make(chan int)
	var waitGroup sync.WaitGroup
	var errorMu sync.Mutex
	var firstErr error
	worker := func() {
		defer waitGroup.Done()
		for index := range jobs {
			imageID, err := c.waitForImage(ctx, pending[index].signature, pending[index].fileName)
			if err != nil {
				errorMu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				errorMu.Unlock()
				continue
			}
			imageIDs[pending[index].target] = imageID
		}
	}
	waitGroup.Add(workerCount)
	for index := 0; index < workerCount; index++ {
		go worker()
	}
	for _, index := range pendingIndexes {
		jobs <- index
	}
	close(jobs)
	waitGroup.Wait()
	if firstErr != nil {
		return imageIDs, firstErr
	}
	return imageIDs, nil
}

func (c *Client) prepareImageUpload(ctx context.Context, producer string, input ImageInput) (pendingImage, error) {
	fileName := strings.TrimSpace(filepath.Base(input.FileName))
	mediaType := strings.ToLower(strings.TrimSpace(input.MediaType))
	if fileName == "" || len([]rune(fileName)) > 200 {
		return pendingImage{}, fmt.Errorf("漫画图片文件名不能为空且不能超过 200 个字符")
	}
	if len(input.Content) == 0 {
		return pendingImage{}, fmt.Errorf("漫画图片 %s 内容为空", fileName)
	}
	// 2K 漫画页原图约 5–7MB，超过 N8/巨量单图体积上限会被判“图片过大”。
	// 只压缩超限的副本，本地保存的原图不动；压缩后同步切换文件名后缀和
	// media type，保证申请上传时声明的格式与真实内容一致。
	content, uploadType, err := compressImageForUpload(input.Content, mediaType)
	if err != nil {
		return pendingImage{}, fmt.Errorf("漫画图片 %s %w", fileName, err)
	}
	if uploadType != mediaType {
		fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName)) + imageFileExtension(uploadType)
		mediaType = uploadType
	}
	width, height, err := imageDimensions(content, mediaType)
	if err != nil {
		return pendingImage{}, fmt.Errorf("漫画图片 %s 无效：%w", fileName, err)
	}
	// N8 文档示例使用文件内容 MD5 作为 signature；该值同时用于 TOS
	// 路径和 image/read 查询，必须与回调侧使用的签名完全一致。
	digest := md5.Sum(content)
	signature := hex.EncodeToString(digest[:])
	var upload struct {
		URL    string            `json:"url"`
		Header map[string]string `json:"header"`
		Info   *struct {
			ID int64 `json:"id"`
		} `json:"info"`
	}
	if err := c.post(ctx, uploadPath, map[string]any{
		"producer_admin_name": producer,
		"file_name":           fileName,
		"signature":           signature,
		"file_width":          width,
		"file_height":         height,
		"mime_type":           mediaType,
	}, &upload); err != nil {
		return pendingImage{}, fmt.Errorf("申请上传 %s 失败：%w", fileName, err)
	}
	if upload.Info != nil && upload.Info.ID > 0 {
		return pendingImage{fileName: fileName, imageID: upload.Info.ID}, nil
	}
	if strings.TrimSpace(upload.URL) == "" {
		return pendingImage{}, fmt.Errorf("N8 未返回图片 %s 的上传地址", fileName)
	}
	if err := c.put(ctx, upload.URL, upload.Header, content); err != nil {
		return pendingImage{}, fmt.Errorf("上传图片 %s 失败：%w", fileName, err)
	}
	return pendingImage{fileName: fileName, signature: signature}, nil
}

func (c *Client) waitForImage(ctx context.Context, signature, fileName string) (int64, error) {
	var lastErr error
	for attempt := 0; attempt < c.pollAttempts; attempt++ {
		var result struct {
			ImageID int64 `json:"image_id"`
		}
		err := c.post(ctx, readPath, map[string]any{"signature": signature}, &result)
		if err == nil {
			if result.ImageID <= 0 {
				return 0, fmt.Errorf("N8 图片 %s 已就绪但未返回图片 ID", fileName)
			}
			return result.ImageID, nil
		}
		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.Code != "IMAGE_NOT_READY" {
			return 0, err
		}
		lastErr = err
		if attempt+1 == c.pollAttempts {
			break
		}
		timer := time.NewTimer(c.pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return 0, ctx.Err()
		case <-timer.C:
		}
	}
	if lastErr != nil {
		return 0, fmt.Errorf("等待 N8 图片 %s 入库超时（已等待 %d 秒，最后状态：%w）", fileName, c.pollAttempts, lastErr)
	}
	return 0, fmt.Errorf("等待 N8 图片 %s 入库超时", fileName)
}

func (c *Client) post(ctx context.Context, path string, payload map[string]any, target any) error {
	seconds := c.now().Unix()
	payload["time"] = seconds
	sum := md5.Sum([]byte(c.apiKey + strconv.FormatInt(seconds, 10)))
	payload["sign"] = hex.EncodeToString(sum[:])
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("N8 HTTP 状态码 %d", response.StatusCode)
	}
	var envelope struct {
		Code    string          `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(&envelope); err != nil {
		return fmt.Errorf("解析 N8 响应失败：%w", err)
	}
	if envelope.Code != "SUCCESS" {
		return &APIError{Code: envelope.Code, Message: envelope.Message, Body: compactJSON(envelope)}
	}
	if target == nil || len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return nil
	}
	if err := json.Unmarshal(envelope.Data, target); err != nil {
		return fmt.Errorf("解析 N8 数据失败：%w", err)
	}
	return nil
}

func compactJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func (c *Client) put(ctx context.Context, rawURL string, headers map[string]string, content []byte) error {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return fmt.Errorf("N8 返回了无效上传地址")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPut, parsed.String(), bytes.NewReader(content))
	if err != nil {
		return err
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("TOS HTTP 状态码 %d", response.StatusCode)
	}
	return nil
}

func uniquePositiveIDs(values []int64, label string) ([]int64, error) {
	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			return nil, fmt.Errorf("%s ID 必须大于 0", label)
		}
		if _, exists := seen[value]; exists {
			return nil, fmt.Errorf("%s ID 不能重复", label)
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result, nil
}

// imageFileExtension 返回与媒体类型匹配的文件后缀。转码后必须同步换后缀，
// 否则会出现「.png 实际是 JPEG」这种声明与实际不一致的上传。
func imageFileExtension(mediaType string) string {
	switch strings.ToLower(strings.TrimSpace(mediaType)) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/bmp":
		return ".bmp"
	default:
		return ".png"
	}
}

func imageDimensions(content []byte, mediaType string) (int, int, error) {
	switch mediaType {
	case "image/png", "image/jpeg", "image/gif":
		config, _, err := image.DecodeConfig(bytes.NewReader(content))
		if err != nil || config.Width < 1 || config.Height < 1 {
			return 0, 0, fmt.Errorf("无法读取图片尺寸")
		}
		return config.Width, config.Height, nil
	case "image/bmp":
		if len(content) < 26 || string(content[:2]) != "BM" {
			return 0, 0, fmt.Errorf("BMP 文件头无效")
		}
		width := int(int32(binary.LittleEndian.Uint32(content[18:22])))
		height := int(int32(binary.LittleEndian.Uint32(content[22:26])))
		if width < 0 {
			width = -width
		}
		if height < 0 {
			height = -height
		}
		if width < 1 || height < 1 {
			return 0, 0, fmt.Errorf("BMP 图片尺寸无效")
		}
		return width, height, nil
	default:
		return 0, 0, fmt.Errorf("N8 只接受 JPG、JPEG、GIF、BMP 或 PNG")
	}
}
