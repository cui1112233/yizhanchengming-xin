package httpapi

import "testing"

func TestAdaptiveImageConcurrencyStartsAtStableTwo(t *testing.T) {
	controller := newShuihuoAdaptiveImageConcurrency()
	host := "images.example.com"
	if got := controller.limit(host, 4); got != 2 {
		t.Fatalf("initial limit=%d, want stable limit 2", got)
	}
	controller.report(host, 4, "reset")
	controller.report(host, 4, "body_stalled")
	if got := controller.limit(host, 4); got != 2 {
		t.Fatalf("degraded limit=%d, want 2", got)
	}
}

func TestAdaptiveImageConcurrencyRecoversAfterConsecutiveSuccesses(t *testing.T) {
	controller := newShuihuoAdaptiveImageConcurrency()
	host := "images.example.com"
	for index := 0; index < shuihuoAdaptiveRecoverySuccesses; index++ {
		controller.report(host, 4, "ok")
	}
	if got := controller.limit(host, 4); got != 3 {
		t.Fatalf("first recovery limit=%d, want 3", got)
	}
	for index := 0; index < shuihuoAdaptiveRecoverySuccesses; index++ {
		controller.report(host, 4, "ok")
	}
	if got := controller.limit(host, 4); got != 4 {
		t.Fatalf("full recovery limit=%d, want 4", got)
	}
}
