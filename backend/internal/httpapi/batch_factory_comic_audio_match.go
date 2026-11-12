package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/n8catalog"
)

var comicAudioCategoryByStyle = map[string]string{
	"现代悬疑": "现代悬疑",
	"男频都市": "男频都市",
	"古风通用": "古风通用",
	"玄幻":   "古风通用",
	"历史":   "古风通用",
	"古风虐文": "古风虐文",
	"古风甜文": "古风甜文",
	"现代虐文": "现代虐文",
	"年代虐文": "现代虐文",
	"家庭伤感": "现代虐文",
	"现代甜文": "现代通用",
	"现代通用": "现代通用",
	"年代甜文": "现代通用",
	"年代通用": "现代通用",
	"现代女主": "现代通用",
	"家庭奇葩": "现代通用",
}

func comicAudioMetadataText(metadata map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(fmt.Sprint(metadata[key])); value != "" && value != "<nil>" {
			return value
		}
	}
	return ""
}

func comicAudioCategoryName(book batchfactoryv11.Book) (string, error) {
	metadata := book.SourceMetadata
	status := strings.ToLower(comicAudioMetadataText(metadata, "classifyStatus", "classify_status"))
	if status != "classified" {
		return "", fmt.Errorf("%s 的男女频和风格尚未完成识别", strings.TrimSpace(book.Title))
	}
	style := comicAudioMetadataText(metadata, "style", "styleName", "style_name")
	gender := comicAudioMetadataText(metadata, "gender", "gender_name")
	if name := comicAudioCategoryByStyle[style]; name != "" {
		return name, nil
	}
	switch gender {
	case "男频", "男":
		return "男频都市", nil
	case "女频", "女":
		return "现代通用", nil
	default:
		return "", fmt.Errorf("%s 缺少可用于音频匹配的男女频和风格", strings.TrimSpace(book.Title))
	}
}

type comicAudioSelection struct {
	CategoryName string    `json:"categoryName"`
	CategoryID   int64     `json:"categoryId"`
	AudioName    string    `json:"audioName"`
	AudioID      int64     `json:"audioId"`
	Mode         string    `json:"mode"`
	Gender       string    `json:"gender,omitempty"`
	Style        string    `json:"style,omitempty"`
	SelectedAt   time.Time `json:"selectedAt"`
}

func comicAudioBook(batch batchfactoryv11.Batch, bookID string) (batchfactoryv11.Book, error) {
	for _, book := range batch.Books {
		if book.ID == bookID {
			return book, nil
		}
	}
	return batchfactoryv11.Book{}, batchfactoryv11.ErrNotFound
}

func savedComicAudioSelection(book batchfactoryv11.Book) (comicAudioSelection, bool) {
	raw := book.SettingsState.Patch["comicAudioSelection"]
	if len(raw) == 0 {
		return comicAudioSelection{}, false
	}
	var selection comicAudioSelection
	if json.Unmarshal(raw, &selection) != nil || selection.CategoryID <= 0 || selection.AudioID <= 0 {
		return comicAudioSelection{}, false
	}
	return selection, true
}

func (p *BatchFactoryN8ComicPublisher) randomAudioIndex(upper int) (int, error) {
	if upper <= 0 {
		return 0, fmt.Errorf("没有可用指定音频")
	}
	if p.RandomIndex != nil {
		return p.RandomIndex(upper)
	}
	value, err := rand.Int(rand.Reader, big.NewInt(int64(upper)))
	if err != nil {
		return 0, fmt.Errorf("随机选择音频失败：%w", err)
	}
	return int(value.Int64()), nil
}

// comicAudioFallbackCategories 是自动匹配音频时，按风格映射到的分类没有
// 可用音频后的兜底顺序。N8 后台常常只给通用分类上传音频，风格分类是空的；
// 如果直接报错，整本书会永久卡在发布预检里，必须人工换分类才能继续。
var comicAudioFallbackCategories = []string{"现代通用", "古风通用", "男频都市", "现代虐文", "现代甜文", "古风虐文", "古风甜文"}

