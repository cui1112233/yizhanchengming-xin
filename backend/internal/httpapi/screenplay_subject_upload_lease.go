package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type subjectUploadRedisDo func(context.Context, ...string) (any, error)

func subjectUploadLease(ctx context.Context, do subjectUploadRedisDo, owner, key string) (func(), error) {
	if do == nil {
		return nil, fmt.Errorf("真人素材上传需要Redis租约服务")
	}
	b := make([]byte, 24)
	if _, e := rand.Read(b); e != nil {
		return nil, e
	}
	token := hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(owner + "\x00" + key))
	lock := fmt.Sprintf("qiantie:subject-upload:%x", sum)
	v, e := do(ctx, "SET", lock, token, "NX", "EX", "300")
	if e != nil {
		return nil, e
	}
	if fmt.Sprint(v) != "OK" {
		return nil, fmt.Errorf("该上传正在执行，请稍后核对")
	}
	return func() {
		c, cancel := context.WithTimeout(context.Background(), 5e9)
		defer cancel()
		_, _ = do(c, "EVAL", `if redis.call('GET',KEYS[1])==ARGV[1] then return redis.call('DEL',KEYS[1]) end return 0`, "1", lock, token)
	}, nil
}
