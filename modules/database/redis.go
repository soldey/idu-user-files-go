package database

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"main/modules/common"
	"time"
)

type IRedisService interface {
}

type RedisService struct {
	basis IRedisService

	Redis *redis.Client
}

func NewRedisService() RedisService {
	redisService := RedisService{Redis: redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", common.Config.Get("REDIS_HOST"), common.Config.Get("REDIS_PORT")),
		Password: "",
		DB:       0,
	})}
	ctx := context.Background()
	fmt.Printf("Redis: %+v\n", redisService.GetStringList(&ctx))
	return redisService
}

func (s *RedisService) GetStringList(ctx *context.Context) []string {
	res, _ := s.Redis.ScanType(*ctx, 0, "", 200, "STRING").Val()
	return res
}

func (s *RedisService) SetTTL(ctx *context.Context, key string, ttl int64) *redis.BoolCmd {
	return s.Redis.Expire(*ctx, key, time.Duration(ttl*1000000000))
}

func (s *RedisService) SaveBytes(ctx *context.Context, key string, value []byte, ttl int64) *redis.StatusCmd {
	return s.Redis.Set(*ctx, key, value, time.Duration(ttl*1000000000))
}

func (s *RedisService) GetBytes(ctx *context.Context, key string) ([]byte, error) {
	return s.Redis.Get(*ctx, key).Bytes()
}

var Redis RedisService
