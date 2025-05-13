package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr, password string, db int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisClient{client: rdb}
}

func (r *RedisClient) SetFCMToken(ctx context.Context, userID, platform, token string) error {
	key := fmt.Sprintf("fcm:token:%s:%s", userID, platform)
	return r.client.Set(ctx, key, token, 0).Err()
}

func (r *RedisClient) GetFCMToken(ctx context.Context, userID, platform string) (string, error) {
	key := fmt.Sprintf("fcm:token:%s:%s", userID, platform)
	return r.client.Get(ctx, key).Result()
}
