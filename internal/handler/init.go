package handler

import "github.com/Jseongwon/FirebaseManagementService.git/internal/redis"

var redisClient *redis.RedisClient

func InitHandler(rc *redis.RedisClient) {
	redisClient = rc
}
