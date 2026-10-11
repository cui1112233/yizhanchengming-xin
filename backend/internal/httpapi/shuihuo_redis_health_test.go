package httpapi

import (
	"context"
	"testing"
)

func TestShuihuoRedisHealthRequiresURL(t *testing.T) {
	ready, reason := shuihuoRedisHealth(context.Background(), "")
	if ready {
		t.Fatal("empty Redis URL must not be reported ready")
	}
	if reason != "未配置 Redis 地址" {
		t.Fatalf("reason = %q", reason)
	}
}

func TestShuihuoRedisHealthRejectsInvalidURL(t *testing.T) {
	ready, reason := shuihuoRedisHealth(context.Background(), "not-a-redis-url")
	if ready {
		t.Fatal("invalid Redis URL must not be reported ready")
	}
	if reason != "Redis 地址无效" {
		t.Fatalf("reason = %q", reason)
	}
}