func enabledComicAudioCategories(categories []n8catalog.AudioCategory) []n8catalog.AudioCategory {
	enabled := make([]n8catalog.AudioCategory, 0, len(categories))
	for _, category := range categories {
		if category.ID <= 0 || !strings.EqualFold(strings.TrimSpace(category.Status), "ENABLE") {
			continue
		}
		enabled = append(enabled, category)
	}
	return enabled
}

// pickComicAudioFallback 按兜底顺序找一个已启用且确实有音频的分类。
// used 是刚刚判定为“没有可用音频”的分类，必须排除；单个分类目录请求失败
// 只代表该分类暂时不可用，继续尝试下一个，不影响整本发布。
func (p *BatchFactoryN8ComicPublisher) pickComicAudioFallback(ctx context.Context, organizationID string, used n8catalog.AudioCategory, categories []n8catalog.AudioCategory) (n8catalog.AudioCategory, []n8catalog.Audio, error) {
	enabled := enabledComicAudioCategories(categories)
	ordered := make([]n8catalog.AudioCategory, 0, len(enabled))
	seen := map[int64]bool{}
	for _, name := range comicAudioFallbackCategories {
		for _, category := range enabled {
			if category.ID == used.ID || seen[category.ID] || !strings.EqualFold(strings.TrimSpace(category.Name), name) {
				continue
			}
			seen[category.ID] = true
			ordered = append(ordered, category)
		}
	}
	for _, category := range enabled {
		if category.ID == used.ID || seen[category.ID] {
			continue
		}
		seen[category.ID] = true
		ordered = append(ordered, category)
	}
	for _, category := range ordered {
		audios, err := p.audios(ctx, strconv.FormatInt(category.ID, 10), organizationID)
		if err != nil || len(audios) == 0 {
			continue
		}
		return category, audios, nil
	}
	return n8catalog.AudioCategory{}, nil, fmt.Errorf("当前组织所有启用分类都没有可用音频，请先在任意一个分类上传音频")
}

func enabledComicAudioCategory(categories []n8catalog.AudioCategory, id int64, name string) (n8catalog.AudioCategory, error) {
	matches := make([]n8catalog.AudioCategory, 0, 1)
	for _, category := range categories {
		if category.ID <= 0 || !strings.EqualFold(strings.TrimSpace(category.Status), "ENABLE") {
			continue
		}
		if id > 0 && category.ID != id {
			continue
		}
		// 分类 ID 是 N8 的稳定标识；名称只是保存时的展示快照，后台改名后可能过期。
		// 只有没有 ID 时才按名称查找，避免有效的已保存分类被旧名称误判为不可用。
		if id <= 0 && strings.TrimSpace(name) != "" && strings.TrimSpace(category.Name) != strings.TrimSpace(name) {
			continue
		}
		matches = append(matches, category)
	}
	if len(matches) != 1 {
		return n8catalog.AudioCategory{}, fmt.Errorf("音频分类“%s”不存在、未启用或名称不唯一", strings.TrimSpace(name))
	}
	return matches[0], nil
}

func enabledComicAudio(audios []n8catalog.Audio, id int64) (n8catalog.Audio, error) {
	for _, audio := range audios {
		if audio.ID == id && strings.EqualFold(strings.TrimSpace(audio.Status), "ENABLE") {
			return audio, nil
		}
	}
	return n8catalog.Audio{}, fmt.Errorf("手动指定的音频 ID %d 已不可用，请重新选择", id)
}

