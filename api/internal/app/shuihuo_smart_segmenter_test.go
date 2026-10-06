package app

import (
	"context"
	"errors"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/shuihuo"
)

type smartProviderStub struct {
	output string
	err    error
}

func (s smartProviderStub) Complete(context.Context, generation.TextRequest) (string, error) {
	return s.output, s.err
}

func TestShuihuoSmartSegmenterConsumesSharedProvider(t *testing.T) {
	segmenter := newShuihuoSmartSegmenter(smartProviderStub{output: `[{"text":"镜头一"}]`})
	rows, err := segmenter.Segment(context.Background(), "原文")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Text != "镜头一" || rows[0].Speaker != "旁白" {
		t.Fatalf("rows=%#v", rows)
	}
}

func TestShuihuoSmartSegmenterMapsSharedUnavailable(t *testing.T) {
	segmenter := newShuihuoSmartSegmenter(smartProviderStub{err: generation.ErrUnavailable})
	_, err := segmenter.Segment(context.Background(), "原文")
	if !errors.Is(err, shuihuo.ErrSmartUnavailable) {
		t.Fatalf("err=%v", err)
	}
}
