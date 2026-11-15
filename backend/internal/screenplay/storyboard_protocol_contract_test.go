package screenplay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"qiantie/backend/internal/prompt"
	"strings"
	"testing"
)

func TestRecoveredStoryboardProtocolsPersistAllFields(t *testing.T) {
	for _, id := range []int{137, 146, 147, 175, 192, 193, 202, 307} {
		t.Run(fmt.Sprint(id), func(t *testing.T) {
			s, store, g := newStoryboardService(t, PermissionBypass)
			sceneID := any(1)
			if id == 175 {
				sceneID = "1-1"
			}
			duration := 12
			if id == 146 || id == 175 {
				duration = 15
			}
			if id == 147 {
				duration = 6
			}
			group := map[string]any{"id": 1, "scene_id": sceneID, "duration_sec": duration, "characters": "甲", "scene": "家", "props": "杯子", "video_desc": "甲缓缓推开窗户。", "image_prompt": "甲站在窗边。", "image_desc": "九宫格静帧。", "source_elements": []string{"1-1-001"}, "voice_actor": "甲", "emotion": "平静", "line": "窗外是风景。", "duration_calc": "动作6秒", "timecode_global": "0:00-0:12", "content_density": "适中", "custom_future_field": map[string]any{"take": 2}}
			if id == 137 || id == 146 || id == 192 || id == 193 {
				delete(group, "duration_sec")
			}
			raw, _ := json.Marshal(map[string]any{"storyboard": []any{group}})
			response := string(raw)
			if id == 137 {
				response += "\n【分镜自检】\n场景：1-1 家\n剧本元素总数：1 -> 生成分镜数：1\n台词总数：1句 -> 遗漏：0\n连贯性检查：单组无相邻衔接\n映射完整性：角色1个/场景1个/物品1个"
			}
			g.responses = []string{response}
			req := generationTestRequest(fmt.Sprintf("protocol-%d", id))
			req.PromptID = fmt.Sprint(id)
			out, err := s.Generate(context.Background(), req)
			if err != nil {
				t.Fatal(err)
			}
			board, err := store.GetStoryboard(context.Background(), "alice", "p1", 1, out.Storyboards[0].Engine)
			if err != nil {
				t.Fatal(err)
			}
			shot := board.Shots[0]
			if shot.ImagePrompt != "甲站在窗边。" || shot.Line != "窗外是风景。" || shot.VoiceActor != "甲" || shot.Emotion != "平静" || !strings.Contains(string(shot.ProtocolFields), "custom_future_field") || !strings.Contains(string(shot.ProtocolFields), "parser_version") {
				t.Fatalf("lost protocol fields: %+v", shot)
			}
			if (id == 137) == g.requests[0].JSONResponse {
				t.Fatal("report mode lost")
			}
			state, _ := store.LoadStoryboardGenerations(context.Background(), "alice", "p1")
			attempt := state.Attempts[req.IdempotencyKey]
			if attempt.PromptID == "" || attempt.PromptDigest == "" || attempt.OutputProtocol == "" || attempt.ParserVersion == "" {
				t.Fatal("missing versioned output contract")
			}
		})
	}
}

func TestRecoveredStoryboardRejectsBadDurationTypesAndIncompleteReports(t *testing.T) {
	p, err := prompt.ResolveCreativePrompt("307", "storyboard")
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []any{0, -1, "garbage", 12.5, 40} {
		// Fractional duration is valid within this template's 4..30 second range.
		row := map[string]any{"id": 1, "scene_id": 1, "duration_sec": value, "characters": "甲", "scene": "家", "props": "无", "video_desc": "甲走路"}
		raw, _ := json.Marshal(map[string]any{"storyboard": []any{row}})
		shots, e := parseStoryboardJSON(string(raw))
		if e == nil {
			e = validateSelectedStoryboardProtocol(p, string(raw), shots)
		}
		if value == 12.5 {
			if e != nil {
				t.Fatal(e)
			}
		} else if e == nil {
			t.Fatalf("bad duration accepted %v", value)
		}
	}
	p, _ = prompt.ResolveCreativePrompt("137", "storyboard")
	row := map[string]any{"id": 1, "characters": "甲", "scene": "家", "image_prompt": "甲", "voice_actor": "无", "emotion": "无", "line": "无", "source_elements": []string{"1"}, "video_desc": "甲走路"}
	raw, _ := json.Marshal(map[string]any{"storyboard": []any{row}})
	shots, _ := parseStoryboardJSON(string(raw))
	if e := validateSelectedStoryboardProtocol(p, string(raw)+"\n【分镜自检】", shots); e == nil {
		t.Fatal("empty self-check accepted")
	}
}

func TestRecoveredUnsupportedPromptMakesNoProviderRequest(t *testing.T) {
	for _, id := range []string{"138"} {
		s, _, g := newStoryboardService(t, PermissionBypass)
		req := generationTestRequest("unsupported-" + id)
		req.PromptID = id
		_, err := s.Generate(context.Background(), req)
		if !errors.Is(err, ErrInvalidArgument) || g.callCount() != 0 {
			t.Fatal(id, err, g.callCount())
		}
	}
}
