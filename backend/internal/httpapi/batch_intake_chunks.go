package httpapi

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/mergeworker"
	"qiantie/backend/internal/novelfetchworkshop"
)

const (
	intakeChunkSize  = 512 << 10
	intakeBookLimit  = 24 << 20
	intakeBatchLimit = 48 << 20
	intakeChunkTTL   = "1800" // 30 minutes
)

type intakeChunks struct {
	do func(context.Context, ...string) (any, error)
}

func newIntakeChunks(options RouterOptions) intakeChunks {
	if redis, ok := options.NovelFetchQueue.(*novelfetchworkshop.RedisQueue); ok && redis.Do != nil {
		return intakeChunks{redis.Do}
	}
	if strings.TrimSpace(options.RedisURL) != "" {
		if redis, err := mergeworker.NewRedisStore(options.RedisURL, "qiantie:bf12:sources"); err == nil {
			return intakeChunks{redis.Do}
		}
	}
	return intakeChunks{}
}

type intakeChunkInput struct {
	Ref      string `json:"ref"`
	Platform string `json:"platform"`
	BookID   string `json:"bookId"`
	Bytes    int    `json:"bytes"`
	Index    int    `json:"index"`
	Data     string `json:"data"`
}

func intakeToken() (string, error) {
	var b [24]byte
	_, err := rand.Read(b[:])
	return hex.EncodeToString(b[:]), err
}

func intakeChunkKey(owner, platform, book, ref string) (string, error) {
	token, err := hex.DecodeString(ref)
	if err != nil || len(token) != 24 || book == "" || len(book) > 160 || len(platform) > 160 {
		return "", fmt.Errorf("正文暂存引用或书号无效")
	}
	binding, _ := json.Marshal([3]string{owner, platform, book})
	return fmt.Sprintf("qiantie:bf12:sources:%x:%s", sha256.Sum256(binding), ref), nil
}

func (s intakeChunks) get(ctx context.Context, key string) (string, error) {
	if s.do == nil {
		return "", fmt.Errorf("大正文暂存需要 Redis，请检查服务配置")
	}
	v, err := s.do(ctx, "GET", key)
	if mergeworker.IsRedisNil(err) || (err == nil && (v == nil || v == "")) {
		return "", fmt.Errorf("正文暂存已过期、分块缺失或不属于当前账号和书籍，请重新提交")
	}
	if err != nil {
		return "", err
	}
	value, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("正文暂存读取失败")
	}
	return value, nil
}

func (s intakeChunks) put(ctx context.Context, owner string, in intakeChunkInput) (string, error) {
	if s.do == nil {
		return "", fmt.Errorf("大正文暂存需要 Redis，请检查服务配置")
	}
	if in.Bytes < 1 || in.Bytes > intakeBookLimit || in.Index < 0 || in.Index > (in.Bytes-1)/intakeChunkSize {
		return "", fmt.Errorf("正文分块大小或序号无效")
	}
	part, err := base64.StdEncoding.DecodeString(in.Data)
	if err != nil || len(part) != min(intakeChunkSize, in.Bytes-in.Index*intakeChunkSize) {
		return "", fmt.Errorf("正文分块长度不符")
	}
	first := in.Ref == ""
	if first {
		if in.Index != 0 {
			return "", fmt.Errorf("请从第一块开始暂存")
		}
		if in.Ref, err = intakeToken(); err != nil {
			return "", err
		}
	}
	key, err := intakeChunkKey(owner, in.Platform, in.BookID, in.Ref)
	if err != nil {
		return "", err
	}
	if first {
		if _, err = s.do(ctx, "SET", key, strconv.Itoa(in.Bytes), "EX", intakeChunkTTL, "NX"); err != nil {
			return "", err
		}
	} else {
		n, err := s.get(ctx, key)
		if err != nil {
			return "", err
		}
		if n != strconv.Itoa(in.Bytes) {
			return "", fmt.Errorf("正文暂存大小发生变化")
		}
	}
	partKey := key + ":" + strconv.Itoa(in.Index)
	v, err := s.do(ctx, "SET", partKey, string(part), "EX", intakeChunkTTL, "NX")
	if err != nil && !mergeworker.IsRedisNil(err) {
		return "", err
	}
	if v == nil || mergeworker.IsRedisNil(err) {
		existing, err := s.get(ctx, partKey)
		if err != nil {
			return "", err
		}
		if existing != string(part) {
			return "", fmt.Errorf("正文分块与上一次不同，请重新提交")
		}
	}
	return in.Ref, nil
}

