package agentchat

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

// newID returns a UUID v4 string, matching the original workspaces.json id
// format (no "ws_" prefix). A non-empty prefix produces "<prefix>-<hex>".
func newID(prefix string) string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return prefix
	}
	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80
	value := hex.EncodeToString(raw[:])
	uuid := value[0:8] + "-" + value[8:12] + "-" + value[12:16] + "-" + value[16:20] + "-" + value[20:32]
	if prefix == "" {
		return uuid
	}
	return prefix + "-" + uuid
}

func decodeJSON(raw []byte, target any) error {
	return json.Unmarshal(raw, target)
}

func writeJSONFile(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o600)
}
