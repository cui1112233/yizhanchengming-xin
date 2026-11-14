package screenplay

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

var chapterEpisodeHeading = regexp.MustCompile(`(?m)^#\s*第\s*\d+\s*集[^\n]*\n?`)
var chapterSceneHeading = regexp.MustCompile(`(?m)^##\s*\d+\s*-\s*(\d+)\s*[:：]`)

func chapterCollectionText(chapters []Chapter) (string, error) {
	parts := []string{}
	for i, c := range chapters {
		if strings.TrimSpace(c.ScriptText) == "" {
			continue
		}
		body := chapterEpisodeHeading.ReplaceAllString(c.ScriptText, "")
		body = chapterSceneHeading.ReplaceAllString(body, fmt.Sprintf("## %d-${1}：", i+1))
		parts = append(parts, fmt.Sprintf("# 第%d集：%s\n%s", i+1, c.Title, body))
	}
	if len(parts) == 0 {
		return "", ErrEmptyContent
	}
	return strings.Join(parts, "\n\n"), nil
}

// Validate before changing the authoring revision, so an invalid edit can be corrected and retried.
func ValidateChapterCollection(chapters []Chapter) error {
	content, err := chapterCollectionText(chapters)
	if err != nil {
		return err
	}
	parsed := ParseScreenplayMarkdown(content)
	if len(parsed.Shots) == 0 || len(parsed.Episodes) == 0 {
		return fmt.Errorf("%w: 剧本缺少可识别的集/场标题，请检查“# 第1集”和“## 1-1：日 内 场景”的格式后重试", ErrArtifactInvalid)
	}
	return nil
}
func PublishChapterCollection(ctx context.Context, s Store, o, p string, chapters []Chapter) (Document, error) {
	content, err := chapterCollectionText(chapters)
	if err != nil {
		return Document{}, err
	}
	return PublishChapterScripts(ctx, s, o, p, content)
}
