package mergeworker

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type RedisDoFunc func(context.Context, ...string) (any, error)

type RedisStore struct {
	URL    string
	Prefix string
	Do     RedisDoFunc
}

var _ Store = (*RedisStore)(nil)
var _ Queue = (*RedisStore)(nil)

func NewRedisStore(rawURL, prefix string) (*RedisStore, error) {
	client := &redisClient{rawURL: strings.TrimSpace(rawURL)}
	if err := client.validate(); err != nil {
		return nil, err
	}
	prefix = strings.Trim(strings.TrimSpace(prefix), ":")
	if prefix == "" {
		prefix = "qiantie:bf11:merge"
	}
	return &RedisStore{URL: rawURL, Prefix: prefix, Do: client.do}, nil
}

func (s *RedisStore) jobKey(id string) string { return s.prefix() + ":job:" + id }
func (s *RedisStore) queueKey() string        { return s.prefix() + ":queue" }
func (s *RedisStore) prefix() string {
	value := strings.Trim(strings.TrimSpace(s.Prefix), ":")
	if value == "" {
		return "qiantie:bf11:merge"
	}
	return value
}

func (s *RedisStore) command(ctx context.Context, args ...string) (any, error) {
	if s == nil || s.Do == nil {
		return nil, fmt.Errorf("redis unavailable")
	}
	return s.Do(ctx, args...)
}

func (s *RedisStore) Create(ctx context.Context, job Job) (Job, error) {
	if strings.TrimSpace(job.ID) == "" {
		return Job{}, fmt.Errorf("job id is required")
	}
	now := time.Now().UTC()
	if job.CreatedAt.IsZero() {
		job.CreatedAt = now
	}
	job.UpdatedAt = now
	body, err := json.Marshal(job)
	if err != nil {
		return Job{}, err
	}
	if _, err := s.command(ctx, "SET", s.jobKey(job.ID), string(body), "NX"); err != nil {
		return Job{}, err
	}
	return job, nil
}

func (s *RedisStore) Get(ctx context.Context, id string) (Job, error) {
	value, err := s.command(ctx, "GET", s.jobKey(strings.TrimSpace(id)))
	if err != nil {
		if errors.Is(err, errRedisNil) {
			return Job{}, ErrNotFound
		}
		return Job{}, err
	}
	text, ok := value.(string)
	if !ok || text == "" {
		return Job{}, ErrNotFound
	}
	var job Job
	if err := json.Unmarshal([]byte(text), &job); err != nil {
		return Job{}, fmt.Errorf("decode redis merge job: %w", err)
	}
	return job, nil
}

func (s *RedisStore) Update(ctx context.Context, job Job) (Job, error) {
	if strings.TrimSpace(job.ID) == "" {
		return Job{}, fmt.Errorf("job id is required")
	}
	if _, err := s.Get(ctx, job.ID); err != nil {
		return Job{}, err
	}
	job.UpdatedAt = time.Now().UTC()
	body, err := json.Marshal(job)
	if err != nil {
		return Job{}, err
	}
	if _, err := s.command(ctx, "SET", s.jobKey(job.ID), string(body)); err != nil {
		return Job{}, err
	}
	return job, nil
}

func (s *RedisStore) Enqueue(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("job id is required")
	}
	_, err := s.command(ctx, "RPUSH", s.queueKey(), id)
	return err
}

func (s *RedisStore) Dequeue(ctx context.Context, timeout time.Duration) (string, error) {
	seconds := int(timeout.Round(time.Second) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	value, err := s.command(ctx, "BLPOP", s.queueKey(), strconv.Itoa(seconds))
	if errors.Is(err, errRedisNil) {
		return "", ErrQueueEmpty
	}
	if err != nil {
		return "", err
	}
	values, ok := value.([]any)
	if !ok || len(values) != 2 {
		return "", fmt.Errorf("unexpected redis queue response")
	}
	id, ok := values[1].(string)
	if !ok || strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("unexpected redis queue item")
	}
	return id, nil
}

type redisClient struct{ rawURL string }

type redisEndpoint struct {
	scheme   string
	address  string
	username string
	password string
	db       int
}

func (c *redisClient) validate() error {
	_, err := parseRedisURL(c.rawURL)
	return err
}

