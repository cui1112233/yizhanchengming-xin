package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/n8catalog"
	"qiantie/backend/internal/n8comic"
)

const n8ComicPublishDraftKind = "n8-comic-publish"

// errN8ComicCarouselIncomplete 表示漫画图片已经上传（imageIDs 可直接复用），
// 但 N8 合集还没有创建完成。它是可续建状态：既不能当成发布成功，也不是
// 内容审核失败，重试时只需要复用 imageIDs 重新建合集。
var errN8ComicCarouselIncomplete = errors.New("N8 漫画图片已上传，但合集尚未创建完成")

type n8ComicPublishSettings struct {
	ProducerAdminName          string  `json:"producerAdminName"`
	OrganizationID             string  `json:"organizationId"`
	AudioCategoryID            int64   `json:"audioCategoryId"`
	AudioCategoryName          string  `json:"audioCategoryName"`
	AudioCategorySelectionMode string  `json:"audioCategorySelectionMode"`
	AudioID                    int64   `json:"audioId"`
	SpecifiedAudioID           int64   `json:"specifiedAudioId"`
	SpecifiedAudioName         string  `json:"specifiedAudioName"`
	WorkIDs                    []int64 `json:"workIds"`
	WorkIDsText                string  `json:"workIdsText"`
	WorkMatchMode              string  `json:"workMatchMode"`
}

type N8ComicClient interface {
	Configured() bool
	Publish(context.Context, n8comic.PublishRequest) (n8comic.PublishResult, error)
}

func (p *BatchFactoryN8ComicPublisher) PublishFromSettings(ctx context.Context, owner, batchID, bookID string, collection batchfactoryv11.ComicCollection, settings batchfactoryv11.SettingsPatch) (N8ComicPublishRecord, error) {
	config, err := parseN8ComicPublishSettings(settings)
	if err != nil {
		return N8ComicPublishRecord{}, err
	}
	if len(config.WorkIDs) == 0 {
		for _, token := range strings.FieldsFunc(config.WorkIDsText, func(r rune) bool { return r == ',' || r == '，' || r == ' ' || r == '\n' || r == '\t' }) {
			value, err := strconv.ParseInt(strings.TrimSpace(token), 10, 64)
			if err != nil || value <= 0 {
				return N8ComicPublishRecord{}, fmt.Errorf("漫画发布的作品 ID 无效")
			}
			config.WorkIDs = append(config.WorkIDs, value)
		}
	}
	if len(config.WorkIDs) == 0 {
		if strings.TrimSpace(config.WorkMatchMode) == "" || config.WorkMatchMode == "title_auto" {
			resolved, err := p.resolveComicWorkIDs(ctx, owner, batchID, bookID)
			if err != nil {
				return N8ComicPublishRecord{}, err
			}
			config.WorkIDs = resolved
		} else {
			return N8ComicPublishRecord{}, fmt.Errorf("尚未匹配到漫画后台作品 ID，请先在漫画合集里确认关联作品")
		}
	}
	selection, err := p.resolveBookComicAudio(ctx, owner, batchID, bookID, config)
	if err != nil {
		return N8ComicPublishRecord{}, err
	}
	return p.Publish(ctx, owner, batchID, bookID, collection.ID, n8ComicPublishInput{
		ProducerAdminName: config.ProducerAdminName,
		AudioCategoryID:   selection.CategoryID,
		AudioID:           selection.AudioID,
		WorkIDs:           config.WorkIDs,
	})
}

func parseN8ComicPublishSettings(settings batchfactoryv11.SettingsPatch) (n8ComicPublishSettings, error) {
	raw, ok := settings["comicPublishSettings"]
	if !ok {
		return n8ComicPublishSettings{}, fmt.Errorf("请先在统一配置中填写漫画发布设置")
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return n8ComicPublishSettings{}, fmt.Errorf("漫画发布设置格式无效")
	}
	for _, key := range []string{"audioCategoryId", "audioId", "specifiedAudioId"} {
		value, exists := payload[key]
		if !exists || value == nil || strings.TrimSpace(fmt.Sprint(value)) == "" {
			payload[key] = int64(0)
			continue
		}
		if text, isText := value.(string); isText {
			parsed, err := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
			if err != nil {
				return n8ComicPublishSettings{}, fmt.Errorf("漫画发布设置格式无效")
			}
			payload[key] = parsed
		}
	}
	normalized, err := json.Marshal(payload)
	if err != nil {
		return n8ComicPublishSettings{}, fmt.Errorf("漫画发布设置格式无效")
	}
	var config n8ComicPublishSettings
	if err := json.Unmarshal(normalized, &config); err != nil {
		return n8ComicPublishSettings{}, fmt.Errorf("漫画发布设置格式无效")
	}
	return config, nil
}

// ValidateSettings only reads catalog data. It is used before automation starts
// so a missing default audio cannot waste image-generation work.
func (p *BatchFactoryN8ComicPublisher) ValidateSettings(ctx context.Context, settings batchfactoryv11.SettingsPatch) error {
	config, err := parseN8ComicPublishSettings(settings)
	if err != nil {
		return err
	}
	if strings.TrimSpace(config.OrganizationID) == "" {
		return fmt.Errorf("漫画发布设置不完整：请选择组织")
	}
	manualAudioID := config.AudioID
	if manualAudioID <= 0 && strings.EqualFold(strings.TrimSpace(config.AudioCategorySelectionMode), "manual") {
		manualAudioID = config.SpecifiedAudioID
	}
	categories, warning := p.audioCategoriesWithWarning(ctx)
	if warning != "" {
		return n8DirectoryError(errors.New(warning))
	}
	if manualAudioID <= 0 {
		for _, category := range categories {
			if category.ID > 0 && strings.EqualFold(strings.TrimSpace(category.Status), "ENABLE") {
				return nil
			}
		}
		return fmt.Errorf("漫画发布设置不完整：没有可用指定音频分类")
	}
	category, err := enabledComicAudioCategory(categories, config.AudioCategoryID, config.AudioCategoryName)
	if err != nil {
		return err
	}
	audios, err := p.audios(ctx, strconv.FormatInt(category.ID, 10), config.OrganizationID)
	if err != nil {
		return err
	}
	_, err = enabledComicAudio(audios, manualAudioID)
	return err
}