func (p *BatchFactoryN8ComicPublisher) resolveBookComicAudio(ctx context.Context, owner, batchID, bookID string, config n8ComicPublishSettings) (comicAudioSelection, error) {
	if p == nil || p.Store == nil {
		return comicAudioSelection{}, fmt.Errorf("漫画发布存储尚未配置")
	}
	batch, err := p.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return comicAudioSelection{}, err
	}
	book, err := comicAudioBook(batch, bookID)
	if err != nil {
		return comicAudioSelection{}, err
	}
	manualAudioID := config.AudioID
	if manualAudioID <= 0 && strings.EqualFold(strings.TrimSpace(config.AudioCategorySelectionMode), "manual") {
		manualAudioID = config.SpecifiedAudioID
	}
	if manualAudioID <= 0 {
		if selection, ok := savedComicAudioSelection(book); ok {
			audios, directoryErr := p.audios(ctx, strconv.FormatInt(selection.CategoryID, 10), config.OrganizationID)
			if directoryErr != nil {
				// 目录暂时失败不等于已保存音频失效；继续复用，避免重试时换音频。
				return selection, nil
			}
			if _, unavailable := enabledComicAudio(audios, selection.AudioID); unavailable == nil {
				return selection, nil
			}
			if selection.Mode == "manual" {
				return comicAudioSelection{}, fmt.Errorf("手动指定的音频 ID %d 已不可用，请重新选择", selection.AudioID)
			}
		}
	}

	categoryName := strings.TrimSpace(config.AudioCategoryName)
	categoryID := config.AudioCategoryID
	mode := "manual"
	if manualAudioID <= 0 {
		mode = "auto"
		categoryName, err = comicAudioCategoryName(book)
		if err != nil {
			return comicAudioSelection{}, err
		}
		categoryID = 0
	}
	categories := p.audioCategories(ctx)
	category, err := enabledComicAudioCategory(categories, categoryID, categoryName)
	if err != nil {
		return comicAudioSelection{}, err
	}
	audios, err := p.audios(ctx, strconv.FormatInt(category.ID, 10), config.OrganizationID)
	if err != nil {
		return comicAudioSelection{}, err
	}
	if manualAudioID <= 0 && len(audios) == 0 {
		// 按风格映射到的分类是空的（N8 后台没给该分类上传音频）。继续用通用
		// 分类兜底，避免整本书永久卡在发布预检；选中的分类会写进书设置，
		// 后续重试沿用同一个音频，不会每轮换 BGM。
		fallback, fallbackAudios, fallbackErr := p.pickComicAudioFallback(ctx, config.OrganizationID, category, categories)
		if fallbackErr != nil {
			return comicAudioSelection{}, fmt.Errorf("当前组织的“%s”分类没有可用音频，且没有其它启用分类可用音频", category.Name)
		}
		log.Printf("comic audio fallback book=%s category=%q -> %q", bookID, category.Name, fallback.Name)
		category, audios = fallback, fallbackAudios
	}
	var audio n8catalog.Audio
	if manualAudioID > 0 {
		audio, err = enabledComicAudio(audios, manualAudioID)
		if err != nil {
			return comicAudioSelection{}, err
		}
	} else {
		if len(audios) == 0 {
			return comicAudioSelection{}, fmt.Errorf("当前组织的“%s”分类没有可用音频", category.Name)
		}
		index, randomErr := p.randomAudioIndex(len(audios))
		if randomErr != nil {
			return comicAudioSelection{}, randomErr
		}
		if index < 0 || index >= len(audios) {
			return comicAudioSelection{}, fmt.Errorf("随机音频序号无效")
		}
		audio = audios[index]
	}
	now := time.Now().UTC()
	if p.Now != nil {
		now = p.Now().UTC()
	}
	selection := comicAudioSelection{
		CategoryName: category.Name, CategoryID: category.ID,
		AudioName: audio.Name, AudioID: audio.ID, Mode: mode,
		Gender:     comicAudioMetadataText(book.SourceMetadata, "gender", "gender_name"),
		Style:      comicAudioMetadataText(book.SourceMetadata, "style", "styleName", "style_name"),
		SelectedAt: now,
	}
	raw, err := json.Marshal(selection)
	if err != nil {
		return comicAudioSelection{}, err
	}
	_, err = p.Store.SaveSettings(ctx, owner, batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBook, BatchID: batchID, BookID: bookID}, batchfactoryv11.SettingsUpdate{
		Patch: batchfactoryv11.SettingsPatch{"comicAudioSelection": raw}, ExpectedRevision: book.Revision,
	})
	if err == nil {
		return selection, nil
	}
	if err != batchfactoryv11.ErrConflict {
		return comicAudioSelection{}, err
	}
	latest, loadErr := p.Store.GetBatch(ctx, owner, batchID)
	if loadErr != nil {
		return comicAudioSelection{}, loadErr
	}
	latestBook, loadErr := comicAudioBook(latest, bookID)
	if loadErr != nil {
		return comicAudioSelection{}, loadErr
	}
	if winner, ok := savedComicAudioSelection(latestBook); ok {
		return winner, nil
	}
	return comicAudioSelection{}, err
}
