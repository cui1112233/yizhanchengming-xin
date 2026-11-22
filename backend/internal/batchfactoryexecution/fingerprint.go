package batchfactoryexecution

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// CanonicalJSON returns deterministic JSON for execution snapshots. Object
// keys are sorted while array order and JSON number spelling are preserved.
func CanonicalJSON(value any) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode canonical json: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var decoded any
	if err := decoder.Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode canonical json: %w", err)
	}
	var out bytes.Buffer
	if err := writeCanonicalJSON(&out, decoded); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func writeCanonicalJSON(out *bytes.Buffer, value any) error {
	switch typed := value.(type) {
	case nil:
		out.WriteString("null")
	case bool:
		if typed { out.WriteString("true") } else { out.WriteString("false") }
	case string:
		out.WriteString(strconv.Quote(typed))
	case json.Number:
		number := typed.String()
		parsed, err := strconv.ParseFloat(number, 64)
		if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
			return fmt.Errorf("invalid json number %q", number)
		}
		out.WriteString(number)
	case []any:
		out.WriteByte('[')
		for index, item := range typed {
			if index > 0 { out.WriteByte(',') }
			if err := writeCanonicalJSON(out, item); err != nil { return err }
		}
		out.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed { keys = append(keys, key) }
		sort.Strings(keys)
		out.WriteByte('{')
		for index, key := range keys {
			if index > 0 { out.WriteByte(',') }
			out.WriteString(strconv.Quote(key))
			out.WriteByte(':')
			if err := writeCanonicalJSON(out, typed[key]); err != nil { return err }
		}
		out.WriteByte('}')
	default:
		return fmt.Errorf("unsupported canonical json value %T", value)
	}
	return nil
}

func ConfigDigest(config any) (string, error) {
	canonical, err := CanonicalJSON(config)
	if err != nil { return "", err }
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), nil
}

func StageInputFingerprint(stageKey string, configDigest string, input any) (string, error) {
	stageKey = strings.TrimSpace(stageKey)
	configDigest = strings.TrimSpace(configDigest)
	if stageKey == "" || configDigest == "" {
		return "", fmt.Errorf("stage key and config digest are required")
	}
	return ConfigDigest(map[string]any{"stageKey": stageKey, "configDigest": configDigest, "input": input})
}
