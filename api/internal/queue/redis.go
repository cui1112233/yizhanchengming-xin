package queue

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
	"github.com/redis/go-redis/v9"
)

type Pusher interface {
	Push(context.Context, string, []byte) error
}

type RedisQueue struct {
	key    string
	pusher Pusher
}

type redisPusher struct {
	client *redis.Client
}

func NewRedisQueue(client *redis.Client, key string) *RedisQueue {
	return NewRedisQueueWithPusher(key, &redisPusher{client: client})
}

func NewRedisQueueWithPusher(key string, pusher Pusher) *RedisQueue {
	key = strings.TrimSpace(key)
	if key == "" {
		key = "qiantie:pipeline:ready"
	}
	return &RedisQueue{key: key, pusher: pusher}
}

func (q *RedisQueue) Enqueue(ctx context.Context, job pipeline.Job) error {
	if q == nil || q.pusher == nil {
		return errors.New("redis queue pusher is required")
	}
	if strings.TrimSpace(job.ID) == "" {
		return errors.New("pipeline job id is required")
	}
	payload, err := json.Marshal(map[string]string{"job_id": job.ID})
	if err != nil {
		return err
	}
	return q.pusher.Push(ctx, q.key, payload)
}

func (p *redisPusher) Push(ctx context.Context, key string, payload []byte) error {
	if p == nil || p.client == nil {
		return errors.New("redis client is required")
	}
	return p.client.RPush(ctx, key, payload).Err()
}
