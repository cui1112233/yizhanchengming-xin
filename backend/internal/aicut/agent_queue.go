package aicut

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type AgentWakeQueue interface {
	Notify(context.Context, string) error
	Wait(context.Context, time.Duration) (string, error)
}

type RedisAgentQueue struct {
	prefix string
	do     RedisDoFunc
}

func NewRedisAgentQueue(prefix string, do RedisDoFunc) *RedisAgentQueue {
	prefix = strings.Trim(strings.TrimSpace(prefix), ":")
	if prefix == "" {
		prefix = "qiantie:aicut:agent"
	}
	return &RedisAgentQueue{prefix: prefix, do: do}
}

func (q *RedisAgentQueue) wakeKey() string { return q.prefix + ":wake" }

func (q *RedisAgentQueue) Notify(ctx context.Context, runID string) error {
	if q == nil || q.do == nil {
		return errors.New("Redis Agent 队列未配置")
	}
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return errors.New("Agent 任务编号不能为空")
	}
	_, err := q.do(ctx, "RPUSH", q.wakeKey(), runID)
	return err
}

func (q *RedisAgentQueue) Wait(ctx context.Context, timeout time.Duration) (string, error) {
	if q == nil || q.do == nil {
		return "", errors.New("Redis Agent 队列未配置")
	}
	seconds := int(timeout.Round(time.Second) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	value, err := q.do(ctx, "BLPOP", q.wakeKey(), strconv.Itoa(seconds))
	if err != nil {
		return "", err
	}
	items, ok := value.([]any)
	if !ok || len(items) != 2 {
		return "", fmt.Errorf("Redis Agent 队列响应无效")
	}
	runID := strings.TrimSpace(fmt.Sprint(items[1]))
	if runID == "" {
		return "", fmt.Errorf("Redis Agent 队列任务编号为空")
	}
	return runID, nil
}

func (s *AgentService) RunQueue(ctx context.Context, queue AgentWakeQueue) {
	if s == nil || queue == nil {
		return
	}
	for ctx.Err() == nil {
		runID, err := queue.Wait(ctx, 5*time.Second)
		if err != nil || runID == "" {
			continue
		}
		_, _ = s.Run(ctx, runID)
	}
}
