package httpapi

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Validate before persisting: a saved card must represent executable input,
// not an error deferred indefinitely to the background scanner.
func normalizeNovelFetchScheduleInput(item, settings map[string]any) error {
	inputText := strings.TrimSpace(firstText(item["input_text"], item["inputText"]))
	if inputText == "" {
		return errors.New("请输入要定时处理的书籍清单")
	}
	runAt, err := parseNovelFetchScheduleTime(strings.TrimSpace(firstText(item["run_at"], item["runAt"])))
	if err != nil {
		return errors.New("请选择有效的定时处理时间")
	}
	platformID := strings.TrimSpace(firstText(item["platform_id"], item["platformId"], "2"))
	platformFound := false
	for _, raw := range novelFetchPlatformCatalog(settings) {
		platform, ok := raw.(map[string]any)
		if ok && strings.TrimSpace(fmt.Sprint(platform["id"])) == platformID {
			platformFound = true
			break
		}
	}
	if !platformFound {
		return errors.New("所选平台不存在，请重新选择平台")
	}
	documents, _, _, _ := parseScheduleDocuments(item)
	if len(documents) == 0 {
		return errors.New("书籍清单中没有有效书籍，请填写书籍 ID")
	}
	item["input_text"] = inputText
	item["platform_id"] = platformID
	item["run_at"] = runAt.UTC().Format(time.RFC3339Nano)
	return nil
}
