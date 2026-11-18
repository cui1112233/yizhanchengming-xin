package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"qiantie/backend/internal/novelfetchworkshop"
	"strings"
	"sync"
	"time"
)

const (
	// 锁 TTL 与参考实现的 stale 判定对齐：进程崩溃后最多 6 分钟即可重新提交。
	// 正常提交由心跳每 60 秒续期，不会在中途过期。
	workshopSubmitLockTTLSeconds     = 360
	workshopSubmitLockRefreshSeconds = 60
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

// Serialize target-site writes across application owners and cookie rotation.
// Older sessions may lack a target username, so conservatively lock the site
// rather than let a known and an unknown session write the same account at once.
func workshopExternalSubmitScope(session map[string]any) string {
	base := strings.TrimRight(strings.TrimSpace(firstText(session["baseUrl"], "https://two.121w.com/tttadmin")), "/")
	if parsed, err := url.Parse(base); err == nil && parsed.Host != "" {
		base = strings.ToLower(parsed.Host)
	}
	return "external-site\n" + base
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
	value, err := redis.Do(ctx, "EVAL", `if redis.call('SET',KEYS[1],ARGV[1],'NX','EX',ARGV[2]) then return 1 end return 0`, "1", key, token, fmt.Sprint(workshopSubmitLockTTLSeconds))
	if err != nil {
		return nil, fmt.Errorf("提交防重复检查失败：%w", err)
	}
	if value == nil || fmt.Sprint(value) != "1" {
		return nil, fmt.Errorf("此账号已有上传或核对任务，请等待完成后重试")
	}
	stop := make(chan struct{})
	var stopOnce sync.Once
	go func() {
		ticker := time.NewTicker(workshopSubmitLockRefreshSeconds * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				refreshCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				_, _ = redis.Do(refreshCtx, "EVAL", `if redis.call('GET',KEYS[1]) == ARGV[1] then return redis.call('EXPIRE',KEYS[1],ARGV[2]) end return 0`, "1", key, token, fmt.Sprint(workshopSubmitLockTTLSeconds))
				cancel()
			}
		}
	}()
	return func() {
		stopOnce.Do(func() { close(stop) })
		_, _ = redis.Do(context.Background(), "EVAL", `if redis.call('GET',KEYS[1]) == ARGV[1] then return redis.call('DEL',KEYS[1]) end return 0`, "1", key, token)
	}, nil
}
