package queue

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

type Pusher interface {
	Push(context.Context, string, []byte) error
}

type RedisQueue struct {
	key    string
	pusher Pusher
}

type networkPusher struct {
	address  string
	password string
	db       int
}

func NewRedisQueue(redisURL, key string) (*RedisQueue, error) {
	pusher, err := newNetworkPusher(redisURL)
	if err != nil {
		return nil, err
	}
	return NewRedisQueueWithPusher(key, pusher), nil
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

func newNetworkPusher(raw string) (*networkPusher, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed == nil || parsed.Scheme != "redis" || parsed.Host == "" {
		return nil, errors.New("redis URL must use redis://host:port")
	}
	address := parsed.Host
	if !strings.Contains(address, ":") {
		address += ":6379"
	}
	password := ""
	if parsed.User != nil {
		password, _ = parsed.User.Password()
	}
	db := 0
	if path := strings.Trim(strings.TrimSpace(parsed.Path), "/"); path != "" {
		value, err := strconv.Atoi(path)
		if err != nil || value < 0 {
			return nil, errors.New("invalid redis database number")
		}
		db = value
	}
	return &networkPusher{address: address, password: password, db: db}, nil
}

func (p *networkPusher) Push(ctx context.Context, key string, payload []byte) error {
	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", p.address)
	if err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	reader := bufio.NewReader(conn)
	if p.password != "" {
		if err := sendRedisCommand(conn, reader, "AUTH", p.password); err != nil {
			return err
		}
	}
	if p.db != 0 {
		if err := sendRedisCommand(conn, reader, "SELECT", strconv.Itoa(p.db)); err != nil {
			return err
		}
	}
	return sendRedisCommand(conn, reader, "RPUSH", key, string(payload))
}

func sendRedisCommand(conn net.Conn, reader *bufio.Reader, args ...string) error {
	var builder strings.Builder
	builder.WriteString("*")
	builder.WriteString(strconv.Itoa(len(args)))
	builder.WriteString("\r\n")
	for _, arg := range args {
		builder.WriteString("$")
		builder.WriteString(strconv.Itoa(len(arg)))
		builder.WriteString("\r\n")
		builder.WriteString(arg)
		builder.WriteString("\r\n")
	}
	if _, err := conn.Write([]byte(builder.String())); err != nil {
		return err
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if strings.HasPrefix(line, "-") {
		return errors.New(strings.TrimSpace(strings.TrimPrefix(line, "-")))
	}
	return nil
}
