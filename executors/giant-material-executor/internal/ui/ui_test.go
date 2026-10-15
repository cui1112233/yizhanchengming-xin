package ui

import (
	"context"
	"testing"

	"qiantie/giant-material-executor/internal/agent"
)

func TestValidateConfigRequiresPairCallback(t *testing.T) {
	if err := validateConfig(Config{}); err == nil {
		t.Fatal("expected missing pair callback to fail")
	}
	if err := validateConfig(Config{Pair: func(context.Context, string) error { return nil }, Snapshot: func() agent.Snapshot { return agent.Snapshot{} }}); err != nil {
		t.Fatalf("valid config error=%v", err)
	}
}

func TestBindingStatusLabel(t *testing.T) {
	cases := []struct {
		state agent.BindingState
		want  string
	}{
		{agent.BindingUnpaired, "未绑定"},
		{agent.BindingConnecting, "连接中"},
		{agent.BindingOnline, "在线"},
		{agent.BindingOffline, "已绑定但离线"},
		{agent.BindingNeedsPairing, "需要重新绑定"},
	}
	for _, tc := range cases {
		if got := bindingStatusLabel(tc.state); got != tc.want {
			t.Fatalf("state=%s label=%q, want %q", tc.state, got, tc.want)
		}
	}
}
