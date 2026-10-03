package queue

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

var ErrQueueEmpty = errors.New("redis queue empty")

type Pusher interface {
	Push(context.Context, string, []byte) error
}

type Popper interface {
	Pop(context.Context, string) ([]byte, error)
}

type RedisQueue struct {
	key    string
	pusher Pusher
}

type RedisSource struct {
	key    string
	popper Popper
}

type networkRedis struct {
	address  string
	password string
	db       int
}

func NewRedisQueue(redisURL, key string) (*RedisQueue, error) {
	client, err := newNetworkRedis(redisURL)
	if err != nil {
		return nil, err
	}
	return NewRedisQueueWithPusher(key, client), nil
}

func NewRedisSource(redisURL, key string) (*RedisSource, error) {
	client, err := newNetworkRedis(redisURL)
	if err != nil {
		return nil, err
	}
	return NewRedisSourceWithPopper(key, client), nil
}

func NewRedisQueueWithPusher(key string, pusher Pusher) *RedisQueue {
	return &RedisQueue{key: normalizeKey(key), pusher: pusher}
}

func NewRedisSourceWithPopper(key string, popper Popper) *RedisSource {
	return &RedisSource{key: normalizeKey(key), popper: popper}
}

func normalizeKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return "qiantie:pipeline:ready"
	}
	return key
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

func (s *RedisSource) Next(ctx context.Context) (string, error) {
	if s == nil || s.popper == nil {
		return "", errors.New("redis source popper is required")
	}
	payload, err := s.popper.Pop(ctx, s.key)
	if err != nil {
		return "", err
	}
	var message struct {
		JobID string `json:"job_id"`
	}
	if err := json.Unmarshal(payload, &message); err != nil {
		return "", fmt.Errorf("decode redis job reference: %w", err)
	}
	message.JobID = strings.TrimSpace(message.JobID)
	if message.JobID == "" {
		return "", errors.New("redis job reference missing job_id")
	}
	return message.JobID, nil
}

func newNetworkRedis(raw string) (*networkRedis, error) {
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
	return &networkRedis{address: address, password: password, db: db}, nil
}

func (r *networkRedis) Push(ctx context.Context, key string, payload []byte) error {
	conn, reader, err := r.connect(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	return sendRedisCommand(conn, reader, "RPUSH", key, string(payload))
}

func (r *networkRedis) Pop(ctx context.Context, key string) ([]byte, error) {
	conn, reader, err := r.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	}
	if err := writeRedisCommand(conn, "BLPOP", key, "5"); err != nil {
		return nil, err
	}
	return readBLPopPayload(reader)
}

func (r *networkRedis) connect(ctx context.Context) (net.Conn, *bufio.Reader, error) {
	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", r.address)
	if err != nil {
		return nil, nil, fmt.Errorf("connect redis: %w", err)
	}
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	reader := bufio.NewReader(conn)
	if r.password != "" {
		if err := sendRedisCommand(conn, reader, "AUTH", r.password); err != nil {
			conn.Close()
			return nil, nil, err
		}
	}
	if r.db != 0 {
		if err := sendRedisCommand(conn, reader, "SELECT", strconv.Itoa(r.db)); err != nil {
			conn.Close()
			return nil, nil, err
		}
	}
	return conn, reader, nil
}

func sendRedisCommand(conn net.Conn, reader *bufio.Reader, args ...string) error {
	if err := writeRedisCommand(conn, args...); err != nil {
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

func writeRedisCommand(writer io.Writer, args ...string) error {
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
	_, err := io.WriteString(writer, builder.String())
	return err
}

func readBLPopPayload(reader *bufio.Reader) ([]byte, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)
	if line == "$-1" || line == "*-1" {
		return nil, ErrQueueEmpty
	}
	if line != "*2" {
		if strings.HasPrefix(line, "-") {
			return nil, errors.New(strings.TrimPrefix(line, "-"))
		}
		return nil, fmt.Errorf("unexpected redis BLPOP reply: %s", line)
	}
	if _, err := readBulkString(reader); err != nil {
		return nil, err
	}
	payload, err := readBulkString(reader)
	if err != nil {
		return nil, err
	}
	return []byte(payload), nil
}

func readBulkString(reader *bufio.Reader) (string, error) {
	header, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	header = strings.TrimSpace(header)
	if !strings.HasPrefix(header, "$") {
		return "", fmt.Errorf("unexpected redis bulk header: %s", header)
	}
	size, err := strconv.Atoi(strings.TrimPrefix(header, "$"))
	if err != nil || size < 0 {
		return "", errors.New("invalid redis bulk length")
	}
	buffer := make([]byte, size+2)
	if _, err := io.ReadFull(reader, buffer); err != nil {
		return "", err
	}
	return string(buffer[:size]), nil
}