// ValidateBookSettings verifies the exact category and audio that a book would
// use without choosing a random audio or persisting comicAudioSelection.
func (p *BatchFactoryN8ComicPublisher) ValidateBookSettings(ctx context.Context, owner, batchID, bookID string, settings batchfactoryv11.SettingsPatch) error {
	config, err := parseN8ComicPublishSettings(settings)
	if err != nil {
		return err
	}
	if strings.TrimSpace(config.OrganizationID) == "" {
		return fmt.Errorf("漫画发布设置不完整：请选择组织")
	}
	if p == nil || p.Store == nil {
		return fmt.Errorf("漫画发布存储尚未配置")
	}
	batch, err := p.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return err
	}
	book, err := comicAudioBook(batch, bookID)
	if err != nil {
		return err
	}
	manualAudioID := config.AudioID
	if manualAudioID <= 0 && strings.EqualFold(strings.TrimSpace(config.AudioCategorySelectionMode), "manual") {
		manualAudioID = config.SpecifiedAudioID
	}
	categoryName := strings.TrimSpace(config.AudioCategoryName)
	categoryID := config.AudioCategoryID
	if manualAudioID <= 0 {
		categoryName, err = comicAudioCategoryName(book)
		if err != nil {
			return err
		}
		categoryID = 0
	}
	categories, warning := p.audioCategoriesWithWarning(ctx)
	if warning != "" {
		return n8DirectoryError(errors.New(warning))
	}
	category, err := enabledComicAudioCategory(categories, categoryID, categoryName)
	if err != nil {
		return err
	}
	audios, err := p.audios(ctx, strconv.FormatInt(category.ID, 10), config.OrganizationID)
	if err != nil {
		return err
	}
	if manualAudioID > 0 {
		if _, err = enabledComicAudio(audios, manualAudioID); err != nil {
			return err
		}
	} else {
		available := false
		for _, audio := range audios {
			if audio.ID > 0 && strings.EqualFold(strings.TrimSpace(audio.Status), "ENABLE") {
				available = true
				break
			}
		}
		if !available {
			return fmt.Errorf("当前组织的“%s”分类没有可用音频", category.Name)
		}
	}
	// 书名 → 作品 ID 的匹配也放进预检：书名对不上时要在生成前拦住，
	// 不能等整本图片都出完、上传完才在发布阶段失败。
	if !comicPublishHasExplicitWorkIDs(config) && comicPublishUsesTitleMatch(config.WorkMatchMode) {
		if _, err := p.resolveComicWorkIDs(ctx, owner, batchID, bookID); err != nil {
			return err
		}
	}
	return nil
}

func comicPublishHasExplicitWorkIDs(config n8ComicPublishSettings) bool {
	if len(config.WorkIDs) > 0 {
		return true
	}
	return strings.TrimSpace(config.WorkIDsText) != ""
}

func comicPublishUsesTitleMatch(mode string) bool {
	mode = strings.TrimSpace(mode)
	return mode == "" || mode == "title_auto"
}

func (p *BatchFactoryN8ComicPublisher) resolveComicPublishAudio(ctx context.Context, categoryID, audioID int64, organizationID string) (int64, int64, error) {
	if categoryID <= 0 {
		for _, category := range p.audioCategories(ctx) {
			if category.ID > 0 && strings.EqualFold(strings.TrimSpace(category.Status), "ENABLE") {
				categoryID = category.ID
				break
			}
		}
		if categoryID <= 0 {
			return 0, 0, fmt.Errorf("漫画发布设置不完整：没有可用指定音频分类")
		}
	}
	if audioID > 0 {
		return categoryID, audioID, nil
	}
	audios, err := p.audios(ctx, strconv.FormatInt(categoryID, 10), organizationID)
	if err != nil {
		return 0, 0, fmt.Errorf("读取指定音频失败：%w", err)
	}
	if len(audios) == 0 {
		return 0, 0, fmt.Errorf("漫画发布设置不完整：没有可用指定音频")
	}
	return categoryID, audios[0].ID, nil
}

// n8DirectoryError 把 N8 目录读取失败归类为上游不可用，并附带稳定错误码，
// 避免这类本可解释的失败在统一出口退化成 500 internal error。
func n8DirectoryError(err error) error {
	if err == nil {
		return nil
	}
	return withErrorCode(fmt.Errorf("%w: %s", batchfactoryv11.ErrUpstream, err.Error()), ErrorCodeN8DirectoryRejected)
}

