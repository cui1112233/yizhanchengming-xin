package httpapi

import (
	"fmt"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
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