func (s intakeChunks) register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/batch-factory/v12/intakes/source-chunks", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, 401, map[string]string{"error": "unauthorized"})
			return
		}
		if !limitRequestBody(w, r, 1<<20) {
			return
		}
		var in intakeChunkInput
		if !decodeJSON(w, r, &in) {
			return
		}
		ref, err := s.put(r.Context(), owner, in)
		if err != nil {
			writeJSON(w, 422, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]string{"ref": ref})
	})
}

type stagedBookBody struct {
	SourceText  string `json:"sourceText"`
	OriginalRaw string `json:"originalRaw"`
	TxtText     string `json:"txtText"`
}
type intakeSourceLease struct {
	keys  []string
	token string
}
type intakeSourceRead struct {
	chunks intakeChunks
	leases []intakeSourceLease
	total  int
}

func (s intakeChunks) createNovelIntake(ctx context.Context, owner string, store batchfactoryv11.Store, input batchfactoryv11.NovelFetchIntakeInput) (batchfactoryv11.Intake, error) {
	read := intakeSourceRead{chunks: s}
	committed := false
	defer func() { read.finish(context.WithoutCancel(ctx), committed) }()
	if err := read.hydrate(ctx, owner, input.Books, false); err != nil {
		return batchfactoryv11.Intake{}, fmt.Errorf("%w：%s", batchfactoryv11.ErrInvalid, err)
	}
	intake, err := store.CreateIntake(ctx, owner, input)
	committed = err == nil
	if committed && len(read.leases) > 0 {
		intake.Payload = nil
	}
	return intake, err
}

func (read *intakeSourceRead) finish(ctx context.Context, success bool) {
	// The lease token prevents a timed-out request from clearing a new reader's keys.
	for _, lease := range read.leases {
		keys := lease.keys
		if !success {
			keys = keys[:1]
		}
		args := []string{"EVAL", "if redis.call('GET',KEYS[1])==ARGV[1] then return redis.call('DEL',unpack(KEYS)) end return 0", strconv.Itoa(len(keys))}
		args = append(args, keys...)
		args = append(args, lease.token)
		_, _ = read.chunks.do(ctx, args...)
	}
}

func (read *intakeSourceRead) book(ctx context.Context, owner, platform, bookID, ref string) (stagedBookBody, error) {
	var body stagedBookBody
	key, err := intakeChunkKey(owner, platform, bookID, ref)
	if err != nil {
		return body, err
	}
	n, err := read.chunks.get(ctx, key)
	if err != nil {
		return body, err
	}
	size, err := strconv.Atoi(n)
	if err != nil || size < 1 || size > intakeBookLimit || read.total+size > intakeBatchLimit {
		return body, errors.New("正文超过单书或整批暂存容量")
	}
	token, err := intakeToken()
	if err != nil {
		return body, err
	}
	lock := key + ":use"
	v, err := read.chunks.do(ctx, "SET", lock, token, "NX", "EX", "300")
	if mergeworker.IsRedisNil(err) || (err == nil && v == nil) {
		return body, errors.New("该书正在创建批次，请稍后重试")
	}
	if err != nil {
		return body, err
	}
	read.leases = append(read.leases, intakeSourceLease{keys: []string{lock, key}, token: token})
	lease := &read.leases[len(read.leases)-1]
	var buf bytes.Buffer
	buf.Grow(size)
	for off := 0; off < size; off += intakeChunkSize {
		partKey := key + ":" + strconv.Itoa(off/intakeChunkSize)
		part, err := read.chunks.get(ctx, partKey)
		if err != nil {
			return body, err
		}
		if len(part) != min(intakeChunkSize, size-off) {
			return body, errors.New("正文分块长度不符，请重新提交")
		}
		buf.WriteString(part)
		lease.keys = append(lease.keys, partKey)
	}
	if err := json.Unmarshal(buf.Bytes(), &body); err != nil {
		return body, errors.New("正文暂存内容格式无效")
	}
	read.total += size
	return body, nil
}

func (read *intakeSourceRead) hydrate(ctx context.Context, owner string, books []batchfactoryv11.CreateBookInput, manual bool) error {
	for i := range books {
		book := &books[i]
		if book.SourceRef == "" {
			continue
		}
		if len(books) > 50 {
			return errors.New("每个批次最多暂存 50 本")
		}
		if book.SourceText != "" || book.TxtText != "" || book.OriginalRaw != "" {
			return errors.New("暂存引用不能与正文同时提交")
		}
		id := book.BookID
		if id == "" {
			id = book.ID
		}
		source, err := read.book(ctx, owner, book.Platform, id, book.SourceRef)
		if err != nil {
			return err
		}
		book.SourceRef = ""
		if !manual {
			book.SourceText = source.SourceText
			book.OriginalRaw, book.TxtText = source.OriginalRaw, source.TxtText
			continue
		}
		book.SetManualSource(source.SourceText, source.OriginalRaw)
	}
	return nil
}