func (p *BatchFactoryN8ComicPublisher) resolveComicWorkIDs(ctx context.Context, owner, batchID, bookID string) ([]int64, error) {
	if p == nil || p.Store == nil || p.WorksDirectory == nil || !p.WorksDirectory.Configured() {
		return nil, n8DirectoryError(errors.New("漫画后台作品目录未配置，无法按书名自动匹配作品 ID"))
	}
	batch, err := p.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return nil, err
	}
	bookTitle := ""
	for _, book := range batch.Books {
		if book.ID == bookID {
			bookTitle = strings.TrimSpace(book.Title)
			break
		}
	}
	if bookTitle == "" {
		return nil, fmt.Errorf("找不到当前漫画的书名，无法自动匹配漫画后台作品")
	}
	target := normalizeComicWorkTitle(bookTitle)
	// 先查本地缓存。命中时完全不访问 N8，因此 N8 修改作品接口不会影响
	// 已经匹配过的书。
	if cached := p.workTitleMap().lookup(target); len(cached) > 0 {
		return cached, nil
	}
	works, err := p.WorksDirectory.SearchWorks(ctx, bookTitle)
	if err != nil {
		return nil, n8DirectoryError(fmt.Errorf("按书名查询漫画后台作品失败：%w", err))
	}
	matched := make([]int64, 0, 1)
	for _, work := range works {
		if work.ID > 0 && normalizeComicWorkTitle(work.Name) == target {
			matched = append(matched, work.ID)
		}
	}
	if len(matched) == 0 && len(works) == 1 && works[0].ID > 0 {
		// The N8 directory already applies its safe title-fragment fallback.
		// When it returns exactly one candidate, use it for unattended publish;
		// ambiguous results still require a manual confirmation in the UI.
		matched = append(matched, works[0].ID)
	}
	if len(matched) == 0 {
		return nil, fmt.Errorf("漫画后台未找到与“%s”完全同名的作品，请先在漫画合集里确认关联作品", bookTitle)
	}
	// 匹配成功后写入本地缓存，下一次发布会直接命中。
	p.workTitleMap().store(target, matched)
	return matched, nil
}

func normalizeComicWorkTitle(value string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\r', '\n', '，', ',', '。', '.', '：', ':', '！', '!', '？', '?', '、', '“', '”', '"', '‘', '’', '\'', '（', '(', '）', ')', '【', '[', '】', ']', '—', '-', '_':
			return -1
		default:
			return r
		}
	}, strings.TrimSpace(value))
}

type N8ComicProducerDirectory interface {
	Configured() bool
	ListProducerAccounts(context.Context) ([]string, error)
}

type N8ComicProducerAccountDirectory interface {
	Configured() bool
	ListProducerAccountDirectory(context.Context) ([]n8catalog.ProducerAccount, error)
}

type N8ComicOrganizationDirectory interface {
	Configured() bool
	ListOrganizations(context.Context) ([]n8catalog.Organization, error)
}

type N8ComicAudioCategoryDirectory interface {
	Configured() bool
	ListAudioCategories(context.Context) ([]n8catalog.AudioCategory, error)
}

type N8ComicAudioDirectory interface {
	Configured() bool
	ListAudio(context.Context, string) ([]n8catalog.Audio, error)
}

type N8ComicWorkDirectory interface {
	Configured() bool
	SearchWorks(context.Context, string) ([]n8catalog.Work, error)
}

type N8ComicPublishRecord struct {
	Status            string    `json:"status"`
	Name              string    `json:"name"`
	ProducerAdminName string    `json:"producerAdminName"`
	AudioCategoryID   int64     `json:"audioCategoryId"`
	WorkIDs           []int64   `json:"workIds"`
	CarouselID        int64     `json:"carouselId"`
	ImageIDs          []int64   `json:"imageIds"`
	PublishedAt       time.Time `json:"publishedAt"`
}

type BatchFactoryN8ComicPublisher struct {
	Store                 batchfactoryv11.Store
	LocalFiles            *localartifact.Store
	Client                N8ComicClient
	ProducerDirectory     N8ComicProducerDirectory
	OrganizationDirectory N8ComicOrganizationDirectory
	WorksDirectory        N8ComicWorkDirectory
	ProducerAccounts      []string
	Now                   func() time.Time
	RandomIndex           func(upper int) (int, error)
	// WorkTitleMapPath 是本地“书名 -> N8 作品 ID”缓存路径，为空时使用默认路径。
	WorkTitleMapPath string
	mu               sync.Mutex
	titleMapOnce     sync.Once
	titleMap         *n8WorkTitleMap

	// N8 目录一次读取要 10 秒以上（账号目录要翻页），而统一配置弹窗每次打开
	// 都会取一遍。这里给成功结果加短 TTL 缓存，避免每次打开都卡在空候选上。
	directoryCacheMu sync.Mutex
	directoryCache   map[string]n8DirectoryCacheEntry
}

const n8DirectoryCacheTTL = 5 * time.Minute

type n8DirectoryCacheEntry struct {
	value     any
	expiresAt time.Time
}

type n8ProducerAccountsCache struct {
	names  []string
	source string
}

type n8ProducerDirectorySnapshot struct {
	names         []string
	source        string
	organizations map[string][]string
}