func parseRedisURL(raw string) (redisEndpoint, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (parsed.Scheme != "redis" && parsed.Scheme != "rediss") || parsed.Hostname() == "" {
		return redisEndpoint{}, fmt.Errorf("invalid redis URL")
	}
	port := parsed.Port()
	if port == "" {
		port = "6379"
	}
	endpoint := redisEndpoint{scheme: parsed.Scheme, address: net.JoinHostPort(parsed.Hostname(), port)}
	if parsed.User != nil {
		endpoint.username = parsed.User.Username()
		endpoint.password, _ = parsed.User.Password()
	}
	path := strings.Trim(parsed.Path, "/")
	if path != "" {
		db, err := strconv.Atoi(path)
		if err != nil || db < 0 {
			return redisEndpoint{}, fmt.Errorf("invalid redis database")
		}
		endpoint.db = db
	}
	return endpoint, nil
}

func (c *redisClient) do(ctx context.Context, args ...string) (any, error) {
	endpoint, err := parseRedisURL(c.rawURL)
	if err != nil {
		return nil, err
	}
	dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
	var conn net.Conn
	if endpoint.scheme == "rediss" {
		conn, err = tls.DialWithDialer(dialer, "tcp", endpoint.address, &tls.Config{MinVersion: tls.VersionTLS12, ServerName: strings.Split(endpoint.address, ":")[0]})
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", endpoint.address)
	}
	if err != nil {
		return nil, fmt.Errorf("redis connection failed")
	}
	defer conn.Close()
	deadline := time.Now().Add(10 * time.Second)
	if value, ok := ctx.Deadline(); ok && value.Before(deadline) {
		deadline = value
	}
	_ = conn.SetDeadline(deadline)
	reader := bufio.NewReader(conn)
	if endpoint.password != "" {
		auth := []string{"AUTH"}
		if endpoint.username != "" {
			auth = append(auth, endpoint.username)
		}
		auth = append(auth, endpoint.password)
		if err := writeRESP(conn, auth...); err != nil {
			return nil, fmt.Errorf("redis authentication failed")
		}
		if _, err := readRESP(reader); err != nil {
			return nil, fmt.Errorf("redis authentication failed")
		}
	}
	if endpoint.db != 0 {
		if err := writeRESP(conn, "SELECT", strconv.Itoa(endpoint.db)); err != nil {
			return nil, fmt.Errorf("redis database selection failed")
		}
		if _, err := readRESP(reader); err != nil {
			return nil, fmt.Errorf("redis database selection failed")
		}
	}
	if err := writeRESP(conn, args...); err != nil {
		return nil, fmt.Errorf("redis request failed")
	}
	// A blocking claim may legitimately be silent longer than the normal
	// command timeout. Extend only this read, after auth and request writes.
	if len(args) == 4 && strings.EqualFold(args[0], "BRPOPLPUSH") {
		if wait, parseErr := time.ParseDuration(args[3] + "s"); parseErr == nil && wait >= 10*time.Second {
			readDeadline := time.Now().Add(wait).Add(2 * time.Second)
			if limit, ok := ctx.Deadline(); ok && limit.Before(readDeadline) {
				readDeadline = limit
			}
			_ = conn.SetReadDeadline(readDeadline)
		}
	}
	value, err := readRESP(reader)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func encodeRESPCommand(args ...string) []byte {
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
	return []byte(builder.String())
}

func writeRESP(writer io.Writer, args ...string) error {
	_, err := writer.Write(encodeRESPCommand(args...))
	return err
}

var errRedisNil = errors.New("redis nil")

func readRESP(reader *bufio.Reader) (any, error) {
	prefix, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("redis response failed")
	}
	line := func() (string, error) {
		value, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSuffix(strings.TrimSuffix(value, "\n"), "\r"), nil
	}
	switch prefix {
	case '+':
		return line()
	case '-':
		message, _ := line()
		if len(message) > 160 {
			message = message[:160]
		}
		return nil, fmt.Errorf("redis error: %s", message)
	case ':':
		value, err := line()
		if err != nil {
			return nil, err
		}
		return strconv.ParseInt(value, 10, 64)
	case '$':
		value, err := line()
		if err != nil {
			return nil, err
		}
		length, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		if length == -1 {
			return nil, errRedisNil
		}
		if length < 0 || length > 16<<20 {
			return nil, fmt.Errorf("invalid redis bulk length")
		}
		payload := make([]byte, length+2)
		if _, err := io.ReadFull(reader, payload); err != nil {
			return nil, err
		}
		return string(payload[:length]), nil
	case '*':
		value, err := line()
		if err != nil {
			return nil, err
		}
		count, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		if count == -1 {
			return nil, errRedisNil
		}
		if count < 0 || count > 1024 {
			return nil, fmt.Errorf("invalid redis array length")
		}
		items := make([]any, 0, count)
		for i := 0; i < count; i++ {
			item, err := readRESP(reader)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		return items, nil
	default:
		return nil, fmt.Errorf("unsupported redis response")
	}
}
