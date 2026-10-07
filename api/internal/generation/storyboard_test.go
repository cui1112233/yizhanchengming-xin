package generation

import "testing"

func TestSplitStoryboardPreservesCardBoundariesAndFallback(t *testing.T) {
	cards := splitStoryboard("### 分镜一（总时长：10s）\n甲进入房间\n---\n### 分镜二（总时长：10s）\n乙回头")
	if len(cards) != 2 || cards[0].Position != 1 || cards[1].Title != "分镜二（总时长：10s）" || cards[0].Content != "甲进入房间" {
		t.Fatalf("cards = %#v", cards)
	}
	fallback := splitStoryboard("没有标题的导演内容")
	if len(fallback) != 1 || fallback[0].Title != "分镜一" || fallback[0].Content != "没有标题的导演内容" {
		t.Fatalf("fallback = %#v", fallback)
	}
}
