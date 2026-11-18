package luoshui

import "embed"

// Native standalone parse messages recovered at the supplied ZIP's model boundary.
// These are independent of the selectable degree/image-generation templates.
// Provenance: zip-baseline/13-native-asset-parsing.md; asset-parse-v4.json.
//
//go:embed native_assets/*.txt
var nativeAssets embed.FS

func NativeAssetMessages(kind, content string) (system, user string, ok bool) {
	key := map[string]string{"character": "characters", "scene": "scenes", "prop": "props"}[kind]
	if key == "" {
		return "", "", false
	}
	body, err := nativeAssets.ReadFile("native_assets/" + key + "-system.txt")
	if err != nil {
		return "", "", false
	}
	prefix, err := nativeAssets.ReadFile("native_assets/" + key + "-user.txt")
	if err != nil {
		return "", "", false
	}
	return string(body), string(prefix) + content, true
}
