package ui

import (
	"context"
	"errors"
	"strings"

	"qiantie/giant-material-executor/internal/agent"
)

type Config struct {
	// PublicURL is the control API address used by the executor. It is shown
	// in the native Windows UI so an executor on another computer does not
	// accidentally call its own 127.0.0.1.
	PublicURL string
	SetPublicURL func(string) error
	Pair         func(context.Context, string) error
	Snapshot     func() agent.Snapshot
	Shutdown     func()
	// Version 是编译时注入的执行器版本号，显示在托盘提示里。
	Version string
}

func validateConfig(config Config) error {
	if config.Pair == nil {
		return errors.New("executor UI pair callback is required")
	}
	if config.Snapshot == nil {
		return errors.New("executor UI snapshot callback is required")
	}
	return nil
}

func bindingStatusLabel(state agent.BindingState) string {
	switch strings.TrimSpace(string(state)) {
	case string(agent.BindingUnpaired):
		return "未绑定"
	case string(agent.BindingConnecting):
		return "连接中"
	case string(agent.BindingOnline):
		return "在线"
	case string(agent.BindingOffline):
		return "已绑定但离线"
	case string(agent.BindingNeedsPairing):
		return "需要重新绑定"
	default:
		return "未连接"
	}
}
