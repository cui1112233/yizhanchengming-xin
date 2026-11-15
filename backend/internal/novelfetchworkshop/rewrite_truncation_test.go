package novelfetchworkshop

import (
	"context"
	"errors"
	"testing"
)

func TestRewriteTruncationRaisesBudgetOnceAndKeepsCompleteOutput(t *testing.T) {
	budgets := []int{}
	p := rewriteProviderFunc(func(_ context.Context, req RewriteRequest) (string, error) {
		budgets = append(budgets, req.MaxTokens)
		if req.MaxTokens < 12000 {
			return "", errors.New("文本模型输出被长度截断，请重试或增大输出上限")
		}
		return "她推开书店的门，朋友抬起头，两人终于把当年的误会说清。", nil
	})
	text, err := completeValidatedRewrite(context.Background(), p, RewriteRequest{MaxTokens: 1200}, "book", "plain_text", 1, 40)
	if err != nil || text == "" || len(budgets) != 2 || budgets[0] != 1200 || budgets[1] != 12000 {
		t.Fatalf("truncation did not recover with bounded budget: text=%q budgets=%v err=%v", text, budgets, err)
	}
}

func TestRewriteTruncationStopsAtBudgetAndRequiresManualRetry(t *testing.T) {
	for _, initial := range []int{1200, 24000} {
		calls := 0
		p := rewriteProviderFunc(func(_ context.Context, req RewriteRequest) (string, error) {
			calls++
			if req.MaxTokens > 24000 {
				t.Fatalf("unbounded output budget: %d", req.MaxTokens)
			}
			return "", errors.New("director provider output truncated (finish_reason=length)")
		})
		_, err := completeValidatedRewrite(context.Background(), p, RewriteRequest{MaxTokens: initial}, "book", "plain_text", 1, 40)
		want := 3
		if initial == 24000 {
			want = 1
		}
		if !errors.Is(err, ErrRewriteValidation) || calls != want {
			t.Fatalf("truncated output kept retrying: initial=%d calls=%d err=%v", initial, calls, err)
		}
	}
}

func TestRewriteNonTruncationProviderErrorIsNotRetriedInsideValidation(t *testing.T) {
	calls := 0
	want := errors.New("API key invalid")
	p := rewriteProviderFunc(func(context.Context, RewriteRequest) (string, error) { calls++; return "", want })
	_, err := completeValidatedRewrite(context.Background(), p, RewriteRequest{MaxTokens: 1200}, "book", "plain_text", 1, 40)
	if !errors.Is(err, want) || calls != 1 {
		t.Fatalf("unrelated provider error retried: calls=%d err=%v", calls, err)
	}
}
