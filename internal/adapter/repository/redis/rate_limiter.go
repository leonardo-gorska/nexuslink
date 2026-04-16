package redis

import (
	"context"
	"time"

	redisclient "github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redisclient.Client
	script *redisclient.Script
}

// NewRateLimiter Lua script for Sliding Window algorithm
func NewRateLimiter(client *redisclient.Client) *RateLimiter {
	script := redisclient.NewScript(`
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window_ms = tonumber(ARGV[2])
local current_time = tonumber(ARGV[3])
local member = ARGV[4]

local clear_before = current_time - window_ms

redis.call('ZREMRANGEBYSCORE', key, '-inf', clear_before)
local amount = redis.call('ZCARD', key)
if amount < limit then
    redis.call('ZADD', key, current_time, member)
    redis.call('PEXPIRE', key, window_ms)
    return 1
else
    return 0
end
`)
	return &RateLimiter{
		client: client,
		script: script,
	}
}

func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	now := time.Now().UnixMilli()
	windowMs := window.Milliseconds()
	// Time in nano is unique enough for member within a sliding window for rate limiting
	member := time.Now().UnixNano()

	res, err := rl.script.Run(ctx, rl.client, []string{key}, limit, windowMs, now, member).Result()
	if err != nil {
		return false, err
	}

	allowed, ok := res.(int64)
	if !ok {
		return false, nil
	}

	return allowed == 1, nil
}
