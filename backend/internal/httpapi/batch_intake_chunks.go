package httpapi

import (
	"bytes"
	"context"
	"crypto/hmac"
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
	intakeChunkTTL   = "1800"    // 30 minutes
	intakeFieldLimit = 1<<24 - 1 // MEDIUMTEXT bytes, including utf8mb4.
	intakeOwnerRefs  = 50
)

var errIntakeFieldBytes = fmt.Errorf("每个原文或生产稿字段最多 %d 个 UTF-8 字节", intakeFieldLimit)

type intakeChunks struct {
	do     func(context.Context, ...string) (any, error)
	secret string
}

func newIntakeChunks(options RouterOptions) intakeChunks {
	if redis, ok := options.NovelFetchQueue.(*novelfetchworkshop.RedisQueue); ok && redis.Do != nil {
		return intakeChunks{do: redis.Do, secret: options.BridgeSecret}
	}
	if strings.TrimSpace(options.RedisURL) != "" {
		if redis, err := mergeworker.NewRedisStore(options.RedisURL, "qiantie:bf12:sources"); err == nil {
			return intakeChunks{do: redis.Do, secret: options.BridgeSecret}
		}
	}
	return intakeChunks{}
}

type intakeChunkInput struct {
	Ref           string `json:"ref"`
	Platform      string `json:"platform"`
	BookID        string `json:"bookId"`
	Bytes         int    `json:"bytes"`
	Index         int    `json:"index"`
	Data          string `json:"data"`
	UploadID      string `json:"uploadId,omitempty"`
	MaxFieldBytes int    `json:"maxFieldBytes,omitempty"`
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
	return fmt.Sprintf("%s:book:%x:%s", intakeOwnerKey(owner), sha256.Sum256(binding), ref), nil
}

func intakeOwnerKey(owner string) string {
	return fmt.Sprintf("qiantie:bf12:sources:{%x}", sha256.Sum256([]byte(owner)))
}

