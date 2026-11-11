package httpapi

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
)

const aicutFontUploadMaxBytes = int64(20 << 20)

type aicutFontNameCandidate struct {
	value string
	score int
}

func parseAICutFontFamily(data []byte) (string, error) {
	if len(data) < 12 {
		return "", fmt.Errorf("字体文件头无效")
	}
	numTables := int(binary.BigEndian.Uint16(data[4:6]))
	if numTables <= 0 || 12+numTables*16 > len(data) {
		return "", fmt.Errorf("字体目录无效")
	}
	var nameTable []byte
	for index := 0; index < numTables; index++ {
		record := 12 + index*16
		if string(data[record:record+4]) != "name" {
			continue
		}
		offset := int(binary.BigEndian.Uint32(data[record+8 : record+12]))
		length := int(binary.BigEndian.Uint32(data[record+12 : record+16]))
		if offset < 0 || length < 6 || offset > len(data) || length > len(data)-offset {
			return "", fmt.Errorf("字体名称表无效")
		}
		nameTable = data[offset : offset+length]
		break
	}
	if len(nameTable) < 6 {
		return "", fmt.Errorf("字体缺少名称表")
	}
	count := int(binary.BigEndian.Uint16(nameTable[2:4]))
	stringsOffset := int(binary.BigEndian.Uint16(nameTable[4:6]))
	if 6+count*12 > len(nameTable) || stringsOffset > len(nameTable) {
		return "", fmt.Errorf("字体名称记录无效")
	}
	best := aicutFontNameCandidate{}
	for index := 0; index < count; index++ {
		record := nameTable[6+index*12 : 18+index*12]
		platform := binary.BigEndian.Uint16(record[0:2])
		language := binary.BigEndian.Uint16(record[4:6])
		nameID := binary.BigEndian.Uint16(record[6:8])
		if nameID != 1 && nameID != 16 {
			continue
		}
		length := int(binary.BigEndian.Uint16(record[8:10]))
		offset := stringsOffset + int(binary.BigEndian.Uint16(record[10:12]))
		if length <= 0 || offset < stringsOffset || offset > len(nameTable) || length > len(nameTable)-offset {
			continue
		}
		value := decodeAICutFontName(platform, nameTable[offset:offset+length])
		value = strings.TrimSpace(strings.ReplaceAll(value, "\x00", ""))
		if value == "" || len([]rune(value)) > 128 {
			continue
		}
		score := 0
		if nameID == 16 {
			score += 100
		}
		if platform == 3 {
			score += 20
		}
		if language == 0x0409 {
			score += 10
		}
		if best.value == "" || score > best.score {
			best = aicutFontNameCandidate{value: value, score: score}
		}
	}
	if best.value == "" {
		return "", fmt.Errorf("字体缺少可用家族名称")
	}
	return best.value, nil
}

func decodeAICutFontName(platform uint16, data []byte) string {
	if platform != 0 && platform != 3 {
		return string(data)
	}
	if len(data)%2 != 0 {
		return ""
	}
	encoded := make([]uint16, len(data)/2)
	for index := range encoded {
		encoded[index] = binary.BigEndian.Uint16(data[index*2 : index*2+2])
	}
	return string(utf16.Decode(encoded))
}

func prepareAICutExportFonts(ctx context.Context, owner, projectID, dir string, timeline aicut.Timeline, store aicut.Store, objects agentworkspace.AssetObjectStore) (aicut.Timeline, string, error) {
	timeline.Tracks = append([]aicut.Track(nil), timeline.Tracks...)
	fontsDir := ""
	prepared := map[string]string{}
	for index := range timeline.Tracks {
		fontAssetID := strings.TrimSpace(timeline.Tracks[index].Settings.FontAssetID)
		if fontAssetID == "" {
			continue
		}
		if family := prepared[fontAssetID]; family != "" {
			timeline.Tracks[index].Settings.FontFamily = family
			continue
		}
		asset, err := store.Asset(ctx, owner, projectID, fontAssetID)
		if err != nil {
			return timeline, "", fmt.Errorf("文字轨“%s”的用户字体不存在", timeline.Tracks[index].Name)
		}
		if asset.Kind != "font" {
			return timeline, "", fmt.Errorf("文字轨“%s”引用的素材不是字体", timeline.Tracks[index].Name)
		}
		body, err := objects.Open(ctx, aicutObjectAsset(asset))
		if err != nil {
			return timeline, "", fmt.Errorf("用户字体“%s”读取失败", asset.Name)
		}
		data, readErr := io.ReadAll(io.LimitReader(body, aicutFontUploadMaxBytes+1))
		_ = body.Close()
		if readErr != nil || int64(len(data)) > aicutFontUploadMaxBytes {
			return timeline, "", fmt.Errorf("用户字体“%s”读取失败", asset.Name)
		}
		family, err := parseAICutFontFamily(data)
		if err != nil {
			return timeline, "", fmt.Errorf("用户字体“%s”无效", asset.Name)
		}
		if fontsDir == "" {
			fontsDir = filepath.Join(dir, "fonts")
			if err := os.MkdirAll(fontsDir, 0o700); err != nil {
				return timeline, "", fmt.Errorf("用户字体目录准备失败")
			}
		}
		ext := strings.ToLower(filepath.Ext(asset.Name))
		if ext != ".ttf" && ext != ".otf" {
			if asset.ContentType == "font/otf" {
				ext = ".otf"
			} else {
				ext = ".ttf"
			}
		}
		fontPath := filepath.Join(fontsDir, fmt.Sprintf("font-%03d%s", len(prepared)+1, ext))
		if err := os.WriteFile(fontPath, data, 0o600); err != nil {
			return timeline, "", fmt.Errorf("用户字体“%s”准备失败", asset.Name)
		}
		prepared[fontAssetID] = family
		timeline.Tracks[index].Settings.FontFamily = family
	}
	return timeline, fontsDir, nil
}
