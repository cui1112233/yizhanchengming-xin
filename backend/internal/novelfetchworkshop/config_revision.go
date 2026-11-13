package novelfetchworkshop

import (
	"errors"
	"fmt"
	"strconv"
)

const ConfigRevisionKey = "configRevision"

var ErrConfigConflict = errors.New("配置已被其他页面或任务更新，请刷新配置后重新修改；本次修改没有覆盖新配置")

func configRevision(settings map[string]any) string {
	value := fmt.Sprint(settings[ConfigRevisionKey])
	if value == "<nil>" || value == "" {
		return "0"
	}
	return value
}

func checkedConfigRevision(current, incoming map[string]any) (string, error) {
	previous := configRevision(current)
	if expected, supplied := incoming[ConfigRevisionKey]; supplied && fmt.Sprint(expected) != previous {
		return "", ErrConfigConflict
	}
	n, err := strconv.ParseUint(previous, 10, 64)
	if err != nil {
		return "", fmt.Errorf("读取配置版本失败")
	}
	return strconv.FormatUint(n+1, 10), nil
}