func (s intakeChunks) get(ctx context.Context, key, field string) (string, error) {
	if s.do == nil {
		return "", fmt.Errorf("大正文暂存需要 Redis，请检查服务配置")
	}
	v, err := s.do(ctx, "HGET", key, field)
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

// A single Redis command reserves the entire declared size before accepting any
// bytes. Every chunk lives in the same expiring hash. The small owner hash holds
// at most 50 reservations, including interrupted uploads on other processes.
const putIntakeChunkLua = `
local size=redis.call('HGET',KEYS[1],'bytes')
if not size then
  if ARGV[1]~='1' then return -1 end
  local used,refs=0,0
  local pending=redis.call('HGETALL',KEYS[2])
  for i=1,#pending,2 do
    if redis.call('EXISTS',pending[i])==1 then
      refs=refs+1
      used=used+tonumber(pending[i+1])
    else redis.call('HDEL',KEYS[2],pending[i]) end
  end
  if refs>=tonumber(ARGV[7]) or used+tonumber(ARGV[2])>tonumber(ARGV[6]) then return -2 end
  redis.call('HSET',KEYS[1],'bytes',ARGV[2],'generation',ARGV[9])
  redis.call('EXPIRE',KEYS[1],ARGV[5])
  redis.call('HSET',KEYS[2],KEYS[1],ARGV[2])
  redis.call('EXPIRE',KEYS[2],tonumber(ARGV[5])+60)
elseif size~=ARGV[2] then return -3 end
local old=redis.call('HGET',KEYS[1],ARGV[3])
if old and old~=ARGV[4] then return -3 end
if not old then redis.call('HSET',KEYS[1],ARGV[3],ARGV[4]) end
if redis.call('HEXISTS',KEYS[1],'ready')==1 then return 1 end
if redis.call('HLEN',KEYS[1])==tonumber(ARGV[8])+2 then return redis.call('HGET',KEYS[1],'generation') end
return 1`

func (s intakeChunks) put(ctx context.Context, owner string, in intakeChunkInput) (string, error) {
	if s.do == nil {
		return "", fmt.Errorf("大正文暂存需要 Redis，请检查服务配置")
	}
	if in.MaxFieldBytes > intakeFieldLimit {
		return "", errIntakeFieldBytes
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
		if in.UploadID != "" {
			id, err := hex.DecodeString(in.UploadID)
			if err != nil || len(id) != 24 {
				return "", errors.New("正文上传标识无效")
			}
			mac := hmac.New(sha256.New, []byte(s.secret))
			binding, _ := json.Marshal([3]string{owner, in.Platform, in.BookID})
			mac.Write(binding)
			mac.Write(id)
			in.Ref = hex.EncodeToString(mac.Sum(nil)[:24])
		} else if in.Ref, err = intakeToken(); err != nil {
			return "", err
		}
	}
	key, err := intakeChunkKey(owner, in.Platform, in.BookID, in.Ref)
	if err != nil {
		return "", err
	}
	start := "0"
	generation := ""
	if first {
		start = "1"
		generation, err = intakeToken()
		if err != nil {
			return "", err
		}
	}
	v, err := s.do(ctx, "EVAL", putIntakeChunkLua, "2", key, intakeOwnerKey(owner), start, strconv.Itoa(in.Bytes), "p"+strconv.Itoa(in.Index), string(part), intakeChunkTTL, strconv.Itoa(intakeBatchLimit), strconv.Itoa(intakeOwnerRefs), strconv.Itoa((in.Bytes+intakeChunkSize-1)/intakeChunkSize), generation)
	if err != nil {
		return "", err
	}
	switch v {
	case int64(-1):
		return "", errors.New("正文暂存已过期或不属于当前书籍，请重新提交")
	case int64(-2):
		return "", errors.New("当前账号待处理正文已达 48 MiB 或 50 个引用，请完成已有批次或待暂存到期后重试")
	case int64(-3):
		return "", errors.New("重复上传的分块与已存正文不一致")
	case int64(1):
		return in.Ref, nil
	default:
		generation, ok := v.(string)
		if !ok || generation == "" {
			return "", errors.New("正文暂存状态无效")
		}
		// Verify the actual decoded field sizes before acknowledging a complete
		// upload; clients can omit or understate their size hint.
		if _, err = s.body(ctx, key, in.Bytes); err != nil {
			if errors.Is(err, errIntakeFieldBytes) || errors.Is(err, errIntakeBodyFormat) {
				_, _ = s.do(ctx, "EVAL", "if redis.call('HGET',KEYS[1],'generation')==ARGV[1] then redis.call('DEL',KEYS[1]); return redis.call('HDEL',KEYS[2],KEYS[1]) end return 0", "2", key, intakeOwnerKey(owner), generation)
			}
			return "", err
		}
		result, err := s.do(ctx, "EVAL", "if redis.call('HGET',KEYS[1],'generation')==ARGV[1] then return redis.call('HSET',KEYS[1],'ready',ARGV[1]) end return -1", "1", key, generation)
		if err != nil {
			return "", err
		}
		if result == int64(-1) {
			return "", errors.New("正文暂存已过期")
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
			writeIntakeSourceError(w, err)
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

var errIntakeBodyFormat = errors.New("正文暂存内容格式无效")

func validateIntakeFields(source, raw, txt string) error {
	if len(source) > intakeFieldLimit || len(raw) > intakeFieldLimit || len(txt) > intakeFieldLimit {
		return errIntakeFieldBytes
	}
	return nil
}
func writeIntakeSourceError(w http.ResponseWriter, err error) {
	code := http.StatusUnprocessableEntity
	if errors.Is(err, errIntakeFieldBytes) {
		code = http.StatusRequestEntityTooLarge
	}
	writeJSON(w, code, map[string]string{"error": err.Error()})
}

type intakeSourceLease struct {
	key        string
	token      string
	ownerKey   string
	generation string
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
		return batchfactoryv11.Intake{}, fmt.Errorf("%w：%w", batchfactoryv11.ErrInvalid, err)
	}
	batchfactoryv11.AcceptExternalOriginals(input.Books, "novel_fetch_original_raw")
	intake, err := store.CreateIntake(ctx, owner, input)
	committed = err == nil
	if committed && len(read.leases) > 0 {
		intake.Payload = nil
	}
	return intake, err
}

func (read *intakeSourceRead) finish(ctx context.Context, success bool) {
	// The lease token prevents a timed-out request from clearing a new reader's keys.
	const release = `
if redis.call('GET',KEYS[1])~=ARGV[1] then return 0 end
redis.call('DEL',KEYS[1])
if ARGV[2]=='true' and redis.call('HGET',KEYS[2],'generation')==ARGV[3] then
  redis.call('DEL',KEYS[2])
  redis.call('HDEL',KEYS[3],KEYS[2])
end
return 1`
	for _, lease := range read.leases {
		_, _ = read.chunks.do(ctx, "EVAL", release, "3", lease.key+":use", lease.key, lease.ownerKey, lease.token, strconv.FormatBool(success), lease.generation)
	}
}

func (read *intakeSourceRead) book(ctx context.Context, owner, platform, bookID, ref string) (stagedBookBody, error) {
	var body stagedBookBody
	key, err := intakeChunkKey(owner, platform, bookID, ref)
	if err != nil {
		return body, err
	}
	n, err := read.chunks.get(ctx, key, "bytes")
	if err != nil {
		return body, err
	}
	size, err := strconv.Atoi(n)
	if err != nil || size < 1 || size > intakeBookLimit || read.total+size > intakeBatchLimit {
		return body, errors.New("正文超过单书或整批暂存容量")
	}
	generation, err := read.chunks.get(ctx, key, "ready")
	if err != nil {
		return body, err
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
	read.leases = append(read.leases, intakeSourceLease{key: key, ownerKey: intakeOwnerKey(owner), token: token, generation: generation})
	read.total += size
	body, err = read.chunks.body(ctx, key, size)
	if err != nil {
		return body, err
	}
	current, err := read.chunks.get(ctx, key, "ready")
	if err != nil {
		return body, err
	}
	if current != generation {
		return body, errors.New("正文暂存已更新，请重新提交")
	}
	return body, nil
}

func (s intakeChunks) body(ctx context.Context, key string, size int) (stagedBookBody, error) {
	var body stagedBookBody
	var buf bytes.Buffer
	buf.Grow(size)
	for off := 0; off < size; off += intakeChunkSize {
		part, err := s.get(ctx, key, "p"+strconv.Itoa(off/intakeChunkSize))
		if err != nil {
			return body, err
		}
		if len(part) != min(intakeChunkSize, size-off) {
			return body, errors.New("正文分块长度不符，请重新提交")
		}
		buf.WriteString(part)
	}
	if err := json.Unmarshal(buf.Bytes(), &body); err != nil {
		return body, errIntakeBodyFormat
	}
	return body, validateIntakeFields(body.SourceText, body.OriginalRaw, body.TxtText)
}

func (read *intakeSourceRead) hydrate(ctx context.Context, owner string, books []batchfactoryv11.CreateBookInput, manual bool) error {
	for i := range books {
		book := &books[i]
		if book.SourceRef == "" {
			if err := validateIntakeFields(book.SourceText, book.OriginalRaw, book.TxtText); err != nil {
				return err
			}
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
