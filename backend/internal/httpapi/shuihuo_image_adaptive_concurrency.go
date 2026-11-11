package httpapi

import (
	"context"
	"net/url"
	"strings"
	"sync"
)

const shuihuoAdaptiveRecoverySuccesses = 8

type shuihuoAdaptiveHostState struct {
	limit     int
	failures  int
	successes int
	semaphore *shuihuoImageSemaphore
}

type shuihuoAdaptiveImageConcurrency struct {
	mu    sync.Mutex
	hosts map[string]*shuihuoAdaptiveHostState
}

func newShuihuoAdaptiveImageConcurrency() *shuihuoAdaptiveImageConcurrency {
	return &shuihuoAdaptiveImageConcurrency{hosts: map[string]*shuihuoAdaptiveHostState{}}
}

func (c *shuihuoAdaptiveImageConcurrency) state(host string, configured int) *shuihuoAdaptiveHostState {
	configured = shuihuoClampImageConcurrency(configured)
	host = strings.ToLower(strings.TrimSpace(host))
	state := c.hosts[host]
	if state == nil {
		// 图片网关在冷启动时直接压 4 路，实测会出现四条连接一起不返回。
		// 每个域名先从稳定的 2 路开始，连续成功后再逐级升到用户配置值。
		initial := configured
		if initial > shuihuoImageConcurrencyDefault {
			initial = shuihuoImageConcurrencyDefault
		}
		state = &shuihuoAdaptiveHostState{limit: initial, semaphore: newShuihuoImageSemaphore(initial)}
		c.hosts[host] = state
	}
	if state.limit > configured {
		state.limit = configured
		state.semaphore.setLimit(state.limit)
	}
	return state
}

func (c *shuihuoAdaptiveImageConcurrency) limit(host string, configured int) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state(host, configured).limit
}

func (c *shuihuoAdaptiveImageConcurrency) report(host string, configured int, outcome string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	state := c.state(host, configured)
	configured = shuihuoClampImageConcurrency(configured)
	switch outcome {
	case "reset", "unexpected_eof", "body_stalled", "tls_timeout":
		state.successes = 0
		state.failures++
		if state.failures >= 2 {
			state.limit = 2
			if configured < state.limit {
				state.limit = configured
			}
		}
	case "ok":
		state.failures = 0
		state.successes++
		if state.successes >= shuihuoAdaptiveRecoverySuccesses && state.limit < configured {
			state.limit++
			state.successes = 0
		}
	default:
		state.successes = 0
	}
	state.semaphore.setLimit(state.limit)
}

func (c *shuihuoAdaptiveImageConcurrency) acquire(ctx context.Context, endpoint string, configured int) (func(), string, error) {
	parsed, _ := url.Parse(strings.TrimSpace(endpoint))
	host := strings.ToLower(strings.TrimSpace(parsed.Host))
	if host == "" {
		return func() {}, "", nil
	}
	c.mu.Lock()
	semaphore := c.state(host, configured).semaphore
	c.mu.Unlock()
	release, err := semaphore.acquire(ctx)
	return release, host, err
}

var shuihuoAdaptiveImageLimiter = newShuihuoAdaptiveImageConcurrency()
