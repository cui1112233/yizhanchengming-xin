package aicut

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ExportNotifier interface {
	Notify(context.Context, string) error
}

type ExportWakeQueue interface {
	ExportNotifier
	Wait(context.Context, time.Duration) error
}

type RedisExportQueue struct {
	prefix string
	do     RedisDoFunc
}

func NewRedisExportQueue(prefix string, do RedisDoFunc) *RedisExportQueue {
	prefix = strings.Trim(strings.TrimSpace(prefix), ":")
	if prefix == "" {
		prefix = "qiantie:aicut:export"
	}
	return &RedisExportQueue{prefix: prefix, do: do}
}

func (q *RedisExportQueue) wakeKey() string { return q.prefix + ":wake" }

func (q *RedisExportQueue) Notify(ctx context.Context, jobID string) error {
	if q == nil || q.do == nil {
		return errors.New("Redis 导出队列未配置")
	}
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return errors.New("导出任务编号不能为空")
	}
	_, err := q.do(ctx, "RPUSH", q.wakeKey(), jobID)
	return err
}

func (q *RedisExportQueue) Wait(ctx context.Context, timeout time.Duration) error {
	if q == nil || q.do == nil {
		return errors.New("Redis 导出队列未配置")
	}
	seconds := int(timeout.Round(time.Second) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	value, err := q.do(ctx, "BLPOP", q.wakeKey(), strconv.Itoa(seconds))
	if err != nil {
		return err
	}
	items, ok := value.([]any)
	if !ok || len(items) != 2 {
		return fmt.Errorf("Redis 导出队列响应无效")
	}
	return nil
}