func (p *BatchFactoryN8ComicPublisher) cachedDirectory(key string) (any, bool) {
	p.directoryCacheMu.Lock()
	defer p.directoryCacheMu.Unlock()
	entry, ok := p.directoryCache[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.value, true
}

func (p *BatchFactoryN8ComicPublisher) storeDirectory(key string, value any) {
	p.directoryCacheMu.Lock()
	defer p.directoryCacheMu.Unlock()
	if p.directoryCache == nil {
		p.directoryCache = map[string]n8DirectoryCacheEntry{}
	}
	p.directoryCache[key] = n8DirectoryCacheEntry{value: value, expiresAt: time.Now().Add(n8DirectoryCacheTTL)}
}

// workTitleMap 惰性初始化本地书名映射，避免所有构造点都必须显式赋值。
func (p *BatchFactoryN8ComicPublisher) workTitleMap() *n8WorkTitleMap {
	p.titleMapOnce.Do(func() {
		p.titleMap = newN8WorkTitleMap(p.WorkTitleMapPath)
	})
	return p.titleMap
}

type n8ComicPublishInput struct {
	ProducerAdminName string  `json:"producerAdminName"`
	AudioCategoryID   int64   `json:"audioCategoryId"`
	AudioID           int64   `json:"audioId"`
	WorkIDs           []int64 `json:"workIds"`
}

func registerN8ComicPublishRoutes(mux *http.ServeMux, publisher *BatchFactoryN8ComicPublisher) {
	if mux == nil || publisher == nil {
		return
	}
	mux.HandleFunc("GET /api/batch-factory/v12/comic-publish/status", func(w http.ResponseWriter, r *http.Request) {
		// 这几个目录来自同一个 N8 后台，但不能串行读取。制作者账号目录
		// 偶尔会慢到 20 秒以上；串行读取会让音频分类也跟着一直转圈。
		// 并行读取后，最慢的那一路不会再阻塞其它下拉框。
		var (
			audioCategories        []n8catalog.AudioCategory
			audioCategoriesWarning string
			source, warning        string
			producerOrganizations  map[string][]string
			accountNames           []string
			organizationItems      []n8catalog.Organization
			organizationsWarning   string
		)
		var directoryWG sync.WaitGroup
		directoryWG.Add(3)
		go func() {
			defer directoryWG.Done()
			accountNames, source, warning, producerOrganizations = publisher.producerDirectorySnapshot(r.Context())
		}()
		go func() {
			defer directoryWG.Done()
			organizationItems, organizationsWarning = publisher.organizationsWithWarning(r.Context())
		}()
		go func() {
			defer directoryWG.Done()
			audioCategories, audioCategoriesWarning = publisher.audioCategoriesWithWarning(r.Context())
		}()
		directoryWG.Wait()
		publishKeyConfigured := publisher.Client != nil && publisher.Client.Configured()
		// directoryAvailable 表示“目录现在真的能读到数据”，而不是“配置开关打开”。
		// 只有令牌存在且三路目录都没有报错时才为 true；缺发布密钥不会影响它。
		directoryAvailable := publisher.ProducerDirectory != nil && publisher.ProducerDirectory.Configured() &&
			organizationsWarning == "" && audioCategoriesWarning == "" && warning == ""
		writeJSON(w, http.StatusOK, map[string]any{
			// configured 保持原含义：N8 漫画发布客户端（open api key）是否可用。
			// 它与目录可用性不是同一个开关，前端不得再用它判断目录告警。
			"configured":                   publishKeyConfigured,
			"directoryAvailable":           directoryAvailable,
			"publishKeyConfigured":         publishKeyConfigured,
			"publishKeyHint":               "QIANTIE_N8_COMIC_OPEN_API_KEY",
			"producerAccounts":             accountNames,
			"producerAccountsSource":       source,
			"producerAccountsWarning":      warning,
			"organizations":                organizationItems,
			"organizationsWarning":         organizationsWarning,
			"producerAccountOrganizations": producerOrganizations,
			"audioCategories":              audioCategories,
			"audioCategoriesWarning":       audioCategoriesWarning,
		})
	})
	mux.HandleFunc("GET /api/batch-factory/v12/comic-publish/audio-categories", func(w http.ResponseWriter, r *http.Request) {
		items, warning := publisher.audioCategoriesWithWarning(r.Context())
		writeJSON(w, http.StatusOK, map[string]any{"audioCategories": items, "audioCategoriesWarning": warning})
	})
	mux.HandleFunc("GET /api/batch-factory/v12/comic-publish/audios", func(w http.ResponseWriter, r *http.Request) {
		categoryID, organizationID := strings.TrimSpace(r.URL.Query().Get("audioCategoryId")), strings.TrimSpace(r.URL.Query().Get("organizationId"))
		if categoryID == "" || organizationID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "audioCategoryId 和 organizationId 均为必填项"})
			return
		}
		items, err := publisher.audios(r.Context(), categoryID, organizationID)
		if err != nil {
			if errors.Is(err, batchfactoryv11.ErrUpstream) {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": n8DirectoryWarning("音频目录", err)})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"audios": items})
	})
	mux.HandleFunc("GET /api/batch-factory/v12/comic-publish/works", func(w http.ResponseWriter, r *http.Request) {
		if publisher.WorksDirectory == nil || !publisher.WorksDirectory.Configured() {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "N8 作品查询接口尚未配置管理员令牌"})
			return
		}
		query := strings.TrimSpace(r.URL.Query().Get("query"))
		if query == "" {
			writeJSON(w, http.StatusOK, map[string]any{"works": []n8catalog.Work{}})
			return
		}
		works, err := publisher.WorksDirectory.SearchWorks(r.Context(), query)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "查询漫画后台作品失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"works": works})
	})
	mux.HandleFunc("POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/comic-collections/{collectionId}/publish", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input n8ComicPublishInput
		if !decodeJSON(w, r, &input) {
			return
		}
		record, err := publisher.Publish(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("collectionId"), input)
		if err != nil {
			if errors.Is(err, batchfactoryv11.ErrUpstream) {
				writeStoreError(w, err)
				return
			}
			if err == batchfactoryv11.ErrNotFound || err == batchfactoryv11.ErrInvalid {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"publish": record})
	})
	mux.HandleFunc("GET /api/batch-factory/v12/batches/{batchId}/books/{bookId}/comic-collections/{collectionId}/publish", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if publisher.Store == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "漫画发布记录存储尚未配置"})
			return
		}
		record, found := publisher.savedRecord(r.Context(), owner, r.PathValue("batchId"), r.PathValue("collectionId"))
		if !found {
			writeJSON(w, http.StatusOK, map[string]any{"published": false, "publish": nil})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"published": true, "publish": record})
	})
}

