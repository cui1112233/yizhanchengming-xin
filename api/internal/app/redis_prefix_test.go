package app

import "testing"

func TestRuntimeRedisPrefixKeepsDefaultAndScopesConfiguredNamespace(t *testing.T) {
	if got := runtimeRedisPrefix("", "task9"); got != "task9" {
		t.Fatalf("default prefix=%q want task9", got)
	}
	if got := runtimeRedisPrefix("ycm:staging:", "shuihuo-media"); got != "ycm:staging:shuihuo-media" {
		t.Fatalf("scoped prefix=%q", got)
	}
}
