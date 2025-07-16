package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

// 토큰 정보를 저장하기 위한 구조체
type TokenInfo struct {
	Token          string    `json:"token"`
	UserID         string    `json:"user_id"`
	Platform       string    `json:"platform"`
	DeviceModel    string    `json:"device_model"`
	OSVersion      string    `json:"os_version"`
	AppVersion     string    `json:"app_version"`
	DeviceID       string    `json:"device_id"`
	InstallationID string    `json:"installation_id"`
	GeneratedAt    time.Time `json:"generated_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

func NewRedisClient(addr, password string, db int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisClient{client: rdb}
}

// 기존 함수들
func (r *RedisClient) SetFCMToken(ctx context.Context, userID, platform, token string) error {
	key := fmt.Sprintf("fcm:token:%s:%s", userID, platform)
	return r.client.Set(ctx, key, token, 0).Err()
}

func (r *RedisClient) GetFCMToken(ctx context.Context, userID, platform string) (string, error) {
	key := fmt.Sprintf("fcm:token:%s:%s", userID, platform)
	return r.client.Get(ctx, key).Result()
}

// 새로운 함수들 - 장치 정보를 포함한 토큰 저장
func (r *RedisClient) SetFCMTokenWithDeviceInfo(ctx context.Context, tokenInfo TokenInfo) error {
	key := fmt.Sprintf("fcm:token:detailed:%s:%s:%s", tokenInfo.UserID, tokenInfo.Platform, tokenInfo.DeviceID)

	// JSON으로 직렬화
	tokenData, err := json.Marshal(tokenInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal token info: %w", err)
	}

	// 만료 시간 설정 (기본 1년)
	expiration := time.Until(tokenInfo.ExpiresAt)
	if expiration <= 0 {
		expiration = 365 * 24 * time.Hour // 1년
	}

	return r.client.Set(ctx, key, tokenData, expiration).Err()
}

func (r *RedisClient) GetFCMTokenWithDeviceInfo(ctx context.Context, userID, platform, deviceID string) (*TokenInfo, error) {
	key := fmt.Sprintf("fcm:token:detailed:%s:%s:%s", userID, platform, deviceID)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var tokenInfo TokenInfo
	if err := json.Unmarshal([]byte(data), &tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token info: %w", err)
	}

	return &tokenInfo, nil
}

// 사용자의 모든 장치 토큰 조회
func (r *RedisClient) GetAllUserTokens(ctx context.Context, userID string) ([]TokenInfo, error) {
	pattern := fmt.Sprintf("fcm:token:detailed:%s:*", userID)

	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	var tokens []TokenInfo
	for _, key := range keys {
		data, err := r.client.Get(ctx, key).Result()
		if err != nil {
			continue // 개별 토큰 조회 실패는 무시하고 계속 진행
		}

		var tokenInfo TokenInfo
		if err := json.Unmarshal([]byte(data), &tokenInfo); err != nil {
			continue
		}

		tokens = append(tokens, tokenInfo)
	}

	return tokens, nil
}

// 특정 장치의 토큰 삭제
func (r *RedisClient) DeleteFCMToken(ctx context.Context, userID, platform, deviceID string) error {
	key := fmt.Sprintf("fcm:token:detailed:%s:%s:%s", userID, platform, deviceID)
	return r.client.Del(ctx, key).Err()
}