func (p *BatchFactoryN8ComicPublisher) audios(ctx context.Context, categoryID, organizationID string) ([]n8catalog.Audio, error) {
	directory, ok := p.ProducerDirectory.(N8ComicAudioDirectory)
	if !ok || directory == nil || !directory.Configured() {
		return nil, n8DirectoryError(errors.New("N8 音频目录尚未配置"))
	}
	items, err := directory.ListAudio(ctx, strings.TrimSpace(organizationID))
	if err != nil {
		return nil, n8DirectoryError(err)
	}
	wanted, _ := strconv.ParseInt(strings.TrimSpace(categoryID), 10, 64)
	filtered := make([]n8catalog.Audio, 0, len(items))
	for _, item := range items {
		if wanted > 0 && !containsInt64(item.CategoryIDs, wanted) {
			continue
		}
		if strings.TrimSpace(organizationID) != "" && strconv.FormatInt(item.OrganizationID, 10) != strings.TrimSpace(organizationID) {
			continue
		}
		if !strings.EqualFold(item.Status, "ENABLE") {
			continue
		}
		filtered = append(filtered, item)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if parseN8AudioUpdatedAt(filtered[i].UpdatedAt).Equal(parseN8AudioUpdatedAt(filtered[j].UpdatedAt)) {
			return filtered[i].ID > filtered[j].ID
		}
		return parseN8AudioUpdatedAt(filtered[i].UpdatedAt).After(parseN8AudioUpdatedAt(filtered[j].UpdatedAt))
	})
	return filtered, nil
}

