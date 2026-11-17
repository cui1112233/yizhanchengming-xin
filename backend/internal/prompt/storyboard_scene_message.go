package prompt

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
)

// StoryboardSceneUserMessage follows the supplied ZIP's captured final message.
// Scene routing metadata stays in CM's checkpoint, outside the model input.
func StoryboardSceneUserMessage(original, content string) (string, error) {
	if strings.TrimSpace(original) == "" || strings.TrimSpace(content) == "" {
		return "", errors.New("章节原文或当前场景内容为空，请补充后再生成")
	}
	var scene map[string]json.RawMessage
	if json.Unmarshal([]byte(content), &scene) != nil || scene == nil {
		text, _ := json.Marshal(content)
		scene = map[string]json.RawMessage{"content": text}
	}
	for key := range scene {
		if strings.HasPrefix(key, "_") {
			delete(scene, key)
		}
	}
	if len(scene) == 0 {
		return "", errors.New("当前场景只有内部标记，没有可生成的内容")
	}
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(scene); err != nil {
		return "", err
	}
	return "【章节原文】\n" + original + "\n\n【当前剧本场景JSON（仅1个场景）】\n" + strings.TrimSuffix(b.String(), "\n") + "\n\n【处理要求】\n请根据以上原文和当前剧本场景内容，为这个场景生成完整的分镜JSON。", nil
}
