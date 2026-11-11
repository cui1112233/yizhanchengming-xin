package aicut

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type RedisDoFunc func(context.Context, ...string) (any, error)

type ASRWakeQueue interface {
	ASRNotifier
	Wait(context.Context, time.Duration) error
}

type RedisASRQueue struct {
	prefix string
	do     RedisDoFunc
}

func NewRedisASRQueue(prefix string, do RedisDoFunc) *RedisASRQueue {
	prefix = strings.Trim(strings.TrimSpace(prefix), ":")
	if prefix == "" {
		prefix = "qiantie:aicut:asr"
	}
	return &RedisASRQueue{prefix: prefix, do: do}
}

func (q *RedisASRQueue) wakeKey() string { return q.prefix + ":wake" }

func (q *RedisASRQueue) Notify(ctx context.Context, jobID string) error {
	if q == nil || q.do == nil {
		return errors.New("Redis 队列未配置")
	}
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return errors.New("ASR 任务编号不能为空")
	}
	_, err := q.do(ctx, "RPUSH", q.wakeKey(), jobID)
	return err
}

func (q *RedisASRQueue) Wait(ctx context.Context, timeout time.Duration) error {
	if q == nil || q.do == nil {
		return errors.New("Redis 队列未配置")
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
		return fmt.Errorf("Redis ASR 队列响应无效")
	}
	return nil
}

func (s *ASRService) Run(ctx context.Context, queue ASRWakeQueue) {
	if s == nil || queue == nil {
		return
	}
	for ctx.Err() == nil {
		_ = queue.Wait(ctx, 5*time.Second)
		for ctx.Err() == nil {
			err := s.ProcessNext(ctx)
			if errors.Is(err, ErrASRQueueEmpty) {
				break
			}
			if err != nil {
				continue
			}
		}
	}
}