func parseN8AudioUpdatedAt(value string) time.Time {
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05", "2006-01-02 15:04:05.999999"} {
		if parsed, err := time.Parse(layout, strings.TrimSpace(value)); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

func containsInt64(values []int64, target int64) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// n8DirectoryFailureDetail 把上游错误压成一段可展示的诊断。HTTP 状态与
// 上游业务 code 必须保留，否则“令牌失效”会被前端误判成“目录本来就是空的”。
// 这里只使用状态码、业务 code/message，不包含令牌等敏感值。
func n8DirectoryFailureDetail(err error) string {
	var failure *n8catalog.DirectoryError
	if !errors.As(err, &failure) {
		return ""
	}
	if failure.Status != 0 {
		return fmt.Sprintf("HTTP %d", failure.Status)
	}
	if code := strings.TrimSpace(failure.Code); code != "" {
		detail := "上游 code=" + code
		if message := strings.TrimSpace(failure.Message); message != "" {
			runes := []rune(message)
			if len(runes) > 60 {
				message = string(runes[:60])
			}
			detail += "：" + message
		}
		return detail
	}
	switch failure.Kind {
	case "connection":
		return "无法连接上游"
	case "invalid_response":
		return "上游返回格式无法解析"
	}
	return ""
}

func n8DirectoryWarning(label string, err error) string {
	if detail := n8DirectoryFailureDetail(err); detail != "" {
		return fmt.Sprintf("N8 %s读取失败（%s），请更新后台登录令牌后重试。", label, detail)
	}
	return fmt.Sprintf("N8 %s读取失败，请更新后台登录令牌后重试。", label)
}

func (p *BatchFactoryN8ComicPublisher) audioCategories(ctx context.Context) []n8catalog.AudioCategory {
	items, _ := p.audioCategoriesWithWarning(ctx)
	return items
}

// audioCategoriesWithWarning 在读取失败时同时返回可展示的原因，避免前端把
// 令牌失效误判成“这个组织本来就没有音频分类”。
func (p *BatchFactoryN8ComicPublisher) audioCategoriesWithWarning(ctx context.Context) ([]n8catalog.AudioCategory, string) {
	if cached, ok := p.cachedDirectory("audioCategories"); ok {
		return cached.([]n8catalog.AudioCategory), ""
	}
	directory, ok := p.ProducerDirectory.(N8ComicAudioCategoryDirectory)
	if !ok || directory == nil || !directory.Configured() {
		log.Printf("N8 audio category directory unavailable: not configured")
		return nil, ""
	}
	items, err := directory.ListAudioCategories(ctx)
	if err != nil {
		log.Printf("N8 audio category directory unavailable: %v", err)
		return nil, n8DirectoryWarning("音频分类目录", err)
	}
	log.Printf("N8 audio category directory returned %d enabled categories", len(items))
	p.storeDirectory("audioCategories", items)
	return items, ""
}

func (p *BatchFactoryN8ComicPublisher) producerOrganizations(ctx context.Context) map[string][]string {
	if cached, ok := p.cachedDirectory("producerOrganizations"); ok {
		return cached.(map[string][]string)
	}
	directory, ok := p.ProducerDirectory.(N8ComicProducerAccountDirectory)
	if !ok || directory == nil || !directory.Configured() {
		return map[string][]string{}
	}
	entries, err := directory.ListProducerAccountDirectory(ctx)
	if err != nil {
		return map[string][]string{}
	}
	result := make(map[string][]string, len(entries))
	for _, entry := range entries {
		name := strings.TrimSpace(entry.Name)
		if name == "" || len(entry.OrganizationIDs) == 0 {
			continue
		}
		result[name] = append([]string(nil), entry.OrganizationIDs...)
	}
	p.storeDirectory("producerOrganizations", result)
	return result
}

// producerDirectorySnapshot 一次翻页同时得到账号名和账号所属组织。
// N8 的 ListProducerAccounts 本身也是通过 ListProducerAccountDirectory 翻页；
// 分开调用会把同一份慢目录完整读取两遍。
func (p *BatchFactoryN8ComicPublisher) producerDirectorySnapshot(ctx context.Context) ([]string, string, string, map[string][]string) {
	if cached, ok := p.cachedDirectory("producerDirectorySnapshot"); ok {
		entry := cached.(n8ProducerDirectorySnapshot)
		return append([]string(nil), entry.names...), entry.source, "", cloneProducerOrganizations(entry.organizations)
	}
	fallback := uniqueProducerAccounts(p.ProducerAccounts)
	emptyOrganizations := map[string][]string{}
	if p.ProducerDirectory == nil || !p.ProducerDirectory.Configured() {
		return fallback, "cache", "", emptyOrganizations
	}
	if directory, ok := p.ProducerDirectory.(N8ComicProducerAccountDirectory); ok {
		entries, err := directory.ListProducerAccountDirectory(ctx)
		if err == nil && len(entries) > 0 {
			names := make([]string, 0, len(entries))
			organizations := make(map[string][]string, len(entries))
			for _, entry := range entries {
				name := strings.TrimSpace(entry.Name)
				if name == "" {
					continue
				}
				names = append(names, name)
				if len(entry.OrganizationIDs) > 0 {
					organizations[name] = append([]string(nil), entry.OrganizationIDs...)
				}
			}
			names = uniqueProducerAccounts(names)
			snapshot := n8ProducerDirectorySnapshot{names: names, source: "n8", organizations: organizations}
			p.storeDirectory("producerDirectorySnapshot", snapshot)
			return names, "n8", "", organizations
		}
		return producerDirectoryFallback(fallback, err)
	}
	accounts, err := p.ProducerDirectory.ListProducerAccounts(ctx)
	if err == nil && len(accounts) > 0 {
		names := uniqueProducerAccounts(accounts)
		snapshot := n8ProducerDirectorySnapshot{names: names, source: "n8", organizations: emptyOrganizations}
		p.storeDirectory("producerDirectorySnapshot", snapshot)
		return names, "n8", "", emptyOrganizations
	}
	return producerDirectoryFallback(fallback, err)
}

func producerDirectoryFallback(fallback []string, err error) ([]string, string, string, map[string][]string) {
	if err != nil {
		var failure *n8catalog.DirectoryError
		if errors.As(err, &failure) {
			log.Printf("N8 producer directory failed: kind=%s status=%d", failure.Kind, failure.Status)
		} else {
			log.Print("N8 producer directory failed: kind=unknown")
		}
	}
	if len(fallback) > 0 {
		warning := "N8 制作者目录实时同步暂不可用，当前使用最近同步的账号。"
		if detail := n8DirectoryFailureDetail(err); detail != "" {
			warning = "N8 制作者目录实时同步暂不可用（" + detail + "），当前使用最近同步的账号。"
		}
		return fallback, "cache", warning, map[string][]string{}
	}
	if err != nil {
		return nil, "unavailable", n8DirectoryWarning("制作者目录", err), map[string][]string{}
	}
	return nil, "unavailable", "N8 没有返回启用中的制作者账号。", map[string][]string{}
}

func cloneProducerOrganizations(source map[string][]string) map[string][]string {
	result := make(map[string][]string, len(source))
	for name, organizations := range source {
		result[name] = append([]string(nil), organizations...)
	}
	return result
}

func (p *BatchFactoryN8ComicPublisher) organizations(ctx context.Context) []n8catalog.Organization {
	items, _ := p.organizationsWithWarning(ctx)
	return items
}

// organizationsWithWarning 在读取失败时同时返回可展示的原因，避免前端把
// 令牌失效误判成“当前组织没有账号”。
func (p *BatchFactoryN8ComicPublisher) organizationsWithWarning(ctx context.Context) ([]n8catalog.Organization, string) {
	if cached, ok := p.cachedDirectory("organizations"); ok {
		return cached.([]n8catalog.Organization), ""
	}
	directory := p.OrganizationDirectory
	if directory == nil {
		if candidate, ok := p.ProducerDirectory.(N8ComicOrganizationDirectory); ok {
			directory = candidate
		}
	}
	if directory == nil || !directory.Configured() {
		return nil, ""
	}
	organizations, err := directory.ListOrganizations(ctx)
	if err != nil {
		log.Printf("N8 organization directory unavailable: %v", err)
		return nil, n8DirectoryWarning("组织目录", err)
	}
	p.storeDirectory("organizations", organizations)
	return organizations, ""
}

func (p *BatchFactoryN8ComicPublisher) producerAccounts(ctx context.Context) ([]string, string, string) {
	if cached, ok := p.cachedDirectory("producerAccounts"); ok {
		entry := cached.(n8ProducerAccountsCache)
		return entry.names, entry.source, ""
	}
	fallback := uniqueProducerAccounts(p.ProducerAccounts)
	if p.ProducerDirectory == nil || !p.ProducerDirectory.Configured() {
		return fallback, "cache", ""
	}
	accounts, err := p.ProducerDirectory.ListProducerAccounts(ctx)
	if err == nil && len(accounts) > 0 {
		names := uniqueProducerAccounts(accounts)
		p.storeDirectory("producerAccounts", n8ProducerAccountsCache{names: names, source: "n8"})
		return names, "n8", ""
	}
	if err != nil {
		var failure *n8catalog.DirectoryError
		if errors.As(err, &failure) {
			log.Printf("N8 producer directory failed: kind=%s status=%d", failure.Kind, failure.Status)
		} else {
			log.Print("N8 producer directory failed: kind=unknown")
		}
	} else {
		log.Print("N8 producer directory returned no accounts")
	}
	if len(fallback) > 0 {
		warning := "N8 制作者目录实时同步暂不可用，当前使用最近同步的账号。"
		if detail := n8DirectoryFailureDetail(err); detail != "" {
			warning = "N8 制作者目录实时同步暂不可用（" + detail + "），当前使用最近同步的账号。"
		}
		return fallback, "cache", warning
	}
	if err != nil {
		return nil, "unavailable", n8DirectoryWarning("制作者目录", err)
	}
	return nil, "unavailable", "N8 没有返回启用中的制作者账号。"
}

func uniqueProducerAccounts(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func (p *BatchFactoryN8ComicPublisher) Publish(ctx context.Context, owner, batchID, bookID, collectionID string, input n8ComicPublishInput) (N8ComicPublishRecord, error) {
	if p == nil || p.Store == nil || p.LocalFiles == nil || p.Client == nil || !p.Client.Configured() {
		return N8ComicPublishRecord{}, fmt.Errorf("N8 漫画发布接口尚未配置")
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	if saved, ok := p.savedRecord(ctx, owner, batchID, collectionID); ok {
		return saved, nil
	}
	input.ProducerAdminName = strings.TrimSpace(input.ProducerAdminName)
	missing := make([]string, 0, 4)
	if input.ProducerAdminName == "" {
		missing = append(missing, "制作者")
	}
	if input.AudioCategoryID <= 0 {
		missing = append(missing, "音频分类")
	}
	if input.AudioID <= 0 {
		missing = append(missing, "音频")
	}
	if len(input.WorkIDs) == 0 {
		missing = append(missing, "关联作品")
	}
	if len(missing) > 0 {
		return N8ComicPublishRecord{}, fmt.Errorf("漫画发布设置不完整：缺少%s", strings.Join(missing, "、"))
	}

	batch, err := p.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return N8ComicPublishRecord{}, err
	}
	var book *batchfactoryv11.Book
	for index := range batch.Books {
		if batch.Books[index].ID == bookID {
			book = &batch.Books[index]
			break
		}
	}
	if book == nil {
		return N8ComicPublishRecord{}, batchfactoryv11.ErrNotFound
	}
	comicStore, ok := p.Store.(batchfactoryv11.ComicPageStore)
	if !ok {
		return N8ComicPublishRecord{}, fmt.Errorf("漫画合集存储尚未配置")
	}
	collections, err := comicStore.ListComicCollections(ctx, owner, batchID, bookID)
	if err != nil {
		return N8ComicPublishRecord{}, err
	}
	var collection *batchfactoryv11.ComicCollection
	for index := range collections {
		if collections[index].ID == collectionID {
			collection = &collections[index]
			break
		}
	}
	if collection == nil {
		return N8ComicPublishRecord{}, batchfactoryv11.ErrNotFound
	}
	items := append([]batchfactoryv11.ComicCollectionItem(nil), collection.Items...)
	sort.SliceStable(items, func(i, j int) bool { return items[i].Ordinal < items[j].Ordinal })
	images := make([]n8comic.ImageInput, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.StorageRef) == "" {
			return N8ComicPublishRecord{}, fmt.Errorf("漫画第 %d 页没有可上传的本地主图", item.Ordinal)
		}
		file, openErr := p.LocalFiles.Open(item.StorageRef)
		if openErr != nil {
			return N8ComicPublishRecord{}, fmt.Errorf("读取漫画第 %d 页失败：%w", item.Ordinal, openErr)
		}
		content, readErr := io.ReadAll(io.LimitReader(file, 64<<20))
		closeErr := file.Close()
		if readErr != nil {
			return N8ComicPublishRecord{}, fmt.Errorf("读取漫画第 %d 页失败：%w", item.Ordinal, readErr)
		}
		if closeErr != nil {
			return N8ComicPublishRecord{}, fmt.Errorf("关闭漫画第 %d 页文件失败：%w", item.Ordinal, closeErr)
		}
		images = append(images, n8comic.ImageInput{FileName: fmt.Sprintf("%s-%02d%s", collection.ID, item.Ordinal, comicImageExtension(item.MediaType)), MediaType: item.MediaType, Content: content})
	}

	now := time.Now()
	if p.Now != nil {
		now = p.Now()
	}
	name, err := n8ComicCollectionName(input.ProducerAdminName, book.Title, now)
	if err != nil {
		return N8ComicPublishRecord{}, err
	}
	// 上一次可能已经上传了全部/部分图片但建合集失败。复用这些 imageIDs，
	// 本轮只上传缺图并重新创建合集，避免整本图片重传。
	var uploadedIDs []int64
	if saved, ok := p.savedUploadedRecord(ctx, owner, batchID, collectionID); ok && len(saved.ImageIDs) == len(images) {
		uploadedIDs = saved.ImageIDs
	}
	result, err := p.Client.Publish(ctx, n8comic.PublishRequest{ProducerAdminName: input.ProducerAdminName, Name: name, Images: images, WorkIDs: input.WorkIDs, AudioCategoryID: input.AudioCategoryID, AudioID: input.AudioID, UploadedImageIDs: uploadedIDs})
	if err != nil {
		if uploaded := countComicPublishImageIDs(result.ImageIDs); uploaded > 0 {
			// 只要已经有一部分图片入库，就把它们持久化，标记为独立终态
			// upload_incomplete；不能返回 nil，否则自动化会把“合集未创建”
			// 误记成发布完成。
			record := N8ComicPublishRecord{Status: "upload_incomplete", Name: name, ProducerAdminName: input.ProducerAdminName, AudioCategoryID: input.AudioCategoryID, WorkIDs: append([]int64(nil), input.WorkIDs...), ImageIDs: append([]int64(nil), result.ImageIDs...), PublishedAt: now}
			encoded, encodeErr := json.Marshal(record)
			if encodeErr != nil {
				return N8ComicPublishRecord{}, encodeErr
			}
			if _, saveErr := p.Store.SaveDraft(ctx, owner, batchfactoryv11.Draft{Key: collectionID, Kind: n8ComicPublishDraftKind, Scope: batchID, Content: string(encoded)}); saveErr != nil {
				return N8ComicPublishRecord{}, fmt.Errorf("N8 图片已上传，但保存上传记录失败：%w", saveErr)
			}
			return record, fmt.Errorf("%w：%v", errN8ComicCarouselIncomplete, err)
		}
		return N8ComicPublishRecord{}, err
	}
	record := N8ComicPublishRecord{Status: "published", Name: name, ProducerAdminName: input.ProducerAdminName, AudioCategoryID: input.AudioCategoryID, WorkIDs: append([]int64(nil), input.WorkIDs...), CarouselID: result.CarouselID, ImageIDs: append([]int64(nil), result.ImageIDs...), PublishedAt: now}
	encoded, err := json.Marshal(record)
	if err != nil {
		return N8ComicPublishRecord{}, err
	}
	if _, err := p.Store.SaveDraft(ctx, owner, batchfactoryv11.Draft{Key: collectionID, Kind: n8ComicPublishDraftKind, Scope: batchID, Content: string(encoded)}); err != nil {
		return N8ComicPublishRecord{}, fmt.Errorf("N8 已发布，但保存本地发布记录失败：%w", err)
	}
	return record, nil
}

func comicImageExtension(mediaType string) string {
	switch strings.ToLower(strings.TrimSpace(mediaType)) {
	case "image/jpeg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/bmp":
		return ".bmp"
	default:
		return ".png"
	}
}

// savedRecord 只对真正发布成功的记录短路。uploaded/upload_incomplete 必须
// 继续走后面的建合集流程，否则会把“图片已上传但合集不存在”当成发布完成。
func (p *BatchFactoryN8ComicPublisher) savedRecord(ctx context.Context, owner, batchID, collectionID string) (N8ComicPublishRecord, bool) {
	draft, err := p.Store.GetDraft(ctx, owner, collectionID, n8ComicPublishDraftKind, batchID)
	if err != nil {
		return N8ComicPublishRecord{}, false
	}
	var record N8ComicPublishRecord
	if json.Unmarshal([]byte(draft.Content), &record) != nil || record.Status != "published" || record.CarouselID <= 0 || len(record.ImageIDs) == 0 {
		return N8ComicPublishRecord{}, false
	}
	return record, true
}

// savedUploadedRecord 读取上次“图片已上传、合集未完成”的记录，用于续建时
// 复用 imageIDs。兼容历史上写入的 uploaded 记录。
func (p *BatchFactoryN8ComicPublisher) savedUploadedRecord(ctx context.Context, owner, batchID, collectionID string) (N8ComicPublishRecord, bool) {
	draft, err := p.Store.GetDraft(ctx, owner, collectionID, n8ComicPublishDraftKind, batchID)
	if err != nil {
		return N8ComicPublishRecord{}, false
	}
	var record N8ComicPublishRecord
	if json.Unmarshal([]byte(draft.Content), &record) != nil || (record.Status != "uploaded" && record.Status != "upload_incomplete") || len(record.ImageIDs) == 0 {
		return N8ComicPublishRecord{}, false
	}
	return record, true
}

// countComicPublishImageIDs 统计已经拿到 N8 图片 ID 的页数；0 表示还没
// 有任何图片入库，不需要写 upload_incomplete 记录。
func countComicPublishImageIDs(ids []int64) int {
	count := 0
	for _, id := range ids {
		if id > 0 {
			count++
		}
	}
	return count
}

func n8ComicCollectionName(producer, title string, now time.Time) (string, error) {
	producer = strings.TrimSpace(producer)
	title = strings.TrimSpace(title)
	if producer == "" || title == "" {
		return "", batchfactoryv11.ErrInvalid
	}
	chinaTime := now.In(time.FixedZone("CST", 8*60*60))
	prefix := producer + chinaTime.Format("0102") + "-"
	prefixRunes := []rune(prefix)
	if len(prefixRunes) >= 64 {
		return "", fmt.Errorf("制作者账号过长，无法生成漫画合集名称")
	}
	titleRunes := []rune(title)
	remaining := 64 - len(prefixRunes)
	if len(titleRunes) > remaining {
		titleRunes = titleRunes[:remaining]
	}
	return prefix + string(titleRunes), nil
}
