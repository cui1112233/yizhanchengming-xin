package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"qiantie/backend/internal/novelfetchworkshop"
)

type submitQueueKey struct{}

func withWorkshopSubmitQueue(next http.Handler, queue novelfetchworkshop.ProcessQueue) http.Handler {
	if queue == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), submitQueueKey{}, queue)))
	})
}

func acquireWorkshopSubmit(ctx context.Context, queue novelfetchworkshop.ProcessQueue, owner, book string) (func(), error) {
	redis, ok := queue.(*novelfetchworkshop.RedisQueue)
	if !ok {
		return func() {}, nil
	} // Memory fixtures do not represent a distributed deployment.
	if redis.Do == nil {
		return nil, fmt.Errorf("提交防重复服务不可用，请检查 Redis")
	}
	digest := sha256.Sum256([]byte(owner + "\naccount-upload"))
	key := redis.Prefix + ":submit-lock:" + hex.EncodeToString(digest[:])
	tokenBytes := make([]byte, 24)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(tokenBytes)
	value, err := redis.Do(ctx, "EVAL", `if redis.call('SET',KEYS[1],ARGV[1],'NX','EX',ARGV[2]) then return 1 end return 0`, "1", key, token, "1800")
	if err != nil {
		return nil, fmt.Errorf("提交防重复检查失败：%w", err)
	}
	if value == nil || fmt.Sprint(value) != "1" {
		return nil, fmt.Errorf("此账号已有上传或核对任务，请等待完成后重试")
	}
	return func() {
		_, _ = redis.Do(context.Background(), "EVAL", `if redis.call('GET',KEYS[1]) == ARGV[1] then return redis.call('DEL',KEYS[1]) end return 0`, "1", key, token)
	}, nil
}
