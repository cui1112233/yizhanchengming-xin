package screenplay

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// The original files used both numeric and textual episode values.
func (v *PositioningItem) UnmarshalJSON(raw []byte) error {
	type plain PositioningItem
	var value struct {
		*plain
		Episode json.RawMessage `json:"episode"`
	}
	value.plain = (*plain)(v)
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	if len(value.Episode) == 0 || string(value.Episode) == "null" {
		v.Episode = ""
		return nil
	}
	if err := json.Unmarshal(value.Episode, &v.Episode); err == nil {
		return nil
	}
	var number json.Number
	if err := json.Unmarshal(value.Episode, &number); err != nil {
		return err
	}
	if _, err := strconv.Atoi(number.String()); err != nil {
		return err
	}
	v.Episode = number.String()
	return nil
}

// PositioningEpisodeText preserves a complete episode rather than truncating a scene.
func PositioningEpisodeText(content, episode string) string {
	number, _ := strconv.Atoi(strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(episode, "第"), "集")))
	episodes := splitEpisodes(content)
	for _, v := range episodes {
		if v.No == number {
			return v.Content
		}
	}
	return content
}

type PositioningItem struct {
	ID               string   `json:"id"`
	Index            int      `json:"index"`
	Episode          string   `json:"episode"`
	SceneCode        string   `json:"scene_code"`
	SceneCodes       []string `json:"scene_codes"`
	SceneName        string   `json:"scene_name"`
	Characters       []string `json:"characters"`
	PositionDetail   string   `json:"position_detail"`
	Remark           string   `json:"remark"`
	Reason           string   `json:"reason"`
	ImagePath        string   `json:"image_path"`
	ImageMediaItemID string   `json:"image_media_item_id,omitempty"`
	Status           string   `json:"status,omitempty"`
	Error            string   `json:"error,omitempty"`
	RequestKey       string   `json:"request_key,omitempty"`
}
type PositioningState struct {
	Revision           int64             `json:"revision"`
	Items              []PositioningItem `json:"items"`
	PrefixText         string            `json:"prefix_text"`
	ImageModel         string            `json:"image_model"`
	ShowCharacterNames bool              `json:"show_character_names"`
}

func ValidatePositioning(v *PositioningState) error {
	if len(v.Items) > 2000 {
		return fmt.Errorf("站位最多保存2000条")
	}
	seen := map[string]bool{}
	for i := range v.Items {
		item := &v.Items[i]
		item.ID = strings.TrimSpace(item.ID)
		if item.ID == "" || len(item.ID) > 96 || strings.ContainsAny(item.ID, "/\\?#") || seen[item.ID] {
			return fmt.Errorf("站位ID不能为空、重复或包含路径字符")
		}
		seen[item.ID] = true
		if len([]rune(item.PositionDetail)) > 20000 || len(item.Characters) > 100 {
			return fmt.Errorf("站位描述或角色数量过多")
		}
		item.Index = i + 1
		item.ImagePath = ""
	}
	if v.Items == nil {
		v.Items = []PositioningItem{}
	}
	if len([]rune(v.PrefixText)) > 30000 {
		return fmt.Errorf("站位前缀过长")
	}
	return nil
}
