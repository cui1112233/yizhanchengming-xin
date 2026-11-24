package batchfactorytext

import (
	"context"
	"strings"
	"testing"
)

func BenchmarkTextPageRendering(b *testing.B) {
	renderer := Renderer{}
	input := RenderInput{
		Text:       strings.Repeat("深夜十一点，她收到一封邮件。\n", 8),
		Layout:     DefaultRenderLayout(),
		Background: Background{Kind: "solid", Color: "#111318"},
	}
	if _, _, err := renderer.Render(context.Background(), input); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := renderer.Render(context.Background(), input); err != nil {
			b.Fatal(err)
		}
	}
}
