package screenplay

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

var chapterEpisodeHeading = regexp.MustCompile(`(?m)^#\s*第\s*\d+\s*集[^\n]*\n?`)
var chapterSceneHeading = regexp.MustCompile(`(?m)^##\s*\d+\s*-\s*(\d+)\s*[:：]`)

func PublishChapterCollection(ctx context.Context, s Store, o, p string, chapters []Chapter) (Document, error) {
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
		return Document{}, ErrEmptyContent
	}
	return PublishChapterScripts(ctx, s, o, p, strings.Join(parts, "\n\n"))
}
