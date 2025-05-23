package utils

import (
	"context"

	"premium_cars_app/internal/config"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func SaveToken(rdb *redis.Client, token string) error {
	return rdb.Set(Ctx, token, "1", config.TokenTTL).Err()
}

func RefreshTokenTTL(rdb *redis.Client, token string) error {
	return rdb.Expire(Ctx, token, config.TokenTTL).Err()
}

func TokenExists(rdb *redis.Client, token string) bool {
	val, err := rdb.Exists(Ctx, token).Result()
	return err == nil && val == 1
}
