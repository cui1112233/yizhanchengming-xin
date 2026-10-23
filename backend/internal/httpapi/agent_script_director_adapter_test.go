package httpapi

import (
	"strings"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
)

func TestAdaptDirectorResultToScriptShotsKeepsDirectorTimelineForCanvasGeneration(t *testing.T) {
	result := batchfactoryv11.DirectorResult{Storyboard: []batchfactoryv11.DirectorVideo{{
		DurationSec:  6,
		Characters:   []string{"林夏", "周野"},
		Props:        []string{"手电筒"},
		Scene:        "旧仓库",
		VideoDesc:    "林夏进入仓库，周野在暗处回应。",
		VisualPrompt: "雨夜旧仓库，湿地反光，冷色电影光。",
		Shots: []batchfactoryv11.DirectorShot{
			{StartSec: 0, EndSec: 3, ShotType: "中景", Camera: "缓慢推进", Description: "林夏推开铁门，举起手电筒。", Audio: "林夏：有人吗？"},
			{StartSec: 3, EndSec: 6, ShotType: "近景", Camera: "切至周野", Description: "周野从货架后现身。", Audio: "周野：别出声。"},
		},
	}}}

	shots, err := adaptDirectorResultToScriptShots("script-1", result)
	if err != nil {
		t.Fatal(err)
	}
	if len(shots) != 1 {
		t.Fatalf("shot count=%d", len(shots))
	}
	shot := shots[0]
	if shot.ID != "script-1-shot-1" || shot.Sequence != 1 || shot.Scene != "旧仓库" || shot.Characters != "林夏、周野" {
		t.Fatalf("shot=%+v", shot)
	}
	if shot.ImagePrompt != "雨夜旧仓库，湿地反光，冷色电影光。" {
		t.Fatalf("image prompt=%q", shot.ImagePrompt)
	}
	if !strings.Contains(shot.VideoPrompt, "00:00-00:03") || !strings.Contains(shot.VideoPrompt, "中景｜缓慢推进") || !strings.Contains(shot.VideoPrompt, "周野从货架后现身") {
		t.Fatalf("video prompt=%q", shot.VideoPrompt)
	}
	if !strings.Contains(shot.TTSText, "林夏：有人吗？") || !strings.Contains(shot.TTSText, "周野：别出声。") {
		t.Fatalf("tts text=%q", shot.TTSText)
	}
}
